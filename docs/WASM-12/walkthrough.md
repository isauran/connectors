# Walkthrough - WASM-12: Durable WASM Stability Improvements

We have successfully resolved resource leak risks (TCP sockets, goroutines, and Wasmtime C-heap memory) in the `durable-wasm` package, cleaned up inline WAT fixtures, and completed comprehensive stress verification.

## Changes Made

### 1. Reusable HTTP Client configuration
- Implemented `EngineOption` with a `WithHTTPClient` builder.
- Configured a global `defaultHTTPClient` with optimized connection pooling parameters (`MaxIdleConns: 100`, `MaxIdleConnsPerHost: 10`, `IdleConnTimeout: 90s`) to prevent socket exhaustion.

### 2. Context-Aware Execution
- Modified the signature of `Engine.Execute` to accept `context.Context` as the first argument.
- Propagated the execution context through streaming handles (`handleDownload`, `handleUpload`) and external API HTTP calls (`host_call_api`).
- Added check for context cancellation (`ctx.Err() != nil`) right after the WASM execution call `runFunc.Call(store)` to return `context.Canceled` immediately.

### 3. C-Heap Resource Cleanup
- Added deferred `store.Close()` inside `Execute` to release Wasmtime store allocations, preventing cgo memory leaks.

### 4. SQLite Concurrency Enhancements
- Enforced `db.SetMaxOpenConns(1)` on `SqliteSnapshotStore` to serialise concurrent writes and prevent `database is locked (SQLITE_BUSY)` conflicts.
- Separated PRAGMA calls inside `NewSqliteSnapshotStore` to ensure proper driver execution.

### 5. Test WAT Fixtures Refactoring (Clean Code)
- Extracted all raw inline WAT string definitions from `engine_test.go` into separate `.wat` files in `durable-wasm/testdata/`.
- Created [testdata.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/testdata/testdata.go) next to the `.wat` files to hold `//go:embed` directives.
- Imported the `testdata` package inside `engine_test.go` for cleaner, modular, and maintainable test code.

### 6. Adapted Host Examples
- Adapted all existing examples and tests to support the new signature:
  - `examples/camunda/host/main.go` and `main_test.go`
  - `examples/gotenberg-telegram/host/main.go` and `main_test.go`
  - `examples/process-csv/host/main.go` and `main_test.go`
  - `examples/temporal/host/main.go`
  - `examples/durable-s3/host/main.go`

### 7. Stability and Stress Tests
- Added three new tests in `engine_test.go`:
  - `TestExecuteCancellation`: Verifies that context cancellation halts execution immediately.
  - `TestStorageErrorInjection`: Injects errors during metadata/snapshot operations to ensure robust handling and cleanup.
  - `TestSoakStressTesting`: Concurrent execution of 200 WASM instances using temporary disk databases to simulate high-load production environments.

## Verification Results

We executed the full test suite in `durable-wasm`:
```bash
go test -v ./...
```
All tests passed successfully, including stress tests, cancellation tests, and all five host-app integration tests.
