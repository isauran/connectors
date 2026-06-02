package bpmn

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nativebpm/connectors/wasman"
)

// ProcessInstance represents the state of an active process run.
type ProcessInstance struct {
	ID            string                 `json:"id"`
	ProcessID     string                 `json:"process_id"`
	ActiveTokens  []string               `json:"active_tokens"`  // Elements holding active tokens
	WaitingTokens []string               `json:"waiting_tokens"` // Tokens waiting for external events (e.g. UserTask)
	Variables     map[string]interface{} `json:"variables"`      // Process context variables
	Completed     bool                   `json:"completed"`
}

// Engine coordinates the execution of BPMN processes.
type Engine struct {
	Process      *ParsedProcess
	WasmanEngine *wasman.Engine
	DMNs         map[string]*DMNDefinitions // decisionRef -> DMNDefinitions
}

// NewEngine creates a new process engine for a parsed process definition.
func NewEngine(process *ParsedProcess, wasmanEngine *wasman.Engine) *Engine {
	return &Engine{
		Process:      process,
		WasmanEngine: wasmanEngine,
		DMNs:         make(map[string]*DMNDefinitions),
	}
}

// RegisterDMN registers a parsed DMN definition to be evaluated by BusinessRuleTasks.
func (e *Engine) RegisterDMN(decisionRef string, dmn *DMNDefinitions) {
	e.DMNs[decisionRef] = dmn
}

// StartInstance starts a new process instance.
func (e *Engine) StartInstance(id string, variables map[string]interface{}) (*ProcessInstance, error) {
	if e.Process.StartNodeID == "" {
		return nil, fmt.Errorf("process definitions lack a start event")
	}

	if variables == nil {
		variables = make(map[string]interface{})
	}

	instance := &ProcessInstance{
		ID:           id,
		ProcessID:    e.Process.ID,
		ActiveTokens: []string{e.Process.StartNodeID},
		Variables:    variables,
		Completed:    false,
	}

	return instance, nil
}

// Step advances the execution of active tokens in the process instance.
func (e *Engine) Step(ctx context.Context, instance *ProcessInstance) error {
	if instance.Completed {
		return nil
	}
	if len(instance.ActiveTokens) == 0 {
		if len(instance.WaitingTokens) == 0 {
			instance.Completed = true
		}
		return nil
	}

	// We pop the first active token to process it.
	currentToken := instance.ActiveTokens[0]
	instance.ActiveTokens = instance.ActiveTokens[1:]

	node, exists := e.Process.Nodes[currentToken]
	if !exists {
		return fmt.Errorf("node with ID %s not found in process definition", currentToken)
	}

	switch n := node.(type) {
	case StartEvent:
		slog.Info("[BPMN ENGINE] StartEvent triggered", "node_id", n.ID)
		e.moveToken(instance, n.ID, "")

	case EndEvent:
		slog.Info("[BPMN ENGINE] EndEvent reached", "node_id", n.ID)
		if parentSubID, ok := e.Process.ParentSubProcesses[n.ID]; ok && parentSubID != "" {
			slog.Info("[BPMN ENGINE] Subprocess EndEvent reached, returning to parent flow", "subprocess_id", parentSubID)
			e.moveToken(instance, parentSubID, "")
		} else {
			if len(instance.ActiveTokens) == 0 {
				if len(instance.WaitingTokens) == 0 {
					instance.Completed = true
				}
				slog.Info("[BPMN ENGINE] Process instance completed successfully", "instance_id", instance.ID)
			}
		}

	case ServiceTask:
		slog.Info("[BPMN ENGINE] Executing ServiceTask", "node_id", n.ID, "name", n.Name)
		if n.WasmPath != "" && e.WasmanEngine != nil {
			err := e.executeWasmTask(ctx, instance, n)
			if err != nil {
				// Re-insert token so we can retry or investigate
				instance.ActiveTokens = append([]string{n.ID}, instance.ActiveTokens...)
				return fmt.Errorf("failed to execute WASM task %s: %w", n.ID, err)
			}
		} else {
			slog.Info("[BPMN ENGINE] ServiceTask executed as noop (no WASM path configured)", "node_id", n.ID)
		}
		e.moveToken(instance, n.ID, "")

	case UserTask:
		slog.Info("[BPMN ENGINE] UserTask reached (Wait State)", "node_id", n.ID, "name", n.Name)
		instance.WaitingTokens = append(instance.WaitingTokens, n.ID)

	case ReceiveTask:
		slog.Info("[BPMN ENGINE] ReceiveTask reached (Wait State)", "node_id", n.ID, "name", n.Name)
		instance.WaitingTokens = append(instance.WaitingTokens, n.ID)

	case IntermediateCatchEvent:
		slog.Info("[BPMN ENGINE] IntermediateCatchEvent reached (Wait State)", "node_id", n.ID, "name", n.Name)
		instance.WaitingTokens = append(instance.WaitingTokens, n.ID)

	case BusinessRuleTask:
		slog.Info("[BPMN ENGINE] Executing BusinessRuleTask", "node_id", n.ID, "decision_ref", n.DecisionRef)
		if dmn, ok := e.DMNs[n.DecisionRef]; ok {
			res, err := Evaluate(dmn, n.DecisionRef, instance.Variables)
			if err != nil {
				return fmt.Errorf("failed to evaluate DMN %s: %w", n.DecisionRef, err)
			}
			if n.ResultVariable != "" {
				if n.MapDecisionResult == "singleEntry" {
					for _, v := range res {
						instance.Variables[n.ResultVariable] = v
						break
					}
				} else {
					instance.Variables[n.ResultVariable] = res
				}
			} else {
				for k, v := range res {
					instance.Variables[k] = v
				}
			}
		} else {
			slog.Warn("[BPMN ENGINE] BusinessRuleTask: DMN not registered, execution skipped", "decision_ref", n.DecisionRef)
		}
		e.moveToken(instance, n.ID, "")

	case *SubProcess:
		slog.Info("[BPMN ENGINE] Entering SubProcess", "node_id", n.ID, "name", n.Name)
		if n.StartNodeID != "" {
			instance.ActiveTokens = append(instance.ActiveTokens, n.StartNodeID)
		} else {
			return fmt.Errorf("subprocess %s has no start event", n.ID)
		}

	case SubProcess:
		slog.Info("[BPMN ENGINE] Entering SubProcess", "node_id", n.ID, "name", n.Name)
		if n.StartNodeID != "" {
			instance.ActiveTokens = append(instance.ActiveTokens, n.StartNodeID)
		} else {
			return fmt.Errorf("subprocess %s has no start event", n.ID)
		}

	case ExclusiveGateway:
		slog.Info("[BPMN ENGINE] ExclusiveGateway (XOR) reached", "node_id", n.ID)
		// Evaluate conditional outflows and select the first true path
		outflows := e.Process.Outflows[n.ID]
		activatedFlow := ""
		for _, flow := range outflows {
			condText := ""
			if flow.ConditionExpression != nil {
				condText = flow.ConditionExpression.Text
			}
			if evaluateCondition(condText, instance.Variables) {
				activatedFlow = flow.TargetRef
				break
			}
		}

		if activatedFlow == "" {
			return fmt.Errorf("no conditional paths evaluated to true at XOR gateway %s", n.ID)
		}
		instance.ActiveTokens = append(instance.ActiveTokens, activatedFlow)

	case ParallelGateway:
		slog.Info("[BPMN ENGINE] ParallelGateway reached", "node_id", n.ID)
		inflows := e.Process.Inflows[n.ID]
		outflows := e.Process.Outflows[n.ID]

		if len(inflows) > 1 {
			// This is a Join gateway. We need to check if tokens from all incoming flows have arrived.
			tokensOnGateway := 1 // including the current one we are processing
			for _, active := range instance.ActiveTokens {
				if active == n.ID {
					tokensOnGateway++
				}
			}

			if tokensOnGateway < len(inflows) {
				// Not all paths have arrived yet. Park this token by pushing it back.
				instance.ActiveTokens = append(instance.ActiveTokens, n.ID)
				slog.Info("[BPMN ENGINE] Parallel Join waiting", "arrived", tokensOnGateway, "required", len(inflows))
				return nil
			}

			// All tokens have arrived! Remove all parked tokens of this gateway from ActiveTokens list.
			var cleanedTokens []string
			for _, active := range instance.ActiveTokens {
				if active != n.ID {
					cleanedTokens = append(cleanedTokens, active)
				}
			}
			instance.ActiveTokens = cleanedTokens
			slog.Info("[BPMN ENGINE] Parallel Join satisfied. Proceeding...")
		}

		// Distribute tokens to all target outflows (Fork or Join continuation)
		for _, flow := range outflows {
			instance.ActiveTokens = append(instance.ActiveTokens, flow.TargetRef)
		}
	}

	return nil
}

// CompleteTask resumes a process instance paused at a wait state (UserTask, ReceiveTask, or IntermediateCatchEvent).
func (e *Engine) CompleteTask(instance *ProcessInstance, nodeID string, variables map[string]interface{}) error {
	found := false
	for i, t := range instance.WaitingTokens {
		if t == nodeID {
			instance.WaitingTokens = append(instance.WaitingTokens[:i], instance.WaitingTokens[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("token for node %s is not currently waiting", nodeID)
	}

	if instance.Variables == nil {
		instance.Variables = make(map[string]interface{})
	}
	for k, v := range variables {
		instance.Variables[k] = v
	}

	slog.Info("[BPMN ENGINE] Resuming execution from wait state", "node_id", nodeID, "instance_id", instance.ID)

	// Move the token forward to target nodes
	e.moveToken(instance, nodeID, "")
	return nil
}

// CorrelateMessage correlates a message to trigger an Event Subprocess, Boundary Event or ReceiveTask.
func (e *Engine) CorrelateMessage(instance *ProcessInstance, messageRef string, variables map[string]interface{}) error {
	if instance.Completed {
		return fmt.Errorf("process instance already completed")
	}

	if instance.Variables == nil {
		instance.Variables = make(map[string]interface{})
	}
	for k, v := range variables {
		instance.Variables[k] = v
	}

	// 1. Check for Event Subprocesses listening to this message.
	for _, node := range e.Process.Nodes {
		if sub, ok := node.(*SubProcess); ok && sub.TriggeredByEvent {
			for _, start := range sub.StartEvents {
				if start.MessageEventDefinition != nil && e.resolveMessageName(start.MessageEventDefinition.MessageRef) == messageRef {
					slog.Info("[BPMN ENGINE] Event Subprocess message start event triggered", "node_id", start.ID, "message_ref", messageRef)

					if start.IsInterrupting {
						slog.Info("[BPMN ENGINE] Interrupting event subprocess: clearing all other tokens")
						instance.ActiveTokens = []string{}
						instance.WaitingTokens = []string{}
					}

					instance.ActiveTokens = append(instance.ActiveTokens, start.ID)
					return nil
				}
			}
		}
	}

	// 2. Check active/waiting tokens on tasks that have a BoundaryEvent attached to them.
	allTokens := append([]string{}, instance.ActiveTokens...)
	allTokens = append(allTokens, instance.WaitingTokens...)

	for _, activeNodeID := range allTokens {
		if bEvents, exists := e.Process.BoundaryEventsByNode[activeNodeID]; exists {
			for _, bEvent := range bEvents {
				if bEvent.MessageEventDefinition != nil && e.resolveMessageName(bEvent.MessageEventDefinition.MessageRef) == messageRef {
					slog.Info("[BPMN ENGINE] Boundary message event triggered", "node_id", bEvent.ID, "attached_to", activeNodeID, "message_ref", messageRef)

					instance.ActiveTokens = removeFromStringSlice(instance.ActiveTokens, activeNodeID)
					instance.WaitingTokens = removeFromStringSlice(instance.WaitingTokens, activeNodeID)

					e.moveToken(instance, bEvent.ID, "")
					return nil
				}
			}
		}
	}

	// 3. Check for regular ReceiveTask or IntermediateCatchEvent in WaitingTokens.
	for i, tID := range instance.WaitingTokens {
		node, exists := e.Process.Nodes[tID]
		if !exists {
			continue
		}

		if rt, ok := node.(ReceiveTask); ok && e.resolveMessageName(rt.MessageRef) == messageRef {
			slog.Info("[BPMN ENGINE] ReceiveTask correlated", "node_id", rt.ID, "message_ref", messageRef)
			instance.WaitingTokens = append(instance.WaitingTokens[:i], instance.WaitingTokens[i+1:]...)
			e.moveToken(instance, rt.ID, "")
			return nil
		}

		if ic, ok := node.(IntermediateCatchEvent); ok && (ic.ID == messageRef || ic.Name == messageRef || e.resolveMessageName(ic.ID) == messageRef) {
			slog.Info("[BPMN ENGINE] IntermediateCatchEvent correlated", "node_id", ic.ID)
			instance.WaitingTokens = append(instance.WaitingTokens[:i], instance.WaitingTokens[i+1:]...)
			e.moveToken(instance, ic.ID, "")
			return nil
		}
	}

	return fmt.Errorf("no event handlers correlated for message %s in this process instance", messageRef)
}

// HandleError propagates a BPMN error thrown by a task to trigger a Boundary Error Event.
func (e *Engine) HandleError(instance *ProcessInstance, nodeID string, errorCode string, variables map[string]interface{}) error {
	if instance.Completed {
		return fmt.Errorf("process instance already completed")
	}

	if instance.Variables == nil {
		instance.Variables = make(map[string]interface{})
	}
	for k, v := range variables {
		instance.Variables[k] = v
	}

	curr := nodeID
	for curr != "" {
		if bEvents, exists := e.Process.BoundaryEventsByNode[curr]; exists {
			for _, bEvent := range bEvents {
				if bEvent.ErrorEventDefinition != nil && (e.resolveErrorCode(bEvent.ErrorEventDefinition.ErrorRef) == errorCode || bEvent.ErrorEventDefinition.ErrorRef == "") {
					slog.Info("[BPMN ENGINE] Boundary error event triggered", "node_id", bEvent.ID, "attached_to", curr, "error_code", errorCode)

					instance.ActiveTokens = e.clearScopeTokens(instance.ActiveTokens, curr)
					instance.WaitingTokens = e.clearScopeTokens(instance.WaitingTokens, curr)

					e.moveToken(instance, bEvent.ID, "")
					return nil
				}
			}
		}
		curr = e.Process.ParentSubProcesses[curr]
	}

	return fmt.Errorf("unhandled BPMN error %s at node %s", errorCode, nodeID)
}

func (e *Engine) clearScopeTokens(tokens []string, scopeID string) []string {
	var result []string
	for _, t := range tokens {
		if t == scopeID || e.Process.IsChildOf(t, scopeID) {
			continue
		}
		result = append(result, t)
	}
	return result
}

func (e *Engine) resolveMessageName(messageRef string) string {
	if name, ok := e.Process.Messages[messageRef]; ok {
		return name
	}
	return messageRef
}

func (e *Engine) resolveErrorCode(errorRef string) string {
	if code, ok := e.Process.Errors[errorRef]; ok {
		return code
	}
	return errorRef
}

func removeFromStringSlice(slice []string, val string) []string {
	var result []string
	for _, item := range slice {
		if item != val {
			result = append(result, item)
		}
	}
	return result
}

func (e *Engine) moveToken(instance *ProcessInstance, sourceNodeID string, selectedTarget string) {
	outflows := e.Process.Outflows[sourceNodeID]
	for _, flow := range outflows {
		if selectedTarget == "" || flow.TargetRef == selectedTarget {
			instance.ActiveTokens = append(instance.ActiveTokens, flow.TargetRef)
		}
	}
}

func (e *Engine) executeWasmTask(ctx context.Context, instance *ProcessInstance, task ServiceTask) error {
	updateChan := make(chan map[string]interface{}, 1)
	addr, server, err := startLocalServer(instance.Variables, updateChan)
	if err != nil {
		return err
	}
	defer server.Shutdown(ctx)

	simulateCrash := false
	if val, ok := instance.Variables["simulate_crash"]; ok {
		if b, ok := val.(bool); ok {
			simulateCrash = b
		}
	}

	slog.Info("[BPMN ENGINE] Launching Wasman WASM session", "instance_id", instance.ID, "wasm_path", task.WasmPath, "server", addr)
	crashed, err := e.WasmanEngine.Session(instance.ID).
		WithServer(addr).
		WithCrash(simulateCrash).
		Run(ctx)
	if err != nil {
		return fmt.Errorf("wasman run failed: %w (crashed: %t)", err, crashed)
	}

	select {
	case updatedVars := <-updateChan:
		for k, v := range updatedVars {
			instance.Variables[k] = v
		}
		slog.Info("[BPMN ENGINE] Variables updated from WASM execution", "vars", updatedVars)
	case <-time.After(100 * time.Millisecond):
		slog.Info("[BPMN ENGINE] WASM executed successfully but made no variable updates")
	}

	return nil
}

func startLocalServer(vars map[string]interface{}, updateChan chan map[string]interface{}) (string, *http.Server, error) {
	mux := http.NewServeMux()

	mux.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(vars)
	})

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		var updated map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updated); err == nil {
			updateChan <- updated
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", nil, err
	}

	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		_ = server.Serve(ln)
	}()

	return ln.Addr().String(), server, nil
}

func evaluateCondition(expr string, vars map[string]interface{}) bool {
	expr = strings.TrimSpace(expr)
	if strings.HasPrefix(expr, "${") && strings.HasSuffix(expr, "}") {
		expr = expr[2 : len(expr)-1]
	}
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return true
	}

	parts := strings.Fields(expr)
	if len(parts) != 3 {
		return false
	}
	varName := parts[0]
	op := parts[1]
	valStr := parts[2]

	val, ok := vars[varName]
	if !ok {
		return false
	}

	switch op {
	case "==":
		strVal := fmt.Sprintf("%v", val)
		return strVal == strings.Trim(valStr, `"'`)
	case "!=":
		strVal := fmt.Sprintf("%v", val)
		return strVal != strings.Trim(valStr, `"'`)
	case ">":
		fVal, err1 := strconv.ParseFloat(fmt.Sprintf("%v", val), 64)
		fLimit, err2 := strconv.ParseFloat(valStr, 64)
		if err1 == nil && err2 == nil {
			return fVal > fLimit
		}
	case "<":
		fVal, err1 := strconv.ParseFloat(fmt.Sprintf("%v", val), 64)
		fLimit, err2 := strconv.ParseFloat(valStr, 64)
		if err1 == nil && err2 == nil {
			return fVal < fLimit
		}
	}

	return false
}

func (pp *ParsedProcess) IsChildOf(nodeID string, parentID string) bool {
	curr := nodeID
	for {
		p, ok := pp.ParentSubProcesses[curr]
		if !ok || p == "" {
			return false
		}
		if p == parentID {
			return true
		}
		curr = p
	}
}
