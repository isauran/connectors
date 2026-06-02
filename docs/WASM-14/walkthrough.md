# Walkthrough - WASM-14: Remove Relational Databases (SQLite & Postgres)

В рамках задачи **WASM-14** было выполнено полное удаление поддержки реляционных СУБД (SQLite и PostgreSQL) и связанных с ними DDL-схем, миграций и инструментов (AtlasGo). Решение полностью переведено на использование объектного хранилища `S3SnapshotStore` в качестве основного продакшн-хранилища и `FileSnapshotStore` в качестве легковесного локального бэкенда.

## Список изменений

1. **Удаление устаревших файлов и папок:**
   - Удалены провайдеры баз данных: `sqlite_store.go` и `postgres_store.go`.
   - Полностью удалены директории `queries/` (содержавшая SQL-запросы) и `schema/` (содержавшая DDL-схемы для SQLite/Postgres).
   - Удалены конфигурация `atlas.hcl` и все цели в `Makefile`, связанные с инспектированием и миграциями через инструмент **AtlasGo**.

2. **Вынесение `FileSnapshotStore` в отдельный файл:**
   - Реализация `FileSnapshotStore` и связанные методы вынесены из [engine.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine.go) в новый выделенный файл [fs_store.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/fs_store.go).
   - Это очистило ядро движка и улучшило модульность проекта.

3. **Оптимизация юнит-тестов ядра:**
   - В [engine_test.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine_test.go) добавлена структура `inMemorySnapshotStore` на базе Go `map` с потокобезопасностью через `sync.RWMutex`.
   - Для предотвращения гонок и ошибок `sigpanic` при изменении памяти WASM-машины, все методы сохранения (`Save`, `SaveDeltas`, `SaveOplog`, `SaveWasm`) выполняют глубокое копирование срезов байтов с помощью функции `copy()`.
   - Все юнит-тесты ядра переведены на использование `inMemorySnapshotStore`, что устранило зависимость от внешних файлов БД и ускорило тесты.
   - Удалены тесты для SQLite и Postgres.

4. **Очистка тестов бенчмарков:**
   - Бенчмарки в [engine_bench_test.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine_bench_test.go) переведены на `inMemorySnapshotStore`.

5. **Обновление и удаление примеров:**
   - Устаревший пример `examples/durable-s3` (демонстрировавший репликацию SQLite с помощью Litestream) полностью удален из репозитория.
   - Остальные практические примеры (`examples/camunda`, `examples/temporal`, `examples/process-csv`, `examples/gotenberg-telegram`) переведены на использование `FileSnapshotStore`.
   - Данные чекпоинтов локально пишутся в каталог `snapshots/` (или `snapshots_test/` в тестовых сценариях), который автоматически очищается и создается перед каждым запуском.
   - Удалены встраивания схем `//go:embed sqlite.sql/postgres.sql`.

6. **Очистка зависимостей:**
   - Из [go.mod](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/go.mod) исключены драйверы `modernc.org/sqlite` и `github.com/lib/pq`.

7. **Обновление документации:**
   - В [README.md](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/README.md) и [README.ru.md](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/README.ru.md) удалены инструкции по установке AtlasGo, запуску БД и накатыванию схем. Инструкции и примеры использования переписаны под `FileSnapshotStore` и `S3SnapshotStore`.

---

## Результаты тестирования

Все тесты ядра и примеров компилируются и успешно проходят локально:

```bash
$ go test ./...
ok  	github.com/nativebpm/connectors/durable-wasm	1.540s
ok  	github.com/nativebpm/connectors/durable-wasm/examples/camunda/host	3.800s
ok  	github.com/nativebpm/connectors/durable-wasm/examples/gotenberg-telegram/host	1.295s
ok  	github.com/nativebpm/connectors/durable-wasm/examples/process-csv/host	0.786s
?   	github.com/nativebpm/connectors/durable-wasm/examples/s3-store/host	[no test files]
ok  	github.com/nativebpm/connectors/durable-wasm/examples/temporal/host	2.645s
```
