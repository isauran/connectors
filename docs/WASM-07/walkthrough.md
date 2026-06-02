# Walkthrough - Хранение снапшотов в SQLite и репликация в S3 через Litestream (WASM-07)

В рамках задачи **WASM-07** мы заменили файловое хранилище снапшотов на транзакционную базу данных SQLite, добавили API удаления снимков, а также настроили автоматическое резервирование и репликацию SQLite базы на S3-совместимое хранилище (rclone serve s3) с помощью Litestream.

## Изменения

### 1. Ядро движка (durable-wasm)
* **[go.mod](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/go.mod)**:
  * Добавлена cgo-free зависимость `modernc.org/sqlite v1.21.2` для работы с SQLite на чистом Go.
* **[engine.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine.go)**:
  * В интерфейс `SnapshotStore` добавлен метод `Delete(id string) error`.
  * В структуре `FileSnapshotStore` реализован этот метод (удаляет файл с диска).
* **[sqlite_store.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/sqlite_store.go) [NEW]**:
  * Реализована структура `SqliteSnapshotStore` на базе `database/sql` и cgo-free sqlite драйвера.
  * База данных автоматически настраивается в режиме WAL (`PRAGMA journal_mode=WAL;`), что необходимо для корректной работы Litestream.
  * Реализованы методы `Save`, `Load`, `Delete` и `Close`.

### 2. Примеры использования (Examples)
* **[examples/simple/host/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/simple/host/main.go)**:
  * Полностью переведен на `SqliteSnapshotStore` и файл базы `snapshots.db`.
  * Использует `engine.store.Delete` для очистки сессии при успешном завершении.
* **Другие примеры**: переведены на использование нового метода `Delete` интерфейса `SnapshotStore` вместо ручного вызова `os.Remove`.

### 3. Docker-окружение для репликации в S3 через rclone
* **[examples/simple/docker-compose.yml](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/simple/docker-compose.yml) [NEW]**:
  * Описывает службу `rclone` (`rclone/rclone:latest`), которая запускает S3-совместимый REST API поверх локального каталога `rclone_data` на порту `8334`.
  * Описывает службу `host-app`, запускающую собранное Go-приложение с Litestream-оберткой.
* **[examples/simple/host/Dockerfile](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/simple/host/Dockerfile)**:
  * Собирает статически связанный бинарник хоста.
  * Разворачивает Alpine-окружение с установленными бинарниками `litestream` и entrypoint-скриптом.
* **[examples/simple/litestream.yml](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/simple/litestream.yml) [NEW]**:
  * Configuration-файл для репликации `snapshots.db` на S3-эндпоинт `http://rclone:8334` с интервалом синхронизации WAL-фреймов в `100ms`.
* **[examples/simple/entrypoint.sh](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/simple/entrypoint.sh) [NEW]**:
  * Пытается восстановить SQLite базу из S3 перед стартом (`litestream restore`).
  * Запускает приложение через встроенный механизм репликации Litestream: `exec litestream replicate -exec "/app/host"`.

## Результаты тестирования

* Unit-тесты движка (`make test` в корне `durable-wasm`) успешно выполняются.
* Локальный запуск простого примера (`make -C durable-wasm/examples/simple run`) успешно создает базу данных SQLite, настраивает WAL, производит запись и чтение снапшота в базу, а также завершает процесс без ошибок.
