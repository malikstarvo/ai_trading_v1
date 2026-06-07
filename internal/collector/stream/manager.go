package stream

import (
	"context"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/avav/ai_trading_v1/internal/collector/metrics"
)

type Manager struct {
	conn              Conn
	handlers          *Handlers
	metrics           *metrics.CollectorMetrics
	logger            *slog.Logger
	mu                sync.Mutex
	running           bool
	cancel            context.CancelFunc
	disconnectedCh    chan struct{}
	topics            []string
	lastMessageTime   map[string]time.Time
	lastMessageMu     sync.RWMutex
	pingInterval      time.Duration
	reconnectBase     time.Duration
	reconnectMax      time.Duration
	maxSilentInterval time.Duration
}

func NewManager(conn Conn, handlers *Handlers, m *metrics.CollectorMetrics, logger *slog.Logger) *Manager {
	return &Manager{
		conn:              conn,
		handlers:          handlers,
		metrics:           m,
		logger:            logger.With("module", "ws_manager"),
		disconnectedCh:    make(chan struct{}, 1),
		lastMessageTime:   make(map[string]time.Time),
		pingInterval:      20 * time.Second,
		reconnectBase:     time.Second,
		reconnectMax:      30 * time.Second,
		maxSilentInterval: 180 * time.Second,
	}
}

func (m *Manager) Start(ctx context.Context, topics []string) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return ErrAlreadyRunning
	}
	m.running = true
	ctx, cancel := context.WithCancel(ctx)
	m.cancel = cancel
	m.topics = topics
	m.mu.Unlock()

	go m.run(ctx)
	go m.watchdog(ctx)
	return nil
}

func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.running {
		return nil
	}
	m.running = false
	if m.cancel != nil {
		m.cancel()
	}
	return m.conn.Close()
}

func (m *Manager) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := m.connect(ctx); err != nil {
			m.logger.Error("connect failed", "error", err)
			if !m.waitRetry(ctx) {
				return
			}
			continue
		}

		m.logger.Info("subscribing", "topics", m.topics)
		if err := m.conn.Subscribe(m.topics); err != nil {
			m.logger.Error("subscribe failed", "error", err)
			m.triggerDisconnect()
			continue
		}

		m.readLoop(ctx)
	}
}

func (m *Manager) connect(ctx context.Context) error {
	m.logger.Info("connecting")
	return m.conn.Connect(ctx)
}

func (m *Manager) readLoop(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			m.logger.Error("panic in readLoop", "recover", r)
			if m.metrics != nil {
				m.metrics.ErrorsTotal.WithLabelValues("ws_panic").Inc()
			}
			m.triggerDisconnect()
		}
	}()

	pingTicker := time.NewTicker(m.pingInterval)
	defer pingTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pingTicker.C:
			m.sendPing()
		default:
		}

		msg, err := m.conn.ReadMessage()
		if err != nil {
			m.logger.Error("read error", "error", err)
			if m.metrics != nil {
				m.metrics.ErrorsTotal.WithLabelValues("ws_read").Inc()
			}
			m.triggerDisconnect()
			return
		}

		m.handlers.HandleMessage(msg)
	}
}

func (m *Manager) sendPing() {
	msg := []byte(`{"op":"ping"}`)
	if err := m.conn.WriteMessage(msg); err != nil {
		m.logger.Debug("ping failed", "error", err)
	}
}

func (m *Manager) triggerDisconnect() {
	if m.metrics != nil {
		m.metrics.ReconnectTotal.Inc()
	}
	_ = m.conn.Close()
	select {
	case m.disconnectedCh <- struct{}{}:
	default:
	}
}

func (m *Manager) waitRetry(ctx context.Context) bool {
	backoff := m.reconnectBase
	for attempt := 0; ; attempt++ {
		jitter := time.Duration(rand.Int63n(int64(backoff / 4)))
		wait := backoff + jitter
		m.logger.Info("reconnecting", "attempt", attempt+1, "backoff", wait)
		select {
		case <-ctx.Done():
			return false
		case <-time.After(wait):
		}
		if m.conn.IsConnected() {
			_ = m.conn.Close()
		}
		if err := m.conn.Connect(ctx); err == nil {
			return true
		}
		backoff *= 2
		if backoff > m.reconnectMax {
			backoff = m.reconnectMax
		}
	}
}

func (m *Manager) UpdateMessageTime(topic string) {
	m.lastMessageMu.Lock()
	m.lastMessageTime[topic] = time.Now()
	m.lastMessageMu.Unlock()
}

func (m *Manager) IsConnected() bool {
	return m.conn.IsConnected()
}

func (m *Manager) LastMessageAge(topic string) time.Duration {
	m.lastMessageMu.RLock()
	defer m.lastMessageMu.RUnlock()
	t, ok := m.lastMessageTime[topic]
	if !ok {
		return time.Duration(0)
	}
	return time.Since(t)
}

func (m *Manager) watchdog(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		if !m.conn.IsConnected() {
			continue
		}
		for _, topic := range m.topics {
			age := m.LastMessageAge(topic)
			if age > m.maxSilentInterval {
				m.logger.Warn("topic silent too long, triggering reconnect",
					"topic", topic, "age", age.Round(time.Second))
				if m.metrics != nil {
					m.metrics.ErrorsTotal.WithLabelValues("ws_stale").Inc()
				}
				m.triggerDisconnect()
				break
			}
		}
	}
}
