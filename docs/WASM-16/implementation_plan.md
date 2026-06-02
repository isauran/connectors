# WASM-16: Рефакторинг ядра Durable WASM и ребрендинг в Wasman

Этот план описывает шаги по декомпозиции монолитного файла `engine.go` в модуле `durable-wasm` и последующий полный ребрендинг модуля под имя **`wasman`**.

## User Review Required

> [!IMPORTANT]
> **Переименование модуля и пакета является ломающим изменением (breaking change)**. Все импорты в примерах и внешних проектах, ссылающиеся на `github.com/nativebpm/connectors/durable-wasm`, будут изменены на `github.com/nativebpm/connectors/wasman`. Пакет `durable` будет переименован в `wasman`.

## Open Questions

> [!NOTE]
> На данный момент открытых вопросов нет, так как финальное имя модуля — **`wasman`** — было согласовано с пользователем.

## Proposed Changes

### 1. Декомпозиция ядра Durable WASM

До переименования директорий выполним разделение монолитного `durable-wasm/engine.go` на логические составляющие.

#### [NEW] [types.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/types.go)
Определение базовых структур, ошибок и интерфейсов:
- Ошибки: `ErrWasmVersionMismatch`, `ErrConcurrentExecution`
- Структуры: `InstanceMeta`, `OplogEntry`, `Engine`, `Session`
- Интерфейс: `SnapshotStore`
- Конфигурация: `EngineOption`, `WithHTTPClient`
- Глобальные переменные: `defaultHTTPClient`

#### [NEW] [execution.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/execution.go)
Конструктор сессий и логика управления выполнением:
- Структура: `Execution`
- Методы: `Session()`, `WithEntrypoint()`, `WithServer()`, `WithCrash()`, `Run()`

#### [NEW] [session.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/session.go)
Реализация потоковой передачи ввода-вывода (Stream-first IO):
- Методы: `handleDownload()`, `handleUpload()`

#### [MODIFY] [engine.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/engine.go)
Будет содержать только:
- Конструктор `NewEngine()`
- Метод `Execute()` (инициализация Wasmtime, регистрация хост-функций, восстановление состояния из снапшотов/дельт)

---

### 2. Ребрендинг модуля и переименование

#### [DELETE] [durable-wasm](file:///Users/user/github.com/nativebpm/connectors/durable-wasm)
Переименование директории `durable-wasm` в `wasman` с помощью Git.

#### [NEW] [wasman](file:///Users/user/github.com/nativebpm/connectors/wasman)
Новое местоположение Go-модуля.

#### [MODIFY] [go.work](file:///Users/user/github.com/nativebpm/connectors/go.work)
Замена `./durable-wasm` на `./wasman`.

#### [MODIFY] [go.mod](file:///Users/user/github.com/nativebpm/connectors/wasman/go.mod)
Изменение имени модуля на `github.com/nativebpm/connectors/wasman`.

#### [MODIFY] Go source files in `wasman`
Изменение `package durable` на `package wasman` во всех файлах каталога `wasman/` (включая тесты и заглушки).

#### [MODIFY] Examples (`wasman/examples/...`)
Обновление всех импортов `"github.com/nativebpm/connectors/durable-wasm"` на `"github.com/nativebpm/connectors/wasman"` во всех файлах примеров:
- `examples/camunda/host/main.go`
- `examples/camunda/host/main_test.go`
- `examples/camunda/worker/main.go`
- `examples/gotenberg-telegram/host/main.go`
- `examples/gotenberg-telegram/host/main_test.go`
- `examples/gotenberg-telegram/worker/main.go`
- `examples/process-csv/host/main.go`
- `examples/process-csv/host/main_test.go`
- `examples/process-csv/worker/main.go`
- `examples/s3-store/host/main.go`
- `examples/s3-store/worker/main.go`
- `examples/temporal/host/main.go`
- `examples/temporal/host/main_test.go`
- `examples/temporal/worker/main.go`

#### [MODIFY] Root build files and docs
Обновление README.md, README.ru.md, Makefile и docs/*.md.

---

## Verification Plan

### Automated Tests
1. Очистка неструктурированного каталога `heartbeat` (если он есть).
2. Запуск тестов модуля `wasman`:
   ```bash
   cd wasman && go test -v ./...
   ```
3. Запуск тестов примеров (проверка компиляции и интеграции):
   ```bash
   make -C wasman run
   ```
   (или запуск тестов для отдельных примеров: `go test -v ./wasman/examples/camunda/host/...` и т.д.)
