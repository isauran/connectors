---
task: WASM-12
status: Completed
summary: Durable WASM Engine Stability Improvements and Stress Testing
---

# WASM-12: Durable WASM Engine Stability Improvements and Stress Testing

## Task Description
Implement stability optimizations in the `durable-wasm` package to prevent leaks of file descriptors, network goroutines, and Wasmtime C-heap memory. Cover the engine with new stress and fault-tolerance test suites.

## Requirements
1. **Reusable HTTP Client**:
   - Refactor `http.Client` out of the `Engine` structure or make it configurable to avoid Socket Exhaustion when creating multiple engines.
2. **Context-aware Execution**:
   - Change the `Execute` signature to `Execute(ctx context.Context, instanceID string, entrypoint string, serverAddr string, shouldCrash bool)`.
   - Link the lifecycle of asynchronous network streaming in `handleUpload` and HTTP calls in `host_call_api` with the session context.
3. **Wasmtime C-resource Cleanup**:
   - Call the explicit `store.Close()` method to prevent C-heap expansion (cgo).
4. **Test Cases**:
   - Add context cancellation tests.
   - Add error injection tests for `SnapshotStore`.
   - Add a stress test for concurrent execution of 200 WASM instances.
