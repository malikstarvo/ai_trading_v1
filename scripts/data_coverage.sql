-- Data Coverage Report
-- Usage: PGPASSWORD=trader_pass psql -h localhost -U trader -d ai_trading -f scripts/data_coverage.sql

\x on

\echo '=== CANDLE COVERAGE ==='
SELECT symbol, timeframe,
       COUNT(*) AS total_candles,
       MIN(time) AS first_ts,
       MAX(time) AS last_ts,
       ROUND(EXTRACT(EPOCH FROM (MAX(time) - MIN(time))) / 86400, 1) AS span_days
FROM candles
GROUP BY symbol, timeframe
ORDER BY symbol, timeframe;

\echo ''
\echo '=== CANDLE GAPS (15m, >30min gap — top 10) ==='
WITH cte AS (
  SELECT symbol, time,
         LAG(time) OVER (PARTITION BY symbol ORDER BY time) AS prev_time
  FROM candles WHERE timeframe = '15m'
)
SELECT symbol, prev_time, time,
       ROUND(EXTRACT(EPOCH FROM (time - prev_time)) / 60, 0) AS gap_minutes
FROM cte
WHERE prev_time IS NOT NULL AND time > prev_time + interval '30 minutes'
ORDER BY gap_minutes DESC LIMIT 10;

\echo ''
\echo '=== FEATURE COVERAGE ==='
SELECT symbol, timeframe,
       COUNT(*) AS total_features,
       MIN(ts) AS first_ts,
       MAX(ts) AS last_ts,
       ROUND(EXTRACT(EPOCH FROM (MAX(ts) - MIN(ts))) / 86400, 1) AS span_days
FROM feature_values
GROUP BY symbol, timeframe
ORDER BY symbol, timeframe;

\echo ''
\echo '=== FEATURE vs CANDLE LAG (BTCUSDT 15m) ==='
SELECT c.symbol, c.timeframe,
       MAX(c.time) AS last_candle,
       MAX(f.ts) AS last_feature,
       ROUND(EXTRACT(EPOCH FROM (MAX(c.time) - COALESCE(MAX(f.ts), '1970-01-01'))) / 60, 0) AS lag_minutes
FROM candles c
LEFT JOIN feature_values f ON f.symbol = c.symbol AND f.timeframe = c.timeframe
WHERE c.symbol = 'BTCUSDT' AND c.timeframe = '15m'
GROUP BY c.symbol, c.timeframe;

\echo ''
\echo '=== LABEL COVERAGE ==='
SELECT symbol, timeframe,
       COUNT(*) AS total_labels,
       MIN(ts) AS first_ts,
       MAX(ts) AS last_ts
FROM training_labels
GROUP BY symbol, timeframe
ORDER BY symbol, timeframe;

\echo ''
\echo '=== PAPER TRADING SUMMARY ==='
SELECT 'snapshots' AS tbl, COUNT(*) FROM paper_account_snapshots
UNION ALL
SELECT 'open_positions', COUNT(*) FROM paper_positions WHERE status = 'open'
UNION ALL
SELECT 'closed_positions', COUNT(*) FROM paper_positions WHERE status = 'closed'
UNION ALL
SELECT 'completed_trades', COUNT(*) FROM paper_trades
UNION ALL
SELECT 'orders', COUNT(*) FROM paper_orders
ORDER BY tbl;

\echo ''
\echo '=== LATEST SNAPSHOTS (last 5) ==='
SELECT ts, balance, equity,
       ROUND(unrealized_pnl::numeric, 2) AS unrealized_pnl,
       ROUND(day_pnl::numeric, 2) AS day_pnl,
       day_trades
FROM paper_account_snapshots
ORDER BY ts DESC LIMIT 5;

\echo ''
\echo '=== OPEN POSITIONS ==='
SELECT id, symbol, timeframe, direction, entry_price, quantity,
       stop_price, open_ts, bars_held
FROM paper_positions
WHERE status = 'open'
ORDER BY open_ts DESC;

\echo ''
\echo '=== RECENT TRADES (last 5) ==='
SELECT id, symbol, direction,
       entry_ts, exit_ts,
       entry_price, exit_price,
       ROUND(net_pnl::numeric, 2) AS net_pnl,
       ROUND(return_pct::numeric, 4) AS return_pct,
       holding_bars, exit_reason
FROM paper_trades
ORDER BY exit_ts DESC NULLS LAST LIMIT 5;

\echo ''
\echo '=== NaN COVERAGE PER FEATURE (BTCUSDT 15m) ==='
SELECT
  COUNT(*)                                                                  AS total_rows,
  COUNT(oi_delta_1_pct)                                                     AS non_null_oi1,
  ROUND((1 - COUNT(oi_delta_1_pct)::numeric / COUNT(*)) * 100, 1)          AS nan_pct_oi1,
  COUNT(oi_delta_4_pct)                                                     AS non_null_oi4,
  ROUND((1 - COUNT(oi_delta_4_pct)::numeric / COUNT(*)) * 100, 1)          AS nan_pct_oi4,
  COUNT(oi_delta_12_pct)                                                    AS non_null_oi12,
  ROUND((1 - COUNT(oi_delta_12_pct)::numeric / COUNT(*)) * 100, 1)         AS nan_pct_oi12,
  COUNT(ls_ratio_raw)                                                       AS non_null_ls,
  ROUND((1 - COUNT(ls_ratio_raw)::numeric / COUNT(*)) * 100, 1)            AS nan_pct_ls,
  COUNT(funding_rate)                                                       AS non_null_fr,
  ROUND((1 - COUNT(funding_rate)::numeric / COUNT(*)) * 100, 1)            AS nan_pct_fr,
  COUNT(liq_long_usd)                                                       AS non_null_liq_l,
  ROUND((1 - COUNT(liq_long_usd)::numeric / COUNT(*)) * 100, 1)            AS nan_pct_liq_l,
  COUNT(liq_short_usd)                                                      AS non_null_liq_s,
  ROUND((1 - COUNT(liq_short_usd)::numeric / COUNT(*)) * 100, 1)           AS nan_pct_liq_s
FROM feature_values
WHERE symbol = 'BTCUSDT' AND timeframe = '15m';

\echo ''
\echo '=== ALL TABLES ROW COUNT ==='
SELECT 'candles' AS tbl, COUNT(*) FROM candles
UNION ALL SELECT 'feature_values', COUNT(*) FROM feature_values
UNION ALL SELECT 'training_labels', COUNT(*) FROM training_labels
UNION ALL SELECT 'paper_account_snapshots', COUNT(*) FROM paper_account_snapshots
UNION ALL SELECT 'paper_positions', COUNT(*) FROM paper_positions
UNION ALL SELECT 'paper_trades', COUNT(*) FROM paper_trades
UNION ALL SELECT 'paper_orders', COUNT(*) FROM paper_orders
ORDER BY tbl;

\x off
