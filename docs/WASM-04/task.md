# Checklist - WASM Durable Execution Engine MVP

- `[x]` Инициализация структуры проекта и регистрация модулей в `go.work`
- `[x]` Установка `tinygo` (если отсутствует) и настройка `Makefile`
- `[x]` Разработка WASM Worker в `durable-wasm/worker`
- `[x]` Разработка Go Host с интеграцией `wasmtime-go` и хост-функциями в `durable-wasm/host`
- `[x]` Реализация механизма snapshot / restore
- `[x]` Рефакторинг хоста в переиспользуемый пакет `durable`
- `[x]` Разработка интеграционного сценария (симуляция сбоя и восстановления)
- `[ ]` Подготовка `Dockerfile` на базе `scratch`
- `[ ]` Проверка и верификация (запуск `make run`, тесты)
