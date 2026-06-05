package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type JobStore struct {
	pool *pgxpool.Pool
}

func NewJobStore(pool *pgxpool.Pool) *JobStore {
	return &JobStore{pool: pool}
}

type JobParams struct {
	LabelThreshold  float64 `json:"label_threshold"`
	Horizons        []int   `json:"horizons"`
	MinCandles      int     `json:"min_candles"`
}

func (s *JobStore) Create(ctx context.Context, symbol, timeframe string, featureSetID int, params JobParams) (int64, error) {
	raw, _ := json.Marshal(params)
	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO feature_jobs (symbol, timeframe, feature_set_id, params, status)
		VALUES ($1, $2, $3, $4, 'pending')
		RETURNING id
	`, symbol, timeframe, featureSetID, raw).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create job: %w", err)
	}
	return id, nil
}

func (s *JobStore) Start(ctx context.Context, jobID int64) error {
	now := time.Now()
	_, err := s.pool.Exec(ctx, `
		UPDATE feature_jobs SET status = 'running', started_at = $2, last_heartbeat = $2
		WHERE id = $1
	`, jobID, now)
	return err
}

func (s *JobStore) Heartbeat(ctx context.Context, jobID int64) error {
	_, err := s.pool.Exec(ctx, `UPDATE feature_jobs SET last_heartbeat = NOW() WHERE id = $1`, jobID)
	return err
}

func (s *JobStore) UpdateProgress(ctx context.Context, jobID int64, processedRows int) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE feature_jobs SET processed_rows = $2, last_heartbeat = NOW()
		WHERE id = $1
	`, jobID, processedRows)
	return err
}

func (s *JobStore) Complete(ctx context.Context, jobID int64, processedRows int) error {
	now := time.Now()
	_, err := s.pool.Exec(ctx, `
		UPDATE feature_jobs
		SET status = 'completed', processed_rows = $2, completed_at = $3, last_heartbeat = $3
		WHERE id = $1
	`, jobID, processedRows, now)
	return err
}

func (s *JobStore) Fail(ctx context.Context, jobID int64, errMsg string) error {
	now := time.Now()
	_, err := s.pool.Exec(ctx, `
		UPDATE feature_jobs
		SET status = 'failed', error = $2, completed_at = $3, last_heartbeat = $3
		WHERE id = $1
	`, jobID, errMsg, now)
	return err
}
