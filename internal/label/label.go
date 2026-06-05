package label

import (
	"github.com/avav/ai_trading_v1/internal/model"
)

const defaultSuccessThreshold = 0.25

func ComputeLabels(candles []model.Candle, horizons []int, successThresholdPct float64) ([]model.LabelRow, error) {
	if len(candles) == 0 {
		return nil, nil
	}
	if successThresholdPct <= 0 {
		successThresholdPct = defaultSuccessThreshold
	}

	n := len(candles)
	closes := make([]float64, n)
	for i, c := range candles {
		closes[i] = c.Close
	}

	rows := make([]model.LabelRow, n)
	for i := 0; i < n; i++ {
		rows[i] = model.LabelRow{
			Symbol:    candles[i].Symbol,
			Timeframe: candles[i].Timeframe,
			Ts:        candles[i].Time,
		}
		for _, h := range horizons {
			if i+h >= n {
				continue
			}
			futureRet := (closes[i+h] - closes[i]) / closes[i] * 100
			var success int8
			if futureRet > successThresholdPct {
				success = 1
			}
			switch h {
			case 4:
				rows[i].FutureReturn4 = futureRet
				rows[i].Success4 = success
			case 12:
				rows[i].FutureReturn12 = futureRet
				rows[i].Success12 = success
			case 24:
				rows[i].FutureReturn24 = futureRet
				rows[i].Success24 = success
			}
		}
	}
	return rows, nil
}
