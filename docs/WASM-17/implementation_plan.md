# WASM-17: Создание модуля bpmn для BPMN 2.0 и DMN движков

Этот план описывает шаги по созданию нового Go-модуля `bpmn` в монорепозитории. Новый модуль будет реализовывать парсинг и интерпретацию BPMN 2.0 схем и DMN таблиц решений, используя разработанный движок `wasman` в качестве Durable WASM SDK для выполнения шагов процесса.

## User Review Required

> [!IMPORTANT]
> Новый модуль `bpmn` будет использовать локальную зависимость от модуля `wasman` через `go.work`. Для работы с внешними модулями потребуется обновить `go.work` и синхронизировать зависимости.

## Proposed Changes

### 1. Конфигурация монорепозитория

#### [MODIFY] [go.work](file:///Users/user/github.com/nativebpm/connectors/go.work)
Добавление пути `./bpmn` в список используемых модулей.

#### [NEW] [go.mod](file:///Users/user/github.com/nativebpm/connectors/bpmn/go.mod)
Создание `go.mod` для модуля `bpmn` с зависимостью от:
- `github.com/nativebpm/connectors/wasman v0.0.1` (локальная замена через go workspace)
- стандартных библиотек парсинга XML

---

### 2. Разработка ядра BPMN 2.0

#### [NEW] [parser.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/parser.go)
Парсинг BPMN 2.0 XML в Go-структуры:
- Структуры: `Definitions`, `Process`, `SequenceFlow`, `StartEvent`, `EndEvent`, `ServiceTask`, `ParallelGateway`, `ExclusiveGateway`.
- Метод `ParseBPMN(xmlData []byte) (*Definitions, error)`.

#### [NEW] [engine.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/engine.go)
Движок выполнения процессов (Process Engine):
- Модель состояния экземпляра процесса: `ProcessInstance` (содержит ID, ID схемы, текущие активные токены/узлы, переменные процесса `Variables`).
- Логика обхода графа: выполнение шагов, ветвление на шлюзах (Exclusive/Parallel Gateways).
- Интеграция с `wasman`: если шаг процесса — это `ServiceTask` с исполнением в WASM (например, с указанием пути к `.wasm` файлу), движок запускает сессию `wasman`, передает туда текущие переменные процесса, выполняет воркер, считывает результат и фиксирует checkpoint.
- Интерфейс для сохранения общего состояния экземпляра BPMN (состояние токенов и переменных) в `wasman.SnapshotStore` для обеспечения полной отказоустойчивости не только WASM-памяти, но и всего процесса оркестрации.

---

### 3. Разработка DMN движка

#### [NEW] [dmn.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/dmn.go)
Парсинг и исполнение таблиц решений DMN:
- Структуры: `Definitions` (DMN), `Decision`, `DecisionTable`, `Input`, `Output`, `Rule`.
- Метод `ParseDMN(xmlData []byte) (*DMN, error)`.
- Метод `Evaluate(decisionID string, variables map[string]interface{}) (map[string]interface{}, error)` для сопоставления входных данных с правилами таблицы (Hit Policy: Unique, Any, First).

---

### 4. Тестирование движка

#### [NEW] [engine_test.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/engine_test.go)
Автоматические тесты:
- Тест парсинга BPMN XML схемы с разветвлениями.
- Тест выполнения BPMN-процесса с мок-запуском WASM-шага через `wasman`.
- Тест вычисления правил DMN таблицы.

---

## Verification Plan

### Automated Tests
1. Инициализация модуля и синхронизация зависимостей:
   ```bash
   make tidy
   ```
2. Запуск тестов нового модуля:
   ```bash
   cd bpmn && go test -v ./...
   ```
