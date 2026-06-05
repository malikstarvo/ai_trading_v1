CREATE TABLE IF NOT EXISTS candles (
    time      TIMESTAMPTZ NOT NULL,
    symbol    TEXT NOT NULL,
    timeframe TEXT NOT NULL,
    open      DOUBLE PRECISION,
    high      DOUBLE PRECISION,
    low       DOUBLE PRECISION,
    close     DOUBLE PRECISION,
    volume    DOUBLE PRECISION,
    PRIMARY KEY (time, symbol, timeframe)
);

SELECT create_hypertable('candles', 'time', if_not_exists => TRUE);

CREATE TABLE IF NOT EXISTS open_interest (
    time         TIMESTAMPTZ NOT NULL,
    symbol       TEXT NOT NULL,
    oi           DOUBLE PRECISION,
    oi_value_usd DOUBLE PRECISION,
    PRIMARY KEY (time, symbol)
);

SELECT create_hypertable('open_interest', 'time', if_not_exists => TRUE);

CREATE TABLE IF NOT EXISTS funding_rates (
    time       TIMESTAMPTZ NOT NULL,
    symbol     TEXT NOT NULL,
    rate       DOUBLE PRECISION,
    interval_h INTEGER,
    PRIMARY KEY (time, symbol)
);

SELECT create_hypertable('funding_rates', 'time', if_not_exists => TRUE);

CREATE TABLE IF NOT EXISTS ls_ratios (
    time       TIMESTAMPTZ NOT NULL,
    symbol     TEXT NOT NULL,
    period     TEXT NOT NULL,
    buy_ratio  DOUBLE PRECISION,
    sell_ratio DOUBLE PRECISION,
    PRIMARY KEY (time, symbol, period)
);

SELECT create_hypertable('ls_ratios', 'time', if_not_exists => TRUE);

CREATE TABLE IF NOT EXISTS liquidations (
    time      TIMESTAMPTZ NOT NULL,
    symbol    TEXT NOT NULL,
    side      TEXT,
    size      DOUBLE PRECISION,
    price     DOUBLE PRECISION,
    value_usd DOUBLE PRECISION
);

SELECT create_hypertable('liquidations', 'time', if_not_exists => TRUE);

CREATE TABLE IF NOT EXISTS collector_health (
    service_name   TEXT PRIMARY KEY,
    status         TEXT NOT NULL,
    last_success_at TIMESTAMPTZ,
    last_error_at  TIMESTAMPTZ,
    last_error_msg TEXT,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
