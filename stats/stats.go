package stats

import (
	"sync"
	"time"
)

type Stats struct {
	mu           sync.Mutex
	durations    []time.Duration
	sizes        []int
	successCount int
	requestCount int
}

func (s *Stats) Add(duration time.Duration, size int, success bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.durations = append(s.durations, duration)
	s.sizes = append(s.sizes, size)
	s.requestCount++

	if success {
		s.successCount++
	}
}

func (s *Stats) Get() (minDur, avgDur, maxDur time.Duration, minSize, avgSize, maxSize int, success, total int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.durations) == 0 {
		return 0, 0, 0, 0, 0, 0, 0, 0
	}

	minDur = s.durations[0]
	maxDur = s.durations[0]
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

	avgDur = totalDur / time.Duration(len(s.durations))

	minSize = s.sizes[0]
	maxSize = s.sizes[0]
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

	avgSize = totalSize / len(s.sizes)

	return minDur, avgDur, maxDur, minSize, avgSize, maxSize, s.successCount, s.requestCount
}
