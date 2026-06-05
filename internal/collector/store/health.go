package store

import (
	"context"
	"time"

	"github.com/avav/ai_trading_v1/internal/model"
)

type CollectorHealthStore interface {
	Heartbeat(ctx context.Context, serviceName string, status string) error
	RecordError(ctx context.Context, serviceName string, err error) error
	GetStatus(ctx context.Context, serviceName string) (*model.CollectorHealth, error)
}

type CollectorHealthStoreImpl struct {
	pool interface {
		Exec(ctx context.Context, sql string, args ...interface{}) error
	}
}

func (h *CollectorHealthStoreImpl) Heartbeat(ctx context.Context, serviceName, status string) error {
	return h.pool.Exec(ctx, `
		INSERT INTO collector_health (service_name, status, last_success_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		ON CONFLICT (service_name) DO UPDATE SET
			status = $2,
			last_success_at = NOW(),
			updated_at = NOW()
	`, serviceName, status)
}

func (h *CollectorHealthStoreImpl) RecordError(ctx context.Context, serviceName string, err error) error {
	now := time.Now()
	return h.pool.Exec(ctx, `
		INSERT INTO collector_health (service_name, status, last_error_at, last_error_msg, updated_at)
		VALUES ($1, 'down', $2, $3, $2)
		ON CONFLICT (service_name) DO UPDATE SET
			status = 'down',
			last_error_at = $2,
			last_error_msg = $3,
			updated_at = $2
	`, serviceName, now, err.Error())
}

func (h *CollectorHealthStoreImpl) GetStatus(ctx context.Context, serviceName string) (*model.CollectorHealth, error) {
	var health model.CollectorHealth
	err := h.pool.Exec(ctx, `
		SELECT service_name, status, last_success_at, last_error_at, last_error_msg, updated_at
		FROM collector_health WHERE service_name = $1
	`, serviceName)
	if err != nil {
		return nil, err
	}
	return &health, nil
}
