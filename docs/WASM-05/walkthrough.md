# Walkthrough - Интеграция httpstream в durable-wasm

В рамках задачи **WASM-05** мы заменили стандартный пакет `net/http` на `httpstream` для отправки HTTP-запросов (загрузки и скачивания) в ядре движка `durable-wasm`.

## Изменения

### Durable WASM Core
* **[go.mod](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/go.mod)**:
  * Добавлена прямая зависимость от `github.com/nativebpm/httpstream v0.0.3`.
  * Директива `replace` исключена благодаря использованию Go Workspaces (`go.work`).
* **[engine.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/host/durable/engine.go)**:
  * Импортирован пакет `github.com/nativebpm/httpstream` и `"context"`.
  * Переписан метод `handleDownload` для использования Fluent API:
    ```go
    resp, err := httpstream.NewRequest(context.Background(), *s.httpClient, "GET", url).Send()
    ```
  * Переписан метод `handleUpload` для использования Fluent API и передачи тела запроса в виде ридера с типом содержимого:
    ```go
    resp, err := httpstream.NewRequest(context.Background(), *s.httpClient, "POST", url).
        Body(pipeReader, "application/octet-stream").
        Send()
    ```

## Результаты тестирования

* Все unit-тесты в пакете `durable-wasm` успешно проходят (`make test` выполняется без ошибок).
* Успешно выполнен пример для Camunda (`make -C durable-wasm/examples/camunda run`), демонстрирующий сохранение состояния, симуляцию падения хоста, восстановление из снапшота и потоковый обмен данными с мок-сервисами.
