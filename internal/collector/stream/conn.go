package stream

import (
	"context"
	"crypto/tls"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Conn interface {
	Connect(ctx context.Context) error
	Close() error
	Subscribe(topics []string) error
	ReadMessage() ([]byte, error)
	WriteMessage(data []byte) error
	IsConnected() bool
}

type WSConn struct {
	url    string
	dialer *websocket.Dialer
	conn   *websocket.Conn
	mu     sync.RWMutex
	logger *slog.Logger
}

func NewWSConn(url string, logger *slog.Logger, insecureSkipVerify bool) *WSConn {
	dialer := &websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	if insecureSkipVerify {
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return &WSConn{
		url:    url,
		dialer: dialer,
		logger: logger.With("module", "ws_conn"),
	}
}

func (c *WSConn) Connect(ctx context.Context) error {
	conn, _, err := c.dialer.DialContext(ctx, c.url, nil)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	c.logger.Info("ws connected", "url", c.url)
	return nil
}

func (c *WSConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

func (c *WSConn) Subscribe(topics []string) error {
	msg := struct {
		Op   string   `json:"op"`
		Args []string `json:"args"`
	}{
		Op:   "subscribe",
		Args: topics,
	}

	data, err := encodeJSON(msg)
	if err != nil {
		return err
	}

	return c.WriteMessage(data)
}

func (c *WSConn) ReadMessage() ([]byte, error) {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()
	if conn == nil {
		return nil, ErrNotConnected
	}
	_, msg, err := conn.ReadMessage()
	return msg, err
}

func (c *WSConn) WriteMessage(data []byte) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()
	if conn == nil {
		return ErrNotConnected
	}
	return conn.WriteMessage(websocket.TextMessage, data)
}

func (c *WSConn) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn != nil
}
