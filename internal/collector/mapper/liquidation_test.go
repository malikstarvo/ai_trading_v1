package mapper_test

import (
	"testing"
	"time"

	"github.com/avav/ai_trading_v1/internal/collector/mapper"
)

func TestWSAllLiquidationToLiq(t *testing.T) {
	item := mapper.WSLiquidationData{
		Time:  1707752083964,
		Sym:   "BTCUSDT",
		Side:  "Buy",
		Size:  "0.003",
		Price: "49407.90",
	}

	liq := mapper.WSAllLiquidationToLiq(item)

	if liq.Symbol != "BTCUSDT" {
		t.Errorf("expected BTCUSDT, got %s", liq.Symbol)
	}
	if liq.Side != "Buy" {
		t.Errorf("expected Buy, got %s", liq.Side)
	}
	if liq.Size != 0.003 {
		t.Errorf("expected 0.003, got %f", liq.Size)
	}
	if liq.Price != 49407.90 {
		t.Errorf("expected 49407.90, got %f", liq.Price)
	}
	expectedValue := 0.003 * 49407.90
	if liq.ValueUSD != expectedValue {
		t.Errorf("expected %f, got %f", expectedValue, liq.ValueUSD)
	}
	expectedTime := time.UnixMilli(1707752083964)
	if !liq.Time.Equal(expectedTime) {
		t.Errorf("expected %v, got %v", expectedTime, liq.Time)
	}
}

func TestWSAllLiquidation_Sell(t *testing.T) {
	item := mapper.WSLiquidationData{
		Time:  1707752083964,
		Sym:   "ETHUSDT",
		Side:  "Sell",
		Size:  "10.5",
		Price: "3500.00",
	}

	liq := mapper.WSAllLiquidationToLiq(item)

	if liq.Side != "Sell" {
		t.Errorf("expected Sell, got %s", liq.Side)
	}
	if liq.Size != 10.5 {
		t.Errorf("expected 10.5, got %f", liq.Size)
	}
}
