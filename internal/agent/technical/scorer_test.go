package technical

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
			name: "bullish perfect",
			input: Input{
				Price:    50000,
				EMA20:    49500,
				EMA50:    48500,
				EMA200:   47000,
				RSI14:    55,
				ATR14:    1250,
				Volume:   1500000,
				VolEMA20: 1000000,
				ADX14:    30,
			},
			wantScore: 89,
		},
		{
			name: "bearish kuat",
			input: Input{
				Price:    45000,
				EMA20:    47000,
				EMA50:    48000,
				EMA200:   49000,
				RSI14:    25,
				ATR14:    1800,
				Volume:   500000,
				VolEMA20: 1000000,
				ADX14:    15,
			},
			wantScore: 25,
		},
		{
			name: "pullback in uptrend",
			input: Input{
				Price:    48000,
				EMA20:    49000,
				EMA50:    48500,
				EMA200:   47000,
				RSI14:    42,
				ATR14:    900,
				Volume:   1000000,
				VolEMA20: 1000000,
				ADX14:    22,
			},
			wantScore: 62,
		},
		{
			name: "overbought",
			input: Input{
				Price:    52000,
				EMA20:    50000,
				EMA50:    49000,
				EMA200:   48000,
				RSI14:    82,
				ATR14:    1560,
				Volume:   2000000,
				VolEMA20: 1000000,
				ADX14:    35,
			},
			wantScore: 79,
		},
		{
			name: "zero volume",
			input: Input{
				Price:    50000,
				EMA20:    49500,
				EMA50:    49000,
				EMA200:   48500,
				RSI14:    50,
				ATR14:    1000,
				Volume:   0,
				VolEMA20: 1000000,
				ADX14:    28,
			},
			// trend=35 (base=30+align=5), momentum=24 (RSI=50), vol=0, atr=10 (2.0%<3.5), adx=7
			wantScore: 76,
		},
		{
			name: "flat market",
			input: Input{
				Price:    50000,
				EMA20:    50000,
				EMA50:    50000,
				EMA200:   50000,
				RSI14:    50,
				ATR14:    250,
				Volume:   1000000,
				VolEMA20: 1000000,
				ADX14:    18,
			},
			wantScore: 35,
		},
		{
			name: "RSI extreme oversold",
			input: Input{
				Price:    50000,
				EMA20:    48000,
				EMA50:    46000,
				EMA200:   44000,
				RSI14:    0,
				ATR14:    1100,
				Volume:   1200000,
				VolEMA20: 1000000,
				ADX14:    20,
			},
			wantScore: 62,
		},
		{
			name: "all zero inputs",
			input: Input{
				Price:    0,
				EMA20:    0,
				EMA50:    0,
				EMA200:   0,
				RSI14:    0,
				ATR14:    0,
				Volume:   0,
				VolEMA20: 0,
				ADX14:    0,
			},
			// Price=0 is valid (not NaN/Inf), RSI=0 → 5 from oversold zone, others return 0
			wantScore: 5,
		},
		{
			name: "high volatility",
			input: Input{
				Price:    50000,
				EMA20:    48500,
				EMA50:    47500,
				EMA200:   46500,
				RSI14:    60,
				ATR14:    3000,
				Volume:   2500000,
				VolEMA20: 1000000,
				ADX14:    32,
			},
			wantScore: 89,
		},
		{
			name: "bearish oversold",
			input: Input{
				Price:    45000,
				EMA20:    46000,
				EMA50:    47000,
				EMA200:   48000,
				RSI14:    28,
				ATR14:    1350,
				Volume:   1800000,
				VolEMA20: 1000000,
				ADX14:    25,
			},
			wantScore: 44,
		},
		{
			name: "NaN Price",
			input: Input{
				Price:    math.NaN(),
				EMA20:    49500,
				EMA50:    48500,
				EMA200:   47000,
				RSI14:    55,
				ATR14:    1250,
				Volume:   1500000,
				VolEMA20: 1000000,
				ADX14:    30,
			},
			wantScore: 0,
		},
		{
			name: "Inf Price",
			input: Input{
				Price:    math.Inf(1),
				EMA20:    49500,
				EMA50:    48500,
				EMA200:   47000,
				RSI14:    55,
				ATR14:    1250,
				Volume:   1500000,
				VolEMA20: 1000000,
				ADX14:    30,
			},
			wantScore: 0,
		},
	}

	tolerance := 1.0

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := Calculate(tt.input)
			if !almostEqual(score.TechnicalScore, tt.wantScore, tolerance) {
				t.Errorf("TechnicalScore = %.2f, want %.2f (components: trend=%.1f mom=%.1f vol=%.1f atr=%.1f adx=%.1f)",
					score.TechnicalScore, tt.wantScore,
					score.Components.Trend,
					score.Components.Momentum,
					score.Components.Volume,
					score.Components.Volatility,
					score.Components.ADXBonus)
			}
		})
	}
}

func TestComponentsSum(t *testing.T) {
	tests := []struct {
		name  string
		input Input
	}{
		{"bullish perfect", Input{50000, 49500, 48500, 47000, 55, 1250, 1500000, 1000000, 30}},
		{"bearish kuat", Input{45000, 47000, 48000, 49000, 25, 1800, 500000, 1000000, 15}},
		{"flat market", Input{50000, 50000, 50000, 50000, 50, 250, 1000000, 1000000, 18}},
		{"high volatility", Input{50000, 48500, 47500, 46500, 60, 3000, 2500000, 1000000, 32}},
		{"NaN guard", Input{}},
	}

	tolerance := 1.5

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := Calculate(tt.input)
			sum := score.Components.Trend + score.Components.Momentum +
				score.Components.Volume + score.Components.Volatility + score.Components.ADXBonus
			if !almostEqual(sum, score.TechnicalScore, tolerance) {
				t.Errorf("components sum = %.2f != TechnicalScore = %.2f (trend=%.1f mom=%.1f vol=%.1f atr=%.1f adx=%.1f)",
					sum, score.TechnicalScore,
					score.Components.Trend,
					score.Components.Momentum,
					score.Components.Volume,
					score.Components.Volatility,
					score.Components.ADXBonus)
			}
		})
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
	if valid(math.Inf(-1)) != false {
		t.Error("valid(-Inf) should be false")
	}
}
