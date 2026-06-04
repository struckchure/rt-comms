# RT-Comms

Real-time 1-on-1 chat backend: WebSocket delivery, PostgreSQL persistence and history, and WebRTC **signaling** for voice and video calls. Media stays peer-to-peer; the server never relays audio or video.

This README is aligned with [ASSESSMENT.md](ASSESSMENT.md) and reflects the current implementation. Full request/response shapes are in [docs/API.md](docs/API.md).

## Assessment coverage

| Area | Requirement | Status |
|------|-------------|--------|
| **Health** | `GET /health` when service is up | Done |
| **WebSocket** | User ID on connect; reject missing ID | Done (`?user=` required) |
| | One active connection per user | Done (new connection closes the previous) |
| | Reconnect behavior documented | See [WebSocket](#websocket) |
| | Unregister on disconnect; no delivery to offline sockets | Done |
| | Thread-safe hub (`sync.RWMutex`) | Done |
| **Direct messaging** | 1:1 messages; sender = connected user | Done |
| | No self-messages | Done (`ErrSelfMessage`) |
| | Persist before delivery; error to sender on persist failure | Done |
| | Deliver when recipient online; store-only when offline | Done |
| | Fields: id, sender, recipient, content, timestamp | Done (+ `message_type`, optional `metadata`) |
| | Optional ack after persist | Done (`type: ack`) |
| **Chat history** | Bidirectional thread between two users | Done |
| | Pagination | Done (`limit`, `before` cursor) |
| | Empty list when no messages | Done (`[]`) |
| | Conversation index for scale | Done (see [Database](#database)) |
| **Database** | PostgreSQL + migrations + repository layer | Done |
| **Graceful shutdown** | SIGINT/SIGTERM; stop HTTP/WS; drain resources | Done (see [Shutdown](#graceful-shutdown)) |
| **Voice signaling** | offer / answer / reject / hangup / ICE; P2P media only | Done |
| | Signaling only when peer online; clear offline error | Done (`peer_offline`) |
| | End session when participant disconnects | Done (session cleared; call row persisted as `missed`) |
| | Minimal `testclient.html` + README testing notes | Done |
| **Documentation** | Run, test, signaling, assumptions, trade-offs | This file + `docs/API.md` |

**Beyond the assessment (implemented):** `media: video` signaling and persisted `video` call rows in chat history. Live video is still WebRTC P2P only.

## Prerequisites

- Go 1.25+
- Docker (PostgreSQL via Compose)
- Make (optional; recommended)

## Quick start

```bash
cp .env.sample .env   # edit values if needed (see Configuration)
make dev              # Postgres + migrations + server
```

Server listens on `http://localhost:8080` by default.

## PostgreSQL setup

```bash
make db-up      # docker compose up -d --wait
make db-down    # stop database
make db-logs    # follow Postgres logs
```

Default connection (override in `.env`):

`postgres://postgres:postgres@localhost:5432/postgres`

## Database migrations

```bash
make migrate-up        # apply all migrations
make migrate-down      # roll back one version
make migrate-version   # show current version
```

Migrations run via `go run` against [golang-migrate](https://github.com/golang-migrate/migrate) (no separate binary install). Schema: `messages` with thread index on `(LEAST(sender,recipient), GREATEST(sender,recipient), created_at DESC)`.

## Configuration

Copy `.env.sample` to `.env` before running locally. `.env` is gitignored. The Makefile loads `.env` when present; Viper reads the same variables at runtime.

| Field | Env | Default | Required |
|-------|-----|---------|----------|
| HTTP port | `PORT` | `:8080` | no |
| Database | `DATABASE_URL` | — | yes |
| STUN (client hint) | `STUN_URL` | `stun:stun.l.google.com:19302` | no |

## Starting the application

```bash
make run
# or: go run ./cmd/server
```

Build binary: `make build` → `bin/server`.

## HTTP

```bash
curl -s localhost:8080/health
curl -s "localhost:8080/users/alice/direct/bob/messages?limit=50"
curl -s "localhost:8080/users/alice/direct/bob/messages?limit=50&before=<message-uuid>"
```

**Pagination:** Newest-first pages. Pass `before=<uuid>` from the oldest item in the current page to load older messages. `limit` defaults to 50, max 200.

## WebSocket

**Connect:** `ws://localhost:8080/ws?user=<userId>`

- Missing or empty `user` → HTTP 400.
- **Reconnect:** A second connection for the same `user` closes the previous socket and replaces it in the hub. In-flight work on the old socket stops when its read pump exits.
- **Delivery:** `Hub.DeliverTo` only writes to an online client; offline peers do not receive live frames (messages remain in Postgres).

**Send a text message:**

```json
{"type":"message","message_type":"text","to":"bob","text":"hello"}
```

**Inbound errors** (e.g. self-message, invalid payload, DB failure) return:

```json
{"type":"error","code":"...","message":"..."}
```

Example with [websocat](https://github.com/vi/websocat):

```bash
websocat "ws://localhost:8080/ws?user=alice"
```

## WebRTC signaling and testing

The backend forwards SDP and ICE only. **No** audio, video, or RTP passes through the server.

### Signaling flow

1. **Caller** sends `call_offer` with `to`, `media` (`voice` or `video`), and SDP.
2. If **callee is offline** → caller gets WebSocket `error` with `peer_offline`; no session.
3. If **callee is online** → callee receives `call_offer` with `call_id` and `from`; caller receives `call_started` with `call_id`.
4. **Callee** sends `call_answer` or `call_reject` with `call_id` and SDP (answer).
5. Both sides exchange `call_ice` with `call_id` and candidate payloads.
6. **Hangup** or **reject** → peer notified via signaling; server records a `voice`/`video` history row (`completed`, `rejected`, or `missed`) and pushes it like a chat message when the peer is online.
7. **WebRTC media** flows directly between browsers after signaling.

### Test client (`testclient.html`)

1. Start the server (`make dev`).
2. Open `http://localhost:8080/testclient.html` in **two** browser profiles or machines (not `file://`).
3. Set users (e.g. `alice` / `bob`), same peer field, click **Connect** on both.
4. Send text; use **Load history** or rely on history loaded on connect.
5. **Voice** or **Video**: caller starts call; callee **Accept** or **Decline** (no auto-answer). Allow camera/mic for video.
6. Configure STUN in the UI (defaults to Google STUN). For restrictive NAT, add a TURN server in the client WebRTC config (not provided by this backend).

Signaling message reference: [docs/API.md](docs/API.md).

## Graceful shutdown

On `SIGINT` / `SIGTERM`:

1. Echo stops accepting new HTTP requests (10s graceful timeout).
2. Active calls are cleared (`Call.EndAll`).
3. Hub closes all WebSocket connections (`ConnectionService.Shutdown`).
4. DI scope and `pgx` pool close.

## Stack

| Assessment | This repo |
|------------|-----------|
| Go | Go 1.25+ |
| PostgreSQL + pgx | Yes |
| gorilla/websocket | Yes |
| net/http router | Echo v5 |
| migrate | `go run` … `migrate/v4` via Makefile |

## Design notes

- **History types:** `text`, `voice`, `video`. Terminal call events persist metadata: `call_id`, `status`, `duration_sec`, `media`.
- **Layout:** `cmd/server` (app + DI), `handlers/`, `services/`, `repository/`, `internal/` (config, hub, protocol, errors), `ws/`, `static/testclient.html`.
- **Dependency injection:** [samber/do v2](https://github.com/samber/do) in `cmd/server/di.go`.

## Assumptions

- Identity is the `?user=` query string (no authentication or authorization).
- User IDs are opaque strings; no registration service.
- `CheckOrigin` allows all origins (dev/test client only).
- Call sessions are in-memory (single process).
- Disconnect during a call ends the server session and records `missed`; peers should also handle local WebRTC teardown (test client listens for `call_hangup` and WS close).

## Trade-offs and improvements

- **No clustering** — hub and call store are process-local.
- **No TURN** — signaling-only; NAT traversal may need TURN in production clients.
- **No message read receipts** — only optional persist ack to sender.
- **No rate limiting or auth** — would be required for a public deployment.
- **Postman** — use `docs/API.md` instead of a collection file.

## Tests

```bash
make test
# or: go test ./... -race
```
