package edge

import (
	"context"
	"math"

	"github.com/avav/ai_trading_v1/internal/db"
)

type RollingSummary struct {
	Window    int
	Mean      float64
	Std       float64
	Stability float64
}

func RunRolling(ctx context.Context, store *db.EdgeStore, filter EdgeFilter, feature FeatureInfo, horizon LabelHorizon, windows []int) ([]RollingSummary, error) {
	var summaries []RollingSummary

	for _, w := range windows {
		points, err := store.RollingCorrelation(ctx, feature.Col, horizon.Col, filter.Symbol, filter.Timeframe, filter.FeatureSetID, w)
		if err != nil {
			return nil, err
		}

		var sum, sumSq float64
		var n int
		var overallSign float64

		for _, p := range points {
			if math.IsNaN(p.Corr) {
				continue
			}
			sum += p.Corr
			sumSq += p.Corr * p.Corr
			n++
			overallSign += p.Corr
		}

		if n == 0 {
			summaries = append(summaries, RollingSummary{Window: w})
			continue
		}

		mean := sum / float64(n)
		variance := sumSq/float64(n) - mean*mean
		if variance < 0 {
			variance = 0
		}
		std := math.Sqrt(variance)

		sameSignCount := 0
		for _, p := range points {
			if math.IsNaN(p.Corr) {
				continue
			}
			if (p.Corr >= 0 && overallSign >= 0) || (p.Corr < 0 && overallSign < 0) {
				sameSignCount++
			}
		}

		stability := float64(sameSignCount) / float64(n)

		summaries = append(summaries, RollingSummary{
			Window:    w,
			Mean:      mean,
			Std:       std,
			Stability: stability,
		})
	}

	return summaries, nil
}
