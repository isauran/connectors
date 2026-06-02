# Walkthrough - WASM-15: Go SDK & Fluent API for Durable WASM

В рамках задачи **WASM-15** разработан Go SDK непосредственно в корневом пакете `durable-wasm` (экспортируемый из пакета `durable`), а также реализован полностью Fluent API без использования дженериков для чистой и безопасной передачи состояния.

## Список изменений

1. **Создание Fluent API в Go SDK:**
   - Реализован файл [sdk.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/sdk.go) для среды WASM (`//go:build wasm`), декларирующий низкоуровневые импорты.
   - Реализована простая, не-дженериковая структура `Workflow` с методами:
     - `NewWorkflow()`: инициализирует новый fluent-раннер.
     - `.Step(step func() error)`: добавляет шаг в виде связанного метода экземпляра состояния (например, `state.checkInventory`).
     - `.Run()`: запускает выполнение шагов последовательно с автоматической фиксацией чекпоинтов после каждого шага.
   - Создана заглушка [sdk_stub.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/sdk_stub.go) с флагом `//go:build !wasm` для успешной сборки хоста.

2. **Статическая инициализация состояния и удаление дженериков:**
   - Состояние воркфлоу инициализируется классическим для Go статическим способом при объявлении глобальной переменной:
     ```go
     var state = &State{
         ChatID: 77665544,
         FileID: "file_docx_invoice_102",
     }
     ```
   - Поскольку при запуске `run()` (как в первый раз, так и при восстановлении после сбоя) глобальная переменная `state` гарантированно не равна `nil` (так как хост Wasmtime восстанавливает снимок памяти до вызова `run()`), вызовы связанных методов `state.method` выполняются абсолютно безопасно, исключая риски паники `nil pointer dereference`.
   - Это позволило полностью отказаться от дженериков, method expressions (`(*State).method`) и дополнительных замыканий-инициализаторов.

3. **Миграция всех примеров на чистый Fluent API:**
   - Обновлены воркеры во всех 5 примерах в директории `examples/`:
     - [examples/s3-store/worker/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/s3-store/worker/main.go)
     - [examples/camunda/worker/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/camunda/worker/main.go)
     - [examples/process-csv/worker/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/process-csv/worker/main.go)
     - [examples/gotenberg-telegram/worker/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/gotenberg-telegram/worker/main.go)
     - [examples/temporal/worker/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/temporal/worker/main.go)

4. **Настройка build tags для хост-файлов:**
   - В [engine.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine.go), [s3_store.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/s3_store.go) и [fs_store.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/fs_store.go) добавлен тег сборки `//go:build !wasm`.

5. **Исправление .gitignore**:
   - В [durable-wasm/.gitignore](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/.gitignore) добавлено исключение `!**/host/`.

6. **Результаты тестирования:**
   - Все воркеры скомпилированы в `.wasm`.
   - Интеграционные тесты хост-системы успешно пройдены:
     ```bash
     $ make test
     Running tests...
     go test -v ./...
     ok  	github.com/nativebpm/connectors/durable-wasm	1.820s
     ok  	github.com/nativebpm/connectors/durable-wasm/examples/camunda/host	2.298s
     ok  	github.com/nativebpm/connectors/durable-wasm/examples/gotenberg-telegram/host	1.755s
     ok  	github.com/nativebpm/connectors/durable-wasm/examples/process-csv/host	2.284s
     ?   	github.com/nativebpm/connectors/durable-wasm/examples/s3-store/host	[no test files]
     ok  	github.com/nativebpm/connectors/durable-wasm/examples/temporal/host	3.892s
     ```
