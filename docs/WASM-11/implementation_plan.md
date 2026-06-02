# Implementation Plan - Production-ready Improvements for Durable WASM (Multi-Version Execution)

Внедрение 5 улучшений стабильности для Durable WASM движка с особым фокусом на надежную смену версий WASM-модулей (Multi-Version Execution Registry), предотвращение Split-Brain (OCC), детерминизм времени, периодическую очистку Oplog (Truncation) и канонизацию float/NaN.

## User Review Required

> [!IMPORTANT]
> Для поддержки сценария, когда бэкенд засыпает на длительный срок (например, 6 месяцев), а после обновления просыпается с несовместимой версией WASM-модуля, мы внедряем **Реестр WASM-модулей (Multi-Version Registry)**.
> При инициализации `NewEngine` бинарный код текущего WASM модуля автоматически сохраняется в `SnapshotStore`. При восстановлении инстанса, если его хэш отличается от текущего хэша движка, нужная версия WASM автоматически загружается из базы данных/файлов и компилируется на лету для продолжения выполнения.

> [!WARNING]
> Изменение интерфейса `SnapshotStore` является ломающим изменением (breaking change) для внешних реализаций. Мы обновим все три внутренние реализации в репозитории: `FileSnapshotStore`, `SqliteSnapshotStore` и `PostgresSnapshotStore`.

## Proposed Changes

### durable-wasm

#### [MODIFY] [engine.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine.go)

- Расширить интерфейс `SnapshotStore` методами для хранения WASM бинарников:
  ```go
  type SnapshotStore interface {
      // ... existing methods ...

      // WASM Registry for Multi-Version Support
      SaveWasm(hash string, wasmBytes []byte) error
      LoadWasm(hash string) ([]byte, error)
  }
  ```
- В `FileSnapshotStore` реализовать методы `SaveWasm` и `LoadWasm` (сохранение в файлы `wasm_<hash>.wasm` в целевой директории).
- В `NewEngine`:
  - Рассчитывать SHA256 хэш WASM модуля.
  - Сохранять модуль в реестр через `store.SaveWasm(wasmHash, wasmBytes)`.
- В `Execute`:
  - Проверять совместимость версии модуля. Если `meta.WasmHash != e.wasmHash`, динамически загружать историческую версию из реестра через `e.store.LoadWasm(meta.WasmHash)`, компилировать и использовать её для создания инстанса.
  - Исправить условие полного снапшота: делать полный снапшот на первом чекпоинте (`Version == 1`) или при `Version > 1 && Version % 5 == 0`.
    ```go
    isFullSnapshot := session.meta.Version == 1 || (session.meta.Version > 1 && session.meta.Version%5 == 0)
    ```

#### [MODIFY] [sqlite_store.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/sqlite_store.go)

- Добавить таблицу `wasm_modules`:
  ```sql
  CREATE TABLE IF NOT EXISTS wasm_modules (
      hash TEXT PRIMARY KEY,
      wasm_bytes BLOB NOT NULL,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );
  ```
- Реализовать методы `SaveWasm` и `LoadWasm` для работы с таблицей `wasm_modules`.

#### [MODIFY] [postgres_store.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/postgres_store.go)

- Добавить аналогичную таблицу и реализовать методы `SaveWasm` и `LoadWasm` для Postgres.

#### [MODIFY] [engine_test.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine_test.go)

- Добавить тест `TestMultiVersionWasmExecution`, проверяющий автоматическую загрузку и компиляцию старой версии WASM модуля при смене версии движка.
- Обновить `TestDirtyPageAndOplog`:
  - На первом шаге проверять наличие полного снапшота (`store.Load`).
  - На втором шаге проверять наличие дельт (`store.LoadDeltas`).
- Обновить существующие тесты для совместимости с новой燬хеме.

## Verification Plan

### Automated Tests
- Запуск тестов: `go test -v ./durable-wasm/...`
- Проверка линтера: `golangci-lint run ./durable-wasm/...`
