# WASM-10: Обновление golangci-lint до v2 и статический анализ

В рамках этой задачи мы обновим инструмент статического анализа `golangci-lint` с устаревшей версии (v1.64.x) до актуальной мажорной версии v2.12.2, адаптируем файл конфигурации `.golangci.yml` и проверим кодовую базу.

## User Review Required

> [!IMPORTANT]
> Переход на `golangci-lint` v2 является мажорным обновлением. Нам потребуется обновить структуру файла конфигурации `.golangci.yml` и убедиться, что используемые линтеры совместимы.

## Open Questions

Нет открытых вопросов.

## Proposed Changes

### Build Infrastructure & Configuration

#### [MODIFY] [Makefile](file:///Users/user/github.com/nativebpm/connectors/Makefile)
- Обновить версию устанавливаемого линтера в цели `lint-install` с `v1.64.8` на `v2.12.2`.
- Заменить `go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8` на установку версии `v2.12.2`.

#### [MODIFY] [.golangci.yml](file:///Users/user/github.com/nativebpm/connectors/.golangci.yml)
- Смигрировать формат файла конфигурации под стандарт `golangci-lint` v2 с помощью команды `golangci-lint migrate` или обновить вручную.

## Verification Plan

### Automated Tests
- Запустить `make lint` и убедиться, что `golangci-lint` успешно запускается и не выдает критических ошибок о синтаксисе или несовместимости конфигурации.
- Исправить предупреждения линтера в коде, если новая версия обнаружит новые проблемы.
