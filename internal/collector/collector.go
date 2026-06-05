package collector

import "context"

type Collector interface {
	Start(ctx context.Context) error
	Stop() error
}
