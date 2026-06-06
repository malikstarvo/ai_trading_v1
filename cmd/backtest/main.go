package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/avav/ai_trading_v1/internal/backtest"
	"github.com/avav/ai_trading_v1/internal/db"
)

func main() {
	ctx := context.Background()

	cfg := db.ConfigFromEnv()
	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("db pool: %v", err)
	}
	defer pool.Close()

	if err := db.RunAllMigrations(ctx, pool); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	store := db.NewFeatureStore(pool)

	symbol := getEnv("BACKTEST_SYMBOL", "BTCUSDT")
	timeframe := getEnv("BACKTEST_TIMEFRAME", "15m")
	lookbackDays := getEnvInt("BACKTEST_DAYS", 30)
	holdingBars := getEnvInt("BACKTEST_HOLDING_BARS", 4)
	warmupBars := getEnvInt("BACKTEST_WARMUP_BARS", 200)
	initialCapital := getEnvFloat("BACKTEST_INITIAL_CAPITAL", 10000)

	featureSetID, err := store.EnsureDefaultFeatureSet(ctx)
	if err != nil {
		log.Fatalf("feature set: %v", err)
	}

	after := time.Now().Add(-time.Duration(lookbackDays) * 24 * time.Hour)

	log.Printf("Loading candles for %s %s after %s...", symbol, timeframe, after.Format(time.RFC3339))
	candles, err := store.LoadCandlesAfter(ctx, symbol, timeframe, after, 0)
	if err != nil {
		log.Fatalf("load candles: %v", err)
	}
	log.Printf("Loaded %d candles", len(candles))

	log.Printf("Loading features for %s %s (feature set %d)...", symbol, timeframe, featureSetID)
	features, err := store.LoadFeaturesAfter(ctx, symbol, timeframe, featureSetID, after)
	if err != nil {
		log.Fatalf("load features: %v", err)
	}
	log.Printf("Loaded %d features", len(features))

	if len(candles) < warmupBars+holdingBars+2 {
		log.Fatalf("not enough data: %d candles, need >= %d", len(candles), warmupBars+holdingBars+2)
	}

	minLen := len(candles)
	if len(features) < minLen {
		minLen = len(features)
	}
	candles = candles[:minLen]
	features = features[:minLen]

	btCfg := backtest.Config{
		Symbol:         symbol,
		Timeframe:      timeframe,
		InitialCapital: initialCapital,
		Commission:     0.001,
		Slippage:       0.0005,
		HoldingBars:    holdingBars,
		Direction:      "long",
		ATRMultiplier:  2.0,
		WarmupBars:     warmupBars,
		GateConfig:     backtest.DefaultConfig().GateConfig,
	}

	engine := backtest.New(btCfg)
	log.Printf("Running backtest: %s %s, holding=%d, warmup=%d, capital=%.0f",
		symbol, timeframe, holdingBars, warmupBars, initialCapital)

	start := time.Now()
	summary, err := engine.Run(candles, features)
	elapsed := time.Since(start)
	if err != nil {
		log.Fatalf("backtest failed: %v", err)
	}

	m := summary.Metrics
	fmt.Println()
	fmt.Println("=== Backtest Results ===")
	fmt.Printf("Symbol:       %s\n", symbol)
	fmt.Printf("Timeframe:    %s\n", timeframe)
	fmt.Printf("Period:       %s – %s\n", candles[warmupBars].Time.Format("2006-01-02"), candles[len(candles)-1].Time.Format("2006-01-02"))
	fmt.Printf("Duration:     %v\n", elapsed)
	fmt.Println()
	fmt.Printf("Total Trades:    %d\n", m.TotalTrades)
	fmt.Printf("Winning Trades:  %d\n", m.WinningTrades)
	fmt.Printf("Losing Trades:   %d\n", m.LosingTrades)
	fmt.Printf("Win Rate:        %.1f%%\n", m.WinRate)
	fmt.Printf("Profit Factor:   %.2f\n", m.ProfitFactor)
	fmt.Printf("Net PnL:         $%.2f\n", m.NetPnL)
	fmt.Printf("Sharpe Ratio:    %.2f\n", m.SharpeRatio)
	fmt.Printf("Sortino Ratio:   %.2f\n", m.SortinoRatio)
	fmt.Printf("Max Drawdown:    %.1f%%\n", m.MaxDrawdownPct)
	fmt.Printf("Expectancy:      $%.2f\n", m.Expectancy)
	fmt.Printf("Avg Return:      %.2f%%\n", m.AvgReturnPct)
	fmt.Printf("Avg Holding:     %.1f bars\n", m.AvgHoldingBars)
	fmt.Printf("Exposure:        %.1f%%\n", m.ExposurePct)
	fmt.Println()

	if m.TotalTrades > 0 {
		fmt.Println("Top 5 Trades by PnL:")
		trades := summary.Trades
		for i := 0; i < 5 && i < len(trades); i++ {
			t := trades[i]
			fmt.Printf("  %s → %s | %s | PnL=$%.2f | return=%.2f%% | reason=%s\n",
				t.EntryTime.Format("2006-01-02 15:04"),
				t.ExitTime.Format("2006-01-02 15:04"),
				t.Direction, t.PnL, t.ReturnPct*100, t.ExitReason)
		}
	}
}

func getEnv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func getEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	var n int
	fmt.Sscanf(v, "%d", &n)
	return n
}

func getEnvFloat(key string, def float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	var n float64
	fmt.Sscanf(v, "%f", &n)
	return n
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
