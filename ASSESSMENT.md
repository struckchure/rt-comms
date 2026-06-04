# Backend Developer Assessment

## Assessment Objective

Development of a Real-Time 1-on-1 Chat Backend Service with message persistence, chat history retrieval, and WebRTC voice call signaling.

## Technology Stack

- Golang
- PostgreSQL
- pgx
- gorilla/websocket or coder/websocket
- net/http (or Chi/Echo)
- migrate (or goose)

## Project Tasks

### Section 1 – Real-time Chat Service

#### Health Check

Create a health check endpoint that confirms the service is running and available.

**Requirements**

- Return a successful response when the service is healthy.
- Support service monitoring and health checks.

#### WebSocket Connection Management

Implement WebSocket support for real-time communication between users.

**Requirements**

- Associate each WebSocket connection with a unique user ID.
- Reject connections that do not provide a valid user ID.
- Support only one active connection per user.
- Document how reconnects are handled when a user already has an active connection.
- Remove disconnected users from the active connection registry.
- Prevent messages from being sent to disconnected clients.
- Ensure connection management is safe under concurrent access.

#### Direct Messaging

Implement one-to-one real-time messaging between connected users.

**Requirements**

- Allow users to send messages to a single recipient.
- Ensure the sender matches the connected user.
- Prevent users from sending messages to themselves.
- Persist messages to PostgreSQL before attempting delivery.
- If message persistence fails:
  - Notify the sender.
  - Do not deliver the message.
- If the recipient is online:
  - Deliver the message immediately through WebSocket.
- If the recipient is offline:
  - Store the message in the database only.
- Each stored message must include:
  - Unique identifier
  - Sender
  - Recipient
  - Message content
  - Timestamp
- Optionally send a delivery acknowledgement to the sender after successful persistence.

#### Chat History Retrieval

Implement conversation history retrieval between two users.

**Requirements**

- Return all messages exchanged between two users regardless of direction.
- Support pagination.
- Choose and document a pagination strategy.
- Return an empty list when no conversation exists.
- Ensure conversation queries remain efficient as message volume grows.

**Constraints**

- Messages must always be persisted before delivery.
- Only one active WebSocket connection is allowed per user.
- Users cannot send messages to themselves.
- Message history retrieval must be optimized for performance.
- Message delivery and connection management must be safe under concurrent usage.

---

### Section 2 – Database Design

#### Message Persistence Layer

Use PostgreSQL as the primary data store for chat messages.

**Requirements**

- Store all chat messages in the database.
- Create database migrations for schema management.
- Include message metadata such as:
  - Sender
  - Recipient
  - Content
  - Creation timestamp
- Create appropriate indexes to efficiently retrieve conversations between two users.
- Use proper repository abstractions and data access patterns.
- Ensure data consistency and reliability.

---

### Section 3 – Graceful Shutdown

Implement graceful application shutdown.

**Requirements**

- Handle SIGINT and SIGTERM signals.
- Stop accepting new HTTP requests.
- Stop accepting new WebSocket connections.
- Close active WebSocket connections gracefully.
- Allow in-flight database operations to complete or time out safely.
- Release resources and exit cleanly.

---

### Section 4 – Voice Call Signaling (WebRTC)

#### WebRTC Signaling

Implement signaling support for peer-to-peer voice calls using the existing WebSocket connection.

**Requirements**

- Support call initiation between two users.
- Support call acceptance and rejection.
- Support ICE candidate exchange.
- Support call termination.
- Forward signaling messages only when the recipient is online.
- Return a clear error when the recipient is unavailable.
- Ensure media is transmitted directly between peers using WebRTC.
- Do not transmit audio through WebSockets.
- Do not relay audio through the backend server.
- If either participant disconnects during an active call:
  - Notify the other participant.
  - Terminate the active call session.
- Provide a minimal browser-based test client (`testclient.html`) for testing voice calls.
- Document the signaling flow and voice call testing process in the README.

**Constraints**

- The backend must only handle signaling messages.
- Audio streams must remain peer-to-peer between participants.
- No audio, video, or media data should pass through the backend service.

---

## Documentation

Provide documentation covering:

- Running and testing the application.
- WebRTC setup and signaling flow.
- Testing with a minimal HTML client (`testclient.html`) for voice calls.
- Design decisions and implementation assumptions.

## Deliverables

GitHub repository containing:

- Complete source code.
- Postman Collection or API documentation link.
- README file.

### README Requirements

The README should include:

- PostgreSQL setup instructions.
- Running database migrations.
- Starting the application.
- Environment variable configuration.
- WebSocket connection and testing instructions.
- Chat history retrieval instructions.
- Voice signaling setup and testing instructions.
- Assumptions made during development.
- Trade-offs and potential improvements.
