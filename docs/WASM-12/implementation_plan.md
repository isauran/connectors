# Implementation Plan - WASM-12: Durable WASM Stability Improvements

This plan describes the steps for improving the stability of the `durable-wasm` engine and covering it with stress tests.

## User Review Required

> [!IMPORTANT]
> The signature of the `Engine.Execute` function will be changed to accept `ctx context.Context` as the first argument. This is a breaking change for code that calls the engine directly (e.g., examples and tests). We will need to adapt all `Execute` calls in the codebase.

## Proposed Changes

### [Component: durable-wasm Engine]

#### [MODIFY] [engine.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine.go)
- **HTTP Client Configuration**:
  - Add functional option `WithHTTPClient(client *http.Client) EngineOption` to support passing a reusable HTTP client.
  - Use a optimized global default HTTP client to avoid socket depletion.
- **Context-aware Execution**:
  - Update `Execute` signature:
    ```go
    func (e *Engine) Execute(ctx context.Context, instanceID string, entrypoint string, serverAddr string, shouldCrash bool) (bool, error)
    ```
  - Create a child cancelable context inside `Execute`: `ctx, cancel := context.WithCancel(ctx)`.
  - Defer `cancel()` call on `Execute` termination.
  - Pass this `ctx` to `handleDownload` network calls and `handleUpload` goroutine.
  - Use context-aware HTTP requests in `host_call_api`.
- **Wasmtime C-resource Cleanup**:
  - Add `defer store.Close()` to release C-heap memory explicitly in `Execute`.

#### [MODIFY] [engine_test.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine_test.go)
- Adapt all existing tests to the new `Execute(ctx, ...)` signature, passing `context.Background()`.
- Add new tests:
  - `TestExecuteCancellation`: Runs a WASM executing a long HTTP request, cancels the context, and verifies immediate unlock and resource cleanup.
  - `TestStorageErrorInjection`: Injects snapshot store returning mock I/O errors and verifies engine stability and absence of leaks.
  - `TestSoakStressTesting`: Runs 200 instances concurrently to verify memory usage and socket exhaustion limits.

---

### [Component: Examples and Other Tests]

Adapt all host example calls to the new signature.

#### [MODIFY] [examples](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/)
- Adapt `camunda/host/main.go` and `main_test.go`.
- Adapt `gotenberg-telegram/host/main.go` and `main_test.go`.
- Adapt `process-csv/host/main.go` and `main_test.go`.
- Adapt `temporal/host/main.go`.

---

## Verification Plan

### Automated Tests
- Run all tests in the `durable-wasm` module to validate stability and stress conditions:
  ```bash
  go test -v ./...
  ```
