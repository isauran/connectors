# Результаты реализации: Шифрование Payload и Codec Server (SEC-01)

Реализован рекомендуемый (Способ Б) вариант интеграции шифрования полезных нагрузок в Temporal и автоматической расшифровки в админке.

## Что сделано:

1. **Добавлен шифрующий кодек (`CryptCodec`)**:
   - Реализован в [temporal/codec.go](file:///Users/user/github.com/nativebpm/connectors/temporal/codec.go).
   - Использует стандартный криптографический алгоритм **AES-GCM (256-бит)** для шифрования данных.
   - Хеширует переданный секретный ключ через **SHA-256**, гарантируя корректный размер ключа безопасности для любых паролей.
   - Полностью шифрует оригинальные `Payload` (включая метаданные), сохраняя наружу только служебную пометку `encoding: "binary/encrypted"`.

2. **Интеграция в Temporal Client**:
   - Обновлен клиент в [temporal/client.go](file:///Users/user/github.com/nativebpm/connectors/temporal/client.go).
   - Если задана переменная окружения `TEMPORAL_ENCRYPTION_KEY`, клиент автоматически инициализирует шифрующий `DataConverter` и применяет его для всех отправляемых и получаемых данных.

3. **Реализован Codec Server**:
   - Код сервера находится в [temporal/cmd/codec-server/main.go](file:///Users/user/github.com/nativebpm/connectors/temporal/cmd/codec-server/main.go).
   - Это простой HTTP-сервер, который слушает эндпоинты `/decode` и `/encode` с поддержкой CORS для корректной работы из веб-браузеров (админки).

4. **Интеграция в Docker Compose**:
   - Создан [Dockerfile](file:///Users/user/github.com/nativebpm/connectors/temporal/docker/codec/Dockerfile) для сборки контейнера `codec-server`.
   - В [temporal/docker/docker-compose.yaml](file:///Users/user/github.com/nativebpm/connectors/temporal/docker/docker-compose.yaml) добавлен сервис `codec-server`.
   - Настроен автоматический проброс адреса кодека в админку Temporal через переменную `TEMPORAL_UI_CODEC_ENDPOINT=http://localhost:8082/decode`.

5. **Написаны тесты**:
   - Написан юнит-тест [temporal/codec_test.go](file:///Users/user/github.com/nativebpm/connectors/temporal/codec_test.go), проверяющий шифрование и полную расшифровку с сохранением структуры данных.

---

## Запуск и проверка:

### 1. Запуск инфраструктуры
Перейдите в директорию `temporal/docker/` и поднимите контейнеры:
```bash
docker compose up -d --build
```
Это автоматически соберет `codec-server` и запустит сервер Temporal с проброшенным эндпоинтом дешифрования.

### 2. Запуск тестов
Для проверки работоспособности шифрования в Go выполните:
```bash
go test -v -run=TestCryptCodec_EncodeDecode ./...
```
*(Все тесты успешно пройдены)*.

### 3. Проверка в админке
1. Откройте браузер по адресу `http://localhost:8233`.
2. Запустите любой workflow, передающий данные (например, `make run-loadtest`).
3. В админке Temporal вы увидите данные в **чистом (нешифрованном)** текстовом виде, так как админка на лету обращается к `http://localhost:8082/decode`.
4. Если вы остановите контейнер `codec-server` (`docker compose stop codec-server`), данные в админке превратятся в зашифрованный бинарный код `binary/encrypted`, доказывая надежность защиты.
