# Walkthrough - WASM-13: S3SnapshotStore Integration

В рамках задачи **WASM-13** была успешно реализована поддержка S3-совместимых объектных хранилищ (`S3SnapshotStore`) в качестве нового бэкенда `SnapshotStore` для движка `durable-wasm`.

## Список изменений

1. **Добавлено поле `ETag` в `InstanceMeta`**:
   - В [engine.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine.go#L33) структура `InstanceMeta` дополнена полем `ETag string` с тегом `json:"etag,omitempty"`. Это позволяет использовать Optimistic Concurrency Control (OCC) нативно средствами S3 API.
   - Изменение обратно-совместимо и не задевает SQLite/PostgreSQL хранилища.

2. **Создано S3-совместимое хранилище `S3SnapshotStore`**:
   - В новом файле [s3_store.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/s3_store.go) реализована структура `S3SnapshotStore`, имплементирующая интерфейс `SnapshotStore`.
   - Запись и чтение снимков памяти происходят по путям:
     - Снимки памяти: `instances/{id}/snapshot.bin`
     - Разница памяти (deltas): `instances/{id}/deltas.json`
     - Лог внешних вызовов (oplog): `instances/{id}/oplog.json`
     - Метаданные инстанса: `instances/{id}/meta.json`
     - Модули WASM: `wasm/{hash}.wasm`
   - Реализована блокировка OCC в методе `SaveMetadata` с использованием S3 заголовков:
     - `If-None-Match: *` для первой записи новой версии.
     - `If-Match: <ETag>` для обновления существующих метаданных.
     - При возникновении конфликта версий (`status 412 Precondition Failed`) метод возвращает `false, nil`.

3. **Интеграционное тестирование**:
   - В [engine_test.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine_test.go) добавлен метод `TestS3SnapshotStore`.
   - Тест проверяет сохранение снимков памяти, дельт, oplog, корректность проверки OCC (первичную вставку, обычное обновление и ошибки при устаревшем ETag) и сохранение WASM-модулей.
   - Тест автоматически пропускается, если не заданы переменные окружения S3.

## Результаты тестирования

Все тесты были успешно запущены и пройдены локально с использованием MinIO в Docker:

```
=== RUN   TestS3SnapshotStore
--- PASS: TestS3SnapshotStore (0.06s)
PASS
ok  	github.com/nativebpm/connectors/durable-wasm	0.603s
```
