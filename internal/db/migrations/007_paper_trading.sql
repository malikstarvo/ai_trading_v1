CREATE TABLE IF NOT EXISTS paper_orders (
    id              BIGSERIAL PRIMARY KEY,
    symbol          TEXT NOT NULL,
    timeframe       TEXT NOT NULL,
    direction       TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'created',
    requested_size  DOUBLE PRECISION NOT NULL,
    filled_size     DOUBLE PRECISION NOT NULL DEFAULT 0,
    fill_price      DOUBLE PRECISION,
    slippage_pct    DOUBLE PRECISION,
    commission      DOUBLE PRECISION NOT NULL DEFAULT 0,
    reason          JSONB,
    open_ts         TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS paper_fills (
    id              BIGSERIAL PRIMARY KEY,
    order_id        BIGINT NOT NULL REFERENCES paper_orders(id),
    ts              TIMESTAMPTZ NOT NULL,
    side            TEXT NOT NULL,
    price           DOUBLE PRECISION NOT NULL,
    size            DOUBLE PRECISION NOT NULL,
    fee             DOUBLE PRECISION NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS paper_positions (
    id              BIGSERIAL PRIMARY KEY,
    symbol          TEXT NOT NULL,
    timeframe       TEXT NOT NULL,
    direction       TEXT NOT NULL,
    entry_order_id  BIGINT REFERENCES paper_orders(id),
    quantity        DOUBLE PRECISION NOT NULL,
    entry_price     DOUBLE PRECISION NOT NULL,
    entry_fee       DOUBLE PRECISION NOT NULL DEFAULT 0,
    stop_price      DOUBLE PRECISION,
    open_ts         TIMESTAMPTZ NOT NULL,
    bars_held       INT NOT NULL DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'open'
);

CREATE TABLE IF NOT EXISTS paper_trades (
    id              BIGSERIAL PRIMARY KEY,
    position_id     BIGINT REFERENCES paper_positions(id),
    symbol          TEXT NOT NULL,
    timeframe       TEXT NOT NULL,
    direction       TEXT NOT NULL,
    entry_ts        TIMESTAMPTZ NOT NULL,
    exit_ts         TIMESTAMPTZ,
    entry_price     DOUBLE PRECISION NOT NULL,
    exit_price      DOUBLE PRECISION,
    size            DOUBLE PRECISION NOT NULL,
    gross_pnl       DOUBLE PRECISION,
    commission      DOUBLE PRECISION,
    net_pnl         DOUBLE PRECISION,
    return_pct      DOUBLE PRECISION,
    holding_bars    INT,
    exit_reason     TEXT,
    entry_reason    JSONB,
    feature_snapshot JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS paper_account_snapshots (
    ts              TIMESTAMPTZ NOT NULL,
    balance         DOUBLE PRECISION NOT NULL,
    equity          DOUBLE PRECISION NOT NULL,
    unrealized_pnl  DOUBLE PRECISION NOT NULL,
    day_pnl         DOUBLE PRECISION NOT NULL,
    day_trades      INT NOT NULL
);

SELECT create_hypertable('paper_account_snapshots', 'ts', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_paper_orders_status ON paper_orders(status);
CREATE INDEX IF NOT EXISTS idx_paper_positions_status ON paper_positions(status);
CREATE INDEX IF NOT EXISTS idx_paper_trades_entry_ts ON paper_trades(entry_ts DESC);
