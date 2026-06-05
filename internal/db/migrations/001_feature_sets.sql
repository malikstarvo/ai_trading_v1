CREATE TABLE IF NOT EXISTS feature_sets (
    id          SERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    version     TEXT NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    active      BOOLEAN NOT NULL DEFAULT false,
    UNIQUE(name, version)
);
