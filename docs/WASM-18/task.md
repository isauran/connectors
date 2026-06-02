# WASM-18 Task Checklist

- `[x]` Расширение моделей BPMN в [bpmn.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/bpmn.go)
  - `[x]` Добавить структуру `UserTask`
  - `[x]` Добавить структуру `ReceiveTask`
  - `[x]` Добавить структуру `IntermediateCatchEvent`
  - `[x]` Обновить структуру `Process`
  - `[x]` Обновить функцию `indexProcess`
- `[x]` Поддержка Wait State в движке в [engine.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/engine.go)
  - `[x]` Добавить поле `WaitingTokens` в `ProcessInstance`
  - `[x]` Обновить метод `Step()` для обработки `UserTask`, `ReceiveTask`, `IntermediateCatchEvent`
  - `[x]` Добавить метод `CompleteTask(...)`
- `[x]` Написание тестов в [engine_test.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/engine_test.go)
  - `[x]` Добавить тест с шагами ожидания и проверкой корреляции
- `[x]` Верификация сборки и прогона тестов
