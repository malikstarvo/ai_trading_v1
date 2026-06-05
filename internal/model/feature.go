package model

import "time"

type FeatureRow struct {
	Symbol       string    `db:"symbol"`
	Timeframe    string    `db:"timeframe"`
	Ts           time.Time `db:"ts"`
	FeatureSetID int       `db:"feature_set_id"`

	EMA20            float64 `db:"ema20"`
	EMA50            float64 `db:"ema50"`
	EMA200           float64 `db:"ema200"`
	RSI14            float64 `db:"rsi14"`
	ATR14            float64 `db:"atr14"`
	ADX14            float64 `db:"adx14"`
	VolumeEMA20      float64 `db:"volume_ema20"`
	PriceAboveEMA20  int8    `db:"price_above_ema20"`
	PriceAboveEMA50  int8    `db:"price_above_ema50"`
	PriceAboveEMA200 int8    `db:"price_above_ema200"`

	OIDelta1Pct       float64 `db:"oi_delta_1_pct"`
	OIDelta4Pct       float64 `db:"oi_delta_4_pct"`
	OIDelta12Pct      float64 `db:"oi_delta_12_pct"`
	OIZScore30        float64 `db:"oi_zscore_30"`
	FundingRate       float64 `db:"funding_rate"`
	FundingZScore30   float64 `db:"funding_zscore_30"`
	LSRatioRaw        float64 `db:"ls_ratio_raw"`
	LSRatioNormalized float64 `db:"ls_ratio_normalized"`
	LiqLongUSD        float64 `db:"liq_long_usd"`
	LiqShortUSD       float64 `db:"liq_short_usd"`
	LiqImbalance      float64 `db:"liq_imbalance"`

	Return1      float64 `db:"return_1"`
	Return4      float64 `db:"return_4"`
	Return12     float64 `db:"return_12"`
	Volatility14 float64 `db:"volatility_14"`
	Volatility50 float64 `db:"volatility_50"`
}

// OrderFlowSnapshot aligns order flow data to candle timestamps.
type OrderFlowSnapshot struct {
	Ts          time.Time
	OI          float64
	FundingRate float64
	LSBuyRatio  float64
	LSSellRatio float64
	LiqLongUSD  float64
	LiqShortUSD float64
}
