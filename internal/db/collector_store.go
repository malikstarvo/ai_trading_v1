package db

import (
	"context"
	"fmt"
	"time"

	"github.com/avav/ai_trading_v1/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CandleStore struct {
	pool *pgxpool.Pool
}

func NewCandleStore(pool *pgxpool.Pool) *CandleStore {
	return &CandleStore{pool: pool}
}

func (s *CandleStore) Insert(ctx context.Context, candle *model.Candle) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO candles (time, symbol, timeframe, open, high, low, close, volume)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (time, symbol, timeframe) DO NOTHING
	`, candle.Time, candle.Symbol, candle.Timeframe, candle.Open, candle.High, candle.Low, candle.Close, candle.Volume)
	return err
}

func (s *CandleStore) InsertBatch(ctx context.Context, candles []model.Candle) error {
	if len(candles) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, c := range candles {
		batch.Queue(`
			INSERT INTO candles (time, symbol, timeframe, open, high, low, close, volume)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			ON CONFLICT (time, symbol, timeframe) DO NOTHING
		`, c.Time, c.Symbol, c.Timeframe, c.Open, c.High, c.Low, c.Close, c.Volume)
	}

	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range candles {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("batch insert candle: %w", err)
		}
	}
	return nil
}

func (s *CandleStore) LatestTime(ctx context.Context, symbol string, timeframe string) (time.Time, error) {
	var t time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT MAX(time)
		FROM candles WHERE symbol = $1 AND timeframe = $2
	`, symbol, timeframe).Scan(&t)
	return t, err
}

type OrderFlowStore struct {
	pool *pgxpool.Pool
}

func NewOrderFlowStore(pool *pgxpool.Pool) *OrderFlowStore {
	return &OrderFlowStore{pool: pool}
}

func (s *OrderFlowStore) InsertOpenInterest(ctx context.Context, oi *model.OIRecord) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO open_interest (time, symbol, oi, oi_value_usd)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (time, symbol) DO NOTHING
	`, oi.Time, oi.Symbol, oi.OI, oi.OIValueUSD)
	return err
}

func (s *OrderFlowStore) InsertFundingRate(ctx context.Context, fr *model.FundingRate) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO funding_rates (time, symbol, rate, interval_h)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (time, symbol) DO NOTHING
	`, fr.Time, fr.Symbol, fr.Rate, fr.IntervalH)
	return err
}

func (s *OrderFlowStore) InsertLSRatio(ctx context.Context, ls *model.LSRatio) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO ls_ratios (time, symbol, period, buy_ratio, sell_ratio)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (time, symbol, period) DO NOTHING
	`, ls.Time, ls.Symbol, ls.Period, ls.BuyRatio, ls.SellRatio)
	return err
}

func (s *OrderFlowStore) InsertLiquidation(ctx context.Context, liq *model.Liquidation) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO liquidations (time, symbol, side, size, price, value_usd)
		VALUES ($1,$2,$3,$4,$5,$6)
	`, liq.Time, liq.Symbol, liq.Side, liq.Size, liq.Price, liq.ValueUSD)
	return err
}

type CollectorHealthStore struct {
	pool *pgxpool.Pool
}

func NewCollectorHealthStore(pool *pgxpool.Pool) *CollectorHealthStore {
	return &CollectorHealthStore{pool: pool}
}

func (s *CollectorHealthStore) Heartbeat(ctx context.Context, serviceName, status string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO collector_health (service_name, status, last_success_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		ON CONFLICT (service_name) DO UPDATE SET
			status = $2, last_success_at = NOW(), updated_at = NOW()
	`, serviceName, status)
	return err
}

func (s *CollectorHealthStore) RecordError(ctx context.Context, serviceName string, err error) error {
	now := time.Now()
	_, e := s.pool.Exec(ctx, `
		INSERT INTO collector_health (service_name, status, last_error_at, last_error_msg, updated_at)
		VALUES ($1, 'down', $2, $3, $2)
		ON CONFLICT (service_name) DO UPDATE SET
			status = 'down', last_error_at = $2, last_error_msg = $3, updated_at = $2
	`, serviceName, now, err.Error())
	return e
}

func (s *CollectorHealthStore) GetStatus(ctx context.Context, serviceName string) (*model.CollectorHealth, error) {
	var h model.CollectorHealth
	err := s.pool.QueryRow(ctx, `
		SELECT service_name, status, last_success_at, last_error_at, last_error_msg, updated_at
		FROM collector_health WHERE service_name = $1
	`, serviceName).Scan(&h.ServiceName, &h.Status, &h.LastSuccessAt, &h.LastErrorAt, &h.LastErrorMsg, &h.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &h, nil
}
