# Plan of Implementation - TEMP-12

Реализация обучающего примера `temporal/examples/sequin-outbox` на Go, показывающего интеграцию Sequin CDC с Temporal.

## Proposed Changes

### [Component: Temporal Sequin Outbox Example]

Создание нового примера в каталоге `temporal/examples/sequin-outbox`.

#### [NEW] [workflow.go](file:///Users/user/github.com/nativebpm/connectors/temporal/examples/sequin-outbox/workflow.go)
- Определение воркфлоу `DeleteUserWorkflow(ctx workflow.Context, userID string, email string) error`.
- Вызов активностей `CleanUpExternalSystems` и `SendDeletionConfirmation` с таймаутами и стратегиями повтора.

#### [NEW] [activities.go](file:///Users/user/github.com/nativebpm/connectors/temporal/examples/sequin-outbox/activities.go)
- Определение активностей:
  - `CleanUpExternalSystems(ctx context.Context, userID string) error` (эмуляция удаления из S3, Stripe и т.д.).
  - `SendDeletionConfirmation(ctx context.Context, email string) error` (эмуляция отправки email).

#### [NEW] [server.go](file:///Users/user/github.com/nativebpm/connectors/temporal/examples/sequin-outbox/server.go)
- HTTP-сервер на Go для приема вебхуков от Sequin.
- Обработчик POST `/delete-user`:
  - Читает тело JSON. Запрос содержит структуру Sequin для событий удаления:
    ```json
    {
      "action": "delete",
      "old_record": {
        "id": "uuid",
        "email": "user@example.com"
      }
    }
    ```
  - Инициализирует Temporal клиент (с использованием `temporal.NewClient` из корня проекта).
  - Запускает `DeleteUserWorkflow` асинхронно с `WorkflowID` вида `delete-user-<record.id>`.

#### [NEW] [worker/main.go](file:///Users/user/github.com/nativebpm/connectors/temporal/examples/sequin-outbox/worker/main.go)
- Инициализация и запуск Temporal Worker для обработки воркфлоу `DeleteUserWorkflow`.

#### [NEW] [integration_test.go](file:///Users/user/github.com/nativebpm/connectors/temporal/examples/sequin-outbox/integration_test.go)
- Тестирование воркфлоу с помощью `go.temporal.io/sdk/testsuite`.
- Тестирование вебхук-обработчика в `server.go` с помощью `net/http/httptest` и мокированием Temporal API.

#### [NEW] [README.md](file:///Users/user/github.com/nativebpm/connectors/temporal/examples/sequin-outbox/README.md)
- Инструкция по локальному запуску Postgres, Sequin, Temporal и самого примера.

#### [MODIFY] [Makefile](file:///Users/user/github.com/nativebpm/connectors/temporal/Makefile)
- Добавление таргетов для тестирования и сборки нового примера.

## Verification Plan

### Automated Tests
- Запуск тестов нового примера:
  ```bash
  go test -v ./examples/sequin-outbox/...
  ```
- Запуск всех тестов в модуле `temporal` для проверки отсутствия регрессий:
  ```bash
  go test -v ./...
  ```

### Manual Verification
- Сборка бинарников `server` и `worker`:
  ```bash
  go build -o bin/sequin-server ./examples/sequin-outbox/server.go
  go build -o bin/sequin-worker ./examples/sequin-outbox/worker/main.go
  ```
