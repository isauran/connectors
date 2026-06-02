# WASM-19: Полная поддержка спецификации BPMN 2.0 и интеграционное тестирование

Этот план описывает изменения в модуле `bpmn` для поддержки элементов `BusinessRuleTask` (с интеграцией DMN), граничных событий (`BoundaryEvent` для сообщений, ошибок и таймеров), вложенных подпроцессов (`SubProcess`) и подпроцессов событий (`Event Subprocess`).

## User Review Required

> [!IMPORTANT]
> Мы добавляем новые поля в структуру `ParsedProcess` (определенную в [bpmn.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/bpmn.go)):
> - `ParentSubProcesses map[string]string` — связь каждого внутреннего элемента с его родительским подпроцессом.
> - `SubProcesses map[string]*SubProcess` — быстрый доступ к структурам подпроцессов.
> - `BoundaryEventsByNode map[string][]BoundaryEvent` — связь граничных событий с их родительскими задачами.
>
> Также в [engine.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/engine.go) добавляются новые методы для доставки событий и ошибок:
> - `CorrelateMessage(instance *ProcessInstance, messageRef string, variables map[string]interface{}) error`
> - `HandleError(instance *ProcessInstance, nodeID string, errorCode string, variables map[string]interface{}) error`

## Proposed Changes

### 1. Расширение парсера и моделей BPMN 2.0

#### [MODIFY] [bpmn.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/bpmn.go)
- Добавить структуры:
  - `BusinessRuleTask` с атрибутами `ID`, `Name`, `DecisionRef`, `MapDecisionResult`, `ResultVariable`.
  - `BoundaryEvent` с атрибутами `ID`, `Name`, `AttachedToRef`, `TimerEventDefinition`, `MessageEventDefinition`, `ErrorEventDefinition`.
  - `TimerEventDefinition` с полем `TimeDuration`.
  - `MessageEventDefinition` с полем `MessageRef`.
  - `ErrorEventDefinition` с полем `ErrorRef`.
  - `SubProcess` с атрибутами `ID`, `Name`, `TriggeredByEvent` и списками всех внутренних элементов BPMN (рекурсивно).
- Изменить структуру `Process`, добавив `BusinessRuleTasks`, `BoundaryEvents` и `SubProcesses`.
- Изменить структуру `ParsedProcess`:
  - Добавить `ParentSubProcesses map[string]string`
  - Добавить `SubProcesses map[string]*SubProcess`
  - Добавить `BoundaryEventsByNode map[string][]BoundaryEvent`
- Обновить функцию `indexProcess` и реализовать вспомогательную рекурсивную функцию `indexSubProcess` для индексации внутренних элементов подпроцессов, заполнения родительских связей и привязки граничных событий.

---

### 2. Доработка логики выполнения в движке

#### [MODIFY] [engine.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/engine.go)
- В `Engine` добавить карту DMN определений: `DMNs map[string]*DMNDefinitions` и метод `RegisterDMN(decisionRef string, dmn *DMNDefinitions)`.
- В методе `Step()` добавить обработку новых типов узлов:
  - `BusinessRuleTask`: вычисление ассоциированной DMN-таблицы (через `Evaluate`) и запись результатов в контекст переменных.
  - `SubProcess`: перенос токена на внутренний `StartNodeID` подпроцесса.
  - `EndEvent`: если это конец подпроцесса (проверяется по `ParentSubProcesses`), токен переносится на исходящие потоки родительского подпроцесса, а не завершает весь инстанс.
- Добавить метод корреляции сообщений:
  ```go
  func (e *Engine) CorrelateMessage(instance *ProcessInstance, messageRef string, variables map[string]interface{}) error
  ```
  Который проверяет:
  1. Есть ли активные Event Subprocesses, слушающие это сообщение. Если да и `isInterrupting="true"`, прерывает текущее выполнение и запускает подпроцесс.
  2. Висит ли на какой-то из активных/ожидающих задач граничное событие (`BoundaryEvent`) с этим сообщением. Если да, прерывает задачу и переводит токен на обработчик.
  3. Есть ли в `WaitingTokens` задачи `ReceiveTask` или `IntermediateCatchEvent` с этим сообщением.
- Добавить метод обработки ошибок:
  ```go
  func (e *Engine) HandleError(instance *ProcessInstance, nodeID string, errorCode string, variables map[string]interface{}) error
  ```
  Который ищет `BoundaryEvent` с типом ошибки на самом узле или поднимается по иерархии родительских подпроцессов (`ParentSubProcesses`). При нахождении прерывает выполнение внутри области и переводит токен на обработчик.

---

### 3. Интеграционное тестирование

#### [NEW] [integration_test.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/integration_test.go)
- Создать тесты на основе реальных BPMN схем Camunda:
  - `TestBPMNDMNIntegration`: выполнение `dmn_business_rule.bpmn` с реальным парсингом `decision.dmn` и проверкой результата.
  - `TestBPMNEventsBoundary`: выполнение `events.bpmn` с проверкой граничных сообщений, ошибок и таймеров.
  - `TestBPMNSubprocesses`: выполнение `subprocesses.bpmn` с вложенными подпроцессами и глобальным Event Subprocess для прерывания процесса.

---

## Verification Plan

### Automated Tests
- Запуск тестов модуля `bpmn`:
  ```bash
  cd bpmn && go test -v ./...
  ```
