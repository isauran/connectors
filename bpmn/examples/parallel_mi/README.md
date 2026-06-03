# BPMN 2.0 Multi-Instance & Inclusive Gateway Example

Этот пример демонстрирует совместное использование двух мощных паттернов спецификации BPMN 2.0:
1. **Multi-Instance Tasks**: Динамическое размножение токенов для параллельной обработки элементов коллекции (например, согласование списка товаров).
2. **Inclusive Gateway (OR-шлюз)**: Параллельное разветвление и слияние потоков на основе условных логических выражений.

## Схема процесса

```mermaid
graph TD
  Start([Start]) --> MI_Task[Approve Item (Multi-Instance)]
  MI_Task --> Split{OR Split}
  
  Split -- isUrgent == true --> TaskUrgent[Process Urgent Order]
  Split -- isStandard == true --> TaskStandard[Process Standard Order]
  
  TaskUrgent --> Join{OR Join}
  TaskStandard --> Join
  
  Join --> End([End])
```

## Как запустить

Для запуска примера выполните:

```bash
go run main.go
```

Вы увидите подробный пошаговый вывод работы процесса, инициализацию мульти-инстансных копий задач, их слияние и распараллеливание на инклюзивном шлюзе по условным веткам.
