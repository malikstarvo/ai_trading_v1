package feature

import "math"

func ADX(high, low, close []float64, period int) []float64 {
	n := len(high)
	if period < 1 || n == 0 || len(low) != n || len(close) != n {
		result := make([]float64, n)
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}
	result := make([]float64, n)
	for i := range result {
		result[i] = math.NaN()
	}

	smoothedTR := make([]float64, n)
	smoothedPDM := make([]float64, n)
	smoothedNDM := make([]float64, n)
	dxValues := make([]float64, n)

	var sumTR, sumPDM, sumNDM float64

	for i := 0; i < n; i++ {
		tr := trueRange(high[i], low[i], close, i)
		pdm, ndm := directionalMovement(high, low, i)

		if i < period {
			sumTR += tr
			sumPDM += pdm
			sumNDM += ndm
			continue
		}

		if i == period {
			smoothedTR[i] = sumTR / float64(period)
			smoothedPDM[i] = sumPDM / float64(period)
			smoothedNDM[i] = sumNDM / float64(period)
		} else {
			smoothedTR[i] = (smoothedTR[i-1]*float64(period-1) + tr) / float64(period)
			smoothedPDM[i] = (smoothedPDM[i-1]*float64(period-1) + pdm) / float64(period)
			smoothedNDM[i] = (smoothedNDM[i-1]*float64(period-1) + ndm) / float64(period)
		}

		pdi := 100 * smoothedPDM[i] / smoothedTR[i]
		ndi := 100 * smoothedNDM[i] / smoothedTR[i]
		diSum := pdi + ndi

		if diSum == 0 {
			dxValues[i] = 0
		} else {
			dxValues[i] = 100 * math.Abs(pdi-ndi) / diSum
		}
	}

	var dxCount int
	var dxSum float64

	for i := period + 1; i < n; i++ {
		if dxCount < period {
			dxSum += dxValues[i]
			dxCount++
			if dxCount == period {
				result[i] = dxSum / float64(period)
			}
		} else {
			adx := (result[i-1]*float64(period-1) + dxValues[i]) / float64(period)
			result[i] = adx
		}
	}
	return result
}

func directionalMovement(high, low []float64, i int) (pdm, ndm float64) {
	if i == 0 {
		return 0, 0
	}
	upMove := high[i] - high[i-1]
	downMove := low[i-1] - low[i]

	if upMove > downMove && upMove > 0 {
		return upMove, 0
	}
	if downMove > upMove && downMove > 0 {
		return 0, downMove
	}
	return 0, 0
}
