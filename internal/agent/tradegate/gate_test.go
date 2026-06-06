package tradegate

import (
	"math"
	"testing"
)

func almostEqual(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}

func TestDecisions(t *testing.T) {
	tests := []struct {
		name         string
		input        Input
		setup        func(*Gate)
		wantDecision Decision
		wantSizeMin  float64
		wantSizeMax  float64
		wantReason   string
	}{
		{
			name: "full size: high confidence ideal regime",
			input: Input{
				TechnicalScore: 85,
				OrderFlowScore: 82,
				RegimeScore:    80,
				RegimeLabel:    "trending_low_vol",
				MetaModelProb:  1.0,
			},
			wantDecision: FullSize,
			wantSizeMin:  1.0,
			wantSizeMax:  1.0,
		},
		{
			name: "normal size: moderate confidence",
			input: Input{
				TechnicalScore: 75,
				OrderFlowScore: 70,
				RegimeScore:    65,
				RegimeLabel:    "trending_low_vol",
				MetaModelProb:  1.0,
			},
			wantDecision: NormalSize,
			wantSizeMin:  0.5,
			wantSizeMax:  0.5,
		},
		{
			name: "small size: trending high vol override",
			input: Input{
				TechnicalScore: 65,
				OrderFlowScore: 60,
				RegimeScore:    55,
				RegimeLabel:    "trending_high_vol",
				MetaModelProb:  1.0,
			},
			wantDecision: SmallSize,
			wantSizeMin:  0.18,
			wantSizeMax:  0.19,
		},
		{
			name: "full base but ranging high vol cuts to 0.25",
			input: Input{
				TechnicalScore: 85,
				OrderFlowScore: 80,
				RegimeScore:    75,
				RegimeLabel:    "ranging_high_vol",
				MetaModelProb:  1.0,
			},
			wantDecision: FullSize,
			wantSizeMin:  0.25,
			wantSizeMax:  0.25,
		},
		{
			name: "regime blacklist: ranging_low_vol",
			input: Input{
				TechnicalScore: 90,
				OrderFlowScore: 85,
				RegimeScore:    40,
				RegimeLabel:    "ranging_low_vol",
				MetaModelProb:  1.0,
			},
			wantDecision: NoTrade,
			wantReason:   "regime: ranging_low_vol",
		},
		{
			name: "small size: borderline confidence",
			input: Input{
				TechnicalScore: 55,
				OrderFlowScore: 60,
				RegimeScore:    70,
				RegimeLabel:    "trending_low_vol",
				MetaModelProb:  1.0,
			},
			wantDecision: SmallSize,
			wantSizeMin:  0.25,
			wantSizeMax:  0.25,
		},
		{
			name: "no trade: low confidence",
			input: Input{
				TechnicalScore: 40,
				OrderFlowScore: 30,
				RegimeScore:    50,
				RegimeLabel:    "trending_low_vol",
				MetaModelProb:  1.0,
			},
			wantDecision: NoTrade,
			wantReason:   "low confidence: 38.0",
		},
		{
			name: "no trade: ML below threshold",
			input: Input{
				TechnicalScore: 85,
				OrderFlowScore: 80,
				RegimeScore:    75,
				RegimeLabel:    "trending_low_vol",
				MetaModelProb:  0.40,
			},
			wantDecision: NoTrade,
			wantReason:   "ML below threshold: 0.40",
		},
		{
			name: "no trade: NaN scores",
			input: Input{
				TechnicalScore: math.NaN(),
				OrderFlowScore: 80,
				RegimeScore:    75,
				RegimeLabel:    "trending_low_vol",
				MetaModelProb:  1.0,
			},
			wantDecision: NoTrade,
			wantReason:   "invalid input: NaN or Inf score",
		},
		{
			name: "no trade: cooldown active",
			input: Input{
				TechnicalScore: 85,
				OrderFlowScore: 82,
				RegimeScore:    80,
				RegimeLabel:    "trending_low_vol",
				MetaModelProb:  1.0,
			},
			setup: func(g *Gate) {
				g.OnTradeClosed()
			},
			wantDecision: NoTrade,
			wantReason:   "cooldown: 2 bars remaining",
		},
		{
			name: "no trade: regime unknown",
			input: Input{
				TechnicalScore: 85,
				OrderFlowScore: 80,
				RegimeScore:    75,
				RegimeLabel:    "unknown",
				MetaModelProb:  1.0,
			},
			wantDecision: NoTrade,
			wantReason:   "regime: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := New(DefaultConfig())
			if tt.setup != nil {
				tt.setup(g)
			}

			out := g.Evaluate(tt.input)

			if out.Decision != tt.wantDecision {
				t.Errorf("Decision = %s, want %s (reason=%s conf=%.1f size=%.3f)",
					out.Decision, tt.wantDecision, out.Reason, out.RawConfidence, out.SizeMultiplier)
			}

			if tt.wantSizeMin > 0 || tt.wantSizeMax > 0 {
				if out.SizeMultiplier < tt.wantSizeMin || out.SizeMultiplier > tt.wantSizeMax {
					t.Errorf("SizeMultiplier = %.4f, want [%.4f, %.4f]",
						out.SizeMultiplier, tt.wantSizeMin, tt.wantSizeMax)
				}
			}

			if tt.wantReason != "" && out.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", out.Reason, tt.wantReason)
			}

			if out.Decision == NoTrade && out.Reason == "" {
				t.Error("NoTrade decision must have a non-empty Reason")
			}
		})
	}
}

func TestRawConfidence(t *testing.T) {
	g := New(DefaultConfig())

	// Test raw confidence is computed even for rejected trades
	out := g.Evaluate(Input{
		TechnicalScore: 90,
		OrderFlowScore: 85,
		RegimeScore:    40,
		RegimeLabel:    "ranging_low_vol",
		MetaModelProb:  1.0,
	})
	if out.RawConfidence != 78 {
		t.Errorf("RawConfidence for rejected trade = %.1f, want 78.0", out.RawConfidence)
	}

	// Test raw confidence for accepted trade
	out = g.Evaluate(Input{
		TechnicalScore: 85,
		OrderFlowScore: 82,
		RegimeScore:    80,
		RegimeLabel:    "trending_low_vol",
		MetaModelProb:  1.0,
	})
	if !almostEqual(out.RawConfidence, 82.8, 0.1) {
		t.Errorf("RawConfidence = %.1f, want 82.8", out.RawConfidence)
	}
}

func TestCooldownStateMachine(t *testing.T) {
	g := New(GateConfig{
		CooldownBars:      2,
		MaxTradesPerDay:   5,
		DrawdownLimit:     -0.05,
		StartingCapital:   10_000,
		MetaModelThreshold: 0.45,
		TechWeight:        0.40,
		OFWeight:          0.40,
		RegimeWeight:      0.20,
	})

	input := Input{
		TechnicalScore: 85,
		OrderFlowScore: 82,
		RegimeScore:    80,
		RegimeLabel:    "trending_low_vol",
		MetaModelProb:  1.0,
	}

	// Initial: trade allowed
	out := g.Evaluate(input)
	if out.Decision != FullSize {
		t.Fatalf("initial evaluate should pass: got %s", out.Decision)
	}

	// Simulate trade lifecycle
	g.OnTradePlaced()
	g.OnTradeClosed() // → cooldown, 2 bars remaining

	// Cooldown bar 1
	g.OnBar()
	out = g.Evaluate(input)
	if out.Decision != NoTrade {
		t.Errorf("during cooldown: should be NoTrade, got %s", out.Decision)
	}
	if out.Reason != "cooldown: 1 bars remaining" {
		t.Errorf("expected cooldown: 1 bars, got %s", out.Reason)
	}

	// Cooldown bar 2 → expires
	g.OnBar()
	if g.State() != StateIdle {
		t.Errorf("after cooldown: state should be idle, got %s", g.State())
	}

	// Now trade should be allowed again
	out = g.Evaluate(input)
	if out.Decision != FullSize {
		t.Errorf("after cooldown: should allow trade, got %s", out.Decision)
	}
}

func TestMaxTradesPerDay(t *testing.T) {
	g := New(GateConfig{
		CooldownBars:      0,
		MaxTradesPerDay:   3,
		DrawdownLimit:     -0.05,
		StartingCapital:   10_000,
		MetaModelThreshold: 0.45,
		TechWeight:        0.40,
		OFWeight:          0.40,
		RegimeWeight:      0.20,
	})

	input := Input{
		TechnicalScore: 85,
		OrderFlowScore: 82,
		RegimeScore:    80,
		RegimeLabel:    "trending_low_vol",
		MetaModelProb:  1.0,
	}

	for i := 0; i < 3; i++ {
		out := g.Evaluate(input)
		if out.Decision != FullSize {
			t.Fatalf("trade %d: expected FullSize, got %s", i+1, out.Decision)
		}
		g.OnTradePlaced()
		// No cooldown (cooldownBars=0), but OnTradeClosed still sets cooldown
		g.OnTradeClosed()
		g.OnBar() // expire cooldown immediately
	}

	// 4th trade should be blocked
	out := g.Evaluate(input)
	if out.Decision != NoTrade {
		t.Errorf("after max trades: expected NoTrade, got %s", out.Decision)
	}
	if out.Reason != "max trades per day reached: 3" {
		t.Errorf("expected max trades reason, got %s", out.Reason)
	}

	// New day resets
	g.ResetDay("2026-06-07", 0)
	out = g.Evaluate(input)
	if out.Decision != FullSize {
		t.Errorf("after reset: should allow trade, got %s", out.Decision)
	}
}

func TestDailyDrawdownStop(t *testing.T) {
	g := New(GateConfig{
		CooldownBars:      0,
		MaxTradesPerDay:   5,
		DrawdownLimit:     -0.05,
		StartingCapital:   10_000,
		MetaModelThreshold: 0.45,
		TechWeight:        0.40,
		OFWeight:          0.40,
		RegimeWeight:      0.20,
	})

	input := Input{
		TechnicalScore: 85,
		OrderFlowScore: 82,
		RegimeScore:    80,
		RegimeLabel:    "trending_low_vol",
		MetaModelProb:  1.0,
	}

	// Accumulate losses: -$300, -$200, -$100, -$200 = -$800 (8% > 5%)
	g.UpdatePnl(-300)
	g.UpdatePnl(-200)
	g.UpdatePnl(-100)
	g.UpdatePnl(-200)

	if g.State() != StateStopped {
		t.Fatalf("after drawdown: state should be stopped, got %s", g.State())
	}

	out := g.Evaluate(input)
	if out.Decision != NoTrade {
		t.Errorf("during drawdown: expected NoTrade, got %s", out.Decision)
	}
	if out.Reason != "daily drawdown limit hit" {
		t.Errorf("expected drawdown reason, got %s", out.Reason)
	}

	// New day resets
	reset := g.ResetDay("2026-06-07", 0)
	if !reset {
		t.Error("ResetDay should return true when resuming from stopped")
	}
	if g.State() != StateIdle {
		t.Errorf("after reset: state should be idle, got %s", g.State())
	}

	out = g.Evaluate(input)
	if out.Decision != FullSize {
		t.Errorf("after day reset: should allow trade, got %s", out.Decision)
	}
}

func TestRegimeUnknownRejected(t *testing.T) {
	g := New(DefaultConfig())

	out := g.Evaluate(Input{
		TechnicalScore: 85,
		OrderFlowScore: 82,
		RegimeScore:    80,
		RegimeLabel:    "unknown",
		MetaModelProb:  1.0,
	})

	if out.Decision != NoTrade {
		t.Errorf("unknown regime: expected NoTrade, got %s", out.Decision)
	}
}

func TestBoundaries(t *testing.T) {
	g := New(DefaultConfig())

	// Confidence exactly 60 → SmallSize
	out := g.Evaluate(Input{
		TechnicalScore: 60,
		OrderFlowScore: 60,
		RegimeScore:    60,
		RegimeLabel:    "trending_low_vol",
		MetaModelProb:  1.0,
	})
	if out.Decision != SmallSize {
		t.Errorf("conf=60: expected SmallSize, got %s", out.Decision)
	}

	// Confidence exactly 70 → NormalSize
	out = g.Evaluate(Input{
		TechnicalScore: 70,
		OrderFlowScore: 70,
		RegimeScore:    70,
		RegimeLabel:    "trending_low_vol",
		MetaModelProb:  1.0,
	})
	if out.Decision != NormalSize {
		t.Errorf("conf=70: expected NormalSize, got %s", out.Decision)
	}

	// Confidence exactly 80 → FullSize
	out = g.Evaluate(Input{
		TechnicalScore: 80,
		OrderFlowScore: 80,
		RegimeScore:    80,
		RegimeLabel:    "trending_low_vol",
		MetaModelProb:  1.0,
	})
	if out.Decision != FullSize {
		t.Errorf("conf=80: expected FullSize, got %s", out.Decision)
	}

	// MetaModelProb exactly 0.45 → threshold, should pass
	out = g.Evaluate(Input{
		TechnicalScore: 85,
		OrderFlowScore: 82,
		RegimeScore:    80,
		RegimeLabel:    "trending_low_vol",
		MetaModelProb:  0.45,
	})
	if out.Decision != FullSize {
		t.Errorf("ML prob=0.45: should pass threshold, got %s", out.Decision)
	}
}
