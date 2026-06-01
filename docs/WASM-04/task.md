# Checklist - WASM Durable Execution Engine MVP - Refactoring

- `[x]` Инициализация структуры проекта и регистрация модулей в `go.work`
- `[x]` Установка `tinygo` (если отсутствует) и настройка `Makefile`
- `[x]` Разработка WASM Worker в `durable-wasm/worker`
- `[x]` Разработка Go Host с интеграцией `wasmtime-go` и хост-функциями в `durable-wasm/host`
- `[x]` Реализация механизма snapshot / restore
- `[x]` Рефакторинг хоста в переиспользуемый пакет `durable`
- `[x]` Разработка интеграционного сценария (симуляция сбоя и восстановления)
- `[x]` Подготовка `Dockerfile` на базе `scratch`
- `[x]` Проверка и верификация (запуск `make run`, тесты)
- `[ ]` Рефакторинг: Удаление папки `camunda-temporal`
- `[ ]` Рефакторинг: Создание примера для Temporal Activity (`examples/temporal`)
- `[ ]` Рефакторинг: Создание примера для Camunda External Task (`examples/camunda`)
- `[ ]` Рефакторинг: Обновление корневого Makefile и go.work
- `[ ]` Рефакторинг: Локальный запуск и верификация примеров
