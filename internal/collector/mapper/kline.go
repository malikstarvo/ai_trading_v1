package mapper

import (
	"strconv"
	"time"

	"github.com/avav/ai_trading_v1/internal/collector/bybit"
	"github.com/avav/ai_trading_v1/internal/model"
)

type WSKlineData struct {
	Start     int64  `json:"start"`
	End       int64  `json:"end"`
	Interval  string `json:"interval"`
	Open      string `json:"open"`
	High      string `json:"high"`
	Low       string `json:"low"`
	Close     string `json:"close"`
	Volume    string `json:"volume"`
	Turnover  string `json:"turnover"`
	Confirm   bool   `json:"confirm"`
	Timestamp int64  `json:"timestamp"`
}

func WSKlineToCandle(item WSKlineData, symbol string, timeframe string) model.Candle {
	return model.Candle{
		Time:      time.UnixMilli(item.Start),
		Symbol:    symbol,
		Timeframe: timeframe,
		Open:      parseFloat(item.Open),
		High:      parseFloat(item.High),
		Low:       parseFloat(item.Low),
		Close:     parseFloat(item.Close),
		Volume:    parseFloat(item.Volume),
	}
}

func RestKlineToCandle(item bybit.KlineItem, symbol string, timeframe string) model.Candle {
	start := parseInt64(item[0])
	return model.Candle{
		Time:      time.UnixMilli(start),
		Symbol:    symbol,
		Timeframe: timeframe,
		Open:      parseFloat(item[1]),
		High:      parseFloat(item[2]),
		Low:       parseFloat(item[3]),
		Close:     parseFloat(item[4]),
		Volume:    parseFloat(item[5]),
	}
}

func RestKlinesToCandles(resp *bybit.KlineResponse, symbol string, timeframe string) []model.Candle {
	candles := make([]model.Candle, 0, len(resp.List))
	for _, item := range resp.List {
		candles = append(candles, RestKlineToCandle(item, symbol, timeframe))
	}
	return candles
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func parseInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}
