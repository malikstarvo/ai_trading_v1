package papertrade

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/avav/ai_trading_v1/internal/agent/orderflow"
	"github.com/avav/ai_trading_v1/internal/agent/regime"
	"github.com/avav/ai_trading_v1/internal/agent/technical"
	"github.com/avav/ai_trading_v1/internal/agent/tradegate"
	"github.com/avav/ai_trading_v1/internal/db"
	"github.com/avav/ai_trading_v1/internal/model"
	"github.com/jackc/pgx/v5"
)

type PaperEngine struct {
	cfg    EngineConfig
	store  *PaperStore
	fStore *db.FeatureStore

	gate       *tradegate.Gate
	state      EngineState
	position   *Position
	lastBarTS  time.Time
	equity     float64
	balance    float64
	totalPnL   float64
	dayPnL     float64
	dayTrades  int
	currentDay string
	barCount   int
	startTime  time.Time
}

func New(cfg EngineConfig, store *PaperStore, fStore *db.FeatureStore) *PaperEngine {
	gCfg := cfg.GateConfig
	gCfg.StartingCapital = cfg.InitialCapital
	return &PaperEngine{
		cfg:       cfg,
		store:     store,
		fStore:    fStore,
		gate:      tradegate.New(gCfg),
		state:     StateRunning,
		equity:    cfg.InitialCapital,
		balance:   cfg.InitialCapital,
		startTime: time.Now(),
	}
}

func (e *PaperEngine) Run(ctx context.Context) error {
	log.Printf("[PaperEngine] Starting: %s %s, capital=%.0f, threshold=[long>=%.0f short<=%.0f]",
		e.cfg.Symbol, e.cfg.Timeframe, e.cfg.InitialCapital, e.cfg.LongThreshold, e.cfg.ShortThreshold)

	if err := e.recoverState(ctx); err != nil {
		log.Printf("[PaperEngine] Recovery: %v", err)
	}

	ticker := time.NewTicker(e.cfg.PollInterval)
	defer ticker.Stop()

	log.Println("[PaperEngine] Entering main loop")
	for {
		select {
		case <-ticker.C:
			e.processTick(ctx)
		case <-ctx.Done():
			log.Println("[PaperEngine] Shutting down")
			return nil
		}
	}
}

func (e *PaperEngine) processTick(ctx context.Context) {
	candle, err := e.loadLatestCandle(ctx)
	if err != nil || candle == nil {
		if err != nil {
			log.Printf("[PaperEngine] load candle: %v", err)
		}
		return
	}

	if candle.Time.Equal(e.lastBarTS) {
		return
	}
	isNewBar := !e.lastBarTS.IsZero()
	e.lastBarTS = candle.Time
	e.barCount++

	if isNewBar {
		e.gate.OnBar()
		if e.position != nil {
			e.position.BarsHeld++
			e.store.UpdatePositionBarsHeld(ctx, e.position.ID, e.position.BarsHeld)
		}
		day := candle.Time.Format("2006-01-02")
		if day != e.currentDay && !e.positionOpen() {
			e.resetDay(ctx, day)
		}
	}

	features, err := e.loadFeaturesAt(ctx, candle.Time)
	if err != nil || features == nil {
		if err != nil {
			log.Printf("[PaperEngine] load features: %v", err)
		}
		e.recordSnapshot(ctx, candle.Time)
		return
	}

	techScore := technical.Calculate(technical.Input{
		Price: candle.Close, EMA20: features.EMA20, EMA50: features.EMA50,
		EMA200: features.EMA200, RSI14: features.RSI14, ATR14: features.ATR14,
		ADX14: features.ADX14, Volume: candle.Volume, VolEMA20: features.VolumeEMA20,
	})
	ofScore := orderflow.Calculate(orderflow.Input{
		FundingRate: features.FundingRate, OIDeltaPct: features.OIDelta1Pct,
		LSRatio: features.LSRatioRaw, LongLiqUSD: features.LiqLongUSD,
		ShortLiqUSD: features.LiqShortUSD,
	})
	regScore := regime.Calculate(regime.Input{
		ADX14: features.ADX14, ATR14: features.ATR14,
		Price: candle.Close, Volatility: features.Volatility14,
	})

	dir := decideDirection(techScore.TechnicalScore, e.cfg.LongThreshold, e.cfg.ShortThreshold)

	if e.positionOpen() {
		if e.checkExit(ctx, candle, dir) {
			e.recordSnapshot(ctx, candle.Time)
			return
		}
		e.updateUnrealized(candle.Close)
		e.recordSnapshot(ctx, candle.Time)
		return
	}

	if !isNewBar || e.state == StateStopped || dir == NoTrade {
		e.recordSnapshot(ctx, candle.Time)
		return
	}

	gateInput := tradegate.Input{
		TechnicalScore: techScore.TechnicalScore,
		OrderFlowScore: ofScore.OrderFlowScore,
		RegimeScore:    regScore.RegimeScore,
		RegimeLabel:    regScore.Regime,
		MetaModelProb:  1.0,
	}
	gOut := e.gate.Evaluate(gateInput)
	if gOut.Decision == tradegate.NoTrade {
		e.recordSnapshot(ctx, candle.Time)
		return
	}

	atr := features.ATR14
	if atr <= 0 {
		atr = candle.Close * 0.01
	}
	positionSize := CalcPositionSize(e.equity, candle.Close, atr, e.cfg.ATRMultiplier, e.cfg.RiskPerTradePct)
	if positionSize <= 0 {
		e.recordSnapshot(ctx, candle.Time)
		return
	}
	positionSize *= gOut.SizeMultiplier

	entryResult := SimulateEntry(candle.Close, positionSize, dir, e.cfg, candle.Volume)
	e.openPosition(ctx, candle, entryResult, dir, gOut, techScore, ofScore, regScore, features)
}

func (e *PaperEngine) checkExit(ctx context.Context, candle *model.Candle, dir Direction) bool {
	if e.position == nil {
		return false
	}

	var exitPrice float64
	var reason string
	closed := false

	if (e.position.Direction == Long && candle.Low <= e.position.StopPrice) ||
		(e.position.Direction == Short && candle.High >= e.position.StopPrice) {
		exitPrice = e.position.StopPrice
		reason = string(ExitStopLoss)
		closed = true
	}

	if !closed && e.position.BarsHeld >= e.cfg.HoldingBars {
		switch e.position.Direction {
		case Long:
			exitPrice = candle.Close * (1 - e.cfg.Slippage)
		case Short:
			exitPrice = candle.Close * (1 + e.cfg.Slippage)
		}
		reason = string(ExitMaxHold)
		closed = true
	}

	if !closed && dir != NoTrade && dir != e.position.Direction {
		switch e.position.Direction {
		case Long:
			exitPrice = candle.Close * (1 - e.cfg.Slippage)
		case Short:
			exitPrice = candle.Close * (1 + e.cfg.Slippage)
		}
		reason = string(ExitOppositeSig)
		closed = true
	}

	if !closed {
		if e.position.Direction == Long {
			newStop := candle.Close * (1 - e.cfg.ATRMultiplier*candleATR(candle.Close, 0))
			if newStop > e.position.StopPrice {
				e.position.StopPrice = newStop
			}
		}
		return false
	}

	e.closePosition(ctx, candle, exitPrice, reason)
	return true
}

func (e *PaperEngine) openPosition(ctx context.Context, candle *model.Candle, entry FillResult, dir Direction, gOut tradegate.Output, techOut technical.Score, ofOut orderflow.Score, regOut regime.Score, features *model.FeatureRow) {
	entryReason := map[string]interface{}{
		"technical_score":  techOut.TechnicalScore,
		"orderflow_score":  ofOut.OrderFlowScore,
		"regime_score":     regOut.RegimeScore,
		"confidence_score": gOut.RawConfidence,
		"regime_label":     regOut.Regime,
		"gate_decision":    string(gOut.Decision),
		"direction":        string(dir),
	}
	entryReasonJSON, _ := json.Marshal(entryReason)

	snap := map[string]interface{}{
		"technical_score":  techOut.TechnicalScore,
		"orderflow_score":  ofOut.OrderFlowScore,
		"regime_score":     regOut.RegimeScore,
		"confidence_score": gOut.RawConfidence,
		"regime_label":     regOut.Regime,
		"atr14":            features.ATR14,
		"adx14":            features.ADX14,
	}
	snapJSON, _ := json.Marshal(snap)

	order := &Order{
		Symbol: e.cfg.Symbol, Timeframe: e.cfg.Timeframe, Direction: dir,
		Status: OrderFilled, RequestedSize: entry.Commission / e.cfg.Commission,
		FilledSize: entry.Commission / e.cfg.Commission, FillPrice: entry.FillPrice,
		SlippagePct: entry.SlippagePct, Commission: entry.Commission,
		Reason: string(entryReasonJSON), OpenTS: candle.Time,
	}
	if err := e.store.InsertOrder(ctx, order); err != nil {
		log.Printf("[PaperEngine] insert order: %v", err)
		return
	}

	fill := &Fill{
		OrderID: order.ID, TS: candle.Time,
		Side: func() string { if dir == Long { return "buy" }; return "sell" }(),
		Price: entry.FillPrice, Size: order.FilledSize, Fee: entry.Commission,
	}
	if err := e.store.InsertFill(ctx, fill); err != nil {
		log.Printf("[PaperEngine] insert fill: %v", err)
	}

	stopDistance := features.ATR14 * e.cfg.ATRMultiplier
	if features.ATR14 <= 0 {
		stopDistance = candle.Close * 0.01 * e.cfg.ATRMultiplier
	}
	var stopPrice float64
	switch dir {
	case Long:
		stopPrice = entry.FillPrice - stopDistance
	case Short:
		stopPrice = entry.FillPrice + stopDistance
	}

	pos := &Position{
		Symbol: e.cfg.Symbol, Timeframe: e.cfg.Timeframe, Direction: dir,
		EntryOrderID: order.ID, Quantity: order.FilledSize,
		EntryPrice: entry.FillPrice, EntryFee: entry.Commission,
		StopPrice: stopPrice, OpenTS: candle.Time,
	}
	if err := e.store.InsertPosition(ctx, pos); err != nil {
		log.Printf("[PaperEngine] insert position: %v", err)
		return
	}

	e.position = pos
	e.balance -= entry.Commission
	e.gate.OnTradePlaced()

	log.Printf("[PaperEngine] ENTRY %s %.4f size=%.2f stop=%.4f equity=%.0f snap=%s",
		dir, entry.FillPrice, order.FilledSize, stopPrice, e.equity, snapJSON)
}

func (e *PaperEngine) closePosition(ctx context.Context, candle *model.Candle, exitPrice float64, reason string) {
	if e.position == nil {
		return
	}
	pos := e.position
	exitResult := SimulateExit(exitPrice, pos.Quantity, pos.Direction, e.cfg)

	var grossPnL float64
	switch pos.Direction {
	case Long:
		grossPnL = (exitResult.FillPrice - pos.EntryPrice) / pos.EntryPrice * pos.Quantity
	case Short:
		grossPnL = (pos.EntryPrice - exitResult.FillPrice) / pos.EntryPrice * pos.Quantity
	}

	totalCommission := pos.EntryFee + exitResult.Commission
	netPnL := grossPnL - totalCommission
	returnPct := netPnL / pos.Quantity
	if math.IsNaN(netPnL) || math.IsInf(netPnL, 0) {
		netPnL = 0
		returnPct = 0
	}

	trade := &Trade{
		PositionID: pos.ID, Symbol: e.cfg.Symbol, Timeframe: e.cfg.Timeframe,
		Direction: pos.Direction, EntryTS: pos.OpenTS, ExitTS: candle.Time,
		EntryPrice: pos.EntryPrice, ExitPrice: exitResult.FillPrice,
		Size: pos.Quantity, GrossPnL: grossPnL,
		Commission: totalCommission, NetPnL: netPnL, ReturnPct: returnPct,
		HoldingBars: pos.BarsHeld, ExitReason: reason,
	}
	if err := e.store.InsertTrade(ctx, trade); err != nil {
		log.Printf("[PaperEngine] insert trade: %v", err)
	}

	e.store.ClosePosition(ctx, pos.ID)

	e.balance += netPnL + pos.Quantity
	e.totalPnL += netPnL
	e.dayPnL += netPnL
	e.dayTrades++

	e.gate.UpdatePnl(netPnL)
	e.gate.OnTradeClosed()

	ddPct := e.dayPnL / e.cfg.InitialCapital * 100
	if ddPct <= -e.cfg.MaxDailyDrawdownPct {
		e.state = StateStopped
		log.Printf("[PaperEngine] DAILY DRAWDOWN LIMIT: %.2f%%", ddPct)
	}

	totalDD := e.totalPnL / e.cfg.InitialCapital * 100
	if totalDD <= -e.cfg.MaxTotalDrawdownPct {
		e.state = StateStopped
		log.Printf("[PaperEngine] TOTAL DRAWDOWN LIMIT: %.2f%%", totalDD)
	}

	e.equity = e.balance
	e.position = nil

	log.Printf("[PaperEngine] EXIT %s reason=%s PnL=%.2f return=%.2f%% equity=%.0f",
		pos.Direction, reason, netPnL, returnPct*100, e.equity)
}

func (e *PaperEngine) updateUnrealized(price float64) {
	if e.position == nil {
		return
	}
	var upnl float64
	switch e.position.Direction {
	case Long:
		upnl = (price - e.position.EntryPrice) / e.position.EntryPrice * e.position.Quantity
	case Short:
		upnl = (e.position.EntryPrice - price) / e.position.EntryPrice * e.position.Quantity
	}
	e.equity = e.balance + upnl
}

func (e *PaperEngine) recordSnapshot(ctx context.Context, ts time.Time) {
	snap := &AccountSnapshot{
		TS: ts, Balance: e.balance, Equity: e.equity,
		UnrealizedPnL: e.equity - e.balance,
		DayPnL: e.dayPnL, DayTrades: e.dayTrades,
	}
	e.store.InsertSnapshot(ctx, snap)
}

func (e *PaperEngine) resetDay(ctx context.Context, day string) {
	dayPnL, dayTrades, err := e.store.LoadDailyStats(ctx, day)
	if err == nil {
		e.dayPnL = dayPnL
	}
	e.dayTrades = dayTrades
	e.currentDay = day
	e.gate.ResetDay(day, e.dayPnL)
	if e.state == StateStopped {
		e.state = StateRunning
		log.Printf("[PaperEngine] Day reset %s — resumed", day)
	}
}

func (e *PaperEngine) recoverState(ctx context.Context) error {
	pos, err := e.store.LoadOpenPosition(ctx, e.cfg.Symbol, e.cfg.Timeframe)
	if err != nil {
		return fmt.Errorf("load open position: %w", err)
	}
	if pos != nil {
		e.position = pos
		log.Printf("[PaperEngine] Recovered position #%d %s entry=%.2f stop=%.2f bars=%d",
			pos.ID, pos.Direction, pos.EntryPrice, pos.StopPrice, pos.BarsHeld)
	}

	totalPnL, err := e.store.LoadTotalPnL(ctx)
	if err == nil {
		e.totalPnL = totalPnL
		e.balance = e.cfg.InitialCapital + totalPnL
		e.equity = e.balance
	}

	day := time.Now().Format("2006-01-02")
	dayPnL, dayTrades, err := e.store.LoadDailyStats(ctx, day)
	if err == nil {
		e.dayPnL = dayPnL
		e.dayTrades = dayTrades
	}
	e.currentDay = day

	log.Printf("[PaperEngine] Recovery: balance=%.0f totalPnL=%.0f dayPnL=%.0f dayTrades=%d hasPosition=%v",
		e.balance, e.totalPnL, e.dayPnL, e.dayTrades, pos != nil)
	return nil
}

func (e *PaperEngine) loadLatestCandle(ctx context.Context) (*model.Candle, error) {
	var c model.Candle
	err := e.fStore.Pool().QueryRow(ctx, `
		SELECT time, symbol, timeframe, open, high, low, close, volume
		FROM candles
		WHERE symbol = $1 AND timeframe = $2
		ORDER BY time DESC
		LIMIT 1
	`, e.cfg.Symbol, e.cfg.Timeframe).Scan(&c.Time, &c.Symbol, &c.Timeframe, &c.Open, &c.High, &c.Low, &c.Close, &c.Volume)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (e *PaperEngine) loadFeaturesAt(ctx context.Context, ts time.Time) (*model.FeatureRow, error) {
	featureSetID, err := e.fStore.EnsureDefaultFeatureSet(ctx)
	if err != nil {
		return nil, err
	}
	lookback := time.Duration(e.cfg.HoldingBars*2) * 15 * time.Minute
	rows, err := e.fStore.LoadFeaturesAfter(ctx, e.cfg.Symbol, e.cfg.Timeframe, featureSetID, ts.Add(-lookback))
	if err != nil {
		return nil, err
	}
	var best *model.FeatureRow
	for i := range rows {
		if rows[i].Ts.After(ts) {
			continue
		}
		if best == nil || rows[i].Ts.After(best.Ts) {
			best = &rows[i]
		}
	}
	if best == nil {
		return nil, nil
	}
	staleness := ts.Sub(best.Ts)
	if staleness > 1*time.Hour {
		log.Printf("[PaperEngine] stale features: %v behind candle", staleness)
	}
	return best, nil
}

func (e *PaperEngine) positionOpen() bool {
	return e.position != nil
}

func candleATR(atr, _ float64) float64 {
	if atr <= 0 {
		return 0.01
	}
	return atr
}

type EngineStateResult struct {
	State      EngineState
	Equity     float64
	TotalPnL   float64
	DayPnL     float64
	DayTrades  int
	Position   *Position
	LastBarTS  time.Time
	BarCount   int
	Uptime     time.Duration
}

func (e *PaperEngine) State() EngineStateResult {
	return EngineStateResult{
		State: e.state, Equity: e.equity, TotalPnL: e.totalPnL,
		DayPnL: e.dayPnL, DayTrades: e.dayTrades,
		Position: e.position, LastBarTS: e.lastBarTS,
		BarCount: e.barCount, Uptime: time.Since(e.startTime),
	}
}
