# Task Checklist - WASM-14: Remove Relational Databases (SQLite & Postgres)

- `[x]` Delete obsolete files and directories (sqlite_store.go, postgres_store.go, schema.go, atlas.hcl, queries/, schema/)
- `[x]` Define inMemorySnapshotStore in `engine_test.go` and update core engine tests
- `[x]` Update benchmark tests in `engine_bench_test.go`
- `[x]` Update example `examples/camunda` (host/main.go & main_test.go) to use FileSnapshotStore
- `[x]` Update example `examples/temporal` (host/main.go) to use FileSnapshotStore
- `[x]` Update example `examples/process-csv` (host/main_test.go) to use FileSnapshotStore
- `[x]` Update example `examples/gotenberg-telegram` (host/main_test.go) to use FileSnapshotStore
- `[x]` Delete obsolete example `examples/durable-s3` entirely
- `[x]` Clean up Makefile (remove AtlasGo and SQL schemas targets)
- `[x]` Update README.md and README.ru.md
- `[x]` Run `go mod tidy` to clean up dependencies
- `[x]` Move FileSnapshotStore from `engine.go` to `fs_store.go`
- `[x]` Run `go test ./...` to verify all tests and examples compile and pass
