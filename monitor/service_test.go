package monitor

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"cli-web-monitor/stats"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

// TestService_OneRequestAtATimePerURL testing of serial processing of requests for single url
func TestService_OneRequestAtATimePerURL(t *testing.T) {
	t.Parallel()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// mock URL
	url := "http://example.com/success"

	var (
		activeCalls int
		mu          sync.Mutex
	)

	// mock responses
	httpmock.RegisterResponder("GET", url,
		func(req *http.Request) (*http.Response, error) {
			mu.Lock()
			activeCalls++

			// evaluate
			require.Equal(t, 1, activeCalls, "Expected only one request for url at the time")

			mu.Unlock()

			time.Sleep(150 * time.Millisecond)

			mu.Lock()
			activeCalls--
			mu.Unlock()

			return httpmock.NewStringResponse(200, "OK"), nil
		})

	// run
	mockRenderer := func(m map[string]*stats.Stats) {}

	cfg := MonitorConfig{
		URLs:       []string{url},
		Client:     &http.Client{},
		Renderer:   mockRenderer,
		TickPeriod: 30 * time.Millisecond, // ticks faster than handler finishes
	}

	svc := NewService(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := svc.StartMonitoring(ctx)
	require.NoError(t, err)

	<-ctx.Done()
}

// TestService_ParallelRequestsAcrossDifferentURLs testing of processing each url parallel way
func TestService_ParallelRequestsAcrossDifferentURLs(t *testing.T) {
	t.Parallel()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// mock URLs
	urls := []string{
		"http://example.com/one",
		"http://example.com/two",
		"http://example.com/three",
	}

	var (
		mu            sync.Mutex
		activeURLs    = make(map[string]bool)
		maxConcurrent int
	)

	// mock responses
	for _, url := range urls {
		u := url // capture loop var

		httpmock.RegisterResponder("GET", u,
			func(req *http.Request) (*http.Response, error) {
				mu.Lock()
				activeURLs[u] = true

				if len(activeURLs) > maxConcurrent {
					maxConcurrent = len(activeURLs)
				}

				mu.Unlock()

				time.Sleep(150 * time.Millisecond)

				mu.Lock()
				delete(activeURLs, u)
				mu.Unlock()

				return httpmock.NewStringResponse(200, "OK"), nil
			})
	}

	// run
	renderer := func(m map[string]*stats.Stats) {}

	cfg := MonitorConfig{
		URLs:       urls,
		Client:     &http.Client{},
		Renderer:   renderer,
		TickPeriod: 20 * time.Millisecond,
	}

	svc := NewService(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := svc.StartMonitoring(ctx)
	require.NoError(t, err)

	<-ctx.Done()

	// evaluate
	require.Equal(t, len(urls), maxConcurrent, "Expected concurrent processing across all URLs")
}

// TestService_AllCombinations testing some general combinations
func TestService_AllCombinations(t *testing.T) {
	t.Parallel()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// mock URL
	urls := []string{
		"http://example.com/success",
		"http://example.com/failure",
		"http://example.com/timeout",
		"http://example.com/large",
	}

	// mock responses
	httpmock.RegisterResponder("GET", urls[0],
		httpmock.NewStringResponder(200, `{"status":"ok"}`)) // 200 OK

	httpmock.RegisterResponder("GET", urls[1],
		httpmock.NewStringResponder(404, `{"error":"not found"}`)) // 404 Not Found

	httpmock.RegisterResponder("GET", urls[2],
		func(req *http.Request) (*http.Response, error) {
			time.Sleep(200 * time.Millisecond) // Simulate timeout
			return httpmock.NewStringResponse(200, `{"status":"delayed ok"}`), nil
		})

	largeBody := strings.Repeat("x", 1024) // 1 KB
	httpmock.RegisterResponder("GET", urls[3],
		httpmock.NewStringResponder(200, largeBody)) // Large 200 response

	// run
	mockRenderer := func(m map[string]*stats.Stats) {}

	cfg := MonitorConfig{
		URLs:       urls,
		Client:     &http.Client{Timeout: 300 * time.Millisecond},
		Renderer:   mockRenderer,
		TickPeriod: 100 * time.Millisecond,
	}

	svc := NewService(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := svc.StartMonitoring(ctx)
	require.NoError(t, err)

	<-ctx.Done()

	// evaluate
	var sumSuccess, sumFail, sumTotal, sumBodyBytes int

	for _, url := range urls {
		stat := svc.GetStats()[url]
		_, _, _, _, bodyBytes, _, success, total := stat.Get()

		sumSuccess += success
		sumTotal += total
		sumBodyBytes += bodyBytes
	}

	sumFail = sumTotal - sumSuccess

	require.GreaterOrEqual(t, sumSuccess, 2)   // success, large
	require.GreaterOrEqual(t, sumFail, 1)      // 404
	require.GreaterOrEqual(t, sumTotal, 3)     // timeout might be skipped if too slow
	require.GreaterOrEqual(t, sumBodyBytes, 1) // large response is 1 KB, rest is too small
}
