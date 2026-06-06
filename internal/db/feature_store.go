package db

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/avav/ai_trading_v1/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FeatureStore struct {
	pool *pgxpool.Pool
}

func NewFeatureStore(pool *pgxpool.Pool) *FeatureStore {
	return &FeatureStore{pool: pool}
}

func (s *FeatureStore) Pool() *pgxpool.Pool {
	return s.pool
}

func (s *FeatureStore) GetActiveFeatureSetID(ctx context.Context) (int, error) {
	var id int
	err := s.pool.QueryRow(ctx, `SELECT id FROM feature_sets WHERE active = true LIMIT 1`).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("no active feature set: %w", err)
	}
	return id, nil
}

func (s *FeatureStore) EnsureDefaultFeatureSet(ctx context.Context) (int, error) {
	var id int
	err := s.pool.QueryRow(ctx, `
		INSERT INTO feature_sets (name, version, description, active)
		VALUES ('Feature Set V1', '1.0', 'Initial: 27 technical + orderflow + market features', true)
		ON CONFLICT (name, version) DO UPDATE SET active = true
		RETURNING id
	`).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("ensure feature set: %w", err)
	}
	return id, nil
}

func (s *FeatureStore) LoadCandlesAfter(ctx context.Context, symbol, timeframe string, after time.Time, limit int) ([]model.Candle, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT time, symbol, timeframe, open, high, low, close, volume
		FROM candles
		WHERE symbol = $1 AND timeframe = $2 AND time > $3
		ORDER BY time ASC
		LIMIT $4
	`, symbol, timeframe, after, limit)
	if err != nil {
		return nil, fmt.Errorf("query candles: %w", err)
	}
	defer rows.Close()

	var candles []model.Candle
	for rows.Next() {
		var c model.Candle
		if err := rows.Scan(&c.Time, &c.Symbol, &c.Timeframe, &c.Open, &c.High, &c.Low, &c.Close, &c.Volume); err != nil {
			return nil, fmt.Errorf("scan candle: %w", err)
		}
		candles = append(candles, c)
	}
	return candles, rows.Err()
}

func (s *FeatureStore) LoadOrderFlow(ctx context.Context, symbol, timeframe string, timestamps []time.Time) ([]model.OrderFlowSnapshot, error) {
	if len(timestamps) == 0 {
		return nil, nil
	}
	tfDuration := parseDuration(timeframe)
	if tfDuration == 0 {
		tfDuration = 15 * time.Minute
	}

	start := timestamps[0]
	end := timestamps[len(timestamps)-1]

	oiRecords := s.loadOI(ctx, symbol, start, end)
	frRecords := s.loadFunding(ctx, symbol, start, end)
	lsRecords := s.loadLS(ctx, symbol, start, end)

	liqData, err := s.loadLiqAggregated(ctx, symbol, start, end, tfDuration)
	if err != nil {
		return nil, fmt.Errorf("liq aggregated: %w", err)
	}

	var oiIdx, frIdx, lsIdx int
	result := make([]model.OrderFlowSnapshot, len(timestamps))

	for i, ts := range timestamps {
		snap := model.OrderFlowSnapshot{Ts: ts}

		for oiIdx < len(oiRecords) && !oiRecords[oiIdx].t.After(ts) {
			snap.OI = oiRecords[oiIdx].oi
			oiIdx++
		}

		for frIdx < len(frRecords) && !frRecords[frIdx].t.After(ts) {
			snap.FundingRate = frRecords[frIdx].rate
			frIdx++
		}

		for lsIdx < len(lsRecords) && !lsRecords[lsIdx].t.After(ts) {
			snap.LSBuyRatio = lsRecords[lsIdx].buyRatio
			snap.LSSellRatio = lsRecords[lsIdx].sellRatio
			lsIdx++
		}

		bucket := ts.Truncate(tfDuration)
		if ld, ok := liqData[bucket]; ok {
			snap.LiqLongUSD = ld.longUSD
			snap.LiqShortUSD = ld.shortUSD
		}

		result[i] = snap
	}
	return result, nil
}

type liqSnapshot struct {
	longUSD  float64
	shortUSD float64
}

type oiRecord struct {
	t  time.Time
	oi float64
}

type frRecord struct {
	t    time.Time
	rate float64
}

type lsRecord struct {
	t         time.Time
	buyRatio  float64
	sellRatio float64
}

func (s *FeatureStore) loadOI(ctx context.Context, symbol string, start, end time.Time) []oiRecord {
	rows, _ := s.pool.Query(ctx, `
		SELECT time, oi FROM open_interest
		WHERE symbol = $1 AND time BETWEEN $2 AND $3
		ORDER BY time ASC
	`, symbol, start, end)
	defer rows.Close()

	var records []oiRecord
	for rows.Next() {
		var r oiRecord
		rows.Scan(&r.t, &r.oi)
		records = append(records, r)
	}
	return records
}

func (s *FeatureStore) loadFunding(ctx context.Context, symbol string, start, end time.Time) []frRecord {
	rows, _ := s.pool.Query(ctx, `
		SELECT time, rate FROM funding_rates
		WHERE symbol = $1 AND time BETWEEN $2 AND $3
		ORDER BY time ASC
	`, symbol, start, end)
	defer rows.Close()

	var records []frRecord
	for rows.Next() {
		var r frRecord
		rows.Scan(&r.t, &r.rate)
		records = append(records, r)
	}
	return records
}

func (s *FeatureStore) loadLS(ctx context.Context, symbol string, start, end time.Time) []lsRecord {
	rows, _ := s.pool.Query(ctx, `
		SELECT time, buy_ratio, sell_ratio FROM ls_ratios
		WHERE symbol = $1 AND time BETWEEN $2 AND $3
		ORDER BY time ASC
	`, symbol, start, end)
	defer rows.Close()

	var records []lsRecord
	for rows.Next() {
		var r lsRecord
		rows.Scan(&r.t, &r.buyRatio, &r.sellRatio)
		records = append(records, r)
	}
	return records
}

func (s *FeatureStore) loadLiqAggregated(ctx context.Context, symbol string, start, end time.Time, bucket time.Duration) (map[time.Time]liqSnapshot, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			time_bucket($4::interval, time) AS bucket,
			COALESCE(SUM(value_usd) FILTER (WHERE side = 'Sell'), 0) AS long_liq_usd,
			COALESCE(SUM(value_usd) FILTER (WHERE side = 'Buy'), 0) AS short_liq_usd
		FROM liquidations
		WHERE symbol = $1 AND time BETWEEN $2 AND $3
		GROUP BY bucket
		ORDER BY bucket ASC
	`, symbol, start, end, bucket.String())
	if err != nil {
		return nil, fmt.Errorf("query liq: %w", err)
	}
	defer rows.Close()

	result := make(map[time.Time]liqSnapshot)
	for rows.Next() {
		var bucket time.Time
		var ls liqSnapshot
		if err := rows.Scan(&bucket, &ls.longUSD, &ls.shortUSD); err != nil {
			return nil, fmt.Errorf("scan liq: %w", err)
		}
		result[bucket] = ls
	}
	return result, rows.Err()
}

func (s *FeatureStore) UpsertFeatures(ctx context.Context, rows []model.FeatureRow) error {
	if len(rows) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, r := range rows {
		batch.Queue(`
			INSERT INTO feature_values (
				symbol, timeframe, ts, feature_set_id,
				ema20, ema50, ema200, rsi14, atr14, adx14, volume_ema20,
				price_above_ema20, price_above_ema50, price_above_ema200,
				oi_delta_1_pct, oi_delta_4_pct, oi_delta_12_pct,
				oi_zscore_30, funding_rate, funding_zscore_30,
				ls_ratio_raw, ls_ratio_normalized,
				liq_long_usd, liq_short_usd, liq_imbalance,
				return_1, return_4, return_12,
				volatility_14, volatility_50
			) VALUES (
				$1,$2,$3,$4,
				$5,$6,$7,$8,$9,$10,$11,
				$12,$13,$14,
				$15,$16,$17,$18,$19,$20,
				$21,$22,$23,$24,$25,
				$26,$27,$28,$29,$30
			) ON CONFLICT (symbol, timeframe, ts, feature_set_id) DO UPDATE SET
				ema20 = EXCLUDED.ema20,
				ema50 = EXCLUDED.ema50,
				ema200 = EXCLUDED.ema200,
				rsi14 = EXCLUDED.rsi14,
				atr14 = EXCLUDED.atr14,
				adx14 = EXCLUDED.adx14,
				volume_ema20 = EXCLUDED.volume_ema20,
				price_above_ema20 = EXCLUDED.price_above_ema20,
				price_above_ema50 = EXCLUDED.price_above_ema50,
				price_above_ema200 = EXCLUDED.price_above_ema200,
				oi_delta_1_pct = EXCLUDED.oi_delta_1_pct,
				oi_delta_4_pct = EXCLUDED.oi_delta_4_pct,
				oi_delta_12_pct = EXCLUDED.oi_delta_12_pct,
				oi_zscore_30 = EXCLUDED.oi_zscore_30,
				funding_rate = EXCLUDED.funding_rate,
				funding_zscore_30 = EXCLUDED.funding_zscore_30,
				ls_ratio_raw = EXCLUDED.ls_ratio_raw,
				ls_ratio_normalized = EXCLUDED.ls_ratio_normalized,
				liq_long_usd = EXCLUDED.liq_long_usd,
				liq_short_usd = EXCLUDED.liq_short_usd,
				liq_imbalance = EXCLUDED.liq_imbalance,
				return_1 = EXCLUDED.return_1,
				return_4 = EXCLUDED.return_4,
				return_12 = EXCLUDED.return_12,
				volatility_14 = EXCLUDED.volatility_14,
				volatility_50 = EXCLUDED.volatility_50
		`,
			r.Symbol, r.Timeframe, r.Ts, r.FeatureSetID,
			nanToNull(r.EMA20), nanToNull(r.EMA50), nanToNull(r.EMA200),
			nanToNull(r.RSI14), nanToNull(r.ATR14), nanToNull(r.ADX14),
			nanToNull(r.VolumeEMA20),
			r.PriceAboveEMA20, r.PriceAboveEMA50, r.PriceAboveEMA200,
			nanToNull(r.OIDelta1Pct), nanToNull(r.OIDelta4Pct), nanToNull(r.OIDelta12Pct),
			nanToNull(r.OIZScore30), nanToNull(r.FundingRate), nanToNull(r.FundingZScore30),
			nanToNull(r.LSRatioRaw), nanToNull(r.LSRatioNormalized),
			nanToNull(r.LiqLongUSD), nanToNull(r.LiqShortUSD), nanToNull(r.LiqImbalance),
			nanToNull(r.Return1), nanToNull(r.Return4), nanToNull(r.Return12),
			nanToNull(r.Volatility14), nanToNull(r.Volatility50),
		)
	}

	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range rows {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("upsert feature row: %w", err)
		}
	}
	return nil
}

func (s *FeatureStore) UpsertLabels(ctx context.Context, rows []model.LabelRow, featureSetID int) error {
	if len(rows) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, r := range rows {
		batch.Queue(`
			INSERT INTO training_labels (
				symbol, timeframe, ts, feature_set_id,
				future_return_4, future_return_12, future_return_24,
				success_4, success_12, success_24
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
			ON CONFLICT (symbol, timeframe, ts, feature_set_id) DO UPDATE SET
				future_return_4 = EXCLUDED.future_return_4,
				future_return_12 = EXCLUDED.future_return_12,
				future_return_24 = EXCLUDED.future_return_24,
				success_4 = EXCLUDED.success_4,
				success_12 = EXCLUDED.success_12,
				success_24 = EXCLUDED.success_24
		`,
			r.Symbol, r.Timeframe, r.Ts, featureSetID,
			nanToNull(r.FutureReturn4), nanToNull(r.FutureReturn12), nanToNull(r.FutureReturn24),
			r.Success4, r.Success12, r.Success24,
		)
	}

	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range rows {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("upsert label row: %w", err)
		}
	}
	return nil
}

func nanToNull(v float64) *float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return nil
	}
	return &v
}

func (s *FeatureStore) LoadFeaturesAfter(ctx context.Context, symbol, timeframe string, featureSetID int, after time.Time) ([]model.FeatureRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			symbol, timeframe, ts, feature_set_id,
			ema20, ema50, ema200, rsi14, atr14, adx14, volume_ema20,
			price_above_ema20, price_above_ema50, price_above_ema200,
			COALESCE(oi_delta_1_pct, 0), COALESCE(oi_delta_4_pct, 0), COALESCE(oi_delta_12_pct, 0),
			COALESCE(oi_zscore_30, 0), COALESCE(funding_rate, 0), COALESCE(funding_zscore_30, 0),
			COALESCE(ls_ratio_raw, 0), COALESCE(ls_ratio_normalized, 0),
			COALESCE(liq_long_usd, 0), COALESCE(liq_short_usd, 0), COALESCE(liq_imbalance, 0),
			COALESCE(return_1, 0), COALESCE(return_4, 0), COALESCE(return_12, 0),
			COALESCE(volatility_14, 0), COALESCE(volatility_50, 0)
		FROM feature_values
		WHERE symbol = $1 AND timeframe = $2 AND feature_set_id = $3 AND ts > $4
		ORDER BY ts ASC
	`, symbol, timeframe, featureSetID, after)
	if err != nil {
		return nil, fmt.Errorf("query features: %w", err)
	}
	defer rows.Close()

	var features []model.FeatureRow
	for rows.Next() {
		var f model.FeatureRow
		var pAbove20, pAbove50, pAbove200 int8
		if err := rows.Scan(
			&f.Symbol, &f.Timeframe, &f.Ts, &f.FeatureSetID,
			&f.EMA20, &f.EMA50, &f.EMA200, &f.RSI14, &f.ATR14, &f.ADX14, &f.VolumeEMA20,
			&pAbove20, &pAbove50, &pAbove200,
			&f.OIDelta1Pct, &f.OIDelta4Pct, &f.OIDelta12Pct,
			&f.OIZScore30, &f.FundingRate, &f.FundingZScore30,
			&f.LSRatioRaw, &f.LSRatioNormalized,
			&f.LiqLongUSD, &f.LiqShortUSD, &f.LiqImbalance,
			&f.Return1, &f.Return4, &f.Return12,
			&f.Volatility14, &f.Volatility50,
		); err != nil {
			return nil, fmt.Errorf("scan feature: %w", err)
		}
		f.PriceAboveEMA20 = pAbove20
		f.PriceAboveEMA50 = pAbove50
		f.PriceAboveEMA200 = pAbove200
		features = append(features, f)
	}
	return features, rows.Err()
}

func parseDuration(timeframe string) time.Duration {
	switch timeframe {
	case "15m":
		return 15 * time.Minute
	case "1h":
		return time.Hour
	case "4h":
		return 4 * time.Hour
	case "1d":
		return 24 * time.Hour
	default:
		return 0
	}
}
