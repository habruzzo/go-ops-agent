package analyzers

import (
	"testing"
	"time"

	"github.com/habruzzo/agent/core"
)

func BenchmarkAnomalyAnalyzer_Analyze(b *testing.B) {
	analyzer := NewAnomalyAnalyzer("bench-analyzer")
	analyzer.Configure(map[string]interface{}{"threshold": 2.0})

	// Create test data
	data := make([]core.DataPoint, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = core.DataPoint{
			Timestamp: time.Now(),
			Source:    "benchmark",
			Metric:    "cpu_usage",
			Value:     float64(i % 100), // Some variation
			Labels:    map[string]string{"instance": "test"},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.Analyze(data)
		if err != nil {
			b.Fatalf("Analyze failed: %v", err)
		}
	}
}

func BenchmarkAnomalyAnalyzer_AnalyzeSmallDataset(b *testing.B) {
	analyzer := NewAnomalyAnalyzer("bench-analyzer")
	analyzer.Configure(map[string]interface{}{"threshold": 2.0})

	// Create small test data
	data := make([]core.DataPoint, 10)
	for i := 0; i < 10; i++ {
		data[i] = core.DataPoint{
			Timestamp: time.Now(),
			Source:    "benchmark",
			Metric:    "cpu_usage",
			Value:     float64(i % 10),
			Labels:    map[string]string{"instance": "test"},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.Analyze(data)
		if err != nil {
			b.Fatalf("Analyze failed: %v", err)
		}
	}
}

func BenchmarkAnomalyAnalyzer_AnalyzeLargeDataset(b *testing.B) {
	analyzer := NewAnomalyAnalyzer("bench-analyzer")
	analyzer.Configure(map[string]interface{}{"threshold": 2.0})

	// Create large test data
	data := make([]core.DataPoint, 10000)
	for i := 0; i < 10000; i++ {
		data[i] = core.DataPoint{
			Timestamp: time.Now(),
			Source:    "benchmark",
			Metric:    "cpu_usage",
			Value:     float64(i % 1000),
			Labels:    map[string]string{"instance": "test"},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.Analyze(data)
		if err != nil {
			b.Fatalf("Analyze failed: %v", err)
		}
	}
}

func BenchmarkAnomalyAnalyzer_CanAnalyze(b *testing.B) {
	analyzer := NewAnomalyAnalyzer("bench-analyzer")

	data := make([]core.DataPoint, 100)
	for i := 0; i < 100; i++ {
		data[i] = core.DataPoint{
			Timestamp: time.Now(),
			Source:    "benchmark",
			Metric:    "cpu_usage",
			Value:     float64(i),
			Labels:    map[string]string{"instance": "test"},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.CanAnalyze(data)
	}
}
