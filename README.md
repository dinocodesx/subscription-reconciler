# Subscription Reconciler

A robust Go-based backend service designed to reconcile premium subscription access from multiple sales channels (In-app Store, Mobile Carrier, and Third-party Marketplace). It maintains the "canonical truth" of a user's entitlement and triggers notifications before access expires.

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

#### Example Request

```json
{
  "eventId": "evt_abc123",
  "userId": "u_42",
  "type": "RENEWAL",
  "eventTimeMs": 1716700000000,
  "productId": "premium_monthly"
}
```

#### Example Response

```json
{
  "duplicate": false,
  "entitlement": {
    "active": true,
    "source": "STORE",
    "expiresAt": "2026-06-10T00:00:00Z",
    "lastChangedAt": "2026-05-20T11:23:00Z",
    "reason": "RENEWAL"
  }
}
```

#### 2. Marketplace Bulk Revoke

- **Endpoint:** `POST /webhooks/marketplace/revoke`
- **Logic:** A monthly bulk revocation for marketplace users. It is scoped strictly to the `MARKETPLACE` source. If a user has a concurrent active subscription via the `STORE`, revoking their marketplace access will not impact their overall premium status.

#### Example Request

```json
{
  "userIds": ["u_42", "u_91", "u_133"]
}
```

#### Example Response

```json
{
  "revokedCount": 3
}
```

#### 3. Entitlement Read Endpoint

- **Endpoint:** `GET /users/:id/entitlement`
- **Logic:** Returns the canonical entitlement state. If a user has multiple active sources, it resolves precedence in this order: `STORE` > `CARRIER` > `MARKETPLACE`.

#### Example Response

```json
{
  "active": true,
  "source": "STORE",
  "expiresAt": "2026-06-25T10:00:00Z",
  "lastChangedAt": "2026-05-26T10:00:00Z",
  "reason": "INITIAL_PURCHASE"
}
```

---

## 🏗️ Design Decisions & Tradeoffs

### Event Replay (Event Sourcing Lite)

For the In-app Store, we store every unique `eventId`. When a new event arrives, we fetch all events for that user, sort them by `eventTimeMs`, and "reduce" them to find the latest state. This naturally solves the problem of out-of-order and late-arriving events without complex state-machine logic.

### Claim-Based Workers

Both the **Carrier Poller** and the **Notification Sender** use a "Claim" pattern (`FOR UPDATE SKIP LOCKED`). This allows us to run multiple instances of the service horizontally without two workers ever processing the same user or notification simultaneously.

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

## 🛠️ Future Improvements

If I had more time and not exams:
