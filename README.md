# Subscription Reconciler

A robust Go-based backend service designed to reconcile premium subscription access from multiple sales channels (In-app Store, Mobile Carrier, and Third-party Marketplace). It maintains the "canonical truth" of a user's entitlement, triggers notifications before access expires, and maintains a comprehensive **Audit Log** for every state transition.

## 🚀 Quick Start

### Prerequisites

- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/install/)

### Setup & Run

The entire system, including the service, Postgres database, and mock carrier endpoint, can be started with a single command:

```bash
docker compose up --build
```

The API will be available at `http://localhost:8080`.

## 📖 API Documentation

### Interactive Swagger UI

The service provides a built-in Swagger UI for exploring and testing the API endpoints.

**Endpoint:** `http://localhost:8080/api-docs`

### OpenAPI Specification

You can also access the raw OpenAPI 3.0 specification file directly:

**Endpoint:** `http://localhost:8080/openapi.yaml`

---

## Endpoints Overview

### 1. Store Webhook Ingestion

- **Endpoint:** `POST /webhooks/store`
- **Logic:** Handles asynchronous, out-of-order, and late-arriving events from the In-app Store. It persists every unique event and "replays" the user's history to reconstruct the current state. This ensures that a late-arriving "Initial Purchase" doesn't overwrite a more recent "Cancellation".

### 2. Marketplace Bulk Revoke

- **Endpoint:** `POST /webhooks/marketplace/revoke`
- **Logic:** A monthly bulk revocation for marketplace users. It is scoped strictly to the `MARKETPLACE` source. If a user has a concurrent active subscription via the `STORE`, revoking their marketplace access will not impact their overall premium status.

### 3. Entitlement Read Endpoint

- **Endpoint:** `GET /users/:id/entitlement`
- **Logic:** Returns the canonical entitlement state. If a user has multiple active sources, it resolves precedence in this order: `STORE` > `CARRIER` > `MARKETPLACE`.

### 4. Entitlement Timeline (Stretch Goal)

- **Endpoint:** `GET /users/:id/timeline`
- **Logic:** Returns the reconstructed history of the user's entitlement changes from the internal Audit Log. This provides a transparent view of every state transition, including the reason and the event that triggered it.

---

## 🏗️ Design Decisions & Tradeoffs

### Event Replay (Event Sourcing Lite)

For the In-app Store, we store every unique `eventId`. When a new event arrives, we fetch all events for that user, sort them by `eventTimeMs`, and "reduce" them to find the latest state. This naturally solves the problem of out-of-order and late-arriving events without complex state-machine logic.

### Claim-Based Workers

Both the **Carrier Poller** and the **Notification Sender** use a "Claim" pattern (`FOR UPDATE SKIP LOCKED`). This allows us to run multiple instances of the service horizontally without two workers ever processing the same user or notification simultaneously.

### Secure Error Handling

The API follows a strict security philosophy regarding error exposure. Internal storage or query details are never leaked to the client. All internal errors are logged on the server with relevant context (including user ID where available), while the client receives a generic JSON error response like `{ "error": "Internal server error" }`.

### Database Precedence

Precedence is resolved at the domain layer (`internal/domain/entitlement/resolver.go`). This keeps the business rules out of SQL queries and makes them easily testable with unit tests.

## 🧪 Testing

Run the unit tests with:

```bash
go test ./test/unit/...
```

The test suite covers:

- Out-of-order store events.
- Precedence resolution (Store > Carrier > Marketplace).
- Notification scheduling logic.
- Marketplace revocation scoping.
- Audit log consistency.

## 🛠️ Future Improvements (If I had more time)

### 1. Observability & Monitoring

I would integrate **Prometheus** metrics to track event ingestion latency and worker success rates. Additionally, implementing **OpenTelemetry** for distributed tracing would be invaluable for debugging complex entitlement resolution flows across multiple sales channels.

### 2. Resilience: Dead Letter Queues

For webhook ingestion, I would implement a Dead Letter Queue (DLQ) pattern. If an event fails to process due to a transient error or malformed data, it should be moved to a separate table for manual inspection and replay, ensuring no customer signals are ever lost.

### 3. Scalability: Message Queues

While the `FOR UPDATE SKIP LOCKED` pattern is effective for current scales, I would evaluate transitioning to a dedicated message queue (like **AWS SQS** or **RabbitMQ**) for the notification and carrier polling workers. This would decouple the database from the worker orchestration and allow for even higher throughput.

### 4. Cache Layer

Adding a **Redis** cache for the `GET /users/:id/entitlement` endpoint would significantly reduce database load for high-frequency checks, using a cache-invalidation strategy triggered by new webhook events.
