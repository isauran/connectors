# HTTP-01: Исправление зависания ConcurrencyMiddleware при отмене контекста

Этот план описывает исправление проблемы с бесконечной блокировкой в `ConcurrencyMiddleware` при получении слота из семафора ограничения параллельных запросов, если контекст запроса был отменен.

## User Review Required

Изменения не ломают публичный API пакета `httpstream`. Это багфикс внутренней работы промежуточного ПО ограничения конкурентности.

## Open Questions

Нет.

## Proposed Changes

### Компонент `httpstream` (HTTP клиент)

---

#### [MODIFY] [limit.go](file:///Users/user/github.com/nativebpm/connectors/httpstream/limit.go)
Изменение метода `RoundTrip` для поддержки неблокирующего выхода по отмене контекста:
```go
func (c *concurrencyLimiter) RoundTrip(req *http.Request) (*http.Response, error) {
	select {
	case c.sem <- struct{}{}:
	case <-req.Context().Done():
		return nil, req.Context().Err()
	}
	defer func() { <-c.sem }()
	return c.next.RoundTrip(req)
}
```

#### [NEW] [limit_test.go](file:///Users/user/github.com/nativebpm/connectors/httpstream/limit_test.go)
Добавление тестов для верификации:
- Ограничение конкурентности (пропускной способности семафора).
- Быстрый возврат ошибки при отмене контекста заблокированного в очереди запроса.

## Verification Plan

### Automated Tests
1. Запуск тестов модуля `httpstream`:
   ```bash
   go test -v ./httpstream/...
   ```
