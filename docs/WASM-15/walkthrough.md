# Walkthrough - WASM-15: Go SDK for Durable WASM

В рамках задачи **WASM-15** разработан Go SDK непосредственно в корневом пакете `durable-wasm` (экспортируемый из пакета `durable`), призванный упростить написание бизнес-логики для Durable WASM воркеров, скрыв детали работы с хост-функциями, указателями, буферами и стейт-машинами с помощью Fluent API.

## Список изменений

1. **Создание Go SDK в корне пакета:**
   - Реализован файл [sdk.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/sdk.go) для среды WASM (`//go:build wasm`), декларирующий низкоуровневые импорты (`checkpoint`, `host_get_time`, `host_call_api`, `stream_data`).
   - Написаны `Reader` (`io.Reader`) и `Writer` (`io.WriteCloser`) для работы с сетевыми потоками хоста.
   - Реализована структура `Workflow` с Fluent методами `.Step(...)` и `.Run()` для связывания и последовательного выполнения шагов с автоматической фиксацией чекпоинтов.
   - Реализована структура `APICall` с Fluent методами `Call(...)`, `.WithPayload(...)` и `.Send()` для удобного детерминированного взаимодействия с API.
   - Создана заглушка [sdk_stub.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/sdk_stub.go) с флагом `//go:build !wasm` для обеспечения успешной сборки и тестирования SDK на хост-платформах.

2. **Настройка build tags для хост-файлов:**
   - В [engine.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine.go), [s3_store.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/s3_store.go) и [fs_store.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/fs_store.go) добавлен тег сборки `//go:build !wasm`.
   - Это гарантирует, что TinyGo при компиляции воркера в WebAssembly проигнорирует cgo-зависимости Wasmtime, которые не могут быть собраны под WASM.

3. **Миграция примера:**
   - Код воркера [examples/s3-store/worker/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/s3-store/worker/main.go) переписан на использование Fluent API пакета `durable` (`durable.NewWorkflow()`).
   - Код сократился, стал полностью линейным и очистился от `unsafe.Pointer` вызовов.

4. **Результаты тестирования:**
   - Воркер скомпилирован TinyGo в `.wasm` файл.
   - Тесты хост-системы и интеграционные тесты `go test ./...` успешно выполнены:
     ```bash
     $ go test ./...
     ok  	github.com/nativebpm/connectors/durable-wasm	1.878s
     ok  	github.com/nativebpm/connectors/durable-wasm/examples/camunda/host	2.304s
     ok  	github.com/nativebpm/connectors/durable-wasm/examples/gotenberg-telegram/host	1.045s
     ok  	github.com/nativebpm/connectors/durable-wasm/examples/process-csv/host	1.967s
     ?   	github.com/nativebpm/connectors/durable-wasm/examples/s3-store/host	[no test files]
     ok  	github.com/nativebpm/connectors/durable-wasm/examples/temporal/host	3.269s
     ```
