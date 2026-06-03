# Walkthrough: WASM-20 — Реализация расширенной спецификации BPMN 2.0

В ходе выполнения задачи была полностью протестирована и финализирована поддержка расширенных элементов BPMN 2.0 в Go-модуле `bpmn`.

## Изменения

### 1. Движок процессов
- В [engine.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/engine.go) добавлена поддержка Signal Event Subprocess в методе `BroadcastSignal`, аналогично тому, как поддерживается Message Event Subprocess.

### 2. Тестовое покрытие
Создан файл [advanced_test.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/advanced_test.go), содержащий комплексные юнит-тесты:
- **Inclusive Gateway**: Тестирует ветвление (OR-Split) и слияние (OR-Join). Проверяет, что шлюз ожидает только потенциально достижимые токены (через поиск пути BFS).
- **Не-прерывающие граничные события**: Проверяет, что при наступлении Message Boundary Event с `cancelActivity="false"` исходная задача остается активной (в `WaitingTokens`), а параллельный поток запускается.
- **Широковещательные сигналы**: Проверяет рассылку сигнала (метод `BroadcastSignal`), активирующую несколько слушателей (Boundary Event и Intermediate Catch Event) одновременно.
- **Компенсации (Saga)**: Проверяет вызов компенсирующей задачи, связанной через Association, с отслеживанием истории завершенных задач (`CompletedTasks`).
- **Мульти-инстансы**: Проверяет инициализацию параллельных копий задач по коллекции (`loopDataInputRef`) с последующим сбором токенов на виртуальном Parallel Join (`#join`).

### 3. Исправление GitHub Workflow
- В файле [.github/workflows/publish-module-release.yml](file:///Users/user/github.com/nativebpm/connectors/.github/workflows/publish-module-release.yml) исправлен шаг извлечения аннотации тега. Ранее при создании обычного (lightweight) тега команда `git cat-file tag` падала с ошибкой 128, что приводило к сбою шага из-за pipefail. Теперь тип тега проверяется безопасно через `git cat-file -t`.

---

## Результаты тестирования

Все тесты были успешно пройдены:
```bash
ok  	github.com/nativebpm/connectors/bpmn	0.537s
ok  	github.com/nativebpm/connectors/camunda	2.769s
ok  	github.com/nativebpm/connectors/temporal	1.590s
```
