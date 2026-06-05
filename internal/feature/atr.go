package feature

import "math"

func ATR(high, low, close []float64, period int) []float64 {
	n := len(high)
	if period < 1 || n == 0 || len(low) != n || len(close) != n {
		result := make([]float64, n)
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}
	result := make([]float64, n)

	var sumTR float64
	for i := 0; i < n; i++ {
		tr := trueRange(high[i], low[i], close, i)
		if i < period {
			sumTR += tr
			result[i] = math.NaN()
		} else if i == period {
			atr := sumTR / float64(period)
			result[i] = atr
		} else {
			atr := (result[i-1]*float64(period-1) + tr) / float64(period)
			result[i] = atr
		}
	}
	return result
}

func trueRange(h, l float64, close []float64, i int) float64 {
	hl := h - l
	if i == 0 {
		return hl
	}
	hc := math.Abs(h - close[i-1])
	lc := math.Abs(l - close[i-1])
	return math.Max(hl, math.Max(hc, lc))
}
