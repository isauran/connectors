package bpmn

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

// Definitions represents the root element of a BPMN XML document.
type Definitions struct {
	XMLName   xml.Name  `xml:"definitions"`
	Processes []Process `xml:"process"`
	Messages  []Message `xml:"message"`
	Errors    []Error   `xml:"error"`
}

// Message represents a global BPMN message definition.
type Message struct {
	ID   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
}

// Error represents a global BPMN error definition.
type Error struct {
	ID        string `xml:"id,attr"`
	Name      string `xml:"name,attr"`
	ErrorCode string `xml:"errorCode,attr"`
}

// Process represents a BPMN process definition.
type Process struct {
	ID                      string                   `xml:"id,attr"`
	Name                    string                   `xml:"name,attr"`
	IsExecutable            bool                     `xml:"isExecutable,attr"`
	StartEvents             []StartEvent             `xml:"startEvent"`
	EndEvents               []EndEvent               `xml:"endEvent"`
	ServiceTasks            []ServiceTask            `xml:"serviceTask"`
	Tasks                   []ServiceTask            `xml:"task"`
	UserTasks               []UserTask               `xml:"userTask"`
	ReceiveTasks            []ReceiveTask            `xml:"receiveTask"`
	IntermediateCatchEvents []IntermediateCatchEvent `xml:"intermediateCatchEvent"`
	SequenceFlows           []SequenceFlow           `xml:"sequenceFlow"`
	ExclusiveGateways       []ExclusiveGateway       `xml:"exclusiveGateway"`
	ParallelGateways        []ParallelGateway        `xml:"parallelGateway"`
	BusinessRuleTasks       []BusinessRuleTask       `xml:"businessRuleTask"`
	BoundaryEvents          []BoundaryEvent          `xml:"boundaryEvent"`
	SubProcesses            []SubProcess             `xml:"subProcess"`
}

// StartEvent represents a BPMN start event.
type StartEvent struct {
	ID                     string                  `xml:"id,attr"`
	Name                   string                  `xml:"name,attr"`
	IsInterrupting         bool                    `xml:"isInterrupting,attr"` // for event subprocess start events
	MessageEventDefinition *MessageEventDefinition `xml:"messageEventDefinition"`
}

// EndEvent represents a BPMN end event.
type EndEvent struct {
	ID   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
}

// ServiceTask represents a BPMN service task.
type ServiceTask struct {
	ID           string `xml:"id,attr"`
	Name         string `xml:"name,attr"`
	Topic        string `xml:"topic,attr"`
	WasmPath     string `xml:"wasmPath,attr"` // Custom attribute for wasman orchestration
	CamundaTopic string `xml:"type,attr"`
}

// UserTask represents a BPMN user task (wait state).
type UserTask struct {
	ID   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
}

// ReceiveTask represents a BPMN receive task (wait state for a message).
type ReceiveTask struct {
	ID         string `xml:"id,attr"`
	Name       string `xml:"name,attr"`
	MessageRef string `xml:"messageRef,attr"`
}

// IntermediateCatchEvent represents an intermediate catch event (e.g. message wait).
type IntermediateCatchEvent struct {
	ID   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
}

// BusinessRuleTask represents a BPMN business rule task calling DMN.
type BusinessRuleTask struct {
	ID                string `xml:"id,attr"`
	Name              string `xml:"name,attr"`
	DecisionRef       string `xml:"decisionRef,attr"`
	MapDecisionResult string `xml:"mapDecisionResult,attr"`
	ResultVariable    string `xml:"resultVariable,attr"`
}

// BoundaryEvent represents a BPMN boundary event attached to a task.
type BoundaryEvent struct {
	ID                     string                  `xml:"id,attr"`
	Name                   string                  `xml:"name,attr"`
	AttachedToRef          string                  `xml:"attachedToRef,attr"`
	TimerEventDefinition   *TimerEventDefinition   `xml:"timerEventDefinition"`
	MessageEventDefinition *MessageEventDefinition `xml:"messageEventDefinition"`
	ErrorEventDefinition   *ErrorEventDefinition   `xml:"errorEventDefinition"`
}

// TimerEventDefinition defines a timer trigger duration.
type TimerEventDefinition struct {
	TimeDuration string `xml:"timeDuration"`
}

// MessageEventDefinition defines a message correlation reference.
type MessageEventDefinition struct {
	MessageRef string `xml:"messageRef,attr"`
}

// ErrorEventDefinition defines an error catch code.
type ErrorEventDefinition struct {
	ErrorRef string `xml:"errorRef,attr"`
}

// SubProcess represents an embedded subprocess block.
type SubProcess struct {
	ID                      string                   `xml:"id,attr"`
	Name                    string                   `xml:"name,attr"`
	TriggeredByEvent        bool                     `xml:"triggeredByEvent,attr"`
	StartEvents             []StartEvent             `xml:"startEvent"`
	EndEvents               []EndEvent               `xml:"endEvent"`
	ServiceTasks            []ServiceTask            `xml:"serviceTask"`
	Tasks                   []ServiceTask            `xml:"task"`
	UserTasks               []UserTask               `xml:"userTask"`
	ReceiveTasks            []ReceiveTask            `xml:"receiveTask"`
	IntermediateCatchEvents []IntermediateCatchEvent `xml:"intermediateCatchEvent"`
	SequenceFlows           []SequenceFlow           `xml:"sequenceFlow"`
	ExclusiveGateways       []ExclusiveGateway       `xml:"exclusiveGateway"`
	ParallelGateways        []ParallelGateway        `xml:"parallelGateway"`
	BusinessRuleTasks       []BusinessRuleTask       `xml:"businessRuleTask"`
	SubProcesses            []SubProcess             `xml:"subProcess"`
	BoundaryEvents          []BoundaryEvent          `xml:"boundaryEvent"`

	// Internal execution cache
	StartNodeID             string
}

// SequenceFlow represents a BPMN transition flow between two elements.
type SequenceFlow struct {
	ID                  string               `xml:"id,attr"`
	SourceRef           string               `xml:"sourceRef,attr"`
	TargetRef           string               `xml:"targetRef,attr"`
	ConditionExpression *ConditionExpression `xml:"conditionExpression"`
}

// ConditionExpression represents formal conditions on transitions.
type ConditionExpression struct {
	Type string `xml:"type,attr"`
	Text string `xml:",chardata"`
}

// ExclusiveGateway represents an exclusive decision point (XOR gateway).
type ExclusiveGateway struct {
	ID   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
}

// ParallelGateway represents an parallel execution join/fork point (AND gateway).
type ParallelGateway struct {
	ID   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
}

// ParsedProcess is an indexed representation of a process layout for fast navigation.
type ParsedProcess struct {
	ID                   string
	Name                 string
	StartNodeID          string
	Nodes                map[string]interface{}
	Outflows             map[string][]SequenceFlow
	Inflows              map[string][]SequenceFlow
	ParentSubProcesses   map[string]string          // nodeID -> parentSubProcessID
	SubProcesses         map[string]*SubProcess     // subProcessID -> SubProcess
	BoundaryEventsByNode map[string][]BoundaryEvent // nodeID -> BoundaryEvents
	Messages             map[string]string          // MessageID -> MessageName
	Errors               map[string]string          // ErrorID -> ErrorCode
}

// ParseBPMN parses raw BPMN XML data and indexes the first executable process found.
func ParseBPMN(xmlData []byte) (*ParsedProcess, error) {
	var defs Definitions
	dec := xml.NewDecoder(bytes.NewReader(xmlData))
	dec.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		return input, nil
	}
	err := dec.Decode(&defs)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal BPMN XML: %w", err)
	}

	for _, p := range defs.Processes {
		if p.IsExecutable {
			return indexProcess(&p, &defs)
		}
	}

	// fallback to first process if none are marked executable
	if len(defs.Processes) > 0 {
		return indexProcess(&defs.Processes[0], &defs)
	}

	return nil, fmt.Errorf("no processes found in BPMN document")
}

func indexProcess(p *Process, defs *Definitions) (*ParsedProcess, error) {
	pp := &ParsedProcess{
		ID:                   p.ID,
		Name:                 p.Name,
		Nodes:                make(map[string]interface{}),
		Outflows:             make(map[string][]SequenceFlow),
		Inflows:              make(map[string][]SequenceFlow),
		ParentSubProcesses:   make(map[string]string),
		SubProcesses:         make(map[string]*SubProcess),
		BoundaryEventsByNode: make(map[string][]BoundaryEvent),
		Messages:             make(map[string]string),
		Errors:               make(map[string]string),
	}

	// Index global messages and errors
	for _, msg := range defs.Messages {
		pp.Messages[msg.ID] = msg.Name
	}
	for _, err := range defs.Errors {
		pp.Errors[err.ID] = err.ErrorCode
	}

	// 1. Index all nodes
	for _, n := range p.StartEvents {
		pp.Nodes[n.ID] = n
		if pp.StartNodeID == "" {
			pp.StartNodeID = n.ID
		}
	}
	for _, n := range p.EndEvents {
		pp.Nodes[n.ID] = n
	}
	for _, n := range p.ServiceTasks {
		pp.Nodes[n.ID] = n
	}
	for _, n := range p.Tasks {
		pp.Nodes[n.ID] = n
	}
	for _, n := range p.UserTasks {
		pp.Nodes[n.ID] = n
	}
	for _, n := range p.ReceiveTasks {
		pp.Nodes[n.ID] = n
	}
	for _, n := range p.IntermediateCatchEvents {
		pp.Nodes[n.ID] = n
	}
	for _, n := range p.ExclusiveGateways {
		pp.Nodes[n.ID] = n
	}
	for _, n := range p.ParallelGateways {
		pp.Nodes[n.ID] = n
	}
	for _, n := range p.BusinessRuleTasks {
		pp.Nodes[n.ID] = n
	}
	for _, n := range p.BoundaryEvents {
		pp.Nodes[n.ID] = n
		pp.BoundaryEventsByNode[n.AttachedToRef] = append(pp.BoundaryEventsByNode[n.AttachedToRef], n)
	}

	// Index Subprocesses
	for _, sub := range p.SubProcesses {
		indexSubProcess(&sub, pp, "")
	}

	// 2. Index outflows and inflows
	for _, flow := range p.SequenceFlows {
		pp.Outflows[flow.SourceRef] = append(pp.Outflows[flow.SourceRef], flow)
		pp.Inflows[flow.TargetRef] = append(pp.Inflows[flow.TargetRef], flow)
	}

	return pp, nil
}

func indexSubProcess(sub *SubProcess, pp *ParsedProcess, parentID string) {
	pp.SubProcesses[sub.ID] = sub
	pp.Nodes[sub.ID] = sub
	if parentID != "" {
		pp.ParentSubProcesses[sub.ID] = parentID
	}

	for _, n := range sub.StartEvents {
		pp.Nodes[n.ID] = n
		pp.ParentSubProcesses[n.ID] = sub.ID
		if sub.StartNodeID == "" {
			sub.StartNodeID = n.ID
		}
	}
	for _, n := range sub.EndEvents {
		pp.Nodes[n.ID] = n
		pp.ParentSubProcesses[n.ID] = sub.ID
	}
	for _, n := range sub.ServiceTasks {
		pp.Nodes[n.ID] = n
		pp.ParentSubProcesses[n.ID] = sub.ID
	}
	for _, n := range sub.Tasks {
		pp.Nodes[n.ID] = n
		pp.ParentSubProcesses[n.ID] = sub.ID
	}
	for _, n := range sub.UserTasks {
		pp.Nodes[n.ID] = n
		pp.ParentSubProcesses[n.ID] = sub.ID
	}
	for _, n := range sub.ReceiveTasks {
		pp.Nodes[n.ID] = n
		pp.ParentSubProcesses[n.ID] = sub.ID
	}
	for _, n := range sub.IntermediateCatchEvents {
		pp.Nodes[n.ID] = n
		pp.ParentSubProcesses[n.ID] = sub.ID
	}
	for _, n := range sub.ExclusiveGateways {
		pp.Nodes[n.ID] = n
		pp.ParentSubProcesses[n.ID] = sub.ID
	}
	for _, n := range sub.ParallelGateways {
		pp.Nodes[n.ID] = n
		pp.ParentSubProcesses[n.ID] = sub.ID
	}
	for _, n := range sub.BusinessRuleTasks {
		pp.Nodes[n.ID] = n
		pp.ParentSubProcesses[n.ID] = sub.ID
	}
	for _, n := range sub.BoundaryEvents {
		pp.Nodes[n.ID] = n
		pp.ParentSubProcesses[n.ID] = sub.ID
		pp.BoundaryEventsByNode[n.AttachedToRef] = append(pp.BoundaryEventsByNode[n.AttachedToRef], n)
	}

	for _, inner := range sub.SubProcesses {
		indexSubProcess(&inner, pp, sub.ID)
	}

	for _, flow := range sub.SequenceFlows {
		pp.Outflows[flow.SourceRef] = append(pp.Outflows[flow.SourceRef], flow)
		pp.Inflows[flow.TargetRef] = append(pp.Inflows[flow.TargetRef], flow)
	}
}
