package feature

import (
	"math"
	"testing"
)

func TestOIDeltaPct_Basic(t *testing.T) {
	v := OIDeltaPct([]float64{100, 110, 121, 133.1}, 1)
	if !math.IsNaN(v[0]) {
		t.Fatal("expected NaN at index 0")
	}
	// (110-100)/100 * 100 = 10%
	if math.Abs(v[1]-10.0) > 0.001 {
		t.Fatalf("expected 10%%, got %v", v[1])
	}
}

func TestOIDeltaPct_Period4(t *testing.T) {
	v := OIDeltaPct([]float64{100, 102, 104, 106, 110}, 4)
	for i := 0; i < 4; i++ {
		if !math.IsNaN(v[i]) {
			t.Fatalf("expected NaN at index %d", i)
		}
	}
	// (110-100)/100 * 100 = 10%
	if math.Abs(v[4]-10.0) > 0.001 {
		t.Fatalf("expected 10%%, got %v", v[4])
	}
}

func TestOIDeltaPct_ZeroPrev(t *testing.T) {
	v := OIDeltaPct([]float64{0, 100}, 1)
	if !math.IsNaN(v[0]) {
		t.Fatal("expected NaN at index 0")
	}
	if !math.IsNaN(v[1]) {
		t.Fatal("expected NaN when prev=0")
	}
}

func TestNormalizedImbalance_Basic(t *testing.T) {
	v := NormalizedImbalance(
		[]float64{100, 200, 0, 300, 50},
		[]float64{50, 100, 0, 100, 150},
	)
	// (100-50)/(100+50) = 50/150 ≈ 0.333
	if math.Abs(v[0]-0.3333) > 0.001 {
		t.Fatalf("expected ~0.333, got %v", v[0])
	}
	// (200-100)/(200+100) = 100/300 ≈ 0.333
	if math.Abs(v[1]-0.3333) > 0.001 {
		t.Fatalf("expected ~0.333, got %v", v[1])
	}
	// 0+0=0 → result = 0
	if v[2] != 0 {
		t.Fatalf("expected 0, got %v", v[2])
	}
	// (300-100)/(300+100) = 200/400 = 0.5
	if math.Abs(v[3]-0.5) > 0.001 {
		t.Fatalf("expected 0.5, got %v", v[3])
	}
	// (50-150)/(50+150) = -100/200 = -0.5
	if math.Abs(v[4]-(-0.5)) > 0.001 {
		t.Fatalf("expected -0.5, got %v", v[4])
	}
}

func TestNormalizedImbalance_Mismatch(t *testing.T) {
	v := NormalizedImbalance([]float64{1, 2}, []float64{1})
	for _, val := range v {
		if !math.IsNaN(val) {
			t.Fatal("expected NaN for mismatched lengths")
		}
	}
}

func TestLSRatioRaw(t *testing.T) {
	v := LSRatioRaw([]float64{1.5, 2.0, 0}, []float64{0.5, 1.0, 1.0})
	// 1.5/0.5 = 3.0
	if math.Abs(v[0]-3.0) > 0.001 {
		t.Fatalf("expected 3.0, got %v", v[0])
	}
	// 2.0/1.0 = 2.0
	if math.Abs(v[1]-2.0) > 0.001 {
		t.Fatalf("expected 2.0, got %v", v[1])
	}
	// 0/1.0 = 0
	if v[2] != 0 {
		t.Fatalf("expected 0, got %v", v[2])
	}
}

func TestLSRatioRaw_DivByZero(t *testing.T) {
	v := LSRatioRaw([]float64{1.0}, []float64{0})
	if !math.IsNaN(v[0]) {
		t.Fatal("expected NaN when sell=0")
	}
}
