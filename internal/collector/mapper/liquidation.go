package mapper

import (
	"time"

	"github.com/avav/ai_trading_v1/internal/model"
)

type WSLiquidationData struct {
	Time int64  `json:"T"`
	Sym  string `json:"s"`
	Side string `json:"S"`
	Size string `json:"v"`
	Price string `json:"p"`
}

func WSAllLiquidationToLiq(item WSLiquidationData) model.Liquidation {
	price := parseFloat(item.Price)
	size := parseFloat(item.Size)
	return model.Liquidation{
		Time:     time.UnixMilli(item.Time),
		Symbol:   item.Sym,
		Side:     item.Side,
		Size:     size,
		Price:    price,
		ValueUSD: price * size,
	}
}
