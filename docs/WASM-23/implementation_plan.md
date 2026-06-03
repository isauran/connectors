# WASM-23: Вынесение гостевого раннера в отдельный Go-модуль wasman/runner

Этот план описывает вынесение файлов `runner.go` и `runner_stub.go` из основного пакета `wasman` в отдельный изолированный Go-модуль `wasman/runner` с собственным файлом `go.mod` (пространство имён `runner`). Это полностью исключает влияние хост-зависимостей (таких как Wasmtime, AWS SDK и т.д.) при сборке гостевых воркеров с помощью TinyGo.

## User Review Required

Изменение является ломающим для импортов гостевого кода:
- Гостевые воркеры в примерах теперь должны импортировать `"github.com/nativebpm/connectors/wasman/runner"` вместо `"github.com/nativebpm/connectors/wasman"`.
- Обращения к раннеру изменятся с `wasman.NewWorkflow()` на `runner.NewWorkflow()`, `wasman.Call()` на `runner.Call()` и т.д.

Публичный API хоста (`wasman.NewEngine`, `wasman.Engine.Execute` и др.) никак не меняется.

## Open Questions

Нет.

## Proposed Changes

### Компонент `wasman` (Хост и Гость)

---

#### [NEW] [go.mod](file:///Users/user/github.com/nativebpm/connectors/wasman/runner/go.mod)
Создание нового файла `go.mod` для модуля `wasman/runner`.
- Имя модуля: `github.com/nativebpm/connectors/wasman/runner`.

#### [MODIFY] [go.work](file:///Users/user/github.com/nativebpm/connectors/go.work)
Регистрация нового пути `./wasman/runner` в Go-воркспейсе.

#### [MODIFY] [go.mod](file:///Users/user/github.com/nativebpm/connectors/wasman/go.mod)
Добавление зависимости от нового модуля `wasman/runner`:
- Добавить `require github.com/nativebpm/connectors/wasman/runner v0.0.1`.
- Добавить директиву `replace github.com/nativebpm/connectors/wasman/runner => ./runner` для локального разрешения зависимостей вне воркспейса.

#### [NEW] [runner.go](file:///Users/user/github.com/nativebpm/connectors/wasman/runner/runner.go)
Создание файла раннера в новом пакете `runner`. Пакет будет объявлен как `package runner`.

#### [NEW] [runner_stub.go](file:///Users/user/github.com/nativebpm/connectors/wasman/runner/runner_stub.go)
Создание заглушек раннера в новом пакете `runner`. Пакет будет объявлен как `package runner`.

### Примеры воркеров (Обновление импортов)

---

#### [MODIFY] [main.go](file:///Users/user/github.com/nativebpm/connectors/wasman/examples/s3-store/worker/main.go)
#### [MODIFY] [main.go](file:///Users/user/github.com/nativebpm/connectors/wasman/examples/camunda/worker/main.go)
#### [MODIFY] [main.go](file:///Users/user/github.com/nativebpm/connectors/wasman/examples/gotenberg-telegram/worker/main.go)
#### [MODIFY] [main.go](file:///Users/user/github.com/nativebpm/connectors/wasman/examples/process-csv/worker/main.go)
#### [MODIFY] [main.go](file:///Users/user/github.com/nativebpm/connectors/wasman/examples/temporal/worker/main.go)
#### [MODIFY] [main.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/examples/orchestration/worker/main.go)

В каждом из этих файлов изменить импорт `"github.com/nativebpm/connectors/wasman"` на `"github.com/nativebpm/connectors/wasman/runner"`, а также заменить вызовы функций/типов префикса `wasman.` на `runner.`.

## Verification Plan

### Automated Tests
1. Сборка всех WASM-воркеров в примерах:
   ```bash
   make -C wasman build-worker
   make -C bpmn/examples/orchestration build-worker
   ```
2. Запуск unit- и интеграционных тестов:
   ```bash
   make -C wasman test
   go test -v ./bpmn/...
   ```
