CREATE TABLE IF NOT EXISTS training_labels (
    symbol              TEXT NOT NULL,
    timeframe           TEXT NOT NULL,
    ts                  TIMESTAMPTZ NOT NULL,
    feature_set_id      INT NOT NULL DEFAULT 1,

    future_return_4     DOUBLE PRECISION,
    future_return_12    DOUBLE PRECISION,
    future_return_24    DOUBLE PRECISION,
    success_4           SMALLINT,
    success_12          SMALLINT,
    success_24          SMALLINT,

    PRIMARY KEY (symbol, timeframe, ts, feature_set_id)
);

SELECT create_hypertable('training_labels', 'ts', if_not_exists => TRUE);
