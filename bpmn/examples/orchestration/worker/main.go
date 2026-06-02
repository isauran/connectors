//go:build wasm

package main

import (
	"encoding/json"
	"fmt"

	"github.com/nativebpm/connectors/wasman"
)

// Variables represents the schema of input/output variables of the process.
type Variables map[string]interface{}

// State holds the execution state of our workflow steps inside WASM.
type State struct {
	Vars Variables
}

var state = &State{
	Vars: make(Variables),
}

//export run
func run() int32 {
	return wasman.NewWorkflow().
		Step(state.loadVariables).
		Step(state.executeBusinessLogic).
		Step(state.saveVariables).
		Run()
}

func main() {}

func (s *State) loadVariables() error {
	println("[WASM WORKER] Step 1: Loading variables from host...")
	
	// Decode JSON variables from the host's download stream
	err := json.NewDecoder(wasman.Reader).Decode(&s.Vars)
	if err != nil {
		return fmt.Errorf("failed to decode variables: %w", err)
	}
	
	fmt.Printf("[WASM WORKER] Received variables: %+v\n", s.Vars)
	return nil
}

func (s *State) executeBusinessLogic() error {
	println("[WASM WORKER] Step 2: Executing business logic...")

	approvedVal, ok := s.Vars["approved"]
	if !ok {
		return fmt.Errorf("variable 'approved' not found")
	}

	approved, ok := approvedVal.(bool)
	if !ok {
		return fmt.Errorf("variable 'approved' must be boolean")
	}

	if approved {
		println("[WASM WORKER] Processing payout payment...")
		s.Vars["payment_status"] = "success"
		s.Vars["transaction_id"] = "TXN-987654321"
	} else {
		println("[WASM WORKER] Logging rejection details...")
		s.Vars["rejection_logged"] = true
		s.Vars["rejection_reason"] = "DMN rules evaluation returned false"
	}

	return nil
}

func (s *State) saveVariables() error {
	println("[WASM WORKER] Step 3: Saving updated variables to host...")

	// Encode and stream the updated variables to the host's upload stream
	err := json.NewEncoder(wasman.Writer).Encode(s.Vars)
	if err != nil {
		return fmt.Errorf("failed to encode variables: %w", err)
	}

	// Close the writer to signal EOF
	return wasman.Writer.Close()
}
