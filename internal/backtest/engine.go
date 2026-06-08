package backtest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/avav/ai_trading_v1/internal/agent/orderflow"
	"github.com/avav/ai_trading_v1/internal/agent/regime"
	"github.com/avav/ai_trading_v1/internal/agent/technical"
	"github.com/avav/ai_trading_v1/internal/agent/tradegate"
	"github.com/avav/ai_trading_v1/internal/model"
)

type Engine struct {
	cfg Config
}

func New(cfg Config) *Engine {
	return &Engine{cfg: cfg}
}

func (e *Engine) Run(candles []model.Candle, features []model.FeatureRow) (*Summary, error) {
	if len(candles) != len(features) {
		return nil, fmt.Errorf("candle/feature length mismatch: %d vs %d", len(candles), len(features))
	}

	minBars := e.cfg.WarmupBars + e.cfg.HoldingBars + 2
	if len(candles) < minBars {
		return nil, fmt.Errorf("need >= %d bars (warmup=%d holding=%d), got %d",
			minBars, e.cfg.WarmupBars, e.cfg.HoldingBars, len(candles))
	}

	for i := range candles {
		if !candles[i].Time.Equal(features[i].Ts) {
			return nil, fmt.Errorf("feature/candle timestamp mismatch at bar %d: candle.Time=%v feature.Ts=%v",
				i, candles[i].Time, features[i].Ts)
		}
	}

	mlProbs := e.prefetchMLProbs(candles)

	gate := tradegate.New(e.cfg.GateConfig)
	totalLiveBars := len(candles) - e.cfg.WarmupBars
	equityCurve := make([]float64, 0, totalLiveBars)
	cumulativePnL := 0.0
	totalFees := 0.0

	var trades []Trade
	var position *Trade
	entryBar := -1
	pendingBar := -1
	var pending *pendingInfo
	barsInMarket := 0

	for i := e.cfg.WarmupBars; i < len(candles); i++ {
		c := candles[i]
		f := features[i]

		techOut := technical.Calculate(technical.Input{
			Price:    c.Close,
			EMA20:    f.EMA20,
			EMA50:    f.EMA50,
			EMA200:   f.EMA200,
			RSI14:    f.RSI14,
			ATR14:    f.ATR14,
			Volume:   c.Volume,
			VolEMA20: f.VolumeEMA20,
			ADX14:    f.ADX14,
		})

		ofOut := orderflow.Calculate(orderflow.Input{
			FundingRate:  f.FundingRate,
			OIDeltaPct:   f.OIDelta1Pct,
			LSRatio:      f.LSRatioRaw,
			LongLiqUSD:   f.LiqLongUSD,
			ShortLiqUSD:  f.LiqShortUSD,
		})

		regimeOut := regime.Calculate(regime.Input{
			ADX14:      f.ADX14,
			ATR14:      f.ATR14,
			Price:      c.Close,
			Volatility: f.Volatility14,
		})

		mlProb := 1.0
		if mlProbs != nil {
			if p, ok := mlProbs[c.Time]; ok {
				mlProb = p
			}
		}

		gateInput := tradegate.Input{
			TechnicalScore:    techOut.TechnicalScore,
			OrderFlowScore:    ofOut.OrderFlowScore,
			RegimeScore:       regimeOut.RegimeScore,
			RegimeLabel:       regimeOut.Regime,
			MetaModelProb:     mlProb,
		}
		decision := gate.Evaluate(gateInput)

		// Signal at bar i → decide direction and schedule entry at bar i+1
		dir := decideDir(techOut.TechnicalScore, e.cfg.LongThreshold, e.cfg.ShortThreshold)
		if decision.Decision != tradegate.NoTrade && position == nil && pendingBar < 0 && i < len(candles)-1 {
			// Direction filter
			switch e.cfg.Direction {
			case "long":
				if dir != DirLong {
					continue
				}
			case "short":
				if dir != DirShort {
					continue
				}
			}

			entryPrice := entryPriceForDir(candles[i+1].Open, dir, e.cfg.Slippage)
			atr := f.ATR14
			if atr <= 0 {
				atr = entryPrice * 0.01
			}
			stopPrice := stopPriceForDir(entryPrice, dir, atr, e.cfg.ATRMultiplier)

			sizeFrac := decision.SizeMultiplier
			if sizeFrac <= 0 {
				sizeFrac = 0.25
			}
			positionSize := e.cfg.InitialCapital * sizeFrac

			pendingBar = i + 1
			pending = &pendingInfo{
				entryBar:    pendingBar,
				entryPrice:  entryPrice,
				stopPrice:   stopPrice,
				size:        positionSize,
				direction:   dir,
				techScore:   techOut.TechnicalScore,
				ofScore:     ofOut.OrderFlowScore,
				regimeScore: regimeOut.RegimeScore,
				regimeLabel: regimeOut.Regime,
				confidence:  decision.RawConfidence,
			}
			gate.OnTradePlaced()
		}

		// Execute pending entry at bar i (if this is the scheduled bar)
		if position == nil && pending != nil && i == pending.entryBar {
			position = &Trade{
				EntryTime:    c.Time,
				Direction:    pending.direction,
				EntryPrice:   pending.entryPrice,
				Size:         pending.size,
				StopPrice:    pending.stopPrice,
				TechnicalScore: pending.techScore,
				OrderFlowScore: pending.ofScore,
				RegimeScore:    pending.regimeScore,
				Confidence:     pending.confidence,
				RegimeLabel:    pending.regimeLabel,
			}
			entryBar = i
			pendingBar = -1
			pending = nil
		}

		// Check exit for open position
		if position != nil {
			barsHeld := i - entryBar + 1
			position.HoldingBars = barsHeld

			closed := false
			// Check stop loss (direction-aware)
			if (position.Direction == DirLong && c.Low <= position.StopPrice) ||
				(position.Direction == DirShort && c.High >= position.StopPrice) {
				position.ExitPrice = position.StopPrice
				position.ExitTime = c.Time
				position.ExitReason = "stop"
				closed = true
			} else if barsHeld >= e.cfg.HoldingBars {
				// Exit at close when holding period reached (direction-aware)
				position.ExitPrice = exitPriceForDir(c.Close, position.Direction, e.cfg.Slippage)
				position.ExitTime = c.Time
				position.ExitReason = "expiry"
				closed = true
			}

			if closed {
				totalFees += closeTrade(position, e.cfg)
				trades = append(trades, *position)
				cumulativePnL += position.PnL
				barsInMarket += position.HoldingBars
				gate.UpdatePnl(position.PnL)
				gate.OnTradeClosed()
				position = nil
			}
		}

		gate.OnBar()

		// Record equity at end of bar (direction-aware unrealized PnL)
		equity := e.cfg.InitialCapital + cumulativePnL
		if position != nil {
			var upnl float64
			switch position.Direction {
			case DirLong:
				upnl = (c.Close - position.EntryPrice) / position.EntryPrice * position.Size
			case DirShort:
				upnl = (position.EntryPrice - c.Close) / position.EntryPrice * position.Size
			}
			equity += upnl
		}
		equityCurve = append(equityCurve, equity)
	}

	// Force close remaining position at last candle
	if position != nil {
		last := candles[len(candles)-1]
		position.ExitPrice = exitPriceForDir(last.Close, position.Direction, e.cfg.Slippage)
		position.ExitTime = last.Time
		position.ExitReason = "end_of_data"
		totalFees += closeTrade(position, e.cfg)
		trades = append(trades, *position)
		cumulativePnL += position.PnL
		barsInMarket += position.HoldingBars
	}

	metrics := CalculateMetrics(trades, e.cfg.InitialCapital, totalLiveBars, barsInMarket, totalFees)

	return &Summary{
		Config:      e.cfg,
		Trades:      trades,
		Metrics:     metrics,
		EquityCurve: equityCurve,
	}, nil
}

func closeTrade(t *Trade, cfg Config) float64 {
	if t.ExitPrice <= 0 || t.EntryPrice <= 0 {
		return 0
	}
	var grossPnl float64
	switch t.Direction {
	case DirLong:
		grossPnl = (t.ExitPrice - t.EntryPrice) / t.EntryPrice * t.Size
	case DirShort:
		grossPnl = (t.EntryPrice - t.ExitPrice) / t.EntryPrice * t.Size
	}
	entryFee := t.Size * cfg.Commission
	exitFee := t.ExitPrice / t.EntryPrice * t.Size * cfg.Commission
	fee := entryFee + exitFee
	t.PnL = grossPnl - fee
	t.ReturnPct = t.PnL / t.Size
	if math.IsNaN(t.PnL) || math.IsInf(t.PnL, 0) {
		t.PnL = 0
		t.ReturnPct = 0
		fee = 0
	}
	return fee
}

func entryPriceForDir(refPrice float64, dir Direction, slippage float64) float64 {
	switch dir {
	case DirLong:
		return refPrice * (1 + slippage)
	case DirShort:
		return refPrice * (1 - slippage)
	default:
		return refPrice * (1 + slippage)
	}
}

func exitPriceForDir(refPrice float64, dir Direction, slippage float64) float64 {
	switch dir {
	case DirLong:
		return refPrice * (1 - slippage)
	case DirShort:
		return refPrice * (1 + slippage)
	default:
		return refPrice * (1 - slippage)
	}
}

func stopPriceForDir(entryPrice float64, dir Direction, atr float64, atrMultiplier float64) float64 {
	switch dir {
	case DirLong:
		return entryPrice * (1 - atrMultiplier*atr/entryPrice)
	case DirShort:
		return entryPrice * (1 + atrMultiplier*atr/entryPrice)
	default:
		return entryPrice * (1 - atrMultiplier*atr/entryPrice)
	}
}

func (e *Engine) prefetchMLProbs(candles []model.Candle) map[time.Time]float64 {
	if e.cfg.MLAPIURL == "" {
		return nil
	}

	// Collect live bar timestamps
	tss := make([]string, 0, len(candles)-e.cfg.WarmupBars)
	for i := e.cfg.WarmupBars; i < len(candles); i++ {
		tss = append(tss, candles[i].Time.Format(time.RFC3339))
	}

	body := map[string]interface{}{
		"timestamps": tss,
		"horizon":    4,
		"symbol":     e.cfg.Symbol,
		"timeframe":  e.cfg.Timeframe,
	}
	data, _ := json.Marshal(body)

	url := fmt.Sprintf("%s/api/model/predict-batch", e.cfg.MLAPIURL)
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		log.Printf("[Backtest] ML batch predict error: %v", err)
		return nil
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[Backtest] ML batch read error: %v", err)
		return nil
	}

	var result struct {
		Results []struct {
			TS   string  `json:"ts"`
			Prob float64 `json:"prob"`
		} `json:"results"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Printf("[Backtest] ML batch parse error: %v", err)
		return nil
	}

	probs := make(map[time.Time]float64, len(result.Results))
	for _, r := range result.Results {
		ts, err := time.Parse(time.RFC3339, r.TS)
		if err == nil {
			probs[ts] = r.Prob
		}
	}
	return probs
}
