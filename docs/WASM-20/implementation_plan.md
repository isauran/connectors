# WASM-20: Реализация расширенной спецификации BPMN 2.0

Этот план описывает шаги по расширению Go-модуля `bpmn` для поддержки инклюзивных шлюзов (OR), не-прерывающих граничных событий, сигналов, компенсаций и мульти-инстансов.

## User Review Required

> [!IMPORTANT]
> Базовые структуры в [bpmn.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/bpmn.go) и логика движка в [engine.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/engine.go) уже расширены на предыдущих этапах:
> - Добавлена история выполненных шагов `CompletedTasks` в `ProcessInstance`.
> - Добавлена поддержка Inclusive Gateway (Split/Join).
> - Реализованы методы `BroadcastSignal` и `TriggerCompensation`.
> - Реализованы не-прерывающие события и мульти-инстансы.
>
> **Текущий этап**: Нам необходимо написать комплексные тесты в [advanced_test.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/advanced_test.go) для проверки всей добавленной функциональности, а также убедиться, что они проходят.
> Дополнительно мы внесем небольшое улучшение в метод `BroadcastSignal` для поддержки Signal Event Subprocess (аналогично тому, как поддерживается Message Event Subprocess).

## Proposed Changes

### 1. Доработка движка

#### [MODIFY] [engine.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/engine.go)
- Добавить в метод `BroadcastSignal` обработку Signal Event Subprocess (запуск соответствующих событий в подпроцессах при рассылке сигналов).

### 2. Разработка тестов

#### [NEW] [advanced_test.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/advanced_test.go)
- **TestBPMNInclusiveGateway**: Проверка OR Split/Join. Схема с разветвлением на 2 ветки по условиям и их последующим инклюзивным слиянием.
- **TestBPMNNonInterruptingBoundary**: Проверка не-прерывающего Message Boundary Event (`cancelActivity="false"`). При корреляции сообщения исходная задача должна оставаться активной, а параллельно запускаться ветка обработчика события.
- **TestBPMNSignals**: Проверка широковещательного Signal Start/Boundary Event. Сигнал должен доходить до всех активных слушателей.
- **TestBPMNCompensations**: Проверка механизма компенсации. Завершение задачи, вызов `TriggerCompensation` для запуска компенсирующей задачи, связанной через Association.
- **TestBPMNMultiInstance**: Проверка мульти-инстансов (параллельные User Tasks/Service Tasks). Запуск `N` копий задач на основе коллекции в переменных и сбор их в виртуальном Parallel Join (`#join`).

---

## Verification Plan

### Automated Tests
- Запуск тестов модуля `bpmn`:
  ```bash
  cd bpmn && go test -v ./...
  ```
