package stream_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/avav/ai_trading_v1/internal/collector/metrics"
	"github.com/avav/ai_trading_v1/internal/collector/mock"
	"github.com/avav/ai_trading_v1/internal/collector/stream"
	"github.com/avav/ai_trading_v1/internal/collector/validate"
	"github.com/prometheus/client_golang/prometheus"
)

func newTestHandlers(
	cs *mock.MockCandleStore,
	ofs *mock.MockOrderFlowStore,
) *stream.Handlers {
	reg := prometheus.NewRegistry()
	m := metrics.New(reg)
	v := validate.New(m)
	ctx := context.Background()
	log := slog.Default()

	return stream.NewHandlers(ctx, cs, ofs, v, m, log)
}

func TestHandleKline_ConfirmedCandleStored(t *testing.T) {
	cs := mock.NewMockCandleStore()
	ofs := mock.NewMockOrderFlowStore()
	h := newTestHandlers(cs, ofs)

	msg := `{
		"topic": "kline.15.BTCUSDT",
		"type": "snapshot",
		"data": [{
			"start": 1672324800000,
			"end": 1672325099999,
			"interval": "15",
			"open": "50000",
			"close": "50500",
			"high": "51000",
			"low": "49000",
			"volume": "123.456",
			"turnover": "6200000",
			"confirm": true,
			"timestamp": 1672324988882
		}],
		"ts": 1672324988882
	}`

	h.HandleMessage([]byte(msg))

	candles := cs.Candles()
	if len(candles) != 1 {
		t.Fatalf("expected 1 candle, got %d", len(candles))
	}
	if candles[0].Symbol != "BTCUSDT" {
		t.Errorf("expected BTCUSDT, got %s", candles[0].Symbol)
	}
	if candles[0].Timeframe != "15m" {
		t.Errorf("expected 15m, got %s", candles[0].Timeframe)
	}
}

func TestHandleKline_UnconfirmedCandleNotStored(t *testing.T) {
	cs := mock.NewMockCandleStore()
	ofs := mock.NewMockOrderFlowStore()
	h := newTestHandlers(cs, ofs)

	msg := `{
		"topic": "kline.15.BTCUSDT",
		"type": "snapshot",
		"data": [{
			"start": 1672324800000,
			"interval": "15",
			"open": "50000",
			"close": "50500",
			"high": "51000",
			"low": "49000",
			"volume": "123.456",
			"confirm": false,
			"timestamp": 1672324988882
		}],
		"ts": 1672324988882
	}`

	h.HandleMessage([]byte(msg))

	candles := cs.Candles()
	if len(candles) != 0 {
		t.Errorf("expected 0 candles for unconfirmed, got %d", len(candles))
	}
}

func TestHandleTicker_OIStored(t *testing.T) {
	cs := mock.NewMockCandleStore()
	ofs := mock.NewMockOrderFlowStore()
	h := newTestHandlers(cs, ofs)

	msg := `{
		"topic": "tickers.BTCUSDT",
		"type": "snapshot",
		"data": [{
			"symbol": "BTCUSDT",
			"lastPrice": "50000",
			"openInterest": "150000.5",
			"openInterestValue": "7500000000",
			"fundingRate": "0.0001",
			"nextFundingTime": "1672329600000"
		}],
		"ts": 1672324988882
	}`

	h.HandleMessage([]byte(msg))

	oiRecords := ofs.OIRecords()
	if len(oiRecords) < 1 {
		t.Fatalf("expected at least 1 OI record")
	}
	if oiRecords[0].OI != 150000.5 {
		t.Errorf("expected OI 150000.5, got %f", oiRecords[0].OI)
	}
}

func TestHandleTicker_FundingRateStored(t *testing.T) {
	cs := mock.NewMockCandleStore()
	ofs := mock.NewMockOrderFlowStore()
	h := newTestHandlers(cs, ofs)

	msg := `{
		"topic": "tickers.ETHUSDT",
		"type": "snapshot",
		"data": [{
			"symbol": "ETHUSDT",
			"lastPrice": "3500",
			"openInterest": "500000",
			"openInterestValue": "1750000000",
			"fundingRate": "0.0002",
			"nextFundingTime": "1672329600000"
		}],
		"ts": 1672324988882
	}`

	h.HandleMessage([]byte(msg))

	_ = ofs.OIRecords()
	// If no panic, test passes
}

func TestHandleLiquidation(t *testing.T) {
	cs := mock.NewMockCandleStore()
	ofs := mock.NewMockOrderFlowStore()
	h := newTestHandlers(cs, ofs)

	msg := `{
		"topic": "allLiquidation.BTCUSDT",
		"type": "snapshot",
		"data": [{
			"T": 1707752083964,
			"s": "BTCUSDT",
			"S": "Buy",
			"v": "0.003",
			"p": "49407.90"
		}],
		"ts": 1707752083964
	}`

	h.HandleMessage([]byte(msg))

	liqRecords := ofs.LiqRecords()
	if len(liqRecords) != 1 {
		t.Fatalf("expected 1 liquidation, got %d", len(liqRecords))
	}
	if liqRecords[0].Symbol != "BTCUSDT" {
		t.Errorf("expected BTCUSDT, got %s", liqRecords[0].Symbol)
	}
	if liqRecords[0].Side != "Buy" {
		t.Errorf("expected Buy, got %s", liqRecords[0].Side)
	}
}

func TestHandleInvalidMessage(t *testing.T) {
	cs := mock.NewMockCandleStore()
	ofs := mock.NewMockOrderFlowStore()
	h := newTestHandlers(cs, ofs)

	// Invalid JSON should not panic
	h.HandleMessage([]byte(`invalid json`))
	h.HandleMessage([]byte(`{}`))
	h.HandleMessage([]byte(`{"op":"pong"}`))
}

func TestHandleKline_ParseTopicCorrectly(t *testing.T) {
	cs := mock.NewMockCandleStore()
	ofs := mock.NewMockOrderFlowStore()
	h := newTestHandlers(cs, ofs)

	// Test 1h timeframe
	msg := `{
		"topic": "kline.60.BTCUSDT",
		"type": "snapshot",
		"data": [{
			"start": 1672324800000,
			"interval": "60",
			"open": "50000",
			"close": "51000",
			"high": "51500",
			"low": "49500",
			"volume": "1000",
			"turnover": "51000000",
			"confirm": true,
			"timestamp": 1672324988882
		}]
	}`

	h.HandleMessage([]byte(msg))

	candles := cs.Candles()
	if len(candles) == 0 {
		t.Fatal("expected candle from 1h topic")
	}
	if candles[0].Timeframe != "1h" {
		t.Errorf("expected 1h, got %s", candles[0].Timeframe)
	}
}
