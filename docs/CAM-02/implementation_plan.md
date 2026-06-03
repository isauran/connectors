# CAM-02: Полноценный пример с Sequin CDC воркером

Создание нового полноценного примера `camunda/examples/loan-granting-cdc`, который демонстрирует работу с бизнес-процессом кредитного конвейера с помощью Change Data Capture (CDC) воркера (`camunda.SequinWorker`) на основе Sequin.

## User Review Required

> [!IMPORTANT]
> В ходе тестирования CDC-воркера были обнаружены два критических ограничения в текущей реализации `SequinWorker` в SDK:
> 1. **Отсутствие локальных переменных**: Воркер запрашивал только глобальные переменные процесса (`/process-instance/{id}/variables`). Из-за этого переменные локальной области видимости (такие как `score` внутри параллельного цикла Multi-Instance) не попадали в обработчик, что приводило к ошибке `score variable not found in task`.
> 2. **Отсутствие BusinessKey**: Таблица `act_ru_ext_task` в базе данных Camunda не содержит колонку `business_key_`. Соответственно, Sequin присылал пустое значение, и обработчики не могли найти данные в in-memory хранилище по бизнес-ключу.
>
> **Предлагаемое решение**:
> - Расширить Camunda REST клиент методами для получения переменных по Execution ID (`/execution/{id}/variables`) и получения самого процесса для извлечения `businessKey`.
> - Обновить `SequinWorker` для использования этих новых методов.

## Open Questions

Нет открытых вопросов.

## Proposed Changes

---

### Ядро SDK (Camunda SDK)

#### [MODIFY] [camunda.go](file:///Users/user/github.com/nativebpm/connectors/camunda/camunda.go)
Добавить два новых метода в `Client`:
1. `GetExecutionVariables(ctx, executionID)` — для загрузки как глобальных, так и локальных переменных задачи.
2. `GetProcessInstanceBusinessKey(ctx, processInstanceID)` — для загрузки `businessKey` процесса по его ID.

#### [MODIFY] [camunda_sequin.go](file:///Users/user/github.com/nativebpm/connectors/camunda/camunda_sequin.go)
Обновить логику `processMessage`:
- Использовать `GetExecutionVariables` вместо `GetProcessVariables` для правильной загрузки переменной `score` из Multi-Instance подпроцесса.
- Запрашивать `businessKey` через `GetProcessInstanceBusinessKey` и подставлять его в `ExternalTask.BusinessKey`.

---

### Примеры использования (Examples)

#### [NEW] [main.go](file:///Users/user/github.com/nativebpm/connectors/camunda/examples/loan-granting-cdc/main.go)
Создать файл точки входа для примера. Он будет выполнять:
- Загрузку схемы процесса `bpmn/loan-granting.bpmn`.
- Настройку `camunda.SequinWorker`, подключенного к Sequin CDC-потоку.
- Регистрацию переиспользуемых обработчиков из `loan-granting/handlers`.
- Запуск 5 параллельных экземпляров кредитного процесса.
- Запуск воркера Sequin.

#### [NEW] [loan-granting.bpmn](file:///Users/user/github.com/nativebpm/connectors/camunda/examples/loan-granting-cdc/bpmn/loan-granting.bpmn)
Скопировать файл схемы BPMN из примера `loan-granting/bpmn/loan-granting.bpmn`.

#### [NEW] [README.md](file:///Users/user/github.com/nativebpm/connectors/camunda/examples/loan-granting-cdc/README.md)
Написать руководство по запуску примера, объясняющее разницу между REST-опросом и CDC-воркером.

---

## Verification Plan

### Automated Tests
Интеграционные тесты пакета `camunda` (включая работу с Sequin) запускаются командой:
```bash
cd camunda
go test -v .
```

### Manual Verification
1. Запуск инфраструктуры:
   ```bash
   cd camunda
   make camunda
   ```
2. Запуск нового примера:
   ```bash
   cd camunda/examples/loan-granting-cdc
   go run main.go
   ```
3. Проверка по логам, что задачи обрабатываются воркером через Sequin без ошибок отсутствия переменных или бизнес-ключа.
