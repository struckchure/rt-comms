# Implementation Spec (Checklist)

> Traceability: sections map to [ASSESSMENT.md](ASSESSMENT.md). Use this file to track implementation progress.

| Document | Role |
|----------|------|
| [ASSESSMENT.md](ASSESSMENT.md) | Interviewer requirements (source of truth) |
| **spec.md** | Working checklist: endpoints, protocol, files, verification |

**Design extension (document in README):** Assessment focuses on text + voice signaling. This build also persists **voice** and **video** calls as `message_type` rows in chat history. Live media stays **WebRTC P2P** — signaling and call records only through the server.

---

## Progress summary

- [x] Section 1 – Real-time chat ([ASSESSMENT §1](ASSESSMENT.md#section-1--real-time-chat-service))
- [x] Section 2 – Database ([ASSESSMENT §2](ASSESSMENT.md#section-2--database-design))
- [x] Section 3 – Graceful shutdown ([ASSESSMENT §3](ASSESSMENT.md#section-3--graceful-shutdown))
- [x] Section 4 – Calls in chat ([ASSESSMENT §4](ASSESSMENT.md#section-4--voice-call-signaling-webrtc) + extension)
- [x] Documentation & deliverables

---

## Locked decisions

| Topic | Choice |
|-------|--------|
| HTTP | [Echo v5](https://github.com/labstack/echo) — `func(c *echo.Context) error` |
| WebSocket | `github.com/gorilla/websocket` |
| DB | PostgreSQL + **pgx** |
| Migrations | `golang-migrate/migrate` under `migrations/` |
| Config | **viper** — [`.env`](.env) + `AutomaticEnv()` → typed `Config` struct |
| User on WS | `GET /ws?user=alice` |
| Reconnect | **Replace** prior socket for same user (document in README) |
| History | `GET /users/:user/direct/:peer/messages` |
| Pagination | `limit` (default 50, max 200), `before` = message UUID, **newest first** |
| Message types | `text`, `voice`, `video` — DB + HTTP history + WS delivery |
| STUN (test client) | `stun:stun.l.google.com:19302` (configurable via `STUN_URL`) |

### WebSocket protocol (summary)

| Purpose | `type` | Notes |
|---------|--------|--------|
| Text chat | `message` | `message_type: "text"`, `to`, `text` |
| Errors | `error` | `code`, `message` |
| Ack (optional) | `ack` | `id`, `message_type`, `created_at` |
| Call signaling | `call_offer`, `call_answer`, `call_reject`, `call_hangup`, `call_ice` | `media`: `voice` \| `video`; opaque `sdp` / `candidate` |
| Call → history | — | On terminal call event → persist `voice`/`video` row → deliver like text |

---

## Package layout

```text
cmd/server/main.go              # LoadConfig → NewApp(cfg).Run()

handlers/                       # I*Handler + *Handler + New*Handler (no routes.go)
  health_handler.go
  history_handler.go
  websocket_handler.go
  static_handler.go

services/                       # I*Service + *Service + New*Service
  chat_service.go
  connection_service.go
  call_service.go

repository/
  message_repository.go         # IMessageRepository + MessageRepository + NewMessageRepository

ws/
  client.go
  pumps.go

internal/                       # package internal — flat files only
  app.go                        # Echo, ALL routes, AppDeps, shutdown
  config.go                     # Config + LoadConfig (viper)
  hub.go                        # IHub + Hub + NewHub
  callsession.go
  protocol.go
  errors.go

migrations/
static/testclient.html
models/message.go               # replace stub models/chat.go
```

### Dependency rules

- **handlers** → services, ws, internal (errors, protocol)
- **services** → repository, internal (hub, callsession, errors), ws
- **repository** → pgx only
- **internal** support files (`hub.go`, etc.) → must **not** import handlers, services, repository
- **internal/app.go** → may import echo, handlers, services, repository (composition root)
- **Routing only in `internal/app.go`** — no `handlers/routes.go`

### Interface naming (`I*` / struct / `New*`)

```go
type IChatService interface { /* ... */ }
type ChatService struct { /* ... */ }
func NewChatService(...) IChatService { return &ChatService{...} }
```

| Layer | Interface | Struct | Constructor |
|-------|-----------|--------|-------------|
| `repository/` | `IMessageRepository` | `MessageRepository` | `NewMessageRepository(...) IMessageRepository` |
| `services/` | `IChatService` | `ChatService` | `NewChatService(...) IChatService` |
| `services/` | `IConnectionService` | `ConnectionService` | `NewConnectionService(...) IConnectionService` |
| `services/` | `ICallService` | `CallService` | `NewCallService(...) ICallService` |
| `handlers/` | `IHealthHandler` | `HealthHandler` | `NewHealthHandler(...) IHealthHandler` |
| `handlers/` | `IHistoryHandler` | `HistoryHandler` | `NewHistoryHandler(...) IHistoryHandler` |
| `handlers/` | `IWebSocketHandler` | `WebSocketHandler` | `NewWebSocketHandler(...) IWebSocketHandler` |
| `internal/hub.go` | `IHub` | `Hub` | `NewHub() IHub` |

`AppDeps` (in `internal/app.go`) holds **only `I*` fields**. Tests use manual fakes implementing `I*` interfaces in `*_test.go`.

---

## Unified chat messages (extension)

| `message_type` | When created | `content` | `metadata` (JSONB) |
|----------------|--------------|-----------|---------------------|
| `text` | User sends chat | body | null |
| `voice` | Call ends / rejected / missed | optional label | `call_id`, `status`, `duration_sec`, `media: "voice"` |
| `video` | Same for video call | optional label | same with `media: "video"` |

Active calls: WS = signaling only. Media = **P2P WebRTC**, not stored on server.

---

## Module scaffolding

- [ ] Create dirs: `handlers/`, `services/`, `repository/`, `ws/`, `internal/`, `migrations/`, `static/`
- [ ] `go.mod`: `echo/v5`, `gorilla/websocket`, `pgx`, `viper`, `migrate`
- [ ] `internal/config.go` — viper + typed `Config` (see below)
- [ ] `internal/app.go` — `NewApp`, `AppDeps`, route registration, shutdown
- [ ] `internal/hub.go`, `callsession.go`, `protocol.go`, `errors.go`
- [ ] Handler files: `*_handler.go` with `I*` + struct + `New*`
- [ ] Service files: `*_service.go` with `I*` + struct + `New*`
- [ ] `repository/message_repository.go` — `IMessageRepository` + impl
- [ ] `models/message.go`; remove or replace [models/chat.go](models/chat.go)

### Config (`internal/config.go` + `.env`)

- [ ] `Config` struct with `mapstructure` tags: `PORT`, `DATABASE_URL`, `STUN_URL` (optional)
- [ ] `LoadConfig()`: viper `SetConfigFile(".env")`, `SetConfigType("env")`, `ReadInConfig`, `AutomaticEnv`, `Unmarshal`
- [ ] Defaults: `PORT=:8080`; validate `DATABASE_URL` required
- [ ] [`.env`](.env) aligned with tags; add `PORT` if missing
- [ ] [cmd/server/main.go](cmd/server/main.go): `cfg, _ := internal.LoadConfig()` → `internal.NewApp(cfg).Run()`

Example `.env`:

```env
PORT=:8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres
STUN_URL=stun:stun.l.google.com:19302
```

---

## Section 1 – Real-time chat service

*Maps to [ASSESSMENT §1](ASSESSMENT.md#section-1--real-time-chat-service)*

### 1.0 Echo bootstrap

- [ ] `echo.New()` in `internal/app.go`; `middleware.RequestLogger`, `Recover`
- [ ] All routes registered in `app.go` only
- [ ] `e.Start(cfg.Port)`; `e.Shutdown(ctx)` on SIGINT/SIGTERM

### 1.1 Health check

- [ ] `IHealthHandler` / `HealthHandler` / `NewHealthHandler`
- [ ] `GET /health` → `200` `{"status":"ok"}` (optional DB ping — document if added)
- [ ] Route: `e.GET("/health", deps.Health.Health)`

### 1.2 WebSocket connection management

- [ ] `GET /ws` — `IWebSocketHandler`; require `?user=`; reject invalid/empty
- [ ] `IConnectionService` + `internal.IHub`: one active connection per user; **replace** on reconnect
- [ ] Unregister on disconnect; no writes after unregister
- [ ] `ws/`: read/write pumps; race-safe hub tests

### 1.3 Direct messaging

- [ ] WS `type: message`, `message_type: text`, `to`, `text`
- [ ] Sender = connected user; block self-send
- [ ] **Persist before deliver:** `IMessageRepository.Insert` → then `IConnectionService.DeliverTo` if online
- [ ] DB failure → `type: error` to sender; no delivery
- [ ] Offline recipient → DB only
- [ ] Payload: `id`, `sender`, `recipient`, `message_type`, `content`, `metadata`, `created_at`
- [ ] Optional `type: ack` after persist

### 1.4 Chat history retrieval

- [ ] `IHistoryHandler` depends on `IChatService` only
- [ ] `GET /users/:user/direct/:peer/messages` — query `limit`, `before`
- [ ] Returns **text + voice + video**; both directions; empty → `[]`
- [ ] Indexed thread query in repository

### 1.5 Section 1 constraints

- [ ] Persist-before-deliver verified
- [ ] One socket per user
- [ ] No self-messages
- [ ] `go test -race` on hub / connection tests

---

## Section 2 – Database design

*Maps to [ASSESSMENT §2](ASSESSMENT.md#section-2--database-design)*

- [ ] Migration `messages`: `id` (UUID), `sender`, `recipient`, `message_type`, `content`, `metadata` (JSONB), `created_at`
- [ ] Index for thread by pair + `created_at`
- [ ] `IMessageRepository`: `InsertMessage`, `ListThread`
- [ ] `NewMessageRepository(pool)` wired from `cfg.DatabaseURL`
- [ ] `models/message.go`: `MessageType`, call `Metadata` struct

---

## Section 3 – Graceful shutdown

*Maps to [ASSESSMENT §3](ASSESSMENT.md#section-3--graceful-shutdown)*

- [ ] `signal.Notify` SIGINT, SIGTERM
- [ ] Stop new HTTP/WS; `IConnectionService.Shutdown`; `ICallService.EndAll`
- [ ] `e.Shutdown` with timeout (~10s); drain DB ops
- [ ] Manual: Ctrl+C, no panic

---

## Section 4 – Calls in chat (voice + video)

*Maps to [ASSESSMENT §4](ASSESSMENT.md#section-4--voice-call-signaling-webrtc) + history extension*

### Signaling (P2P media)

- [ ] Same `/ws`: route `call_*` to `ICallService`
- [ ] `call_offer` with `media: voice|video`; forward SDP/ICE only
- [ ] `call_answer`, `call_reject`, `call_hangup`, `call_ice`
- [ ] Callee offline → clear WS error (e.g. `peer_offline`)
- [ ] **No** audio/video bytes on WS or through Go
- [ ] `internal/callsession` for active `call_id` state
- [ ] Disconnect mid-call → notify peer; end session

### Call records in chat

- [ ] Terminal events → `IChatService.RecordCallMessage` → persist → deliver if online
- [ ] `message_type` `voice` or `video` with metadata (`status`, `duration_sec`, etc.)
- [ ] Rejected / missed calls still create history rows
- [ ] `static/testclient.html` — text + voice + video call buttons
- [ ] `e.Static` or file route for test client in `app.go`

---

## Documentation & deliverables

*Maps to [ASSESSMENT Documentation](ASSESSMENT.md#documentation) & [Deliverables](ASSESSMENT.md#deliverables)*

- [ ] [README.md](README.md): Postgres ([docker-compose.yaml](docker-compose.yaml)), migrations, start app
- [ ] `.env` / viper / typed `Config` table
- [ ] WebSocket + websocat examples
- [ ] History `curl` examples
- [ ] WebRTC signaling + `testclient.html` steps
- [ ] **Design extension** (text/voice/video in history vs assessment)
- [ ] Assumptions, trade-offs, reconnect policy
- [ ] Postman collection **or** API doc (OpenAPI/markdown)
- [ ] `go build ./...` and `go test ./...` pass

---

## Implementation order

1. `internal/config.go` + `app.go` skeleton + `IMessageRepository` + migrations + health route  
2. `IHub` + `IConnectionService` + `IWebSocketHandler` + `IChatService` (text) + `IHistoryHandler`  
3. Graceful shutdown  
4. `ICallService` + call messages in history + `testclient.html`  
5. README, API doc, interface-based tests  

---

## Verification matrix

| # | Check | Command / action |
|---|--------|------------------|
| 1 | Health | `curl -s localhost:8080/health` |
| 2 | Live text | `websocat 'ws://localhost:8080/ws?user=alice'` → message to bob; bob receives |
| 3 | Isolation | charlie does not receive alice→bob message |
| 4 | History | `curl 'localhost:8080/users/alice/direct/bob/messages?limit=50'` includes message |
| 5 | Offline | bob offline → message in history only after reconnect |
| 6 | Race | `go test ./... -race` |
| 7 | Voice history | after voice call → history has `message_type: voice` |
| 8 | Video history | after video call → `message_type: video`; media was P2P |

---

## Out of scope

- Group chat, rooms, multi-recipient
- Server-side media storage or relay through Go/WS
- Full auth (query-string `user` is fine for assessment)
- Clustering / multi-instance
- Persisted call CDRs beyond chat `messages` rows

**In scope:** video signaling + `message_type: video` in history (P2P only).

---

## Existing scaffold (reference)

| Path | Status |
|------|--------|
| [cmd/server/main.go](cmd/server/main.go) | Stub — wire `LoadConfig` + `NewApp` |
| [models/chat.go](models/chat.go) | Replace with `models/message.go` |
| [docker-compose.yaml](docker-compose.yaml) | Postgres ready |
| [.env](.env) | Has `DATABASE_URL` — add `PORT`, `STUN_URL` |
