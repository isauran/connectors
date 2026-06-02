# Task Checklist - WASM-12

- `[x]` Implement `WithHTTPClient` option and a default global HTTP client in `engine.go`
- `[x]` Add `context.Context` to the `Execute` signature and context cancellation in `engine.go`
- `[x]` Implement explicit Wasmtime store cleanup (`store.Close()`) in `Execute`
- `[x]` Adapt existing tests in `engine_test.go` to the new `Execute` signature
- `[x]` Adapt `examples/durable-s3/host/main.go` to the new `Execute` signature
- `[x]` Add `TestExecuteCancellation` (context cancellation) test in `engine_test.go`
- `[x]` Add `TestStorageErrorInjection` (error injection mock for SnapshotStore) in `engine_test.go`
- `[x]` Add `TestSoakStressTesting` (concurrent runs of 200 instances) in `engine_test.go`
- `[x]` Run all tests and verify engine functionality and stability
