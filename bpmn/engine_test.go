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

const parallelBPMN = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://www.omg.org/spec/BPMN/20100524/MODEL" targetNamespace="http://bpmn.io/schema/bpmn">
  <process id="parallel_process" isExecutable="true">
    <startEvent id="start" />
    <sequenceFlow id="flow1" sourceRef="start" targetRef="fork" />
    <parallelGateway id="fork" name="Split" />
    
    <sequenceFlow id="flow_to_b1" sourceRef="fork" targetRef="task_b1" />
    <serviceTask id="task_b1" name="Branch 1 Task" />
    <sequenceFlow id="flow_to_join1" sourceRef="task_b1" targetRef="join" />
    
    <sequenceFlow id="flow_to_b2" sourceRef="fork" targetRef="task_b2" />
    <serviceTask id="task_b2" name="Branch 2 Task" />
    <sequenceFlow id="flow_to_join2" sourceRef="task_b2" targetRef="join" />
    
    <parallelGateway id="join" name="Merge" />
    <sequenceFlow id="flow_to_end" sourceRef="join" targetRef="end" />
    <endEvent id="end" />
  </process>
</definitions>`

func TestBPMNParallelGateway(t *testing.T) {
	pp, err := ParseBPMN([]byte(parallelBPMN))
	require.NoError(t, err)
	assert.Equal(t, "parallel_process", pp.ID)

	engine := NewEngine(pp, nil)
	instance, err := engine.StartInstance("instance-parallel", nil)
	require.NoError(t, err)

	// 1. start -> fork
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "fork")

	// 2. fork -> [task_b1, task_b2]
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "task_b1")
	assert.Contains(t, instance.ActiveTokens, "task_b2")

	// 3. Обрабатываем первый токен (task_b1) -> переходит на join.
	// Остается task_b2 и join.
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "task_b2")
	assert.Contains(t, instance.ActiveTokens, "join")

	// 4. Обрабатываем task_b2 -> переходит на join.
	// Теперь оба токена пришли на join. ActiveTokens: [join, join].
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "join")
	assert.Equal(t, 2, len(instance.ActiveTokens))

	// 5. Обрабатываем join -> переходит на end.
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "end")

	// 6. end -> complete
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.True(t, instance.Completed)
}

