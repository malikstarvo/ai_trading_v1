package mapper

import (
	"time"

	"github.com/avav/ai_trading_v1/internal/collector/bybit"
	"github.com/avav/ai_trading_v1/internal/model"
)

func RestToOI(item bybit.OIItem, symbol string) model.OIRecord {
	return model.OIRecord{
		Time:   time.UnixMilli(item.Timestamp),
		Symbol: symbol,
		OI:     parseFloat(item.OpenInterest),
	}
}

func RestToLSRatio(item bybit.LSRatioItem, symbol, period string) model.LSRatio {
	return model.LSRatio{
		Time:      time.UnixMilli(item.Timestamp),
		Symbol:    symbol,
		Period:    period,
		BuyRatio:  parseFloat(item.BuyRatio),
		SellRatio: parseFloat(item.SellRatio),
	}
}

func RestToFundingRate(item bybit.FundingRateItem) model.FundingRate {
	return model.FundingRate{
		Time:   time.UnixMilli(item.FundingTime),
		Symbol: item.Symbol,
		Rate:   parseFloat(item.FundingRate),
	}
}
