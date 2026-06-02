# WASM-19 Task Checklist

- `[x]` Расширение моделей BPMN в [bpmn.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/bpmn.go)
  - `[x]` Добавить структуру `BusinessRuleTask`
  - `[x]` Добавить структуру `BoundaryEvent` и определения событий
  - `[x]` Добавить структуру `SubProcess`
  - `[x]` Обновить структуру `Process` и `ParsedProcess`
  - `[x]` Реализовать рекурсивную индексацию подпроцессов в `indexProcess`/`indexSubProcess`
- `[x]` Логика выполнения в движке в [engine.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/engine.go)
  - `[x]` Интеграция DMN и поддержка `BusinessRuleTask`
  - `[x]` Поддержка жизненного цикла вложенных подпроцессов (`SubProcess` и `EndEvent`)
  - `[x]` Реализация доставки сообщений `CorrelateMessage` (Event Subprocesses, Boundary Messages, Receive Tasks)
  - `[x]` Реализация обработки ошибок `HandleError` (Boundary Errors по иерархии подпроцессов)
- `[x]` Написание тестов в [integration_test.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/integration_test.go)
  - `[x]` Тест DMN интеграции на схеме `dmn_business_rule.bpmn`
  - `[x]` Тест граничных событий на схеме `events.bpmn`
  - `[x]` Тест подпроцессов на схеме `subprocesses.bpmn`
- `[x]` Верификация сборки и прогона тестов
