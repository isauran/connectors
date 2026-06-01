# Implementation Plan - Объединение в единый Go-модуль durable-wasm

Этот план описывает шаги по объединению отдельных модулей `durable-wasm/host` и `durable-wasm/worker` в единый Go-модуль `github.com/nativebpm/connectors/durable-wasm` на уровне корня каталога `durable-wasm/`. Это упростит публикацию библиотеки в публичный репозиторий и ее импорт другими проектами (потребуется один Git-тег вида `durable-wasm/vX.Y.Z`).

## User Review Required

> [!IMPORTANT]
> - **Удаление отдельных модулей**: Модули `github.com/nativebpm/connectors/durable-wasm/host` и `github.com/nativebpm/connectors/durable-wasm/worker` будут удалены и заменены единым модулем `github.com/nativebpm/connectors/durable-wasm`.
> - **Совместимость импорта**: Все пути импорта пакетов (например, `github.com/nativebpm/connectors/durable-wasm/host/durable`) в коде примеров останутся неизменными, так как подкаталоги `host/` и `worker/` теперь станут частью общего модуля. Понадобится обновить только файлы `go.mod` примеров.

## Proposed Changes

Мы настроим единый `go.mod` в корне `durable-wasm` и обновим конфигурацию рабочей области и примеров.

---

### Root Module Configuration

#### [NEW] [go.mod](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/go.mod)
* Создать новый файл модуля `github.com/nativebpm/connectors/durable-wasm` в корне директории `durable-wasm`.
* Объявить в нем зависимости от `wasmtime-go/v20` и `testify` (перенос зависимостей из старого `host/go.mod`).

#### [DELETE] [go.mod](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/host/go.mod)
* Удалить старый файл модуля хоста.

#### [DELETE] [go.mod](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/worker/go.mod)
* Удалить старый файл модуля воркера.

---

### Workspace Configuration

#### [MODIFY] [go.work](file:///Users/user/github.com/nativebpm/connectors/go.work)
* Заменить записи `./durable-wasm/host` и `./durable-wasm/worker` на `./durable-wasm`.

---

### Examples Configuration

#### [MODIFY] [go.mod](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/camunda/host/go.mod)
* Заменить `github.com/nativebpm/connectors/durable-wasm/host v0.0.0` на `github.com/nativebpm/connectors/durable-wasm v0.0.0`.
* Заменить replace-директиву на `github.com/nativebpm/connectors/durable-wasm => ../../../`.

#### [MODIFY] [go.mod](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/temporal/host/go.mod)
* Заменить `github.com/nativebpm/connectors/durable-wasm/host v0.0.0` на `github.com/nativebpm/connectors/durable-wasm v0.0.0`.
* Заменить replace-директиву на `github.com/nativebpm/connectors/durable-wasm => ../../../`.

#### [MODIFY] [go.mod](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/process-csv/host/go.mod)
* Заменить `github.com/nativebpm/connectors/durable-wasm/host v0.0.0` на `github.com/nativebpm/connectors/durable-wasm v0.0.0`.
* Заменить replace-директиву на `github.com/nativebpm/connectors/durable-wasm => ../../../`.

#### [MODIFY] [go.mod](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/gotenberg-telegram/host/go.mod)
* Заменить `github.com/nativebpm/connectors/durable-wasm/host v0.0.0` на `github.com/nativebpm/connectors/durable-wasm v0.0.0`.
* Заменить replace-директиву на `github.com/nativebpm/connectors/durable-wasm => ../../../`.

---

## Verification Plan

### Automated Tests
1. Запустить обновление зависимостей:
   ```bash
   go work sync
   make tidy
   ```
2. Запустить unit-тесты ядра:
   ```bash
   make -C durable-wasm test
   ```
3. Запустить примеры для проверки работоспособности:
   ```bash
   make -C durable-wasm/examples/temporal run
   make -C durable-wasm/examples/camunda run
   ```

### Manual Verification
* Проверить отсутствие ошибок компиляции и конфликтов зависимостей.
