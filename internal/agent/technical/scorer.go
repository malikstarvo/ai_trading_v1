package technical

import (
	"math"
)

type Input struct {
	Price    float64
	EMA20    float64
	EMA50    float64
	EMA200   float64
	RSI14    float64
	ATR14    float64
	Volume   float64
	VolEMA20 float64
	ADX14    float64
}

type Components struct {
	Trend      float64
	Momentum   float64
	Volume     float64
	Volatility float64
	ADXBonus   float64
}

type Score struct {
	TechnicalScore float64
	Components    Components
}

func valid(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

func Calculate(input Input) Score {
	if !valid(input.Price) {
		return Score{}
	}

	trend := calcTrend(input)
	momentum := calcMomentum(input.RSI14)
	volume := calcVolume(input.Volume, input.VolEMA20)
	volatility := calcVolatility(input.ATR14, input.Price)
	adxBonus := calcADXBonus(input.ADX14)

	total := trend + momentum + volume + volatility + adxBonus
	if total > 100 {
		total = 100
	}
	if total < 0 {
		total = 0
	}

	return Score{
		TechnicalScore: total,
		Components: Components{
			Trend:      trend,
			Momentum:   momentum,
			Volume:     volume,
			Volatility: volatility,
			ADXBonus:   adxBonus,
		},
	}
}

func calcTrend(input Input) float64 {
	if !valid(input.EMA20) || !valid(input.EMA50) || !valid(input.EMA200) {
		return 0
	}

	price, e20, e50, e200 := input.Price, input.EMA20, input.EMA50, input.EMA200

	var base float64
	switch {
	case price > e20 && e20 > e50 && e50 > e200:
		base = 30
	case price > e20 && e20 > e50:
		base = 25
	case price > e20:
		base = 20
	case price < e20 && e20 > e50 && e50 > e200:
		base = 20
	case price > e50:
		base = 15
	case price > e200:
		base = 10
	default:
		base = 0
	}

	var alignment float64
	if e20 > e50 && e50 > e200 {
		alignment = 5
	}

	result := base + alignment
	if result > 35 {
		result = 35
	}
	return result
}

func calcMomentum(rsi float64) float64 {
	if !valid(rsi) {
		return 0
	}

	switch {
	case rsi >= 45 && rsi <= 65:
		// Sweet spot: trending zone, highest score
		return 22 + ((rsi-45)/20)*8
	case rsi >= 30 && rsi < 45:
		// Weak bearish, rising to sweet spot
		return 15 + ((rsi-30)/15)*7
	case rsi > 65 && rsi <= 75:
		// Overbought early, decreasing from sweet spot
		return 22 - ((rsi-65)/10)*7
	case rsi > 75:
		// Extreme overbought, tapers to 0
		val := 15 - ((rsi-75)/25)*15
		if val < 0 {
			val = 0
		}
		return val
	case rsi >= 0 && rsi < 30:
		// Oversold zone, low but rising (potential reversal)
		return 5 + (rsi/30)*10
	default:
		return 0
	}
}

func calcVolume(volume, volEMA20 float64) float64 {
	if !valid(volume) || !valid(volEMA20) || volEMA20 <= 0 {
		return 0
	}

	ratio := volume / volEMA20
	score := math.Log1p(ratio) * 12
	if score > 20 {
		score = 20
	}
	if score < 0 {
		score = 0
	}
	return score
}

func calcVolatility(atr, price float64) float64 {
	if !valid(atr) || !valid(price) || price <= 0 || atr <= 0 {
		return 0
	}

	atrPct := atr / price * 100

	switch {
	case atrPct < 1.0:
		return 3
	case atrPct < 2.0:
		return 5
	case atrPct < 3.5:
		return 10
	case atrPct < 5.0:
		return 7
	default:
		return 4
	}
}

func calcADXBonus(adx float64) float64 {
	if !valid(adx) || adx <= 0 {
		return 0
	}

	switch {
	case adx >= 35:
		return 10
	case adx >= 25:
		return 7
	case adx >= 20:
		return 3
	default:
		return 0
	}
}
