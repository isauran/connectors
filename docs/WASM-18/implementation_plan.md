# WASM-18: Поддержка асинхронных шагов ожидания (Wait State) в BPMN-движке

Этот план описывает изменения в модуле `bpmn` для поддержки шагов ожидания (Wait State) при выполнении BPMN процессов, с возможностью корреляции по внешнему сигналу.

## User Review Required

> [!IMPORTANT]
> Мы добавляем новое поле `WaitingTokens []string` в структуру `ProcessInstance` (определенную в [engine.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/engine.go)). Это поле сериализуется в JSON и сохраняется в состоянии процесса.
> Также добавляется метод `CompleteTask(instance *ProcessInstance, nodeID string, variables map[string]interface{}) error` в структуру `Engine`.

## Proposed Changes

### 1. Расширение парсера BPMN 2.0

#### [MODIFY] [bpmn.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/bpmn.go)
- Добавить структуры:
  - `UserTask` с атрибутами `ID` и `Name`.
  - `ReceiveTask` с атрибутами `ID`, `Name` и `MessageRef`.
  - `IntermediateCatchEvent` с атрибутами `ID` и `Name`.
- Изменить структуру `Process`, объявив `UserTasks []UserTask`, `ReceiveTasks []ReceiveTask` и `IntermediateCatchEvents []IntermediateCatchEvent`.
- В функции `indexProcess` добавить индексацию новых типов узлов в общую карту `Nodes`.

---

### 2. Поддержка Wait State и Correlation в движке

#### [MODIFY] [engine.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/engine.go)
- Добавить поле `WaitingTokens []string` в структуру `ProcessInstance` с тегом `json:"waiting_tokens"`.
- В методе `Step()` добавить ветки switch-case для `UserTask`, `ReceiveTask` и `IntermediateCatchEvent`. При их достижении:
  - Выводить информационный лог в `slog`.
  - Добавлять ID текущего узла в `instance.WaitingTokens`.
  - Прекращать дальнейшее продвижение токена (т.е. не вызывать `moveToken`).
- Добавить новый метод:
  ```go
  func (e *Engine) CompleteTask(instance *ProcessInstance, nodeID string, variables map[string]interface{}) error
  ```
  Который:
  - Проверяет наличие `nodeID` в `instance.WaitingTokens`.
  - Если токен найден, удаляет его из списка ожидания.
  - Объединяет переданные `variables` с переменными процесса `instance.Variables`.
  - Продвигает токен дальше, вызывая `e.moveToken(instance, nodeID, "")`.

---

### 3. Написание тестов

#### [MODIFY] [engine_test.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/engine_test.go)
- Добавить XML схему `waitStateBPMN` с `userTask` и `receiveTask`.
- Добавить тест `TestBPMNWaitStates`:
  - Проверить, что процесс приостанавливается на `userTask`, токен попадает в `WaitingTokens`, и вызов `Step()` более не двигает его вперед.
  - Вызвать `CompleteTask` для `userTask`, проверить, что токен переходит на следующий шаг.
  - Проверить аналогичное поведение для `receiveTask`.

---

## Verification Plan

### Automated Tests
- Запуск тестов модуля `bpmn`:
  ```bash
  cd bpmn && go test -v ./...
  ```
- Проверка сборки и тестов всего проекта:
  ```bash
  go test -v ./...
  ```
