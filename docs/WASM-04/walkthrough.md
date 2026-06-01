# Walkthrough - WASM Durable Execution Engine MVP

This document summarizes the changes, architecture, and verification results for the Durable Execution Engine based on WebAssembly (WASM).

## Accomplishments
1. **Reusable `durable` Package**:
   - Created a modular, highly-reusable package `durable` in the Go host orchestator ([engine.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/host/durable/engine.go)).
   - Exposes `Engine`, `Session`, and `SnapshotStore` (e.g. `FileSnapshotStore`) to manage execution lifecycles, memory restores, and stream-first HTTP transactions.
2. **TinyGo WASM Worker**:
   - Developed a step-based state machine worker ([main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/worker/main.go)) that simulates sequential business steps: Initialization, Data Streaming with transformation, and Finalization.
3. **$O(1)$ RAM Stream-first HTTP**:
   - Developed a memory-efficient streaming protocol that transfers data in chunks (4KB) directly to/from WASM memory to the network using `io.Pipe` and Go's `http.Client`. This guarantees $O(1)$ memory consumption.
4. **Snapshot & Restore Mechanism**:
   - Implemented checkpointing: the WASM module halts and asks the Go Host to dump its linear memory (`[]byte`) to disk.
   - Designed a recovery routine where a brand new, clean WASM instance is spawned, its linear memory is overwritten with the snapshot, and the execution is resumed from the last saved state machine step.
5. **Real-world Backend Examples**:
   - Created a realistic backend ETL pipeline example in [durable-wasm/examples/process-csv](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/process-csv).
   - The worker wraps the low-level `stream_data` host function into a standard `io.Reader` and `io.Writer` interface, enabling the use of Go's built-in `encoding/csv` and `encoding/json` streaming libraries.
   - Streams mock CSV users from the host, validates emails and parse amounts, transforms them into target JSON records, and posts them back chunk-by-chunk with strict $O(1)$ RAM usage.
6. **Scaffolding and Automation**:
   - Written a custom [Makefile](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/Makefile) to build the worker, run the host orchestrator, clean build artifacts, and generate a minimal scratch-based Docker image.
   - Prepared [Dockerfile](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/host/Dockerfile) utilizing multi-stage static builds deployed on `scratch`.

## Architecture Diagram
```
WASM Worker (TinyGo)                   Go Host (Wasmtime-Go)              Mock REST API
--------------------                   ---------------------              -------------
Step 0: Init         ---------->       checkpoint() -> Write Snapshot
(Simulated Crash)                               (Abort Execution)

               *** Restarting Host - Reloading Memory Snapshot ***

Step 1: Stream       ---------->       stream_data(0: read chunk)    <--- HTTP GET /download
                     ---------->       stream_data(1: write chunk)   ---> HTTP POST /upload
                     ---------->       stream_data(1: EOF)           ---> Close POST Request

Step 2: Finalize     ---------->       checkpoint() -> Final Snapshot
Step 3: Completed    ---------->       Result: 1 (Complete)
```

## Validation & Verification

### Automated Unit Tests
We developed a complete automated integration unit test [engine_test.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/host/durable/engine_test.go).
Running `make test` outputs:
```bash
$ make test
Building WASM worker using TinyGo...
tinygo build -o worker/worker.wasm -target=wasi worker/main.go
Running tests...
cd host && go test -v ./...
?   	github.com/nativebpm/connectors/durable-wasm/host	[no test files]
=== RUN   TestDurableExecutionLifecycle
[ENGINE] Invoking entrypoint 'run'...
[WASM WORKER] Step 0: Starting initialization...
[WASM WORKER] Step 0 completed. Initiating checkpoint.
[ENGINE] 'checkpoint' invoked for instance 'test-worker-instance'
[ENGINE] Snapshot successfully saved (131072 bytes)
[ENGINE] Simulating host crash. Aborting WASM execution.
[ENGINE] Found saved snapshot for 'test-worker-instance'. Restoring memory...
[ENGINE] Memory snapshot successfully restored.
[ENGINE] Invoking entrypoint 'run'...
[WASM WORKER] Step 1: Processing data stream...
[ENGINE] GET Request to http://localhost:18081/download (Stream-first)
[ENGINE] GET Stream EOF. Closing response.
[ENGINE] POST Request to http://localhost:18081/upload (Stream-first via io.Pipe)
[WASM WORKER] Stream EOF. All data received.
[ENGINE] Closing upload stream (EOF). Waiting for response...
[ENGINE] POST completed successfully.
[WASM WORKER] Step 1 completed. Initiating checkpoint.
[ENGINE] 'checkpoint' invoked for instance 'test-worker-instance'
[ENGINE] Snapshot successfully saved (262144 bytes)
[WASM WORKER] Step 2: Finalizing business process...
[WASM WORKER] Total bytes processed and transformed: 31
[WASM WORKER] Step 2 completed. Initiating final checkpoint.
[ENGINE] 'checkpoint' invoked for instance 'test-worker-instance'
[ENGINE] Snapshot successfully saved (262144 bytes)
[WASM WORKER] Execution already completed.
[ENGINE] Execution completed. Result: 1
--- PASS: TestDurableExecutionLifecycle (0.08s)
PASS
ok  	github.com/nativebpm/connectors/durable-wasm/host/durable	0.633s
```

### CSV Processing Pipeline Example (Backend Logic Demo)
We verified the ETL pipeline example by running `make run-csv-example` inside `durable-wasm/`:
```bash
$ make run-csv-example
Building WASM worker for CSV processing example using TinyGo...
tinygo build -o worker/worker.wasm -target=wasi worker/main.go
Running CSV processing example host...
cd host && go run main.go
[HOST] Starting CSV-to-JSON Pipeline Durable Execution Example...
[MOCK SERVER] Listening on http://localhost:18082

--- RUN 1: Executing WASM CSV pipeline with simulated crash ---
[ENGINE] Invoking entrypoint 'run'...
[CSV WORKER] Step 0: Initializing CSV processor...
[CSV WORKER] Step 0 completed. Saving checkpoint.
[ENGINE] 'checkpoint' invoked for instance 'csv-worker-instance'
[ENGINE] Snapshot successfully saved (131072 bytes)
[ENGINE] Simulating host crash. Aborting WASM execution.
[HOST] Execution successfully suspended/crashed: error while executing at wasm backtrace:
    0: 0x4afa5 - main!run
note: using the `WASMTIME_BACKTRACE_DETAILS=1` environment variable may show more debugging information

Caused by:
    simulated_host_crash
[HOST] Verified that snapshot file was written to disk.

--- RUN 2: Restoring from snapshot and processing CSV stream ---
[ENGINE] Found saved snapshot for 'csv-worker-instance'. Restoring memory...
[ENGINE] Memory snapshot successfully restored.
[ENGINE] Invoking entrypoint 'run'...
[CSV WORKER] Step 1: Processing CSV stream and generating JSON output...
[ENGINE] GET Request to http://localhost:18082/download (Stream-first)
[ENGINE] GET Stream EOF. Closing response.
[ENGINE] POST Request to http://localhost:18082/upload (Stream-first via io.Pipe)
[MOCK SERVER] Received transformed JSON stream:
{"id":"1","name":"Alice Johnson","email":"alice@example.com","amount":120.5,"status":"active"}
{"id":"2","name":"Bob Smith","email":"bob-invalid-email","amount":250,"status":"invalid_email"}
{"id":"3","name":"Charlie Brown","email":"charlie@example.com","amount":0,"status":"invalid_amount"}
{"id":"4","name":"David Miller","email":"david@example.com","amount":450,"status":"active"}
[ENGINE] Closing upload stream (EOF). Waiting for response...
{"id":"5","name":"Eve Adams","email":"eve@example.com","amount":90.25,"status":"active"}
[ENGINE] POST completed successfully.
[CSV WORKER] Step 1 completed. Saving checkpoint.
[ENGINE] 'checkpoint' invoked for instance 'csv-worker-instance'
[ENGINE] Snapshot successfully saved (262144 bytes)
[CSV WORKER] Step 2: Finalizing business validation...
[CSV WORKER] Total valid records processed: 3
[CSV WORKER] Sum of active user amounts: 660.75
[CSV WORKER] Step 2 completed. Final checkpoint.
[ENGINE] 'checkpoint' invoked for instance 'csv-worker-instance'
[ENGINE] Snapshot successfully saved (262144 bytes)
[CSV WORKER] Execution completed.
[ENGINE] Execution completed. Result: 1

[HOST] Durable WASM CSV pipeline example completed successfully.
```
