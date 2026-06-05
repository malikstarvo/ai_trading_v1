package feature

import (
	"math"
	"testing"
)

func TestADX_ShortInput(t *testing.T) {
	v := ADX([]float64{1, 2, 3}, []float64{1, 2, 3}, []float64{1, 2, 3}, 14)
	for _, val := range v {
		if !math.IsNaN(val) {
			t.Fatal("expected NaN for short input")
		}
	}
}

func TestADX_WarmupPeriod(t *testing.T) {
	n := 30
	high := make([]float64, n)
	low := make([]float64, n)
	close := make([]float64, n)
	for i := 0; i < n; i++ {
		high[i] = float64(100 + i)
		low[i] = float64(99 + i)
		close[i] = 99.5 + float64(i)
	}
	v := ADX(high, low, close, 14)
	// First valid ADX should be at index 28 (2*14)
	for i := 0; i < 28; i++ {
		if !math.IsNaN(v[i]) {
			t.Fatalf("expected NaN at index %d for warmup phase", i)
		}
	}
	if math.IsNaN(v[28]) {
		t.Fatal("expected valid ADX at index 28")
	}
}

func TestADX_TrendingUp(t *testing.T) {
	n := 60
	high := make([]float64, n)
	low := make([]float64, n)
	close := make([]float64, n)
	for i := 0; i < n; i++ {
		high[i] = float64(100 + i)
		low[i] = float64(99 + i)
		close[i] = 99.5 + float64(i)
	}
	v := ADX(high, low, close, 14)
	last := v[len(v)-1]
	if math.IsNaN(last) {
		t.Fatal("expected valid ADX")
	}
	// Strong trend → ADX should be > 25
	if last < 25 {
		t.Fatalf("expected ADX > 25 for strong trend, got %v", last)
	}
}

func TestADX_Ranging(t *testing.T) {
	n := 60
	high := make([]float64, n)
	low := make([]float64, n)
	close := make([]float64, n)
	for i := 0; i < n; i++ {
		high[i] = 101
		low[i] = 99
		close[i] = 100
	}
	v := ADX(high, low, close, 14)
	last := v[len(v)-1]
	if math.IsNaN(last) {
		t.Fatal("expected valid ADX")
	}
	// Ranging market → ADX should be < 25
	if last > 25 {
		t.Fatalf("expected ADX < 25 for ranging market, got %v", last)
	}
}
