package bpmn

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 1. Test XML for Inclusive Gateway (OR Split/Join)
const inclusiveBPMN = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://www.omg.org/spec/BPMN/20100524/MODEL" targetNamespace="http://bpmn.io/schema/bpmn">
  <process id="inclusive_process" isExecutable="true">
    <startEvent id="start" />
    <sequenceFlow id="f_start" sourceRef="start" targetRef="split" />
    <inclusiveGateway id="split" name="Split" />
    
    <sequenceFlow id="flow1" sourceRef="split" targetRef="task1">
      <conditionExpression>goA == true</conditionExpression>
    </sequenceFlow>
    <serviceTask id="task1" name="Task A" />
    <sequenceFlow id="f_t1" sourceRef="task1" targetRef="join" />

    <sequenceFlow id="flow2" sourceRef="split" targetRef="task2">
      <conditionExpression>goB == true</conditionExpression>
    </sequenceFlow>
    <serviceTask id="task2" name="Task B" />
    <sequenceFlow id="f_t2" sourceRef="task2" targetRef="join" />

    <sequenceFlow id="flow3" sourceRef="split" targetRef="task3">
      <conditionExpression>goD == true</conditionExpression>
    </sequenceFlow>
    <serviceTask id="task3" name="Task D" />
    <sequenceFlow id="f_t3" sourceRef="task3" targetRef="join" />

    <inclusiveGateway id="join" name="Join" />
    <sequenceFlow id="f_end" sourceRef="join" targetRef="end" />
    <endEvent id="end" />
  </process>
</definitions>`

func TestBPMNInclusiveGateway(t *testing.T) {
	pp, err := ParseBPMN([]byte(inclusiveBPMN))
	require.NoError(t, err)

	// Scenario 1: goA == true, goB == true -> Both task1 and task2 should trigger, wait at join, and complete
	engine := NewEngine(pp, nil)
	instance, err := engine.StartInstance("inst-inc-c", map[string]interface{}{
		"goA": true,
		"goB": true,
		"goD": false,
	})
	require.NoError(t, err)

	// start -> split
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "split")

	// split -> task1, task2 (both goA == true and goB == true are true)
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "task1")
	assert.Contains(t, instance.ActiveTokens, "task2")
	assert.Equal(t, 2, len(instance.ActiveTokens))

	// step task1 -> moves to join, ActiveTokens: [task2, join]
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "task2")
	assert.Contains(t, instance.ActiveTokens, "join")

	// step task2 -> moves to join, ActiveTokens: [join, join]
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "join")
	assert.Equal(t, 2, len(instance.ActiveTokens))

	// step join -> satisfies because no other tokens can reach it. Moves to end.
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "end")

	// step end -> complete
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.True(t, instance.Completed)

	// Scenario 2: goA == true, goB == false -> Only task1 triggers. Join immediately satisfies.
	instance2, err := engine.StartInstance("inst-inc-a", map[string]interface{}{
		"goA": true,
		"goB": false,
		"goD": false,
	})
	require.NoError(t, err)

	_ = engine.Step(context.Background(), instance2) // start -> split
	_ = engine.Step(context.Background(), instance2) // split -> task1 (only)

	assert.Contains(t, instance2.ActiveTokens, "task1")
	assert.Equal(t, 1, len(instance2.ActiveTokens))

	_ = engine.Step(context.Background(), instance2) // task1 -> join. ActiveTokens: [join]
	assert.Contains(t, instance2.ActiveTokens, "join")

	_ = engine.Step(context.Background(), instance2) // join -> end. Immediately proceeds because task2 cannot reach.
	assert.Contains(t, instance2.ActiveTokens, "end")

	_ = engine.Step(context.Background(), instance2) // end -> complete
	assert.True(t, instance2.Completed)
}

// 2. Test XML for Non-interrupting boundary events
const nonInterruptingBPMN = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://www.omg.org/spec/BPMN/20100524/MODEL" targetNamespace="http://bpmn.io/schema/bpmn">
  <message id="MsgRef" name="nonInterruptingMsg" />
  <process id="non_interrupting_process" isExecutable="true">
    <startEvent id="start" />
    <sequenceFlow id="f1" sourceRef="start" targetRef="user_task" />
    <userTask id="user_task" name="Main Task" />
    <sequenceFlow id="f2" sourceRef="user_task" targetRef="end_main" />
    <endEvent id="end_main" />

    <boundaryEvent id="boundary_event" attachedToRef="user_task" cancelActivity="false">
      <messageEventDefinition messageRef="MsgRef" />
    </boundaryEvent>
    <sequenceFlow id="f3" sourceRef="boundary_event" targetRef="boundary_task" />
    <serviceTask id="boundary_task" name="Boundary Handler" />
    <sequenceFlow id="f4" sourceRef="boundary_task" targetRef="end_boundary" />
    <endEvent id="end_boundary" />
  </process>
</definitions>`

func TestBPMNNonInterruptingBoundary(t *testing.T) {
	pp, err := ParseBPMN([]byte(nonInterruptingBPMN))
	require.NoError(t, err)

	engine := NewEngine(pp, nil)
	instance, err := engine.StartInstance("inst-non-int", nil)
	require.NoError(t, err)

	// start -> user_task
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)

	// process user_task -> park at WaitingTokens
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.Contains(t, instance.WaitingTokens, "user_task")
	assert.Empty(t, instance.ActiveTokens)

	// Correlate message "nonInterruptingMsg"
	err = engine.CorrelateMessage(instance, "nonInterruptingMsg", nil)
	require.NoError(t, err)

	// Since cancelActivity="false", user_task remains in WaitingTokens,
	// and boundary_task's transition (boundary_event -> boundary_task) is added to ActiveTokens
	assert.Contains(t, instance.WaitingTokens, "user_task")
	assert.Contains(t, instance.ActiveTokens, "boundary_task")

	// step boundary_task -> end_boundary
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "end_boundary")

	// step end_boundary -> finishes boundary flow, user_task is still waiting
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.Empty(t, instance.ActiveTokens)
	assert.Contains(t, instance.WaitingTokens, "user_task")
	assert.False(t, instance.Completed)

	// Now complete the main user_task
	err = engine.CompleteTask(instance, "user_task", nil)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "end_main")

	// step end_main -> complete
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.True(t, instance.Completed)
}

// 3. Test XML for Broadcast Signal
const signalBPMN = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://www.omg.org/spec/BPMN/20100524/MODEL" targetNamespace="http://bpmn.io/schema/bpmn">
  <signal id="SigRef" name="broadcastAlert" />
  <process id="signal_process" isExecutable="true">
    <startEvent id="start" />
    <sequenceFlow id="f_start" sourceRef="start" targetRef="fork" />
    <parallelGateway id="fork" name="Fork" />

    <sequenceFlow id="f_t1" sourceRef="fork" targetRef="user_task1" />
    <userTask id="user_task1" name="User Task 1" />
    <sequenceFlow id="f_e1" sourceRef="user_task1" targetRef="end1" />
    <endEvent id="end1" />

    <boundaryEvent id="boundary_sig" attachedToRef="user_task1" cancelActivity="true">
      <signalEventDefinition signalRef="SigRef" />
    </boundaryEvent>
    <sequenceFlow id="f_sb" sourceRef="boundary_sig" targetRef="handler_task" />
    <serviceTask id="handler_task" name="Signal Boundary Handler" />
    <sequenceFlow id="f_eb" sourceRef="handler_task" targetRef="end_b" />
    <endEvent id="end_b" />

    <sequenceFlow id="f_t2" sourceRef="fork" targetRef="catch_sig" />
    <intermediateCatchEvent id="catch_sig" name="Wait for Signal">
      <signalEventDefinition signalRef="SigRef" />
    </intermediateCatchEvent>
    <sequenceFlow id="f_e2" sourceRef="catch_sig" targetRef="end2" />
    <endEvent id="end2" />
  </process>
</definitions>`

func TestBPMNSignals(t *testing.T) {
	pp, err := ParseBPMN([]byte(signalBPMN))
	require.NoError(t, err)

	engine := NewEngine(pp, nil)
	instance, err := engine.StartInstance("inst-sig", nil)
	require.NoError(t, err)

	// start -> fork
	_ = engine.Step(context.Background(), instance)
	// fork -> [user_task1, catch_sig]
	_ = engine.Step(context.Background(), instance)
	// step user_task1 -> WaitingTokens
	_ = engine.Step(context.Background(), instance)
	// step catch_sig -> WaitingTokens
	_ = engine.Step(context.Background(), instance)

	assert.Contains(t, instance.WaitingTokens, "user_task1")
	assert.Contains(t, instance.WaitingTokens, "catch_sig")

	// Broadcast signal "broadcastAlert"
	err = engine.BroadcastSignal(instance, "broadcastAlert", nil)
	require.NoError(t, err)

	// Both should wake up!
	// 1. user_task1 signal boundary event is interrupting, so user_task1 is removed, and handler_task is activated.
	// 2. catch_sig intermediate catch event is activated.
	assert.NotContains(t, instance.WaitingTokens, "user_task1")
	assert.NotContains(t, instance.WaitingTokens, "catch_sig")
	assert.Contains(t, instance.ActiveTokens, "handler_task")
	assert.Contains(t, instance.ActiveTokens, "end2") // catch_sig points directly to end2
}

// 4. Test XML for Compensations (Saga Pattern)
const compensationBPMN = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://www.omg.org/spec/BPMN/20100524/MODEL" targetNamespace="http://bpmn.io/schema/bpmn">
  <process id="compensation_process" isExecutable="true">
    <startEvent id="start" />
    <sequenceFlow id="f1" sourceRef="start" targetRef="task_main" />
    <serviceTask id="task_main" name="Reserve Hotel" />
    <sequenceFlow id="f2" sourceRef="task_main" targetRef="checkpoint" />
    <userTask id="checkpoint" name="Checkpoint" />
    <sequenceFlow id="f3" sourceRef="checkpoint" targetRef="end" />
    <endEvent id="end" />

    <boundaryEvent id="comp_boundary" attachedToRef="task_main" />
    <association id="assoc1" sourceRef="comp_boundary" targetRef="task_compensate" />
    <serviceTask id="task_compensate" name="Cancel Hotel Reservation" />
  </process>
</definitions>`

func TestBPMNCompensations(t *testing.T) {
	pp, err := ParseBPMN([]byte(compensationBPMN))
	require.NoError(t, err)

	engine := NewEngine(pp, nil)
	instance, err := engine.StartInstance("inst-comp", nil)
	require.NoError(t, err)

	// start -> task_main
	_ = engine.Step(context.Background(), instance)
	// task_main (ServiceTask) executes automatically, records to CompletedTasks, and moves to checkpoint
	_ = engine.Step(context.Background(), instance)

	assert.Contains(t, instance.CompletedTasks, "task_main")
	assert.Contains(t, instance.ActiveTokens, "checkpoint")

	// checkpoint -> WaitingTokens
	_ = engine.Step(context.Background(), instance)
	assert.Contains(t, instance.WaitingTokens, "checkpoint")

	// Trigger compensation for task_main
	err = engine.TriggerCompensation(instance, "task_main")
	require.NoError(t, err)

	// task_compensate should be activated
	assert.Contains(t, instance.ActiveTokens, "task_compensate")

	// step task_compensate -> executes as noop and token disappears (no outflows)
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.Empty(t, instance.ActiveTokens)

	// Now complete checkpoint
	err = engine.CompleteTask(instance, "checkpoint", nil)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "end")

	// end -> complete
	_ = engine.Step(context.Background(), instance)
	assert.True(t, instance.Completed)
}

// 5. Test XML for Multi-Instance Tasks
const multiInstanceBPMN = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://www.omg.org/spec/BPMN/20100524/MODEL" targetNamespace="http://bpmn.io/schema/bpmn">
  <process id="mi_process" isExecutable="true">
    <startEvent id="start" />
    <sequenceFlow id="f1" sourceRef="start" targetRef="mi_task" />
    <userTask id="mi_task" name="Multi-Instance Approval">
      <multiInstanceLoopCharacteristics>
        <loopDataInputRef>items</loopDataInputRef>
      </multiInstanceLoopCharacteristics>
    </userTask>
    <sequenceFlow id="f2" sourceRef="mi_task" targetRef="end" />
    <endEvent id="end" />
  </process>
</definitions>`

func TestBPMNMultiInstance(t *testing.T) {
	pp, err := ParseBPMN([]byte(multiInstanceBPMN))
	require.NoError(t, err)

	engine := NewEngine(pp, nil)
	// We pass a collection "items" with 3 elements
	instance, err := engine.StartInstance("inst-mi", map[string]interface{}{
		"items": []string{"item1", "item2", "item3"},
	})
	require.NoError(t, err)

	// start -> mi_task
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "mi_task")

	// Step mi_task -> should initialize and split into 3 copies: mi_task#0, mi_task#1, mi_task#2
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "mi_task#0")
	assert.Contains(t, instance.ActiveTokens, "mi_task#1")
	assert.Contains(t, instance.ActiveTokens, "mi_task#2")
	assert.NotContains(t, instance.ActiveTokens, "mi_task") // base token is gone
	assert.Equal(t, 3, len(instance.ActiveTokens))

	// Step mi_task#0 -> moves it to WaitingTokens
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	// Step mi_task#1 -> moves it to WaitingTokens
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	// Step mi_task#2 -> moves it to WaitingTokens
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)

	assert.Contains(t, instance.WaitingTokens, "mi_task#0")
	assert.Contains(t, instance.WaitingTokens, "mi_task#1")
	assert.Contains(t, instance.WaitingTokens, "mi_task#2")
	assert.Empty(t, instance.ActiveTokens)

	// Complete copy #0 -> redirects to mi_task#join. ActiveTokens: [mi_task#join]
	err = engine.CompleteTask(instance, "mi_task#0", nil)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "mi_task#join")

	// Step mi_task#join -> not all 3 arrived yet, so it reparks at the end of ActiveTokens
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "mi_task#join")

	// Complete copy #1 -> redirects to mi_task#join. ActiveTokens: [mi_task#join, mi_task#join]
	err = engine.CompleteTask(instance, "mi_task#1", nil)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "mi_task#join")
	assert.Equal(t, 2, len(instance.ActiveTokens))

	// Complete copy #2 -> redirects to mi_task#join. ActiveTokens: [mi_task#join, mi_task#join, mi_task#join]
	err = engine.CompleteTask(instance, "mi_task#2", nil)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "mi_task#join")
	assert.Equal(t, 3, len(instance.ActiveTokens))

	// Step mi_task#join -> now 3 copies have arrived. Satisfies, cleans up join tokens, and moves to end!
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "end")

	// end -> complete
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.True(t, instance.Completed)
}
