package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ResearchStore struct {
	pool *pgxpool.Pool
}

func NewResearchStore(pool *pgxpool.Pool) *ResearchStore {
	return &ResearchStore{pool: pool}
}

func (s *ResearchStore) CreateRun(ctx context.Context, name, symbol, timeframe string, featureSetID int, config json.RawMessage) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO research_runs (name, symbol, timeframe, feature_set_id, config, status)
		VALUES ($1, $2, $3, $4, $5, 'pending')
		RETURNING id
	`, name, symbol, timeframe, featureSetID, config).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create research run: %w", err)
	}
	return id, nil
}

func (s *ResearchStore) StartRun(ctx context.Context, runID int64) error {
	now := time.Now()
	_, err := s.pool.Exec(ctx, `
		UPDATE research_runs SET status = 'running', started_at = $2 WHERE id = $1
	`, runID, now)
	return err
}

func (s *ResearchStore) CompleteRun(ctx context.Context, runID int64) error {
	now := time.Now()
	_, err := s.pool.Exec(ctx, `
		UPDATE research_runs SET status = 'completed', completed_at = $2 WHERE id = $1
	`, runID, now)
	return err
}

func (s *ResearchStore) FailRun(ctx context.Context, runID int64, errMsg string) error {
	now := time.Now()
	_, err := s.pool.Exec(ctx, `
		UPDATE research_runs SET status = 'failed', completed_at = $2 WHERE id = $1
	`, runID, now)
	if err != nil {
		return err
	}
	return fmt.Errorf("research run failed: %s", errMsg)
}

func (s *ResearchStore) SaveResult(ctx context.Context, runID int64, featureName, metricName string, value float64, samples int, metadata json.RawMessage) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO research_results (research_run_id, feature_name, metric_name, metric_value, samples, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, runID, featureName, metricName, value, samples, metadata)
	return err
}

func (s *ResearchStore) GetResults(ctx context.Context, runID int64) ([]ResearchResultRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, research_run_id, feature_name, metric_name, metric_value, samples, metadata, created_at
		FROM research_results
		WHERE research_run_id = $1
		ORDER BY metric_value DESC
	`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ResearchResultRow
	for rows.Next() {
		var r ResearchResultRow
		if err := rows.Scan(&r.ID, &r.ResearchRunID, &r.FeatureName, &r.MetricName, &r.MetricValue, &r.Samples, &r.Metadata, &r.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func (s *ResearchStore) SaveArtifact(ctx context.Context, runID int64, artifactType, filePath, description string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO research_artifacts (research_run_id, artifact_type, file_path, description)
		VALUES ($1, $2, $3, $4)
	`, runID, artifactType, filePath, description)
	return err
}
