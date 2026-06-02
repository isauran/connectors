# Walkthrough - Production-ready Improvements for Durable WASM

Успешно реализованы все улучшения стабильности для Durable WASM движка с полной поддержкой безопасной смены версий модулей и миграцией схем баз данных под управление **Atlas Go (https://atlasgo.io/)**.

## Изменения

1. **Реестр WASM-модулей (Multi-Version Registry)**:
   - В интерфейс `SnapshotStore` добавлены методы `SaveWasm(hash string, wasmBytes []byte)` и `LoadWasm(hash string) ([]byte, error)`.
   - Реализована поддержка сохранения и загрузки бинарников WASM в `FileSnapshotStore` (файлы `wasm_<hash>.wasm`), `SqliteSnapshotStore` (таблица `wasm_modules`) и `PostgresSnapshotStore` (таблица `wasm_modules`).
   - В `NewEngine` бинарный код WASM модуля автоматически сохраняется в реестр.
   - В `Execute` при несовпадении текущего хэша с требуемым (из метаданных инстанса) нужная версия WASM автоматически загружается из реестра и компилируется на лету для возобновления работы инстанса. Это позволяет бэкенду проснуться через 6+ месяцев и продолжить выполнение на оригинальной версии кода без конфликтов и падений.

2. **Миграции баз данных под управлением Atlas Go**:
   - Перенесены все схемы баз данных в стандартные каталоги версионных миграций Atlas:
     - [`durable-wasm/migrations/sqlite/20260602191000_init_sqlite.sql`](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/migrations/sqlite/20260602191000_init_sqlite.sql) — начальная миграция схем таблиц для SQLite.
     - [`durable-wasm/migrations/postgres/20260602191000_init_postgres.sql`](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/migrations/postgres/20260602191000_init_postgres.sql) — начальная миграция схем таблиц для PostgreSQL.
   - Сгенерированы валидные файлы контрольных сумм `atlas.sum` для обоих типов СУБД с помощью Docker-образа `arigaio/atlas:latest migrate hash`.
   - Настроен `//go:embed` в [`sqlite_store.go`](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/sqlite_store.go) и [`postgres_store.go`](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/postgres_store.go), указывающий на первый файл миграций Atlas. Это позволило использовать единый декларативный источник схем для in-memory тестов и устранить дублирование SQL DDL-запросов.

3. **Корректировка Snapshot-стратегии**:
   - На первом чекпоинте (`Version == 1`) теперь всегда пишется полный снапшот памяти в `store.Save`, что обеспечивает базовый снимок памяти.
   - При последующих чекпоинтах (`Version > 1`) пишутся дельты (`SaveDeltas`), а периодические полные чекпоинты (каждые 5 шагов) делают Truncation Oplog и Deltas.

4. **Верификация тестами**:
   - Обновлен тест `TestDirtyPageAndOplog` для соответствия стратегии создания полного снапшота на версии 1 и проверки дельт на версии 2.
   - Реализован новый комплексный тест `TestMultiVersionWasmExecution`, который проверяет корректную загрузку и replay старой версии WASM-модуля при смене версии движка.
   - Исправлен тест `TestWasmModuleHashMismatch` для симуляции отсутствия модуля в реестре.

## Результаты верификации

- Все тесты (включая примеры Gotenberg, CSV, Temporal и новые тесты версионирования) успешно выполняются:
  ```
  PASS
  ok  	github.com/nativebpm/connectors/durable-wasm	0.774s
  ```
- Статический анализ `go vet ./durable-wasm/...` пройден успешно без предупреждений.
