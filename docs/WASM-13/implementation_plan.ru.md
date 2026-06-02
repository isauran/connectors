# Plan of Implementation - WASM-13: S3SnapshotStore Integration

Этот план описывает шаги по интеграции нового S3-совместимого хранилища снимков памяти в движок `durable-wasm`.

## User Review Required

> [!IMPORTANT]
> Для корректной поддержки Optimistic Concurrency Control (OCC) в S3 без дополнительных БД мы расширим структуру `InstanceMeta` полем `ETag string` (с тегом `json:"etag,omitempty"`). Это обратно-совместимое изменение, которое не сломает другие провайдеры (SQLite и Postgres будут игнорировать или просто сохранять это поле в JSON/БД).

## Proposed Changes

### [Component: durable-wasm S3 Storage Provider]

#### [MODIFY] [engine.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine.go)
- Добавить поле `ETag string` в структуру `InstanceMeta`:
  ```go
  type InstanceMeta struct {
      InstanceID string `json:"instance_id"`
      WasmHash   string `json:"wasm_hash"`
      Version    int    `json:"version"`
      ETag       string `json:"etag,omitempty"`
  }
  ```

#### [NEW] [s3_store.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/s3_store.go)
- Создать `S3SnapshotStore` со следующими полями:
  ```go
  type S3SnapshotStore struct {
      client *s3.Client
      bucket string
  }
  ```
- Реализовать метод `NewS3SnapshotStore(ctx context.Context, bucket string, opts ...func(*s3.Options))` для создания клиента через AWS SDK v2.
- Реализовать методы `Save(id, data)` и `Load(id)`: считывают/записывают бинарный объект `instances/{id}/snapshot.bin`.
- Реализовать `SaveDeltas`, `LoadDeltas`, `SaveOplog`, `LoadOplog` с сериализацией в JSON по путям `instances/{id}/deltas.json` и `instances/{id}/oplog.json` соответственно.
- Реализовать `SaveMetadata` с OCC:
  - Использовать `If-None-Match: *` для новой записи (версия 0).
  - Использовать `If-Match: ETag` для обновления (версия > 0).
  - Если возвращается ошибка `PreconditionFailed` (код 412), возвращать `false, nil` (OCC конфликт).
  - При успешной записи сохранять новый ETag в `meta.ETag`.
- Добавить статическую проверку: `var _ SnapshotStore = (*S3SnapshotStore)(nil)`.

#### [MODIFY] [engine_test.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine_test.go)
- Добавить тест `TestS3SnapshotStore` аналогично `TestPostgresSnapshotStore`.
- Тест будет пытаться подключиться к локальному S3/MinIO (используя переменные окружения `S3_ENDPOINT`, `S3_BUCKET`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`).
- Если окружение не настроено, тест пропускается (`t.Skip`).

---

## Verification Plan

### Automated Tests
- Запуск тестов для проверки компиляции и интеграции:
  ```bash
  go test -v ./...
  ```
