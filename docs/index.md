# Индекс задач (Semantic Store)

| ID | Title | Status | Semantic Summary |
|:---|:---|:---|:---|
| CAM-02 | Полноценный пример с Sequin CDC воркером | Completed | Создан пример использования Sequin CDC воркера для обработки задач кредитного конвейера |
| CAM-03 | Outbox CDC пример без REST-запросов к Camunda | Completed | Создание примера с использованием триггера авто-блокировки БД и Sequin SQL Enrichment |
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
| WASM-15 | Разработка Go SDK для упрощения написания бизнес-логики | Completed | Создание пакета sdk для инкапсуляции WASM-импортов, ввода-вывода и автоматического управления шагами |
| WASM-16 | Рефакторинг и декомпозиция Durable WASM и ребрендинг модуля | Completed | Рефакторинг engine.go и переименование модуля durable-wasm в wasman |
| WASM-17 | Создание модуля bpmn для BPMN 2.0 и DMN движков | Completed | Разработка Go-модуля bpmn, использующего wasman в качестве Durable WASM SDK |
| WASM-18 | Поддержка асинхронных шагов ожидания (Wait State) | Completed | Добавление поддержки UserTask и ReceiveTask в BPMN-движке |
| WASM-19 | Полная поддержка элементов спецификации BPMN 2.0 | Completed | Реализация Boundary Events, Subprocesses и BusinessRuleTask с интеграцией DMN |
| WASM-20 | Реализация расширенной спецификации BPMN 2.0 | Completed | Добавление Inclusive Gateway, не-прерывающих событий, сигналов, компенсаций и MI |



