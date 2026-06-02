# WASM-17: Создание модуля bpmn для BPMN 2.0 и DMN движков

Этот план описывает шаги по созданию нового Go-модуля `bpmn` в монорепозитории. Новый модуль будет реализовывать парсинг и интерпретацию BPMN 2.0 схем и DMN таблиц решений, используя разработанный движок `wasman` в качестве Durable WASM SDK для выполнения шагов процесса.

## User Review Required

> [!IMPORTANT]
> Мы добавляем поддержку **Wait State (точек ожидания)** для `UserTask`. Это изменит структуру `ProcessInstance` (добавится поле `WaitingTokens []string`) и добавит новый метод корреляции `CompleteUserTask` в `bpmn.Engine`.

## Proposed Changes

### 1. Расширение моделей BPMN 2.0

#### [MODIFY] [bpmn.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/bpmn.go)
- [NEW] Добавление структуры `UserTask` для парсинга `<userTask>`.
- Изменение структуры `Process`: добавление поля `UserTasks []UserTask`.
- Изменение `indexProcess`: индексация `UserTask` в общую карту узлов `Nodes`.

---

### 2. Поддержка Wait State и Correlation в движке

#### [MODIFY] [engine.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/engine.go)
- Изменение `ProcessInstance`: добавление поля `WaitingTokens []string` для хранения токенов, заблокированных на шагах ожидания.
- Изменение метода `Step()`: при обработке `UserTask` токен переносится из `ActiveTokens` в `WaitingTokens`, а шаг завершается (Wait State). Автоматического перехода дальше не происходит.
- [NEW] Добавление метода `CompleteUserTask(instance *ProcessInstance, nodeID string, vars map[string]interface{}) error` для возобновления выполнения процесса внешним сигналом (разблокировка токена и переход на следующий шаг).

---

### 3. Тестирование

#### [MODIFY] [engine_test.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/engine_test.go)
- [NEW] Тест `TestBPMNUserTaskWaitState`: проверка, что процесс засыпает на `UserTask` и корректно возобновляется при вызове `CompleteUserTask` с обновлением переменных.

---

## Verification Plan

### Automated Tests
1. Запуск тестов модуля `bpmn`:
   ```bash
   cd bpmn && go test -v ./...
   ```
2. Сборка и запуск сквозного примера:
   ```bash
   cd bpmn/examples/orchestration && make build && make run
   ```ors/bpmn/examples/orchestration/host/main.go): Go-хост, который запускает `bpmn.Engine`, выполняет шаги, эмулирует падение (crash) в процессе работы WASM-воркера, восстанавливает сессию из снапшота и продолжает исполнение.
- [NEW] [worker/main.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/examples/orchestration/worker/main.go): Go-код воркера, компилируемый в WASM, реализующий шаги процесса с помощью Durable SDK `wasman`.
- [NEW] [Makefile](file:///Users/user/github.com/nativebpm/connectors/bpmn/examples/orchestration/Makefile): Скрипт сборки воркера в `.wasm` и запуска примера.
- [NEW] [README.md](file:///Users/user/github.com/nativebpm/connectors/bpmn/examples/orchestration/README.md): Описание работы примера.

---

## Verification Plan

### Automated Tests
1. Запуск тестов модуля `bpmn`:
   ```bash
   cd bpmn && go test -v ./...
   ```
2. Сборка и запуск нового примера:
   ```bash
   cd bpmn/examples/orchestration && make build && make run
   ```
