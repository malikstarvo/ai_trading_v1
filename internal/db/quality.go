package db

import (
	"fmt"
	"math"
	"time"

	"github.com/avav/ai_trading_v1/internal/model"
)

func ValidateFeatureRows(rows []model.FeatureRow) error {
	if len(rows) == 0 {
		return fmt.Errorf("empty feature rows")
	}

	if err := checkDuplicateTimestamps(rows); err != nil {
		return err
	}

	floats := countFeatureFloats(rows)
	nanCount := countNan(rows)
	infCount := countInf(rows)

	nanRatio := float64(nanCount) / float64(floats)
	infRatio := float64(infCount) / float64(floats)

	if nanRatio > 0.05 {
		return fmt.Errorf("NaN ratio %.4f exceeds 5%% threshold", nanRatio)
	}
	if infRatio > 0.001 {
		return fmt.Errorf("Inf ratio %.4f exceeds 0.1%% threshold", infRatio)
	}

	return nil
}

func checkDuplicateTimestamps(rows []model.FeatureRow) error {
	seen := make(map[time.Time]bool)
	for _, r := range rows {
		if seen[r.Ts] {
			return fmt.Errorf("duplicate timestamp: %v", r.Ts)
		}
		seen[r.Ts] = true
	}
	return nil
}

func countFeatureFloats(rows []model.FeatureRow) int {
	const floatsPerRow = 24
	return len(rows) * floatsPerRow
}

func countNan(rows []model.FeatureRow) int {
	var n int
	for _, r := range rows {
		if math.IsNaN(r.EMA20) { n++ }
		if math.IsNaN(r.EMA50) { n++ }
		if math.IsNaN(r.EMA200) { n++ }
		if math.IsNaN(r.RSI14) { n++ }
		if math.IsNaN(r.ATR14) { n++ }
		if math.IsNaN(r.ADX14) { n++ }
		if math.IsNaN(r.VolumeEMA20) { n++ }
		if math.IsNaN(r.OIDelta1Pct) { n++ }
		if math.IsNaN(r.OIDelta4Pct) { n++ }
		if math.IsNaN(r.OIDelta12Pct) { n++ }
		if math.IsNaN(r.OIZScore30) { n++ }
		if math.IsNaN(r.FundingRate) { n++ }
		if math.IsNaN(r.FundingZScore30) { n++ }
		if math.IsNaN(r.LSRatioRaw) { n++ }
		if math.IsNaN(r.LSRatioNormalized) { n++ }
		if math.IsNaN(r.LiqLongUSD) { n++ }
		if math.IsNaN(r.LiqShortUSD) { n++ }
		if math.IsNaN(r.LiqImbalance) { n++ }
		if math.IsNaN(r.Return1) { n++ }
		if math.IsNaN(r.Return4) { n++ }
		if math.IsNaN(r.Return12) { n++ }
		if math.IsNaN(r.Volatility14) { n++ }
		if math.IsNaN(r.Volatility50) { n++ }
	}
	return n
}

func countInf(rows []model.FeatureRow) int {
	var n int
	for _, r := range rows {
		if math.IsInf(r.EMA20, 0) { n++ }
		if math.IsInf(r.EMA50, 0) { n++ }
		if math.IsInf(r.EMA200, 0) { n++ }
		if math.IsInf(r.RSI14, 0) { n++ }
		if math.IsInf(r.ATR14, 0) { n++ }
		if math.IsInf(r.ADX14, 0) { n++ }
		if math.IsInf(r.VolumeEMA20, 0) { n++ }
		if math.IsInf(r.OIDelta1Pct, 0) { n++ }
		if math.IsInf(r.OIDelta4Pct, 0) { n++ }
		if math.IsInf(r.OIDelta12Pct, 0) { n++ }
		if math.IsInf(r.OIZScore30, 0) { n++ }
		if math.IsInf(r.FundingRate, 0) { n++ }
		if math.IsInf(r.FundingZScore30, 0) { n++ }
		if math.IsInf(r.LSRatioRaw, 0) { n++ }
		if math.IsInf(r.LSRatioNormalized, 0) { n++ }
		if math.IsInf(r.LiqLongUSD, 0) { n++ }
		if math.IsInf(r.LiqShortUSD, 0) { n++ }
		if math.IsInf(r.LiqImbalance, 0) { n++ }
		if math.IsInf(r.Return1, 0) { n++ }
		if math.IsInf(r.Return4, 0) { n++ }
		if math.IsInf(r.Return12, 0) { n++ }
		if math.IsInf(r.Volatility14, 0) { n++ }
		if math.IsInf(r.Volatility50, 0) { n++ }
	}
	return n
}
