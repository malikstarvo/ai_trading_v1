package mock

import (
	"context"
	"sync"
	"time"

	"github.com/avav/ai_trading_v1/internal/model"
)

type MockCandleStore struct {
	mu           sync.Mutex
	candles      []model.Candle
	insertFn     func(ctx context.Context, candle *model.Candle) error
	insertBatchFn func(ctx context.Context, candles []model.Candle) error
	latestTimeFn func(ctx context.Context, symbol string, timeframe string) (time.Time, error)
}

func NewMockCandleStore() *MockCandleStore {
	return &MockCandleStore{
		candles: make([]model.Candle, 0),
	}
}

func (m *MockCandleStore) Insert(ctx context.Context, candle *model.Candle) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.candles = append(m.candles, *candle)
	if m.insertFn != nil {
		return m.insertFn(ctx, candle)
	}
	return nil
}

func (m *MockCandleStore) InsertBatch(ctx context.Context, candles []model.Candle) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.candles = append(m.candles, candles...)
	if m.insertBatchFn != nil {
		return m.insertBatchFn(ctx, candles)
	}
	return nil
}

func (m *MockCandleStore) LatestTime(ctx context.Context, symbol string, timeframe string) (time.Time, error) {
	if m.latestTimeFn != nil {
		return m.latestTimeFn(ctx, symbol, timeframe)
	}
	latest := time.Time{}
	for _, c := range m.candles {
		if c.Symbol == symbol && c.Timeframe == timeframe {
			if c.Time.After(latest) {
				latest = c.Time
			}
		}
	}
	return latest, nil
}

func (m *MockCandleStore) Candles() []model.Candle {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]model.Candle, len(m.candles))
	copy(result, m.candles)
	return result
}

func (m *MockCandleStore) SetInsertFn(fn func(ctx context.Context, candle *model.Candle) error) {
	m.insertFn = fn
}

func (m *MockCandleStore) SetLatestTimeFn(fn func(ctx context.Context, symbol string, timeframe string) (time.Time, error)) {
	m.latestTimeFn = fn
}

type MockOrderFlowStore struct {
	mu          sync.Mutex
	oiRecords   []model.OIRecord
	frRecords   []model.FundingRate
	lsRecords   []model.LSRatio
	liqRecords  []model.Liquidation
}

func NewMockOrderFlowStore() *MockOrderFlowStore {
	return &MockOrderFlowStore{}
}

func (m *MockOrderFlowStore) InsertOpenInterest(ctx context.Context, oi *model.OIRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.oiRecords = append(m.oiRecords, *oi)
	return nil
}

func (m *MockOrderFlowStore) InsertFundingRate(ctx context.Context, fr *model.FundingRate) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.frRecords = append(m.frRecords, *fr)
	return nil
}

func (m *MockOrderFlowStore) InsertLSRatio(ctx context.Context, ls *model.LSRatio) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lsRecords = append(m.lsRecords, *ls)
	return nil
}

func (m *MockOrderFlowStore) InsertLiquidation(ctx context.Context, liq *model.Liquidation) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.liqRecords = append(m.liqRecords, *liq)
	return nil
}

func (m *MockOrderFlowStore) OIRecords() []model.OIRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.oiRecords
}

func (m *MockOrderFlowStore) LiqRecords() []model.Liquidation {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.liqRecords
}

type MockCollectorHealthStore struct {
	statuses map[string]*model.CollectorHealth
	mu       sync.Mutex
}

func NewMockCollectorHealthStore() *MockCollectorHealthStore {
	return &MockCollectorHealthStore{
		statuses: make(map[string]*model.CollectorHealth),
	}
}

func (m *MockCollectorHealthStore) Heartbeat(ctx context.Context, serviceName string, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	m.statuses[serviceName] = &model.CollectorHealth{
		ServiceName: serviceName,
		Status:      status,
		UpdatedAt:   now,
	}
	return nil
}

func (m *MockCollectorHealthStore) RecordError(ctx context.Context, serviceName string, err error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	m.statuses[serviceName] = &model.CollectorHealth{
		ServiceName:  serviceName,
		Status:       "down",
		LastErrorAt:  &now,
		LastErrorMsg: err.Error(),
		UpdatedAt:    now,
	}
	return nil
}

func (m *MockCollectorHealthStore) GetStatus(ctx context.Context, serviceName string) (*model.CollectorHealth, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.statuses[serviceName], nil
}
