package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// TrafficPattern represents different types of traffic patterns
type TrafficPattern struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Duration    time.Duration `json:"duration"`
	RampUp      time.Duration `json:"ramp_up"`
	HoldTime    time.Duration `json:"hold_time"`
	RampDown    time.Duration `json:"ramp_down"`
	BaseLoad    int           `json:"base_load"`
	PeakLoad    int           `json:"peak_load"`
	Frequency   time.Duration `json:"frequency"`
}

// TrafficGenerator generates realistic traffic patterns
type TrafficGenerator struct {
	patterns   []TrafficPattern
	active     bool
	mu         sync.RWMutex
	metrics    *TrafficMetrics
	httpClient *http.Client
	targetURL  string
}

// TrafficMetrics tracks traffic generation statistics
type TrafficMetrics struct {
	TotalRequests  int64     `json:"total_requests"`
	SuccessfulReqs int64     `json:"successful_requests"`
	FailedReqs     int64     `json:"failed_requests"`
	AverageLatency float64   `json:"average_latency_ms"`
	MaxLatency     float64   `json:"max_latency_ms"`
	MinLatency     float64   `json:"min_latency_ms"`
	CurrentRPS     float64   `json:"current_rps"`
	StartTime      time.Time `json:"start_time"`
	LastUpdate     time.Time `json:"last_update"`
	mu             sync.RWMutex
}

// TrafficSpike represents a specific traffic spike event
type TrafficSpike struct {
	ID          string         `json:"id"`
	Pattern     TrafficPattern `json:"pattern"`
	StartTime   time.Time      `json:"start_time"`
	EndTime     time.Time      `json:"end_time"`
	CurrentLoad int            `json:"current_load"`
	Status      string         `json:"status"` // "pending", "running", "completed", "failed"
}

func main() {
	var (
		targetURL   = flag.String("target", "http://localhost:8080", "Target URL to send traffic to")
		pattern     = flag.String("pattern", "gradual", "Traffic pattern: gradual, spike, burst, chaos, black-friday, ddos")
		duration    = flag.Duration("duration", 5*time.Minute, "Total duration of traffic generation")
		baseLoad    = flag.Int("base-load", 10, "Base requests per second")
		peakLoad    = flag.Int("peak-load", 100, "Peak requests per second")
		metricsPort = flag.Int("metrics-port", 8081, "Port for metrics endpoint")
		configFile  = flag.String("config", "", "JSON config file for custom patterns")
		interactive = flag.Bool("interactive", false, "Run in interactive mode")
	)
	flag.Parse()

	// Set up logging
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// Create traffic generator
	generator := NewTrafficGenerator(*targetURL)

	// Load patterns
	if *configFile != "" {
		if err := generator.LoadPatternsFromFile(*configFile); err != nil {
			log.Fatalf("Failed to load patterns from file: %v", err)
		}
	} else {
		generator.LoadDefaultPatterns()
	}

	// Start metrics server
	go generator.StartMetricsServer(*metricsPort)

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("Shutting down traffic generator...")
		cancel()
	}()

	if *interactive {
		runInteractiveMode(ctx, generator)
	} else {
		// Run specified pattern
		pattern := generator.GetPattern(*pattern)
		if pattern == nil {
			log.Fatalf("Unknown pattern: %s", *pattern)
		}

		// Override pattern settings with command line flags
		pattern.BaseLoad = *baseLoad
		pattern.PeakLoad = *peakLoad
		pattern.Duration = *duration

		slog.Info("Starting traffic generation",
			"pattern", pattern.Name,
			"target", *targetURL,
			"duration", pattern.Duration,
			"base_load", pattern.BaseLoad,
			"peak_load", pattern.PeakLoad)

		if err := generator.RunPattern(ctx, *pattern); err != nil {
			log.Fatalf("Failed to run pattern: %v", err)
		}
	}

	slog.Info("Traffic generator stopped")
}

func NewTrafficGenerator(targetURL string) *TrafficGenerator {
	return &TrafficGenerator{
		patterns:   make([]TrafficPattern, 0),
		metrics:    &TrafficMetrics{StartTime: time.Now()},
		httpClient: &http.Client{Timeout: 30 * time.Second},
		targetURL:  targetURL,
	}
}

func (tg *TrafficGenerator) LoadDefaultPatterns() {
	tg.patterns = []TrafficPattern{
		{
			Name:        "gradual",
			Description: "Gradual ramp-up and ramp-down",
			Duration:    10 * time.Minute,
			RampUp:      3 * time.Minute,
			HoldTime:    4 * time.Minute,
			RampDown:    3 * time.Minute,
			BaseLoad:    10,
			PeakLoad:    100,
		},
		{
			Name:        "spike",
			Description: "Sudden spike with quick recovery",
			Duration:    2 * time.Minute,
			RampUp:      10 * time.Second,
			HoldTime:    30 * time.Second,
			RampDown:    10 * time.Second,
			BaseLoad:    10,
			PeakLoad:    500,
		},
		{
			Name:        "burst",
			Description: "Periodic bursts of traffic",
			Duration:    5 * time.Minute,
			RampUp:      5 * time.Second,
			HoldTime:    10 * time.Second,
			RampDown:    5 * time.Second,
			BaseLoad:    5,
			PeakLoad:    200,
			Frequency:   30 * time.Second,
		},
		{
			Name:        "chaos",
			Description: "Random chaotic traffic patterns",
			Duration:    3 * time.Minute,
			RampUp:      5 * time.Second,
			HoldTime:    5 * time.Second,
			RampDown:    5 * time.Second,
			BaseLoad:    5,
			PeakLoad:    300,
			Frequency:   10 * time.Second,
		},
		{
			Name:        "black-friday",
			Description: "Black Friday style traffic surge",
			Duration:    2 * time.Hour,
			RampUp:      30 * time.Minute,
			HoldTime:    90 * time.Minute,
			RampDown:    30 * time.Minute,
			BaseLoad:    50,
			PeakLoad:    1000,
		},
		{
			Name:        "ddos",
			Description: "DDoS attack simulation",
			Duration:    1 * time.Minute,
			RampUp:      5 * time.Second,
			HoldTime:    30 * time.Second,
			RampDown:    5 * time.Second,
			BaseLoad:    10,
			PeakLoad:    2000,
		},
	}
}

func (tg *TrafficGenerator) GetPattern(name string) *TrafficPattern {
	for i := range tg.patterns {
		if tg.patterns[i].Name == name {
			return &tg.patterns[i]
		}
	}
	return nil
}

func (tg *TrafficGenerator) RunPattern(ctx context.Context, pattern TrafficPattern) error {
	tg.mu.Lock()
	tg.active = true
	tg.mu.Unlock()

	defer func() {
		tg.mu.Lock()
		tg.active = false
		tg.mu.Unlock()
	}()

	// Reset metrics
	tg.metrics.mu.Lock()
	tg.metrics.TotalRequests = 0
	tg.metrics.SuccessfulReqs = 0
	tg.metrics.FailedReqs = 0
	tg.metrics.AverageLatency = 0
	tg.metrics.MaxLatency = 0
	tg.metrics.MinLatency = math.MaxFloat64
	tg.metrics.StartTime = time.Now()
	tg.metrics.LastUpdate = time.Now()
	tg.metrics.mu.Unlock()

	slog.Info("Starting traffic pattern", "pattern", pattern.Name)

	// Calculate load progression
	loadProgression := tg.calculateLoadProgression(pattern)

	// Start traffic generation
	var wg sync.WaitGroup
	done := make(chan struct{})

	// Start load progression
	wg.Add(1)
	go func() {
		defer wg.Done()
		tg.runLoadProgression(ctx, loadProgression, done)
	}()

	// Start metrics updater
	wg.Add(1)
	go func() {
		defer wg.Done()
		tg.updateMetrics(ctx, done)
	}()

	// Wait for completion or cancellation
	select {
	case <-ctx.Done():
		slog.Info("Traffic generation cancelled")
	case <-done:
		slog.Info("Traffic pattern completed")
	}

	// Stop all goroutines
	close(done)
	wg.Wait()

	// Print final metrics
	tg.printFinalMetrics()

	return nil
}

func (tg *TrafficGenerator) calculateLoadProgression(pattern TrafficPattern) []LoadPoint {
	var progression []LoadPoint
	now := time.Now()

	// Ramp up phase
	rampUpSteps := int(pattern.RampUp.Seconds())
	for i := 0; i < rampUpSteps; i++ {
		progress := float64(i) / float64(rampUpSteps)
		load := pattern.BaseLoad + int(float64(pattern.PeakLoad-pattern.BaseLoad)*progress)
		progression = append(progression, LoadPoint{
			Time: now.Add(time.Duration(i) * time.Second),
			Load: load,
		})
	}

	// Hold phase
	holdSteps := int(pattern.HoldTime.Seconds())
	for i := 0; i < holdSteps; i++ {
		progression = append(progression, LoadPoint{
			Time: now.Add(pattern.RampUp + time.Duration(i)*time.Second),
			Load: pattern.PeakLoad,
		})
	}

	// Ramp down phase
	rampDownSteps := int(pattern.RampDown.Seconds())
	for i := 0; i < rampDownSteps; i++ {
		progress := float64(i) / float64(rampDownSteps)
		load := pattern.PeakLoad - int(float64(pattern.PeakLoad-pattern.BaseLoad)*progress)
		progression = append(progression, LoadPoint{
			Time: now.Add(pattern.RampUp + pattern.HoldTime + time.Duration(i)*time.Second),
			Load: load,
		})
	}

	return progression
}

type LoadPoint struct {
	Time time.Time
	Load int
}

func (tg *TrafficGenerator) runLoadProgression(ctx context.Context, progression []LoadPoint, done chan struct{}) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	progressionIndex := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case <-ticker.C:
			if progressionIndex >= len(progression) {
				close(done)
				return
			}

			currentLoad := progression[progressionIndex].Load
			tg.generateLoad(ctx, currentLoad)
			progressionIndex++
		}
	}
}

func (tg *TrafficGenerator) generateLoad(ctx context.Context, rps int) {
	if rps <= 0 {
		return
	}

	interval := time.Second / time.Duration(rps)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Generate requests for 1 second
	timeout := time.After(time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-timeout:
			return
		case <-ticker.C:
			go tg.sendRequest(ctx)
		}
	}
}

func (tg *TrafficGenerator) sendRequest(ctx context.Context) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, "GET", tg.targetURL+"/health", nil)
	if err != nil {
		tg.recordFailedRequest(0)
		return
	}

	resp, err := tg.httpClient.Do(req)
	latency := float64(time.Since(start).Milliseconds())

	if err != nil {
		tg.recordFailedRequest(latency)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		tg.recordSuccessfulRequest(latency)
	} else {
		tg.recordFailedRequest(latency)
	}
}

func (tg *TrafficGenerator) recordSuccessfulRequest(latency float64) {
	tg.metrics.mu.Lock()
	defer tg.metrics.mu.Unlock()

	tg.metrics.TotalRequests++
	tg.metrics.SuccessfulReqs++
	tg.updateLatencyStats(latency)
}

func (tg *TrafficGenerator) recordFailedRequest(latency float64) {
	tg.metrics.mu.Lock()
	defer tg.metrics.mu.Unlock()

	tg.metrics.TotalRequests++
	tg.metrics.FailedReqs++
	tg.updateLatencyStats(latency)
}

func (tg *TrafficGenerator) updateLatencyStats(latency float64) {
	if latency > tg.metrics.MaxLatency {
		tg.metrics.MaxLatency = latency
	}
	if latency < tg.metrics.MinLatency {
		tg.metrics.MinLatency = latency
	}

	// Simple moving average
	if tg.metrics.TotalRequests == 1 {
		tg.metrics.AverageLatency = latency
	} else {
		tg.metrics.AverageLatency = (tg.metrics.AverageLatency*float64(tg.metrics.TotalRequests-1) + latency) / float64(tg.metrics.TotalRequests)
	}
}

func (tg *TrafficGenerator) updateMetrics(ctx context.Context, done chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case <-ticker.C:
			tg.metrics.mu.Lock()
			now := time.Now()
			elapsed := now.Sub(tg.metrics.StartTime).Seconds()
			if elapsed > 0 {
				tg.metrics.CurrentRPS = float64(tg.metrics.TotalRequests) / elapsed
			}
			tg.metrics.LastUpdate = now
			tg.metrics.mu.Unlock()

			slog.Info("Traffic metrics",
				"total_requests", tg.metrics.TotalRequests,
				"successful", tg.metrics.SuccessfulReqs,
				"failed", tg.metrics.FailedReqs,
				"current_rps", fmt.Sprintf("%.2f", tg.metrics.CurrentRPS),
				"avg_latency_ms", fmt.Sprintf("%.2f", tg.metrics.AverageLatency),
				"max_latency_ms", fmt.Sprintf("%.2f", tg.metrics.MaxLatency))
		}
	}
}

func (tg *TrafficGenerator) printFinalMetrics() {
	tg.metrics.mu.RLock()
	defer tg.metrics.mu.RUnlock()

	fmt.Println("\n=== Final Traffic Metrics ===")
	fmt.Printf("Total Requests: %d\n", tg.metrics.TotalRequests)
	fmt.Printf("Successful: %d (%.2f%%)\n", tg.metrics.SuccessfulReqs,
		float64(tg.metrics.SuccessfulReqs)/float64(tg.metrics.TotalRequests)*100)
	fmt.Printf("Failed: %d (%.2f%%)\n", tg.metrics.FailedReqs,
		float64(tg.metrics.FailedReqs)/float64(tg.metrics.TotalRequests)*100)
	fmt.Printf("Average RPS: %.2f\n", tg.metrics.CurrentRPS)
	fmt.Printf("Average Latency: %.2f ms\n", tg.metrics.AverageLatency)
	fmt.Printf("Max Latency: %.2f ms\n", tg.metrics.MaxLatency)
	fmt.Printf("Min Latency: %.2f ms\n", tg.metrics.MinLatency)
	fmt.Printf("Duration: %v\n", time.Since(tg.metrics.StartTime))
}

func (tg *TrafficGenerator) StartMetricsServer(port int) {
	http.HandleFunc("/metrics", tg.handleMetrics)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	slog.Info("Starting metrics server", "port", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		slog.Error("Failed to start metrics server", "error", err)
	}
}

func (tg *TrafficGenerator) handleMetrics(w http.ResponseWriter, r *http.Request) {
	tg.metrics.mu.RLock()
	defer tg.metrics.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tg.metrics)
}

func (tg *TrafficGenerator) LoadPatternsFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var patterns []TrafficPattern
	if err := json.Unmarshal(data, &patterns); err != nil {
		return err
	}

	tg.patterns = patterns
	return nil
}

func runInteractiveMode(ctx context.Context, generator *TrafficGenerator) {
	fmt.Println("Traffic Generator - Interactive Mode")
	fmt.Println("Available patterns:")
	for _, pattern := range generator.patterns {
		fmt.Printf("  - %s: %s\n", pattern.Name, pattern.Description)
	}
	fmt.Println("Type 'help' for commands, 'quit' to exit")
	fmt.Println()

	// Simple interactive loop (you could enhance this)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("traffic> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		command := parts[0]

		switch command {
		case "quit", "exit":
			fmt.Println("Goodbye!")
			return
		case "help":
			fmt.Println("Commands:")
			fmt.Println("  run <pattern> [duration] [base-load] [peak-load]")
			fmt.Println("  list")
			fmt.Println("  quit")
		case "list":
			for _, pattern := range generator.patterns {
				fmt.Printf("  - %s: %s\n", pattern.Name, pattern.Description)
			}
		case "run":
			if len(parts) < 2 {
				fmt.Println("Usage: run <pattern> [duration] [base-load] [peak-load]")
				continue
			}

			patternName := parts[1]
			pattern := generator.GetPattern(patternName)
			if pattern == nil {
				fmt.Printf("Unknown pattern: %s\n", patternName)
				continue
			}

			// Parse optional parameters
			if len(parts) > 2 {
				if duration, err := time.ParseDuration(parts[2]); err == nil {
					pattern.Duration = duration
				}
			}
			if len(parts) > 3 {
				if baseLoad, err := strconv.Atoi(parts[3]); err == nil {
					pattern.BaseLoad = baseLoad
				}
			}
			if len(parts) > 4 {
				if peakLoad, err := strconv.Atoi(parts[4]); err == nil {
					pattern.PeakLoad = peakLoad
				}
			}

			fmt.Printf("Running pattern: %s\n", pattern.Name)
			if err := generator.RunPattern(ctx, *pattern); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		default:
			fmt.Printf("Unknown command: %s. Type 'help' for commands.\n", command)
		}
	}
}
