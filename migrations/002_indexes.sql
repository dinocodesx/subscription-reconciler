CREATE INDEX IF NOT EXISTS idx_store_events_user_time
    ON store_events (user_id, event_time_ms, received_at);

CREATE INDEX IF NOT EXISTS idx_source_entitlements_carrier_next_poll
    ON source_entitlements (next_poll_at)
    WHERE source = 'CARRIER';

CREATE INDEX IF NOT EXISTS idx_entitlements_source
    ON entitlements (source);

CREATE INDEX IF NOT EXISTS idx_notifications_due
    ON notifications (scheduled_for)
    WHERE sent_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uniq_pending_expiry_notifications
    ON notifications (user_id, type, scheduled_for)
    WHERE sent_at IS NULL;
