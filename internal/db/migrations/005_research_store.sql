CREATE TABLE IF NOT EXISTS research_runs (
    id              BIGSERIAL PRIMARY KEY,
    name            TEXT NOT NULL,
    symbol          TEXT NOT NULL,
    timeframe       TEXT NOT NULL,
    feature_set_id  INT REFERENCES feature_sets(id),
    config          JSONB,
    status          TEXT NOT NULL DEFAULT 'pending',
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS research_results (
    id              BIGSERIAL PRIMARY KEY,
    research_run_id BIGINT NOT NULL REFERENCES research_runs(id),
    feature_name    TEXT NOT NULL,
    metric_name     TEXT NOT NULL,
    metric_value    DOUBLE PRECISION NOT NULL,
    samples         INT NOT NULL,
    metadata        JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
