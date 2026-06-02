# Implementation Plan - Интеграция httpstream в durable-wasm

Этот план описывает шаги по интеграции пакета `httpstream` в модуль `durable-wasm`. Это позволит использовать Fluent API для стриминга данных (загрузка и выгрузка) и унифицировать HTTP-клиенты.

## Proposed Changes

Мы добавим зависимость от `httpstream` в `durable-wasm/go.mod` и перепишем логику HTTP-запросов в `engine.go`.

---

### Durable WASM Module

#### [MODIFY] [go.mod](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/go.mod)
* Добавить зависимость от `github.com/nativebpm/httpstream v0.0.3`.

#### [MODIFY] [engine.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/host/durable/engine.go)
* Импортировать пакет `github.com/nativebpm/httpstream`.
* В методе `handleDownload`:
  * Заменить прямой вызов `s.httpClient.Get` на `httpstream.NewRequest(context.Background(), *s.httpClient, "GET", url).Send()`.
* В методе `handleUpload`:
  * Внутри горутины заменить ручную подготовку `http.NewRequest` и `s.httpClient.Do` на цепочку вызовов `httpstream.NewRequest`:
    ```go
    resp, err := httpstream.NewRequest(context.Background(), *s.httpClient, "POST", url).
        Body(pipeReader, "application/octet-stream").
        Send()
    ```

---

## Verification Plan

### Automated Tests
1. Синхронизировать зависимости в Go-воркспейсе:
   ```bash
   go work sync
   ```
2. Прогнать тесты модуля `durable-wasm`:
   ```bash
   make -C durable-wasm test
   ```
3. Запустить примеры для проверки сквозной работоспособности (например, Camunda):
   ```bash
   make -C durable-wasm/examples/camunda run
   ```

### Manual Verification
* Убедиться в отсутствии предупреждений компилятора и линтера.
