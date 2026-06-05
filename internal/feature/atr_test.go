package feature

import (
	"math"
	"testing"
)

func TestATR_ShortInput(t *testing.T) {
	v := ATR([]float64{10, 11}, []float64{9, 10}, []float64{9.5, 10.5}, 14)
	if len(v) != 2 {
		t.Fatalf("expected 2, got %d", len(v))
	}
	for _, val := range v {
		if !math.IsNaN(val) {
			t.Fatal("expected NaN for short input")
		}
	}
}

func TestATR_Basic(t *testing.T) {
	high := []float64{10, 11, 12, 11, 10, 9, 8, 9, 10, 11, 12, 11, 10, 9, 8}
	low := []float64{8, 9, 10, 9, 8, 7, 6, 7, 8, 9, 10, 9, 8, 7, 6}
	close := []float64{9, 10, 11, 10, 9, 8, 7, 8, 9, 10, 11, 9.5, 9, 8, 7}
	v := ATR(high, low, close, 14)
	if len(v) != 15 {
		t.Fatalf("expected 15, got %d", len(v))
	}
	for i := 0; i < 14; i++ {
		if !math.IsNaN(v[i]) {
			t.Fatalf("expected NaN at index %d", i)
		}
	}
	if math.IsNaN(v[14]) {
		t.Fatal("expected valid ATR at index 14")
	}
	if v[14] <= 0 {
		t.Fatal("ATR should be positive")
	}
}

func TestATR_Constant(t *testing.T) {
	n := 20
	high := make([]float64, n)
	low := make([]float64, n)
	close := make([]float64, n)
	for i := 0; i < n; i++ {
		high[i] = 11
		low[i] = 9
		close[i] = 10
	}
	v := ATR(high, low, close, 14)
	last := v[len(v)-1]
	// ATR should converge to 2.0 (high-low)
	if math.Abs(last-2.0) > 0.01 {
		t.Fatalf("expected ~2.0, got %v", last)
	}
}

func TestATR_MismatchedLengths(t *testing.T) {
	v := ATR([]float64{1, 2}, []float64{1}, []float64{1, 2}, 14)
	if len(v) != 2 {
		t.Fatalf("expected 2 results, got %d", len(v))
	}
	for _, val := range v {
		if !math.IsNaN(val) {
			t.Fatal("expected all NaN for mismatched lengths")
		}
	}
}
