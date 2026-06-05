package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/avav/ai_trading_v1/internal/collector"
	"github.com/avav/ai_trading_v1/internal/collector/metrics"
	"github.com/avav/ai_trading_v1/internal/db"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	ctx := context.Background()

	cfg := db.ConfigFromEnv()
	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		log.Error("db pool failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := db.RunAllMigrations(ctx, pool); err != nil {
		log.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	candleStore := db.NewCandleStore(pool)
	orderFlowStore := db.NewOrderFlowStore(pool)
	healthStore := db.NewCollectorHealthStore(pool)

	collectorCfg := collector.DefaultConfig()
	collectorCfg.Testnet = os.Getenv("COLLECTOR_TESTNET") != "false"

	if u := os.Getenv("COLLECTOR_WS_URL"); u != "" {
		collectorCfg.WSURL = u
	}
	if u := os.Getenv("COLLECTOR_REST_URL"); u != "" {
		collectorCfg.BaseURL = u
	}
	collectorCfg.InsecureSkipVerify = os.Getenv("COLLECTOR_INSECURE_SKIP_VERIFY") == "true"

	reg := prometheus.NewRegistry()
	m := metrics.New(reg)

	svc := collector.NewService(collectorCfg, candleStore, orderFlowStore, healthStore, m, log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Info("signal received", "signal", sig)
		cancel()
	}()

	if err := svc.Start(ctx); err != nil {
		log.Error("start failed", "error", err)
		os.Exit(1)
	}

	<-ctx.Done()

	if err := svc.Stop(); err != nil {
		log.Error("stop failed", "error", err)
	}

	log.Info("shutdown complete")
}
