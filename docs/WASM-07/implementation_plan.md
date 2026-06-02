# Implementation Plan - Хранение снапшотов в SQLite и репликация через Litestream (WASM-07)

Этот план описывает шаги по переносу хранения снапшотов WASM из бинарных файлов на диске в базу данных SQLite, а также настройку репликации базы данных на S3 с помощью Litestream.

## Proposed Changes

Мы добавим чистый Go-драйвер SQLite (`modernc.org/sqlite`), расширим интерфейс `SnapshotStore` методом `Delete()`, реализуем `SqliteSnapshotStore`, переведем пример `examples/simple` на использование SQLite и подготовим Docker-окружение для репликации через Litestream.

---

### Core Module Upgrades

#### [MODIFY] [go.mod](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/go.mod)
* Добавить зависимость от `modernc.org/sqlite v1.21.2` (или актуальной совместимой версии).

#### [MODIFY] [engine.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine.go)
* Обновить интерфейс `SnapshotStore`, добавив метод `Delete(id string) error`.
* В структуру `FileSnapshotStore` добавить реализацию метода `Delete(id string) error`, который вызывает `os.Remove`.
* В методе `Execute` после успешного выполнения и удаления временного файла снапшота (в коде хостов) теперь будет вызываться `store.Delete(id)`.

#### [NEW] [sqlite_store.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/sqlite_store.go)
* Создать новую структуру `SqliteSnapshotStore`:
  ```go
  type SqliteSnapshotStore struct {
      db *sql.DB
  }
  ```
* Реализовать конструктор `NewSqliteSnapshotStore(dbPath string) (*SqliteSnapshotStore, error)`:
  * Открывать SQLite через `database/sql` с драйвером `"sqlite"`.
  * Включать WAL-режим: `PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000;`.
  * Создавать таблицу `CREATE TABLE IF NOT EXISTS snapshots (id TEXT PRIMARY KEY, snapshot BLOB, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)`.
* Реализовать методы `Save(id string, snapshot []byte) error`, `Load(id string) ([]byte, error)` и `Delete(id string) error`.
* Реализовать метод `Close() error` для корректного закрытия БД.

---

### Refactoring Examples

#### [MODIFY] [examples/simple/host/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/simple/host/main.go)
* Заменить `FileSnapshotStore` на `SqliteSnapshotStore`.
* Базу данных инициализировать в файле `snapshots.db`.
* Заменить прямой вызов `os.Remove("worker-instance-42.bin")` на `engine.store.Delete(instanceID)`.
* Закрывать соединение с БД при завершении работы хоста.

#### [MODIFY] [examples/camunda/host/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/camunda/host/main.go)
* Заменить вызов `os.Remove(snapshotPath)` на `engine.store.Delete(businessKey)`.

---

### Litestream & S3 Replication (Simple Example)

#### [NEW] [docker-compose.yml](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/simple/docker-compose.yml)
* Создать docker-compose файл для запуска примера в связке с Litestream и локальным MinIO (S3-совместимое хранилище) для репликации, аналогично структуре в `pocketstream`.
* Описать два сервиса:
  * `host-app`: наше Go-приложение с восстановлением через `litestream restore` перед запуском.
  * `litestream-sidecar`: демон репликации `litestream replicate`.
  * `rclone`: локальный S3-совместимый сервер на базе `rclone serve s3`.

#### [NEW] [Dockerfile.litestream](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/simple/Dockerfile.litestream)
* Подготовить Dockerfile для фоновой репликации базы `snapshots.db` на MinIO.

---

## Verification Plan

### Automated Tests
1. Проверить unit-тесты ядра движка (убедиться, что `FileSnapshotStore` по-прежнему работает корректно с новым методом `Delete`):
   ```bash
   make -C durable-wasm test
   ```
2. Синхронизировать зависимости в Go-воркспейсе:
   ```bash
   go work sync
   make tidy
   ```
3. Запустить пример `simple` локально в обычном режиме (без Docker) для проверки корректности SQLite-хранилища:
   ```bash
   make -C durable-wasm/examples/simple run
   ```

### Manual Verification
* Развернуть Docker-окружение примера `simple` и проверить, что база данных реплицируется в MinIO S3-бакет и восстанавливается при перезапусках.
