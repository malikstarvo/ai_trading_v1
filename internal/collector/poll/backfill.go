package poll

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/avav/ai_trading_v1/internal/collector/bybit"
	"github.com/avav/ai_trading_v1/internal/collector/mapper"
	"github.com/avav/ai_trading_v1/internal/collector/store"
)

const maxRetries = 3
const batchSize = 1000

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
	var wg sync.WaitGroup
	var idx int
	for _, symbol := range b.symbols {
		for _, tf := range b.timeframes {
			if idx > 0 {
				time.Sleep(500 * time.Millisecond)
			}
			idx++
			wg.Add(1)
			go func(sym, name, interval string) {
				defer wg.Done()
				if err := b.backfillSymbol(ctx, sym, name, interval); err != nil {
					b.logger.Error("backfill failed",
						"symbol", sym,
						"timeframe", name,
						"error", err,
					)
				}
			}(symbol, tf.Name, tf.BybitInterval)
		}
	}
	wg.Wait()
	return nil
}

func (b *Backfill) backfillSymbol(ctx context.Context, symbol, timeframe, bybitInterval string) error {
	end := time.Now()
	latest := end.AddDate(0, 0, -b.daysHistory)

	b.logger.Info("backfilling",
		"symbol", symbol,
		"timeframe", timeframe,
		"from", latest,
	)

	startUnix := latest.UnixMilli()
	cursor := end.UnixMilli()
	total := 0
	for cursor > startUnix {
		var resp *bybit.KlineResponse
		var err error
		for attempt := 0; attempt < maxRetries; attempt++ {
			resp, err = b.client.GetKlines(ctx, symbol, bybitInterval, startUnix, cursor, batchSize)
			if err == nil {
				break
			}
			if strings.Contains(err.Error(), "10006") {
				sleep := time.Duration(math.Pow(2, float64(attempt))) * time.Second
				b.logger.Warn("rate limited, retrying",
					"symbol", symbol,
					"timeframe", timeframe,
					"attempt", attempt+1,
					"max", maxRetries,
					"sleep", sleep,
				)
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(sleep):
				}
				continue
			}
			return fmt.Errorf("fetch %s %s: %w", symbol, timeframe, err)
		}
		if err != nil {
			return fmt.Errorf("fetch %s %s after %d retries: %w", symbol, timeframe, maxRetries, err)
		}
		if len(resp.List) == 0 {
			break
		}
		candles := mapper.RestKlinesToCandles(resp, symbol, timeframe)
		sort.Slice(candles, func(i, j int) bool {
			return candles[i].Time.Before(candles[j].Time)
		})
		if err := b.store.InsertBatch(ctx, candles); err != nil {
			return fmt.Errorf("store %s %s: %w", symbol, timeframe, err)
		}
		total += len(candles)
		if len(candles) < batchSize {
			break
		}
		cursor = candles[0].Time.UnixMilli() - 1
	}

	b.logger.Info("backfill complete",
		"symbol", symbol,
		"timeframe", timeframe,
		"total_candles", total,
	)
	return nil
}
