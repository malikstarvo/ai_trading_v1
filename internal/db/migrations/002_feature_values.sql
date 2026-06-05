CREATE TABLE IF NOT EXISTS feature_values (
    symbol              TEXT NOT NULL,
    timeframe           TEXT NOT NULL,
    ts                  TIMESTAMPTZ NOT NULL,
    feature_set_id      INT NOT NULL REFERENCES feature_sets(id),

    ema20               DOUBLE PRECISION,
    ema50               DOUBLE PRECISION,
    ema200              DOUBLE PRECISION,
    rsi14               DOUBLE PRECISION,
    atr14               DOUBLE PRECISION,
    adx14               DOUBLE PRECISION,
    volume_ema20        DOUBLE PRECISION,
    price_above_ema20   SMALLINT,
    price_above_ema50   SMALLINT,
    price_above_ema200  SMALLINT,

    oi_delta_1_pct      DOUBLE PRECISION,
    oi_delta_4_pct      DOUBLE PRECISION,
    oi_delta_12_pct     DOUBLE PRECISION,
    oi_zscore_30        DOUBLE PRECISION,
    funding_rate        DOUBLE PRECISION,
    funding_zscore_30   DOUBLE PRECISION,
    ls_ratio_raw        DOUBLE PRECISION,
    ls_ratio_normalized DOUBLE PRECISION,
    liq_long_usd        DOUBLE PRECISION,
    liq_short_usd       DOUBLE PRECISION,
    liq_imbalance       DOUBLE PRECISION,

    return_1            DOUBLE PRECISION,
    return_4            DOUBLE PRECISION,
    return_12           DOUBLE PRECISION,
    volatility_14       DOUBLE PRECISION,
    volatility_50       DOUBLE PRECISION,

    PRIMARY KEY (symbol, timeframe, ts, feature_set_id)
);

SELECT create_hypertable('feature_values', 'ts', if_not_exists => TRUE);
