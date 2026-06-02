# Task Checklist - WASM-15: Go SDK for Durable WASM

- `[x]` Implement SDK in root of `durable-wasm` (`sdk.go` & `sdk_stub.go` with build tags)
  - `[x]` Import host functions (`checkpoint`, `host_get_time`, `host_call_api`, `stream_data`)
  - `[x]` Implement SDK API wrappers (`GetTime`, `CallAPI`)
  - `[x]` Implement `Reader` (`io.Reader`) and `Writer` (`io.WriteCloser`)
  - `[x]` Implement Fluent API (`Workflow` and `APICall` builders)
- `[x]` Add build tags `//go:build !wasm` to host files (`engine.go`, `s3_store.go`, `fs_store.go`)
- `[x]` Refactor `examples/s3-store/worker/main.go` to use Fluent API of `durable` package
- `[x]` Rebuild `s3-store` worker WebAssembly binary and run all tests to verify stability
