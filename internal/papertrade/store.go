package papertrade

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaperStore struct {
	pool *pgxpool.Pool
}

func NewPaperStore(pool *pgxpool.Pool) *PaperStore {
	return &PaperStore{pool: pool}
}

func (s *PaperStore) InsertOrder(ctx context.Context, o *Order) error {
	return s.pool.QueryRow(ctx, `
		INSERT INTO paper_orders (symbol, timeframe, direction, status, requested_size, filled_size, fill_price, slippage_pct, commission, reason, open_ts)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at
	`, o.Symbol, o.Timeframe, string(o.Direction), string(o.Status), o.RequestedSize, o.FilledSize, o.FillPrice, o.SlippagePct, o.Commission, o.Reason, o.OpenTS).Scan(&o.ID, &o.CreatedAt)
}

func (s *PaperStore) UpdateOrderFill(ctx context.Context, orderID int64, fillPrice, filledSize, commission float64, slippagePct float64) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE paper_orders SET status = 'filled', fill_price = $2, filled_size = $3, commission = $4, slippage_pct = $5, updated_at = NOW()
		WHERE id = $1
	`, orderID, fillPrice, filledSize, commission, slippagePct)
	return err
}

func (s *PaperStore) InsertFill(ctx context.Context, fill *Fill) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO paper_fills (order_id, ts, side, price, size, fee)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, fill.OrderID, fill.TS, fill.Side, fill.Price, fill.Size, fill.Fee)
	return err
}

func (s *PaperStore) InsertPosition(ctx context.Context, pos *Position) error {
	return s.pool.QueryRow(ctx, `
		INSERT INTO paper_positions (symbol, timeframe, direction, entry_order_id, quantity, entry_price, entry_fee, stop_price, open_ts)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`, pos.Symbol, pos.Timeframe, string(pos.Direction), pos.EntryOrderID, pos.Quantity, pos.EntryPrice, pos.EntryFee, pos.StopPrice, pos.OpenTS).Scan(&pos.ID)
}

func (s *PaperStore) ClosePosition(ctx context.Context, positionID int64) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE paper_positions SET status = 'closed' WHERE id = $1
	`, positionID)
	return err
}

func (s *PaperStore) UpdatePositionBarsHeld(ctx context.Context, positionID int64, barsHeld int) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE paper_positions SET bars_held = $2 WHERE id = $1
	`, positionID, barsHeld)
	return err
}

func (s *PaperStore) InsertTrade(ctx context.Context, t *Trade) error {
	return s.pool.QueryRow(ctx, `
		INSERT INTO paper_trades (position_id, symbol, timeframe, direction, entry_ts, exit_ts, entry_price, exit_price, size, gross_pnl, commission, net_pnl, return_pct, holding_bars, exit_reason, entry_reason, feature_snapshot)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id
	`, t.PositionID, t.Symbol, t.Timeframe, string(t.Direction), t.EntryTS, t.ExitTS, t.EntryPrice, t.ExitPrice, t.Size, t.GrossPnL, t.Commission, t.NetPnL, t.ReturnPct, t.HoldingBars, t.ExitReason, t.EntryReason, t.FeatureSnapshot).Scan(&t.ID)
}

func (s *PaperStore) InsertSnapshot(ctx context.Context, snap *AccountSnapshot) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO paper_account_snapshots (ts, balance, equity, unrealized_pnl, day_pnl, day_trades)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, snap.TS, snap.Balance, snap.Equity, snap.UnrealizedPnL, snap.DayPnL, snap.DayTrades)
	return err
}

func (s *PaperStore) LoadOpenPosition(ctx context.Context, symbol, timeframe string) (*Position, error) {
	pos := &Position{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, symbol, timeframe, direction, entry_order_id, quantity, entry_price, entry_fee, stop_price, open_ts, bars_held, status
		FROM paper_positions
		WHERE symbol = $1 AND timeframe = $2 AND status = 'open'
		ORDER BY id DESC LIMIT 1
	`, symbol, timeframe).Scan(
		&pos.ID, &pos.Symbol, &pos.Timeframe, &pos.Direction, &pos.EntryOrderID,
		&pos.Quantity, &pos.EntryPrice, &pos.EntryFee, &pos.StopPrice, &pos.OpenTS,
		&pos.BarsHeld, &pos.Status,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return pos, err
}

func (s *PaperStore) LoadDailyStats(ctx context.Context, day string) (dayPnL float64, dayTrades int, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(net_pnl), 0), COUNT(*)
		FROM paper_trades
		WHERE entry_ts::date = $1::date
	`, day).Scan(&dayPnL, &dayTrades)
	return
}

func (s *PaperStore) LoadTotalPnL(ctx context.Context) (float64, error) {
	var total float64
	err := s.pool.QueryRow(ctx, `SELECT COALESCE(SUM(net_pnl), 0) FROM paper_trades`).Scan(&total)
	return total, err
}

func (s *PaperStore) FeatureSnapshotJSON(techScore, ofScore, regimeScore, confScore, regimeLabel string, atr14, adx14 float64) string {
	snap := AgentSnapshot{
		TechnicalScore:  parseFloat(techScore),
		OrderFlowScore:  parseFloat(ofScore),
		RegimeScore:     parseFloat(regimeScore),
		ConfidenceScore: parseFloat(confScore),
		RegimeLabel:     regimeLabel,
		ATR14:           atr14,
		ADX14:           adx14,
	}
	b, _ := json.Marshal(snap)
	return string(b)
}

func parseFloat(s string) float64 {
	var v float64
	fmt.Sscanf(s, "%f", &v)
	return v
}
