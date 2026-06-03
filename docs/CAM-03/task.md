# CAM-03 Checklist (Explicit Lock)

- [x] Удаление триггера авто-блокировки БД
  - [x] Дропнуть триггер и функцию в БД Camunda Postgres
  - [x] Удалить файл миграции `20260603000000_add_task_autolock_trigger.sql`
  - [x] Пересчитать хэши миграций Atlas (`make atlas-hash`)
  - [x] Сбросить и перезапустить docker-контейнеры Camunda (`docker compose down -v && docker compose up -d`)
  - [x] Накатить чистые миграции (`make atlas-apply`)
- [x] Доработка ядра коннектора (`camunda_sequin.go`)
  - [x] Изменить метод `processMessage` для выполнения явного `Lock` перед хэндлером во всех режимах (CDC и legacy)
- [x] Верификация сборки и тестов
  - [x] Собрать проект (`go build ./camunda/...`)
  - [x] Прогнать тесты (`go test -v ./camunda`)
- [x] Сквозное тестирование примера
  - [x] Запустить пример: `go run camunda/examples/loan-granting-cdc-outbox/main.go`
  - [x] Убедиться по логам воркера в успешной обработке задач (ровно два REST-запроса на задачу: `Lock` и `Complete`)
- [x] Документирование и коммит
  - [x] Отметить все задачи как выполненные в `docs/CAM-03/task.md`
  - [x] Обновить `docs/CAM-03/walkthrough.md`
  - [x] Сделать коммит
