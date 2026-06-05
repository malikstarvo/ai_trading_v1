package db

import (
	"fmt"
	"log"
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

	// Mandatory features: candle-based — must pass thresholds
	mfloats := countMandatoryFloats(rows)
	mnan := countMandatoryNaN(rows)
	minf := countMandatoryInf(rows)

	mnanRatio := float64(mnan) / float64(mfloats)
	minfRatio := float64(minf) / float64(mfloats)

	if mnanRatio > 0.05 {
		return fmt.Errorf("mandatory NaN ratio %.4f exceeds 5%% threshold", mnanRatio)
	}
	if minfRatio > 0.001 {
		return fmt.Errorf("mandatory Inf ratio %.4f exceeds 0.1%% threshold", minfRatio)
	}

	// Optional features: orderflow — log warning, never reject
	if counts := countOptionalNaNByField(rows); len(counts) > 0 {
		reportOptionalNaN(counts, len(rows))
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

const mandatoryFieldsPerRow = 12

func countMandatoryFloats(rows []model.FeatureRow) int {
	return len(rows) * mandatoryFieldsPerRow
}

func countMandatoryNaN(rows []model.FeatureRow) int {
	var n int
	for _, r := range rows {
		if math.IsNaN(r.EMA20) { n++ }
		if math.IsNaN(r.EMA50) { n++ }
		if math.IsNaN(r.EMA200) { n++ }
		if math.IsNaN(r.RSI14) { n++ }
		if math.IsNaN(r.ATR14) { n++ }
		if math.IsNaN(r.ADX14) { n++ }
		if math.IsNaN(r.VolumeEMA20) { n++ }
		if math.IsNaN(r.Return1) { n++ }
		if math.IsNaN(r.Return4) { n++ }
		if math.IsNaN(r.Return12) { n++ }
		if math.IsNaN(r.Volatility14) { n++ }
		if math.IsNaN(r.Volatility50) { n++ }
	}
	return n
}

func countMandatoryInf(rows []model.FeatureRow) int {
	var n int
	for _, r := range rows {
		if math.IsInf(r.EMA20, 0) { n++ }
		if math.IsInf(r.EMA50, 0) { n++ }
		if math.IsInf(r.EMA200, 0) { n++ }
		if math.IsInf(r.RSI14, 0) { n++ }
		if math.IsInf(r.ATR14, 0) { n++ }
		if math.IsInf(r.ADX14, 0) { n++ }
		if math.IsInf(r.VolumeEMA20, 0) { n++ }
		if math.IsInf(r.Return1, 0) { n++ }
		if math.IsInf(r.Return4, 0) { n++ }
		if math.IsInf(r.Return12, 0) { n++ }
		if math.IsInf(r.Volatility14, 0) { n++ }
		if math.IsInf(r.Volatility50, 0) { n++ }
	}
	return n
}

type optionalFieldCount struct {
	Name string
	NaN  int
}

func countOptionalNaNByField(rows []model.FeatureRow) []optionalFieldCount {
	total := len(rows)
	if total == 0 {
		return nil
	}

	names := []string{
		"OIDelta1Pct", "OIDelta4Pct", "OIDelta12Pct",
		"OIZScore30", "FundingRate", "FundingZScore30",
		"LSRatioRaw", "LSRatioNormalized",
		"LiqLongUSD", "LiqShortUSD", "LiqImbalance",
	}
	counts := make([]optionalFieldCount, len(names))
	for i, name := range names {
		counts[i].Name = name
	}

	for _, r := range rows {
		if math.IsNaN(r.OIDelta1Pct) { counts[0].NaN++ }
		if math.IsNaN(r.OIDelta4Pct) { counts[1].NaN++ }
		if math.IsNaN(r.OIDelta12Pct) { counts[2].NaN++ }
		if math.IsNaN(r.OIZScore30) { counts[3].NaN++ }
		if math.IsNaN(r.FundingRate) { counts[4].NaN++ }
		if math.IsNaN(r.FundingZScore30) { counts[5].NaN++ }
		if math.IsNaN(r.LSRatioRaw) { counts[6].NaN++ }
		if math.IsNaN(r.LSRatioNormalized) { counts[7].NaN++ }
		if math.IsNaN(r.LiqLongUSD) { counts[8].NaN++ }
		if math.IsNaN(r.LiqShortUSD) { counts[9].NaN++ }
		if math.IsNaN(r.LiqImbalance) { counts[10].NaN++ }
	}

	return counts
}

func reportOptionalNaN(counts []optionalFieldCount, total int) {
	hasAny := false
	for _, c := range counts {
		if c.NaN > 0 {
			hasAny = true
			break
		}
	}
	if !hasAny {
		return
	}

	log.Printf("WARNING: orderflow features unavailable")
	for _, c := range counts {
		if c.NaN > 0 {
			pct := float64(c.NaN) * 100 / float64(total)
			log.Printf("  %-20s %.0f%% NaN (%d/%d)", c.Name, pct, c.NaN, total)
		}
	}
}
