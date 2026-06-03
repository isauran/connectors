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
	ID             string                 `json:"id"`
	ProcessID      string                 `json:"process_id"`
	ActiveTokens   []string               `json:"active_tokens"`   // Elements holding active tokens
	WaitingTokens  []string               `json:"waiting_tokens"`  // Tokens waiting for external events (e.g. UserTask)
	CompletedTasks []string               `json:"completed_tasks"` // History of executed task IDs (for compensation)
	Variables      map[string]interface{} `json:"variables"`        // Process context variables
	Completed      bool                   `json:"completed"`
}

// ServiceTaskHandler is a function that executes a service task and can read/write process instance variables.
type ServiceTaskHandler func(ctx context.Context, instance *ProcessInstance, task ServiceTask) error

// Engine coordinates the execution of BPMN processes.
type Engine struct {
	Process             *ParsedProcess
	WasmanEngine        *wasman.Engine
	DMNs                map[string]*DMNDefinitions    // decisionRef -> DMNDefinitions
	ServiceTaskHandlers map[string]ServiceTaskHandler // topic -> handler
}

// NewEngine creates a new process engine for a parsed process definition.
func NewEngine(process *ParsedProcess, wasmanEngine *wasman.Engine) *Engine {
	return &Engine{
		Process:             process,
		WasmanEngine:        wasmanEngine,
		DMNs:                make(map[string]*DMNDefinitions),
		ServiceTaskHandlers: make(map[string]ServiceTaskHandler),
	}
}

// RegisterServiceTaskHandler registers a custom handler function for service tasks with a matching Topic name.
func (e *Engine) RegisterServiceTaskHandler(topic string, handler ServiceTaskHandler) {
	e.ServiceTaskHandlers[topic] = handler
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
		ID:             id,
		ProcessID:      e.Process.ID,
		ActiveTokens:   []string{e.Process.StartNodeID},
		CompletedTasks: []string{},
		Variables:      variables,
		Completed:      false,
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

	// Virtual Multi-Instance Join Handling
	if strings.HasSuffix(currentToken, "#join") {
		baseID := strings.TrimSuffix(currentToken, "#join")
		requiredCount := e.getMultiInstanceCount(instance, baseID)
		
		// Count how many join tokens have accumulated
		joinCount := 1 // including the current one
		for _, t := range instance.ActiveTokens {
			if t == currentToken {
				joinCount++
			}
		}

		if joinCount < requiredCount {
			// Not all parallel multi-instance tasks finished yet. Repark this token at the end of active list.
			instance.ActiveTokens = append(instance.ActiveTokens, currentToken)
			slog.Info("[BPMN ENGINE] Multi-Instance Join waiting", "node_id", baseID, "arrived", joinCount, "required", requiredCount)
			return nil
		}

		// All copies completed! Clean all join tokens of this task from ActiveTokens.
		instance.ActiveTokens = removeAllFromStringSlice(instance.ActiveTokens, currentToken)
		slog.Info("[BPMN ENGINE] Multi-Instance Join satisfied. Proceeding...", "node_id", baseID)
		
		e.recordCompletedTask(instance, baseID)
		e.moveToken(instance, baseID, "")
		return nil
	}

	// Base Node ID resolution for Multi-Instance runs (e.g. task_id#0 -> task_id)
	baseNodeID := currentToken
	isMultiInstanceCopy := false
	if strings.Contains(currentToken, "#") {
		baseNodeID = strings.Split(currentToken, "#")[0]
		isMultiInstanceCopy = true
	}

	node, exists := e.Process.Nodes[baseNodeID]
	if !exists {
		return fmt.Errorf("node with ID %s not found in process definition", baseNodeID)
	}

	// Multi-Instance entry initialization (only if it is not already a copy)
	if !isMultiInstanceCopy {
		miChar := e.getMultiInstanceCharacteristics(node)
		if miChar != nil {
			count := e.getMultiInstanceCount(instance, baseNodeID)
			slog.Info("[BPMN ENGINE] Initializing Multi-Instance Task", "node_id", baseNodeID, "copies", count)
			
			if count == 0 {
				// Empty collection: skip task completely and move forward
				e.recordCompletedTask(instance, baseNodeID)
				e.moveToken(instance, baseNodeID, "")
				return nil
			}

			// Generate N parallel copies: node_id#0, node_id#1, ...
			for i := 0; i < count; i++ {
				copyToken := fmt.Sprintf("%s#%d", baseNodeID, i)
				instance.ActiveTokens = append(instance.ActiveTokens, copyToken)
			}
			return nil
		}
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
		slog.Info("[BPMN ENGINE] Executing ServiceTask", "node_id", currentToken, "base_node", n.ID, "name", n.Name)
		if handler, ok := e.ServiceTaskHandlers[n.Topic]; ok {
			slog.Info("[BPMN ENGINE] Executing local service task handler", "topic", n.Topic, "node_id", currentToken)
			err := handler(ctx, instance, n)
			if err != nil {
				// Re-insert token so we can retry or investigate
				instance.ActiveTokens = append([]string{currentToken}, instance.ActiveTokens...)
				return fmt.Errorf("failed to execute local handler for service task %s (topic: %s): %w", currentToken, n.Topic, err)
			}
		} else if n.WasmPath != "" && e.WasmanEngine != nil {
			err := e.executeWasmTask(ctx, instance, n)
			if err != nil {
				// Re-insert token so we can retry or investigate
				instance.ActiveTokens = append([]string{currentToken}, instance.ActiveTokens...)
				return fmt.Errorf("failed to execute WASM task %s: %w", currentToken, err)
			}
		} else {
			slog.Info("[BPMN ENGINE] ServiceTask executed as noop (no local handler or WASM path configured)", "node_id", currentToken)
		}
		
		if isMultiInstanceCopy {
			// Complete this copy and redirect to the virtual join
			instance.ActiveTokens = append(instance.ActiveTokens, fmt.Sprintf("%s#join", n.ID))
		} else {
			e.recordCompletedTask(instance, n.ID)
			e.moveToken(instance, n.ID, "")
		}

	case UserTask:
		slog.Info("[BPMN ENGINE] UserTask reached (Wait State)", "node_id", currentToken, "name", n.Name)
		instance.WaitingTokens = append(instance.WaitingTokens, currentToken)

	case ReceiveTask:
		slog.Info("[BPMN ENGINE] ReceiveTask reached (Wait State)", "node_id", currentToken, "name", n.Name)
		instance.WaitingTokens = append(instance.WaitingTokens, currentToken)

	case IntermediateCatchEvent:
		slog.Info("[BPMN ENGINE] IntermediateCatchEvent reached (Wait State)", "node_id", currentToken, "name", n.Name)
		instance.WaitingTokens = append(instance.WaitingTokens, currentToken)

	case BusinessRuleTask:
		slog.Info("[BPMN ENGINE] Executing BusinessRuleTask", "node_id", currentToken, "decision_ref", n.DecisionRef)
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
		
		if isMultiInstanceCopy {
			instance.ActiveTokens = append(instance.ActiveTokens, fmt.Sprintf("%s#join", n.ID))
		} else {
			e.recordCompletedTask(instance, n.ID)
			e.moveToken(instance, n.ID, "")
		}

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
			// Join gateway logic
			tokensOnGateway := 1
			for _, active := range instance.ActiveTokens {
				if active == n.ID {
					tokensOnGateway++
				}
			}

			if tokensOnGateway < len(inflows) {
				instance.ActiveTokens = append(instance.ActiveTokens, n.ID)
				slog.Info("[BPMN ENGINE] Parallel Join waiting", "arrived", tokensOnGateway, "required", len(inflows))
				return nil
			}

			instance.ActiveTokens = removeAllFromStringSlice(instance.ActiveTokens, n.ID)
			slog.Info("[BPMN ENGINE] Parallel Join satisfied. Proceeding...")
		}

		for _, flow := range outflows {
			instance.ActiveTokens = append(instance.ActiveTokens, flow.TargetRef)
		}

	case InclusiveGateway:
		slog.Info("[BPMN ENGINE] InclusiveGateway (OR) reached", "node_id", n.ID)
		inflows := e.Process.Inflows[n.ID]
		outflows := e.Process.Outflows[n.ID]

		if len(inflows) > 1 {
			// OR Join logic: wait if there's any other token that can reach this gateway
			canMoreTokensArrive := false
			for _, active := range instance.ActiveTokens {
				if active != n.ID && e.Process.HasPath(active, n.ID) {
					canMoreTokensArrive = true
					break
				}
			}
			if !canMoreTokensArrive {
				for _, waiting := range instance.WaitingTokens {
					if waiting != n.ID && e.Process.HasPath(waiting, n.ID) {
						canMoreTokensArrive = true
						break
					}
				}
			}

			if canMoreTokensArrive {
				instance.ActiveTokens = append(instance.ActiveTokens, n.ID)
				slog.Info("[BPMN ENGINE] Inclusive Join waiting: more tokens can arrive", "node_id", n.ID)
				return nil
			}

			instance.ActiveTokens = removeAllFromStringSlice(instance.ActiveTokens, n.ID)
			slog.Info("[BPMN ENGINE] Inclusive Join satisfied. Proceeding...", "node_id", n.ID)
		}

		// OR Split logic: activate ALL branches that evaluate to true
		activatedAny := false
		for _, flow := range outflows {
			condText := ""
			if flow.ConditionExpression != nil {
				condText = flow.ConditionExpression.Text
			}
			if evaluateCondition(condText, instance.Variables) {
				instance.ActiveTokens = append(instance.ActiveTokens, flow.TargetRef)
				activatedAny = true
			}
		}

		// Fallback to first flow if none evaluated to true
		if !activatedAny && len(outflows) > 0 {
			slog.Info("[BPMN ENGINE] Inclusive Split fallback to first outflow", "node_id", n.ID)
			instance.ActiveTokens = append(instance.ActiveTokens, outflows[0].TargetRef)
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

	baseNodeID := nodeID
	if strings.Contains(nodeID, "#") {
		baseNodeID = strings.Split(nodeID, "#")[0]
	}

	if strings.Contains(nodeID, "#") {
		// Multi-instance copy completed
		instance.ActiveTokens = append(instance.ActiveTokens, fmt.Sprintf("%s#join", baseNodeID))
	} else {
		e.recordCompletedTask(instance, baseNodeID)
		e.moveToken(instance, nodeID, "")
	}
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
		baseID := activeNodeID
		if strings.Contains(activeNodeID, "#") {
			baseID = strings.Split(activeNodeID, "#")[0]
		}
		if bEvents, exists := e.Process.BoundaryEventsByNode[baseID]; exists {
			for _, bEvent := range bEvents {
				if bEvent.MessageEventDefinition != nil && e.resolveMessageName(bEvent.MessageEventDefinition.MessageRef) == messageRef {
					slog.Info("[BPMN ENGINE] Boundary message event triggered", "node_id", bEvent.ID, "attached_to", activeNodeID, "message_ref", messageRef)

					if bEvent.IsInterrupting() {
						instance.ActiveTokens = removeFromStringSlice(instance.ActiveTokens, activeNodeID)
						instance.WaitingTokens = removeFromStringSlice(instance.WaitingTokens, activeNodeID)
					}

					e.moveToken(instance, bEvent.ID, "")
					return nil
				}
			}
		}
	}

	// 3. Check for regular ReceiveTask or IntermediateCatchEvent in WaitingTokens.
	for i, tID := range instance.WaitingTokens {
		baseID := tID
		if strings.Contains(tID, "#") {
			baseID = strings.Split(tID, "#")[0]
		}
		node, exists := e.Process.Nodes[baseID]
		if !exists {
			continue
		}

		if rt, ok := node.(ReceiveTask); ok && e.resolveMessageName(rt.MessageRef) == messageRef {
			slog.Info("[BPMN ENGINE] ReceiveTask correlated", "node_id", tID, "message_ref", messageRef)
			instance.WaitingTokens = append(instance.WaitingTokens[:i], instance.WaitingTokens[i+1:]...)
			
			if strings.Contains(tID, "#") {
				instance.ActiveTokens = append(instance.ActiveTokens, fmt.Sprintf("%s#join", baseID))
			} else {
				e.recordCompletedTask(instance, baseID)
				e.moveToken(instance, rt.ID, "")
			}
			return nil
		}

		if ic, ok := node.(IntermediateCatchEvent); ok && (ic.ID == messageRef || ic.Name == messageRef || e.resolveMessageName(ic.ID) == messageRef) {
			slog.Info("[BPMN ENGINE] IntermediateCatchEvent correlated", "node_id", tID)
			instance.WaitingTokens = append(instance.WaitingTokens[:i], instance.WaitingTokens[i+1:]...)
			e.moveToken(instance, ic.ID, "")
			return nil
		}
	}

	return fmt.Errorf("no event handlers correlated for message %s in this process instance", messageRef)
}

// BroadcastSignal broadcasts a signal to all listening handlers (triggers StartEvents, BoundaryEvents, and CatchEvents).
func (e *Engine) BroadcastSignal(instance *ProcessInstance, signalRef string, variables map[string]interface{}) error {
	if instance.Completed {
		return fmt.Errorf("process instance already completed")
	}

	if instance.Variables == nil {
		instance.Variables = make(map[string]interface{})
	}
	for k, v := range variables {
		instance.Variables[k] = v
	}

	triggered := false

	// 1. Check for Event Subprocesses listening to this signal.
	for _, node := range e.Process.Nodes {
		if sub, ok := node.(*SubProcess); ok && sub.TriggeredByEvent {
			for _, start := range sub.StartEvents {
				if start.SignalEventDefinition != nil && e.resolveSignalName(start.SignalEventDefinition.SignalRef) == signalRef {
					slog.Info("[BPMN ENGINE] Event Subprocess signal start event triggered", "node_id", start.ID, "signal", signalRef)

					if start.IsInterrupting {
						slog.Info("[BPMN ENGINE] Interrupting event subprocess: clearing all other tokens")
						instance.ActiveTokens = []string{}
						instance.WaitingTokens = []string{}
					}

					instance.ActiveTokens = append(instance.ActiveTokens, start.ID)
					triggered = true
				}
			}
		}
	}

	// 2. Check active/waiting tokens on tasks that have a Signal BoundaryEvent attached to them.
	allTokens := append([]string{}, instance.ActiveTokens...)
	allTokens = append(allTokens, instance.WaitingTokens...)

	for _, activeNodeID := range allTokens {
		baseID := activeNodeID
		if strings.Contains(activeNodeID, "#") {
			baseID = strings.Split(activeNodeID, "#")[0]
		}
		if bEvents, exists := e.Process.BoundaryEventsByNode[baseID]; exists {
			for _, bEvent := range bEvents {
				if bEvent.SignalEventDefinition != nil && e.resolveSignalName(bEvent.SignalEventDefinition.SignalRef) == signalRef {
					slog.Info("[BPMN ENGINE] Boundary signal event triggered", "node_id", bEvent.ID, "attached_to", activeNodeID, "signal", signalRef)

					if bEvent.IsInterrupting() {
						instance.ActiveTokens = removeFromStringSlice(instance.ActiveTokens, activeNodeID)
						instance.WaitingTokens = removeFromStringSlice(instance.WaitingTokens, activeNodeID)
					}

					e.moveToken(instance, bEvent.ID, "")
					triggered = true
				}
			}
		}
	}

	// 2. Check for IntermediateCatchEvent / ReceiveTask waiting for signals in WaitingTokens.
	var remainingWaiting []string
	for _, tID := range instance.WaitingTokens {
		baseID := tID
		if strings.Contains(tID, "#") {
			baseID = strings.Split(tID, "#")[0]
		}
		node, exists := e.Process.Nodes[baseID]
		if !exists {
			remainingWaiting = append(remainingWaiting, tID)
			continue
		}

		matched := false
		if ic, ok := node.(IntermediateCatchEvent); ok && ic.SignalEventDefinition != nil && e.resolveSignalName(ic.SignalEventDefinition.SignalRef) == signalRef {
			slog.Info("[BPMN ENGINE] IntermediateCatchEvent signal correlated", "node_id", tID)
			e.moveToken(instance, ic.ID, "")
			matched = true
			triggered = true
		} else if rt, ok := node.(ReceiveTask); ok && rt.MultiInstanceCharacteristics == nil { // Simple ReceiveTask waiting for signal ID as fallback
			// For flexibility: match by SignalRef if defined, otherwise skip
		}

		if !matched {
			remainingWaiting = append(remainingWaiting, tID)
		}
	}
	instance.WaitingTokens = remainingWaiting

	if !triggered {
		slog.Warn("[BPMN ENGINE] BroadcastSignal: no listeners found for signal", "signal", signalRef)
	}
	return nil
}

// TriggerCompensation executes the compensation activity associated with the completed task.
func (e *Engine) TriggerCompensation(instance *ProcessInstance, activityID string) error {
	// Find boundary compensation event associated with the completed activity
	var boundaryCompEventID string
	if bEvents, exists := e.Process.BoundaryEventsByNode[activityID]; exists {
		for _, bEvent := range bEvents {
			// A boundary event is a compensation event if it contains no other definitions, or is targeted by a compensation association
			// In BPMN XML, compensation boundary events usually have no error/timer/message definitions
			if bEvent.ErrorEventDefinition == nil && bEvent.TimerEventDefinition == nil && bEvent.MessageEventDefinition == nil && bEvent.SignalEventDefinition == nil {
				boundaryCompEventID = bEvent.ID
				break
			}
		}
	}

	if boundaryCompEventID == "" {
		return fmt.Errorf("no compensation boundary event attached to activity %s", activityID)
	}

	// Find the association targeting the compensation activity from this boundary event
	var targetCompensationActivityID string
	for _, assoc := range e.Process.Associations {
		if assoc.SourceRef == boundaryCompEventID {
			targetCompensationActivityID = assoc.TargetRef
			break
		}
	}

	if targetCompensationActivityID == "" {
		return fmt.Errorf("no compensation association found for boundary event %s", boundaryCompEventID)
	}

	// Verify if the target task was indeed executed in the past
	executed := false
	for _, id := range instance.CompletedTasks {
		if id == activityID {
			executed = true
			break
		}
	}

	if !executed {
		slog.Info("[BPMN ENGINE] Skipping compensation: activity was not executed", "activity_id", activityID)
		return nil
	}

	slog.Info("[BPMN ENGINE] Triggering compensation activity", "activity_id", targetCompensationActivityID, "for_completed_task", activityID)
	// Add target compensation activity to active tokens list to run it
	instance.ActiveTokens = append(instance.ActiveTokens, targetCompensationActivityID)
	return nil
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
	if strings.Contains(nodeID, "#") {
		curr = strings.Split(nodeID, "#")[0]
	}

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
		baseID := t
		if strings.Contains(t, "#") {
			baseID = strings.Split(t, "#")[0]
		}
		if baseID == scopeID || e.Process.IsChildOf(baseID, scopeID) {
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

func (e *Engine) resolveSignalName(signalRef string) string {
	if name, ok := e.Process.Signals[signalRef]; ok {
		return name
	}
	return signalRef
}

func (e *Engine) recordCompletedTask(instance *ProcessInstance, nodeID string) {
	for _, id := range instance.CompletedTasks {
		if id == nodeID {
			return
		}
	}
	instance.CompletedTasks = append(instance.CompletedTasks, nodeID)
}

func (e *Engine) getMultiInstanceCharacteristics(node interface{}) *MultiInstanceLoopCharacteristics {
	switch n := node.(type) {
	case ServiceTask:
		return n.MultiInstanceCharacteristics
	case UserTask:
		return n.MultiInstanceCharacteristics
	case ReceiveTask:
		return n.MultiInstanceCharacteristics
	case BusinessRuleTask:
		return n.MultiInstanceCharacteristics
	}
	return nil
}

func (e *Engine) getMultiInstanceCount(instance *ProcessInstance, baseID string) int {
	mi := e.getMultiInstanceCharacteristics(e.Process.Nodes[baseID])
	if mi == nil || mi.LoopDataInputRef == "" {
		return 1
	}
	val, ok := instance.Variables[mi.LoopDataInputRef]
	if !ok {
		return 0
	}
	// Try parsing length by serializing to JSON
	var arr []interface{}
	bytesData, _ := json.Marshal(val)
	if json.Unmarshal(bytesData, &arr) == nil {
		return len(arr)
	}
	return 0
}

func (pp *ParsedProcess) HasPath(src, dest string) bool {
	visited := make(map[string]bool)
	queue := []string{src}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		if curr == dest {
			return true
		}
		if visited[curr] {
			continue
		}
		visited[curr] = true
		for _, flow := range pp.Outflows[curr] {
			queue = append(queue, flow.TargetRef)
		}
	}
	return false
}

func removeAllFromStringSlice(slice []string, val string) []string {
	var result []string
	for _, item := range slice {
		if item != val {
			result = append(result, item)
		}
	}
	return result
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
	baseID := sourceNodeID
	if strings.Contains(sourceNodeID, "#") {
		baseID = strings.Split(sourceNodeID, "#")[0]
	}
	outflows := e.Process.Outflows[baseID]
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
