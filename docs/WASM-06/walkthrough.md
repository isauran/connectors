# Walkthrough - Реструктуризация durable-wasm модуля

В рамках задачи **WASM-06** мы реструктурировали модуль `durable-wasm`, очистив его корень от демонстрационных файлов и перенеся их в директорию примеров.

## Изменения

### 1. Ядро движка (Public API)
* Файлы `engine.go` и `engine_test.go` перенесены непосредственно в корень модуля `durable-wasm/` под пакетом `durable`.
* Файл-обертка `durable.go` удален, так как все типы ядра теперь доступны потребителям библиотеки напрямую из корня модуля.
* Каталог `host/durable/` полностью удален.

### 2. Примеры использования (Examples)
* Создан новый пример `examples/simple/`, куда перенесены демонстрационные файлы песочницы:
  * `host/main.go` -> [examples/simple/host/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/simple/host/main.go)
  * `host/Dockerfile` -> [examples/simple/host/Dockerfile](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/simple/host/Dockerfile)
  * `worker/main.go` -> [examples/simple/worker/main.go](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/simple/worker/main.go)
* Создан [examples/simple/Makefile](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/examples/simple/Makefile) для сборки TinyGo-воркера и запуска примера.
* Созданы файлы `go.mod` для хоста и воркера песочницы для изоляции зависимостей.

### 3. Сборочные скрипты и Рабочее окружение
* Обновлен корневой [go.work](file:///Users/user/github.com/nativebpm/connectors/go.work): добавлены пути к новым Go-модулям `simple/host` и `simple/worker`.
* Обновлен [durable-wasm/Makefile](file:///Users/user/github.com/nativebpm/connectors/durable-wasm/Makefile): сборочные цели и запуск unit-тестов перенаправлены на новые пути.

## Результаты тестирования

* Запуск `make -C durable-wasm test` успешно собирает воркер примера и запускает unit-тесты ядра движка (тесты завершились успешно).
* Пример `make -C durable-wasm/examples/simple run` успешно отрабатывает весь жизненный цикл.
* Проверена обратная совместимость: интеграционные примеры (Camunda, Temporal, CSV) компилируются и работают без изменений в их коде.
