package regime

import "math"

type Input struct {
	ADX14      float64
	ATR14      float64
	Price      float64
	Volatility float64
}

type Components struct {
	TrendScore float64
	VolScore   float64
}

type Score struct {
	RegimeScore float64
	Regime      string
	Components  Components
}

func valid(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

// Calculate returns RegimeScore 0-100 and a regime label.
// NaN/Inf per-field falls back to neutral midpoint (25).
// All inputs invalid → Regime="unknown" (not a market signal).
func Calculate(input Input) Score {
	if !valid(input.ADX14) && !valid(input.ATR14) && !valid(input.Price) && !valid(input.Volatility) {
		return Score{
			RegimeScore: 50,
			Regime:      "unknown",
			Components:  Components{TrendScore: 25, VolScore: 25},
		}
	}

	trendScore := calcTrendScore(input.ADX14)

	atrVolScore := calcVolFromATR(input.ATR14, input.Price)
	featureVolScore := calcVolFromFeature(input.Volatility)

	volScore := (atrVolScore + featureVolScore) / 2

	total := trendScore + volScore
	if total > 100 {
		total = 100
	}
	if total < 0 {
		total = 0
	}

	regime := classifyRegime(input.ADX14, volScore)

	return Score{
		RegimeScore: total,
		Regime:      regime,
		Components: Components{
			TrendScore: trendScore,
			VolScore:   volScore,
		},
	}
}

// calcTrendScore returns 0-50 from ADX14.
// ADX < 20 = ranging → 0. ADX ≥ 35 = strong trend → 50.
// ADX 20-35 = linear interpolation 0-50.
func calcTrendScore(adx float64) float64 {
	if !valid(adx) {
		return 25
	}
	if adx < 20 {
		return 0
	}
	if adx >= 35 {
		return 50
	}
	return (adx - 20) / 15 * 50
}

// calcVolFromATR returns 0-50 from ATR as % of price.
func calcVolFromATR(atr, price float64) float64 {
	if !valid(atr) || !valid(price) || price <= 0 || atr <= 0 {
		return 25
	}

	atrPct := atr / price * 100

	switch {
	case atrPct < 0.5:
		return 5
	case atrPct < 1.5:
		return 5 + (atrPct-0.5)/1.0*20
	case atrPct < 3.0:
		return 25 + (atrPct-1.5)/1.5*20
	case atrPct < 5.0:
		return 45 - (atrPct-3.0)/2.0*20
	default:
		return 15
	}
}

// calcVolFromFeature returns 0-50 from Volatility14 (log-return stddev %).
func calcVolFromFeature(vol float64) float64 {
	if !valid(vol) || vol <= 0 {
		return 25
	}

	switch {
	case vol < 0.3:
		return 5
	case vol < 1.0:
		return 5 + (vol-0.3)/0.7*20
	case vol < 2.5:
		return 25 + (vol-1.0)/1.5*20
	case vol < 4.0:
		return 45 - (vol-2.5)/1.5*15
	default:
		return 20
	}
}

// classifyRegime returns regime label based on ADX and volScore.
func classifyRegime(adx float64, volScore float64) string {
	if !valid(adx) {
		return "unknown"
	}

	isTrending := adx >= 25
	isHighVol := volScore >= 25

	switch {
	case isTrending && isHighVol:
		return "trending_high_vol"
	case isTrending && !isHighVol:
		return "trending_low_vol"
	case !isTrending && isHighVol:
		return "ranging_high_vol"
	default:
		return "ranging_low_vol"
	}
}
