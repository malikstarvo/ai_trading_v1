package feature

import (
	"math"
	"testing"
)

func TestEMA_ShortInput(t *testing.T) {
	v := EMA([]float64{10, 20}, 5)
	if len(v) != 2 {
		t.Fatalf("expected 2 results, got %d", len(v))
	}
	if !math.IsNaN(v[0]) || !math.IsNaN(v[1]) {
		t.Fatal("expected NaN for short input")
	}
}

func TestEMA_ExactPeriod(t *testing.T) {
	v := EMA([]float64{1, 2, 3, 4, 5}, 5)
	if len(v) != 5 {
		t.Fatalf("expected 5 results, got %d", len(v))
	}
	for i := 0; i < 4; i++ {
		if !math.IsNaN(v[i]) {
			t.Fatalf("expected NaN at index %d", i)
		}
	}
	// SMA = 3, EMA = 2/6*5 + 4/6*3 = 1.6667 + 2.0 = 3.6667
	expected := 2.0/6.0*5.0 + 4.0/6.0*3.0
	if math.Abs(v[4]-expected) > 1e-10 {
		t.Fatalf("expected %v, got %v", expected, v[4])
	}
}

func TestEMA_Convergence(t *testing.T) {
	v := EMA([]float64{10, 10, 10, 10, 10, 10, 10, 10}, 3)
	// After SMA seed of 10, EMA should converge to 10
	last := v[len(v)-1]
	if math.Abs(last-10) > 1e-10 {
		t.Fatalf("expected 10, got %v", last)
	}
}

func TestEMA_Empty(t *testing.T) {
	v := EMA(nil, 5)
	if v != nil {
		t.Fatal("expected nil for empty input")
	}
	v = EMA([]float64{}, 5)
	if v != nil {
		t.Fatal("expected nil for empty input")
	}
}

func TestPriceAboveEMA(t *testing.T) {
	closes := []float64{5, 10, 15, 20, 25}
	ema := []float64{math.NaN(), math.NaN(), 12, 18, 22}
	result := PriceAboveEMA(closes, ema)
	expected := []int8{0, 0, 1, 1, 1}
	for i := range expected {
		if result[i] != expected[i] {
			t.Fatalf("index %d: expected %d, got %d", i, expected[i], result[i])
		}
	}
}

func TestPriceAboveEMA_Negative(t *testing.T) {
	closes := []float64{5, 10, 8}
	ema := []float64{math.NaN(), math.NaN(), 11}
	result := PriceAboveEMA(closes, ema)
	if result[2] != 0 {
		t.Fatal("expected 0 when close below ema")
	}
}
