# Walkthrough - WASM-13: S3SnapshotStore Integration

As part of the task **WASM-13**, support for S3-compatible object stores (`S3SnapshotStore`) was successfully implemented as a new `SnapshotStore` backend for the `durable-wasm` engine.

## Changes list

1. **Added `ETag` field to `InstanceMeta`**:
   - In [engine.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine.go#L33), the `InstanceMeta` struct has been extended with an `ETag string` field (tagged with `json:"etag,omitempty"`). This enables Optimistic Concurrency Control (OCC) directly using S3 API capabilities.
   - The change is backward-compatible and does not affect SQLite/PostgreSQL stores.

2. **Created S3-compatible snapshot store `S3SnapshotStore`**:
   - Implemented the `S3SnapshotStore` struct in a new file [s3_store.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/s3_store.go), satisfying the `SnapshotStore` interface.
   - Storage paths for instance data:
     - Memory snapshots: `instances/{id}/snapshot.bin`
     - Memory deltas: `instances/{id}/deltas.json`
     - Execution logs (oplog): `instances/{id}/oplog.json`
     - Instance metadata: `instances/{id}/meta.json`
     - WASM modules: `wasm/{hash}.wasm`
   - Optimistic lock check in `SaveMetadata` using S3 headers:
     - `If-None-Match: *` for the initial metadata insert.
     - `If-Match: <ETag>` for updates on existing metadata.
     - When a concurrency conflict occurs (`status 412 Precondition Failed`), the method returns `false, nil`.

3. **Integration Testing**:
   - Added `TestS3SnapshotStore` in [engine_test.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine_test.go).
   - Verifies saving snapshots, deltas, oplog, correct OCC locking checks (insert, update, stale ETag write failures) and saving/loading WASM modules.
   - The test skips gracefully if S3 configuration variables are not set.

## Test Results

All tests have been successfully verified locally against MinIO running in Docker:

```
=== RUN   TestS3SnapshotStore
--- PASS: TestS3SnapshotStore (0.06s)
PASS
ok  	github.com/nativebpm/connectors/durable-wasm	0.603s
```
