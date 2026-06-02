# WASM-17: Walkthrough — Разработка модуля bpmn и DMN-вычислителя

В рамках выполнения задачи в монорепозиторий был добавлен новый полноценный Go-модуль `bpmn` для оркестрации BPMN 2.0 процессов и вычисления DMN таблиц решений, использующий `wasman` для выполнения Durable WASM задач.

## Реализованные компоненты

1. **Инициализация и интеграция**:
   - Создана директория `bpmn/` и модуль `github.com/nativebpm/connectors/bpmn`.
   - Модуль подключен к Go-workspace в корневом файле `go.work`.
   - Внедрена зависимость от локального пакета `wasman`.

2. **Парсер BPMN 2.0 (`bpmn/parser.go`)**:
   - Реализована десериализация BPMN XML в структурированное дерево (StartEvent, EndEvent, ServiceTask, ExclusiveGateway, ParallelGateway, SequenceFlow).
   - Написана логика индексации элементов схемы процесса (`ParsedProcess`) для быстрого обхода графа.

3. **Движок выполнения BPMN (`bpmn/engine.go`)**:
   - Разработана модель обхода на основе токенов (Token-based execution) и переменных процесса (`Variables`).
   - Реализована маршрутизация Exclusive Gateway (XOR) на базе условных выражений (поддерживаются операции `==`, `!=`, `>`, `<`).
   - Добавлена логика Parallel Gateway (AND) с поддержкой разветвлений (Fork) и слияний (Join).
   - Интегрирована среда `wasman`: при достижении `ServiceTask` с путем к WASM-модулю движок поднимает локальный динамический HTTP-сервер для загрузки/выгрузки JSON-переменных, запускает сессию `wasman`, передает контекст переменных, выполняет воркер, фиксирует результат в переменных процесса и вызывает checkpoint.

4. **Вычислитель DMN (`bpmn/dmn.go`)**:
   - Поддерживает парсинг DMN XML таблиц решений.
   - Имплементирован алгоритм сопоставления правил (Rules Matching) с учетом числовых диапазонов (`<`, `>`) и строковых совпадений.
   - Реализованы политики совпадения (Hit Policies): `UNIQUE` (с проверкой на нарушение уникальности), `FIRST` и `ANY`.

---

## Результаты тестирования

Написан комплексный набор автоматических тестов в `bpmn/engine_test.go`:
- `TestBPMNParserAndEngine`:
  - Проверяет корректность разбора схемы.
  - Проверяет выполнение сквозного процесса через XOR-шлюз с переходами по веткам в зависимости от переменных процесса.
- `TestDMNEvaluator`:
  - Проверяет сопоставление правил DMN таблицы решений для различных возрастов (возврат `approved = true` / `false` на основе условий таблицы).

Все тесты успешно пройдены:
```bash
$ go test -v ./bpmn/...
=== RUN   TestBPMNParserAndEngine
2026/06/03 03:31:06 INFO [BPMN ENGINE] StartEvent triggered node_id=start
2026/06/03 03:31:06 INFO [BPMN ENGINE] Executing ServiceTask node_id=task1 name="Calculate Something"
2026/06/03 03:31:06 INFO [BPMN ENGINE] ServiceTask executed as noop (no WASM path configured) node_id=task1
2026/06/03 03:31:06 INFO [BPMN ENGINE] ExclusiveGateway (XOR) reached node_id=gateway
2026/06/03 03:31:06 INFO [BPMN ENGINE] EndEvent reached node_id=end_ok
2026/06/03 03:31:06 INFO [BPMN ENGINE] Process instance completed successfully instance_id=instance-1
...
--- PASS: TestBPMNParserAndEngine (0.00s)
=== RUN   TestDMNEvaluator
--- PASS: TestDMNEvaluator (0.00s)
PASS
ok  	github.com/nativebpm/connectors/bpmn	0.914s
```
