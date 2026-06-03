# BPMN 2.0 & DMN Execution Engine

Легковесный интерпретатор и парсер BPMN 2.0 процессов и таблиц решений DMN на языке Go. Используется для управления сложными сценариями оркестрации, поддержки асинхронных шагов (Wait States) и запуска изолированных воркеров.

---

## Возможности

1. **Парсинг BPMN 2.0**: Преобразование стандартных XML-файлов BPMN в индексированный граф для быстрого обхода.
2. **Движок выполнения (`bpmn.Engine`)**:
   - Пошаговое продвижение токенов (`Step`).
   - Поддержка разветвлений и слияний: Exclusive Gateway (XOR), Parallel Gateway (AND), Inclusive Gateway (OR).
   - Поддержка Boundary Events (ошибки, таймеры, сигналы, сообщения).
   - Поддержка встроенных Subprocesses и Event Subprocesses.
   - Поддержка компенсаций (Compensation/Saga Pattern).
   - Поддержка Multi-Instance (параллельный запуск нескольких копий задач).
3. **Wait States (Шаги ожидания)**: Приостановка выполнения на `UserTask` или `ReceiveTask` с возможностью последующего возобновления (`CompleteTask` или `CorrelateMessage`).
4. **Интеграция DMN**: Вычисление таблиц решений на шаге `BusinessRuleTask`.
5. **Durable WASM Workers**: Запуск безопасных WebAssembly воркеров через `wasman` с поддержкой сохранения снимков памяти (durable execution).
6. **Local ServiceTask Handlers**: Регистрация стандартных Go-функций в качестве воркеров по имени топика (`Topic`), аналогично паттерну Camunda External Task.

---

## Быстрый старт с локальным воркером

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/nativebpm/connectors/bpmn"
)

func main() {
	// 1. Чтение и парсинг BPMN
	xmlData, _ := os.ReadFile("process.bpmn")
	pp, _ := bpmn.ParseBPMN(xmlData)

	// 2. Инициализация движка
	engine := bpmn.NewEngine(pp, nil)

	// 3. Регистрация обработчика (воркера) для топика
	engine.RegisterServiceTaskHandler("calculateInterest", func(ctx context.Context, instance *bpmn.ProcessInstance, task bpmn.ServiceTask) error {
		balance := instance.Variables["balance"].(float64)
		instance.Variables["interest"] = balance * 0.05
		return nil
	})

	// 4. Запуск экземпляра процесса
	instance, _ := engine.StartInstance("instance-1", map[string]interface{}{
		"balance": 1000.0,
	})

	// 5. Пошаговое выполнение
	engine.Step(context.Background(), instance) // StartEvent
	engine.Step(context.Background(), instance) // Выполнение calculateInterest (запуск воркера)
	engine.Step(context.Background(), instance) // EndEvent
}
```

---

## Структура примеров

В папке [examples](file:///Users/user/github.com/nativebpm/connectors/bpmn/examples) находятся готовые примеры использования:
* [worker](file:///Users/user/github.com/nativebpm/connectors/bpmn/examples/worker) — демонстрация локальных обработчиков (ServiceTaskHandlers) с использованием Camunda-like топиков.
* [saga](file:///Users/user/github.com/nativebpm/connectors/bpmn/examples/saga) — реализация паттерна Saga (компенсация выполненных шагов при сбое).
* [orchestration](file:///Users/user/github.com/nativebpm/connectors/bpmn/examples/orchestration) — интеграция с `wasman` для выполнения распределенных шагов в WebAssembly.

---

## Тестирование

Для запуска тестов воспользуйтесь командой:
```bash
go test -v ./...
```
