package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/avav/ai_trading_v1/internal/agent/tradegate"
	"github.com/avav/ai_trading_v1/internal/db"
	papertrade "github.com/avav/ai_trading_v1/internal/papertrade"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := db.ConfigFromEnv()
	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("db pool: %v", err)
	}
	defer pool.Close()

	if err := db.RunAllMigrations(ctx, pool); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	paperStore := papertrade.NewPaperStore(pool)
	fStore := db.NewFeatureStore(pool)

	engineCfg := papertrade.EngineConfig{
		Symbol:              getEnv("PAPER_SYMBOL", "BTCUSDT"),
		Timeframe:           getEnv("PAPER_TIMEFRAME", "15m"),
		InitialCapital:      getEnvFloat("PAPER_INITIAL_CAPITAL", 10000),
		Commission:          getEnvFloat("PAPER_COMMISSION", 0.00055),
		Slippage:            getEnvFloat("PAPER_SLIPPAGE", 0.0005),
		ATRMultiplier:       getEnvFloat("PAPER_ATR_MULTIPLIER", 2.0),
		HoldingBars:         getEnvInt("PAPER_HOLDING_BARS", 24),
		PollInterval:        time.Duration(getEnvInt("PAPER_POLL_INTERVAL_SEC", 60)) * time.Second,
		RiskPerTradePct:     getEnvFloat("PAPER_RISK_PCT", 1.0),
		MaxDailyDrawdownPct: getEnvFloat("PAPER_MAX_DAILY_DD_PCT", 5.0),
		MaxTotalDrawdownPct: getEnvFloat("PAPER_MAX_TOTAL_DD_PCT", 15.0),
		LongThreshold:       getEnvFloat("PAPER_LONG_THRESHOLD", 60.0),
		ShortThreshold:      getEnvFloat("PAPER_SHORT_THRESHOLD", 40.0),
		GateConfig:          tradegate.DefaultConfig(),
		MLAPIURL:            getEnv("PAPER_ML_API_URL", ""),
	}
	engineCfg.GateConfig.StartingCapital = engineCfg.InitialCapital

	engine := papertrade.New(engineCfg, paperStore, fStore)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Shutting down...")
		cancel()
	}()

	// Print status periodically
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				st := engine.State()
				fmt.Println()
				fmt.Println("=== Paper Trader Status ===")
				fmt.Printf("Uptime:        %s\n", st.Uptime.Round(time.Second))
				fmt.Printf("State:         %s\n", st.State)
				fmt.Printf("Bars seen:     %d\n", st.BarCount)
				fmt.Printf("Equity:        $%.2f\n", st.Equity)
				fmt.Printf("Total PnL:     $%.2f (%.2f%%)\n", st.TotalPnL, st.TotalPnL/engineCfg.InitialCapital*100)
				fmt.Printf("Day PnL:       $%.2f\n", st.DayPnL)
				fmt.Printf("Day Trades:    %d\n", st.DayTrades)
				if st.Position != nil {
					fmt.Printf("Position:      %s #%d entry=%.2f stop=%.2f bars=%d\n",
						st.Position.Direction, st.Position.ID, st.Position.EntryPrice, st.Position.StopPrice, st.Position.BarsHeld)
				} else {
					fmt.Printf("Position:      none\n")
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	log.Printf("Paper Trader started: %s %s capital=%.0f long>=%.0f short<=%.0f",
		engineCfg.Symbol, engineCfg.Timeframe, engineCfg.InitialCapital,
		engineCfg.LongThreshold, engineCfg.ShortThreshold)

	if err := engine.Run(ctx); err != nil {
		log.Fatalf("engine: %v", err)
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
