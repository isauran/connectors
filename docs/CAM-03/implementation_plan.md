# CAM-03: Camunda CDC с использованием Sequin Enrichment и явным Lock

Реализация архитектуры обработки внешних задач Camunda 7 с использованием **Sequin SQL Enrichment** (нативное обогащение данных в Sequin) и явного вызова метода `Lock` через REST API перед выполнением задачи.

Мы полностью отказываемся от триггеров авто-блокировки в базе данных Camunda, оставляя БД полностью "коробочной". Блокировка (`Lock`) и завершение (`Complete`) выполняются явно через стандартный REST API воркера, а переменные процесса и бизнес-ключ поступают реактивно из CDC-сообщения Sequin (благодаря SQL Enrichment).

## User Review Required

> [!IMPORTANT]
> - **Сброс БД**: Для удаления уже примененной миграции триггера в локальном окружении потребуется сбросить и перезапустить контейнеры Camunda (`docker compose down -v && docker compose up -d`). Это очистит базу данных и пересоздаст ее на основе обновленного списка миграций.
> - **Сетевые запросы**: В CDC-режиме воркер теперь будет делать **2 REST-запроса** к Camunda на задачу: `Lock` и `Complete` (вместо 1 запроса). Это по-прежнему исключает дорогостоящие запросы на получение переменных (`GetVariables`) и бизнес-ключа.

## Open Questions

Нет открытых вопросов.

## Proposed Changes

---

### 1. База данных Camunda (Migrations)

#### [DELETE] [20260603000000_add_task_autolock_trigger.sql](file:///Users/user/github.com/nativebpm/connectors/camunda/docker/camunda/migrations/20260603000000_add_task_autolock_trigger.sql)
Удалить файл миграции с триггером авто-блокировки.

#### [MODIFY] [atlas.sum](file:///Users/user/github.com/nativebpm/connectors/camunda/docker/camunda/migrations/atlas.sum)
Пересчитать хэши миграций (запустить `make atlas-hash`), чтобы исключить удаленную миграцию из контрольной суммы Atlas.

---

### 2. Ядро коннектора (Camunda SDK)

#### [MODIFY] [camunda_sequin.go](file:///Users/user/github.com/nativebpm/connectors/camunda/camunda_sequin.go)
Обновить логику метода `processMessage` в `SequinWorker`:
- **Lock для всех режимов**: Вызов REST API `Lock` должен выполняться ВСЕГДА (как в CDC-режиме, так и в legacy-режиме) перед запуском бизнес-обработчика.
- **Условный сбор переменных**:
  - Если в сообщении есть `metadata.enrichment`, переменные и бизнес-ключ берутся из сообщения (запросы `GetExecutionVariables` и `GetProcessInstanceBusinessKey` пропускаются).
  - Если метаданных нет, переменные и бизнес-ключ запрашиваются по сети через REST API.

---

## Verification Plan

### Automated Tests
Убедиться, что сборка проекта проходит успешно:
```bash
go build ./camunda/...
go test -v ./camunda -run TestSequinWorker
```

### Manual Verification
1. Сбросить локальную БД, пересоздать контейнеры и применить чистые миграции:
   ```bash
   cd camunda/docker/camunda
   docker compose down -v
   docker compose up -d
   ```
2. Выполнить накат миграций (без триггера):
   ```bash
   cd ../..
   make atlas-apply
   ```
3. Запустить пример `loan-granting-cdc-outbox`:
   ```bash
   cd camunda/examples/loan-granting-cdc-outbox
   go run main.go
   ```
4. Убедиться по логам воркера, что задачи выполняются успешно. В логах Camunda REST API должно быть ровно два запроса на задачу: `Lock` и `Complete`.
