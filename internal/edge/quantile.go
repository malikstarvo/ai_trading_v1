package edge

import (
	"context"

	"github.com/avav/ai_trading_v1/internal/db"
)

func RunQuantile(ctx context.Context, store *db.EdgeStore, filter EdgeFilter, feature FeatureInfo, horizon LabelHorizon, nBuckets int) ([]db.QuantileRow, error) {
	buckets, err := store.Quantiles(ctx, feature.Col, horizon.Col, filter.Symbol, filter.Timeframe, filter.FeatureSetID, nBuckets)
	if err != nil {
		return nil, err
	}
	if len(buckets) == 0 {
		return nil, nil
	}

	return buckets, nil
}
