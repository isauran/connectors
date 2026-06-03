# Walkthrough - WASM-23: Вынесение гостевого раннера в отдельный Go-модуль

В рамках задачи **WASM-23** файлы гостевого WASM-раннера (`runner.go` и `runner_stub.go`) были успешно перенесены в новый Go-модуль `wasman/runner` с собственным файлом `go.mod`. Это полностью изолировало код, компилируемый под TinyGo, от любых хост-зависимостей пакета `wasman`.

## Список изменений

1. **Создание нового Go-модуля:**
   - Создан файл [go.mod](file:///Users/user/github.com/nativebpm/connectors/wasman/runner/go.mod) в директории `connectors/wasman/runner`.
   - Новый путь `./wasman/runner` зарегистрирован в конфигурации Go-воркспейса [go.work](file:///Users/user/github.com/nativebpm/connectors/go.work).
   - В [go.mod](file:///Users/user/github.com/nativebpm/connectors/wasman/go.mod) хост-модуля добавлена зависимость `github.com/nativebpm/connectors/wasman/runner` с директивой `replace`.

2. **Перенос и переименование файлов рантайма:**
   - Перенесены файлы:
     - `connectors/wasman/runner.go` -> [runner/runner.go](file:///Users/user/github.com/nativebpm/connectors/wasman/runner/runner.go)
     - `connectors/wasman/runner_stub.go` -> [runner/runner_stub.go](file:///Users/user/github.com/nativebpm/connectors/wasman/runner/runner_stub.go)
   - В обоих файлах пакет изменён с `wasman` на `runner`.
   - Удалены старые файлы из корня `wasman/`.

3. **Обновление примеров воркеров:**
   - Во всех 6 примерах воркеров изменён импорт с `"github.com/nativebpm/connectors/wasman"` на `"github.com/nativebpm/connectors/wasman/runner"`, а обращения переписаны с `wasman.` на `runner.`:
     - [examples/s3-store/worker/main.go](file:///Users/user/github.com/nativebpm/connectors/wasman/examples/s3-store/worker/main.go)
     - [examples/camunda/worker/main.go](file:///Users/user/github.com/nativebpm/connectors/wasman/examples/camunda/worker/main.go)
     - [examples/gotenberg-telegram/worker/main.go](file:///Users/user/github.com/nativebpm/connectors/wasman/examples/gotenberg-telegram/worker/main.go)
     - [examples/process-csv/worker/main.go](file:///Users/user/github.com/nativebpm/connectors/wasman/examples/process-csv/worker/main.go)
     - [examples/temporal/worker/main.go](file:///Users/user/github.com/nativebpm/connectors/wasman/examples/temporal/worker/main.go)
     - [bpmn/examples/orchestration/worker/main.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/examples/orchestration/worker/main.go)

## Результаты тестирования

1. Успешная сборка всех WASM-воркеров TinyGo в примерах `wasman`:
   ```bash
   make -C wasman build-worker
   ```
2. Успешный запуск unit-тестов модуля `wasman`:
   ```bash
   go test -v .
   ```
3. Успешная сборка WASM-воркера для примера в модуле `bpmn`:
   ```bash
   make -C bpmn/examples/orchestration build-worker
   ```
4. Успешный запуск всех unit/интеграционных тестов `bpmn` модуля:
   ```bash
   go test -v ./...
   ```
