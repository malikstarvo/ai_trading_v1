package orderflow

import (
	"math"
	"testing"
)

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

func TestCalculate(t *testing.T) {
	tests := []struct {
		name      string
		input     Input
		wantScore float64
	}{
		{
			name: "bullish confluence",
			input: Input{
				FundingRate: 0.00002,
				OIDeltaPct:  0.5,
				LSRatio:     1.2,
				LongLiqUSD:  1_000_000,
				ShortLiqUSD: 500_000,
			},
			// fund=11 (0.2bps sweet spot), oi=20, ls=16, liq=12.71
			wantScore: 60,
		},
		{
			name: "bearish confluence",
			input: Input{
				FundingRate: -0.00003,
				OIDeltaPct:  -0.8,
				LSRatio:     0.7,
				LongLiqUSD:  200_000,
				ShortLiqUSD: 1_000_000,
			},
			// fund=9.1, oi=7, ls=6, liq=12.48
			wantScore: 35,
		},
		{
			name: "short squeeze top",
			input: Input{
				FundingRate: 0.00008,
				OIDeltaPct:  1.0,
				LSRatio:     1.8,
				LongLiqUSD:  500_000,
				ShortLiqUSD: 10_000_000,
			},
			// fund=14 (0.8bps sweet spot), oi=25, ls=12.29, liq=11.93
			wantScore: 63,
		},
		{
			name: "long capitulation",
			input: Input{
				FundingRate: 0.00001,
				OIDeltaPct:  0.2,
				LSRatio:     1.05,
				LongLiqUSD:  5_000_000,
				ShortLiqUSD: 500_000,
			},
			// fund=10.5 (0.1bps sweet spot), oi=17, ls=13, liq=13.72
			wantScore: 54,
		},
		{
			name: "extreme bearish",
			input: Input{
				FundingRate: -0.00010,
				OIDeltaPct:  -1.5,
				LSRatio:     0.4,
				LongLiqUSD:  10_000_000,
				ShortLiqUSD: 300_000,
			},
			// fund=7 (-1bps mild bear), oi=0, ls=0, liq=14.98
			wantScore: 22,
		},
		{
			name: "blowoff top",
			input: Input{
				FundingRate: 0.00015,
				OIDeltaPct:  2.0,
				LSRatio:     2.5,
				LongLiqUSD:  20_000_000,
				ShortLiqUSD: 1_000_000,
			},
			// fund=17.5 (1.5bps sweet spot), oi=25, ls=6, liq=17.43
			wantScore: 66,
		},
		{
			name: "neutral no activity",
			input: Input{
				FundingRate: 0,
				OIDeltaPct:  0,
				LSRatio:     1.0,
				LongLiqUSD:  0,
				ShortLiqUSD: 0,
			},
			// fund=10, oi=15, ls=12, liq=12.5
			wantScore: 50,
		},
		{
			name: "mixed signals",
			input: Input{
				FundingRate: 0.00005,
				OIDeltaPct:  -0.3,
				LSRatio:     1.4,
				LongLiqUSD:  100_000,
				ShortLiqUSD: 100_000,
			},
			// fund=12.5 (0.5bps sweet spot), oi=12, ls=16.86, liq=12.52
			wantScore: 54,
		},
		{
			name: "long wipeout",
			input: Input{
				FundingRate: 0.00002,
				OIDeltaPct:  0.3,
				LSRatio:     1.1,
				LongLiqUSD:  50_000_000,
				ShortLiqUSD: 500_000,
			},
			// fund=11 (0.2bps sweet spot), oi=18, ls=14, liq=24.84
			wantScore: 68,
		},
		{
			name: "NaN guard missing data",
			input: Input{
				FundingRate: math.NaN(),
				OIDeltaPct:  math.NaN(),
				LSRatio:     math.NaN(),
				LongLiqUSD:  math.NaN(),
				ShortLiqUSD: math.NaN(),
			},
			// all fields fall back to neutral: fund=10, oi=12.5, ls=10, liq=12.5
			wantScore: 45,
		},
	}

	tolerance := 1.0

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := Calculate(tt.input)
			if !almostEqual(score.OrderFlowScore, tt.wantScore, tolerance) {
				t.Errorf("OrderFlowScore = %.2f, want %.2f (fund=%.1f oi=%.1f ls=%.1f liq=%.1f)",
					score.OrderFlowScore, tt.wantScore,
					score.Components.FundingScore,
					score.Components.OIScore,
					score.Components.LSScore,
					score.Components.LiqScore)
			}
		})
	}
}

func TestComponentsSum(t *testing.T) {
	tests := []struct {
		name  string
		input Input
	}{
		{"bullish", Input{0.00002, 0.5, 1.2, 1_000_000, 500_000}},
		{"bearish", Input{-0.00003, -0.8, 0.7, 200_000, 1_000_000}},
		{"neutral", Input{0, 0, 1.0, 0, 0}},
		{"NaN", Input{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}},
	}

	tolerance := 1.5

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := Calculate(tt.input)
			sum := score.Components.FundingScore + score.Components.OIScore +
				score.Components.LSScore + score.Components.LiqScore
			if !almostEqual(sum, score.OrderFlowScore, tolerance) {
				t.Errorf("components sum = %.2f != OrderFlowScore = %.2f (fund=%.1f oi=%.1f ls=%.1f liq=%.1f)",
					sum, score.OrderFlowScore,
					score.Components.FundingScore,
					score.Components.OIScore,
					score.Components.LSScore,
					score.Components.LiqScore)
			}
		})
	}
}

func TestBounds(t *testing.T) {
	// Extreme bull case: all max
	extremeBull := Input{
		FundingRate: 0.0001,   // 1 bps, sweet spot → 15
		OIDeltaPct:  5.0,      // cap at 25
		LSRatio:     1.15,     // 12 + 0.15/0.3*6 = 15
		LongLiqUSD:  500_000_000, // well above $50M cap
		ShortLiqUSD: 0,
	}
	score := Calculate(extremeBull)
	if score.OrderFlowScore < 50 {
		t.Errorf("extreme bull should score high: got %.2f (fund=%.1f oi=%.1f ls=%.1f liq=%.1f)",
			score.OrderFlowScore,
			score.Components.FundingScore,
			score.Components.OIScore,
			score.Components.LSScore,
			score.Components.LiqScore)
	}
	if score.OrderFlowScore > 100 {
		t.Errorf("score exceeds 100: %.2f", score.OrderFlowScore)
	}

	// Extreme bear case: all min
	extremeBear := Input{
		FundingRate: -0.001,  // -10 bps → 0
		OIDeltaPct:  -10.0,   // cap at 0
		LSRatio:     0.01,    // cap at 0
		LongLiqUSD:  0,
		ShortLiqUSD: 500_000_000,
	}
	score = Calculate(extremeBear)
	if score.OrderFlowScore > 20 {
		t.Errorf("extreme bear should score low: got %.2f (fund=%.1f oi=%.1f ls=%.1f liq=%.1f)",
			score.OrderFlowScore,
			score.Components.FundingScore,
			score.Components.OIScore,
			score.Components.LSScore,
			score.Components.LiqScore)
	}
}

func TestValid(t *testing.T) {
	if valid(42.0) != true {
		t.Error("valid(42.0) should be true")
	}
	if valid(0.0) != true {
		t.Error("valid(0.0) should be true")
	}
	if valid(math.NaN()) != false {
		t.Error("valid(NaN) should be false")
	}
	if valid(math.Inf(1)) != false {
		t.Error("valid(+Inf) should be false")
	}
}
