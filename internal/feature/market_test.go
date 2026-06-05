package feature

import (
	"math"
	"testing"
)

func TestLogReturn_Basic(t *testing.T) {
	v := LogReturn([]float64{100, 110, 121, 133.1}, 1)
	if len(v) != 4 {
		t.Fatalf("expected 4, got %d", len(v))
	}
	if !math.IsNaN(v[0]) {
		t.Fatal("expected NaN at index 0")
	}
	// ln(110/100) * 100 ≈ 9.53
	expected := math.Log(110.0/100.0) * 100
	if math.Abs(v[1]-expected) > 0.01 {
		t.Fatalf("expected %v, got %v", expected, v[1])
	}
}

func TestLogReturn_Negative(t *testing.T) {
	v := LogReturn([]float64{100, 90, 81, 72.9}, 1)
	// ln(90/100) * 100 ≈ -10.54
	expected := math.Log(90.0/100.0) * 100
	if math.Abs(v[1]-expected) > 0.01 {
		t.Fatalf("expected %v, got %v", expected, v[1])
	}
}

func TestLogReturn_Period4(t *testing.T) {
	v := LogReturn([]float64{100, 101, 102, 103, 110}, 4)
	if !math.IsNaN(v[0]) || !math.IsNaN(v[1]) || !math.IsNaN(v[2]) || !math.IsNaN(v[3]) {
		t.Fatal("expected NaN before period 4")
	}
	expected := math.Log(110.0/100.0) * 100
	if math.Abs(v[4]-expected) > 0.01 {
		t.Fatalf("expected %v, got %v", expected, v[4])
	}
}

func TestVolatility_Basic(t *testing.T) {
	// Constant returns → volatility = 0
	ret := []float64{1, 1, 1, 1, 1}
	v := Volatility(ret, 3)
	if !math.IsNaN(v[0]) || !math.IsNaN(v[1]) {
		t.Fatal("expected NaN before window")
	}
	for i := 2; i < len(ret); i++ {
		if v[i] != 0 {
			t.Fatalf("expected 0 for constant returns, got %v", v[i])
		}
	}
}

func TestVolatility_Oscillating(t *testing.T) {
	ret := []float64{1, -1, 1, -1, 1, -1}
	v := Volatility(ret, 3)
	last := v[len(v)-1]
	if last <= 0 {
		t.Fatal("expected positive volatility for oscillating returns")
	}
}

func TestVolatility_TooSmall(t *testing.T) {
	v := Volatility([]float64{1}, 3)
	if len(v) != 1 || !math.IsNaN(v[0]) {
		t.Fatal("expected NaN for window > input")
	}
}

func TestZScore_Basic(t *testing.T) {
	v := ZScore([]float64{10, 10, 10, 10, 10, 20}, 5)
	// First 4 are NaN
	for i := 0; i < 4; i++ {
		if !math.IsNaN(v[i]) {
			t.Fatalf("expected NaN at index %d", i)
		}
	}
	// At index 4: values{10,10,10,10,10} → mean=10, std=0 → z=0
	if v[4] != 0 {
		t.Fatalf("expected 0 at index 4, got %v", v[4])
	}
	// At index 5: values{10,10,10,10,20} → mean=12, std=4 → z=(20-12)/4 = 2
	if math.Abs(v[5]-2.0) > 0.01 {
		t.Fatalf("expected ~2.0 at index 5, got %v", v[5])
	}
}

func TestZScore_ShortInput(t *testing.T) {
	v := ZScore([]float64{1}, 5)
	if len(v) != 1 || !math.IsNaN(v[0]) {
		t.Fatal("expected NaN for short input")
	}
}
