package feature

import (
	"math"

	"github.com/avav/ai_trading_v1/internal/model"
)

func ComputeFeatures(candles []model.Candle, orderflow []model.OrderFlowSnapshot, featureSetID int) ([]model.FeatureRow, error) {
	n := len(candles)
	if n < 200 {
		return nil, nil
	}

	closes := extractFloats(candles, func(c model.Candle) float64 { return c.Close })
	highs := extractFloats(candles, func(c model.Candle) float64 { return c.High })
	lows := extractFloats(candles, func(c model.Candle) float64 { return c.Low })
	volumes := extractFloats(candles, func(c model.Candle) float64 { return c.Volume })

	ema20 := EMA(closes, 20)
	ema50 := EMA(closes, 50)
	ema200 := EMA(closes, 200)
	rsi14 := RSI(closes, 14)
	atr14 := ATR(highs, lows, closes, 14)
	adx14 := ADX(highs, lows, closes, 14)
	volEMA20 := EMA(volumes, 20)

	priceAboveEMA20 := PriceAboveEMA(closes, ema20)
	priceAboveEMA50 := PriceAboveEMA(closes, ema50)
	priceAboveEMA200 := PriceAboveEMA(closes, ema200)

	return1 := LogReturn(closes, 1)
	return4 := LogReturn(closes, 4)
	return12 := LogReturn(closes, 12)

	returnDiffs := make([]float64, n)
	for i := range return1 {
		if math.IsNaN(return1[i]) {
			returnDiffs[i] = math.NaN()
		} else {
			returnDiffs[i] = return1[i]
		}
	}
	volatility14 := Volatility(returnDiffs, 14)
	volatility50 := Volatility(returnDiffs, 50)

	rows := make([]model.FeatureRow, n)
	for i := 0; i < n; i++ {
		rows[i] = model.FeatureRow{
			Symbol:       candles[i].Symbol,
			Timeframe:    candles[i].Timeframe,
			Ts:           candles[i].Time,
			FeatureSetID: featureSetID,

			EMA20:            ema20[i],
			EMA50:            ema50[i],
			EMA200:           ema200[i],
			RSI14:            rsi14[i],
			ATR14:            atr14[i],
			ADX14:            adx14[i],
			VolumeEMA20:      volEMA20[i],
			PriceAboveEMA20:  priceAboveEMA20[i],
			PriceAboveEMA50:  priceAboveEMA50[i],
			PriceAboveEMA200: priceAboveEMA200[i],

			Return1:      return1[i],
			Return4:      return4[i],
			Return12:     return12[i],
			Volatility14: volatility14[i],
			Volatility50: volatility50[i],
		}
	}

	if len(orderflow) == n && n > 0 {
		oi := extractOFloats(orderflow, func(o model.OrderFlowSnapshot) float64 { return o.OI })
		funding := extractOFloats(orderflow, func(o model.OrderFlowSnapshot) float64 { return o.FundingRate })
		buyRatio := extractOFloats(orderflow, func(o model.OrderFlowSnapshot) float64 { return o.LSBuyRatio })
		sellRatio := extractOFloats(orderflow, func(o model.OrderFlowSnapshot) float64 { return o.LSSellRatio })
		liqLong := extractOFloats(orderflow, func(o model.OrderFlowSnapshot) float64 { return o.LiqLongUSD })
		liqShort := extractOFloats(orderflow, func(o model.OrderFlowSnapshot) float64 { return o.LiqShortUSD })

		oiDelta1 := OIDeltaPct(oi, 1)
		oiDelta4 := OIDeltaPct(oi, 4)
		oiDelta12 := OIDeltaPct(oi, 12)
		oiZ30 := ZScore(oi, 30)
		fundingZ30 := ZScore(funding, 30)
		lsRaw := LSRatioRaw(buyRatio, sellRatio)
		lsNorm := NormalizedImbalance(buyRatio, sellRatio)
		liqImb := NormalizedImbalance(liqLong, liqShort)

		for i := 0; i < n; i++ {
			rows[i].OIDelta1Pct = oiDelta1[i]
			rows[i].OIDelta4Pct = oiDelta4[i]
			rows[i].OIDelta12Pct = oiDelta12[i]
			rows[i].OIZScore30 = oiZ30[i]
			rows[i].FundingRate = funding[i]
			rows[i].FundingZScore30 = fundingZ30[i]
			rows[i].LSRatioRaw = lsRaw[i]
			rows[i].LSRatioNormalized = lsNorm[i]
			rows[i].LiqLongUSD = liqLong[i]
			rows[i].LiqShortUSD = liqShort[i]
			rows[i].LiqImbalance = liqImb[i]
		}
	}

	return rows, nil
}

func extractFloats(candles []model.Candle, fn func(model.Candle) float64) []float64 {
	result := make([]float64, len(candles))
	for i, c := range candles {
		result[i] = fn(c)
	}
	return result
}

func extractOFloats(snapshots []model.OrderFlowSnapshot, fn func(model.OrderFlowSnapshot) float64) []float64 {
	result := make([]float64, len(snapshots))
	for i, s := range snapshots {
		result[i] = fn(s)
	}
	return result
}
