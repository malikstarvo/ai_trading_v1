package poll

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/avav/ai_trading_v1/internal/collector/bybit"
	"github.com/avav/ai_trading_v1/internal/collector/mapper"
	"github.com/avav/ai_trading_v1/internal/collector/metrics"
	"github.com/avav/ai_trading_v1/internal/collector/store"
	"github.com/avav/ai_trading_v1/internal/collector/validate"
)

type LSRatioPoller struct {
	client   *bybit.Client
	symbols  []string
	store    store.OrderFlowStore
	validate validate.Validator
	metrics  *metrics.CollectorMetrics
	logger   *slog.Logger
}

func NewLSRatioPoller(
	client *bybit.Client,
	symbols []string,
	orderFlowStore store.OrderFlowStore,
	v validate.Validator,
	m *metrics.CollectorMetrics,
	logger *slog.Logger,
) *LSRatioPoller {
	return &LSRatioPoller{
		client:   client,
		symbols:  symbols,
		store:    orderFlowStore,
		validate: v,
		metrics:  m,
		logger:   logger.With("module", "poller_ls_ratio"),
	}
}

func (p *LSRatioPoller) Name() string {
	return "long_short_ratio"
}

func (p *LSRatioPoller) Interval() time.Duration {
	return 15 * time.Minute
}

func (p *LSRatioPoller) Poll(ctx context.Context) error {
	for _, symbol := range p.symbols {
		resp, err := p.client.GetLongShortRatio(ctx, symbol, "15min", 500)
		if err != nil {
			return fmt.Errorf("%s %s: %w", p.Name(), symbol, err)
		}
		for _, item := range resp.List {
			ls := mapper.RestToLSRatio(item, symbol, "15min")
			if err := p.validate.ValidateLSRatio(&ls); err != nil {
				p.logger.Warn("validation failed", "error", err)
				continue
			}
			if err := p.store.InsertLSRatio(ctx, &ls); err != nil {
				return fmt.Errorf("store ls_ratio %s: %w", symbol, err)
			}
			if p.metrics != nil {
				p.metrics.StoredTotal.WithLabelValues("long_short_ratio").Inc()
			}
		}
	}
	return nil
}
