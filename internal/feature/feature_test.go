package feature

import (
	"math"
	"testing"
	"time"

	"github.com/avav/ai_trading_v1/internal/model"
)

func TestComputeFeatures_ShortInput(t *testing.T) {
	candles := make([]model.Candle, 50)
	for i := range candles {
		candles[i] = model.Candle{Time: time.Now(), Close: 100, High: 101, Low: 99, Volume: 1000}
	}
	result, err := ComputeFeatures(candles, nil, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil for <200 candles")
	}
}

func TestComputeFeatures_Basic(t *testing.T) {
	n := 210
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	candles := make([]model.Candle, n)
	for i := 0; i < n; i++ {
		candles[i] = model.Candle{
			Time:   now.Add(time.Duration(i) * time.Minute * 15),
			Symbol: "BTCUSDT",
			Timeframe: "15m",
			Open:   100 + float64(i)*0.1,
			High:   101 + float64(i)*0.1,
			Low:    99 + float64(i)*0.1,
			Close:  100 + float64(i)*0.1,
			Volume: 1000 + float64(i),
		}
	}

	of := make([]model.OrderFlowSnapshot, n)
	for i := 0; i < n; i++ {
		of[i] = model.OrderFlowSnapshot{
			Ts:          candles[i].Time,
			OI:          1e9 + float64(i)*1e6,
			FundingRate: 0.0001 + float64(i)*0.000001,
			LSBuyRatio:  0.5 + float64(i)*0.001,
			LSSellRatio: 0.5 - float64(i)*0.001,
			LiqLongUSD:  1e6 + float64(i)*1e4,
			LiqShortUSD: 1e6 - float64(i)*1e4,
		}
	}

	result, err := ComputeFeatures(candles, of, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result) != n {
		t.Fatalf("expected %d rows, got %d", n, len(result))
	}

	last := result[n-1]

	if math.IsNaN(last.EMA20) {
		t.Fatal("expected valid EMA20")
	}
	if math.IsNaN(last.EMA50) {
		t.Fatal("expected valid EMA50")
	}
	if math.IsNaN(last.EMA200) {
		t.Fatal("expected valid EMA200")
	}
	if math.IsNaN(last.RSI14) {
		t.Fatal("expected valid RSI14")
	}
	if math.IsNaN(last.ATR14) {
		t.Fatal("expected valid ATR14")
	}
	if math.IsNaN(last.ADX14) {
		t.Fatal("expected valid ADX14")
	}
	if math.IsNaN(last.VolumeEMA20) {
		t.Fatal("expected valid VolumeEMA20")
	}
	if math.IsNaN(last.Return1) {
		t.Fatal("expected valid Return1")
	}
	if math.IsNaN(last.Volatility14) {
		t.Fatal("expected valid Volatility14")
	}

	if math.IsNaN(last.OIDelta1Pct) {
		t.Fatal("expected valid OIDelta1Pct")
	}
	if math.IsNaN(last.OIZScore30) {
		t.Fatal("expected valid OIZScore30")
	}
	if math.IsNaN(last.FundingRate) {
		t.Fatal("expected valid FundingRate")
	}
	if math.IsNaN(last.FundingZScore30) {
		t.Fatal("expected valid FundingZScore30")
	}
	if math.IsNaN(last.LSRatioRaw) {
		t.Fatal("expected valid LSRatioRaw")
	}
	if math.IsNaN(last.LiqImbalance) {
		t.Fatal("expected valid LiqImbalance")
	}
}

func TestComputeFeatures_NoOrderFlow(t *testing.T) {
	n := 210
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	candles := make([]model.Candle, n)
	for i := 0; i < n; i++ {
		candles[i] = model.Candle{
			Time:   now.Add(time.Duration(i) * time.Minute * 15),
			Symbol: "BTCUSDT",
			Timeframe: "15m",
			Close:  100 + float64(i)*0.1,
		}
	}

	result, err := ComputeFeatures(candles, nil, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	last := result[n-1]
	if math.IsNaN(last.EMA20) {
		t.Fatal("expected valid EMA20 even without order flow")
	}
	if last.OIDelta1Pct != 0 {
		t.Fatal("expected 0 OI delta when no order flow data")
	}
}
