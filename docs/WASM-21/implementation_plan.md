# WASM-21: Агрегированное S3-индексирование для визуализации процессов в Cockpit

Внедрение поддержки единого индексного файла активных процессов `instances/active_index.json` в S3-совместимое хранилище `SnapshotStore` и автоматическая синхронизация этого индекса при любых изменениях состояния токенов в `bpmn.Engine`.

## User Review Required

> [!IMPORTANT]
> - **Сетевой оверхед**: Синхронизация индекса требует записи в S3 при каждом шаге `Step` или завершении задачи. Для обеспечения атомарности и защиты от конкурентной записи (OCC) используется 5 попыток повтора (backoff retry). В высоконагруженных окружениях это может создавать небольшую задержку на запись, но исключает необходимость в базах данных.

## Proposed Changes

---

### [Component: wasman]

#### [MODIFY] [types.go](file:///Users/user/github.com/nativebpm/connectors/wasman/types.go)
Добавить методы для работы с индексом в интерфейс `SnapshotStore`:
```go
	// Active Index for Cockpit visualization
	UpdateActiveIndex(id string, info []byte, completed bool) error
	LoadActiveIndex() ([]byte, error)
```

#### [MODIFY] [wasman.go](file:///Users/user/github.com/nativebpm/connectors/wasman/wasman.go)
Добавить публичный геттер для получения `store` из `Engine`:
```go
// Store returns the SnapshotStore associated with the Engine
func (e *Engine) Store() SnapshotStore {
	return e.store
}
```

#### [MODIFY] [s3_store.go](file:///Users/user/github.com/nativebpm/connectors/wasman/s3_store.go)
Реализовать методы `UpdateActiveIndex` (с OCC-логикой и `If-Match`) и `LoadActiveIndex` для S3-хранилища.

#### [MODIFY] [fs_store.go](file:///Users/user/github.com/nativebpm/connectors/wasman/fs_store.go)
Реализовать методы `UpdateActiveIndex` и `LoadActiveIndex` для локального файлового хранилища.

---

### [Component: bpmn]

#### [MODIFY] [engine.go](file:///Users/user/github.com/nativebpm/connectors/bpmn/engine.go)
1. Реализовать хелпер `syncIndex(instance *ProcessInstance)` для сериализации состояния (`active_tokens`, `waiting_tokens`, `completed`, `updated_at`) и обновления индекса через `SnapshotStore`.
2. Вызывать `syncIndex` в точках мутации состояния:
   - В конце `StartInstance`
   - В конце `Step` (перед успешным `return nil`)
   - В конце `CompleteTask` (перед успешным `return nil`)
   - В конце `CorrelateMessage` (перед успешным `return nil`)
   - В конце `BroadcastSignal` (перед успешным `return nil`)
   - В конце `HandleError` (перед успешным `return nil`)

---

## Verification Plan

### Automated Tests
1. Написать новые unit-тесты в `wasman/engine_test.go` для проверки корректности работы `UpdateActiveIndex` (добавление, обновление, удаление завершенных инстансов, разрешение OCC коллизий).
2. Запустить все тесты модулей `wasman` и `bpmn`:
   ```bash
   cd wasman && go test -v ./...
   cd ../bpmn && go test -v ./...
   ```

### Manual Verification
Запустить один из примеров в `bpmn/examples/worker` или `bpmn/examples/parallel_mi` и убедиться, что в каталоге хранилища (файловом или S3) появляется файл `active_index.json` с актуальным списком токенов.
