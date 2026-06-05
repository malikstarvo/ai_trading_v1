package mapper

import (
	"time"

	"github.com/avav/ai_trading_v1/internal/model"
)

type WSTickerData struct {
	Symbol            string `json:"symbol"`
	LastPrice         string `json:"lastPrice"`
	IndexPrice        string `json:"indexPrice"`
	MarkPrice         string `json:"markPrice"`
	OpenInterest      string `json:"openInterest"`
	OpenInterestValue string `json:"openInterestValue"`
	FundingRate       string `json:"fundingRate"`
	NextFundingTime   string `json:"nextFundingTime"`
	Volume24h         string `json:"volume24h"`
	Turnover24h       string `json:"turnover24h"`
}

func TickerToOI(data WSTickerData, ts time.Time) model.OIRecord {
	return model.OIRecord{
		Time:       ts,
		Symbol:     data.Symbol,
		OI:         parseFloat(data.OpenInterest),
		OIValueUSD: parseFloat(data.OpenInterestValue),
	}
}

func TickerToFundingRate(data WSTickerData, ts time.Time) model.FundingRate {
	return model.FundingRate{
		Time:   ts,
		Symbol: data.Symbol,
		Rate:   parseFloat(data.FundingRate),
	}
}
