# Пример локального воркера/обработчика (BPMN ServiceTask)

Этот пример демонстрирует использование встроенного в модуль `bpmn` локального реестра обработчиков задач (`ServiceTaskHandler`) для выполнения шагов схемы BPMN в Go-коде.

## Описание сценария

1. Процесс начинается со `StartEvent`.
2. Шаг `calculate_interest` (ServiceTask с топиком `calculateInterest`) вычисляет 5% годовых от текущего баланса и записывает в переменную `calculated_interest`.
3. Шаг `apply_interest` (ServiceTask с топиком `applyInterest`) прибавляет начисленный процент к балансу.
4. Процесс завершается в `EndEvent`.

Оба шага реализованы как обычные Go-функции, зарегистрированные в движке по имени топика (`Topic`), аналогично концепции External Task в Camunda.

## Как запустить пример

```bash
cd bpmn/examples/worker
go run main.go
```

## Связь схемы BPMN и кода Go

В файле `process.bpmn` сервис-таски настроены со специальным атрибутом `topic`:

```xml
<serviceTask id="calculate_interest" name="Calculate Interest" topic="calculateInterest" />
```

В коде Go мы регистрируем обработчик для этого топика в `bpmn.Engine`:

```go
engine.RegisterServiceTaskHandler("calculateInterest", func(ctx context.Context, instance *bpmn.ProcessInstance, task bpmn.ServiceTask) error {
    // Чтение переменных и бизнес-логика
    balance := instance.Variables["balance"].(float64)
    interest := balance * 0.05
    
    // Запись результатов обратно в контекст процесса
    instance.Variables["calculated_interest"] = interest
    return nil
})
```
