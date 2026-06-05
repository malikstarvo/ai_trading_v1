package store

import (
	"context"
	"time"

	"github.com/avav/ai_trading_v1/internal/model"
)

type CandleStore interface {
	Insert(ctx context.Context, candle *model.Candle) error
	InsertBatch(ctx context.Context, candles []model.Candle) error
	LatestTime(ctx context.Context, symbol string, timeframe string) (time.Time, error)
}
