package bpmn

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testBPMN = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://www.omg.org/spec/BPMN/20100524/MODEL" targetNamespace="http://bpmn.io/schema/bpmn">
  <process id="test_process" isExecutable="true">
    <startEvent id="start" />
    <sequenceFlow id="flow1" sourceRef="start" targetRef="task1" />
    <serviceTask id="task1" name="Calculate Something" />
    <sequenceFlow id="flow2" sourceRef="task1" targetRef="gateway" />
    <exclusiveGateway id="gateway" name="Is Valid?" />
    <sequenceFlow id="flow3" sourceRef="gateway" targetRef="end_ok">
      <conditionExpression>valid == true</conditionExpression>
    </sequenceFlow>
    <sequenceFlow id="flow4" sourceRef="gateway" targetRef="end_fail">
      <conditionExpression>valid == false</conditionExpression>
    </sequenceFlow>
    <endEvent id="end_ok" />
    <endEvent id="end_fail" />
  </process>
</definitions>`

const testDMN = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" id="definitions" name="definitions" namespace="http://camunda.org/schema/1.0/dmn">
  <decision id="approve_loan" name="Approve Loan">
    <decisionTable id="decisionTable" hitPolicy="UNIQUE">
      <input id="input1" label="Age">
        <inputExpression id="inputExpression1" typeRef="integer">
          <text>age</text>
        </inputExpression>
      </input>
      <output id="output1" label="Approved" name="approved" typeRef="boolean" />
      <rule id="rule1">
        <inputEntry id="inputEntry1">
          <text>&lt;18</text>
        </inputEntry>
        <outputEntry id="outputEntry1">
          <text>false</text>
        </outputEntry>
      </rule>
      <rule id="rule2">
        <inputEntry id="inputEntry2">
          <text>&gt;17</text>
        </inputEntry>
        <outputEntry id="outputEntry2">
          <text>true</text>
        </outputEntry>
      </rule>
    </decisionTable>
  </decision>
</definitions>`

func TestBPMNParserAndEngine(t *testing.T) {
	// 1. Test Parser
	pp, err := ParseBPMN([]byte(testBPMN))
	require.NoError(t, err)
	assert.Equal(t, "test_process", pp.ID)
	assert.Equal(t, "start", pp.StartNodeID)
	assert.Contains(t, pp.Nodes, "task1")
	assert.Contains(t, pp.Nodes, "gateway")

	// 2. Test Engine Execution (True branch)
	engine := NewEngine(pp, nil) // no WASM engine for unit test
	instance, err := engine.StartInstance("instance-1", map[string]interface{}{
		"valid": true,
	})
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "start")

	// Step 1: StartEvent -> Task1
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "task1")

	// Step 2: ServiceTask -> ExclusiveGateway
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "gateway")

	// Step 3: XOR Gateway -> end_ok (since valid is true)
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "end_ok")

	// Step 4: EndEvent reach -> complete
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.True(t, instance.Completed)

	// 3. Test Engine Execution (False branch)
	instance2, err := engine.StartInstance("instance-2", map[string]interface{}{
		"valid": false,
	})
	require.NoError(t, err)

	// Quick steps to gateway
	_ = engine.Step(context.Background(), instance2) // start -> task1
	_ = engine.Step(context.Background(), instance2) // task1 -> gateway

	// XOR Gateway -> end_fail (since valid is false)
	err = engine.Step(context.Background(), instance2)
	require.NoError(t, err)
	assert.Contains(t, instance2.ActiveTokens, "end_fail")
}

func TestDMNEvaluator(t *testing.T) {
	dmn, err := ParseDMN([]byte(testDMN))
	require.NoError(t, err)

	// Test case: age < 18 -> approved = false
	res, err := Evaluate(dmn, "approve_loan", map[string]interface{}{
		"age": 16,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, false, res["approved"])

	// Test case: age > 17 -> approved = true
	res2, err := Evaluate(dmn, "approve_loan", map[string]interface{}{
		"age": 20,
	})
	require.NoError(t, err)
	require.NotNil(t, res2)
	assert.Equal(t, true, res2["approved"])
}
