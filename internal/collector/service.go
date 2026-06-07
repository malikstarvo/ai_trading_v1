package collector

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/avav/ai_trading_v1/internal/collector/bybit"
	"github.com/avav/ai_trading_v1/internal/collector/metrics"
	"github.com/avav/ai_trading_v1/internal/collector/poll"
	"github.com/avav/ai_trading_v1/internal/collector/recovery"
	"github.com/avav/ai_trading_v1/internal/collector/store"
	"github.com/avav/ai_trading_v1/internal/collector/stream"
	"github.com/avav/ai_trading_v1/internal/collector/validate"
)

var ErrAlreadyRunning = errors.New("collector already running")

type Service struct {
	cfg    Config
	log    *slog.Logger
	metrics  *metrics.CollectorMetrics

	bybitClient  *bybit.Client
	stream       *stream.Manager
	handlers     *stream.Handlers
	backfill     *poll.Backfill
	lsRatioPoller  *poll.LSRatioPoller
	recoverySched  *recovery.Scheduler

	candleStore    store.CandleStore
	orderFlowStore store.OrderFlowStore
	healthStore    store.CollectorHealthStore

	mu      sync.Mutex
	running bool
	cancel  context.CancelFunc
}

func NewService(
	cfg Config,
	candleStore store.CandleStore,
	orderFlowStore store.OrderFlowStore,
	healthStore store.CollectorHealthStore,
	m *metrics.CollectorMetrics,
	log *slog.Logger,
) *Service {
	log = log.With("module", "collector")

	rl := bybit.NewTokenBucket(cfg.RateLimit.RequestsPerSecond, cfg.RateLimit.Burst)
	bybitClient := bybit.NewClient(cfg.BaseURL, rl, log, cfg.InsecureSkipVerify, cfg.ProxyURL)

	validator := validate.New(m)

	ctx := context.Background()

	handlers := stream.NewHandlers(ctx, candleStore, orderFlowStore, validator, m, log)
	wsConn := stream.NewWSConn(cfg.WSURL, log, cfg.InsecureSkipVerify, cfg.ProxyURL)
	mgr := stream.NewManager(wsConn, handlers, m, log)

	bf := poll.NewBackfill(bybitClient, cfg.Symbols, candleStore, cfg.BackfillDays, log)
	lsPoller := poll.NewLSRatioPoller(bybitClient, cfg.Symbols, orderFlowStore, validator, m, log)

	detector := recovery.NewDetector(candleStore, bybitClient, cfg.Symbols, m, log, cfg.Recovery.GapBars)
	filler := recovery.NewFiller(bybitClient, candleStore, log)
	recSched := recovery.NewScheduler(detector, filler, cfg.Recovery.CheckInterval, log)

	return &Service{
		cfg:            cfg,
		log:            log,
		metrics:        m,
		bybitClient:    bybitClient,
		stream:         mgr,
		handlers:       handlers,
		backfill:       bf,
		lsRatioPoller:  lsPoller,
		recoverySched:  recSched,
		candleStore:    candleStore,
		orderFlowStore: orderFlowStore,
		healthStore:    healthStore,
	}
}

func (s *Service) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return ErrAlreadyRunning
	}
	s.running = true
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.mu.Unlock()

	s.log.Info("starting collector",
		"symbols", s.cfg.Symbols,
		"ws_url", s.cfg.WSURL,
		"rest_url", s.cfg.BaseURL,
	)

	// 1. Backfill historical data (non-fatal)
	if err := s.backfill.Run(ctx); err != nil {
		s.log.Warn("backfill failed (continuing)", "error", err)
	}

	// 2. Start WebSocket stream
	topics := s.buildTopics()
	if err := s.stream.Start(ctx, topics); err != nil {
		return err
	}

	// 3. Start L/S Ratio poller
	go s.runPoller(ctx, s.lsRatioPoller)

	// 4. Start recovery gap checker
	if s.cfg.Recovery.Enabled {
		go s.recoverySched.Run(ctx)
	}

	// 5. Start health heartbeat
	go s.heartbeatLoop(ctx)

	s.log.Info("collector started")
	return nil
}

func (s *Service) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return nil
	}
	s.running = false
	if s.cancel != nil {
		s.cancel()
	}
	return s.stream.Stop()
}

func (s *Service) buildTopics() []string {
	var topics []string
	for _, sym := range s.cfg.Symbols {
		topics = append(topics,
			"kline.15."+sym,
			"kline.60."+sym,
			"tickers."+sym,
			"allLiquidation."+sym,
		)
	}
	return topics
}

func (s *Service) runPoller(ctx context.Context, p poll.Poller) {
	defer func() {
		if r := recover(); r != nil {
			s.log.Error("panic in poller", "poller", p.Name(), "recover", r)
		}
	}()

	ticker := time.NewTicker(p.Interval())
	defer ticker.Stop()

	// Run immediately on start
	if err := p.Poll(ctx); err != nil {
		s.log.Error("initial poll failed", "poller", p.Name(), "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.Poll(ctx); err != nil {
				s.log.Error("poll failed", "poller", p.Name(), "error", err)
				if s.metrics != nil {
					s.metrics.ErrorsTotal.WithLabelValues("poll_" + p.Name()).Inc()
				}
			}
		}
	}
}

func (s *Service) heartbeatLoop(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			s.log.Error("panic in heartbeatLoop", "recover", r)
		}
	}()

	heartbeatTicker := time.NewTicker(30 * time.Second)
	statsTicker := time.NewTicker(5 * time.Minute)
	defer heartbeatTicker.Stop()
	defer statsTicker.Stop()

	// Log initial stats when collector starts
	s.handlers.LogStats()

	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeatTicker.C:
			if s.healthStore != nil {
				status := "healthy"
				if !s.stream.IsConnected() {
					status = "degraded"
				}
				_ = s.healthStore.Heartbeat(ctx, "collector", status)
			}
		case <-statsTicker.C:
			s.handlers.LogStats()
		}
	}
}

func (s *Service) IsConnected() bool {
	return s.stream.IsConnected()
}

func (s *Service) resolveSymbols(ctx context.Context) error {
	var resolved []string
	for _, sym := range s.cfg.Symbols {
		_, err := s.bybitClient.GetKlines(ctx, sym, "15", 0, 1, 1)
		if err != nil {
			if strings.Contains(err.Error(), "10001") {
				s.log.Warn("symbol not available on exchange, skipping", "symbol", sym)
				continue
			}
			return fmt.Errorf("check symbol %s: %w", sym, err)
		}
		resolved = append(resolved, sym)
	}

	if len(resolved) == 0 {
		return fmt.Errorf("no configured symbols are available (from config: %v)", s.cfg.Symbols)
	}

	original := s.cfg.Symbols
	s.cfg.Symbols = resolved
	s.log.Info("symbols resolved", "resolved", resolved, "from_config", original)
	return nil
}
