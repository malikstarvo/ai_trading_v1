package model

import "time"

type Candle struct {
	Time      time.Time `db:"time"`
	Symbol    string    `db:"symbol"`
	Timeframe string    `db:"timeframe"`
	Open      float64   `db:"open"`
	High      float64   `db:"high"`
	Low       float64   `db:"low"`
	Close     float64   `db:"close"`
	Volume    float64   `db:"volume"`
}

type Timeframe string

const (
	Timeframe15m Timeframe = "15m"
	Timeframe1h  Timeframe = "1h"
)

func (t Timeframe) BybitInterval() string {
	switch t {
	case Timeframe15m:
		return "15"
	case Timeframe1h:
		return "60"
	default:
		return string(t)
	}
}

func (t Timeframe) Duration() time.Duration {
	switch t {
	case Timeframe15m:
		return 15 * time.Minute
	case Timeframe1h:
		return time.Hour
	default:
		return 0
	}
}
