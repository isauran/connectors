---
task: TEMP-12
status: In Progress
summary: Создание примера интеграции Sequin Outbox CDC и Temporal на Go
---

# TEMP-12: Реализация примера Sequin Outbox CDC -> Temporal на Go

## Описание задачи
Реализовать обучающий пример в каталоге `temporal/examples/sequin-outbox`, демонстрирующий паттерн Transactional Outbox с использованием Sequin (CDC) и Temporal.

## Требования к примеру
1. **Имитация БД и CDC**:
   - При удалении пользователя (`DELETE` из таблицы `users`) Sequin захватывает изменения из WAL и отправляет POST вебхук.
   - Мы напишем Go HTTP-сервер, который принимает вебхук от Sequin и запускает соответствующий Temporal Workflow.
2. **Temporal Workflow & Activities**:
   - `DeleteUserWorkflow`: принимает ID пользователя и Email, последовательно выполняет очистку внешних систем и отправку письма.
   - `CleanUpExternalSystems`: активность, имитирующая очистку данных в S3 и внешних API.
   - `SendDeletionConfirmation`: активность, имитирующая отправку email.
3. **Интеграционные тесты**:
   - Написать юнит-тесты для Workflow с использованием `go.temporal.io/sdk/testsuite`.
   - Написать тесты для HTTP-обработчика вебхуков.
