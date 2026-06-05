package mock

import (
	"context"
	"sync"

	"github.com/avav/ai_trading_v1/internal/collector/stream"
)

type MockWSConn struct {
	mu           sync.Mutex
	connected    bool
	connectFn    func(ctx context.Context) error
	closeFn      func() error
	subscribeFn  func(topics []string) error
	readMsgFn    func() ([]byte, error)
	writeMsgFn   func(data []byte) error
}

func NewMockWSConn() *MockWSConn {
	return &MockWSConn{
		connected: false,
	}
}

func (m *MockWSConn) Connect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = true
	if m.connectFn != nil {
		return m.connectFn(ctx)
	}
	return nil
}

func (m *MockWSConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = false
	if m.closeFn != nil {
		return m.closeFn()
	}
	return nil
}

func (m *MockWSConn) Subscribe(topics []string) error {
	if m.subscribeFn != nil {
		return m.subscribeFn(topics)
	}
	return nil
}

func (m *MockWSConn) ReadMessage() ([]byte, error) {
	if m.readMsgFn != nil {
		return m.readMsgFn()
	}
	return nil, stream.ErrNotConnected
}

func (m *MockWSConn) WriteMessage(data []byte) error {
	if m.writeMsgFn != nil {
		return m.writeMsgFn(data)
	}
	return nil
}

func (m *MockWSConn) IsConnected() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.connected
}

func (m *MockWSConn) SetConnectFn(fn func(ctx context.Context) error) {
	m.connectFn = fn
}

func (m *MockWSConn) SetReadMsgFn(fn func() ([]byte, error)) {
	m.readMsgFn = fn
}
