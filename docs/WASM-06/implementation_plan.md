# Implementation Plan - Реструктуризация durable-wasm модуля

Этот план описывает шаги по реструктуризации модуля `durable-wasm` для выноса демонстрационного кода (песочницы) в папку `examples/simple` и очистки структуры публичной библиотеки.

## Proposed Changes

Мы перенесем ядро движка прямо в корень `durable-wasm/`, удалим промежуточные файлы и папки `host/` и `worker/` из корня модуля, а демонстрационную песочницу перенесем в новый пример `durable-wasm/examples/simple/`.

---

### Core Library Restructuring

#### [NEW] [engine.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine.go)
* Перенести файл `host/durable/engine.go` в корень `durable-wasm/engine.go` с сохранением пакета `durable`.

#### [NEW] [engine_test.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine_test.go)
* Перенести файл `host/durable/engine_test.go` в корень `durable-wasm/engine_test.go` с сохранением пакета `durable`.
* Скорректировать относительный путь к тестовому WASM-файлу в `engine_test.go`:
  Заменить `filepath.Join("..", "..", "worker", "worker.wasm")` на `filepath.Join("examples", "simple", "worker", "worker.wasm")`.

#### [DELETE] [durable.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/durable.go)
* Удалить обертку-реэкспорт, так как типы теперь будут доступны напрямую из пакета `durable` в корне модуля.

#### [DELETE] [host/durable](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/host/durable)
* Удалить старую директорию ядра.

---

### Moving Demo Playground to Examples

#### [NEW] [simple example](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/simple)
Создать новую директорию примера `durable-wasm/examples/simple` и перенести туда демо-песочницу:
* Перенести `host/main.go` в `examples/simple/host/main.go`.
* Перенести `host/Dockerfile` в `examples/simple/host/Dockerfile`.
* Перенести `worker/main.go` в `examples/simple/worker/main.go`.
* Создать `examples/simple/Makefile` для сборки и запуска:
  ```makefile
  .PHONY: build run clean
  build:
  	tinygo build -o worker/worker.wasm -target=wasi worker/main.go
  run: build
  	cd host && go run main.go
  clean:
  	rm -f worker/worker.wasm
  ```

#### [DELETE] [host/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/host/main.go)
* Удалить демо-хост из корня.

#### [DELETE] [host/Dockerfile](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/host/Dockerfile)
* Удалить Dockerfile из корня.

#### [DELETE] [worker/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/worker/main.go)
* Удалить демо-воркер из корня.

---

### Root Configuration and Workspace

#### [MODIFY] [Makefile](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/Makefile)
* Скорректировать цели в `durable-wasm/Makefile`:
  * Заменить пути сборки воркера песочницы на `examples/simple/worker/main.go`.
  * Изменить пути запуска тестов на локальный пакет в корне.

#### [MODIFY] [go.work](file:///Users/user/github.com/nativebpm/connectors/go.work)
* Убрать запись `./durable-wasm/host` и `./durable-wasm/worker` (если они там остались), убедиться, что есть `./durable-wasm`.
* Добавить `./durable-wasm/examples/simple/host` и `./durable-wasm/examples/simple/worker`.

---

## Verification Plan

### Automated Tests
1. Синхронизировать зависимости:
   ```bash
   go work sync
   make tidy
   ```
2. Запустить unit-тесты ядра движка (теперь в корне `durable-wasm`):
   ```bash
   make -C durable-wasm test
   ```
3. Запустить перенесенный playground-пример:
   ```bash
   make -C durable-wasm/examples/simple run
   ```
4. Запустить остальные примеры (Camunda, Temporal) для проверки обратной совместимости.

### Manual Verification
* Убедиться в отсутствии лишних демо-файлов в корне `durable-wasm`.
