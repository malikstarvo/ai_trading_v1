package db

import "time"

type QuantileRow struct {
	Bucket       int
	Trades       int
	AvgReturn    float64
	WinRate      float64
	ProfitFactor float64
}

type RollingPoint struct {
	Ts   time.Time
	Corr float64
}

type RegimeRow struct {
	TrendRegime string
	VolRegime   string
	Corr        float64
	Samples     int
}

type MetricsRow struct {
	FeatureName  string
	LabelHorizon int
	MetricName   string
	MetricValue  float64
	Samples      int
	Metadata     map[string]interface{}
}

type DecayResult struct {
	PeakCorr    float64
	AvgCorr     float64
	DecayRate   float64
	Persistence float64
}

type RollingSummary struct {
	Window    int
	Mean      float64
	Std       float64
	Stability float64
}

type FeatureScore struct {
	FeatureName    string
	CompositeScore float64
	Rank           int
}
