package edge

import "sort"

type ComponentScore struct {
	FeatureName           string
	QuantilePF            float64
	QuantileWRDelta       float64
	RollingStability      float64
	RegimeConsistencyVal  float64
	AvgAbsCorrelation     float64
	CompositeScore        float64
}

func ComputeRanking(scores []ComponentScore) []ComponentScore {
	if len(scores) == 0 {
		return nil
	}

	ranked := make([]ComponentScore, len(scores))
	copy(ranked, scores)

	ranks := func(getter func(ComponentScore) float64) []int {
		vals := make([]float64, len(ranked))
		for i, s := range ranked {
			vals[i] = getter(s)
		}
		order := make([]int, len(vals))
		for i := range order {
			order[i] = i
		}
		sort.Slice(order, func(i, j int) bool {
			return vals[order[i]] > vals[order[j]]
		})
		out := make([]int, len(vals))
		for rank, idx := range order {
			out[idx] = rank + 1
		}
		return out
	}

	pfRanks := ranks(func(s ComponentScore) float64 { return s.QuantilePF })
	wrRanks := ranks(func(s ComponentScore) float64 { return s.QuantileWRDelta })
	stabilityRanks := ranks(func(s ComponentScore) float64 { return s.RollingStability })
	regimeRanks := ranks(func(s ComponentScore) float64 { return s.RegimeConsistencyVal })
	corrRanks := ranks(func(s ComponentScore) float64 { return s.AvgAbsCorrelation })

	n := len(ranked)

	for i := range ranked {
		pfScore := 1.0 - float64(pfRanks[i]-1)/float64(n-1)
		wrScore := 1.0 - float64(wrRanks[i]-1)/float64(n-1)
		stabilityScore := 1.0 - float64(stabilityRanks[i]-1)/float64(n-1)
		regimeScore := 1.0 - float64(regimeRanks[i]-1)/float64(n-1)
		corrScore := 1.0 - float64(corrRanks[i]-1)/float64(n-1)

		ranked[i].CompositeScore = 0.40*pfScore + 0.25*wrScore + 0.15*stabilityScore + 0.10*regimeScore + 0.10*corrScore
	}

	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].CompositeScore > ranked[j].CompositeScore
	})

	return ranked
}
