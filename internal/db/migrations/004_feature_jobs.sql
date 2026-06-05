CREATE TABLE IF NOT EXISTS feature_jobs (
    id              BIGSERIAL PRIMARY KEY,
    symbol          TEXT NOT NULL,
    timeframe       TEXT NOT NULL,
    feature_set_id  INT REFERENCES feature_sets(id),
    params          JSONB,
    status          TEXT NOT NULL DEFAULT 'pending',
    processed_rows  INT DEFAULT 0,
    error           TEXT,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    last_heartbeat  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
