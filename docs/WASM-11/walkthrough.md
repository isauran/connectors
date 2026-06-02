# Walkthrough - Production-ready Improvements for Durable WASM

Успешно реализованы все улучшения стабильности для Durable WASM движка с полной поддержкой безопасной смены версий модулей и рефакторингом схемы БД с использованием `go:embed`.

## Изменения

1. **Реестр WASM-модулей (Multi-Version Registry)**:
   - В интерфейс `SnapshotStore` добавлены методы `SaveWasm(hash string, wasmBytes []byte)` и `LoadWasm(hash string) ([]byte, error)`.
   - Реализована поддержка сохранения и загрузки бинарников WASM в `FileSnapshotStore` (файлы `wasm_<hash>.wasm`), `SqliteSnapshotStore` (таблица `wasm_modules`) и `PostgresSnapshotStore` (таблица `wasm_modules`).
   - В `NewEngine` бинарный код WASM модуля автоматически сохраняется в реестр.
   - В `Execute` при несовпадении текущего хэша с требуемым (из метаданных инстанса) нужная версия WASM автоматически загружается из реестра и компилируется на лету для возобновления работы инстанса. Это позволяет бэкенду проснуться через 6+ месяцев и продолжить выполнение на оригинальной версии кода без конфликтов и падений.

2. **Внедрение схем базы данных через `go:embed`**:
   - Вынесены все SQL-запросы инициализации таблиц DDL из Go-кода в отдельные файлы:
     - [`durable-wasm/schema/sqlite.sql`](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/schema/sqlite.sql)
     - [`durable-wasm/schema/postgres.sql`](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/schema/postgres.sql)
   - В [`sqlite_store.go`](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/sqlite_store.go) и [`postgres_store.go`](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/postgres_store.go) добавлено использование директивы `//go:embed` для автоматического встраивания этих схем в скомпилированный Go-бинарник.
   - Это значительно упростило функции `NewSqliteSnapshotStore` и `NewPostgresSnapshotStore`, удалив более 100 строк жестко зашитого SQL из Go-файлов и сделав архитектуру чище и поддерживаемее.

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
