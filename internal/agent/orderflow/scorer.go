package orderflow

import (
	"math"
)

type Input struct {
	FundingRate float64 // decimal, e.g. 0.00002 = 0.002 bps
	OIDeltaPct  float64 // % change in open interest, e.g. 0.5 = +0.5%
	LSRatio     float64 // longs/shorts ratio, e.g. 1.2
	LongLiqUSD  float64 // long liquidation volume in USD
	ShortLiqUSD float64 // short liquidation volume in USD
}

type Components struct {
	FundingScore float64
	OIScore      float64
	LSScore      float64
	LiqScore     float64
}

type Score struct {
	OrderFlowScore float64
	Components     Components
}

func valid(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

// Calculate returns OrderFlowScore (0-100).
// NaN/Inf per-field falls back to neutral — not a market signal, just missing data.
func Calculate(input Input) Score {
	fundingScore := calcFundingScore(input.FundingRate)
	oiScore := calcOIScore(input.OIDeltaPct)
	lsScore := calcLSScore(input.LSRatio)
	liqScore := calcLiqScore(input.LongLiqUSD, input.ShortLiqUSD)

	total := fundingScore + oiScore + lsScore + liqScore
	if total > 100 {
		total = 100
	}
	if total < 0 {
		total = 0
	}

	return Score{
		OrderFlowScore: total,
		Components: Components{
			FundingScore: fundingScore,
			OIScore:      oiScore,
			LSScore:      lsScore,
			LiqScore:     liqScore,
		},
	}
}

// calcFundingScore — 0-20.
// Funding rate in bps (rate * 10000). Negative = bearish, 0-2 bps = sweet spot, >5 = overcrowded.
func calcFundingScore(fundingRate float64) float64 {
	if !valid(fundingRate) {
		return 10
	}

	bps := fundingRate * 10000

	switch {
	case bps <= 0:
		// Negative → bearish, linearly decreasing
		score := 10 + bps*3
		if score < 0 {
			score = 0
		}
		return score
	case bps < 2:
		// 0-2 bps sweet spot
		return 10 + bps*5
	case bps < 5:
		// 2-5 bps elevated
		score := 20 - (bps-2)*3.33
		if score < 10 {
			score = 10
		}
		return score
	default:
		// >=5 bps overcrowded
		score := 10 - (bps-5)*2
		if score < 0 {
			score = 0
		}
		return score
	}
}

// calcOIScore — 0-25.
// Rising OI = trend conviction. Diluted slope: score = 15 + oiDelta * 10, clamped.
func calcOIScore(oiDeltaPct float64) float64 {
	if !valid(oiDeltaPct) {
		return 12.5
	}

	score := 15 + oiDeltaPct*10
	if score > 25 {
		score = 25
	}
	if score < 0 {
		score = 0
	}
	return score
}

// calcLSScore — 0-30.
// 1.1-1.3 = healthy long bias. >2.0 = crowded. <0.7 = extreme bearish.
func calcLSScore(lsRatio float64) float64 {
	if !valid(lsRatio) || lsRatio <= 0 {
		return 10
	}

	switch {
	case lsRatio >= 1.0 && lsRatio <= 1.3:
		return 12 + (lsRatio-1.0)/0.3*6
	case lsRatio > 1.3 && lsRatio <= 2.0:
		score := 18 - (lsRatio-1.3)/0.7*8
		if score < 10 {
			score = 10
		}
		return score
	case lsRatio > 2.0:
		score := 10 - (lsRatio-2.0)*8
		if score < 2 {
			score = 2
		}
		return score
	default:
		// lsRatio < 1.0
		score := 12 - (1.0-lsRatio)*20
		if score < 0 {
			score = 0
		}
		return score
	}
}

// calcLiqScore — 0-25.
// Primary edge. Rewards large liquidation events with bonus for directional alignment.
// imbalance = (shortLiq - longLiq) / totalLiq. Negative = more long liq → bullish.
func calcLiqScore(longLiq, shortLiq float64) float64 {
	if !valid(longLiq) || !valid(shortLiq) {
		return 12.5
	}

	totalLiq := longLiq + shortLiq
	if totalLiq <= 0 {
		return 12.5
	}

	imbalance := (shortLiq - longLiq) / totalLiq // [-1, +1]
	norm := totalLiq / 50_000_000                 // normalized to $50M cap
	if norm > 1 {
		norm = 1
	}
	if norm < 0 {
		norm = 0
	}

	// direction: negative imbalance (more long liq) → adds to score
	// magnitude: raw event size reward, independent of direction
	direction := imbalance * norm * 8
	magnitude := norm * 4.5

	score := 12.5 - direction + magnitude
	if score > 25 {
		score = 25
	}
	if score < 0 {
		score = 0
	}
	return score
}
