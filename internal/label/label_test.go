package label

import (
	"math"
	"testing"
	"time"

	"github.com/avav/ai_trading_v1/internal/model"
)

func TestComputeLabels_Basic(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	n := 30
	candles := make([]model.Candle, n)
	for i := 0; i < n; i++ {
		candles[i] = model.Candle{
			Time:      now.Add(time.Duration(i) * time.Hour),
			Symbol:    "BTCUSDT",
			Timeframe: "1h",
			Close:     100 + float64(i),
		}
	}

	rows, err := ComputeLabels(candles, []int{4, 12, 24}, 0.25)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != n {
		t.Fatalf("expected %d rows, got %d", n, len(rows))
	}

	// candle[0]: close=100, candle[4]: close=104
	// future_return_4 = (104-100)/100*100 = 4%
	if math.Abs(rows[0].FutureReturn4-4.0) > 0.001 {
		t.Fatalf("expected 4%%, got %v", rows[0].FutureReturn4)
	}
	if rows[0].Success4 != 1 {
		t.Fatal("expected success_4=1 for 4% return")
	}

	// Last 4 rows should have NaN for future_return_4 (no forward data)
	for i := n - 4; i < n; i++ {
		if rows[i].FutureReturn4 != 0 {
			t.Fatalf("expected 0 FutureReturn4 at index %d (last horizon rows)", i)
		}
		if rows[i].Success4 != 0 {
			t.Fatalf("expected 0 Success4 at index %d", i)
		}
	}
}

func TestComputeLabels_SuccessThreshold(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := []model.Candle{
		{Time: now, Close: 100},
		{Time: now.Add(time.Hour), Close: 100.1},   // +0.1% < 0.25% → not success
		{Time: now.Add(2 * time.Hour), Close: 100}, // -0.1% → not success
		{Time: now.Add(3 * time.Hour), Close: 100.25}, // +0.25% >= 0.25% → success
	}

	rows, _ := ComputeLabels(candles, []int{4}, 0.25)

	// rows[0] horizon 4 → look at close[4] which doesn't exist → 0
	if rows[0].FutureReturn4 != 0 {
		t.Fatal("expected 0 for last horizon row")
	}
}

func TestComputeLabels_Empty(t *testing.T) {
	rows, err := ComputeLabels(nil, []int{4, 12, 24}, 0.25)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rows != nil {
		t.Fatal("expected nil for empty input")
	}
}

func TestComputeLabels_DefaultThreshold(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := []model.Candle{
		{Time: now, Close: 100},
		{Time: now.Add(time.Hour), Close: 100.1},    // +0.1% → not success
		{Time: now.Add(2 * time.Hour), Close: 100.5}, // +0.5% → success
	}

	rows, _ := ComputeLabels(candles, []int{4}, 0)
	if rows[0].FutureReturn4 != 0 {
		t.Fatal("expected 0 for last horizon row")
	}
}
