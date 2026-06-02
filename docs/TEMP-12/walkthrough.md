# Walkthrough - TEMP-12

В рамках задачи **TEMP-12** был успешно реализован обучающий пример `temporal/examples/sequin-outbox`, демонстрирующий интеграцию Sequin CDC и Temporal на языке Go.

## Список изменений

1. **Созданы новые файлы примера в `temporal/examples/sequin-outbox`**:
   - `workflow.go` — Temporal воркфлоу `DeleteUserWorkflow` для координации удаления.
   - `activities.go` — Активности для имитации очистки данных во внешних системах и отправки почты.
   - `handler.go` — Выделенная бизнес-логика разбора вебхуков от Sequin и запуска воркфлоу в Temporal с использованием интерфейса `WorkflowStarter`.
   - `server/main.go` — Исполняемый HTTP-сервер, слушающий POST-запросы от Sequin на порту `3333`.
   - `worker/main.go` — Исполняемый Temporal воркер, регистрирующий воркфлоу и активности.
   - `integration_test.go` — Юнит-тестирование воркфлоу и сквозное тестирование HTTP-обработчика с мокированием клиента Temporal.
   - `README.md` — Подробная документация по локальному запуску Postgres, Sequin, Temporal и самого примера.

2. **Обновлен `temporal/Makefile`**:
   - Добавлены новые таргеты для простого локального запуска:
     - `make run-server-sequin-outbox`
     - `make run-worker-sequin-outbox`

3. **Оформлена документация Semantic Store**:
   - Создана директория `docs/TEMP-12/` с описанием задачи, планом, чек-листом и данным отчетом.
   - Обновлен глобальный индекс задач `docs/index.md`.

## Результаты тестирования

Все тесты были успешно пройдены в локальной тестовой среде.

### Вывод команды `go test -v ./examples/sequin-outbox/...`

```
=== RUN   TestUnitTestSuite
=== RUN   TestUnitTestSuite/Test_DeleteUserWorkflow_CleanupFailure
2026/06/02 21:07:16 ERROR Activity error. ... Error cleanup failed
...
=== RUN   TestUnitTestSuite/Test_DeleteUserWorkflow_Success
--- PASS: TestUnitTestSuite (0.03s)
    --- PASS: TestUnitTestSuite/Test_DeleteUserWorkflow_CleanupFailure (0.03s)
    --- PASS: TestUnitTestSuite/Test_DeleteUserWorkflow_Success (0.00s)
=== RUN   TestSequinWebhookHandler
=== RUN   TestSequinWebhookHandler/Success_Delete_Action
2026/06/02 21:07:16 [Webhook] Received delete event for user: user-123 (john.doe@example.com)
2026/06/02 21:07:16 [Webhook] Successfully started DeleteUserWorkflow. ID: delete-user-user-123, RunID: run-123
=== RUN   TestSequinWebhookHandler/Ignore_Non-Delete_Action
2026/06/02 21:07:16 Ignoring non-delete action: insert
=== RUN   TestSequinWebhookHandler/Invalid_JSON_Payload
2026/06/02 21:07:16 Error decoding JSON payload: ...
=== RUN   TestSequinWebhookHandler/Missing_Old_Record_on_Delete
2026/06/02 21:07:16 Error: delete action received but old_record is nil
--- PASS: TestSequinWebhookHandler (0.00s)
    --- PASS: TestSequinWebhookHandler/Success_Delete_Action (0.00s)
    --- PASS: TestSequinWebhookHandler/Ignore_Non-Delete_Action (0.00s)
    --- PASS: TestSequinWebhookHandler/Invalid_JSON_Payload (0.00s)
    --- PASS: TestSequinWebhookHandler/Missing_Old_Record_on_Delete (0.00s)
PASS
ok  	github.com/nativebpm/connectors/temporal/examples/sequin-outbox	0.771s
```
