package backtest

import "math"

func CalculateMetrics(trades []Trade, initialCapital float64, totalBars, barsInMarket int, totalFeesPaid float64) Metrics {
	m := Metrics{
		TotalFeesPaid: totalFeesPaid,
	}

	if len(trades) == 0 {
		return m
	}

	m.TotalTrades = len(trades)
	m.TotalBars = totalBars

	var grossProfit, grossLoss float64
	var totalReturnPct, totalWinnerPct, totalLoserPct float64
	var largestWinner, largestLoser float64
	var totalHoldingBars int
	var returns []float64

	for _, t := range trades {
		returns = append(returns, t.ReturnPct)
		totalReturnPct += t.ReturnPct
		totalHoldingBars += t.HoldingBars
		m.NetPnL += t.PnL
		m.Expectancy += t.PnL

		if t.ReturnPct > 0 {
			m.WinningTrades++
			grossProfit += t.ReturnPct
			totalWinnerPct += t.ReturnPct
			if t.ReturnPct > largestWinner {
				largestWinner = t.ReturnPct
			}
		} else {
			m.LosingTrades++
			grossLoss += math.Abs(t.ReturnPct)
			totalLoserPct += t.ReturnPct
			if t.ReturnPct < largestLoser {
				largestLoser = t.ReturnPct
			}
		}
	}

	m.WinRate = float64(m.WinningTrades) / float64(m.TotalTrades) * 100
	m.Expectancy /= float64(m.TotalTrades)
	m.AvgReturnPct = totalReturnPct / float64(m.TotalTrades) * 100
	m.AvgHoldingBars = float64(totalHoldingBars) / float64(m.TotalTrades)

	if totalBars > 0 {
		m.ExposurePct = float64(barsInMarket) / float64(totalBars) * 100
	}

	if m.WinningTrades > 0 {
		m.AvgWinnerPct = totalWinnerPct / float64(m.WinningTrades) * 100
		m.LargestWinnerPct = largestWinner * 100
	}
	if m.LosingTrades > 0 {
		m.AvgLoserPct = totalLoserPct / float64(m.LosingTrades) * 100
		m.LargestLoserPct = largestLoser * 100
	}

	if grossLoss > 0 {
		m.ProfitFactor = grossProfit / grossLoss
	}

	if len(returns) > 1 {
		meanVal := mean(returns)
		std := stdDev(returns, meanVal)

		if std > 0 {
			barsPerYear := 35040.0
			sqrtBars := math.Sqrt(barsPerYear)
			m.SharpeRatio = meanVal / std * sqrtBars

			var downside []float64
			for _, r := range returns {
				if r < 0 {
					downside = append(downside, r)
				}
			}
			if len(downside) > 0 {
				downsideStd := stdDev(downside, mean(downside))
				if downsideStd > 0 {
					m.SortinoRatio = meanVal / downsideStd * sqrtBars
				}
			}
		}
	}

	m.MaxDrawdownPct = calcMaxDrawdown(trades, initialCapital)

	m.TradingDays, m.CalendarDays = calcTradeDistribution(trades)
	if m.CalendarDays > 0 {
		months := float64(m.CalendarDays) / 30.0
		m.TradesPerMonth = float64(m.TotalTrades) / months
		m.TradesPerYear = float64(m.TotalTrades) / months * 12
	}

	return m
}

func calcTradeDistribution(trades []Trade) (tradingDays, calendarDays int) {
	if len(trades) == 0 {
		return 0, 0
	}

	daySet := make(map[string]struct{})
	first := trades[0].EntryTime
	last := trades[0].EntryTime

	for _, t := range trades {
		day := t.EntryTime.Format("2006-01-02")
		daySet[day] = struct{}{}
		if t.EntryTime.Before(first) {
			first = t.EntryTime
		}
		if t.EntryTime.After(last) {
			last = t.EntryTime
		}
		if !t.ExitTime.IsZero() {
			exitDay := t.ExitTime.Format("2006-01-02")
			daySet[exitDay] = struct{}{}
			if t.ExitTime.After(last) {
				last = t.ExitTime
			}
		}
	}

	tradingDays = len(daySet)
	calendarDays = int(last.Sub(first).Hours()/24) + 1
	if calendarDays < 1 {
		calendarDays = 1
	}
	return tradingDays, calendarDays
}



func mean(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	s := 0.0
	for _, v := range vals {
		s += v
	}
	return s / float64(len(vals))
}

func stdDev(vals []float64, m float64) float64 {
	if len(vals) < 2 {
		return 0
	}
	var sumSq float64
	for _, v := range vals {
		d := v - m
		sumSq += d * d
	}
	return math.Sqrt(sumSq / float64(len(vals)-1))
}

func calcMaxDrawdown(trades []Trade, initialCapital float64) float64 {
	if len(trades) == 0 {
		return 0
	}

	cumulativePnL := 0.0
	peak := initialCapital
	maxDD := 0.0

	for _, t := range trades {
		cumulativePnL += t.PnL
		equity := initialCapital + cumulativePnL
		if equity > peak {
			peak = equity
		}
		if peak > 0 {
			dd := (peak - equity) / peak
			if dd > maxDD {
				maxDD = dd
			}
		}
	}

	return maxDD * 100
}
