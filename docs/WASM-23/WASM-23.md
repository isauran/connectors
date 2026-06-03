---
task: WASM-23
status: In Progress
summary: Вынесение гостевого WASM-раннера в отдельный Go-модуль wasman/runner с собственным go.mod
---

# WASM-23: Вынесение гостевого WASM-раннера в отдельный Go-модуль wasman/runner

Необходимо изолировать код гостевого рантайма (TinyGo-совместимого раннера) от хост-зависимостей (таких как Wasmtime, AWS SDK и т.д.), вынеся его в отдельный Go-модуль `wasman/runner` с собственным файлом `go.mod`.

## Требования:
1. Создать директорию `connectors/wasman/runner`.
2. Инициализировать в ней Go-модуль с файлом `go.mod` (имя модуля `github.com/nativebpm/connectors/wasman/runner`).
3. Зарегистрировать путь `./wasman/runner` в корневом файле `connectors/go.work`.
4. Перенести `connectors/wasman/runner.go` -> `connectors/wasman/runner/runner.go` и изменить `package wasman` на `package runner`.
5. Перенести `connectors/wasman/runner_stub.go` -> `connectors/wasman/runner/runner_stub.go` и изменить `package wasman` на `package runner`.
6. Добавить зависимость от нового модуля `github.com/nativebpm/connectors/wasman/runner` в `wasman/go.mod` (для компиляции примеров воркеров, находящихся внутри модуля `wasman`).
7. Обновить импорты и вызовы в примерах воркеров (заменить `wasman.NewWorkflow` на `runner.NewWorkflow` и т.д.):
   - `connectors/wasman/examples/s3-store/worker/main.go`
   - `connectors/wasman/examples/camunda/worker/main.go`
   - `connectors/wasman/examples/gotenberg-telegram/worker/main.go`
   - `connectors/wasman/examples/process-csv/worker/main.go`
   - `connectors/wasman/examples/temporal/worker/main.go`
   - `connectors/bpmn/examples/orchestration/worker/main.go`
8. Убедиться, что сборка воркеров через TinyGo и запуск тестов хостов проходят без ошибок.
