# Implementation Plan - WASM Durable Execution Engine MVP

This document describes the plan for creating a WebAssembly-based Durable Execution Engine (WASM-04).

## User Review Required

> [!IMPORTANT]
> - **TinyGo Compiler**: TinyGo must be installed on the local machine (`brew install tinygo`). If it's not present, we will automate its installation during the implementation phase.
> - **Wasmtime-Go Dependency**: We will use the official `github.com/bytecodealliance/wasmtime-go/v20` library. It requires CGO to compile.
> - **Simulation of Failure**: We will demonstrate durability by running the WASM module, executing a checkpoint, simulating a host crash/restart, restoring the linear memory snapshot, and continuing execution from the last saved state.

## Proposed Changes

We will create a new directory `durable-wasm` in the root of the repository containing the host orchestrator and the WASM worker.

---

### Root Workspace Configuration

#### [MODIFY] [go.work](file:///Users/user/github.com/nativebpm/connectors/go.work)
Add paths to the new host and worker modules:
- `./durable-wasm/host`
- `./durable-wasm/worker`

---

### Component: Durable WASM Engine

#### [NEW] [Makefile](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/Makefile)
Automate building the WASM worker, running the host, executing integration tests, and building the scratch Docker image.

#### [NEW] [worker/go.mod](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/worker/go.mod)
Go module file for the TinyGo WASM worker.

#### [NEW] [worker/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/worker/main.go)
Implementation of the business-logic worker in Go (compiled via TinyGo to WASM).
- Implements a state machine to track step execution.
- Imports `checkpoint` and `stream_data` host functions.
- Processes data in chunks of 4KB to maintain $O(1)$ memory consumption.

#### [NEW] [host/go.mod](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/host/go.mod)
Go module file for the orchestrator host app, requiring `wasmtime-go`.

#### [NEW] [host/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/host/main.go)
Implementation of the Go orchestrator.
- Sets up Wasmtime environment with WASI.
- Implements `stream_data` host function using `io.Pipe` for $O(1)$ RAM upload/download.
- Implements `checkpoint` host function to snapshot linear memory (`[]byte`) to a local file.
- Spawns a test HTTP server to simulate external REST API endpoints (`/download` and `/upload`).
- Orchestrates execution: runs worker, stops at checkpoint (simulated failure), initializes a clean WASM instance, restores memory, and continues execution.

#### [NEW] [host/Dockerfile](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/host/Dockerfile)
Multi-stage build Dockerfile:
- Stage 1: Build Go host with `wasmtime-go` statically linked (using a base image containing required C libraries).
- Stage 2: Deploy to a `scratch` container.

---

## Verification Plan

### Automated Tests
We will write an automated integration flow in `host/main.go` that:
1. Starts the mock HTTP server.
2. Runs the worker from scratch.
3. Interrupts the worker at Step 1 and writes memory snapshot.
4. Spawns a new clean Wasmtime instance, loads the snapshot, and executes again.
5. Verifies that the data was fully processed, transformed, and uploaded to the mock server, matching the initial input data.

We will run:
```bash
make -C durable-wasm build
make -C durable-wasm run
```
And verify that the logs show successful execution, crash simulation, memory restore, and finalization.
