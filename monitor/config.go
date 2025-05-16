package monitor

import (
	"net/http"
	"time"

	"cli-web-monitor/stats"
)

type MonitorConfig struct {
	URLs       []string
	Client     *http.Client
	Renderer   func(statsMap map[string]*stats.Stats)
	TickPeriod time.Duration
}
