# Implementation Plan - WASM-14: Remove Relational Databases (SQLite & Postgres)

Этот план описывает шаги по удалению поддержки реляционных СУБД (SQLite и PostgreSQL) из движка `durable-wasm`. В результате останется поддержка только двух хранилищ:
1. `S3SnapshotStore` — основное облачное хранилище с нативным OCC через ETag.
2. `FileSnapshotStore` — простое локальное файловое хранилище для локального тестирования, отладки и примеров.

## Преимущества
- **Упрощение архитектуры:** Отсутствие реляционных схем, транзакций и DDL-миграций.
- **Удаление зависимостей:** Устраняются зависимости от `modernc.org/sqlite` и `github.com/lib/pq`, что уменьшает размер собранного бинарного файла и время компиляции.
- **Удаление AtlasGo:** Больше не требуется настраивать и запускать инструмент AtlasGo для миграций.
- **Простота локального запуска:** Примеры и тесты будут использовать файловую систему или in-memory хранилище, не требуя внешних баз данных или контейнеров.

## User Review Required

> [!WARNING]
> Это ломающее изменение (breaking change): все интеграции, которые рассчитывали на работу с SQLite или Postgres, должны будут перейти на `S3SnapshotStore` или `FileSnapshotStore`.

---

## Proposed Changes

### [Component: durable-wasm Core]

#### [DELETE] [sqlite_store.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/sqlite_store.go)
- Полное удаление файла реализации SQLite провайдера.

#### [DELETE] [postgres_store.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/postgres_store.go)
- Полное удаление файла реализации PostgreSQL провайдера.

#### [DELETE] [schema.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/schema.go)
- Удаление вспомогательного файла экспорта схем (был создан ранее для тестов).

#### [DELETE] [atlas.hcl](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/atlas.hcl)
- Удаление конфигурационного файла AtlasGo.

#### [DELETE] [queries](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/queries)
- Полное удаление каталога SQL-запросов (`queries/sqlite/` и `queries/postgres/`).

#### [DELETE] [schema](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/schema)
- Полное удаление каталога SQL-схем.

---

### [Component: durable-wasm Testing]

#### [MODIFY] [engine_test.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine_test.go)
- Добавить реализацию `inMemorySnapshotStore` (на базе `map` с потокобезопасностью через `sync.RWMutex`) внутри тестового файла для изоляции юнит-тестов от диска.
- Заменить все вызовы `NewSqliteSnapshotStore` на инициализацию `inMemorySnapshotStore`.
- Удалить функции `initSqliteStore` и `initPostgresStore`, а также тест `TestPostgresSnapshotStore`.

#### [MODIFY] [engine_bench_test.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine_bench_test.go)
- Заменить использование `NewSqliteSnapshotStore` на `inMemorySnapshotStore` для чистых in-memory бенчмарков.

---

### [Component: Examples]

#### [MODIFY] [examples/camunda/host/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/camunda/host/main.go)
- Заменить использование `NewSqliteSnapshotStore` на `&durable.FileSnapshotStore{Dir: "snapshots"}`.
- Создавать директорию `snapshots` через `os.MkdirAll` при старте.

#### [MODIFY] [examples/camunda/host/main_test.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/camunda/host/main_test.go)
- Удалить встраивание (`//go:embed`) SQL-схем.
- Заменить инициализацию хранилища на `FileSnapshotStore` (или `inMemorySnapshotStore`).

#### [MODIFY] [examples/temporal/host/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/temporal/host/main.go)
- Удалить блок выбора хранилища (Postgres/SQLite). Оставить только `FileSnapshotStore`.
- Удалить встраивание SQL-схем.

#### [MODIFY] [examples/process-csv/host/main_test.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/process-csv/host/main_test.go)
- Удалить встраивание SQL-схем и заменить инициализацию на `FileSnapshotStore`.

#### [MODIFY] [examples/gotenberg-telegram/host/main_test.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/gotenberg-telegram/host/main_test.go)
- Удалить встраивание SQL-схем и заменить инициализацию на `FileSnapshotStore`.

#### [MODIFY] [examples/durable-s3/host/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/durable-s3/host/main.go)
- Заменить использование SQLite на `FileSnapshotStore`.

---

### [Component: Project Infrastructure]

#### [MODIFY] [Makefile](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/Makefile)
- Удалить цели, связанные с генерацией схем и запуском миграций AtlasGo.

#### [MODIFY] [README.md](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/README.md) & [README.ru.md](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/README.ru.md)
- Убрать разделы про настройку БД, DDL, AtlasGo.
- Обновить примеры кода с использованием `FileSnapshotStore` и `S3SnapshotStore`.

#### [MODIFY] [go.mod](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/go.mod)
- Запустить `go mod tidy` для автоматического удаления зависимостей `modernc.org/sqlite` и `github.com/lib/pq`.

---

## Verification Plan

### Automated Tests
- Запуск тестов всего репозитория `durable-wasm`:
  ```bash
  go test -v ./...
  ```
- Выполнение `make tidy` для проверки чистоты зависимостей.
