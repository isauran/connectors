# Plan of Implementation - WASM-12: Durable WASM Stability Improvements

Этот план описывает шаги по повышению стабильности движка `durable-wasm` и покрытию его стресс-тестами.

## User Review Required

> [!IMPORTANT]
> Будет изменена сигнатура функции `Engine.Execute` — в нее добавится первым аргументом `ctx context.Context`. Это ломающее изменение для кода, который запускает движок напрямую (например, примеры/тесты). Нам нужно будет адаптировать все вызовы `Execute` в кодовой базе.

## Proposed Changes

### [Component: durable-wasm Engine]

#### [MODIFY] [engine.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine.go)
- **Конфигурация HTTP-клиента**:
  - Добавить функциональную опцию `WithHTTPClient(client *http.Client) EngineOption` для передачи переиспользуемого клиента.
  - По умолчанию использовать глобально оптимизированный HTTP-клиент, чтобы не плодить TCP-соединения.
- **Context-aware Execution**:
  - Изменить сигнатуру `Execute`:
    ```go
    func (e *Engine) Execute(ctx context.Context, instanceID string, entrypoint string, serverAddr string, shouldCrash bool) (bool, error)
    ```
  - Внутри `Execute` создавать дочерний контекст с отменой: `ctx, cancel := context.WithCancel(ctx)`.
  - Вызывать `cancel()` в `defer` при завершении `Execute`.
  - Передавать этот `ctx` в сетевые вызовы `handleDownload` и в горутину `handleUpload`.
- **Очистка ресурсов Wasmtime**:
  - Добавить `defer store.Close()` (или `linker.Close()`, если требуется) для явного высвобождения C-памяти в `Execute`.

#### [MODIFY] [engine_test.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine_test.go)
- Адаптировать все существующие тесты под новую сигнатуру `Execute(ctx, ...)`, передавая `context.Background()`.
- Добавить новые тесты:
  - `TestExecuteCancellation`: запускает Wasm, выполняющий долгий HTTP-запрос, отменяет контекст и проверяет быструю разблокировку и закрытие ресурсов.
  - `TestStorageErrorInjection`: подсовывает мок-хранилище, возвращающее I/O ошибки, проверяет стабильность движка и отсутствие утечек сокетов/горутин.
  - `TestSoakStressTesting`: запускает в конкурентном режиме (goroutines) выполнение 200 инстансов, измеряя использование дескрипторов и памяти.

---

### [Component: Examples and Other Tests]

Потребуется изменить вызовы `Execute` в примерах.

#### [MODIFY] [main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/durable-s3/main.go) (если он существует)
- Обновить вызов `engine.Execute` на `engine.Execute(context.Background(), )`.

---

## Verification Plan

### Automated Tests
- Запуск всех тестов в модуле `durable-wasm` для валидации стабильности и прохождения стресс-тестов:
  ```bash
  go test -v -race ./...
  ```
