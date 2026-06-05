package recovery

import (
	"context"
	"log/slog"
	"time"
)

type Scheduler struct {
	detector     *Detector
	filler       *Filler
	interval     time.Duration
	logger       *slog.Logger
}

func NewScheduler(detector *Detector, filler *Filler, interval time.Duration, logger *slog.Logger) *Scheduler {
	return &Scheduler{
		detector: detector,
		filler:   filler,
		interval: interval,
		logger:   logger.With("module", "recovery_scheduler"),
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run immediately on start
	s.runOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runOnce(ctx)
		}
	}
}

func (s *Scheduler) runOnce(ctx context.Context) {
	reports, err := s.detector.Run(ctx)
	if err != nil {
		s.logger.Error("detection failed", "error", err)
		return
	}

	for _, report := range reports {
		if err := s.filler.Fill(ctx, report); err != nil {
			s.logger.Error("fill failed", "error", err, "symbol", report.Symbol)
		}
	}
}
