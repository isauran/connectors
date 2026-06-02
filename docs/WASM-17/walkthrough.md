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

5. **Тестирование и примеры (Текущий этап)**:
   - В `bpmn/engine_test.go` добавлен тест `TestBPMNParallelGateway` для параллельного ветвления и слияния (Fork/Join).
   - Создан комплексный сквозной пример `bpmn/examples/orchestration`, объединяющий парсинг BPMN и DMN, вычисление DMN-правил на хосте и выполнение WASM-воркера.
   - Пример демонстрирует **Durable Execution**: симуляцию падения (crash) воркера на первом чекпоинте, сохранение состояния памяти на диск и последующее успешное возобновление работы из снапшота.

---

## Результаты тестирования

1. Все тесты модуля `bpmn` успешно пройдены:
```bash
$ go test -v ./...
=== RUN   TestBPMNParserAndEngine
--- PASS: TestBPMNParserAndEngine (0.00s)
=== RUN   TestDMNEvaluator
--- PASS: TestDMNEvaluator (0.00s)
=== RUN   TestBPMNParallelGateway
--- PASS: TestBPMNParallelGateway (0.00s)
PASS
ok  	github.com/nativebpm/connectors/bpmn	0.793s
```

2. Успешный лог выполнения примера `bpmn/examples/orchestration`:
```bash
$ make run
Building WASM worker for orchestration example using TinyGo...
tinygo build -o worker/worker.wasm -target=wasi worker/main.go
Running orchestration example host...
cd host && go run main.go
2026/06/03 03:44:03 INFO [HOST] Starting BPMN + DMN + Wasman Durable Orchestration Example
2026/06/03 03:44:03 INFO [HOST] Parsed BPMN process id=loan_process
2026/06/03 03:44:03 INFO [HOST] Parsed DMN decision table id=loan_decision
2026/06/03 03:44:03 INFO [HOST] Next step token to execute token=start
2026/06/03 03:44:03 INFO [BPMN ENGINE] StartEvent triggered node_id=start
2026/06/03 03:44:03 INFO [HOST] Next step token to execute token=check_rules
2026/06/03 03:44:03 INFO [HOST] Evaluating DMN rules for decision... age=25 income=1200
2026/06/03 03:44:03 INFO [HOST] DMN decision evaluated successfully result=map[approved:true]
2026/06/03 03:44:03 INFO [BPMN ENGINE] Executing ServiceTask node_id=check_rules name="Check Loan Rules"
2026/06/03 03:44:03 INFO [BPMN ENGINE] ServiceTask executed as noop (no WASM path configured) node_id=check_rules
2026/06/03 03:44:03 INFO [HOST] Next step token to execute token=is_approved
2026/06/03 03:44:03 INFO [BPMN ENGINE] ExclusiveGateway (XOR) reached node_id=is_approved
2026/06/03 03:44:03 INFO [HOST] Next step token to execute token=execute_payment
2026/06/03 03:44:03 INFO [BPMN ENGINE] Executing ServiceTask node_id=execute_payment name="Execute Payment Payout"
2026/06/03 03:44:03 INFO [BPMN ENGINE] Launching Wasman WASM session instance_id=loan-orchestration-instance wasm_path=../worker/worker.wasm server=127.0.0.1:62694
2026/06/03 03:44:03 INFO [ENGINE] Invoking entrypoint entrypoint=run
[WASM WORKER] Step 1: Loading variables from host...
2026/06/03 03:44:03 INFO [ENGINE] GET Request (Stream-first) url=http://127.0.0.1:62694/download
2026/06/03 03:44:03 INFO [ENGINE] GET Stream EOF. Closing response
[WASM WORKER] Received variables: map[age:25 approved:true income:1200 simulate_crash:true]
2026/06/03 03:44:03 INFO [ENGINE] 'checkpoint' invoked instance_id=loan-orchestration-instance
2026/06/03 03:44:03 INFO [ENGINE] Writing Full Memory Snapshot version=1
2026/06/03 03:44:03 WARN [ENGINE] Simulating host crash. Aborting WASM execution.
2026/06/03 03:44:03 WARN [HOST] Caught expected simulated WASM crash!
2026/06/03 03:44:03 INFO [HOST] Disabling crash flag and retrying execution to demonstrate Durable Recovery...
2026/06/03 03:44:03 INFO [HOST] Verified memory snapshot exists on disk. Resuming...
2026/06/03 03:44:03 INFO [HOST] Next step token to execute token=execute_payment
2026/06/03 03:44:03 INFO [BPMN ENGINE] Executing ServiceTask node_id=execute_payment name="Execute Payment Payout"
2026/06/03 03:44:03 INFO [BPMN ENGINE] Launching Wasman WASM session instance_id=loan-orchestration-instance wasm_path=../worker/worker.wasm server=127.0.0.1:62696
2026/06/03 03:44:03 INFO [ENGINE] Found saved full snapshot. Restoring memory... instance_id=loan-orchestration-instance
2026/06/03 03:44:03 INFO [ENGINE] Growing memory pages=2
2026/06/03 03:44:03 INFO [ENGINE] Memory successfully restored from full snapshot
2026/06/03 03:44:03 INFO [ENGINE] Invoking entrypoint entrypoint=run
[WASM WORKER] Step 2: Executing business logic...
[WASM WORKER] Processing payout payment...
2026/06/03 03:44:03 INFO [ENGINE] 'checkpoint' invoked instance_id=loan-orchestration-instance
2026/06/03 03:44:03 INFO [ENGINE] Memory deltas successfully saved dirty_blocks=5
[WASM WORKER] Step 3: Saving updated variables to host...
2026/06/03 03:44:03 INFO [ENGINE] POST Request (Stream-first via io.Pipe) url=http://127.0.0.1:62696/upload
2026/06/03 03:44:03 INFO [ENGINE] Closing upload stream (EOF). Waiting for response
2026/06/03 03:44:03 INFO [ENGINE] POST completed successfully
2026/06/03 03:44:03 INFO [ENGINE] 'checkpoint' invoked instance_id=loan-orchestration-instance
2026/06/03 03:44:03 INFO [ENGINE] Memory deltas successfully saved dirty_blocks=6
2026/06/03 03:44:03 INFO [ENGINE] Execution completed result=0
2026/06/03 03:44:03 INFO [BPMN ENGINE] Variables updated from WASM execution vars="map[age:25 approved:true income:1200 payment_status:success simulate_crash:true transaction_id:TXN-987654321]"
2026/06/03 03:44:03 INFO [HOST] Next step token to execute token=end
2026/06/03 03:44:03 INFO [BPMN ENGINE] EndEvent reached node_id=end
2026/06/03 03:44:03 INFO [BPMN ENGINE] Process instance completed successfully instance_id=loan-orchestration-instance
2026/06/03 03:44:03 INFO [HOST] Durable Orchestration completed successfully!
2026/06/03 03:44:03 INFO [HOST] Final process variables variables="map[age:25 approved:true income:1200 payment_status:success simulate_crash:true transaction_id:TXN-987654321]"
2026/06/03 03:44:03 INFO [HOST] Verification PASSED: Payout executed successfully via WASM worker after recovery!
2026/06/03 03:44:03 INFO [HOST] Payment status status=success txn=TXN-987654321
```

