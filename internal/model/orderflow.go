package model

import "time"

type OIRecord struct {
	Time       time.Time `db:"time"`
	Symbol     string    `db:"symbol"`
	OI         float64   `db:"oi"`
	OIValueUSD float64   `db:"oi_value_usd"`
}

type FundingRate struct {
	Time      time.Time `db:"time"`
	Symbol    string    `db:"symbol"`
	Rate      float64   `db:"rate"`
	IntervalH int       `db:"interval_h"`
}

type LSRatio struct {
	Time      time.Time `db:"time"`
	Symbol    string    `db:"symbol"`
	Period    string    `db:"period"`
	BuyRatio  float64   `db:"buy_ratio"`
	SellRatio float64   `db:"sell_ratio"`
}

type Liquidation struct {
	Time     time.Time `db:"time"`
	Symbol   string    `db:"symbol"`
	Side     string    `db:"side"`
	Size     float64   `db:"size"`
	Price    float64   `db:"price"`
	ValueUSD float64   `db:"value_usd"`
}
