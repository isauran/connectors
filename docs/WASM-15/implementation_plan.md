# Implementation Plan - Host Execution Fluent API

Этот план описывает внедрение Fluent API для настройки и запуска WASM-сессий на стороне хоста (в пакете `durable`), а также миграцию всех примеров, тестов и файлов документации (README) на новый синтаксис.

## User Review Required

> [!NOTE]
> Внедряется структура `Execution` с методами-билдерами `.WithServer()`, `.WithCrash()`, `.WithEntrypoint()` и методом запуска `.Run(ctx)`. Метод `engine.Execute(...)` сохраняется для обратной совместимости, но все внутренние примеры и тесты переводятся на новый Fluent-интерфейс, который делает код хоста чище и читаемее.

## Proposed Changes

### [Component: Durable Engine (Host)]

#### [MODIFY] [durable-wasm/engine.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine.go)
- Добавить структуру `Execution` с полями `engine`, `instanceID`, `entrypoint`, `serverAddr` и `shouldCrash`.
- Добавить метод `func (e *Engine) Session(instanceID string) *Execution`.
- Добавить методы-билдеры `WithServer`, `WithCrash`, `WithEntrypoint`.
- Добавить метод `func (ex *Execution) Run(ctx context.Context) (crashed bool, err error)`, который внутренне вызывает `engine.Execute(...)`.

---

### [Component: Examples & Tests]

Мы перепишем вызовы `engine.Execute` во всех примерах и тестах на новый Fluent-синтаксис:

#### [MODIFY] [examples/s3-store/host/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/s3-store/host/main.go)
#### [MODIFY] [examples/camunda/host/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/camunda/host/main.go)
#### [MODIFY] [examples/camunda/host/main_test.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/camunda/host/main_test.go)
#### [MODIFY] [examples/process-csv/host/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/process-csv/host/main.go)
#### [MODIFY] [examples/process-csv/host/main_test.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/process-csv/host/main_test.go)
#### [MODIFY] [examples/gotenberg-telegram/host/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/gotenberg-telegram/host/main.go)
#### [MODIFY] [examples/gotenberg-telegram/host/main_test.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/gotenberg-telegram/host/main_test.go)
#### [MODIFY] [examples/temporal/host/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/temporal/host/main.go)
#### [MODIFY] [engine_test.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine_test.go)

---

### [Component: Documentation]

#### [MODIFY] [README.md](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/README.md) и [README.ru.md](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/README.ru.md)
- Обновить примеры запуска хоста в README файлах, используя `engine.Session(...)` Fluent API.

---

## Verification Plan

### Automated Tests
1. Перекомпилировать все воркеры:
   ```bash
   cd durable-wasm
   make build-worker
   ```
2. Запустить все тесты хоста:
   ```bash
   go test -v ./...
   ```
