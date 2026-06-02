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
}

// NewEngine creates a new process engine for a parsed process definition.
func NewEngine(process *ParsedProcess, wasmanEngine *wasman.Engine) *Engine {
	return &Engine{
		Process:      process,
		WasmanEngine: wasmanEngine,
	}
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
		if len(instance.ActiveTokens) == 0 {
			instance.Completed = true
			slog.Info("[BPMN ENGINE] Process instance completed successfully", "instance_id", instance.ID)
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
			// Count how many tokens currently reside on this gateway or have arrived.
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

	case UserTask:
		slog.Info("[BPMN ENGINE] UserTask reached (Wait State)", "node_id", n.ID, "name", n.Name)
		instance.WaitingTokens = append(instance.WaitingTokens, n.ID)

	case ReceiveTask:
		slog.Info("[BPMN ENGINE] ReceiveTask reached (Wait State)", "node_id", n.ID, "name", n.Name)
		instance.WaitingTokens = append(instance.WaitingTokens, n.ID)

	case IntermediateCatchEvent:
		slog.Info("[BPMN ENGINE] IntermediateCatchEvent reached (Wait State)", "node_id", n.ID, "name", n.Name)
		instance.WaitingTokens = append(instance.WaitingTokens, n.ID)
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

	// Create and compile specific WASM engine configuration for this step if task-level compilation is needed
	// For simplicity, we execute via our shared engine using the Instance ID and task parameters
	slog.Info("[BPMN ENGINE] Launching Wasman WASM session", "instance_id", instance.ID, "wasm_path", task.WasmPath, "server", addr)
	crashed, err := e.WasmanEngine.Session(instance.ID).
		WithServer(addr).
		WithCrash(simulateCrash).
		Run(ctx)
	if err != nil {
		return fmt.Errorf("wasman run failed: %w (crashed: %t)", err, crashed)
	}

	// Capture updated variables from worker upload
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
