# Implementation Plan — Telegram Bot API Connector (TEL-03)

This plan outlines the implementation of a Telegram Bot API connector module in the `nativebpm/connectors` monorepository. The connector will leverage `httpstream` for zero-buffer file uploads (documents, photos) and support sending messages, handling callbacks, webhooks, and long-polling.

## User Review Required

> [!IMPORTANT]
> The connector will be implemented as a new Go module `./telegram`.
> To run tests and verify the code, we will add `./telegram` to `go.work`.

> [!NOTE]
> We will use `httptest.Server` in the tests to simulate Telegram's REST API, ensuring that we verify zero-buffer streaming behavior and query/multipart parameter passing.

## Proposed Changes

### Connectors Monorepo Workspace Configuration

#### [MODIFY] [go.work](file:///Users/user/github.com/nativebpm/connectors/go.work)
Add `./telegram` to the workspace use list.

---

### Telegram Connector Module

#### [NEW] [go.mod](file:///Users/user/github.com/nativebpm/connectors/telegram/go.mod)
Create a new Go module `github.com/nativebpm/connectors/telegram` requiring `github.com/nativebpm/httpstream`.

#### [NEW] [telegram.go](file:///Users/user/github.com/nativebpm/connectors/telegram/telegram.go)
Implement the core structures and `Client` for Telegram Bot API:
- **Client**: Holds the `httpstream.Client` pointing to `https://api.telegram.org/bot<token>`.
- **SendMessage**: Sends text messages with optional Markdown/HTML and markup keyboards.
- **SendDocument**: Streams a file from `io.Reader` using `httpstream.Multipart` directly to `/sendDocument`.
- **SendPhoto**: Streams an image from `io.Reader` using `httpstream.Multipart` directly to `/sendPhoto`.
- **AnswerCallbackQuery**: Sends response to interactive callback queries.
- **SetWebhook**: Configures Telegram webhook.
- **GetUpdates**: Performs long-polling to retrieve incoming updates.
- **Data Models**:
  - `Update`, `Message`, `Chat`, `User`, `CallbackQuery`
  - `InlineKeyboardMarkup`, `InlineKeyboardButton`, `ReplyKeyboardMarkup`, `ReplyParameters`

#### [NEW] [telegram_test.go](file:///Users/user/github.com/nativebpm/connectors/telegram/telegram_test.go)
Implement tests to verify:
1. `SendMessage` correctly serializes JSON and handles successful/error API responses.
2. `SendDocument` and `SendPhoto` perform streaming multipart file transfers and parse parameters.
3. Callback response and update fetching.

#### [NEW] [Makefile](file:///Users/user/github.com/nativebpm/connectors/telegram/Makefile)
Add a standard Makefile for linting and testing the module.

#### [NEW] [README.md](file:///Users/user/github.com/nativebpm/connectors/telegram/README.md)
Add standard module documentation and usage examples.

---

## Verification Plan

### Automated Tests
- Run `go test -v ./telegram/...` to ensure all tests pass successfully.
- Run `golangci-lint run ./telegram/...` to verify compliance with formatting/lint rules.

### Manual Verification
- We will include an example code showing how a client can be initialized and used.
