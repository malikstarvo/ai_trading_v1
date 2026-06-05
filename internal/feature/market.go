package feature

import "math"

func LogReturn(closes []float64, periods int) []float64 {
	if periods < 1 || len(closes) == 0 {
		result := make([]float64, len(closes))
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}
	result := make([]float64, len(closes))
	for i := range closes {
		if i < periods {
			result[i] = math.NaN()
		} else {
			result[i] = math.Log(closes[i]/closes[i-periods]) * 100
		}
	}
	return result
}

func Volatility(logReturns []float64, window int) []float64 {
	if window < 2 || len(logReturns) == 0 {
		result := make([]float64, len(logReturns))
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}
	result := make([]float64, len(logReturns))
	for i := range logReturns {
		if i < window-1 {
			result[i] = math.NaN()
		} else {
			result[i] = stdDev(logReturns[i-window+1 : i+1])
		}
	}
	return result
}

func ZScore(values []float64, window int) []float64 {
	if window < 2 || len(values) == 0 {
		result := make([]float64, len(values))
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}
	result := make([]float64, len(values))
	for i := range values {
		if i < window-1 {
			result[i] = math.NaN()
		} else {
			slice := values[i-window+1 : i+1]
			mean := mean(slice)
			sd := stdDev(slice)
			if sd == 0 {
				result[i] = 0
			} else {
				result[i] = (values[i] - mean) / sd
			}
		}
	}
	return result
}

func mean(v []float64) float64 {
	var sum float64
	for _, x := range v {
		sum += x
	}
	return sum / float64(len(v))
}

func stdDev(v []float64) float64 {
	m := mean(v)
	var sumSq float64
	for _, x := range v {
		d := x - m
		sumSq += d * d
	}
	return math.Sqrt(sumSq / float64(len(v)))
}
