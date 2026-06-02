# Walkthrough - WASM-12: Durable WASM Stability Improvements

Мы успешно устранили риски утечек ресурсов (TCP-сокеты, сетевые горутины и C-память Wasmtime) в пакете `durable-wasm` и провели всестороннее стресс-тестирование движка.

## Внесенные изменения

### 1. Переиспользуемый HTTP-клиент
- Реализована опция `EngineOption` с конструктором `WithHTTPClient`.
- Настроен глобальный `defaultHTTPClient` по умолчанию с оптимизированными параметрами пула соединений (`MaxIdleConns: 100`, `MaxIdleConnsPerHost: 10`, `IdleConnTimeout: 90s`) для предотвращения исчерпания сокетов.

### 2. Поддержка context.Context
- Изменена сигнатура `Engine.Execute` — теперь первым аргументом передается `context.Context`.
- Контекст выполнения проброшен во все сетевые методы: стриминг скачивания (`handleDownload`), стриминг загрузки (`handleUpload`) и стандартные HTTP-запросы API (`host_call_api`).
- Добавлена проверка отмены контекста (`ctx.Err() != nil`) сразу после возврата из WASM вызова `runFunc.Call(store)`, чтобы незамедлительно вернуть `context.Canceled`.

### 3. Освобождение C-ресурсов Wasmtime
- Добавлен `defer store.Close()` внутри `Execute` для явного высвобождения памяти Wasmtime, выделенной через cgo.

### 4. Оптимизация SQLite для конкурентного доступа
- Для `SqliteSnapshotStore` установлено ограничение `db.SetMaxOpenConns(1)`, чтобы сериализовать параллельные записи и предотвратить ошибки `database is locked (SQLITE_BUSY)`.
- Разделены вызовы PRAGMA команд внутри `NewSqliteSnapshotStore` для корректного выполнения драйвером.

### 5. Адаптация примеров хост-приложений
- Все хост-приложения примеров обновлены под новую сигнатуру `Execute`:
  - `examples/camunda/host/main.go` и `main_test.go`
  - `examples/gotenberg-telegram/host/main.go` and `main_test.go`
  - `examples/process-csv/host/main.go` and `main_test.go`
  - `examples/temporal/host/main.go`
  - `examples/durable-s3/host/main.go`

### 6. Новые тесты стабильности и стресс-тестирования
- В `engine_test.go` добавлены три класса тестов:
  - `TestExecuteCancellation`: Проверяет быструю отмену выполнения и освобождение ресурсов при отмене контекста.
  - `TestStorageErrorInjection`: Эмулирует сбои SnapshotStore при сохранении метаданных и снапшотов.
  - `TestSoakStressTesting`: Конкурентный запуск 200 инстансов WASM на временной дисковой БД для симуляции реальной нагрузки.

## Результаты тестирования

Все тесты в модуле `durable-wasm` успешно пройдены:
```bash
go test -v ./...
```
Тесты отмены, инъекции ошибок, стресс-тест и все 5 интеграционных тестов примеров завершились успешно.
