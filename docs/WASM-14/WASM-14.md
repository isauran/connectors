---
task: WASM-14
status: Completed
summary: Удаление поддержки SQLite и PostgreSQL, переход на S3SnapshotStore и FileSnapshotStore
---

# WASM-14: Удаление поддержки реляционных БД (SQLite/Postgres)

## Описание задачи
Удалить поддержку реляционных СУБД (SQLite и PostgreSQL) из движка `durable-wasm`, избавиться от DDL-схем, генерации SQL и рантайм-зависимости от инструмента миграций (AtlasGo). Решение полностью переводится на использование объектного хранилища `S3SnapshotStore` в качестве основного продакшн-хранилища и `FileSnapshotStore` в качестве легковесного локального файлового бэкенда для тестирования и разработки.

## Результаты
1. Удалены файлы `sqlite_store.go`, `postgres_store.go`, SQL-запросы в папке `queries/` и схемы в `schema/`.
2. Удалены конфигурация AtlasGo и цели в `Makefile`.
3. Локальный `FileSnapshotStore` перенесен из `engine.go` в отдельный файл `fs_store.go`.
4. Реализован `inMemorySnapshotStore` в `engine_test.go` для быстрой и надежной работы юнит-тестов без записи на диск.
5. Примеры обновлены на использование `FileSnapshotStore` в каталогах `snapshots/`.
6. Очищены неиспользуемые зависимости драйверов БД из `go.mod`.
