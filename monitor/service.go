package monitor

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"

	"cli-web-monitor/stats"
)

type Service interface {
	StartMonitoring(context.Context) error
	GetStats() map[string]*stats.Stats
}

type service struct {
	config      MonitorConfig
	statsMap    map[string]*stats.Stats
	renderChan  chan struct{}
	tickerChans map[string]chan struct{}
	wg          *sync.WaitGroup
}

func NewService(cfg MonitorConfig) Service {
	statsMap := make(map[string]*stats.Stats)

	for _, url := range cfg.URLs {
		statsMap[url] = &stats.Stats{}
	}

	return &service{
		config:      cfg,
		statsMap:    statsMap,
		renderChan:  make(chan struct{}, 1),
		tickerChans: make(map[string]chan struct{}),
		wg:          &sync.WaitGroup{},
	}
}

func (srv *service) StartMonitoring(ctx context.Context) error {
	// monitor each url
	for _, url := range srv.config.URLs {
		tickChan := make(chan struct{}, 1)
		srv.tickerChans[url] = tickChan
		srv.wg.Add(1)
		go monitorURL(ctx, url, srv.statsMap[url], srv.wg, tickChan, srv.renderChan, srv.config.Client)
	}

	// first update on start
	for _, ch := range srv.tickerChans {
		select {
		case ch <- struct{}{}:
		default:
		}
	}

	// periodic updates
	ticker := time.NewTicker(srv.config.TickPeriod)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-ticker.C:
			for _, ch := range srv.tickerChans {
				select {
				case ch <- struct{}{}:
				default:
				}
			}
		case <-srv.renderChan:
			srv.config.Renderer(srv.statsMap)
		case <-ctx.Done():
			break loop
		}
	}

	srv.wg.Wait()
	srv.config.Renderer(srv.statsMap)

	return nil
}

func (srv *service) GetStats() map[string]*stats.Stats {
	return srv.statsMap
}

func monitorURL(
	ctx context.Context,
	url string,
	stats *stats.Stats,
	wg *sync.WaitGroup,
	tickChan <-chan struct{},
	renderChan chan<- struct{},
	client *http.Client,
) {
	defer wg.Done()
	busy := false

	for {
		select {
		case <-ctx.Done():
			return
		case <-tickChan:
			if busy {
				// skip this tick if a last request is still running
				continue
			}

			busy = true

			go func() {
				defer func() { busy = false }()

				start := time.Now()
				resp, err := client.Get(url)
				duration := time.Since(start)

				success := false
				size := 0

				if err == nil {
					defer resp.Body.Close()
					success = resp.StatusCode >= 200 && resp.StatusCode < 400

					bodyBytes, _ := io.ReadAll(resp.Body)

					size += len(bodyBytes) / 1024
				}

				stats.Add(duration, size, success)

				select {
				case renderChan <- struct{}{}:
				default:
				}
			}()
		}
	}
}
