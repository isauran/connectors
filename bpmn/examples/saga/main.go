package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/nativebpm/connectors/bpmn"
)

func main() {
	fmt.Println("=== BPMN 2.0 Saga Pattern (Compensation) Example ===")

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
	engine := bpmn.NewEngine(pp, nil) // no WASM support needed for local mock tasks
	
	// Start process instance
	instance, err := engine.StartInstance("saga-instance-1", map[string]interface{}{
		"booking_id": "BK-9921",
	})
	if err != nil {
		log.Fatalf("failed to start instance: %v", err)
	}

	fmt.Printf("Started process '%s' with ID '%s'\n", pp.ID, instance.ID)
	fmt.Printf("Active tokens: %v\n", instance.ActiveTokens)

	// Step 1: Start -> Charge Payment
	fmt.Println("\n--- Step 1: Start Event ---")
	err = engine.Step(context.Background(), instance)
	if err != nil {
		log.Fatalf("step failed: %v", err)
	}
	fmt.Printf("Active tokens: %v\n", instance.ActiveTokens)

	// Step 2: Charge Payment -> Reserve Hotel
	fmt.Println("\n--- Step 2: Executing 'Charge Payment' ---")
	err = engine.Step(context.Background(), instance)
	if err != nil {
		log.Fatalf("step failed: %v", err)
	}
	fmt.Printf("Completed tasks history: %v\n", instance.CompletedTasks)
	fmt.Printf("Active tokens: %v\n", instance.ActiveTokens)

	// Step 3: Reserve Hotel -> Checkpoint (Wait State)
	fmt.Println("\n--- Step 3: Executing 'Reserve Hotel' ---")
	err = engine.Step(context.Background(), instance)
	if err != nil {
		log.Fatalf("step failed: %v", err)
	}
	fmt.Printf("Completed tasks history: %v\n", instance.CompletedTasks)
	fmt.Printf("Active tokens: %v\n", instance.ActiveTokens)

	// Step 4: Checkpoint is UserTask, executing Step parks it to WaitingTokens
	fmt.Println("\n--- Step 4: Reaching Checkpoint (Wait State) ---")
	err = engine.Step(context.Background(), instance)
	if err != nil {
		log.Fatalf("step failed: %v", err)
	}
	fmt.Printf("Active tokens: %v\n", instance.ActiveTokens)
	fmt.Printf("Waiting tokens: %v\n", instance.WaitingTokens)

	// SIMULATING A FAILURE (e.g. Booking confirmation failed downstream)
	fmt.Println("\n--- CRITICAL: downstream flight reservation failed. Initiating Saga compensations... ---")
	
	// We compensate the completed activities in reverse order (hotel first, then payment)
	// Compensate Hotel
	fmt.Println("Triggering compensation for 'reserve_hotel'...")
	err = engine.TriggerCompensation(instance, "reserve_hotel")
	if err != nil {
		log.Fatalf("failed to trigger compensation for reserve_hotel: %v", err)
	}
	fmt.Printf("Active tokens: %v\n", instance.ActiveTokens)

	// Execute compensating task 'cancel_hotel'
	err = engine.Step(context.Background(), instance)
	if err != nil {
		log.Fatalf("failed to execute compensation task: %v", err)
	}
	fmt.Println("Compensation activity 'cancel_hotel' completed.")
	fmt.Printf("Active tokens: %v\n", instance.ActiveTokens)

	// Compensate Payment
	fmt.Println("\nTriggering compensation for 'charge_payment'...")
	err = engine.TriggerCompensation(instance, "charge_payment")
	if err != nil {
		log.Fatalf("failed to trigger compensation for charge_payment: %v", err)
	}
	fmt.Printf("Active tokens: %v\n", instance.ActiveTokens)

	// Execute compensating task 'refund_payment'
	err = engine.Step(context.Background(), instance)
	if err != nil {
		log.Fatalf("failed to execute compensation task: %v", err)
	}
	fmt.Println("Compensation activity 'refund_payment' completed.")
	fmt.Printf("Active tokens: %v\n", instance.ActiveTokens)

	// Now that saga rolled back successfully, we complete the wait state or cancel the process
	fmt.Println("\n--- Saga rollback completed successfully. Cancelling the checkpoint task... ---")
	
	// For simplicity, we resume and complete the process
	err = engine.CompleteTask(instance, "checkpoint", nil)
	if err != nil {
		log.Fatalf("failed to complete wait state: %v", err)
	}
	fmt.Printf("Active tokens: %v\n", instance.ActiveTokens)

	// Reaching End Event
	err = engine.Step(context.Background(), instance)
	if err != nil {
		log.Fatalf("step failed: %v", err)
	}
	fmt.Printf("Process completed: %t\n", instance.Completed)
}
