package stats

import (
	"sync"
	"time"

	"cli-web-monitor/jsonmodel"
)

type Stats struct {
	mu           sync.Mutex
	durations    []time.Duration
	sizes        []int
	successCount int
	requestCount int
	Responses    []jsonmodel.RequestResponse
}

func (s *Stats) Add(duration time.Duration, size int, success bool, date time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.durations = append(s.durations, duration)
	s.sizes = append(s.sizes, size)
	s.requestCount++

	if success {
		s.successCount++
	}

	s.Responses = append(s.Responses, jsonmodel.RequestResponse{Date: date, OK: success, DurationMs: float64(duration)})
}

func (s *Stats) Get() *StatsReport {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.durations) == 0 {
		return nil
	}

	minDur := s.durations[0]
	maxDur := s.durations[0]
	totalDur := time.Duration(0)

	for _, d := range s.durations {
		if d < minDur {
			minDur = d
		}

		if d > maxDur {
			maxDur = d
		}

		totalDur += d
	}

	avgDur := totalDur / time.Duration(len(s.durations))

	minSize := s.sizes[0]
	maxSize := s.sizes[0]
	totalSize := 0

	for _, sz := range s.sizes {
		if sz < minSize {
			minSize = sz
		}

		if sz > maxSize {
			maxSize = sz
		}

		totalSize += sz
	}

	avgSize := totalSize / len(s.sizes)

	return &StatsReport{
		MinDuration: minDur,
		AvgDuration: avgDur,
		MaxDuration: maxDur,
		MinSize:     minSize,
		AvgSize:     avgSize,
		MaxSize:     maxSize,
		Success:     s.successCount,
		Total:       s.requestCount,
	}
}
