package edge

type FeatureInfo struct {
	Name string
	Col  string
}

type EdgeFilter struct {
	Symbol       string
	Timeframe    string
	FeatureSetID int
}

type LabelHorizon struct {
	Name string
	Col  string
}

type StudyConfig struct {
	Symbols          []string
	Timeframes       []string
	FeatureSetID     int
	LabelHorizons    []LabelHorizon
	RollingWindows   []int
	QuantileNBuckets int
	RegimePct        float64
	Features         []FeatureInfo
}
