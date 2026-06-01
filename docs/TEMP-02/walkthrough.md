# TEMP-02: Результаты демонстрации Activity Heartbeats

В рамках этой задачи разработан и успешно протестирован демонстрационный пример для Activity Heartbeats с сохранением и восстановлением состояния.

## Реализованные файлы

1. **Workflow**: [workflow.go](file:///Users/user/github.com/nativebpm/connectors/temporal/examples/heartbeat/workflow.go)
   - Конфигурирует `ActivityOptions` с `HeartbeatTimeout: 2 * time.Second` и политикой повторных попыток.
2. **Activity**: [activity.go](file:///Users/user/github.com/nativebpm/connectors/temporal/examples/heartbeat/activity.go)
   - Реализует 10 пошаговых итераций.
   - Каждую итерацию отправляет `activity.RecordHeartbeat`.
   - На первой попытке имитирует зависание на 4 секунды (превышая `HeartbeatTimeout`).
   - На второй попытке восстанавливает прогресс через `activity.GetHeartbeatDetails`.
3. **Worker**: [main.go](file:///Users/user/github.com/nativebpm/connectors/temporal/examples/heartbeat/worker/main.go)
   - Запуск воркера, регистрирующего разработанные Workflow и Activity.
4. **Starter**: [main.go](file:///Users/user/github.com/nativebpm/connectors/temporal/examples/heartbeat/starter/main.go)
   - Инициализация и запуск Workflow.
5. **Makefile**: [Makefile](file:///Users/user/github.com/nativebpm/connectors/temporal/Makefile)
   - Добавлены команды `run-worker-heartbeat` и `run-starter-heartbeat`.

---

## Анализ логов выполнения

Ниже приведены логи воркера при запуске примера:

```text
2026/06/02 01:18:08 Воркер heartbeat успешно запущен для Task Queue: default-task-queue
2026/06/02 01:18:15 DEBUG ExecuteActivity ... Attempt 1 ActivityID 5 ActivityType HeartbeatActivity

// Attempt 1: Начинается выполнение с 1-го шага
2026/06/02 01:18:15 INFO  Starting activity processing ... StartStep 1 TotalSteps 10 Attempt 1
2026/06/02 01:18:15 INFO  Processing step ... Step 1 Attempt 1
2026/06/02 01:18:16 INFO  Heartbeat recorded successfully ... CompletedStep 1 Attempt 1
2026/06/02 01:18:16 INFO  Processing step ... Step 2 Attempt 1
2026/06/02 01:18:17 INFO  Heartbeat recorded successfully ... CompletedStep 2 Attempt 1
2026/06/02 01:18:17 INFO  Processing step ... Step 3 Attempt 1
2026/06/02 01:18:18 INFO  Heartbeat recorded successfully ... CompletedStep 3 Attempt 1
2026/06/02 01:18:18 INFO  Processing step ... Step 4 Attempt 1

// Имитация зависания на Attempt 1: Воркер "засыпает" на 4 секунды, не присылая Heartbeat
2026/06/02 01:18:19 WARN  [SIMULATION] Freezing worker on Attempt 1 at step 4 (sleeping 4s without heartbeating)...

// Attempt 2: Поскольку HeartbeatTimeout = 2s, сервер считает попытку 1 упавшей и запускает попытку 2
// Попытка 2 находит детали предыдущего успешного Heartbeat (CompletedStep 3) и возобновляет работу с шага 4!
2026/06/02 01:18:22 INFO  Found heartbeat details. Resuming progress ... CompletedStep 3 Attempt 2
2026/06/02 01:18:22 INFO  Starting activity processing ... StartStep 4 TotalSteps 10 Attempt 2
2026/06/02 01:18:22 INFO  Processing step ... Step 4 Attempt 2

// Попытка 1 просыпается спустя 4 секунды и пытается отправить Heartbeat, но сервер отвечает ошибкой,
// так как Activity для этой попытки уже признана завершившейся по таймауту.
2026/06/02 01:18:23 WARN  RecordActivityHeartbeat with error ... Error invalid activityID or activity already timed out
2026/06/02 01:18:23 WARN  Activity context was cancelled ... Attempt 1
2026/06/02 01:18:23 ERROR Activity error. ... Attempt 1 Error context canceled

// Попытка 2 успешно продолжает выполнение до конца
2026/06/02 01:18:23 INFO  Heartbeat recorded successfully ... CompletedStep 4 Attempt 2
2026/06/02 01:18:23 INFO  Processing step ... Step 5 Attempt 2
...
2026/06/02 01:18:29 INFO  Heartbeat recorded successfully ... CompletedStep 10 Attempt 2
```

Workflow завершается успешно, возвращая результат: `All 10 steps completed successfully on attempt 2!`.
