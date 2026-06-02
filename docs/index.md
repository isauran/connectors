# Индекс задач (Semantic Store)

| ID | Title | Status | Semantic Summary |
|:---|:---|:---|:---|
| SEC-01 | Безопасность интеграции с Camunda и Temporal | Completed | Реализовано шифрование Payload (Zero-Trust) и Codec Server для админки Temporal |
| TEMP-02 | Пример демонстрации Activity Heartbeats в Temporal | Completed | Создан пример для демонстрации Activity Heartbeats в Go SDK с восстановлением прогресса |
| TEL-03 | Драйвер Telegram Bot API для httpstream | Completed | Создание коннектора для Telegram Bot API с поддержкой потоковой загрузки медиа |
| WASM-04 | Durable Execution Engine на базе WASM | Completed | Разработка MVP WASM-движка с snapshotting и O(1) RAM. Объединено в единый переиспользуемый Go-модуль. |
| WASM-05 | Интеграция httpstream в durable-wasm | Completed | Переход на Fluent API httpstream для скачивания и загрузки данных в WASM-движке. |
| WASM-06 | Реструктуризация модуля durable-wasm | Completed | Перенос демо-файлов и песочницы в каталог examples, очистка корня модуля. |
| WASM-07 | Хранение снапшотов в SQLite и репликация в S3 | Completed | Хранение снимков WASM в SQLite и настройка репликации через Litestream. |
| WASM-08 | Исследование архитектуры Golem Cloud | Completed | Проведен детальный анализ архитектуры Golem Cloud (Oplog, Delta Snapshots) и ограничений SQLite. |
| WASM-09 | Проектирование целевой масштабируемой архитектуры | Completed | Спроектирована распределенная архитектура Durable WASM (Stateless хосты, CockroachDB, Oplog, Delta Snapshots). |
| WASM-10 | Обновление golangci-lint до v2 и статический анализ | Completed | Обновление golangci-lint до версии v2, адаптация конфигурации и запуск статического анализа. |
| WASM-11 | Продакшн-ready улучшения для Durable WASM | Completed | Миграция версий WASM, OCC, Truncation Oplog, детерминированное время и NaN canonicalization |
| TEMP-12 | Пример Sequin Outbox CDC к Temporal | In Progress | Обучающий пример интеграции Sequin Outbox CDC и Temporal на Go с HTTP-вебхуками и воркером. |
| WASM-12 | Повышение стабильности и стресс-тестирование Durable WASM | Completed | Оптимизация утечек ресурсов (TCP сокетов, cgo ресурсов Wasmtime) и добавление стресс-тестов в durable-wasm. |
| WASM-13 | Интеграция S3-совместимого хранилища снимков памяти | Completed | Реализация S3SnapshotStore для распределенного хранения снапшотов с OCC оптимистичной блокировкой. |
| WASM-14 | Удаление поддержки реляционных БД (SQLite/Postgres) | Completed | Удаление DDL-миграций, SQLite/Postgres и переход на S3 и File Snapshot Store |
