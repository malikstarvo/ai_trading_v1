package tradegate

import (
	"fmt"
	"math"
	"sync"
)

type Decision string

const (
	NoTrade    Decision = "no_trade"
	SmallSize  Decision = "small_size"
	NormalSize Decision = "normal_size"
	FullSize   Decision = "full_size"
)

type GateState string

const (
	StateIdle     GateState = "idle"
	StateCooldown GateState = "cooldown"
	StateStopped  GateState = "stopped"
)

type Input struct {
	TechnicalScore     float64
	OrderFlowScore     float64
	RegimeScore        float64
	RegimeLabel        string
	MetaModelProb      float64
}

type Output struct {
	Decision        Decision
	SizeMultiplier  float64
	RawConfidence   float64
	FinalConfidence float64
	Reason          string
}

type GateConfig struct {
	CooldownBars      int
	MaxTradesPerDay   int
	DrawdownLimit     float64
	StartingCapital   float64
	MetaModelThreshold float64
	TechWeight        float64
	OFWeight          float64
	RegimeWeight      float64
}

func DefaultConfig() GateConfig {
	return GateConfig{
		CooldownBars:      2,
		MaxTradesPerDay:   5,
		DrawdownLimit:     -0.05,
		StartingCapital:   10_000,
		MetaModelThreshold: 0.45,
		TechWeight:        0.40,
		OFWeight:          0.40,
		RegimeWeight:      0.20,
	}
}

type Gate struct {
	mu      sync.Mutex
	cfg     GateConfig

	state                 GateState
	cooldownBarsRemaining int
	tradeCount            int
	dayPnl                float64
	currentDay            string
}

func New(cfg GateConfig) *Gate {
	return &Gate{
		cfg:   cfg,
		state: StateIdle,
	}
}

// Evaluate processes the input through the complete gate pipeline:
// NaN → MetaModel → Regime blacklist → Drawdown → MaxTrades → Cooldown → Confidence sizing → Regime override.
func (g *Gate) Evaluate(input Input) Output {
	if !valid(input.TechnicalScore) || !valid(input.OrderFlowScore) || !valid(input.RegimeScore) {
		return noTrade(0, "invalid input: NaN or Inf score")
	}

	rawConf := input.TechnicalScore*g.cfg.TechWeight +
		input.OrderFlowScore*g.cfg.OFWeight +
		input.RegimeScore*g.cfg.RegimeWeight

	if rawConf > 100 {
		rawConf = 100
	}
	if rawConf < 0 {
		rawConf = 0
	}

	if input.MetaModelProb < g.cfg.MetaModelThreshold {
		return noTrade(rawConf, fmt.Sprintf("ML below threshold: %.2f", input.MetaModelProb))
	}

	if input.RegimeLabel == "ranging_low_vol" || input.RegimeLabel == "unknown" {
		return noTrade(rawConf, "regime: "+input.RegimeLabel)
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.state == StateStopped {
		return noTrade(rawConf, "daily drawdown limit hit")
	}

	if g.tradeCount >= g.cfg.MaxTradesPerDay {
		return noTrade(rawConf, fmt.Sprintf("max trades per day reached: %d", g.cfg.MaxTradesPerDay))
	}

	if g.state == StateCooldown {
		return noTrade(rawConf, fmt.Sprintf("cooldown: %d bars remaining", g.cooldownBarsRemaining))
	}

	baseSize := sizeFromConfidence(rawConf)
	if baseSize == 0 {
		return noTrade(rawConf, fmt.Sprintf("low confidence: %.1f", rawConf))
	}

	finalSize := applyRegimeOverride(baseSize, input.RegimeLabel)

	decision := decisionFromSize(baseSize)
	return Output{
		Decision:        decision,
		SizeMultiplier:  finalSize,
		RawConfidence:   rawConf,
		FinalConfidence: rawConf,
		Reason:          "",
	}
}

// OnTradePlaced is called when a trade is taken. Increments trade count.
func (g *Gate) OnTradePlaced() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.tradeCount++
}

// OnTradeClosed transitions to cooldown state.
func (g *Gate) OnTradeClosed() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.state = StateCooldown
	g.cooldownBarsRemaining = g.cfg.CooldownBars
}

// OnBar decrements cooldown. If cooldown expires, returns to idle.
func (g *Gate) OnBar() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.state == StateCooldown {
		g.cooldownBarsRemaining--
		if g.cooldownBarsRemaining <= 0 {
			g.state = StateIdle
		}
	}
}

// UpdatePnl adds realized PnL from a closed trade. Checks drawdown limit.
func (g *Gate) UpdatePnl(pnl float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.dayPnl += pnl
	ddPct := g.dayPnl / g.cfg.StartingCapital
	if ddPct <= g.cfg.DrawdownLimit {
		g.state = StateStopped
	}
}

// ResetDay resets daily counters and checks if a new day started.
// Returns true if state was reset from stopped to idle.
func (g *Gate) ResetDay(newDay string, startingPnl float64) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.tradeCount = 0
	g.dayPnl = startingPnl
	g.currentDay = newDay
	if g.state == StateStopped {
		g.state = StateIdle
		return true
	}
	return false
}

// State returns current gate state (thread-safe).
func (g *Gate) State() GateState {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.state
}

// TradeCount returns trades today (thread-safe).
func (g *Gate) TradeCount() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.tradeCount
}

// --- internal helpers ---

func valid(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

func noTrade(rawConf float64, reason string) Output {
	return Output{
		Decision:        NoTrade,
		SizeMultiplier:  0,
		RawConfidence:   rawConf,
		FinalConfidence: rawConf,
		Reason:          reason,
	}
}

func sizeFromConfidence(conf float64) float64 {
	switch {
	case conf >= 80:
		return 1.00
	case conf >= 70:
		return 0.50
	case conf >= 60:
		return 0.25
	default:
		return 0
	}
}

func decisionFromSize(size float64) Decision {
	switch {
	case size >= 1.00:
		return FullSize
	case size >= 0.50:
		return NormalSize
	case size > 0:
		return SmallSize
	default:
		return NoTrade
	}
}

func applyRegimeOverride(size float64, label string) float64 {
	if size <= 0 {
		return 0
	}

	switch label {
	case "trending_low_vol":
		return size * 1.00
	case "trending_high_vol":
		return size * 0.75
	case "ranging_high_vol":
		return size * 0.25
	default:
		return 0
	}
}
