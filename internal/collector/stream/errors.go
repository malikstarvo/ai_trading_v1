package stream

import "errors"

var (
	ErrNotConnected   = errors.New("websocket not connected")
	ErrAlreadyRunning = errors.New("stream manager already running")
)
