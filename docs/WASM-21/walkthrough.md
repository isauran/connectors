# WASM-21 Walkthrough: Агрегированное S3-индексирование для Cockpit

Реализована поддержка единого индексного файла активных процессов `instances/active_index.json` в S3-совместимом хранилище и локальной файловой системе с авто-синхронизацией состояния токенов.

## Выполненные изменения

### 1. Модуль `wasman`
- **Интерфейс `SnapshotStore` ([types.go](file:///Users/user/github.com/nativebpm/connectors/wasman/types.go))**:
  Добавлены методы `UpdateActiveIndex(id string, info []byte, completed bool) error` и `LoadActiveIndex() ([]byte, error)`.
- **`Engine` ([wasman.go](file:///Users/user/github.com/nativebpm/connectors/wasman/wasman.go))**:
  Добавлен публичный метод `Store() SnapshotStore` для получения доступа к хранилищу из BPMN-движка.
- **`S3SnapshotStore` ([s3_store.go](file:///Users/user/github.com/nativebpm/connectors/wasman/s3_store.go))**:
  Реализованы методы `UpdateActiveIndex` и `LoadActiveIndex` с использованием OCC (Optimistic Concurrency Control, заголовки `If-Match` / `If-None-Match: "*"`) и экспоненциального бэкапа (до 5 попыток) при конфликтах параллельной записи.
  Инициализация `nextIndex := make([]map[string]interface{}, 0)` гарантирует запись корректного пустого JSON-массива `[]` (вместо `null`), когда в системе нет активных инстансов.
- **`FileSnapshotStore` ([fs_store.go](file:///Users/user/github.com/nativebpm/connectors/wasman/fs_store.go))**:
  Реализованы локальные версии методов для работы с файлом `active_index.json` в каталоге отладки.

### 2. Модуль `bpmn`
- **Синхронизация в `bpmn.Engine` ([engine.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/engine.go))**:
  - Реализован метод `syncIndex(instance *ProcessInstance)`, формирующий JSON с метаданными инстанса (`instance_id`, `process_id`, `active_tokens`, `waiting_tokens`, `completed`, `updated_at`) и отправляющий его в `SnapshotStore`.
  - Точки мутации состояния процесса обернуты в публичные методы, которые после выполнения оригинальной логики (переименованной в `*Internal` методы) автоматически вызывают `syncIndex`:
    - `StartInstance`
    - `Step` (обертка над `stepInternal`)
    - `CompleteTask` (обертка над `completeTaskInternal`)
    - `CorrelateMessage` (обертка над `correlateMessageInternal`)
    - `BroadcastSignal` (обертка над `broadcastSignalInternal`)
    - `TriggerCompensation` (обертка над `triggerCompensationInternal`)
    - `HandleError` (обертка над `handleErrorInternal`)

---

## Тестирование и верификация

### Автоматические тесты
1. **Тесты модуля `wasman`**:
   - Написаны новые unit-тесты `TestActiveIndex` в [engine_test.go](file:///Users/user/github.com/nativebpm/connectors/wasman/engine_test.go), проверяющие корректность добавления, обновления и удаления записей из `active_index.json` (как для in-memory заглушки, так и для реального `FileSnapshotStore`).
   - Добавлен шаг интеграционного тестирования `UpdateActiveIndex` на S3 с проверкой OCC в `TestS3SnapshotStore`.
   - Результат запуска: `PASS` (все тесты успешно пройдены).
2. **Тесты модуля `bpmn`**:
   - Добавлен комплексный unit-тест `TestBPMNEngineActiveIndexSync` в [engine_test.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/engine_test.go), проверяющий сквозную синхронизацию файла-индекса при прохождении токенов через события, XOR-шлюзы, пользовательские и принимающие задачи.
   - Результат запуска: `PASS`.

### Ручная проверка
Запущена демонстрационная программа `bpmn/examples/orchestration` с помощью `make run`. 
Успешно проверено:
- После `StartInstance` в `active_index.json` появился токен `start`.
- По ходу выполнения шагов `Step` токен в файле изменялся на `check_rules` -> `is_approved` -> `execute_payment` -> `end`.
- При падении WASM-воркера состояние сохранилось, индекс оставался актуальным. После возобновления и завершения процесса инстанс был полностью удален, а файл `active_index.json` принял корректный пустой вид `[]`.
