package papertrade

import (
	"math"
)

type FillResult struct {
	FillPrice  float64
	SlippagePct float64
	Commission float64
}

// SimulateEntry calculates the fill price, slippage, and commission for an entry order.
// For long: buy at refPrice * (1 + slippage) — worse price (higher).
// For short: sell at refPrice * (1 - slippage) — worse price (lower).
func SimulateEntry(refPrice float64, size float64, direction Direction, cfg EngineConfig, candleVolume float64) FillResult {
	slippage := cfg.Slippage
	// Volume premium: if size > 1% of candle volume, add extra slippage
	if candleVolume > 0 && size > candleVolume*0.01 {
		ratio := size / (candleVolume * 0.01)
		slippage += cfg.Slippage * math.Min(ratio, 2.0)
	}

	var fillPrice float64
	switch direction {
	case Long:
		fillPrice = refPrice * (1 + slippage)
	case Short:
		fillPrice = refPrice * (1 - slippage)
	default:
		fillPrice = refPrice
	}

	commission := size * cfg.Commission
	return FillResult{
		FillPrice:   fillPrice,
		SlippagePct: slippage,
		Commission:  commission,
	}
}

// SimulateExit calculates the fill price and commission for an exit.
// For long: sell at refPrice * (1 - slippage).
// For short: buy at refPrice * (1 + slippage).
func SimulateExit(refPrice float64, size float64, direction Direction, cfg EngineConfig) FillResult {
	var fillPrice float64
	switch direction {
	case Long:
		fillPrice = refPrice * (1 - cfg.Slippage)
	case Short:
		fillPrice = refPrice * (1 + cfg.Slippage)
	default:
		fillPrice = refPrice
	}

	commission := size * cfg.Commission
	return FillResult{
		FillPrice:   fillPrice,
		SlippagePct: cfg.Slippage,
		Commission:  commission,
	}
}

// CalcPositionSize calculates position size based on fixed-fraction risk.
// positionSize = equity * riskPct / stopDistance
// where stopDistance = entryPrice * ATRMultiplier * (atr / entryPrice) = ATRMultiplier * atr
func CalcPositionSize(equity float64, entryPrice float64, atr float64, atrMultiplier float64, riskPct float64) float64 {
	if atr <= 0 || equity <= 0 || entryPrice <= 0 {
		return 0
	}
	stopDistance := atr * atrMultiplier
	if stopDistance <= 0 {
		return 0
	}
	size := equity * (riskPct / 100.0) / stopDistance
	if size <= 0 {
		return 0
	}
	return size
}

func decideDirection(techScore float64, longThreshold, shortThreshold float64) Direction {
	if techScore >= longThreshold {
		return Long
	}
	if techScore <= shortThreshold {
		return Short
	}
	return NoTrade
}
