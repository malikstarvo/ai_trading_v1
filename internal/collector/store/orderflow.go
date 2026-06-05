package store

import (
	"context"

	"github.com/avav/ai_trading_v1/internal/model"
)

type OrderFlowStore interface {
	InsertOpenInterest(ctx context.Context, oi *model.OIRecord) error
	InsertFundingRate(ctx context.Context, fr *model.FundingRate) error
	InsertLSRatio(ctx context.Context, ls *model.LSRatio) error
	InsertLiquidation(ctx context.Context, liq *model.Liquidation) error
}
