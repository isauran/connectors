# Walkthrough - WASM-22: Переименование файлов SDK в Runner API

В рамках задачи **WASM-22** выполнено переименование файлов `sdk.go` и `sdk_stub.go` в `runner.go` и `runner_stub.go` в соответствии с их ролью в качестве Runner API гостевого WASM-модуля. Также очищены ссылки на "SDK" в коде и логах.

## Список изменений

1. **Переименование и модификация файлов:**
   - Переименован [sdk.go](file:///Users/user/github.com/nativebpm/connectors/wasman/sdk.go) -> [runner.go](file:///Users/user/github.com/nativebpm/connectors/wasman/runner.go). Логирование ошибок шагов обновлено с префиксом `[WASMAN ERROR]` вместо `[SDK ERROR]`.
   - Переименован [sdk_stub.go](file:///Users/user/github.com/nativebpm/connectors/wasman/sdk_stub.go) -> [runner_stub.go](file:///Users/user/github.com/nativebpm/connectors/wasman/runner_stub.go). Текст ошибки о сборке без флага `wasm` обновлен для соответствия новому имени (указано "wasman guest runner").
   - Удалены устаревшие файлы `sdk.go` и `sdk_stub.go`.

2. **Обновление документации:**
   - В [walkthrough.md](file:///Users/user/github.com/nativebpm/connectors/docs/WASM-15/walkthrough.md) для WASM-15 ссылки обновлены с `sdk.go` / `sdk_stub.go` на `runner.go` / `runner_stub.go` с исправленным путем к модулю `wasman`.
   - В [task.md](file:///Users/user/github.com/nativebpm/connectors/docs/WASM-16/task.md) для WASM-16 ссылки также обновлены.

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
