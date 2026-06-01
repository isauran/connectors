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
   - **CSV processing pipeline** ([process-csv](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/process-csv)): streams mock CSV users, validates fields, transforms records into JSON, and posts them back chunk-by-chunk with strict $O(1)$ RAM usage.
   - **Camunda & Temporal orchestration** ([camunda-temporal](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/camunda-temporal)): simulates billing transactions (Temporal Activity) and updating CRM databases (Camunda Task) utilizing intermediate checkpoints to ensure execution safety.
   - **Gotenberg & Telegram document pipeline** ([gotenberg-telegram](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/gotenberg-telegram)): streams a DOCX document from Telegram, converts it to PDF via Gotenberg API, and uploads the PDF back to the Telegram user. The file bytes are kept in global slices in WASM linear memory, meaning they are preserved across host failures.
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
Running `make test` outputs `PASS` successfully.

### Example 1: CSV Processing Pipeline
Executed by running `make run-csv-example` inside `durable-wasm/`. Parses and validates records, transforming CSV inputs into structured JSON streams.

### Example 2: Camunda & Temporal Orchestration
Executed by running `make run-camunda-temporal-example` inside `durable-wasm/`. Demonstrates a transactional billing step (Temporal Activity) followed by CRM updating (Camunda Task), validating transactional persistence across checkpoints.

### Example 3: Gotenberg & Telegram Document Pipeline
Executed by running `make run-gotenberg-telegram-example` inside `durable-wasm/`. Downloads a DOCX document from Telegram Bot API, converts it via Gotenberg, and streams the PDF back to the user chat. Demonstrates automatic caching of downloaded documents in the WASM memory snapshot.
