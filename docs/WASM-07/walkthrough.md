# Walkthrough - Хранение снапшотов в SQLite и репликация в S3 через Litestream (WASM-07)

В рамках задачи **WASM-07** мы полностью перевели хранение снапшотов движка Durable WASM из файлового хранилища в базу данных SQLite, добавили API удаления снимков, а также настроили автоматическое резервирование и репликацию SQLite базы на локальное S3-совместимое облачное хранилище **SeaweedFS** с помощью Litestream с политикой удержания (retention) истории на 5 дней и созданием полных снимков каждые 24 часа.

## Изменения

### 1. Ядро движка (durable-wasm)
* **[go.mod](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/go.mod)**:
  * Добавлена cgo-free зависимость `modernc.org/sqlite v1.21.2` для работы с SQLite на чистом Go.
* **[engine.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine.go)**:
  * В интерфейс `SnapshotStore` добавлен метод `Delete(id string) error`.
  * В структуре `FileSnapshotStore` реализован этот метод (удаляет файл с диска).
  * Всё логирование переведено с `fmt.Printf`/`fmt.Println` на стандартную библиотеку структурированного логирования `log/slog`.
* **[sqlite_store.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/sqlite_store.go)**:
  * Реализована структура `SqliteSnapshotStore` на базе `database/sql` и cgo-free sqlite драйвера.
  * База данных автоматически настраивается в режиме WAL (`PRAGMA journal_mode=WAL;`), что необходимо для корректной работы Litestream.
  * Реализованы методы `Save`, `Load`, `Delete` и `Close`.

### 2. Полный перевод всех примеров (Examples) на SQLite и slog
Все хост-приложения примеров переведены на использование `SqliteSnapshotStore` и логирование `log/slog`:
* **[examples/simple/host/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/simple/host/main.go)**
* **[examples/process-csv/host/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/process-csv/host/main.go)**
* **[examples/camunda/host/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/camunda/host/main.go)**
* **[examples/gotenberg-telegram/host/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/gotenberg-telegram/host/main.go)**
* **[examples/temporal/host/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/temporal/host/main.go)**

Каждое из этих приложений теперь:
* Инициализирует SQLite-базу с помощью `durable.NewSqliteSnapshotStore("snapshots.db")`.
* Проверяет и загружает снапшоты через базу SQLite.
* Логирует ход выполнения процесса через структурированные сообщения `slog.Info`, `slog.Warn`, `slog.Error`.
* Удаляет временные файлы/снимки из SQLite с помощью `store.Delete(id)` в конце работы.

* **[examples/simple/.gitignore](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/simple/.gitignore)**:
  * Добавлены правила игнорирования временных файлов SQLite (`*.db`, `*.db-wal`, `*.db-shm`), скомпилированного бинарника `host/host`, WASM-файла `worker/worker.wasm` и каталога локальных данных SeaweedFS `seaweed_data/`.

### 3. Docker-окружение для репликации в S3 через SeaweedFS
* **[examples/simple/docker-compose.yml](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/simple/docker-compose.yml)**:
  * Описывает службу `seaweedfs` (`chrislusf/seaweedfs:latest`), которая запускает S3-совместимый API на порту `8333` и Filer на порту `8888` с монтированием каталога `seaweed_data`.
  * Описывает службу `host-app`, запускающую собранное Go-приложение с Litestream-оберткой.
* **[examples/simple/host/Dockerfile](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/simple/host/Dockerfile)**:
  * Собирает статически связанный бинарник хоста.
  * Разворачивает Alpine-окружение с установленными бинарниками `litestream` и entrypoint-скриптом.
* **[examples/simple/litestream.yml](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/simple/litestream.yml)**:
  * Конфигурационный файл для репликации `snapshots.db` на S3-эндпоинт SeaweedFS `http://seaweedfs:8333` с интервалом синхронизации WAL-фреймов в `100ms`.
  * Настроена политика удержания `retention: 5d` и частота полных снимков `snapshot-interval: 24h` для обеспечения возможности восстановления на любой момент времени в течение последних 5 дней.
* **[examples/simple/entrypoint.sh](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/simple/entrypoint.sh)**:
  * Опрашивает порты SeaweedFS Filer (8888) и S3 API (8333), ожидая их готовности.
  * Автоматически создает бакет `snapshots-bucket` через POST-запрос в Filer, корректно обрабатывая ситуацию, если бакет уже существует (возвращает 409 Conflict).
  * Пытается восстановить SQLite базу из S3 перед стартом (`litestream restore`).
  * Запускает приложение через встроенный механизм репликации Litestream: `exec litestream replicate -exec "/app/host"`.

## Результаты тестирования

* Unit-тесты движка (`make test` в корне `durable-wasm`) переведены на использование in-memory базы данных SQLite (`SqliteSnapshotStore(":memory:")`). Тесты успешно выполняются без обращения к диску, подтверждая корректность логики snapshotting, восстановления памяти и обработки сбоев.
* Все примеры успешно компилируются: `go build` проходит для каждого хост-приложения.
* **Проверка репликации**: В процессе работы приложения в Docker-контейнере Litestream успешно инициализирует мониторы удаления L0 файлов (`retention=5d`) и создания снимков (`interval=24h`). Файлы изменений LTX и снимков загружаются в SeaweedFS S3-бакет.
* **Проверка восстановления**: При выполнении `docker compose down -v && docker compose up` (удаление локального volume `shared-data` с БД), Litestream при запуске автоматически определяет отставание локальной базы от реплики в SeaweedFS и скачивает последний snapshot и WAL-кадры, полностью восстанавливая состояние БД.
