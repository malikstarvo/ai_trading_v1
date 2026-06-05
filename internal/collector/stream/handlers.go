package stream

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/avav/ai_trading_v1/internal/collector/mapper"
	"github.com/avav/ai_trading_v1/internal/collector/metrics"
	"github.com/avav/ai_trading_v1/internal/collector/store"
	"github.com/avav/ai_trading_v1/internal/collector/validate"
)

type Handlers struct {
	candleStore    store.CandleStore
	orderFlowStore store.OrderFlowStore
	validator      validate.Validator
	metrics        *metrics.CollectorMetrics
	logger         *slog.Logger
	ctx            context.Context
}

func NewHandlers(
	ctx context.Context,
	candleStore store.CandleStore,
	orderFlowStore store.OrderFlowStore,
	validator validate.Validator,
	m *metrics.CollectorMetrics,
	logger *slog.Logger,
) *Handlers {
	return &Handlers{
		candleStore:    candleStore,
		orderFlowStore: orderFlowStore,
		validator:      validator,
		metrics:        m,
		logger:         logger.With("module", "ws_handlers"),
		ctx:            ctx,
	}
}

type wsMessage struct {
	Topic   string          `json:"topic"`
	Type    string          `json:"type"`
	Data    json.RawMessage `json:"data"`
	Success *bool           `json:"success,omitempty"`
	RetMsg  string          `json:"ret_msg,omitempty"`
	Op      string          `json:"op,omitempty"`
}

func (h *Handlers) HandleMessage(msg []byte) {
	var raw wsMessage
	if err := json.Unmarshal(msg, &raw); err != nil {
		h.logger.Debug("parse failed", "error", err)
		return
	}

	if h.metrics != nil {
		topicLabel := raw.Topic
		if topicLabel == "" {
			topicLabel = raw.Op
		}
		h.metrics.MessagesTotal.WithLabelValues(topicLabel).Inc()
	}

	switch {
	case strings.HasPrefix(raw.Topic, "kline."):
		h.handleKline(raw.Data, raw.Topic)
	case strings.HasPrefix(raw.Topic, "tickers."):
		h.handleTicker(raw.Data, raw.Topic)
	case strings.HasPrefix(raw.Topic, "allLiquidation."):
		h.handleLiquidation(raw.Data, raw.Topic)
	case raw.Op == "pong":
	default:
		h.logger.Debug("unhandled message", "topic", raw.Topic, "op", raw.Op)
	}
}

func (h *Handlers) handleKline(data json.RawMessage, topic string) {
	var klines []mapper.WSKlineData
	if err := json.Unmarshal(data, &klines); err != nil {
		h.logger.Error("parse kline data", "error", err)
		return
	}

	symbol, timeframe := parseKlineTopic(topic)
	for _, k := range klines {
		if !k.Confirm {
			continue
		}
		candle := mapper.WSKlineToCandle(k, symbol, timeframe)
		if err := h.validator.ValidateCandle(&candle); err != nil {
			h.logger.Warn("candle validation failed", "error", err)
			continue
		}
		if err := h.candleStore.Insert(h.ctx, &candle); err != nil {
			h.logger.Error("store candle", "error", err, "symbol", candle.Symbol)
			if h.metrics != nil {
				h.metrics.ErrorsTotal.WithLabelValues("store_candle").Inc()
			}
			continue
		}
		if h.metrics != nil {
			h.metrics.StoredTotal.WithLabelValues("candles").Inc()
		}
	}
}

func (h *Handlers) handleTicker(data json.RawMessage, topic string) {
	var ticker mapper.WSTickerData
	if err := json.Unmarshal(data, &ticker); err != nil {
		h.logger.Error("parse ticker data", "error", err)
		return
	}

	now := time.Now()

	oi := mapper.TickerToOI(ticker, now)
	if err := h.validator.ValidateOI(&oi); err == nil {
		if err := h.orderFlowStore.InsertOpenInterest(h.ctx, &oi); err != nil {
			h.logger.Error("store OI", "error", err)
		} else {
			if h.metrics != nil {
				h.metrics.StoredTotal.WithLabelValues("open_interest").Inc()
			}
		}
	}

	fr := mapper.TickerToFundingRate(ticker, now)
	if err := h.validator.ValidateFunding(&fr); err == nil {
		if err := h.orderFlowStore.InsertFundingRate(h.ctx, &fr); err != nil {
			h.logger.Error("store funding rate", "error", err)
		} else {
			if h.metrics != nil {
				h.metrics.StoredTotal.WithLabelValues("funding_rate").Inc()
			}
		}
	}
}

func (h *Handlers) handleLiquidation(data json.RawMessage, topic string) {
	var liquidations []mapper.WSLiquidationData
	if err := json.Unmarshal(data, &liquidations); err != nil {
		h.logger.Error("parse liquidation data", "error", err)
		return
	}

	for _, liq := range liquidations {
		l := mapper.WSAllLiquidationToLiq(liq)
		if err := h.validator.ValidateLiquidation(&l); err != nil {
			h.logger.Warn("liquidation validation failed", "error", err)
			continue
		}
		if err := h.orderFlowStore.InsertLiquidation(h.ctx, &l); err != nil {
			h.logger.Error("store liquidation", "error", err)
			continue
		}
		if h.metrics != nil {
			h.metrics.StoredTotal.WithLabelValues("liquidations").Inc()
		}
	}
}

func parseKlineTopic(topic string) (symbol string, timeframe string) {
	parts := strings.Split(topic, ".")
	if len(parts) >= 3 {
		return parts[2], klineIntervalToTF(parts[1])
	}
	return "", ""
}

func klineIntervalToTF(interval string) string {
	switch interval {
	case "15":
		return "15m"
	case "60":
		return "1h"
	default:
		return interval
	}
}
