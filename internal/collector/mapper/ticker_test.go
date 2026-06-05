package mapper_test

import (
	"testing"
	"time"

	"github.com/avav/ai_trading_v1/internal/collector/mapper"
)

func TestTickerToOI(t *testing.T) {
	data := mapper.WSTickerData{
		Symbol:            "BTCUSDT",
		OpenInterest:      "150000.5",
		OpenInterestValue: "7500000000",
		FundingRate:       "0.0001",
	}

	ts := time.Now()
	oi := mapper.TickerToOI(data, ts)

	if oi.Symbol != "BTCUSDT" {
		t.Errorf("expected BTCUSDT, got %s", oi.Symbol)
	}
	if oi.OI != 150000.5 {
		t.Errorf("expected 150000.5, got %f", oi.OI)
	}
	if oi.OIValueUSD != 7500000000 {
		t.Errorf("expected 7500000000, got %f", oi.OIValueUSD)
	}
}

func TestTickerToFundingRate(t *testing.T) {
	data := mapper.WSTickerData{
		Symbol:      "ETHUSDT",
		FundingRate: "0.0001",
	}

	ts := time.Now()
	fr := mapper.TickerToFundingRate(data, ts)

	if fr.Symbol != "ETHUSDT" {
		t.Errorf("expected ETHUSDT, got %s", fr.Symbol)
	}
	if fr.Rate != 0.0001 {
		t.Errorf("expected 0.0001, got %f", fr.Rate)
	}
}

func TestTickerToFundingRate_Zero(t *testing.T) {
	data := mapper.WSTickerData{
		Symbol:      "SOLUSDT",
		FundingRate: "0",
	}

	ts := time.Now()
	fr := mapper.TickerToFundingRate(data, ts)

	if fr.Rate != 0 {
		t.Errorf("expected 0, got %f", fr.Rate)
	}
}
