# Task Checklist - WASM-12

- `[x]` Реализовать опцию `WithHTTPClient` и глобальный клиент по умолчанию в `engine.go`
- `[x]` Добавить поддержку `context.Context` в сигнатуру `Execute` и отмену контекста в `engine.go`
- `[x]` Реализовать явное высвобождение ресурсов Wasmtime (`store.Close()`) в `Execute`
- `[x]` Адаптировать существующие тесты в `engine_test.go` под новую сигнатуру `Execute`
- `[x]` Адаптировать пример `examples/durable-s3/host/main.go` под новую сигнатуру `Execute`
- `[x]` Добавить тест `TestExecuteCancellation` (отмена контекста) в `engine_test.go`
- `[x]` Добавить тест `TestStorageErrorInjection` (мок SnapshotStore с ошибками) в `engine_test.go`
- `[x]` Добавить стресс-тест `TestSoakStressTesting` (конкурентные запуски 200 инстансов) в `engine_test.go`
- `[x]` Запустить тесты и верифицировать работоспособность
