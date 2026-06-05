package edge

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/avav/ai_trading_v1/internal/db"
)

type Study struct {
	store      *db.EdgeStore
	resStore   *db.ResearchStore
	cfg        StudyConfig
	allResults []MetricResult
}

type MetricResult struct {
	FeatureName  string
	LabelHorizon string
	MetricName   string
	MetricValue  float64
	Samples      int
	Metadata     map[string]interface{}
}

func NewStudy(store *db.EdgeStore, resStore *db.ResearchStore, cfg StudyConfig) *Study {
	return &Study{
		store:    store,
		resStore: resStore,
		cfg:      cfg,
	}
}

func (s *Study) RunAll(ctx context.Context) error {
	count, err := s.store.CountFeatures(ctx, s.cfg.FeatureSetID)
	if err != nil {
		return fmt.Errorf("check feature_values: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("feature_values is empty for feature_set_id=%d — run feature-backfill first", s.cfg.FeatureSetID)
	}
	for _, sym := range s.cfg.Symbols {
		for _, tf := range s.cfg.Timeframes {
			filter := EdgeFilter{
				Symbol:       sym,
				Timeframe:    tf,
				FeatureSetID: s.cfg.FeatureSetID,
			}
			log.Printf("Running edge study: %s %s", sym, tf)
			if err := s.runForFilter(ctx, filter); err != nil {
				return fmt.Errorf("%s %s: %w", sym, tf, err)
			}
		}
	}
	return nil
}

func (s *Study) runForFilter(ctx context.Context, filter EdgeFilter) error {
	var featureScores []ComponentScore

	for _, feature := range s.cfg.Features {
		featureScores = append(featureScores, s.analyzeFeature(ctx, filter, feature))
	}

	ranked := ComputeRanking(featureScores)

	for i, fs := range ranked {
		s.addResult(fs.FeatureName, "", "composite_score", fs.CompositeScore, 0, map[string]interface{}{
			"rank": i + 1, "symbol": filter.Symbol, "timeframe": filter.Timeframe,
		})
		log.Printf("  Rank %d: %s (score=%.3f)", i+1, fs.FeatureName, fs.CompositeScore)
	}

	return nil
}

func (s *Study) analyzeFeature(ctx context.Context, filter EdgeFilter, feature FeatureInfo) ComponentScore {
	var cs ComponentScore
	cs.FeatureName = feature.Name

	var horizonCorrs []float64
	var stabilitySum float64
	var stabilityCount int

	for _, h := range s.cfg.LabelHorizons {
		p, sp, samples, err := RunCorrelation(ctx, s.store, filter, feature, h)
		if err != nil {
			log.Printf("  WARN: corr %s/%s: %v", feature.Name, h.Name, err)
			continue
		}

		s.addResult(feature.Name, h.Name, "pearson", p, samples, map[string]interface{}{"symbol": filter.Symbol, "timeframe": filter.Timeframe})
		s.addResult(feature.Name, h.Name, "spearman", sp, samples, map[string]interface{}{"symbol": filter.Symbol, "timeframe": filter.Timeframe})

		avgCorr := (math.Abs(p) + math.Abs(sp)) / 2
		horizonCorrs = append(horizonCorrs, avgCorr)

		buckets, err := RunQuantile(ctx, s.store, filter, feature, h, s.cfg.QuantileNBuckets)
		if err != nil {
			log.Printf("  WARN: quantile %s/%s: %v", feature.Name, h.Name, err)
		} else if len(buckets) > 0 {
			top := buckets[len(buckets)-1]
			bottom := buckets[0]

			s.addResult(feature.Name, h.Name, "q_top_pf", top.ProfitFactor, top.Trades, nil)
			s.addResult(feature.Name, h.Name, "q_top_wr", top.WinRate, top.Trades, nil)
			s.addResult(feature.Name, h.Name, "q_bottom_pf", bottom.ProfitFactor, bottom.Trades, nil)
			s.addResult(feature.Name, h.Name, "q_bottom_wr", bottom.WinRate, bottom.Trades, nil)

			if top.ProfitFactor > cs.QuantilePF {
				cs.QuantilePF = top.ProfitFactor
			}
			wrDelta := top.WinRate - bottom.WinRate
			if wrDelta > cs.QuantileWRDelta {
				cs.QuantileWRDelta = wrDelta
			}
		}

		summaries, err := RunRolling(ctx, s.store, filter, feature, h, s.cfg.RollingWindows)
		if err != nil {
			log.Printf("  WARN: rolling %s/%s: %v", feature.Name, h.Name, err)
		} else {
			for _, rs := range summaries {
				s.addResult(feature.Name, h.Name, fmt.Sprintf("rolling_mean_%d", rs.Window), rs.Mean, 0, nil)
				s.addResult(feature.Name, h.Name, fmt.Sprintf("rolling_std_%d", rs.Window), rs.Std, 0, nil)
				s.addResult(feature.Name, h.Name, fmt.Sprintf("rolling_stability_%d", rs.Window), rs.Stability, 0, nil)
				stabilitySum += rs.Stability
				stabilityCount++
			}
		}

		regimes, err := RunRegime(ctx, s.store, filter, feature, h, s.cfg.RegimePct)
		if err != nil {
			log.Printf("  WARN: regime %s/%s: %v", feature.Name, h.Name, err)
		} else {
			for _, r := range regimes {
				s.addResult(feature.Name, h.Name, fmt.Sprintf("regime_%s_%s", r.TrendRegime, r.VolRegime), r.Corr, r.Samples, nil)
			}
			rc := RegimeConsistency(regimes)
			if rc > cs.RegimeConsistencyVal {
				cs.RegimeConsistencyVal = rc
			}
		}
	}

	if stabilityCount > 0 {
		cs.RollingStability = stabilitySum / float64(stabilityCount)
	}

	var totalAbs float64
	for _, c := range horizonCorrs {
		totalAbs += c
	}
	if len(horizonCorrs) > 0 {
		cs.AvgAbsCorrelation = totalAbs / float64(len(horizonCorrs))
	}

	decay := AnalyzeDecay(horizonCorrs)
	s.addResult(feature.Name, "", "decay_peak", decay.PeakCorr, 0, nil)
	s.addResult(feature.Name, "", "decay_avg", decay.AvgCorr, 0, nil)
	s.addResult(feature.Name, "", "decay_rate", decay.DecayRate, 0, nil)
	s.addResult(feature.Name, "", "decay_persistence", decay.Persistence, 0, nil)

	return cs
}

func (s *Study) addResult(featureName, labelHorizon, metricName string, value float64, samples int, meta map[string]interface{}) {
	if meta == nil {
		meta = make(map[string]interface{})
	}
	s.allResults = append(s.allResults, MetricResult{
		FeatureName:  featureName,
		LabelHorizon: labelHorizon,
		MetricName:   metricName,
		MetricValue:  value,
		Samples:      samples,
		Metadata:     meta,
	})
}

func (s *Study) GetResults() []MetricResult {
	return s.allResults
}

func (s *Study) SaveToDB(ctx context.Context, runID int64) error {
	for _, r := range s.allResults {
		metaJSON, _ := json.Marshal(r.Metadata)
		labelHorizon := 0
		if r.LabelHorizon == "future_return_4" {
			labelHorizon = 4
		} else if r.LabelHorizon == "future_return_12" {
			labelHorizon = 12
		} else if r.LabelHorizon == "future_return_24" {
			labelHorizon = 24
		}
		if r.Metadata == nil {
			r.Metadata = make(map[string]interface{})
		}
		r.Metadata["horizon"] = labelHorizon
		if labelHorizon > 0 {
			metaJSON, _ = json.Marshal(r.Metadata)
		}

		if err := s.resStore.SaveResult(ctx, runID, r.FeatureName, r.MetricName, r.MetricValue, r.Samples, metaJSON); err != nil {
			return fmt.Errorf("save result %s/%s: %w", r.FeatureName, r.MetricName, err)
		}
	}
	return nil
}

func (s *Study) TopFeatures(n int) []string {
	unique := make(map[string]float64)
	for _, r := range s.allResults {
		if r.MetricName == "composite_score" {
			if existing, ok := unique[r.FeatureName]; !ok || r.MetricValue > existing {
				unique[r.FeatureName] = r.MetricValue
			}
		}
	}

	type pair struct {
		name  string
		score float64
	}
	var pairs []pair
	for name, score := range unique {
		pairs = append(pairs, pair{name, score})
	}

	for i := 0; i < len(pairs); i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].score > pairs[i].score {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}

	if n > len(pairs) {
		n = len(pairs)
	}

	var top []string
	for i := 0; i < n; i++ {
		top = append(top, pairs[i].name)
	}
	return top
}

func init() {
	_ = time.Now
}
