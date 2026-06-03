# WASM-22: Переименование файлов SDK в Runner API

Переименование файлов `sdk.go` и `sdk_stub.go` в пакете `wasman` в `runner.go` и `runner_stub.go`, так как они представляют собой API воркера-раннера для WASM, а не полноценный SDK. Также очищаются все упоминания "SDK" в коде этих файлов.

## User Review Required

Никаких ломающих изменений в публичном Go API пакета `wasman` не планируется. Все функции (`NewWorkflow`, `Run`, `CallAPI`, `Checkpoint` и др.) и типы (`Workflow`, `APICall`) останутся без изменений, поэтому пользовательский код в воркерах не сломается. Изменяются только имена самих исходных файлов и сообщения об ошибках / префиксы логирования (например, `[SDK ERROR]` -> `[WASMAN ERROR]`).

## Open Questions

Нет.

## Proposed Changes

### Компонент `wasman` (Durable WASM guest API)

---

#### [DELETE] [sdk.go](file:///Users/user/github.com/nativebpm/connectors/wasman/sdk.go)
Удаление старого файла с реализацией для WASM.

#### [NEW] [runner.go](file:///Users/user/github.com/nativebpm/connectors/wasman/runner.go)
Создание нового файла на замене `sdk.go`.
- Переименовать префикс логирования с `[SDK ERROR]` на `[WASMAN ERROR]`.

#### [DELETE] [sdk_stub.go](file:///Users/user/github.com/nativebpm/connectors/wasman/sdk_stub.go)
Удаление старого файла с заглушками для хоста.

#### [NEW] [runner_stub.go](file:///Users/user/github.com/nativebpm/connectors/wasman/runner_stub.go)
Создание нового файла на замене `sdk_stub.go`.
- Переименовать ошибку `wasman sdk is only supported...` на `wasman guest runner is only supported...`.

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
