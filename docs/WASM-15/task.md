# Task Checklist - WASM-15: Go SDK Workflow State Pattern Migration

- `[x]` Migrate `examples/s3-store/worker/main.go` to State Struct pattern
- `[x]` Migrate `examples/camunda/worker/main.go` to State Struct pattern
- `[x]` Migrate `examples/process-csv/worker/main.go` to State Struct pattern
- `[x]` Migrate `examples/gotenberg-telegram/worker/main.go` to State Struct pattern
- `[x]` Migrate `examples/temporal/worker/main.go` to State Struct pattern
- `[ ]` Rebuild all WebAssembly binaries (`worker.wasm`) using TinyGo
- `[ ]` Run host tests (`go test -v ./...`) to verify snapshot and restore functions
- `[ ]` Commit changes, update release tag `durable-wasm/v0.0.5`
