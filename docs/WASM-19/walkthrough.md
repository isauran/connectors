# WASM-19 Walkthrough: Полная спецификация BPMN 2.0 и интеграция DMN

Реализована полная поддержка спецификации BPMN 2.0, покрывающая все элементы, используемые в реальных Camunda-схемах проекта.

## Изменения

### 1. Парсер BPMN
В файле [bpmn.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/bpmn.go):
- Добавлены новые структуры: `BusinessRuleTask`, `BoundaryEvent` (с таймерами, сообщениями и ошибками) и `SubProcess`.
- Внедрена поддержка глобальных списков сообщений (`Message`) и ошибок (`Error`) в корневом элементе `Definitions` для разрешения их `messageRef`/`errorRef` в реальные имена/коды.
- Внедрены новые карты связей в `ParsedProcess`:
  - `ParentSubProcesses map[string]string` (связь дочерних узлов с их родительским подпроцессом).
  - `SubProcesses map[string]*SubProcess` (доступ к объектам подпроцессов).
  - `BoundaryEventsByNode map[string][]BoundaryEvent` (привязка граничных событий к задачам).
- Написана рекурсивная функция индексации подпроцессов `indexSubProcess`.

### 2. Движок процессов
В файле [engine.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/engine.go):
- **BusinessRuleTask**: Внедрена интеграция с DMN (`DMNs map[string]*DMNDefinitions` в `Engine` и метод `RegisterDMN`). При достижении `BusinessRuleTask` автоматически вызывается `Evaluate` и результаты записываются в контекст процесса.
- **Вложенные подпроцессы**: При входе токен переносится на внутренний `StartEvent` подпроцесса, а при достижении внутреннего `EndEvent` — возвращается в родительский поток через исходящие переходы самого подпроцесса.
- **Доставка сообщений (`CorrelateMessage`)**:
  - Поддерживает запуск прерывающих Event Subprocesses (`triggeredByEvent="true"` и `isInterrupting="true"`) по сообщению с очисткой остальных токенов.
  - Поддерживает прерывание активных/ожидающих задач при срабатывании граничных событий (`BoundaryEvent` с типом сообщения).
  - Поддерживает стандартные `ReceiveTask` и `IntermediateCatchEvent`.
- **Обработка ошибок (`HandleError`)**:
  - Ищет граничные события ошибок на самом узле, а при их отсутствии поднимается по иерархии родительских подпроцессов (`ParentSubProcesses`). При срабатывании полностью очищает все дочерние токены в этой области видимости и переводит токен на обработчик.

### 3. Интеграционное тестирование
Создан новый тестовый набор [integration_test.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/integration_test.go), который тестирует работу движка на трех **реальных файлах схем Camunda** из каталога `camunda/examples/bpmn-spec/bpmn/`:
- `TestBPMNDMNIntegration` (запуск `dmn_business_rule.bpmn` и вычисление `decision.dmn`).
- `TestBPMNEventsBoundary` (граничные сообщения и ошибки в `events.bpmn`).
- `TestBPMNSubprocesses` (подпроцессы, граничные ошибки подпроцессов и Event Subprocesses в `subprocesses.bpmn`).

## Верификация
Все тесты успешно пройдены:
```bash
go test -v ./...
# ok  	github.com/nativebpm/connectors/bpmn	0.404s
```
Интеграционные тесты сравнения с Camunda:
```bash
go test -v -run TestBPMNVSRealCamunda
# PASS
# ok  	github.com/nativebpm/connectors/camunda	2.496s
```
