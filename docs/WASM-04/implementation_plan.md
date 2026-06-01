# Implementation Plan - Разделение Camunda и Temporal примеров

Этот план описывает процесс разделения демонстрационного примера `camunda-temporal` на два независимых примера: `temporal` и `camunda`. Это позволит продемонстрировать чистую интеграцию WebAssembly Durable Execution Engine с каждой из систем оркестрации отдельно.

## User Review Required

> [!IMPORTANT]
> - **Схема BPMN для Camunda**: Новый пример `camunda` будет содержать очищенную BPMN-схему [process.bpmn](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/camunda/bpmn/process.bpmn) только с шагами для Camunda External Task.
> - **Гибридный подход**: Оба примера продолжат использовать рекомендованный гибридный подход: временный снапшот памяти воркера во время выполнения транзакции и сброс финальных данных в классическую БД (Mock API/файл JSON) по завершении с очисткой снапшота.

## Proposed Changes

Мы удалим пример `camunda-temporal` и создадим вместо него две независимые директории в `durable-wasm/examples/`.

---

### Root Workspace & Build Configuration

#### [MODIFY] [go.work](file:///Users/user/github.com/nativebpm/connectors/go.work)
* Удалить старые пути:
  - `./durable-wasm/examples/camunda-temporal/host`
  - `./durable-wasm/examples/camunda-temporal/worker`
* Добавить новые пути:
  - `./durable-wasm/examples/temporal/host`
  - `./durable-wasm/examples/temporal/worker`
  - `./durable-wasm/examples/camunda/host`
  - `./durable-wasm/examples/camunda/worker`

#### [MODIFY] [Makefile](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/Makefile)
* Заменить цель `run-camunda-temporal-example` на `run-temporal-example` и `run-camunda-example`.
* Обновить цель `clean` для очистки артефактов в новых папках.

---

### Component: Temporal Activity Example (`examples/temporal`)

Демонстрирует запуск WASM-кода в качестве долгоживущей Activity в Temporal, поддерживающей промежуточные чекпоинты (Heartbeats / Progress Restoration).

#### [NEW] [Makefile](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/temporal/Makefile)
* Автоматизация сборки `worker/worker.wasm` через TinyGo и запуска `host/main.go`.

#### [NEW] [worker/go.mod](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/temporal/worker/go.mod)
* Go module для TinyGo воркера.

#### [NEW] [worker/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/temporal/worker/main.go)
* Специфичный для Temporal воркер. Реализует многошаговые тяжелые вычисления (например, скоринг или процессинг):
  - Step 0: Инициализация Activity.
  - Step 1: Вызов внешнего API для скачивания расчетных параметров с чекпоинтом.
  - Step 2: Выполнение тяжелой калькуляции с чекпоинтом.
  - Step 3: Сохранение финальных результатов калькуляции в БД хоста (гибридный подход) с чекпоинтом.
  - Step 4: Завершение.

#### [NEW] [host/go.mod](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/temporal/host/go.mod)
* Go module для Go хоста, импортирующий библиотеку `durable`.

#### [NEW] [host/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/temporal/host/main.go)
* Эмулирует Temporal Activity Runner. 
* Запускает WASM, имитирует краш на первом шаге, сохраняет снапшот `temporal-activity.bin`.
* Перезапускает воркер, восстанавливает состояние из снапшота, выполняет до конца.
* Сохраняет финальную запись в `database_temporal.json`, удаляет временный снапшот.

#### [NEW] [.gitignore](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/temporal/.gitignore)
* Игнорирование `.bin`, `.json` (кроме кода) и `worker.wasm`.

---

### Component: Camunda External Task Example (`examples/camunda`)

Демонстрирует запуск WASM-кода в качестве воркера для Camunda External Task, управляемого схемой BPMN.

#### [NEW] [Makefile](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/camunda/Makefile)
* Автоматизация сборки и запуска.

#### [NEW] [bpmn/process.bpmn](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/camunda/bpmn/process.bpmn)
* BPMN-схема процесса Camunda. Содержит одну Service Task типа `external` с топиком `durable-wasm-task` (вместо разделения на шаги, которые были в предыдущей схеме).

#### [NEW] [worker/go.mod](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/camunda/worker/go.mod)
* Go module для воркера.

#### [NEW] [worker/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/camunda/worker/main.go)
* TinyGo воркер, реализующий логику обработки одной задачи Camunda в несколько шагов:
  - Step 0: Инициализация задачи.
  - Step 1: Проверка доступности товаров (Inventory Check) через HTTP API.
  - Step 2: Списание средств (Payment Capture) через HTTP API.
  - Step 3: Сохранение статуса заказа в Master-БД.
  - Step 4: Завершение и возврат флага успеха в Camunda.

#### [NEW] [host/go.mod](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/camunda/host/go.mod)
* Go module для хоста.

#### [NEW] [host/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/camunda/host/main.go)
* Эмулирует Camunda External Task Client. Поддерживает Mock API-сервер для инвентаря, платежей и БД.
* Выполняет прогон с симуляцией падения хоста, восстановлением памяти из `camunda-task.bin`, записью в `database_camunda.json` и очисткой снапшота.

#### [NEW] [.gitignore](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/camunda/.gitignore)

---

### [DELETE] [camunda-temporal](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/camunda-temporal)
* Полное удаление старой директории объединенного примера.

---

## Verification Plan

### Automated Steps
1. Запуск сборки и тестов обоих примеров:
   ```bash
   make -C durable-wasm/examples/temporal run
   make -C durable-wasm/examples/camunda run
   ```
2. Проверка, что:
   - Снапшоты создаются при сбое.
   - Восстановление памяти проходит корректно (данные не теряются).
   - Базы данных `database_temporal.json` и `database_camunda.json` создаются с финальным результатом.
   - Снапшоты удаляются после успешного завершения.

### Manual Steps
* Проверка компиляции TinyGo и форматирования кода с помощью `go fmt`.
