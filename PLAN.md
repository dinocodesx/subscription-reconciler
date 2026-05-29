# Backend Engineer Assignment — Subscription Reconciler

# Premium Entitlement Reconciler Assignment

## The problem

We sell a premium subscription through three different sales channels, and a user's premium access can be granted or revoked by **any** of them.

We need a single backend service that:

- Ingests signals from all three channels
- Maintains the canonical truth for:
  > Is this user premium right now, and why?
- Triggers a notification when access is about to lapse

The three channels behave very differently.

---

## Sales channels

### 1. In-app store

The in-app store pushes webhooks to us.

Important behaviour:

- Delivery is **at-least-once**
- Events have **no ordering guarantee**
- Events may occasionally arrive **days late**

---

### 2. Mobile carrier

The mobile carrier does **not** push updates.

We must poll their API on a schedule to discover whether the user's plan is still active.

---

### 3. Third-party marketplace

The marketplace pushes a single bulk request once a month:

> Revoke premium access for these users.

---

# What to build

## 1. Store webhook ingestion

Implement:

`POST /webhooks/store`

Request body:

```
{
  "eventId":"evt_abc123",
  "userId":"u_42",
  "type":"INITIAL_PURCHASE | RENEWAL | CANCELLATION | BILLING_ISSUE | EXPIRATION | UN_CANCELLATION",
  "eventTimeMs":1716700000000,
  "productId":"premium_monthly"
}
```

The service should ingest the event and update the user's entitlement state when appropriate.

Things to consider:

- Duplicate events
- Out-of-order events
- Late-arriving events
- State transitions caused by different event types

---

## 2. Carrier polling

A background job should run **every 5 minutes**.

For each user whose current entitlement source is:

`CARRIER`

Call the mocked carrier endpoint:

```
GET /mock/carrier/plan?userId=u_42
```

Example response:

```
{
  "status":"active"
}
```

Possible statuses:

```
active
inactive
api_error
```

Implement the mock carrier endpoint as a stub in your repo.

The response should be randomised, for example:

- 85% active
- 10% inactive
- 5% api_error

Important production constraint:

> More than one instance of the worker may be live at the same time.

Design the polling job accordingly.

---

## 3. Marketplace bulk revoke

Implement:

`POST /webhooks/marketplace/revoke`

Request body:

```
{
  "userIds": ["u_42","u_91","u_133"]
}
```

This should revoke marketplace-granted access for every listed user.

Only marketplace-granted access should be revoked by this endpoint.

---

## 4. Entitlement read endpoint

Implement:

`GET /users/:id/entitlement`

Example response:

```
{
  "active":true,
  "source":"STORE | CARRIER | MARKETPLACE | NONE",
  "expiresAt":"2026-06-10T00:00:00Z",
  "lastChangedAt":"2026-05-20T11:23:00Z",
  "reason":"RENEWAL"
}
```

The endpoint should return the canonical current entitlement state for the user.

Possible sources:

```
STORE
CARRIER
MARKETPLACE
NONE
```

---

## 5. Expiration notification

When a user's access is set to expire within the next 24 hours, schedule a one-time notification:

```
your premium expires soon
```

For this assignment, sending the notification means inserting a row into a `notifications` table.

Notification row shape:

```
{
  "userId":"u_42",
  "type":"PREMIUM_EXPIRES_SOON",
  "scheduledFor":"2026-06-09T00:00:00Z",
  "sentAt":null
}
```

A separate worker should pick up notifications at the scheduled time and mark `sentAt`.

Important rule:

> A user should receive each expiring-soon notification at most once.

---

# Constraints

## Storage

Use one of:

- Postgres
- SQLite

## Language

Use one of:

- TypeScript / Node.js
- Go

Pick one.

## Runnable setup

The project must be runnable with a single command:

```
docker compose up
```

Or an equivalent single command.

This should boot:

- The service
- The database
- The mock carrier endpoint

---

# Deliverables

## 1. Git repository

Submit either:

- A GitHub / GitLab repository link
- Or a `.zip` file

We will read your commit history.

Please commit as you would on a real branch.

---

## 2. README.md

Your `README.md` should cover:

- How to run the project
- API examples
- Design decisions
- Tradeoffs considered
- What you would change if you had another week

---

## 3. Tests

Include tests.

We are not prescribing exactly what to test.

Pick the cases you would want to defend in a code review.

Examples of valuable test areas:

- Duplicate store webhook events
- Out-of-order store webhook events
- Late-arriving store webhook events
- Marketplace revoke only affecting marketplace entitlements
- Carrier polling handling inactive plans
- Carrier polling handling API errors
- Expiring-soon notification being scheduled once only
- Concurrent workers not double-processing the same work

---

## 4. docker-compose.yml

Include a `docker-compose.yml` file so the full system can run with one command.

---

# Stretch work

Optional.

Do not start stretch work until the core is solid.

## 1. Audit log

Add an audit log of every entitlement state transition.

Each audit log entry should include:

- Triggering event ID, if available
- Source
- Previous state
- Next state
- Timestamp

---

## 2. Timeline endpoint

Implement:

`GET /users/:id/timeline`

This should return the reconstructed history of the user's entitlement changes from the audit log.

---

## Stretch submission rule

If you do stretch work:

1. Commit the core implementation first
2. Commit stretch work on top

This lets us clearly see the line between core and stretch work.

---

# Submission

Reply to the email with:

- The repo link or attached `.zip`
- Any notes you want to flag
