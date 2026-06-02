# Implementation Plan - WASM-15: Go SDK for Durable WASM

Этот план описывает этапы создания SDK в корне пакета `durable-wasm` для упрощения разработки отказоустойчивых WASM-воркеров.

## Proposed Changes

### [Component: Go SDK]

#### [NEW] [durable-wasm/sdk.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/sdk.go)
- Реализовать SDK в корне пакета `durable`.
- Импортировать все необходимые хост-функции (`checkpoint`, `host_get_time`, `host_call_api`, `stream_data`) с build tag `//go:build wasm`.
- Предоставить API для работы с детерминированным временем и вызовами хоста: `GetTime()`, `CallAPI()`.
- Реализовать `Reader` и `Writer` на базе `stream_data` с поддержкой стандартных интерфейсов `io.Reader` и `io.WriteCloser`.
- Предоставить Fluent API структуры `Workflow` и `APICall` для объединения шагов и внешних вызовов в цепочки.

#### [NEW] [durable-wasm/sdk_stub.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/sdk_stub.go)
- Реализовать заглушки всех структур и функций SDK с build tag `//go:build !wasm` для успешной компиляции на хосте.

#### [MODIFY] [durable-wasm/engine.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine.go), [s3_store.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/s3_store.go), [fs_store.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/fs_store.go)
- Добавить `//go:build !wasm` в начало файлов хоста, чтобы TinyGo игнорировал их при сборке воркера в WASM (это предотвращает ошибки компиляции CGO-зависимостей Wasmtime).

---

### [Component: Examples]

#### [MODIFY] [examples/s3-store/worker/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/s3-store/worker/main.go)
- Удалить все импорты хост-функций и низкоуровневые манипуляции с памятью/указателями.
- Переписать логику воркера с использованием `durable.NewWorkflow()` Fluent API и типизированного потокового ввода-вывода (`durable.Reader`, `durable.Writer`).

---

## Verification Plan

### Automated Tests
- Собрать воркер с помощью TinyGo и запустить интеграционный тест `s3-store` хоста:
  ```bash
  cd durable-wasm
  # Сборка воркера
  tinygo build -o examples/s3-store/worker/worker.wasm -target=wasi -panic=trap examples/s3-store/worker/main.go
  # Запуск тестов хоста
  go test -v ./...
  ```
