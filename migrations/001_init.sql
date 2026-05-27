CREATE TABLE IF NOT EXISTS store_events (
    event_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    type TEXT NOT NULL,
    event_time_ms BIGINT NOT NULL,
    product_id TEXT NOT NULL,
    received_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS source_entitlements (
    user_id TEXT NOT NULL,
    source TEXT NOT NULL,
    active BOOLEAN NOT NULL,
    expires_at TIMESTAMPTZ NULL,
    last_changed_at TIMESTAMPTZ NOT NULL,
    reason TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    next_poll_at TIMESTAMPTZ NULL,
    PRIMARY KEY (user_id, source)
);

CREATE TABLE IF NOT EXISTS entitlements (
    user_id TEXT PRIMARY KEY,
    active BOOLEAN NOT NULL,
    source TEXT NOT NULL,
    expires_at TIMESTAMPTZ NULL,
    last_changed_at TIMESTAMPTZ NOT NULL,
    reason TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id TEXT NOT NULL,
    type TEXT NOT NULL,
    scheduled_for TIMESTAMPTZ NOT NULL,
    sent_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL
);
