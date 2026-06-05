package model

import "time"

type LabelRow struct {
	Symbol    string    `db:"symbol"`
	Timeframe string    `db:"timeframe"`
	Ts        time.Time `db:"ts"`

	FutureReturn4  float64 `db:"future_return_4"`
	FutureReturn12 float64 `db:"future_return_12"`
	FutureReturn24 float64 `db:"future_return_24"`
	Success4       int8    `db:"success_4"`
	Success12      int8    `db:"success_12"`
	Success24      int8    `db:"success_24"`
}
