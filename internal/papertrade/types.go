package papertrade

import (
	"time"

	"github.com/avav/ai_trading_v1/internal/agent/tradegate"
)

type Direction string

const (
	Long    Direction = "long"
	Short   Direction = "short"
	NoTrade Direction = "no_trade"
)

type EngineState string

const (
	StateRunning EngineState = "running"
	StateStopped EngineState = "stopped"
)

type OrderStatus string

const (
	OrderCreated  OrderStatus = "created"
	OrderFilled   OrderStatus = "filled"
	OrderCancelled OrderStatus = "cancelled"
)

type ExitReason string

const (
	ExitStopLoss     ExitReason = "stop_loss"
	ExitMaxHold      ExitReason = "max_hold"
	ExitOppositeSig  ExitReason = "opposite_signal"
	ExitEndOfData    ExitReason = "end_of_data"
	ExitStopLossDaily ExitReason = "stop_loss_daily"
)

type EngineConfig struct {
	Symbol               string
	Timeframe            string
	InitialCapital       float64
	Commission           float64
	Slippage             float64
	ATRMultiplier        float64
	HoldingBars          int
	PollInterval         time.Duration
	RiskPerTradePct      float64
	MaxDailyDrawdownPct  float64
	MaxTotalDrawdownPct  float64
	LongThreshold        float64
	ShortThreshold       float64
	GateConfig           tradegate.GateConfig
}

func DefaultConfig() EngineConfig {
	return EngineConfig{
		Symbol:              "BTCUSDT",
		Timeframe:           "15m",
		InitialCapital:      10_000,
		Commission:          0.00055,
		Slippage:            0.0005,
		ATRMultiplier:       2.0,
		HoldingBars:         24,
		PollInterval:        60 * time.Second,
		RiskPerTradePct:     1.0,
		MaxDailyDrawdownPct: 5.0,
		MaxTotalDrawdownPct: 15.0,
		LongThreshold:       60.0,
		ShortThreshold:      40.0,
		GateConfig:          tradegate.DefaultConfig(),
	}
}

type Order struct {
	ID            int64       `db:"id"`
	Symbol        string      `db:"symbol"`
	Timeframe     string      `db:"timeframe"`
	Direction     Direction   `db:"direction"`
	Status        OrderStatus `db:"status"`
	RequestedSize float64     `db:"requested_size"`
	FilledSize    float64     `db:"filled_size"`
	FillPrice     float64     `db:"fill_price"`
	SlippagePct   float64     `db:"slippage_pct"`
	Commission    float64     `db:"commission"`
	Reason        string      `db:"reason"`
	OpenTS        time.Time   `db:"open_ts"`
	CreatedAt     time.Time   `db:"created_at"`
}

type Fill struct {
	ID      int64     `db:"id"`
	OrderID int64     `db:"order_id"`
	TS      time.Time `db:"ts"`
	Side    string    `db:"side"`
	Price   float64   `db:"price"`
	Size    float64   `db:"size"`
	Fee     float64   `db:"fee"`
}

type Position struct {
	ID           int64       `db:"id"`
	Symbol       string      `db:"symbol"`
	Timeframe    string      `db:"timeframe"`
	Direction    Direction   `db:"direction"`
	EntryOrderID int64       `db:"entry_order_id"`
	Quantity     float64     `db:"quantity"`
	EntryPrice   float64     `db:"entry_price"`
	EntryFee     float64     `db:"entry_fee"`
	StopPrice    float64     `db:"stop_price"`
	OpenTS       time.Time   `db:"open_ts"`
	BarsHeld     int         `db:"bars_held"`
	Status       string      `db:"status"`
}

type Trade struct {
	ID              int64       `db:"id"`
	PositionID      int64       `db:"position_id"`
	Symbol          string      `db:"symbol"`
	Timeframe       string      `db:"timeframe"`
	Direction       Direction   `db:"direction"`
	EntryTS         time.Time   `db:"entry_ts"`
	ExitTS          time.Time   `db:"exit_ts"`
	EntryPrice      float64     `db:"entry_price"`
	ExitPrice       float64     `db:"exit_price"`
	Size            float64     `db:"size"`
	GrossPnL        float64     `db:"gross_pnl"`
	Commission      float64     `db:"commission"`
	NetPnL          float64     `db:"net_pnl"`
	ReturnPct       float64     `db:"return_pct"`
	HoldingBars     int         `db:"holding_bars"`
	ExitReason      string      `db:"exit_reason"`
	EntryReason     string      `db:"entry_reason"`
	FeatureSnapshot string      `db:"feature_snapshot"`
}

type AccountSnapshot struct {
	TS            time.Time `db:"ts"`
	Balance       float64   `db:"balance"`
	Equity        float64   `db:"equity"`
	UnrealizedPnL float64   `db:"unrealized_pnl"`
	DayPnL        float64   `db:"day_pnl"`
	DayTrades     int       `db:"day_trades"`
}

type AgentSnapshot struct {
	TechnicalScore float64 `json:"technical_score,omitempty"`
	OrderFlowScore float64 `json:"orderflow_score,omitempty"`
	RegimeScore    float64 `json:"regime_score,omitempty"`
	ConfidenceScore float64 `json:"confidence_score,omitempty"`
	RegimeLabel    string  `json:"regime_label,omitempty"`
	ATR14          float64 `json:"atr14,omitempty"`
	ADX14          float64 `json:"adx14,omitempty"`
}

type TickResult struct {
	CandleTS    time.Time
	TechScore   float64
	OFScore     float64
	RegimeScore float64
	RegimeLabel string
	Confidence  float64
	Direction   Direction
	GateOutput  tradegate.Output
	TradeOpened bool
	TradeClosed bool
	ExitReason  string
	Equity      float64
	DayPnL      float64
	EngineState EngineState
}
