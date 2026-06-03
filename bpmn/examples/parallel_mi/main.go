package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/nativebpm/connectors/bpmn"
)

func main() {
	fmt.Println("=== BPMN 2.0 Multi-Instance & Inclusive Gateway Example ===")

	// 1. Read and parse BPMN schema
	xmlData, err := os.ReadFile("process.bpmn")
	if err != nil {
		log.Fatalf("failed to read process.bpmn: %v", err)
	}

	pp, err := bpmn.ParseBPMN(xmlData)
	if err != nil {
		log.Fatalf("failed to parse BPMN: %v", err)
	}

	// 2. Initialize process engine
	engine := bpmn.NewEngine(pp, nil)

	// Start process instance with variables:
	// - "items": collection for Multi-Instance (2 items)
	// - "isUrgent": true (triggers urgent path)
	// - "isStandard": true (triggers standard path)
	instance, err := engine.StartInstance("mi-instance-1", map[string]interface{}{
		"items":      []string{"iPhone 15", "MacBook Pro"},
		"isUrgent":   true,
		"isStandard": true,
	})
	if err != nil {
		log.Fatalf("failed to start instance: %v", err)
	}

	fmt.Printf("Started process '%s' with ID '%s'\n", pp.ID, instance.ID)
	fmt.Printf("Active tokens: %v\n\n", instance.ActiveTokens)

	// Step 1: Start -> mi_task
	fmt.Println("--- Step 1: Start Event ---")
	_ = engine.Step(context.Background(), instance)
	fmt.Printf("Active tokens: %v\n\n", instance.ActiveTokens)

	// Step 2: mi_task entry -> splits into 2 parallel copies (mi_task#0, mi_task#1)
	fmt.Println("--- Step 2: Initializing Multi-Instance Task ---")
	_ = engine.Step(context.Background(), instance)
	fmt.Printf("Active tokens: %v\n\n", instance.ActiveTokens)

	// Step 3: Move all active tokens to WaitingTokens (since UserTask is a wait state)
	fmt.Println("--- Step 3: Reaching Wait State for Multi-Instance tasks ---")
	_ = engine.Step(context.Background(), instance) // parks mi_task#0
	_ = engine.Step(context.Background(), instance) // parks mi_task#1
	fmt.Printf("Active tokens: %v\n", instance.ActiveTokens)
	fmt.Printf("Waiting tokens: %v\n\n", instance.WaitingTokens)

	// Step 4: Complete copy #0
	fmt.Println("--- Step 4: Completing 'Approve Item' #0 (iPhone 15) ---")
	err = engine.CompleteTask(instance, "mi_task#0", nil)
	if err != nil {
		log.Fatalf("failed to complete task: %v", err)
	}
	fmt.Printf("Active tokens: %v\n", instance.ActiveTokens)
	fmt.Printf("Waiting tokens: %v\n\n", instance.WaitingTokens)

	// Try to execute the join token (reparks because copy #1 is still waiting)
	_ = engine.Step(context.Background(), instance)
	fmt.Println("Checked join condition. Not satisfied yet, reparked join token.")
	fmt.Printf("Active tokens: %v\n\n", instance.ActiveTokens)

	// Step 5: Complete copy #1
	fmt.Println("--- Step 5: Completing 'Approve Item' #1 (MacBook Pro) ---")
	err = engine.CompleteTask(instance, "mi_task#1", nil)
	if err != nil {
		log.Fatalf("failed to complete task: %v", err)
	}
	fmt.Printf("Active tokens: %v\n", instance.ActiveTokens)
	fmt.Printf("Waiting tokens: %v\n\n", instance.WaitingTokens)

	// Step 6: Step the join (now satisfied, moves token forward to 'split')
	fmt.Println("--- Step 6: Multi-Instance Join satisfied, moving to Inclusive Split ---")
	_ = engine.Step(context.Background(), instance)
	fmt.Printf("Active tokens: %v\n\n", instance.ActiveTokens)

	// Step 7: Step the split -> evaluates conditions and activates both task_urgent and task_standard
	fmt.Println("--- Step 7: Evaluating Inclusive Gateway (OR-Split) ---")
	_ = engine.Step(context.Background(), instance)
	fmt.Printf("Active tokens: %v\n\n", instance.ActiveTokens)

	// Step 8: Execute task_urgent (moves to join)
	fmt.Println("--- Step 8: Executing 'Process Urgent Order' ---")
	_ = engine.Step(context.Background(), instance)
	fmt.Printf("Active tokens: %v\n\n", instance.ActiveTokens)

	// Step 9: Execute task_standard (moves to join)
	fmt.Println("--- Step 9: Executing 'Process Standard Order' ---")
	_ = engine.Step(context.Background(), instance)
	fmt.Printf("Active tokens: %v\n\n", instance.ActiveTokens)

	// Step 10: Step the join gateway (satisfied, moves to end)
	fmt.Println("--- Step 10: Inclusive Join satisfied, moving to End Event ---")
	_ = engine.Step(context.Background(), instance)
	fmt.Printf("Active tokens: %v\n\n", instance.ActiveTokens)

	// Step 11: End Event reached -> complete
	fmt.Println("--- Step 11: Reaching End Event ---")
	_ = engine.Step(context.Background(), instance)
	fmt.Printf("Process completed: %t\n", instance.Completed)
}
