package recovery

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/avav/ai_trading_v1/internal/collector/bybit"
	"github.com/avav/ai_trading_v1/internal/collector/metrics"
	"github.com/avav/ai_trading_v1/internal/collector/store"
	"github.com/avav/ai_trading_v1/internal/model"
)

type Detector struct {
	candleStore store.CandleStore
	bybitClient *bybit.Client
	symbols     []string
	timeframes  []model.Timeframe
	logger      *slog.Logger
	metrics     *metrics.CollectorMetrics
	gapBars     int
}

func NewDetector(
	candleStore store.CandleStore,
	bybitClient *bybit.Client,
	symbols []string,
	m *metrics.CollectorMetrics,
	logger *slog.Logger,
	gapBars int,
) *Detector {
	return &Detector{
		candleStore: candleStore,
		bybitClient: bybitClient,
		symbols:     symbols,
		timeframes:  []model.Timeframe{model.Timeframe15m, model.Timeframe1h},
		logger:      logger.With("module", "gap_detector"),
		metrics:     m,
		gapBars:     gapBars,
	}
}

func (d *Detector) Run(ctx context.Context) ([]GapReport, error) {
	var reports []GapReport

	for _, symbol := range d.symbols {
		for _, tf := range d.timeframes {
			report, err := d.detectGap(ctx, symbol, tf)
			if err != nil {
				d.logger.Error("detect gap failed", "symbol", symbol, "timeframe", tf, "error", err)
				continue
			}
			if report.HasGap {
				reports = append(reports, report)
				if d.metrics != nil {
					d.metrics.GapEventsTotal.WithLabelValues(symbol, string(tf)).Inc()
				}
			}
		}
	}
	return reports, nil
}

func (d *Detector) detectGap(ctx context.Context, symbol string, tf model.Timeframe) (GapReport, error) {
	dbLatest, err := d.candleStore.LatestTime(ctx, symbol, string(tf))
	if err != nil {
		return GapReport{}, fmt.Errorf("db latest %s %s: %w", symbol, tf, err)
	}
	if dbLatest.IsZero() {
		return GapReport{}, nil
	}

	exchangeLatest, err := d.bybitClient.LatestCandleTime(ctx, symbol, tf.BybitInterval())
	if err != nil {
		return GapReport{}, fmt.Errorf("exchange latest %s %s: %w", symbol, tf, err)
	}

	diff := exchangeLatest.Sub(dbLatest)
	threshold := tf.Duration() * time.Duration(d.gapBars)

	if diff > threshold {
		missingBars := int(diff / tf.Duration())
		return GapReport{
			Symbol:      symbol,
			Timeframe:   string(tf),
			GapStart:    dbLatest,
			GapEnd:      exchangeLatest,
			MissingBars: missingBars,
			HasGap:      true,
		}, nil
	}

	return GapReport{HasGap: false}, nil
}
