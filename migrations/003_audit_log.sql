CREATE TABLE IF NOT EXISTS audit_log (
    id BIGSERIAL PRIMARY KEY,
    user_id TEXT NOT NULL,
    event_id TEXT NULL,
    source TEXT NOT NULL,
    previous_active BOOLEAN NOT NULL,
    previous_source TEXT NOT NULL,
    previous_reason TEXT NOT NULL,
    next_active BOOLEAN NOT NULL,
    next_source TEXT NOT NULL,
    next_reason TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_audit_log_user_time
    ON audit_log (user_id, created_at DESC, id DESC);
