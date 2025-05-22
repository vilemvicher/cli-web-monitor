package stats

import "time"

type StatsReport struct {
	MinDuration time.Duration
	AvgDuration time.Duration
	MaxDuration time.Duration
	MinSize     int
	AvgSize     int
	MaxSize     int
	Success     int
	Total       int
}
