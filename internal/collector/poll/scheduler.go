package poll

import (
	"context"
	"time"
)

type Poller interface {
	Name() string
	Interval() time.Duration
	Poll(ctx context.Context) error
}
