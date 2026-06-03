# CAM-03: Результаты реализации Camunda CDC с использованием Sequin Enrichment и авто-блокировки

## Выполненные изменения

1. **Ядро коннектора (Camunda SDK)**
   - Внедрена CDC-логика (zero-lookup/outbox) непосредственно в стандартный `camunda.SequinWorker` в [camunda_sequin.go](file:///Users/user/github.com/nativebpm/connectors/camunda/camunda_sequin.go).
   - Воркер выполняет raw POST запросы к эндпоинту `/receive` Sequin для получения и парсинга полной JSON-структуры сообщения, включая метаданные обогащения `metadata.enrichment`.
   - Добавлена адаптивная логика: если в сообщении CDC присутствует `metadata.enrichment`, воркер переключается в zero-lookup режим (переменные и бизнес-ключ извлекаются из сообщения, запросы `Lock`, `GetExecutionVariables` и `GetProcessInstanceBusinessKey` к Camunda REST API пропускаются).
   - Сохранена полная обратная совместимость: при отсутствии обогащенных метаданных воркер прозрачно откатывается на стандартные REST-запросы.

2. **База данных Camunda (Migrations)**
   - Создана миграция `20260603000000_add_task_autolock_trigger.sql`, добавляющая триггер авто-блокировки `lock_camunda_external_task_trigger` на таблицу `act_ru_ext_task`.
   - Логика триггера ограничена списком топиков кредитного процесса (`creditScoreChecker`, `decider`, `loanGranter`, `requestRejecter`), что предотвращает нежелательную блокировку задач в других процессах и интеграционных тестах.

3. **Код примера (Go Application)**
   - Пример `camunda/examples/loan-granting-cdc-outbox` переведен на использование стандартного `camunda.SequinWorker`. Весь дублирующийся код `OutboxWorker` полностью удален из `main.go`.

## Результаты тестирования

1. **Интеграционные тесты**
   - Все юнит- и интеграционные тесты пакета `camunda` (включая `TestCamundaIntegration_RealServer`, `TestBPMNVSRealCamunda` и `TestSequinWorker_AsyncDelegation`) успешно проходят.

2. **Сквозной тест (E2E)**
   - Запущенный пример успешно обработал 5 кредитных процессов. В логах воркера четко видна активация zero-lookup режима:
     `CDC zero-lookup mode activated: using enriched metadata...`
   - Все локальные переменные (такие как `score`) и бизнес-ключ были корректно получены напрямую из CDC-сообщений. Воркер выполнил ровно по одному REST-запросу `Complete` на каждую задачу.
