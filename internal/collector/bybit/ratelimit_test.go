package bybit_test

import (
	"context"
	"testing"
	"time"

	"github.com/avav/ai_trading_v1/internal/collector/bybit"
)

func TestTokenBucket_Take(t *testing.T) {
	tb := bybit.NewTokenBucket(10, 5)

	ctx := context.Background()
	start := time.Now()
	for i := 0; i < 5; i++ {
		if err := tb.Wait(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	// First 5 tokens should be instantaneous (burst = 5)
	elapsed := time.Since(start)
	if elapsed > 100*time.Millisecond {
		t.Errorf("expected burst to be fast, took %v", elapsed)
	}
}

func TestTokenBucket_RateLimit(t *testing.T) {
	tb := bybit.NewTokenBucket(100, 1)

	ctx := context.Background()
	start := time.Now()
	for i := 0; i < 3; i++ {
		if err := tb.Wait(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	// With rate=100, waiting for 3 tokens should take about 20ms
	elapsed := time.Since(start)
	if elapsed < 5*time.Millisecond {
		t.Errorf("expected rate limiting to slow down, took %v", elapsed)
	}
}

func TestTokenBucket_ContextCancel(t *testing.T) {
	tb := bybit.NewTokenBucket(1, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := tb.Wait(ctx)
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}

func TestTokenBucket_Concurrent(t *testing.T) {
	tb := bybit.NewTokenBucket(1000, 10)

	ctx := context.Background()
	done := make(chan struct{}, 10)

	for i := 0; i < 10; i++ {
		go func() {
			if err := tb.Wait(ctx); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			done <- struct{}{}
		}()
	}

	timeout := time.After(time.Second)
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-timeout:
			t.Fatal("timeout waiting for concurrent token bucket")
		}
	}
}
