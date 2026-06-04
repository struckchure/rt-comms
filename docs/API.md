# RT-Comms API

## HTTP

### GET /health

Response `200`:

```json
{"status":"ok"}
```

### GET /users/:user/direct/:peer/messages

Query parameters:

| Name | Default | Max | Description |
|------|---------|-----|-------------|
| limit | 50 | 200 | Page size |
| before | — | — | UUID cursor; returns messages older than this id |

Response `200`: JSON array of messages (newest first).

```json
[
  {
    "id": "uuid",
    "sender": "alice",
    "recipient": "bob",
    "message_type": "text",
    "content": "hello",
    "metadata": null,
    "created_at": "2026-06-03T12:00:00Z"
  }
]
```

## WebSocket GET /ws?user={userId}

### Client → server

**Text message**

```json
{"type":"message","message_type":"text","to":"bob","text":"hello"}
```

**Call signaling** (opaque SDP/ICE forwarded to peer)

After `call_offer`, the server sends the caller `call_started` with `call_id` for ICE/hangup.

```json
{"type":"call_offer","to":"bob","media":"voice","sdp":{}}
{"type":"call_started","call_id":"...","to":"bob","media":"voice"}
{"type":"call_offer","call_id":"...","from":"alice","media":"voice","sdp":{}}
{"type":"call_answer","call_id":"...","sdp":{}}
{"type":"call_reject","call_id":"..."}
{"type":"call_hangup","call_id":"..."}
{"type":"call_ice","call_id":"...","candidate":{}}
```

### Server → client

**Message** (text or call history row)

```json
{
  "type": "message",
  "id": "uuid",
  "sender": "alice",
  "recipient": "bob",
  "message_type": "text",
  "content": "hello",
  "created_at": "..."
}
```

**Ack** (optional, after persist)

```json
{"type":"ack","id":"uuid","created_at":"..."}
```

**Error**

```json
{"type":"error","code":"peer_offline","message":"peer offline"}
```
