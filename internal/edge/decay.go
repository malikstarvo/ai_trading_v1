package edge

import "math"

type DecayResult struct {
	PeakCorr    float64
	AvgCorr     float64
	DecayRate   float64
	Persistence float64
}

func AnalyzeDecay(correlations []float64) DecayResult {
	var absCorrs []float64
	for _, c := range correlations {
		absCorrs = append(absCorrs, math.Abs(c))
	}

	if len(absCorrs) == 0 {
		return DecayResult{}
	}

	peak := absCorrs[0]
	sum := absCorrs[0]
	for i := 1; i < len(absCorrs); i++ {
		if absCorrs[i] > peak {
			peak = absCorrs[i]
		}
		sum += absCorrs[i]
	}
	avgCorr := sum / float64(len(absCorrs))

	var decayRate float64
	if len(absCorrs) >= 2 {
		decayRate = (absCorrs[0] - absCorrs[len(absCorrs)-1]) / float64(len(absCorrs)-1)
	}

	persistence := absCorrs[0]
	for _, c := range absCorrs {
		if c < persistence {
			persistence = c
		}
	}

	return DecayResult{
		PeakCorr:    peak,
		AvgCorr:     avgCorr,
		DecayRate:   decayRate,
		Persistence: persistence,
	}
}
