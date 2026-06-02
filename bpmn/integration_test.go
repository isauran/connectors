package bpmn

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBPMNDMNIntegration(t *testing.T) {
	// 1. Read files
	bpmnData, err := os.ReadFile("../camunda/examples/bpmn-spec/bpmn/dmn_business_rule.bpmn")
	require.NoError(t, err)

	dmnData, err := os.ReadFile("../camunda/examples/bpmn-spec/bpmn/decision.dmn")
	require.NoError(t, err)

	// 2. Parse DMN and BPMN
	dmn, err := ParseDMN(dmnData)
	require.NoError(t, err)

	pp, err := ParseBPMN(bpmnData)
	require.NoError(t, err)
	assert.Equal(t, "dmn_process", pp.ID)

	// 3. Register DMN inside the BPMN Engine
	engine := NewEngine(pp, nil)
	engine.RegisterDMN("determine_discount", dmn)

	// 4. Start instance with membership = gold
	instance, err := engine.StartInstance("inst-dmn", map[string]interface{}{
		"membership": "gold",
	})
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "StartEvent_1")

	// Step 1: StartEvent_1 -> Activity_DMN_Discount (BusinessRuleTask)
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "Activity_DMN_Discount")

	// Step 2: Activity_DMN_Discount -> Evaluate DMN -> writes to variable 'discount' -> Activity_Apply_Discount
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "Activity_Apply_Discount")

	// Check if DMN output is correctly populated in process context variables
	// Since gold membership grants 20% discount
	assert.Equal(t, 20.0, instance.Variables["discount"])

	// Step 3: Activity_Apply_Discount -> EndEvent_1
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.Contains(t, instance.ActiveTokens, "EndEvent_1")

	// Step 4: EndEvent_1 -> Complete
	err = engine.Step(context.Background(), instance)
	require.NoError(t, err)
	assert.True(t, instance.Completed)
}

func TestBPMNEventsBoundary(t *testing.T) {
	bpmnData, err := os.ReadFile("../camunda/examples/bpmn-spec/bpmn/events.bpmn")
	require.NoError(t, err)

	pp, err := ParseBPMN(bpmnData)
	require.NoError(t, err)
	assert.Equal(t, "events_process", pp.ID)

	engine := NewEngine(pp, nil)

	// SCENARIO 1: Message boundary interruption (cancelOrder)
	t.Run("Boundary Message Interruption", func(t *testing.T) {
		instance, err := engine.StartInstance("inst-msg", nil)
		require.NoError(t, err)

		// 1. Step: StartEvent_1 -> Activity_Process_Order (UserTask - Wait State)
		err = engine.Step(context.Background(), instance)
		require.NoError(t, err)
		assert.Contains(t, instance.ActiveTokens, "Activity_Process_Order")

		err = engine.Step(context.Background(), instance)
		require.NoError(t, err)
		assert.Empty(t, instance.ActiveTokens)
		assert.Contains(t, instance.WaitingTokens, "Activity_Process_Order")

		// 2. Correlate cancelOrder message
		err = engine.CorrelateMessage(instance, "cancelOrder", nil)
		require.NoError(t, err)
		assert.Empty(t, instance.WaitingTokens)
		assert.Contains(t, instance.ActiveTokens, "Activity_Cancel_Cleanup")

		// 3. Step: Activity_Cancel_Cleanup -> Event_Cancelled_End
		err = engine.Step(context.Background(), instance)
		require.NoError(t, err)
		assert.Contains(t, instance.ActiveTokens, "Event_Cancelled_End")

		// 4. Step: Event_Cancelled_End -> Complete
		err = engine.Step(context.Background(), instance)
		require.NoError(t, err)
		assert.True(t, instance.Completed)
	})

	// SCENARIO 2: Boundary Error Interruption (PAYMENT_FAILED)
	t.Run("Boundary Error Interruption", func(t *testing.T) {
		instance, err := engine.StartInstance("inst-error", nil)
		require.NoError(t, err)

		// 1. Step: StartEvent_1 -> Activity_Process_Order
		err = engine.Step(context.Background(), instance)
		require.NoError(t, err)

		// 2. Step: Process Activity_Process_Order -> WaitingTokens
		err = engine.Step(context.Background(), instance)
		require.NoError(t, err)

		// 3. Complete user task
		err = engine.CompleteTask(instance, "Activity_Process_Order", nil)
		require.NoError(t, err)
		assert.Contains(t, instance.ActiveTokens, "Activity_Process_Payment")

		// 4. Trigger payment error boundary event
		err = engine.HandleError(instance, "Activity_Process_Payment", "PAYMENT_FAILED", nil)
		require.NoError(t, err)
		assert.Contains(t, instance.ActiveTokens, "Activity_Payment_Failed_Handle")

		// 5. Step: Activity_Payment_Failed_Handle -> Event_Payment_Failed_End
		err = engine.Step(context.Background(), instance)
		require.NoError(t, err)
		assert.Contains(t, instance.ActiveTokens, "Event_Payment_Failed_End")

		// 6. Step: Event_Payment_Failed_End -> Complete
		err = engine.Step(context.Background(), instance)
		require.NoError(t, err)
		assert.True(t, instance.Completed)
	})
}

func TestBPMNSubprocesses(t *testing.T) {
	bpmnData, err := os.ReadFile("../camunda/examples/bpmn-spec/bpmn/subprocesses.bpmn")
	require.NoError(t, err)

	pp, err := ParseBPMN(bpmnData)
	require.NoError(t, err)
	assert.Equal(t, "subprocesses_process", pp.ID)

	engine := NewEngine(pp, nil)

	// SCENARIO 1: Success Subprocess Execution path
	t.Run("Success Subprocess Path", func(t *testing.T) {
		instance, err := engine.StartInstance("inst-sub-success", nil)
		require.NoError(t, err)

		// 1. StartEvent_1 -> Activity_Before_Sub (Initialize Order)
		err = engine.Step(context.Background(), instance)
		require.NoError(t, err)
		assert.Contains(t, instance.ActiveTokens, "Activity_Before_Sub")

		// 2. Activity_Before_Sub -> SubProcess_Embedded
		err = engine.Step(context.Background(), instance)
		require.NoError(t, err)
		assert.Contains(t, instance.ActiveTokens, "SubProcess_Embedded")

		// 3. Enter SubProcess -> SubStart (inner start event)
		err = engine.Step(context.Background(), instance)
		require.NoError(t, err)
		assert.Contains(t, instance.ActiveTokens, "SubStart")

		// 4. SubStart -> Activity_Sub_Pack (Pack Items)
		err = engine.Step(context.Background(), instance)
		require.NoError(t, err)
		assert.Contains(t, instance.ActiveTokens, "Activity_Sub_Pack")

		// 5. Activity_Sub_Pack -> Activity_Sub_Ship (Ship Items)
		err = engine.Step(context.Background(), instance)
		require.NoError(t, err)
		assert.Contains(t, instance.ActiveTokens, "Activity_Sub_Ship")

		// 6. Activity_Sub_Ship -> SubEnd
		err = engine.Step(context.Background(), instance)
		require.NoError(t, err)
		assert.Contains(t, instance.ActiveTokens, "SubEnd")

		// 7. SubEnd -> returns to parent flow (EndEvent_Success)
		err = engine.Step(context.Background(), instance)
		require.NoError(t, err)
		assert.Contains(t, instance.ActiveTokens, "EndEvent_Success")

		// 8. EndEvent_Success -> Complete
		err = engine.Step(context.Background(), instance)
		require.NoError(t, err)
		assert.True(t, instance.Completed)
	})

	// SCENARIO 2: Boundary Error Interruption on Subprocess level
	t.Run("Boundary Error on Subprocess", func(t *testing.T) {
		instance, err := engine.StartInstance("inst-sub-error", nil)
		require.NoError(t, err)

		// Steps up to Ship Items
		_ = engine.Step(context.Background(), instance) // StartEvent_1 -> Activity_Before_Sub
		_ = engine.Step(context.Background(), instance) // Activity_Before_Sub -> SubProcess_Embedded
		_ = engine.Step(context.Background(), instance) // SubProcess_Embedded -> SubStart
		_ = engine.Step(context.Background(), instance) // SubStart -> Activity_Sub_Pack
		_ = engine.Step(context.Background(), instance) // Activity_Sub_Pack -> Activity_Sub_Ship
		assert.Contains(t, instance.ActiveTokens, "Activity_Sub_Ship")

		// Trigger shipping failure error inside the subprocess.
		// Parent subprocess SubProcess_Embedded boundary error handler should catch it.
		err = engine.HandleError(instance, "Activity_Sub_Ship", "SHIPPING_FAILED", nil)
		require.NoError(t, err)
		assert.Contains(t, instance.ActiveTokens, "Activity_Refund_Customer")

		// Step: Activity_Refund_Customer -> Event_Sub_Error_End
		err = engine.Step(context.Background(), instance)
		require.NoError(t, err)
		assert.Contains(t, instance.ActiveTokens, "Event_Sub_Error_End")

		// Step: Event_Sub_Error_End -> Complete
		err = engine.Step(context.Background(), instance)
		require.NoError(t, err)
		assert.True(t, instance.Completed)
	})

	// SCENARIO 3: Event Subprocess Interruption (cancelFulfillment message)
	t.Run("Event Subprocess Interruption", func(t *testing.T) {
		instance, err := engine.StartInstance("inst-sub-event", nil)
		require.NoError(t, err)

		// Steps up to Pack Items
		_ = engine.Step(context.Background(), instance) // StartEvent_1 -> Activity_Before_Sub
		_ = engine.Step(context.Background(), instance) // Activity_Before_Sub -> SubProcess_Embedded
		_ = engine.Step(context.Background(), instance) // SubProcess_Embedded -> SubStart
		_ = engine.Step(context.Background(), instance) // SubStart -> Activity_Sub_Pack
		assert.Contains(t, instance.ActiveTokens, "Activity_Sub_Pack")

		// Trigger event subprocess cancelFulfillment message
		err = engine.CorrelateMessage(instance, "cancelFulfillment", nil)
		require.NoError(t, err)
		// ActiveTokens should be cleared and set to EventSubStart
		assert.Equal(t, []string{"EventSubStart"}, instance.ActiveTokens)

		// Step: EventSubStart -> Activity_Cancel_Release
		err = engine.Step(context.Background(), instance)
		require.NoError(t, err)
		assert.Contains(t, instance.ActiveTokens, "Activity_Cancel_Release")

		// Step: Activity_Cancel_Release -> EventSubEnd
		err = engine.Step(context.Background(), instance)
		require.NoError(t, err)
		assert.Contains(t, instance.ActiveTokens, "EventSubEnd")

		// Step: EventSubEnd -> parent flow check (ActiveTokens becomes empty)
		err = engine.Step(context.Background(), instance)
		require.NoError(t, err)
		assert.Empty(t, instance.ActiveTokens)
		assert.False(t, instance.Completed)

		// Step: check completion -> Complete
		err = engine.Step(context.Background(), instance)
		require.NoError(t, err)
		assert.True(t, instance.Completed)
	})
}
