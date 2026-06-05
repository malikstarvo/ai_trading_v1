package feature

import (
	"math"
	"testing"
)

func TestRSI_ShortInput(t *testing.T) {
	v := RSI([]float64{10, 20, 30}, 14)
	if len(v) != 3 {
		t.Fatalf("expected 3, got %d", len(v))
	}
	for _, val := range v {
		if !math.IsNaN(val) {
			t.Fatal("expected NaN for short input")
		}
	}
}

func TestRSI_AllUp(t *testing.T) {
	closes := []float64{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24}
	v := RSI(closes, 14)
	if len(v) != 15 {
		t.Fatalf("expected 15, got %d", len(v))
	}
	// First 14 are NaN (index 0 is the first change, so index 14 is bar 14 = first valid)
	for i := 0; i < 14; i++ {
		if !math.IsNaN(v[i]) {
			t.Fatalf("expected NaN at index %d", i)
		}
	}
	// All up → RSI should be 100
	if v[14] != 100 {
		t.Fatalf("expected 100 for all-up trend, got %v", v[14])
	}
}

func TestRSI_AllDown(t *testing.T) {
	closes := []float64{30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17, 16}
	v := RSI(closes, 14)
	for i := 0; i < 14; i++ {
		if !math.IsNaN(v[i]) {
			t.Fatalf("expected NaN at index %d", i)
		}
	}
	// All down → RSI should be 0
	if v[14] != 0 {
		t.Fatalf("expected 0 for all-down trend, got %v", v[14])
	}
}

func TestRSI_Flat(t *testing.T) {
	closes := []float64{}
	for i := 0; i < 20; i++ {
		closes = append(closes, 100)
	}
	v := RSI(closes, 14)
	// Flat → avgGain=avgLoss=0 → RSI = 50
	last := v[len(v)-1]
	if last != 50 {
		t.Fatalf("expected 50 for flat market, got %v", last)
	}
}

func TestRSI_Alternating(t *testing.T) {
	closes := []float64{100, 101, 100, 101, 100, 101, 100, 101, 100, 101, 100, 101, 100, 101, 100, 101}
	v := RSI(closes, 14)
	last := v[len(v)-1]
	if last < 40 || last > 60 {
		t.Fatalf("expected RSI near 50 for alternating, got %v", last)
	}
}
