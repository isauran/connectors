package bpmn

import (
	"encoding/xml"
	"fmt"
)

// Definitions represents the root element of a BPMN XML document.
type Definitions struct {
	XMLName   xml.Name  `xml:"definitions"`
	Processes []Process `xml:"process"`
}

// Process represents a BPMN process definition.
type Process struct {
	ID                string             `xml:"id,attr"`
	Name              string             `xml:"name,attr"`
	IsExecutable      bool               `xml:"isExecutable,attr"`
	StartEvents       []StartEvent       `xml:"startEvent"`
	EndEvents         []EndEvent         `xml:"endEvent"`
	ServiceTasks      []ServiceTask      `xml:"serviceTask"`
	SequenceFlows     []SequenceFlow     `xml:"sequenceFlow"`
	ExclusiveGateways []ExclusiveGateway `xml:"exclusiveGateway"`
	ParallelGateways  []ParallelGateway  `xml:"parallelGateway"`
}

// StartEvent represents a BPMN start event.
type StartEvent struct {
	ID   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
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
	ID           string
	Name         string
	StartNodeID  string
	Nodes        map[string]interface{}
	Outflows     map[string][]SequenceFlow
	Inflows      map[string][]SequenceFlow
}

// ParseBPMN parses raw BPMN XML data and indexes the first executable process found.
func ParseBPMN(xmlData []byte) (*ParsedProcess, error) {
	var defs Definitions
	err := xml.Unmarshal(xmlData, &defs)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal BPMN XML: %w", err)
	}

	for _, p := range defs.Processes {
		if p.IsExecutable {
			return indexProcess(&p)
		}
	}

	// fallback to first process if none are marked executable
	if len(defs.Processes) > 0 {
		return indexProcess(&defs.Processes[0])
	}

	return nil, fmt.Errorf("no processes found in BPMN document")
}

func indexProcess(p *Process) (*ParsedProcess, error) {
	pp := &ParsedProcess{
		ID:       p.ID,
		Name:     p.Name,
		Nodes:    make(map[string]interface{}),
		Outflows: make(map[string][]SequenceFlow),
		Inflows:  make(map[string][]SequenceFlow),
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
	for _, n := range p.ExclusiveGateways {
		pp.Nodes[n.ID] = n
	}
	for _, n := range p.ParallelGateways {
		pp.Nodes[n.ID] = n
	}

	// 2. Index outflows and inflows
	for _, flow := range p.SequenceFlows {
		pp.Outflows[flow.SourceRef] = append(pp.Outflows[flow.SourceRef], flow)
		pp.Inflows[flow.TargetRef] = append(pp.Inflows[flow.TargetRef], flow)
	}

	return pp, nil
}
