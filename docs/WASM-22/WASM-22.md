---
task: WASM-22
status: In Progress
summary: Переименование sdk.go и sdk_stub.go в fluent.go и fluent_stub.go, очистка упоминаний SDK
---

# WASM-22: Переименование файлов SDK в Fluent API

Необходимо переименовать файлы `sdk.go` и `sdk_stub.go` в пакете `wasman` во что-то более подходящее, так как они представляют собой Fluent API / Runner для WASM-воркеров, а не полноценный SDK.

## Требования:
1. Переименовать `connectors/wasman/sdk.go` -> `connectors/wasman/fluent.go`
2. Переименовать `connectors/wasman/sdk_stub.go` -> `connectors/wasman/fluent_stub.go`
3. Очистить или переименовать упоминания "SDK" в логах, сообщениях об ошибках и комментариях в этих файлах (например, использовать `wasman` или `fluent runner`).
4. Обновить документацию и комментарии, ссылающиеся на эти файлы.
