package regime

import (
	"math"
	"testing"
)

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

func TestCalculate(t *testing.T) {
	tests := []struct {
		name       string
		input      Input
		wantScore  float64
		wantRegime string
	}{
		{
			name: "strong trend active vol",
			input: Input{
				ADX14:      40,
				ATR14:      1250,
				Price:      50000,
				Volatility: 1.5,
			},
			// trend=50, atrVol=38.33 (2.5%), featVol=31.67, vol=35, total=85
			wantScore:  85,
			wantRegime: "trending_high_vol",
		},
		{
			name: "strong trend low vol",
			input: Input{
				ADX14:      40,
				ATR14:      400,
				Price:      50000,
				Volatility: 0.6,
			},
			// trend=50, atrVol=11 (0.8%), featVol=13.57, vol=12.29, total=62
			wantScore:  62,
			wantRegime: "trending_low_vol",
		},
		{
			name: "moderate trend active vol",
			input: Input{
				ADX14:      30,
				ATR14:      1500,
				Price:      50000,
				Volatility: 2.0,
			},
			// trend=33.33, atrVol=45 (3.0%), featVol=38.33, vol=41.67, total=75
			wantScore:  75,
			wantRegime: "trending_high_vol",
		},
		{
			name: "weak trend low vol",
			input: Input{
				ADX14:      22,
				ATR14:      500,
				Price:      50000,
				Volatility: 0.7,
			},
			// trend=6.67, atrVol=15 (1.0%), featVol=16.43, vol=15.71, total=22
			wantScore:  22,
			wantRegime: "ranging_low_vol",
		},
		{
			name: "dead flat",
			input: Input{
				ADX14:      15,
				ATR14:      150,
				Price:      50000,
				Volatility: 0.2,
			},
			// trend=0, atrVol=5 (0.3%), featVol=5, vol=5, total=5
			wantScore:  5,
			wantRegime: "ranging_low_vol",
		},
		{
			name: "choppy high vol",
			input: Input{
				ADX14:      18,
				ATR14:      1750,
				Price:      50000,
				Volatility: 2.5,
			},
			// trend=0, atrVol=40 (3.5%), featVol=45, vol=42.5, total=42.5
			wantScore:  42,
			wantRegime: "ranging_high_vol",
		},
		{
			name: "strong trend extreme vol",
			input: Input{
				ADX14:      40,
				ATR14:      3000,
				Price:      50000,
				Volatility: 4.0,
			},
			// trend=50, atrVol=15 (6.0%), featVol=20, vol=17.5, total=67.5
			wantScore:  68,
			wantRegime: "trending_low_vol",
		},
		{
			name: "trend boundary low vol",
			input: Input{
				ADX14:      25,
				ATR14:      700,
				Price:      50000,
				Volatility: 0.9,
			},
			// trend=16.67, atrVol=23 (1.4%), featVol=22.14, vol=22.57, total=39
			wantScore:  39,
			wantRegime: "trending_low_vol",
		},
		{
			name: "strong trend moderate high vol",
			input: Input{
				ADX14:      35,
				ATR14:      1000,
				Price:      50000,
				Volatility: 1.2,
			},
			// trend=50, atrVol=31.67 (2.0%), featVol=27.67, vol=29.67, total=80
			wantScore:  80,
			wantRegime: "trending_high_vol",
		},
		{
			name: "NaN guard all invalid",
			input: Input{
				ADX14:      math.NaN(),
				ATR14:      math.NaN(),
				Price:      math.NaN(),
				Volatility: math.NaN(),
			},
			wantScore:  50,
			wantRegime: "unknown",
		},
		{
			name: "partial NaN ADX only",
			input: Input{
				ADX14:      math.NaN(),
				ATR14:      1000,
				Price:      50000,
				Volatility: 1.5,
			},
			// trend=25 (NaN neutral), atrVol=31.67 (2.0%), featVol=31.67, vol=31.67, total=56.67
			wantScore:  57,
			wantRegime: "unknown",
		},
	}

	tolerance := 1.0

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := Calculate(tt.input)
			if !almostEqual(score.RegimeScore, tt.wantScore, tolerance) {
				t.Errorf("RegimeScore = %.2f, want %.2f (trend=%.1f vol=%.1f label=%s)",
					score.RegimeScore, tt.wantScore,
					score.Components.TrendScore,
					score.Components.VolScore,
					score.Regime)
			}
			if tt.wantRegime != "" && score.Regime != tt.wantRegime {
				t.Errorf("Regime = %s, want %s (score=%.2f trend=%.1f vol=%.1f)",
					score.Regime, tt.wantRegime,
					score.RegimeScore,
					score.Components.TrendScore,
					score.Components.VolScore)
			}
		})
	}
}

func TestComponentsSum(t *testing.T) {
	tests := []struct {
		name  string
		input Input
	}{
		{"strong trend high vol", Input{40, 1250, 50000, 1.5}},
		{"dead flat", Input{15, 150, 50000, 0.2}},
		{"choppy high vol", Input{18, 1750, 50000, 2.5}},
		{"NaN all", Input{math.NaN(), math.NaN(), math.NaN(), math.NaN()}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := Calculate(tt.input)
			sum := score.Components.TrendScore + score.Components.VolScore
			if !almostEqual(sum, score.RegimeScore, 1.5) {
				t.Errorf("components sum = %.2f != RegimeScore = %.2f (trend=%.1f vol=%.1f label=%s)",
					sum, score.RegimeScore,
					score.Components.TrendScore,
					score.Components.VolScore,
					score.Regime)
			}
		})
	}
}

func TestBounds(t *testing.T) {
	// Max trend + max vol
	maxInput := Input{
		ADX14:      100,
		ATR14:      1750,
		Price:      50000,
		Volatility: 2.0,
	}
	score := Calculate(maxInput)
	if score.RegimeScore > 100 {
		t.Errorf("score exceeds 100: %.2f", score.RegimeScore)
	}
	if score.Components.TrendScore > 50 {
		t.Errorf("trend exceeds 50: %.2f", score.Components.TrendScore)
	}

	// Negative values (invalid but should not crash)
	negInput := Input{
		ADX14:      -5,
		ATR14:      -100,
		Price:      50000,
		Volatility: -0.1,
	}
	score = Calculate(negInput)
	if score.RegimeScore < 0 {
		t.Errorf("negative score: %.2f", score.RegimeScore)
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
