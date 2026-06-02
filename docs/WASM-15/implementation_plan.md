# Implementation Plan - WASM-15: Go SDK Workflow State Pattern Migration

Этот план описывает внедрение паттерна структуры состояния (State Struct) с ленивой инициализацией во все примеры Durable WASM воркеров. Этот паттерн решает проблему безопасной передачи переменных между шагами воркфлоу, предотвращая загрязнение глобальной области видимости и гарантируя сохранность данных при восстановлении из чекпоинтов.

## User Review Required

> [!NOTE]
> При использовании структуры состояния, методы которой являются шагами воркфлоу, инициализация структуры должна выполняться лениво в точке входа воркера `run()` (только если глобальный указатель `state == nil`). Это гарантирует, что при восстановлении воркера из сохраненного снимка памяти хоста, куча с уже заполненным объектом состояния будет восстановлена, а повторный вызов `run()` не перезапишет её новым пустым объектом.

## Pattern Description: State Struct with Lazy Initialization

Для передачи данных между шагами мы группируем все переменные в одну структуру:

```go
type State struct {
	OrderID     string
	InventoryOk bool
	PaymentOk   bool
}

var state *State

//export run
func run() int32 {
	if state == nil {
		state = &State{
			OrderID: "ORD-CAM-8899", // Начальное состояние
		}
	}
	return durable.NewWorkflow().
		Step(state.checkInventory).
		Step(state.capturePayment).
		Run()
}

func (s *State) checkInventory() error {
	// Чтение или запись в s.InventoryOk
	s.InventoryOk = true
	return nil
}

func (s *State) capturePayment() error {
	// Доступ к s.InventoryOk и запись в s.PaymentOk
	if s.InventoryOk {
		s.PaymentOk = true
	}
	return nil
}
```

## Proposed Changes

### [Component: Examples]

Мы перепишем все 5 примеров воркеров на использование структуры состояния:

#### [MODIFY] [examples/s3-store/worker/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/s3-store/worker/main.go)
- Перенести глобальную переменную `processedBytes` в структуру `State`.
- Сделать функции `initialize`, `processStream`, `finalizeWorkflow` методами `*State`.
- Выполнить ленивую инициализацию `state` в `run()`.

#### [MODIFY] [examples/camunda/worker/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/camunda/worker/main.go)
- Перенести `orderID`, `inventoryOk`, `paymentOk` в структуру `State`.
- Заменить обычные функции шагов на методы `*State`.
- Ленивая инициализация `state` в `run()`.

#### [MODIFY] [examples/process-csv/worker/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/process-csv/worker/main.go)
- Перенести `totalAmount` и `validRecords` в структуру `State`.
- Заменить функции шагов на методы `*State`.
- Ленивая инициализация `state` в `run()`.

#### [MODIFY] [examples/gotenberg-telegram/worker/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/gotenberg-telegram/worker/main.go)
- Перенести `chatID`, `fileID`, `docxBytes` и `pdfBytes` в структуру `State`.
- Заменить функции шагов на методы `*State`.
- Ленивая инициализация `state` в `run()`.

#### [MODIFY] [examples/temporal/worker/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/temporal/worker/main.go)
- Перенести `activityID`, `baseRate`, `multiplier` и `calculatedVal` в структуру `State`.
- Заменить функции шагов на методы `*State`.
- Ленивая инициализация `state` в `run()`.

---

## Verification Plan

### Automated Tests
1. Перекомпилировать все воркеры с помощью TinyGo (для каждого примера):
   ```bash
   cd durable-wasm
   tinygo build -o examples/s3-store/worker/worker.wasm -target=wasi -panic=trap examples/s3-store/worker/main.go
   tinygo build -o examples/camunda/worker/worker.wasm -target=wasi -panic=trap examples/camunda/worker/main.go
   tinygo build -o examples/process-csv/worker/worker.wasm -target=wasi -panic=trap examples/process-csv/worker/main.go
   tinygo build -o examples/gotenberg-telegram/worker/worker.wasm -target=wasi -panic=trap examples/gotenberg-telegram/worker/main.go
   tinygo build -o examples/temporal/worker/worker.wasm -target=wasi -panic=trap examples/temporal/worker/main.go
   ```
2. Запустить все тесты хост-движка в `durable-wasm` для подтверждения, что функциональность snapshot/restore работает без сбоев:
   ```bash
   go test -v ./...
   ```
