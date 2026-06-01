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
5. **Scaffolding and Automation**:
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

We successfully ran the integration scenario via `make run` verifying the entire lifecycle:
1. **Run 1**: The worker executed Step 0, called `checkpoint`, and the host successfully saved a `131KB` linear memory snapshot before aborting execution with a simulated crash trap.
2. **Run 2**: The host initialized a fresh, clean WASM instance, read `snapshot.bin` off the disk, restored it directly to WASM memory, and resumed execution.
3. **Stream Verification**: The worker resumed from Step 1, initiated a download stream (GET /download), transformed the data (lowercase to uppercase), and streamed it back via an upload pipe (POST /upload).
4. **Mock Server Statistics**: The mock server verified receipt of all `42,500 bytes` successfully, with 100% of lowercase letters properly capitalized.
5. **Finalization**: The worker transitioned to Step 2, logged final stats, and completed at Step 3 returning `Result: 1` successfully.
