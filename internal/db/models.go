package db

import (
	"encoding/json"
	"time"
)

type FeatureJobRow struct {
	ID             int64      `db:"id"`
	Symbol         string     `db:"symbol"`
	Timeframe      string     `db:"timeframe"`
	FeatureSetID   int        `db:"feature_set_id"`
	Params         []byte     `db:"params"`
	Status         string     `db:"status"`
	ProcessedRows  int        `db:"processed_rows"`
	Error          *string    `db:"error"`
	StartedAt      *time.Time `db:"started_at"`
	CompletedAt    *time.Time `db:"completed_at"`
	LastHeartbeat  *time.Time `db:"last_heartbeat"`
	CreatedAt      time.Time  `db:"created_at"`
}

type ResearchRunRow struct {
	ID            int64           `db:"id"`
	Name          string          `db:"name"`
	Symbol        string          `db:"symbol"`
	Timeframe     string          `db:"timeframe"`
	FeatureSetID  int             `db:"feature_set_id"`
	Config        json.RawMessage `db:"config"`
	Status        string          `db:"status"`
	StartedAt     *time.Time      `db:"started_at"`
	CompletedAt   *time.Time      `db:"completed_at"`
	CreatedAt     time.Time       `db:"created_at"`
}

type ResearchResultRow struct {
	ID            int64           `db:"id"`
	ResearchRunID int64           `db:"research_run_id"`
	FeatureName   string          `db:"feature_name"`
	MetricName    string          `db:"metric_name"`
	MetricValue   float64         `db:"metric_value"`
	Samples       int             `db:"samples"`
	Metadata      json.RawMessage `db:"metadata"`
	CreatedAt     time.Time       `db:"created_at"`
}
