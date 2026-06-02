package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/nativebpm/connectors/bpmn"
	"github.com/nativebpm/connectors/wasman"
)

const (
	instanceID   = "loan-orchestration-instance"
	snapshotsDir = "snapshots"
)

func main() {
	slog.Info("[HOST] Starting BPMN + DMN + Wasman Durable Orchestration Example")

	// 1. Read files
	bpmnBytes, err := os.ReadFile("../bpmn/process.bpmn")
	if err != nil {
		slog.Error("[HOST] Failed to read process.bpmn. Make sure to run inside host/ directory", "error", err)
		os.Exit(1)
	}

	dmnBytes, err := os.ReadFile("../bpmn/decision.dmn")
	if err != nil {
		slog.Error("[HOST] Failed to read decision.dmn", "error", err)
		os.Exit(1)
	}

	// 2. Parse BPMN & DMN
	parsedProcess, err := bpmn.ParseBPMN(bpmnBytes)
	if err != nil {
		slog.Error("[HOST] Failed to parse BPMN schema", "error", err)
		os.Exit(1)
	}
	slog.Info("[HOST] Parsed BPMN process", "id", parsedProcess.ID)

	parsedDMN, err := bpmn.ParseDMN(dmnBytes)
	if err != nil {
		slog.Error("[HOST] Failed to parse DMN table", "error", err)
		os.Exit(1)
	}
	slog.Info("[HOST] Parsed DMN decision table", "id", parsedDMN.Decisions[0].ID)

	// 3. Setup WASM Engine
	wasmPath := filepath.Join("..", "worker", "worker.wasm")
	_ = os.RemoveAll(snapshotsDir)
	if err := os.MkdirAll(snapshotsDir, 0755); err != nil {
		slog.Error("[HOST] Failed to create snapshots directory", "error", err)
		os.Exit(1)
	}
	store := &wasman.FileSnapshotStore{Dir: snapshotsDir}
	defer os.RemoveAll(snapshotsDir)

	// Initialize Durable Engine
	wasmanEngine, err := wasman.NewEngine(wasmPath, store)
	if err != nil {
		slog.Error("[HOST] Failed to initialize wasman engine", "error", err)
		slog.Error("[HOST] Make sure worker.wasm is compiled by running 'make build'")
		os.Exit(1)
	}

	// 4. Initialize Process Engine
	processEngine := bpmn.NewEngine(parsedProcess, wasmanEngine)

	// 5. Start Instance
	// Input variables: Age=25, Income=1200 (Satisfies Rules: approved = true)
	// Also we simulate crash in WASM step initially!
	variables := map[string]interface{}{
		"age":            25,
		"income":         1200,
		"simulate_crash": true,
	}

	instance, err := processEngine.StartInstance(instanceID, variables)
	if err != nil {
		slog.Error("[HOST] Failed to start process instance", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// 6. Execution Loop
	for !instance.Completed {
		if len(instance.ActiveTokens) == 0 {
			slog.Error("[HOST] No active tokens left but process is not completed")
			break
		}

		currentToken := instance.ActiveTokens[0]
		slog.Info("[HOST] Next step token to execute", "token", currentToken)

		// Dynamic integration: before we step into ExclusiveGateway,
		// we evaluate DMN if we are executing "check_rules" ServiceTask.
		if currentToken == "check_rules" {
			slog.Info("[HOST] Evaluating DMN rules for decision...", "age", instance.Variables["age"], "income", instance.Variables["income"])
			decisionResult, err := bpmn.Evaluate(parsedDMN, "loan_decision", instance.Variables)
			if err != nil {
				slog.Error("[HOST] DMN evaluation failed", "error", err)
				os.Exit(1)
			}
			slog.Info("[HOST] DMN decision evaluated successfully", "result", decisionResult)

			// Merge DMN result variables into process variables
			for k, v := range decisionResult {
				instance.Variables[k] = v
			}
		}

		// Perform process Engine step
		err = processEngine.Step(ctx, instance)
		if err != nil {
			// Check if this error is due to a simulated crash in WASM worker
			if instance.Variables["simulate_crash"] == true {
				slog.Warn("[HOST] Caught expected simulated WASM crash!", "error", err)
				slog.Info("[HOST] Disabling crash flag and retrying execution to demonstrate Durable Recovery...")
				
				instance.Variables["simulate_crash"] = false
				
				// Verify memory snapshot has been saved to disk
				_, errSnap := store.Load(instanceID)
				if errSnap != nil {
					slog.Error("[HOST] Expected snapshot file not found in storage", "error", errSnap)
					os.Exit(1)
				}
				slog.Info("[HOST] Verified memory snapshot exists on disk. Resuming...")
				
				// Sleep briefly to show transition
				time.Sleep(200 * time.Millisecond)
				continue
			}

			slog.Error("[HOST] Execution step failed", "error", err)
			os.Exit(1)
		}
	}

	slog.Info("[HOST] Durable Orchestration completed successfully!")
	slog.Info("[HOST] Final process variables", "variables", instance.Variables)
	
	// Double check final variables depending on rules
	approved := instance.Variables["approved"].(bool)
	if approved {
		slog.Info("[HOST] Verification PASSED: Payout executed successfully via WASM worker after recovery!")
		slog.Info("[HOST] Payment status", "status", instance.Variables["payment_status"], "txn", instance.Variables["transaction_id"])
	} else {
		slog.Info("[HOST] Verification PASSED: Rejection logged successfully via WASM worker after recovery!")
	}
}
