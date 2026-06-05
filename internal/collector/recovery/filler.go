package recovery

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/avav/ai_trading_v1/internal/collector/bybit"
	"github.com/avav/ai_trading_v1/internal/collector/mapper"
	"github.com/avav/ai_trading_v1/internal/collector/store"
)

type Filler struct {
	bybitClient *bybit.Client
	candleStore store.CandleStore
	logger      *slog.Logger
}

func NewFiller(
	bybitClient *bybit.Client,
	candleStore store.CandleStore,
	logger *slog.Logger,
) *Filler {
	return &Filler{
		bybitClient: bybitClient,
		candleStore: candleStore,
		logger:      logger.With("module", "gap_filler"),
	}
}

func (f *Filler) Fill(ctx context.Context, report GapReport) error {
	if !report.HasGap {
		return nil
	}

	f.logger.Warn("filling gap",
		"symbol", report.Symbol,
		"timeframe", report.Timeframe,
		"missing_bars", report.MissingBars,
		"from", report.GapStart,
		"to", report.GapEnd,
	)

	bybitInterval := intervalToBybit(report.Timeframe)
	cursor := report.GapStart.UnixMilli()
	end := report.GapEnd.UnixMilli()
	total := 0

	for cursor < end {
		resp, err := f.bybitClient.GetKlines(ctx, report.Symbol, bybitInterval, cursor, end, 200)
		if err != nil {
			return fmt.Errorf("fetch gap: %w", err)
		}
		if len(resp.List) == 0 {
			break
		}
		candles := mapper.RestKlinesToCandles(resp, report.Symbol, report.Timeframe)
		if err := f.candleStore.InsertBatch(ctx, candles); err != nil {
			return fmt.Errorf("store gap: %w", err)
		}
		total += len(candles)
		if len(candles) < 200 {
			break
		}
		cursor = candles[len(candles)-1].Time.UnixMilli() + 1
	}

	f.logger.Info("gap filled",
		"symbol", report.Symbol,
		"timeframe", report.Timeframe,
		"total_candles", total,
	)
	return nil
}

func intervalToBybit(tf string) string {
	switch tf {
	case "15m":
		return "15"
	case "1h":
		return "60"
	default:
		return tf
	}
}

func intervalDuration(tf string) time.Duration {
	switch tf {
	case "15m":
		return 15 * time.Minute
	case "1h":
		return time.Hour
	default:
		return 0
	}
}
