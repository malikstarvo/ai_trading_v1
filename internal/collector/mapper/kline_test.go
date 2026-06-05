package mapper_test

import (
	"testing"
	"time"

	"github.com/avav/ai_trading_v1/internal/collector/mapper"
)

func TestWSKlineToCandle(t *testing.T) {
	item := mapper.WSKlineData{
		Start:     1672324800000,
		End:       1672325099999,
		Interval:  "15",
		Open:      "50000.00",
		High:      "51000.00",
		Low:       "49000.00",
		Close:     "50500.00",
		Volume:    "123.456",
		Turnover:  "6200000",
		Confirm:   true,
		Timestamp: 1672324988882,
	}

	candle := mapper.WSKlineToCandle(item, "BTCUSDT", "15m")

	if candle.Symbol != "BTCUSDT" {
		t.Errorf("expected BTCUSDT, got %s", candle.Symbol)
	}
	if candle.Timeframe != "15m" {
		t.Errorf("expected 15m, got %s", candle.Timeframe)
	}
	if candle.Open != 50000 {
		t.Errorf("expected 50000, got %f", candle.Open)
	}
	if candle.Close != 50500 {
		t.Errorf("expected 50500, got %f", candle.Close)
	}
	if candle.Volume != 123.456 {
		t.Errorf("expected 123.456, got %f", candle.Volume)
	}
	expectedTime := time.UnixMilli(1672324800000)
	if !candle.Time.Equal(expectedTime) {
		t.Errorf("expected %v, got %v", expectedTime, candle.Time)
	}
}

func TestWSKlineToCandle_EmptyStrings(t *testing.T) {
	item := mapper.WSKlineData{
		Start:   1672324800000,
		Open:    "",
		High:    "",
		Low:     "",
		Close:   "",
		Volume:  "",
		Confirm: true,
	}

	candle := mapper.WSKlineToCandle(item, "ETHUSDT", "1h")

	if candle.Open != 0 {
		t.Errorf("expected 0, got %f", candle.Open)
	}
}

func TestWSKlineToCandle_Unconfirmed(t *testing.T) {
	item := mapper.WSKlineData{
		Start:   1672324800000,
		Open:    "50000",
		High:    "51000",
		Low:     "49000",
		Close:   "50500",
		Volume:  "100",
		Confirm: false,
	}

	_ = mapper.WSKlineToCandle(item, "BTCUSDT", "15m")
	// Unconfirmed candles should be skipped by handler, but mapper still produces valid output
}
