package recovery

import "time"

type GapReport struct {
	Symbol      string
	Timeframe   string
	GapStart    time.Time
	GapEnd      time.Time
	MissingBars int
	HasGap      bool
}
