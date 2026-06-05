CREATE TABLE IF NOT EXISTS research_artifacts (
    id              BIGSERIAL PRIMARY KEY,
    research_run_id BIGINT NOT NULL REFERENCES research_runs(id),
    artifact_type   TEXT NOT NULL,
    file_path       TEXT NOT NULL,
    description     TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
