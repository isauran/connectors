# CAM-03: Camunda CDC с использованием Sequin Enrichment и авто-блокировки

Реализация архитектуры обработки внешних задач Camunda 7 без дополнительных REST-запросов к Camunda (`Lock`, `GetVariables`) за счет использования **Sequin SQL Enrichment** (нативное обогащение данных в Sequin) и триггера авто-блокировки в БД Camunda.

Вместо отдельного воркера `OutboxWorker` мы дорабатываем штатный `camunda.SequinWorker`, чтобы он автоматически поддерживал zero-lookup режим при наличии обогащенных метаданных от Sequin, и сохранял обратную совместимость (fallback на REST-запросы) при их отсутствии.

## User Review Required

> [!IMPORTANT]
> - **Обратная совместимость**: `SequinWorker` сохраняет полную обратную совместимость. Если Sequin не настроен на обогащение (например, нет метаданных), воркер по-прежнему будет делать REST API запросы `Lock`, `GetExecutionVariables` и `GetProcessInstanceBusinessKey`.
> - **Удаление OutboxWorker**: Локальный `OutboxWorker` из примера `loan-granting-cdc-outbox` будет полностью удален, а пример переведен на использование доработанного `camunda.SequinWorker`.

## Open Questions

Нет открытых вопросов.

## Proposed Changes

---

### 1. База данных Camunda (Migrations)

#### [MODIFY] [20260603000000_add_task_autolock_trigger.sql](file:///Users/user/github.com/nativebpm/connectors/camunda/docker/camunda/migrations/20260603000000_add_task_autolock_trigger.sql)
*(Уже создано)* Триггер авто-блокировки на таблицу `act_ru_ext_task` (`BEFORE INSERT OR UPDATE`). Устанавливает `worker_id_ = 'loan-worker-cdc'` и `lock_exp_time_ = NOW() + INTERVAL '5 minutes'`, если `NEW.worker_id_ IS NULL`.

---

### 2. Конфигурация Sequin

#### [MODIFY] [playground.yml](file:///Users/user/github.com/nativebpm/connectors/camunda/docker/sequin-docker-compose/playground.yml)
*(Уже изменено)* SQL-функция обогащения `camunda_enrichment`, собирающая переменные и бизнес-ключ.

---

### 3. Ядро коннектора (Camunda SDK)

#### [MODIFY] [camunda_sequin.go](file:///Users/user/github.com/nativebpm/connectors/camunda/camunda_sequin.go)
Доработать `SequinWorker`:
- Добавить поля `sequinURL`, `token` и `httpClient`.
- Изменить метод `receiveMessages`, чтобы он отправлял raw POST запросы к Sequin и парсил полный JSON (включая `metadata.enrichment`).
- В `processMessage`, если метаданные обогащения заполнены:
  - Использовать `Variables` и `BusinessKey` из CDC-сообщения.
  - Пропускать шаги REST API `Lock`, `GetExecutionVariables` и `GetProcessInstanceBusinessKey`.
- В противном случае выполнять классические REST API вызовы.

---

### 4. Код примера (Go Application)

#### [MODIFY] [main.go](file:///Users/user/github.com/nativebpm/connectors/camunda/examples/loan-granting-cdc-outbox/main.go)
- Полностью удалить локальный `OutboxWorker` и все его вспомогательные структуры.
- Заменить на использование стандартного `camunda.NewSequinWorker`.

---

## Verification Plan

### Automated Tests
Убедиться, что проект компилируется:
```bash
go build ./camunda/...
```

### Manual Verification
1. Перезапустить docker-окружение:
   ```bash
   cd camunda
   make camunda
   ```
2. Запустить пример `loan-granting-cdc-outbox`:
   ```bash
   cd camunda/examples/loan-granting-cdc-outbox
   go run main.go
   ```
3. Проверить в логах успешную обработку задач:
   - Отсутствие HTTP-запросов к Camunda REST API на `/lock` или `/variables`.
   - Наличие ровно одного HTTP-запроса на задачу: `POST /external-task/{id}/complete`.
