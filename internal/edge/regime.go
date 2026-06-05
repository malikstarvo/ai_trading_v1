package edge

import (
	"context"
	"math"

	"github.com/avav/ai_trading_v1/internal/db"
)

type RegimeResult struct {
	TrendRegime string
	VolRegime   string
	Corr        float64
	Samples     int
}

func RunRegime(ctx context.Context, store *db.EdgeStore, filter EdgeFilter, feature FeatureInfo, horizon LabelHorizon, pctThreshold float64) ([]RegimeResult, error) {
	rows, err := store.RegimeCorrelations(ctx, feature.Col, horizon.Col, filter.Symbol, filter.Timeframe, filter.FeatureSetID, pctThreshold)
	if err != nil {
		return nil, err
	}

	var results []RegimeResult
	for _, r := range rows {
		results = append(results, RegimeResult{
			TrendRegime: r.TrendRegime,
			VolRegime:   r.VolRegime,
			Corr:        r.Corr,
			Samples:     r.Samples,
		})
	}
	return results, nil
}

func RegimeConsistency(regimes []RegimeResult) float64 {
	if len(regimes) == 0 {
		return 0
	}

	var corrs []float64
	for _, r := range regimes {
		if !math.IsNaN(r.Corr) {
			corrs = append(corrs, r.Corr)
		}
	}
	if len(corrs) < 2 {
		return 1.0
	}

	maxC := corrs[0]
	minC := corrs[0]
	for _, c := range corrs {
		if c > maxC {
			maxC = c
		}
		if c < minC {
			minC = c
		}
	}
	return 1.0 - (maxC - minC)
}
