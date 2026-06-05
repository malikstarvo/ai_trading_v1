package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EdgeStore struct {
	pool *pgxpool.Pool
}

func NewEdgeStore(pool *pgxpool.Pool) *EdgeStore {
	return &EdgeStore{pool: pool}
}

func (s *EdgeStore) Correlation(ctx context.Context, featureCol, labelCol string, symbol, timeframe string, featureSetID int) (pearson, spearman float64, samples int, err error) {
	var p, sp *float64
	var cnt int
	err = s.pool.QueryRow(ctx, fmt.Sprintf(`
		WITH ranked AS (
			SELECT fv.%s AS feature, tl.%s AS label,
				percent_rank() OVER (ORDER BY fv.%s) AS f_rank,
				percent_rank() OVER (ORDER BY tl.%s) AS l_rank
			FROM feature_values fv
			JOIN training_labels tl ON fv.symbol = tl.symbol AND fv.timeframe = tl.timeframe AND fv.ts = tl.ts AND fv.feature_set_id = tl.feature_set_id
			WHERE fv.symbol = $1 AND fv.timeframe = $2 AND fv.feature_set_id = $3
			  AND fv.%s IS NOT NULL AND tl.%s IS NOT NULL
		)
		SELECT CORR(feature, label), CORR(f_rank, l_rank), COUNT(*)
		FROM ranked
	`, featureCol, labelCol, featureCol, labelCol, featureCol, labelCol), symbol, timeframe, featureSetID).Scan(&p, &sp, &cnt)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("correlation: %w", err)
	}
	if p != nil {
		pearson = *p
	}
	if sp != nil {
		spearman = *sp
	}
	return pearson, spearman, cnt, nil
}

func (s *EdgeStore) CountFeatures(ctx context.Context, featureSetID int) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM feature_values WHERE feature_set_id = $1`, featureSetID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count features: %w", err)
	}
	return count, nil
}

func (s *EdgeStore) Quantiles(ctx context.Context, featureCol, labelCol string, symbol, timeframe string, featureSetID, nBuckets int) ([]QuantileRow, error) {
	var rows []QuantileRow
	sql := fmt.Sprintf(`
		WITH ranked AS (
			SELECT
				fv.%s AS feature,
				tl.%s AS label,
				NTILE($4) OVER (ORDER BY fv.%s) AS bucket
			FROM feature_values fv
			JOIN training_labels tl ON fv.symbol = tl.symbol AND fv.timeframe = tl.timeframe AND fv.ts = tl.ts AND fv.feature_set_id = tl.feature_set_id
			WHERE fv.symbol = $1 AND fv.timeframe = $2 AND fv.feature_set_id = $3
			  AND fv.%s IS NOT NULL AND tl.%s IS NOT NULL
		)
		SELECT
			bucket,
			COUNT(*),
			AVG(label),
			SUM(CASE WHEN label > 0 THEN 1.0 ELSE 0.0 END) / COUNT(*)::float,
			CASE
				WHEN SUM(CASE WHEN label <= 0 THEN label ELSE 0 END) = 0 THEN 0
				ELSE SUM(CASE WHEN label > 0 THEN label ELSE 0 END) / ABS(SUM(CASE WHEN label <= 0 THEN label ELSE 0 END))
			END
		FROM ranked
		GROUP BY bucket
		ORDER BY bucket
	`, featureCol, labelCol, featureCol, featureCol, labelCol)

	q, err := s.pool.Query(ctx, sql, symbol, timeframe, featureSetID, nBuckets)
	if err != nil {
		return nil, fmt.Errorf("quantiles: %w", err)
	}
	defer q.Close()

	for q.Next() {
		var r QuantileRow
		if err := q.Scan(&r.Bucket, &r.Trades, &r.AvgReturn, &r.WinRate, &r.ProfitFactor); err != nil {
			return nil, fmt.Errorf("scan quantile: %w", err)
		}
		rows = append(rows, r)
	}
	return rows, q.Err()
}

func (s *EdgeStore) RollingCorrelation(ctx context.Context, featureCol, labelCol string, symbol, timeframe string, featureSetID, window int) ([]RollingPoint, error) {
	var points []RollingPoint
	sql := fmt.Sprintf(`
		SELECT
			fv.ts,
			COALESCE(CORR(fv.%s, tl.%s) OVER (ORDER BY fv.ts ROWS BETWEEN %d PRECEDING AND CURRENT ROW), 0) AS rolling_corr
		FROM feature_values fv
		JOIN training_labels tl ON fv.symbol = tl.symbol AND fv.timeframe = tl.timeframe AND fv.ts = tl.ts AND fv.feature_set_id = tl.feature_set_id
		WHERE fv.symbol = $1 AND fv.timeframe = $2 AND fv.feature_set_id = $3
		  AND fv.%s IS NOT NULL AND tl.%s IS NOT NULL
		ORDER BY fv.ts
	`, featureCol, labelCol, window-1, featureCol, labelCol)

	q, err := s.pool.Query(ctx, sql, symbol, timeframe, featureSetID)
	if err != nil {
		return nil, fmt.Errorf("rolling correlation: %w", err)
	}
	defer q.Close()

	for q.Next() {
		var p RollingPoint
		if err := q.Scan(&p.Ts, &p.Corr); err != nil {
			return nil, fmt.Errorf("scan rolling: %w", err)
		}
		points = append(points, p)
	}
	return points, q.Err()
}

func (s *EdgeStore) RegimeCorrelations(ctx context.Context, featureCol, labelCol string, symbol, timeframe string, featureSetID int, pctThreshold float64) ([]RegimeRow, error) {
	var rows []RegimeRow
	sql := fmt.Sprintf(`
		WITH ranked AS (
			SELECT
				fv.%s AS feature,
				tl.%s AS label,
				CASE WHEN PERCENT_RANK() OVER (ORDER BY fv.adx14) >= %f THEN 'trending' ELSE 'ranging' END AS trend_regime,
				CASE WHEN PERCENT_RANK() OVER (ORDER BY fv.atr14) >= %f THEN 'high_vol' ELSE 'low_vol' END AS vol_regime
			FROM feature_values fv
			JOIN training_labels tl ON fv.symbol = tl.symbol AND fv.timeframe = tl.timeframe AND fv.ts = tl.ts AND fv.feature_set_id = tl.feature_set_id
			WHERE fv.symbol = $1 AND fv.timeframe = $2 AND fv.feature_set_id = $3
			  AND fv.%s IS NOT NULL AND tl.%s IS NOT NULL
			  AND fv.adx14 IS NOT NULL AND fv.atr14 IS NOT NULL
		)
		SELECT trend_regime, vol_regime, CORR(feature, label), COUNT(*)
		FROM ranked
		GROUP BY trend_regime, vol_regime
	`, featureCol, labelCol, 1.0-pctThreshold, 1.0-pctThreshold, featureCol, labelCol)

	q, err := s.pool.Query(ctx, sql, symbol, timeframe, featureSetID)
	if err != nil {
		return nil, fmt.Errorf("regime: %w", err)
	}
	defer q.Close()

	for q.Next() {
		var r RegimeRow
		var corr *float64
		if err := q.Scan(&r.TrendRegime, &r.VolRegime, &corr, &r.Samples); err != nil {
			return nil, fmt.Errorf("scan regime: %w", err)
		}
		if corr != nil {
			r.Corr = *corr
		}
		rows = append(rows, r)
	}
	return rows, q.Err()
}

func (s *EdgeStore) LoadTimestamps(ctx context.Context, symbol, timeframe string, featureSetID int) ([]time.Time, error) {
	q, err := s.pool.Query(ctx, `
		SELECT ts FROM feature_values
		WHERE symbol = $1 AND timeframe = $2 AND feature_set_id = $3
		  AND ema20 IS NOT NULL
		ORDER BY ts ASC
	`, symbol, timeframe, featureSetID)
	if err != nil {
		return nil, err
	}
	defer q.Close()

	var ts []time.Time
	for q.Next() {
		var t time.Time
		if err := q.Scan(&t); err != nil {
			return nil, err
		}
		ts = append(ts, t)
	}
	return ts, q.Err()
}
