# Implementation Plan - WASM-13: S3SnapshotStore Integration

This plan describes the steps for integrating a new S3-compatible snapshot store into the `durable-wasm` engine.

## User Review Required

> [!IMPORTANT]
> To support Optimistic Concurrency Control (OCC) directly in S3 without secondary databases, we will extend the `InstanceMeta` structure with an `ETag string` field (tagged with `json:"etag,omitempty"`). This is a backward-compatible change that does not affect SQLite and Postgres stores (they will simply serialize/deserialize it in DBs or JSON).

## Proposed Changes

### [Component: durable-wasm S3 Storage Provider]

#### [MODIFY] [engine.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine.go)
- Add `ETag string` to `InstanceMeta` structure:
  ```go
  type InstanceMeta struct {
      InstanceID string `json:"instance_id"`
      WasmHash   string `json:"wasm_hash"`
      Version    int    `json:"version"`
      ETag       string `json:"etag,omitempty"`
  }
  ```

#### [NEW] [s3_store.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/s3_store.go)
- Create `S3SnapshotStore` struct:
  ```go
  type S3SnapshotStore struct {
      client *s3.Client
      bucket string
  }
  ```
- Implement `NewS3SnapshotStore(ctx context.Context, bucket string, opts ...func(*s3.Options)) (*S3SnapshotStore, error)` to initialize the S3 client.
- Implement `Save(id, data)` and `Load(id)`: Write and read to `instances/{id}/snapshot.bin`.
- Implement `SaveDeltas`, `LoadDeltas`, `SaveOplog`, `LoadOplog` with JSON serialization to `instances/{id}/deltas.json` and `instances/{id}/oplog.json`.
- Implement `SaveMetadata` with OCC:
  - Use `IfNoneMatch = aws.String("*")` for the initial insert (Version 0).
  - Use `IfMatch = aws.String(meta.ETag)` for updates (Version > 0).
  - Catch S3 API errors with code `PreconditionFailed` (status 412) and return `false, nil` indicating OCC conflict.
  - On successful writes, save the response `ETag` to `meta.ETag`.
- Add compile-time interface verification: `var _ SnapshotStore = (*S3SnapshotStore)(nil)`.

#### [MODIFY] [engine_test.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine_test.go)
- Add `TestS3SnapshotStore` integration test in `engine_test.go`.
- The test will try to connect to a local MinIO or S3 endpoint using environment variables: `S3_ENDPOINT`, `S3_BUCKET`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`.
- If environment variables are not set, the test will skip gracefully via `t.Skip`.

---

## Verification Plan

### Automated Tests
- Run all tests in the `durable-wasm` module to verify compilation, interface compatibility, and correctness:
  ```bash
  go test -v ./...
  ```
