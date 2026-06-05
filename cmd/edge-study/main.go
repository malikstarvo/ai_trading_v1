package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/avav/ai_trading_v1/internal/db"
	"github.com/avav/ai_trading_v1/internal/edge"
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

	edgeStore := db.NewEdgeStore(pool)
	resStore := db.NewResearchStore(pool)
	featureStore := db.NewFeatureStore(pool)

	featureSetID, err := featureStore.EnsureDefaultFeatureSet(ctx)
	if err != nil {
		log.Fatalf("feature set: %v", err)
	}

	symbols := getEnvSlice("STUDY_SYMBOLS", []string{"BTCUSDT", "ETHUSDT", "SOLUSDT"})
	timeframes := getEnvSlice("STUDY_TIMEFRAMES", []string{"15m", "1h"})

	features := []edge.FeatureInfo{
		{Name: "EMA20", Col: "ema20"},
		{Name: "EMA50", Col: "ema50"},
		{Name: "EMA200", Col: "ema200"},
		{Name: "RSI14", Col: "rsi14"},
		{Name: "ATR14", Col: "atr14"},
		{Name: "ADX14", Col: "adx14"},
		{Name: "VolumeEMA20", Col: "volume_ema20"},
		{Name: "PriceAboveEMA20", Col: "price_above_ema20"},
		{Name: "PriceAboveEMA50", Col: "price_above_ema50"},
		{Name: "PriceAboveEMA200", Col: "price_above_ema200"},
		{Name: "OIDelta1Pct", Col: "oi_delta_1_pct"},
		{Name: "OIDelta4Pct", Col: "oi_delta_4_pct"},
		{Name: "OIDelta12Pct", Col: "oi_delta_12_pct"},
		{Name: "OIZScore30", Col: "oi_zscore_30"},
		{Name: "FundingRate", Col: "funding_rate"},
		{Name: "FundingZScore30", Col: "funding_zscore_30"},
		{Name: "LSRatioRaw", Col: "ls_ratio_raw"},
		{Name: "LSRatioNormalized", Col: "ls_ratio_normalized"},
		{Name: "LiqLongUSD", Col: "liq_long_usd"},
		{Name: "LiqShortUSD", Col: "liq_short_usd"},
		{Name: "LiqImbalance", Col: "liq_imbalance"},
		{Name: "Return1", Col: "return_1"},
		{Name: "Return4", Col: "return_4"},
		{Name: "Return12", Col: "return_12"},
		{Name: "Volatility14", Col: "volatility_14"},
		{Name: "Volatility50", Col: "volatility_50"},
	}

	studyCfg := edge.StudyConfig{
		Symbols:          symbols,
		Timeframes:       timeframes,
		FeatureSetID:     featureSetID,
		LabelHorizons: []edge.LabelHorizon{
			{Name: "future_return_4", Col: "future_return_4"},
			{Name: "future_return_12", Col: "future_return_12"},
			{Name: "future_return_24", Col: "future_return_24"},
		},
		RollingWindows:   []int{50, 100, 200},
		QuantileNBuckets: 10,
		RegimePct:        0.30,
		Features:         features,
	}

	study := edge.NewStudy(edgeStore, resStore, studyCfg)

	configJSON, _ := json.Marshal(studyCfg)
	runID, err := resStore.CreateRun(ctx, "Edge Validation Study", symbols[0], timeframes[0], featureSetID, configJSON)
	if err != nil {
		log.Fatalf("create run: %v", err)
	}
	resStore.StartRun(ctx, runID)

	log.Printf("Starting edge validation study (run %d)...", runID)
	start := time.Now()

	if err := study.RunAll(ctx); err != nil {
		resStore.FailRun(ctx, runID, err.Error())
		log.Fatalf("study failed: %v", err)
	}

	elapsed := time.Since(start)
	log.Printf("Study completed in %v", elapsed)

	if err := study.SaveToDB(ctx, runID); err != nil {
		log.Fatalf("save to db: %v", err)
	}

	resStore.CompleteRun(ctx, runID)

	exportDir := "research_artifacts"
	os.MkdirAll(exportDir, 0755)

	if err := study.ExportHTML(filepath.Join(exportDir, "edge_report.html")); err != nil {
		log.Printf("WARN: html export: %v", err)
	} else {
		log.Printf("Exported: %s/edge_report.html", exportDir)
	}

	if err := study.ExportCSV(filepath.Join(exportDir, "feature_ranking.csv")); err != nil {
		log.Printf("WARN: csv export: %v", err)
	} else {
		log.Printf("Exported: %s/feature_ranking.csv", exportDir)
	}

	if err := study.ExportJSON(filepath.Join(exportDir, "metrics.json")); err != nil {
		log.Printf("WARN: json export: %v", err)
	} else {
		log.Printf("Exported: %s/metrics.json", exportDir)
	}

	top := study.TopFeatures(5)
	log.Printf("Top 5 features: %v", top)
}

func getEnvSlice(key string, def []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	var result []string
	return append(result, v)
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
