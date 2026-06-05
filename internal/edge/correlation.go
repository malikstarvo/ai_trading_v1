package edge

import (
	"context"

	"github.com/avav/ai_trading_v1/internal/db"
)

func RunCorrelation(ctx context.Context, store *db.EdgeStore, filter EdgeFilter, feature FeatureInfo, horizon LabelHorizon) (pearson, spearman float64, samples int, err error) {
	return store.Correlation(ctx, feature.Col, horizon.Col, filter.Symbol, filter.Timeframe, filter.FeatureSetID)
}
