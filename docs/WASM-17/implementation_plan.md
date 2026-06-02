# WASM-17: Создание модуля bpmn для BPMN 2.0 и DMN движков

Этот план описывает шаги по созданию нового Go-модуля `bpmn` в монорепозитории. Новый модуль будет реализовывать парсинг и интерпретацию BPMN 2.0 схем и DMN таблиц решений, используя разработанный движок `wasman` в качестве Durable WASM SDK для выполнения шагов процесса.

## User Review Required

> [!IMPORTANT]
> Новый модуль `bpmn` использует локальную зависимость от модуля `wasman` через `go.work`. При добавлении примеров мы скомпилируем WASM-воркеры с использованием Go 1.26 под `GOOS=js GOARCH=wasm` (или `GOOS=wasip1 GOARCH=wasm`) и настроим их запуск из хост-процесса.

## Proposed Changes

### 1. Конфигурация монорепозитория

#### [MODIFY] [go.work](file:///Users/user/github.com/nativebpm/connectors/go.work)
Добавление пути `./bpmn` в список используемых модулей. *(Выполнено)*

#### [NEW] [go.mod](file:///Users/user/github.com/nativebpm/connectors/bpmn/go.mod)
Создание `go.mod` для модуля `bpmn`. *(Выполнено)*

---

### 2. Разработка ядра BPMN 2.0

#### [NEW] [parser.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/parser.go)
Парсинг BPMN 2.0 XML в Go-структуры. *(Выполнено)*

#### [NEW] [engine.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/engine.go)
Движок выполнения процессов (Process Engine) с поддержкой шлюзов XOR и Parallel Gateways, а также интеграции с WASM-сессиями через динамический HTTP-сервер. *(Выполнено)*

---

### 3. Разработка DMN движка

#### [NEW] [dmn.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/dmn.go)
Парсинг и исполнение таблиц решений DMN с поддержкой Hit Policies (Unique, First, Any). *(Выполнено)*

---

### 4. Тестирование и примеры (Текущий этап)

#### [MODIFY] [engine_test.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/engine_test.go)
Расширение тестов в `bpmn/`:
- Добавление теста для Parallel Gateway Fork/Join.
- Добавление теста для комплексного процесса с использованием переменных, шлюзов и обработкой ошибок DMN.

#### [NEW] `bpmn/examples/orchestration`
Создание сквозного примера применения BPMN + DMN + Wasman:
- [NEW] [process.bpmn](file:///Users/user/github.com/nativebpm/connectors/bpmn/examples/orchestration/process.bpmn): BPMN XML-схема кредитного конвейера с шагом проверки правил (DMN) и WASM-воркерами (Wasman).
- [NEW] [decision.dmn](file:///Users/user/github.com/nativebpm/connectors/bpmn/examples/orchestration/decision.dmn): DMN-таблица с правилами принятия решения по кредиту на основе возраста и доходов.
- [NEW] [host/main.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/examples/orchestration/host/main.go): Go-хост, который запускает `bpmn.Engine`, выполняет шаги, эмулирует падение (crash) в процессе работы WASM-воркера, восстанавливает сессию из снапшота и продолжает исполнение.
- [NEW] [worker/main.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/examples/orchestration/worker/main.go): Go-код воркера, компилируемый в WASM, реализующий шаги процесса с помощью Durable SDK `wasman`.
- [NEW] [Makefile](file:///Users/user/github.com/nativebpm/connectors/bpmn/examples/orchestration/Makefile): Скрипт сборки воркера в `.wasm` и запуска примера.
- [NEW] [README.md](file:///Users/user/github.com/nativebpm/connectors/bpmn/examples/orchestration/README.md): Описание работы примера.

---

## Verification Plan

### Automated Tests
1. Запуск тестов модуля `bpmn`:
   ```bash
   cd bpmn && go test -v ./...
   ```
2. Сборка и запуск нового примера:
   ```bash
   cd bpmn/examples/orchestration && make build && make run
   ```
