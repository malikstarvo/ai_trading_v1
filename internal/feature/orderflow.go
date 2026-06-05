package feature

import "math"

func OIDeltaPct(OI []float64, periods int) []float64 {
	if periods < 1 || len(OI) == 0 {
		result := make([]float64, len(OI))
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}
	result := make([]float64, len(OI))
	for i := range OI {
		if i < periods {
			result[i] = math.NaN()
		} else {
			prev := OI[i-periods]
			if prev == 0 {
				result[i] = math.NaN()
			} else {
				result[i] = ((OI[i] - prev) / prev) * 100
			}
		}
	}
	return result
}

func NormalizedImbalance(a, b []float64) []float64 {
	n := len(a)
	if n != len(b) {
		result := make([]float64, n)
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}
	result := make([]float64, n)
	for i := 0; i < n; i++ {
		denom := a[i] + b[i]
		if denom == 0 {
			result[i] = 0
		} else {
			result[i] = (a[i] - b[i]) / denom
		}
	}
	return result
}

func LSRatioRaw(buyRatio, sellRatio []float64) []float64 {
	n := len(buyRatio)
	if n != len(sellRatio) {
		result := make([]float64, n)
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}
	result := make([]float64, n)
	for i := 0; i < n; i++ {
		if sellRatio[i] == 0 {
			result[i] = math.NaN()
		} else {
			result[i] = buyRatio[i] / sellRatio[i]
		}
	}
	return result
}
