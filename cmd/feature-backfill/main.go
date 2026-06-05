package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/avav/ai_trading_v1/internal/db"
	"github.com/avav/ai_trading_v1/internal/feature"
	"github.com/avav/ai_trading_v1/internal/label"
	"github.com/avav/ai_trading_v1/internal/model"
)

type symbolTF struct {
	Symbol    string
	Timeframe string
}

const chunkSize = 2200
const warmupBars = 200
const heartbeatInterval = 1000

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

	featureStore := db.NewFeatureStore(pool)
	jobStore := db.NewJobStore(pool)

	activeSetID, err := featureStore.EnsureDefaultFeatureSet(ctx)
	if err != nil {
		log.Fatalf("feature set: %v", err)
	}

	pairs := []symbolTF{
		{"BTCUSDT", "15m"},
		{"BTCUSDT", "1h"},
		{"ETHUSDT", "15m"},
		{"ETHUSDT", "1h"},
		{"SOLUSDT", "15m"},
		{"SOLUSDT", "1h"},
	}

	for _, p := range pairs {
		log.Printf("Backfilling %s %s...", p.Symbol, p.Timeframe)
		if err := processPair(ctx, featureStore, jobStore, p.Symbol, p.Timeframe, activeSetID); err != nil {
			log.Printf("ERROR %s %s: %v", p.Symbol, p.Timeframe, err)
		}
	}
}

func processPair(ctx context.Context, fs *db.FeatureStore, js *db.JobStore, symbol, timeframe string, featureSetID int) error {
	params := db.JobParams{
		LabelThreshold: 0.25,
		Horizons:       []int{4, 12, 24},
		MinCandles:     200,
	}

	jobID, err := js.Create(ctx, symbol, timeframe, featureSetID, params)
	if err != nil {
		return fmt.Errorf("create job: %w", err)
	}
	if err := js.Start(ctx, jobID); err != nil {
		return fmt.Errorf("start job: %w", err)
	}

	totalWritten := 0

	defer func() {
		if r := recover(); r != nil {
			js.Fail(ctx, jobID, fmt.Sprintf("panic: %v", r))
			panic(r)
		}
	}()

	var allCandles []model.Candle
	lastTime := time.Time{}

	for {
		chunk, err := fs.LoadCandlesAfter(ctx, symbol, timeframe, lastTime, chunkSize)
		if err != nil {
			js.Fail(ctx, jobID, err.Error())
			return fmt.Errorf("load candles: %w", err)
		}
		if len(chunk) == 0 {
			break
		}

		allCandles = append(allCandles, chunk...)

		if len(chunk) < warmupBars {
			break
		}

		candleTimes := make([]time.Time, len(chunk))
		for i, c := range chunk {
			candleTimes[i] = c.Time
		}

		of, err := fs.LoadOrderFlow(ctx, symbol, timeframe, candleTimes)
		if err != nil {
			js.Fail(ctx, jobID, err.Error())
			return fmt.Errorf("load orderflow: %w", err)
		}

		rows, err := feature.ComputeFeatures(chunk, of, featureSetID)
		if err != nil {
			js.Fail(ctx, jobID, err.Error())
			return fmt.Errorf("compute features: %w", err)
		}

		valid := rows[warmupBars:]
		if len(valid) == 0 {
			break
		}

		if err := db.ValidateFeatureRows(valid); err != nil {
			js.Fail(ctx, jobID, err.Error())
			return fmt.Errorf("validation: %w", err)
		}

		if err := fs.UpsertFeatures(ctx, valid); err != nil {
			js.Fail(ctx, jobID, err.Error())
			return fmt.Errorf("upsert features: %w", err)
		}

		totalWritten += len(valid)
		log.Printf("  %s %s: %d features written (total %d)", symbol, timeframe, len(valid), totalWritten)

		if totalWritten%heartbeatInterval < len(valid) {
			if err := js.Heartbeat(ctx, jobID); err != nil {
				log.Printf("  WARN: heartbeat failed: %v", err)
			}
		}

		if err := js.UpdateProgress(ctx, jobID, totalWritten); err != nil {
			log.Printf("  WARN: progress update failed: %v", err)
		}

		if len(chunk) < chunkSize {
			break
		}

		lastTime = chunk[len(chunk)-warmupBars-1].Time
	}

	if len(allCandles) >= 200 {
		labels, err := label.ComputeLabels(allCandles, []int{4, 12, 24}, 0.25)
		if err != nil {
			js.Fail(ctx, jobID, err.Error())
			return fmt.Errorf("compute labels: %w", err)
		}

		if err := fs.UpsertLabels(ctx, labels, featureSetID); err != nil {
			js.Fail(ctx, jobID, err.Error())
			return fmt.Errorf("upsert labels: %w", err)
		}
		log.Printf("  %s %s: %d labels written", symbol, timeframe, len(labels))
	}

	if err := js.Complete(ctx, jobID, totalWritten); err != nil {
		return fmt.Errorf("complete job: %w", err)
	}

	log.Printf("  DONE %s %s: %d total features", symbol, timeframe, totalWritten)
	return nil
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
