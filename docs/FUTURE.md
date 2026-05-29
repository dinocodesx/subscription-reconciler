# Future Improvements: Implementation Roadmap

This document provides a highly detailed technical specification for the improvements mentioned in the README. Each section outlines the specific architectural changes, required dependencies, and integration points within the existing codebase.

---

## 1. Observability & Monitoring

To gain deep visibility into the system's performance and health, we will integrate Prometheus for metrics and OpenTelemetry for distributed tracing.

### Prometheus Metrics
We will use the official `github.com/prometheus/client_golang` package.

#### Implementation Details:
1.  **Metric Registry:** Initialize a global Prometheus registry in `internal/app/app.go`.
2.  **Instrumentation:**
    -   **HTTP Handlers:** Wrap the router in `internal/http/router.go` with a middleware that records request duration (histogram) and response codes (counter) for every endpoint.
    -   **Webhook Latency:** In `internal/http/handlers/store_webhook.go`, track the end-to-end processing time for `ApplyStoreEvent` using a histogram: `subscription_webhook_processing_seconds`.
    -   **Worker Jobs:** In `internal/worker/loop.go`, instrument the run loops to export:
        -   `subscription_jobs_total{worker="carrier_poller", status="success|failure"}`
        -   `subscription_jobs_total{worker="notification_sender", status="success|failure"}`
3.  **Metrics Endpoint:** Expose a `/metrics` endpoint on a separate management port (e.g., `9090`) or as a protected internal route to allow Prometheus to scrape the data.

### OpenTelemetry (OTel) Tracing
We will use `go.opentelemetry.io/otel` and `go.opentelemetry.io/otel/sdk`.

#### Implementation Details:
1.  **Tracer Provider:** Initialize the OTel SDK in `internal/app/app.go`, configuring an exporter to send spans to a collector (e.g., Jaeger).
2.  **Context Propagation:**
    -   **HTTP:** Use `otelhttp` middleware in the router to automatically start spans for incoming requests.
    -   **Domain Layer:** Ensure `context.Context` is passed consistently through `internal/domain/entitlement/resolver.go`. Manually create spans for complex logic, such as the precedence resolution loop.
    -   **Database:** Integrate `pgx` with a tracer (like `github.com/exaring/otelpgx`) to automatically capture SQL queries and their performance within the transaction context.
3.  **Visualization:** Add a `jaeger-all-in-one` container to the `docker-compose.yml` for local span visualization.

---

## 2. Resilience: Dead Letter Queues (DLQ)

To ensure high reliability for store webhooks, we will implement a DLQ pattern to handle transient failures (e.g., database deadlocks, network timeouts).

### Database Schema
Create a new migration `migrations/004_dlq.sql`:
```sql
CREATE TABLE store_webhook_dlq (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    raw_payload JSONB NOT NULL,
    error_reason TEXT,
    retry_count INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    next_retry_at TIMESTAMPTZ DEFAULT NOW(),
    last_error_at TIMESTAMPTZ,
    permanently_failed BOOLEAN DEFAULT FALSE
);
CREATE INDEX idx_dlq_retry ON store_webhook_dlq (next_retry_at) WHERE NOT permanently_failed;
```

### Handler Integration
Modify `internal/http/handlers/store_webhook.go`:
1.  When `h.store.ApplyStoreEvent` returns a transient error (which can be identified by wrapping database errors with a specific `IsTransient()` helper), instead of returning a `500 Internal Server Error`, the handler will call a new repository method `h.store.EnqueueDLQ(ctx, payload)`.
2.  Return a `202 Accepted` to the caller, indicating the event is received and will be retried asynchronously.

### DLQ Replay Worker
Create `internal/worker/dlq_replay.go`:
1.  A new background loop that queries the `store_webhook_dlq` table for rows where `next_retry_at <= NOW()` and `permanently_failed = FALSE`.
2.  For each row:
    -   Attempt to process the `raw_payload` using the existing `ApplyStoreEvent` logic.
    -   **On Success:** Delete the row from the DLQ.
    -   **On Failure:** Increment `retry_count`, update `last_error_at`, and calculate the `next_retry_at` using an exponential backoff (e.g., `pow(2, retry_count) * 1 minute`). If `retry_count > 10`, set `permanently_failed = TRUE`.

---

## 3. Scalability: Message Queues

We will transition from a polling-based worker model to an event-driven model using RabbitMQ or AWS SQS to handle higher throughput and decouple processing from the database.

### Event Emission
1.  **Interface:** Define an `EventManager` interface in `internal/domain/events/interface.go`.
2.  **Publisher:** Implement a publisher (e.g., `internal/infrastructure/rabbitmq/publisher.go`) that serializes events to JSON and sends them to specific exchanges.
3.  **Integration:** Whenever the entitlement state changes in `internal/repository/postgres/store_events.go`, emit an `EntitlementChangedEvent`.

### Worker Refactor
Refactor `internal/worker/carrier_poll.go` and `internal/worker/notifications.go`:
1.  **Consumer Logic:** Instead of a `time.Ticker` loop, the workers will use a library like `github.com/rabbitmq/amqp091-go` to maintain a persistent connection to the queue.
2.  **Topic Based Routing:**
    -   **Carrier Worker:** Listens to `carrier.sync.queue`.
    -   **Notification Worker:** Listens to `notification.dispatch.queue`.
3.  **Payloads:** The message payload will contain only the `UserID`. The worker will then fetch the latest state from the database, perform its check, and execute its task.
4.  **Ack/Nack:** Use message acknowledgments (`Ack`) on success and negative acknowledgments (`Nack`) with requeueing for transient failures, leveraging the message queue's built-in retry mechanisms.

---

## 4. Cache Layer (Redis)

To handle high-frequency reads for the `GET /users/:id/entitlement` endpoint, we will introduce Redis as a caching layer.

### Integration & Setup
1.  **Client:** Add `github.com/redis/go-redis/v9` as a dependency.
2.  **Initialization:** In `internal/app/app.go`, initialize the Redis client and inject it into the router and handlers.

### Caching Strategy
Modify `internal/http/handlers/entitlement.go`:
1.  **Read Path (Cache Aside):**
    -   Construct a key: `entitlement:{user_id}`.
    -   Attempt `redis.Get(ctx, key)`.
    -   **On Hit:** Deserialize the JSON and return it immediately.
    -   **On Miss:** Call `h.store.GetEntitlement(ctx, userID)`, serialize the result, and call `redis.Set(ctx, key, result, 1*time.Hour)`.
2.  **Write Path (Cache Invalidation):**
    -   This is the most critical part. To prevent serving stale data, the cache **must** be invalidated whenever a user's state changes.
    -   In `internal/http/handlers/store_webhook.go`, after a successful `ApplyStoreEvent`, call `redis.Del(ctx, "entitlement:"+payload.UserID)`.
    -   In `internal/http/handlers/marketplace_webhook.go`, after a successful bulk revoke, iterate through the list of `userIds` and issue `redis.Del` for each one (potentially using a Redis pipeline for efficiency).
    -   In `internal/worker/carrier_poll.go`, if the carrier poll results in an entitlement update, invalidate the cache for that user.

---

## Implementation Sequence
1.  **Phase 1:** Observability (Prometheus/Jaeger) to establish a baseline for performance.
2.  **Phase 2:** Redis Caching to immediately reduce database read load.
3.  **Phase 3:** DLQ for Store Webhooks to improve ingestion resilience.
4.  **Phase 4:** Message Queues for workers to enable horizontal scalability.
