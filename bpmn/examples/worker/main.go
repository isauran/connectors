package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/nativebpm/connectors/bpmn"
)

func main() {
	fmt.Println("=== BPMN 2.0 Local Worker/Handler Example ===")

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

	// 3. Register worker handlers for topics defined in BPMN ServiceTasks
	engine.RegisterServiceTaskHandler("calculateInterest", func(ctx context.Context, instance *bpmn.ProcessInstance, task bpmn.ServiceTask) error {
		fmt.Printf("[WORKER] Executing 'calculateInterest' for task ID: %s...\n", task.ID)

		balance, ok := instance.Variables["balance"].(float64)
		if !ok {
			balance = 1000.0 // fallback default
		}

		interest := balance * 0.05
		instance.Variables["calculated_interest"] = interest
		fmt.Printf("[WORKER] Balance: %.2f, Calculated Interest (5%%): %.2f\n", balance, interest)
		return nil
	})

	engine.RegisterServiceTaskHandler("applyInterest", func(ctx context.Context, instance *bpmn.ProcessInstance, task bpmn.ServiceTask) error {
		fmt.Printf("[WORKER] Executing 'applyInterest' for task ID: %s...\n", task.ID)

		balance, ok := instance.Variables["balance"].(float64)
		if !ok {
			balance = 1000.0
		}
		interest, _ := instance.Variables["calculated_interest"].(float64)

		newBalance := balance + interest
		instance.Variables["balance"] = newBalance
		fmt.Printf("[WORKER] Applied interest. New Balance: %.2f\n", newBalance)
		return nil
	})

	// 4. Start process instance with initial variables
	instance, err := engine.StartInstance("worker-demo-instance", map[string]interface{}{
		"balance": 5000.0,
	})
	if err != nil {
		log.Fatalf("failed to start instance: %v", err)
	}

	fmt.Printf("Started process '%s' with ID '%s'\n", pp.ID, instance.ID)
	fmt.Printf("Initial variables: %+v\n", instance.Variables)
	fmt.Printf("Active tokens: %v\n", instance.ActiveTokens)

	// Step 1: Start -> calculate_interest
	fmt.Println("\n--- Step 1: Start Event ---")
	err = engine.Step(context.Background(), instance)
	if err != nil {
		log.Fatalf("step failed: %v", err)
	}
	fmt.Printf("Active tokens: %v\n", instance.ActiveTokens)

	// Step 2: Executes 'calculateInterest' worker handler -> moves to apply_interest
	fmt.Println("\n--- Step 2: Executing 'Calculate Interest' Service Task ---")
	err = engine.Step(context.Background(), instance)
	if err != nil {
		log.Fatalf("step failed: %v", err)
	}
	fmt.Printf("Active tokens: %v\n", instance.ActiveTokens)
	fmt.Printf("Variables: %+v\n", instance.Variables)

	// Step 3: Executes 'applyInterest' worker handler -> moves to end
	fmt.Println("\n--- Step 3: Executing 'Apply Interest' Service Task ---")
	err = engine.Step(context.Background(), instance)
	if err != nil {
		log.Fatalf("step failed: %v", err)
	}
	fmt.Printf("Active tokens: %v\n", instance.ActiveTokens)
	fmt.Printf("Variables: %+v\n", instance.Variables)

	// Step 4: end -> complete
	fmt.Println("\n--- Step 4: End Event ---")
	err = engine.Step(context.Background(), instance)
	if err != nil {
		log.Fatalf("step failed: %v", err)
	}
	fmt.Printf("Process completed: %t\n", instance.Completed)
}
