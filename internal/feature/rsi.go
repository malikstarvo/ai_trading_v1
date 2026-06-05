package feature

import "math"

func RSI(closes []float64, period int) []float64 {
	if period < 1 || len(closes) < period+1 {
		result := make([]float64, len(closes))
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}
	result := make([]float64, len(closes))
	for i := range result {
		result[i] = math.NaN()
	}

	var avgGain, avgLoss float64

	for i := 1; i < len(closes); i++ {
		change := closes[i] - closes[i-1]
		gain := math.Max(change, 0)
		loss := math.Max(-change, 0)

		if i < period {
			avgGain += gain
			avgLoss += loss
		} else if i == period {
			avgGain = (avgGain + gain) / float64(period)
			avgLoss = (avgLoss + loss) / float64(period)
			result[i] = computeRSI(avgGain, avgLoss)
		} else {
			avgGain = (avgGain*float64(period-1) + gain) / float64(period)
			avgLoss = (avgLoss*float64(period-1) + loss) / float64(period)
			result[i] = computeRSI(avgGain, avgLoss)
		}
	}
	return result
}

func computeRSI(avgGain, avgLoss float64) float64 {
	if avgLoss == 0 {
		if avgGain == 0 {
			return 50
		}
		return 100
	}
	rs := avgGain / avgLoss
	return 100 - 100/(1+rs)
}
