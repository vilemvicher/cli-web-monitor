package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cli-web-monitor/monitor"
)

const traceRuntime = false

func main() {
	// tracing for dev
	if traceRuntime {
		cleanup, err := startProfiling()

		if err != nil {
			log.Fatalf("Failed to start tracing - %v", err)
		}

		defer cleanup()
	}

	// input
	urls := os.Args[1:]

	if err := validateInputs(urls); err != nil {
		log.Fatalf("Invalid input - error: %v", err)
	}

	// service
	cfg := monitor.MonitorConfig{
		URLs:       urls,
		Client:     &http.Client{Timeout: 10 * time.Second},
		Renderer:   renderTable,
		TickPeriod: 5 * time.Second,
	}

	service := monitor.NewService(cfg)

	// run
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	go func() {
		if err := service.StartMonitoring(ctx); err != nil {
			log.Fatalf("Monitoring failed: %v", err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/stats/{website}", monitor.CreateGetStatsEndpoint(service))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		fmt.Println("Starting server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	<-ctx.Done()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %v\n", err)
	}
}

func validateInputs(inputs []string) error {
	if inputs == nil || len(inputs) == 0 {
		return errors.New("missing input")
	}

	for _, input := range inputs {
		parsed, err := url.ParseRequestURI(input)

		if err != nil {
			return fmt.Errorf("parsing error - %v", err)
		}

		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return errors.New(fmt.Sprintf("invalid connection protocol in %s - %s", input, parsed.Scheme))
		}
	}

	return nil
}
