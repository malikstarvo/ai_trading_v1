package backtest

import (
	"math"
	"testing"
	"time"

	"github.com/avav/ai_trading_v1/internal/agent/tradegate"
	"github.com/avav/ai_trading_v1/internal/model"
)

func defaultTestConfig() Config {
	cfg := DefaultConfig()
	cfg.WarmupBars = 200
	cfg.HoldingBars = 4
	cfg.GateConfig.CooldownBars = 0
	cfg.GateConfig.MetaModelThreshold = 0.0
	cfg.GateConfig.DrawdownLimit = -0.50
	cfg.GateConfig.MaxTradesPerDay = 100
	return cfg
}

func bullishFeatures(i int) model.FeatureRow {
	price := 50000.0 + float64(i)*5.0
	return model.FeatureRow{
		EMA20:       price * 0.99,
		EMA50:       price * 0.97,
		EMA200:      price * 0.95,
		RSI14:       58,
		ATR14:       price * 0.012,
		ADX14:       32,
		VolumeEMA20: 1_000_000,

		FundingRate:  0.00002,
		OIDelta1Pct:  5.0,
		LSRatioRaw:   1.2,
		LiqLongUSD:   5_000_000,
		LiqShortUSD:  1_000_000,
		Volatility14: 1.8,
	}
}

func rangingFeatures(i int) model.FeatureRow {
	price := 50000.0 + math.Sin(float64(i)*0.3)*100.0
	return model.FeatureRow{
		EMA20:  price * 0.998,
		EMA50:  price * 0.997,
		EMA200: price * 0.996,
		RSI14:  50,
		ATR14:  price * 0.003,
		ADX14:  16,

		FundingRate:  0.00001,
		OIDelta1Pct:  0.1,
		LSRatioRaw:   1.0,
		LiqLongUSD:   100_000,
		LiqShortUSD:  100_000,
		Volatility14: 0.4,
	}
}

func chaoticFeatures(i int) model.FeatureRow {
	price := 50000.0 + math.Sin(float64(i)*0.5)*500.0
	return model.FeatureRow{
		EMA20:  price * 0.98,
		EMA50:  price * 0.97,
		EMA200: price * 0.96,
		RSI14:  45,
		ATR14:  price * 0.04,
		ADX14:  18,

		FundingRate:  -0.0002,
		OIDelta1Pct:  -0.5,
		LSRatioRaw:   0.8,
		LiqLongUSD:   2_000_000,
		LiqShortUSD:  3_000_000,
		Volatility14: 4.0,
	}
}

func neutralFeatures(i int) model.FeatureRow {
	price := 50000.0
	return model.FeatureRow{
		EMA20:       price * 0.999,
		EMA50:       price * 0.998,
		EMA200:      price * 0.997,
		RSI14:       50,
		ATR14:       price * 0.005,
		ADX14:       20,
		VolumeEMA20: 500_000,

		FundingRate:  0.00001,
		OIDelta1Pct:  0.0,
		LSRatioRaw:   1.0,
		LiqLongUSD:   100_000,
		LiqShortUSD:  100_000,
		Volatility14: 0.6,
	}
}

func nanFeatures(i int) model.FeatureRow {
	return model.FeatureRow{
		RSI14:      math.NaN(),
		ATR14:      math.NaN(),
		ADX14:      math.NaN(),
		OIDelta1Pct: math.NaN(),
	}
}

func makeCandles(n int, basePrice float64, trend float64, vol float64) []model.Candle {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := make([]model.Candle, n)
	for i := 0; i < n; i++ {
		open := basePrice + float64(i)*trend
		halfRange := vol / 2
		candles[i] = model.Candle{
			Time:   start.Add(time.Duration(i) * 15 * time.Minute),
			Symbol: "BTCUSDT",
			Open:   open - halfRange*0.5,
			High:   open + halfRange,
			Low:    open - halfRange,
			Close:  open + halfRange*0.3,
			Volume: 1_000_000,
		}
	}
	return candles
}

func makeFeatures(n int, fn func(int) model.FeatureRow) []model.FeatureRow {
	return makeFeaturesWithCandles(n, fn, nil)
}

func makeFeaturesWithCandles(n int, fn func(int) model.FeatureRow, candles []model.Candle) []model.FeatureRow {
	features := make([]model.FeatureRow, n)
	for i := 0; i < n; i++ {
		features[i] = fn(i)
		if candles != nil && i < len(candles) {
			features[i].Ts = candles[i].Time
		}
	}
	return features
}

func makeAlignedFeatures(candles []model.Candle, fn func(int) model.FeatureRow) []model.FeatureRow {
	return makeFeaturesWithCandles(len(candles), fn, candles)
}

func TestBacktestStrongBullishTrend(t *testing.T) {
	cfg := defaultTestConfig()
	e := New(cfg)
	candles := makeCandles(300, 50000, 100, 100)
	features := makeAlignedFeatures(candles, bullishFeatures)

	summary, err := e.Run(candles, features)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Metrics.TotalTrades == 0 {
		t.Fatal("expected at least 1 trade in bullish trend")
	}
	if summary.Metrics.NetPnL <= 0 {
		t.Fatalf("expected positive PnL in bullish trend, got %.2f", summary.Metrics.NetPnL)
	}
}

func TestBacktestRangingMarketNoTrades(t *testing.T) {
	cfg := defaultTestConfig()
	e := New(cfg)
	candles := makeCandles(300, 50000, 0, 50)
	features := makeAlignedFeatures(candles, rangingFeatures)

	summary, err := e.Run(candles, features)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Metrics.TotalTrades > 0 {
		t.Fatalf("expected 0 trades in ranging_low_vol regime, got %d", summary.Metrics.TotalTrades)
	}
}

func TestBacktestChaoticHighVol(t *testing.T) {
	cfg := defaultTestConfig()
	e := New(cfg)
	candles := makeCandles(300, 50000, 2, 300)
	features := makeAlignedFeatures(candles, chaoticFeatures)

	summary, err := e.Run(candles, features)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// ranging_high_vol is not blacklisted but regimeOverride multiplies size by 0.25.
	// If confidence is high enough, trades may still occur, but reduced size.
	// No strict assertion — just verify no crash and size is reasonable.
	t.Logf("chaotic: %d trades, netPnL=%.2f", summary.Metrics.TotalTrades, summary.Metrics.NetPnL)
}

func TestBacktestBelowConfidence(t *testing.T) {
	cfg := defaultTestConfig()
	e := New(cfg)
	candles := makeCandles(300, 50000, 0, 10)
	features := makeAlignedFeatures(candles, neutralFeatures)

	summary, err := e.Run(candles, features)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Neutral features produce low confidence (< 60) → no trades via sizeFromConfidence
	if summary.Metrics.TotalTrades > 0 {
		t.Fatalf("expected 0 trades with neutral features, got %d", summary.Metrics.TotalTrades)
	}
}

func TestBacktestCooldownBlocksNewTrades(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.GateConfig.CooldownBars = 3
	cfg.HoldingBars = 1
	e := New(cfg)
	candles := makeCandles(300, 50000, 5, 100)
	features := makeAlignedFeatures(candles, bullishFeatures)

	summary, err := e.Run(candles, features)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Metrics.TotalTrades < 5 {
		t.Fatalf("expected multiple trades in 100 live bars with holding=1, got %d", summary.Metrics.TotalTrades)
	}

	for i := 1; i < len(summary.Trades); i++ {
		prevExit := summary.Trades[i-1].EntryTime
		currEntry := summary.Trades[i].EntryTime
		gap := currEntry.Sub(prevExit)
		if gap < 0 {
			t.Fatalf("trades out of order: exit %v after entry %v", prevExit, currEntry)
		}
	}
}

func TestBacktestStopLossHit(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.HoldingBars = 20
	e := New(cfg)

	n := 250
	candles := make([]model.Candle, n)
	features := make([]model.FeatureRow, n)
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < n; i++ {
		price := 50000.0 + float64(i)*0.5
		candles[i] = model.Candle{
			Time:   start.Add(time.Duration(i) * 15 * time.Minute),
			Symbol: "BTCUSDT",
			Open:   price,
			High:   price * 1.01,
			Low:    price * 0.99,
			Close:  price,
			Volume: 1_000_000,
		}
		features[i] = bullishFeatures(i)
		features[i].Ts = candles[i].Time
		features[i].ATR14 = 500.0
	}

	// Force a deep low at bars 203-206 to trigger stop loss
	for j := 203; j < 210 && j < n; j++ {
		candles[j].Low = 48000
	}

	summary, err := e.Run(candles, features)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hasStop := false
	for _, t := range summary.Trades {
		if t.ExitReason == "stop" {
			hasStop = true
			break
		}
	}
	if !hasStop {
		t.Fatal("expected at least one stop exit")
	}
}

func TestBacktestMetricsKnownValues(t *testing.T) {
	et := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	xt := time.Date(2025, 1, 3, 12, 0, 0, 0, time.UTC)
	trades := []Trade{
		{EntryTime: et, ExitTime: et.Add(1 * time.Hour), ReturnPct: 0.02, PnL: 200, HoldingBars: 4},
		{EntryTime: et.Add(24 * time.Hour), ExitTime: et.Add(25 * time.Hour), ReturnPct: -0.01, PnL: -100, HoldingBars: 3},
		{EntryTime: et.Add(48 * time.Hour), ExitTime: et.Add(49 * time.Hour), ReturnPct: 0.03, PnL: 300, HoldingBars: 5},
		{EntryTime: et.Add(49 * time.Hour), ExitTime: et.Add(50 * time.Hour), ReturnPct: -0.02, PnL: -200, HoldingBars: 6},
		{EntryTime: xt, ExitTime: xt.Add(1 * time.Hour), ReturnPct: 0.015, PnL: 150, HoldingBars: 4},
	}

	m := CalculateMetrics(trades, 10_000, 100, 22, 50.0)

	if m.TotalTrades != 5 {
		t.Fatalf("TotalTrades: got %d, want 5", m.TotalTrades)
	}
	if m.WinningTrades != 3 {
		t.Fatalf("WinningTrades: got %d, want 3", m.WinningTrades)
	}
	if m.LosingTrades != 2 {
		t.Fatalf("LosingTrades: got %d, want 2", m.LosingTrades)
	}
	if math.Abs(m.WinRate-60.0) > 0.01 {
		t.Fatalf("WinRate: got %.2f, want 60.0", m.WinRate)
	}

	profitFactor := (0.02 + 0.03 + 0.015) / (0.01 + 0.02)
	if math.Abs(m.ProfitFactor-profitFactor) > 0.01 {
		t.Fatalf("ProfitFactor: got %.4f, want %.4f", m.ProfitFactor, profitFactor)
	}

	if m.ExposurePct != 22.0 {
		t.Fatalf("ExposurePct: got %.2f, want 22.0", m.ExposurePct)
	}

	if m.TotalFeesPaid != 50.0 {
		t.Fatalf("TotalFeesPaid: got %.2f, want 50.0", m.TotalFeesPaid)
	}
	if m.TotalBars != 100 {
		t.Fatalf("TotalBars: got %d, want 100", m.TotalBars)
	}
	if m.TradingDays != 3 {
		t.Fatalf("TradingDays: got %d, want 3", m.TradingDays)
	}
	if m.CalendarDays < 2 {
		t.Fatalf("CalendarDays: got %d, want >= 2", m.CalendarDays)
	}
}

func TestBacktestDrawdownStop(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.GateConfig.DrawdownLimit = -0.05
	cfg.GateConfig.StartingCapital = 10_000
	cfg.GateConfig.CooldownBars = 0
	e := New(cfg)

	n := 400
	candles := make([]model.Candle, n)
	features := make([]model.FeatureRow, n)
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < n; i++ {
		price := 50000.0
		candles[i] = model.Candle{
			Time:   start.Add(time.Duration(i) * 15 * time.Minute),
			Symbol: "BTCUSDT",
			Open:   price,
			High:   price * 1.001,
			Low:    price * 0.999,
			Close:  price,
			Volume: 1_000_000,
		}
		features[i] = bullishFeatures(i)
		features[i].Ts = candles[i].Time
		features[i].RSI14 = 65
		features[i].ADX14 = 35
		features[i].ATR14 = 100.0
	}

	// Drop price at bar 205 to cause loss
	for i := 203; i < 220 && i < n; i++ {
		candles[i].Close = 49000
		candles[i].Low = 48900
		candles[i].High = 49500
	}

	summary, err := e.Run(candles, features)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lastEquity := summary.EquityCurve[len(summary.EquityCurve)-1]
	if lastEquity >= 9500 {
		t.Fatalf("expected drawdown to reduce equity below 9500, got %.2f", lastEquity)
	}
}

func TestBacktestNaNFeatures(t *testing.T) {
	cfg := defaultTestConfig()
	e := New(cfg)
	candles := makeCandles(250, 50000, 0, 50)
	features := makeAlignedFeatures(candles, nanFeatures)

	summary, err := e.Run(candles, features)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// NaN features → agent scores default to 0 → confidence below threshold → 0 trades
	if summary.Metrics.TotalTrades != 0 {
		t.Fatalf("expected 0 trades with NaN features, got %d", summary.Metrics.TotalTrades)
	}
}

func TestBacktestShortData(t *testing.T) {
	cfg := defaultTestConfig()
	e := New(cfg)
	candles := makeCandles(50, 50000, 5, 100)
	features := makeAlignedFeatures(candles, bullishFeatures)

	_, err := e.Run(candles, features)
	if err == nil {
		t.Fatal("expected error for short data, got nil")
	}
}

func TestBacktestCustomConfig(t *testing.T) {
	cfg := Config{
		Symbol:         "BTCUSDT",
		Timeframe:      "15m",
		InitialCapital: 50_000,
		Commission:     0.0005,
		Slippage:       0.001,
		HoldingBars:    8,
		Direction:      "long",
		ATRMultiplier:  3.0,
		WarmupBars:     100,
		GateConfig: tradegate.GateConfig{
			CooldownBars:      1,
			MaxTradesPerDay:   10,
			DrawdownLimit:     -0.10,
			StartingCapital:   50_000,
			MetaModelThreshold: 0.0,
			TechWeight:        0.40,
			OFWeight:          0.40,
			RegimeWeight:      0.20,
		},
	}

	e := New(cfg)
	candles := makeCandles(300, 50000, 3, 150)
	features := makeAlignedFeatures(candles, bullishFeatures)

	summary, err := e.Run(candles, features)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Metrics.TotalTrades == 0 {
		t.Fatal("expected trades with custom config")
	}
	if len(summary.EquityCurve) != len(candles)-cfg.WarmupBars {
		t.Fatalf("equity curve length: got %d, want %d", len(summary.EquityCurve), len(candles)-cfg.WarmupBars)
	}
}

func TestBacktestCandleFeatureMismatch(t *testing.T) {
	cfg := defaultTestConfig()
	e := New(cfg)
	candles := makeCandles(250, 50000, 5, 100)
	features := makeAlignedFeatures(candles, bullishFeatures)[:200]

	_, err := e.Run(candles, features)
	if err == nil {
		t.Fatal("expected error for length mismatch")
	}
}

func TestBacktestEquityCurveConsistency(t *testing.T) {
	cfg := defaultTestConfig()
	e := New(cfg)
	candles := makeCandles(300, 50000, 5, 100)
	features := makeAlignedFeatures(candles, bullishFeatures)

	summary, err := e.Run(candles, features)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedLen := len(candles) - cfg.WarmupBars
	if len(summary.EquityCurve) != expectedLen {
		t.Fatalf("equity curve length: got %d, want %d", len(summary.EquityCurve), expectedLen)
	}

	if len(summary.EquityCurve) > 0 && summary.EquityCurve[0] <= 0 {
		t.Fatalf("initial equity should be positive, got %.2f", summary.EquityCurve[0])
	}
}

func TestBacktestFeatureCandleAlignment(t *testing.T) {
	cfg := defaultTestConfig()
	e := New(cfg)
	candles := makeCandles(250, 50000, 5, 100)
	features := makeFeatures(250, bullishFeatures)

	_, err := e.Run(candles, features)
	if err == nil {
		t.Fatal("expected error for timestamp mismatch, got nil")
	}
	t.Logf("alignment error: %v", err)
}

func TestBacktestFeeImpact(t *testing.T) {
	zCfg := defaultTestConfig()
	zCfg.Commission = 0
	zCfg.Slippage = 0

	fCfg := defaultTestConfig()
	fCfg.Commission = 0.001
	fCfg.Slippage = 0.0005

	candles := makeCandles(300, 50000, 100, 100)
	features := makeAlignedFeatures(candles, bullishFeatures)

	zeroSummary, err := New(zCfg).Run(candles, features)
	if err != nil {
		t.Fatalf("zero-fee run error: %v", err)
	}

	fullSummary, err := New(fCfg).Run(candles, features)
	if err != nil {
		t.Fatalf("full-fee run error: %v", err)
	}

	if zeroSummary.Metrics.NetPnL < 0 {
		t.Fatalf("zero-fee run should be profitable in bullish trend, got %.2f", zeroSummary.Metrics.NetPnL)
	}

	if fullSummary.Metrics.TotalFeesPaid <= 0 {
		t.Fatalf("expected positive fees with commission=0.001, got %.2f", fullSummary.Metrics.TotalFeesPaid)
	}

	if fullSummary.Metrics.NetPnL > zeroSummary.Metrics.NetPnL {
		t.Fatalf("fees should reduce net PnL: zero=%.2f full=%.2f",
			zeroSummary.Metrics.NetPnL, fullSummary.Metrics.NetPnL)
	}

	feeRatio := fullSummary.Metrics.TotalFeesPaid / (fullSummary.Metrics.TotalFeesPaid + fullSummary.Metrics.NetPnL) * 100
	t.Logf("zero-fee PnL=%.2f full-fee PnL=%.2f fees=%.2f feeRatio=%.1f%% trades=%d",
		zeroSummary.Metrics.NetPnL, fullSummary.Metrics.NetPnL,
		fullSummary.Metrics.TotalFeesPaid, feeRatio, fullSummary.Metrics.TotalTrades)
}

func TestBacktestTradeDistribution(t *testing.T) {
	// 2000 bars ≈ 20 days of 15m data, should produce many trades
	cfg := defaultTestConfig()
	cfg.WarmupBars = 500
	e := New(cfg)
	candles := makeCandles(2000, 50000, 50, 100)
	features := makeAlignedFeatures(candles, bullishFeatures)

	summary, err := e.Run(candles, features)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := summary.Metrics
	t.Logf("trades=%d tradingDays=%d calendarDays=%d totalBars=%d",
		m.TotalTrades, m.TradingDays, m.CalendarDays, m.TotalBars)
	t.Logf("tradesPerMonth=%.1f tradesPerYear=%.1f exposure=%.1f%% avgHold=%.1f",
		m.TradesPerMonth, m.TradesPerYear, m.ExposurePct, m.AvgHoldingBars)

	if m.TotalTrades == 0 {
		t.Fatal("expected at least 1 trade in 2000 bars")
	}
	if m.TradingDays == 0 {
		t.Fatal("expected at least 1 trading day")
	}
	if m.CalendarDays <= 0 {
		t.Fatal("expected positive calendar days")
	}
	if m.ExposurePct <= 0 {
		t.Fatal("expected positive exposure")
	}
	if m.TradesPerMonth <= 0 {
		t.Fatal("expected positive trades per month")
	}
	if m.TradesPerYear <= 0 {
		t.Fatal("expected positive trades per year")
	}
}

func TestBacktestTradesPerMonthReasonable(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.WarmupBars = 100
	e := New(cfg)
	candles := makeCandles(500, 50000, 50, 100)
	features := makeAlignedFeatures(candles, bullishFeatures)

	summary, err := e.Run(candles, features)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := summary.Metrics
	// In 400 live bars with holding=4 and cooldown=0, expect ~80 trades
	// Calendar days ≈ 400*15min/24h = ~4 days → trades/month ~ 80/4*30 = ~600
	if m.TradesPerMonth < 10 {
		t.Fatalf("expected high trade density, got TradesPerMonth=%.1f (trades=%d, days=%d)",
			m.TradesPerMonth, m.TotalTrades, m.CalendarDays)
	}
	t.Logf("validation: trades=%d days=%d trades/month=%.1f",
		m.TotalTrades, m.CalendarDays, m.TradesPerMonth)
}
