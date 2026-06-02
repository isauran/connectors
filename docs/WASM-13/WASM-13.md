---
task: WASM-13
status: In Progress
summary: Implement S3SnapshotStore for distributed snapshot storage with OCC optimistic locking
---

# WASM-13: S3-Compatible Snapshot Store Integration

## Task Description
Implement a new storage provider `S3SnapshotStore` supporting AWS S3 and any S3-compatible object storage (MinIO, Ceph, Cloudflare R2, etc.). This allows distributed Stateless host nodes to horizontally scale Durable WASM executions without hard coupling to local SQLite databases or shared Postgres clusters.

## Requirements
1. **`SnapshotStore` Interface Implementation**:
   Create `S3SnapshotStore` in a new file `s3_store.go` implementing all methods of the `SnapshotStore` interface.
2. **S3 Path Structure**:
   Store instance assets under the following keys:
   - Full snapshots: `instances/{instance_id}/snapshot.bin`
   - Memory deltas: `instances/{instance_id}/deltas.json`
   - Execution Log (Oplog): `instances/{instance_id}/oplog.json`
   - Execution Metadata (OCC): `instances/{instance_id}/meta.json`
   - WASM registry modules: `wasm_modules/{wasm_hash}.wasm`
3. **Optimistic Concurrency Control (OCC)**:
   - Add an `ETag string` field to the `InstanceMeta` structure (in `engine.go`) to track S3 object tags.
   - Use S3 conditional writes (`If-Match` header with the current ETag for updates and `If-None-Match: *` for initial insertions) inside `SaveMetadata` to prevent concurrent instance executions.
4. **Integration Testing**:
   - Write a `TestS3SnapshotStore` integration test in `engine_test.go` that executes with a local MinIO/S3 endpoint or skips if credentials are not configured.
