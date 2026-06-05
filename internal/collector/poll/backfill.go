package poll

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/avav/ai_trading_v1/internal/collector/bybit"
	"github.com/avav/ai_trading_v1/internal/collector/mapper"
	"github.com/avav/ai_trading_v1/internal/collector/store"
)

type Backfill struct {
	client      *bybit.Client
	symbols     []string
	timeframes  []struct {
		Name           string
		BybitInterval  string
	}
	store       store.CandleStore
	logger      *slog.Logger
	daysHistory int
}

func NewBackfill(client *bybit.Client, symbols []string, candleStore store.CandleStore, daysHistory int, logger *slog.Logger) *Backfill {
	return &Backfill{
		client:      client,
		symbols:     symbols,
		timeframes: []struct {
			Name          string
			BybitInterval string
		}{
			{Name: "15m", BybitInterval: "15"},
			{Name: "1h", BybitInterval: "60"},
		},
		store:       candleStore,
		logger:      logger.With("module", "backfill"),
		daysHistory: daysHistory,
	}
}

func (b *Backfill) Run(ctx context.Context) error {
	for _, symbol := range b.symbols {
		for _, tf := range b.timeframes {
			if err := b.backfillSymbol(ctx, symbol, tf.Name, tf.BybitInterval); err != nil {
				b.logger.Error("backfill failed",
					"symbol", symbol,
					"timeframe", tf.Name,
					"error", err,
				)
				return err
			}
		}
	}
	return nil
}

func (b *Backfill) backfillSymbol(ctx context.Context, symbol, timeframe, bybitInterval string) error {
	latest, err := b.store.LatestTime(ctx, symbol, timeframe)
	if err != nil || latest.IsZero() {
		latest = time.Now().AddDate(0, 0, -b.daysHistory)
	}

	end := time.Now()
	if latest.After(end.Add(-time.Hour)) {
		b.logger.Info("skip backfill (recent)", "symbol", symbol, "timeframe", timeframe)
		return nil
	}

	b.logger.Info("backfilling",
		"symbol", symbol,
		"timeframe", timeframe,
		"from", latest,
	)

	cursor := latest.UnixMilli()
	total := 0
	for cursor < end.UnixMilli() {
		resp, err := b.client.GetKlines(ctx, symbol, bybitInterval, cursor, end.UnixMilli(), 200)
		if err != nil {
			return fmt.Errorf("fetch %s %s: %w", symbol, timeframe, err)
		}
		if len(resp.List) == 0 {
			break
		}
		candles := mapper.RestKlinesToCandles(resp, symbol, timeframe)
		if err := b.store.InsertBatch(ctx, candles); err != nil {
			return fmt.Errorf("store %s %s: %w", symbol, timeframe, err)
		}
		total += len(candles)
		if len(candles) < 200 {
			break
		}
		cursor = candles[len(candles)-1].Time.UnixMilli() + 1
	}

	b.logger.Info("backfill complete",
		"symbol", symbol,
		"timeframe", timeframe,
		"total_candles", total,
	)
	return nil
}
