# Чек-лист задачи - WASM-13: Интеграция S3SnapshotStore

- `[x]` Добавить поле `ETag` в структуру `InstanceMeta` в `engine.go`
- `[x]` Реализовать `s3_store.go` с типом `S3SnapshotStore`, удовлетворяющим интерфейсу `SnapshotStore`
- `[x]` Написать интеграционный тест `TestS3SnapshotStore` в `engine_test.go`
- `[x]` Проверить компиляцию и запустить тесты с помощью `go test -v ./...`
- `[x]` Обновить отчеты и документацию (walkthroughs, README)
- `[x]` Закоммитить изменения в Git
