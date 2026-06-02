# Walkthrough - Хранение снапшотов в SQLite и репликация в S3 через Litestream (WASM-07)

В рамках задачи **WASM-07** мы полностью перевели хранение снапшотов движка Durable WASM из файлового хранилища в базу данных SQLite, добавили API удаления снимков, а также настроили автоматическое резервирование и репликацию SQLite базы на локальное S3-совместимое облачное хранилище **SeaweedFS** с помощью Litestream с политикой удержания (retention) истории на 5 дней и созданием полных снимков каждые 24 часа.

Дополнительно вся структура примеров была переработана: все хосты и воркеры примеров объединены под единым Go-модулем `durable-wasm`, что позволило избавиться от 15 избыточных файлов `go.mod`/`go.sum` и очистить `go.work`.

## Изменения

### 1. Ядро движка (durable-wasm)
* **[go.mod](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/go.mod)**:
  * Добавлена cgo-free зависимость `modernc.org/sqlite v1.21.2` для работы с SQLite на чистом Go.
  * Добавлены зависимости, используемые в хостах примеров (`github.com/google/uuid`, `camunda`, `temporal` и `sequin-go`).
  * Для разрешения локальных зависимостей коннекторов в среде разработки используется [go.work](file:///Users/user/github.com/nativebpm/connectors/go.work), что позволило полностью отказаться от лишних директив `replace` в `go.mod`.
* **[engine.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine.go)**:
  * В интерфейс `SnapshotStore` добавлен метод `Delete(id string) error`.
  * В структуре `FileSnapshotStore` реализован этот метод (удаляет файл с диска).
  * Всё логирование переведено с `fmt.Printf`/`fmt.Println` на стандартную библиотеку структурированного логирования `log/slog`.
* **[sqlite_store.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/sqlite_store.go)**:
  * Реализована структура `SqliteSnapshotStore` на базе `database/sql` и cgo-free sqlite драйвера.
  * База данных автоматически настраивается в режиме WAL (`PRAGMA journal_mode=WAL;`), что необходимо для корректной работы Litestream.
  * Реализованы методы `Save`, `Load`, `Delete` и `Close`.

### 2. Единый модуль для примеров (Examples)
Все примеры (`camunda`, `durable-s3`, `gotenberg-telegram`, `process-csv`, `temporal`) переведены на использование единого родительского модуля `durable-wasm`:
* **Удалено 15 файлов `go.mod` и `go.sum`**: Все вложенные модули удалены. Хосты компилируются компилятором Go в контексте родительского модуля `durable-wasm`.
* **Теги сборки воркеров**: В файлы `worker/main.go` всех примеров добавлен тег сборки `//go:build wasm`. Это предотвращает попытки хост-компилятора скомпилировать воркеры при запуске `go test ./...` на хосте, устраняя ошибку `missing function body` для функций с `//go:wasmimport`.
* **[go.work](file:///Users/user/github.com/nativebpm/connectors/go.work)**: Очищен от путей к подмодулям примеров. В секции `use` остался только один корневой путь `./durable-wasm`.

### 3. Docker-окружение для репликации в S3 через SeaweedFS
* **[durable-s3/docker-compose.yml](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/durable-s3/docker-compose.yml)**:
  * Описывает службу `seaweedfs` (`chrislusf/seaweedfs:latest`), которая запускает S3-совместимый API на порту `8333` и Filer на порту `8888`.
  * Описывает службу `host-app`, запускающую собранное Go-приложение с Litestream-оберткой.
* **[durable-s3/host/Dockerfile](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/durable-s3/host/Dockerfile)**:
  * Собирает статически связанный бинарник хоста из единого модуля `durable-wasm`.
  * Разворачивает Alpine-окружение с установленными бинарниками `litestream` и entrypoint-скриптом.
* **[durable-s3/litestream.yml](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/durable-s3/litestream.yml)**:
  * Конфигурационный файл для репликации `snapshots.db` на S3-эндпоинт SeaweedFS `http://seaweedfs:8333` с интервалом синхронизации WAL-фреймов в `100ms`.
  * Настроена политика удержания `retention: 5d` и частота полных снимков `snapshot-interval: 24h` для обеспечения возможности восстановления на любой момент времени в течение последних 5 дней.
* **[durable-s3/entrypoint.sh](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/durable-s3/entrypoint.sh)**:
  * Опрашивает порты SeaweedFS Filer (8888) и S3 API (8333), ожидая их готовности.
  * Автоматически создает бакет `snapshots-bucket` через POST-запрос в Filer, корректно обрабатывая ситуацию, если бакет уже существует (возвращает 409 Conflict).
  * Пытается восстановить SQLite базу из S3 перед стартом (`litestream restore`).
  * Запускает асинхронный фоновый хелпер, который ждет инициализации файла базы данных, проверяет S3 и принудительно создает первый снимок через `litestream snapshot`, если в S3 еще нет резервных копий (чтобы не ждать 24 часа для первого снимка).
  * Запускает приложение через встроенный механизм репликации Litestream: `exec litestream replicate -exec "/app/host"`.

## Результаты тестирования

* Unit-тесты движка (`make test` в корне `durable-wasm`) переведены на использование in-memory базы данных SQLite (`SqliteSnapshotStore(":memory:")`). Тесты успешно выполняются без обращения к диску, подтверждая корректность логики snapshotting, восстановления памяти и обработки сбоев.
* Все примеры успешно компилируются: `go build` проходит для каждого хост-приложения.
* **Проверка репликации**: В процессе работы приложения в Docker-контейнере Litestream успешно инициализирует мониторы удаления L0 файлов (`retention=5d`) и создания снимков (`interval=24h`). Хелпер при старте контейнера успешно проверил реплику и загрузил базовый дамп. Файлы изменений LTX загружаются в SeaweedFS S3-бакет.
* **Проверка восстановления**: При выполнении `docker compose down -v && docker compose up` (удаление локального volume `shared-data` с БД), Litestream при запуске автоматически определяет отставание локальной базы от реплики в SeaweedFS и скачивает последний snapshot и WAL-кадры, полностью восстанавливая состояние БД.
