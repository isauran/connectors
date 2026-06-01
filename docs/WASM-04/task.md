# Checklist - WASM Durable Execution Engine MVP

- `[x]` Инициализация структуры проекта и регистрация модулей в `go.work`
- `[ ]` Установка `tinygo` (если отсутствует) и настройка `Makefile`
- `[ ]` Разработка WASM Worker в `durable-wasm/worker`
- `[ ]` Разработка Go Host с интеграцией `wasmtime-go` и хост-функциями в `durable-wasm/host`
- `[ ]` Реализация механизма snapshot / restore
- `[ ]` Разработка интеграционного сценария (симуляция сбоя и восстановления)
- `[ ]` Подготовка `Dockerfile` на базе `scratch`
- `[ ]` Проверка и верификация (запуск `make run`, тесты)
