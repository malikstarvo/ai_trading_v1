package feature

import "math"

func EMA(values []float64, period int) []float64 {
	if period < 1 || len(values) == 0 {
		return nil
	}
	result := make([]float64, len(values))
	alpha := 2.0 / float64(period+1)

	var sum float64
	for i := 0; i < len(values); i++ {
		if i < period {
			sum += values[i]
			result[i] = math.NaN()
		} else if i == period {
			ema := sum / float64(period)
			ema = alpha*values[i] + (1-alpha)*ema
			result[i] = ema
		} else {
			ema := alpha*values[i] + (1-alpha)*result[i-1]
			result[i] = ema
		}
	}
	return result
}

func PriceAboveEMA(closes []float64, ema []float64) []int8 {
	if len(closes) != len(ema) {
		return nil
	}
	result := make([]int8, len(closes))
	for i := range closes {
		if math.IsNaN(ema[i]) {
			result[i] = 0
		} else if closes[i] > ema[i] {
			result[i] = 1
		} else {
			result[i] = 0
		}
	}
	return result
}
