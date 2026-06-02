---
task: WASM-13
status: In Progress
summary: Реализация S3SnapshotStore для распределенного хранения снапшотов с OCC оптимистичной блокировкой
---

# WASM-13: Интеграция S3-совместимого хранилища снимков памяти

## Описание задачи
Реализовать новый провайдер хранения снимков памяти `S3SnapshotStore`, поддерживающий AWS S3 и любые S3-совместимые объектные хранилища (MinIO, Ceph, Cloudflare R2 и др.). Это позволит распределенным Stateless-нодам горизонтально масштабировать выполнение Durable WASM без жесткой привязки к локальной SQLite или общей СУБД Postgres.

## Требования
1. **Реализация интерфейса `SnapshotStore`**:
   Создать `S3SnapshotStore` в новом файле `s3_store.go`, реализующий все методы интерфейса `SnapshotStore`.
2. **Структура путей в S3**:
   Хранить ресурсы инстансов по следующим ключам:
   - Полные снимки: `instances/{instance_id}/snapshot.bin`
   - Дельты памяти: `instances/{instance_id}/deltas.json`
   - Лог операций (Oplog): `instances/{instance_id}/oplog.json`
   - Метаданные выполнения (OCC): `instances/{instance_id}/meta.json`
   - Модули WASM: `wasm_modules/{wasm_hash}.wasm`
3. **Оптимистичная блокировка (OCC)**:
   - Добавить поле `ETag string` в структуру `InstanceMeta` (в `engine.go`), чтобы хранить ETag заголовки.
   - Метод `SaveMetadata` должен использовать условную запись S3 (заголовок `If-Match` с текущим ETag для обновления и `If-None-Match: *` для новой записи) для надежного предотвращения параллельного выполнения.
4. **Интеграционное тестирование**:
   - Написать интеграционный тест `TestS3SnapshotStore` в `engine_test.go`, который запускается локально с MinIO/S3 (или пропускается, если credentials не заданы).
