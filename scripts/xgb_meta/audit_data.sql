-- Dataset Audit: NaN coverage per feature
-- Run on VPS: psql -h localhost -U trader -d ai_trading -f audit_data.sql

\echo '=== NaN Coverage Per Feature (BTCUSDT 15m) ==='
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
\echo '=== V1 Feature NaN Coverage (BTCUSDT 15m) ==='
SELECT
  COUNT(*)                                                                  AS total_rows,
  COUNT(technical_score)                                                    AS nn_tech,
  COUNT(regime_score)                                                       AS nn_regime,
  COUNT(confidence_score)                                                   AS nn_conf,
  COUNT(atr14)                                                              AS nn_atr,
  COUNT(adx14)                                                              AS nn_adx,
  COUNT(funding_rate)                                                       AS nn_fr,
  COUNT(volume_delta)                                                       AS nn_vol_delta
FROM (
  SELECT
    fv.atr14, fv.adx14, fv.funding_rate,
    -- nullable fields that feed into technical/regime confidence scores
    fv.ema20, fv.ema50, fv.ema200, fv.rsi14, fv.volume, fv.volume_ema20,
    fv.volatility_14
  FROM feature_values fv
  WHERE fv.symbol = 'BTCUSDT' AND fv.timeframe = '15m'
) sub;

\echo ''
\echo '=== Label Distribution (BTCUSDT 15m) ==='
SELECT
  horizon,
  COUNT(*)                                            AS total,
  SUM(success::int)                                   AS wins,
  ROUND(AVG(success::numeric) * 100, 1)               AS win_rate_pct,
  ROUND(PERCENTILE_CONT(0.25) WITHIN GROUP (ORDER BY ret), 4) AS p25_ret,
  ROUND(PERCENTILE_CONT(0.50) WITHIN GROUP (ORDER BY ret), 4) AS p50_ret,
  ROUND(PERCENTILE_CONT(0.75) WITHIN GROUP (ORDER BY ret), 4) AS p75_ret,
  ROUND(AVG(ret)::numeric, 4)                         AS mean_ret,
  ROUND(STDDEV(ret)::numeric, 4)                      AS std_ret
FROM (
  SELECT 4 AS horizon, success_4 AS success, future_return_4 AS ret FROM training_labels
  WHERE symbol = 'BTCUSDT' AND timeframe = '15m'
  UNION ALL
  SELECT 12, success_12, future_return_12 FROM training_labels
  WHERE symbol = 'BTCUSDT' AND timeframe = '15m'
  UNION ALL
  SELECT 24, success_24, future_return_24 FROM training_labels
  WHERE symbol = 'BTCUSDT' AND timeframe = '15m'
) sub
GROUP BY horizon
ORDER BY horizon;

\echo ''
\echo '=== Win Rate at Alternative Thresholds (BTCUSDT 15m) ==='
SELECT
  4   AS horizon,
  ROUND(COUNT(*) FILTER (WHERE future_return_4 >= 0.0025)::numeric / COUNT(*) * 100, 1) AS wr_025pct,
  ROUND(COUNT(*) FILTER (WHERE future_return_4 >= 0.005)::numeric  / COUNT(*) * 100, 1) AS wr_05pct,
  ROUND(COUNT(*) FILTER (WHERE future_return_4 >= 0.01)::numeric   / COUNT(*) * 100, 1) AS wr_1pct
FROM training_labels WHERE symbol = 'BTCUSDT' AND timeframe = '15m'
UNION ALL
SELECT
  12,
  ROUND(COUNT(*) FILTER (WHERE future_return_12 >= 0.0025)::numeric / COUNT(*) * 100, 1),
  ROUND(COUNT(*) FILTER (WHERE future_return_12 >= 0.005)::numeric  / COUNT(*) * 100, 1),
  ROUND(COUNT(*) FILTER (WHERE future_return_12 >= 0.01)::numeric   / COUNT(*) * 100, 1)
FROM training_labels WHERE symbol = 'BTCUSDT' AND timeframe = '15m'
UNION ALL
SELECT
  24,
  ROUND(COUNT(*) FILTER (WHERE future_return_24 >= 0.0025)::numeric / COUNT(*) * 100, 1),
  ROUND(COUNT(*) FILTER (WHERE future_return_24 >= 0.005)::numeric  / COUNT(*) * 100, 1),
  ROUND(COUNT(*) FILTER (WHERE future_return_24 >= 0.01)::numeric   / COUNT(*) * 100, 1)
FROM training_labels WHERE symbol = 'BTCUSDT' AND timeframe = '15m';
