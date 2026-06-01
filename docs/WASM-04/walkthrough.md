# Walkthrough - WASM Durable Execution Engine MVP

This document summarizes the changes, architecture, and verification results for the Durable Execution Engine based on WebAssembly (WASM).

## Accomplishments
1. **Reusable Unified `durable-wasm` Module**:
   - Combined the separate modules `durable-wasm/host` and `durable-wasm/worker` into a single Go module `github.com/nativebpm/connectors/durable-wasm` defined in [go.mod](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/go.mod).
   - This enables developers to import the entire engine package using a single Git tag (e.g. `durable-wasm/vX.Y.Z`) and cleanly import the `durable` subpackage from `github.com/nativebpm/connectors/durable-wasm/host/durable`.
2. **Modular Architecture**:
   - Exposes `Engine`, `Session`, and `SnapshotStore` (e.g. `FileSnapshotStore`) to manage execution lifecycles, memory restores, and stream-first HTTP transactions.
3. **TinyGo WASM Worker**:
   - Developed a step-based state machine worker that simulates sequential business steps: Initialization, Data Streaming with transformation, and Finalization.
4. **$O(1)$ RAM Stream-first HTTP**:
   - Developed a memory-efficient streaming protocol that transfers data in chunks (4KB) directly to/from WASM memory to the network using `io.Pipe` and Go's `http.Client`. This guarantees $O(1)$ memory consumption.
5. **Snapshot & Restore Mechanism**:
   - Implemented checkpointing: the WASM module halts and asks the Go Host to dump its linear memory (`[]byte`) to disk.
   - Designed a recovery routine where a brand new, clean WASM instance is spawned, its linear memory is overwritten with the snapshot, and the execution is resumed from the last saved state machine step.
6. **Real-world Backend Examples**:
   - **CSV processing pipeline** ([process-csv](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/process-csv)): streams mock CSV users, validates fields, transforms records into JSON, and posts them back chunk-by-chunk with strict $O(1)$ RAM usage.
   - **Temporal Activity orchestration** ([temporal](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/temporal)): simulates a long-running activity that fetches parameters, processes math operations, saves final records to database, and uses checkpoints for heartbeating / recovery.
   - **Camunda External Task** ([camunda](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/camunda)): simulates a BPMN Service Task where the worker performs multi-step execution (inventory check + payment capture) with persistent checkpoints.
   - **Gotenberg & Telegram document pipeline** ([gotenberg-telegram](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/gotenberg-telegram)): streams a DOCX document from Telegram, converts it to PDF via Gotenberg API, and uploads the PDF back to the Telegram user. The file bytes are kept in global slices in WASM linear memory, meaning they are preserved across host failures.
7. **Scaffolding and Automation**:
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
Running `make test` outputs `PASS` successfully.

### Example 1: CSV Processing Pipeline
Executed by running `make run-csv-example` inside `durable-wasm/`. Parses and validates records, transforming CSV inputs into structured JSON streams.

### Example 2: Temporal Activity
Executed by running `make run-temporal-example` inside `durable-wasm/`. Demonstrates a multi-step math/processing activity that persists its progress securely on disk.

### Example 3: Camunda External Task
Executed by running `make run-camunda-example` inside `durable-wasm/`. Deploys process definition [process.bpmn](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/camunda/bpmn/process.bpmn) to local Camunda Platform 7 REST API, starts a process instance, polls for tasks, and runs the WASM-based multi-step worker with crash/restore simulation.

### Example 4: Gotenberg & Telegram Document Pipeline
Executed by running `make run-gotenberg-telegram-example` inside `durable-wasm/`. Downloads a DOCX document from Telegram Bot API, converts it via Gotenberg, and streams the PDF back to the user chat.
