package backtest

import (
	"time"

	"github.com/avav/ai_trading_v1/internal/agent/tradegate"
)

type Direction string

const (
	DirLong  Direction = "long"
	DirShort Direction = "short"
)

type Config struct {
	Symbol         string
	Timeframe      string
	StartTime      time.Time
	EndTime        time.Time
	InitialCapital float64
	Commission     float64
	Slippage       float64
	HoldingBars    int
	Direction      string
	ATRMultiplier  float64
	WarmupBars     int
	LongThreshold  float64
	ShortThreshold float64
	GateConfig     tradegate.GateConfig
	MLAPIURL       string
}

func DefaultConfig() Config {
	return Config{
		InitialCapital: 10_000,
		Commission:     0.001,
		Slippage:       0.0005,
		HoldingBars:    4,
		Direction:      "both",
		ATRMultiplier:  2.0,
		WarmupBars:     200,
		LongThreshold:  60.0,
		ShortThreshold: 40.0,
		GateConfig:     tradegate.DefaultConfig(),
		MLAPIURL:       "",
	}
}

type Trade struct {
	EntryTime   time.Time
	ExitTime    time.Time
	Direction   Direction
	EntryPrice  float64
	ExitPrice   float64
	Size        float64
	PnL         float64
	ReturnPct   float64
	HoldingBars int
	ExitReason  string

	TechnicalScore float64
	OrderFlowScore float64
	RegimeScore    float64
	Confidence     float64
	RegimeLabel    string
	StopPrice      float64
}

func decideDir(techScore float64, longThreshold, shortThreshold float64) Direction {
	if techScore >= longThreshold {
		return DirLong
	}
	if techScore <= shortThreshold {
		return DirShort
	}
	return ""
}

type Metrics struct {
	TotalTrades    int
	WinningTrades  int
	LosingTrades   int
	WinRate        float64
	ProfitFactor   float64
	NetPnL         float64
	TotalFeesPaid  float64
	SharpeRatio    float64
	SortinoRatio   float64
	MaxDrawdownPct float64
	Expectancy     float64
	AvgReturnPct   float64

	AvgWinnerPct     float64
	AvgLoserPct      float64
	LargestWinnerPct float64
	LargestLoserPct  float64
	AvgHoldingBars   float64
	ExposurePct      float64

	TotalBars      int
	TradingDays    int
	CalendarDays   int
	TradesPerMonth float64
	TradesPerYear  float64
}

type Summary struct {
	Config      Config
	Trades      []Trade
	Metrics     Metrics
	EquityCurve []float64
}

type pendingInfo struct {
	entryBar    int
	entryPrice  float64
	stopPrice   float64
	size        float64
	direction   Direction
	techScore   float64
	ofScore     float64
	regimeScore float64
	regimeLabel string
	confidence  float64
}
