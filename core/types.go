package core

import (
	"time"
)

// DataPoint represents a single data point from any observability source
type DataPoint struct {
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Metric    string                 `json:"metric"`
	Value     float64                `json:"value"`
	Labels    map[string]string      `json:"labels"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// AnalysisType represents the type of analysis performed
type AnalysisType string

const (
	AnalysisTypeAnomaly     AnalysisType = "anomaly"
	AnalysisTypeTrend       AnalysisType = "trend"
	AnalysisTypeCorrelation AnalysisType = "correlation"
	AnalysisTypeAlert       AnalysisType = "alert"
)

// Analysis represents the result of analyzing data points
type Analysis struct {
	Type       AnalysisType           `json:"type"`
	Confidence float64                `json:"confidence"` // 0.0 to 1.0
	Severity   string                 `json:"severity"`   // low, medium, high, critical
	Summary    string                 `json:"summary"`
	Details    map[string]interface{} `json:"details"`
	DataPoints []DataPoint            `json:"data_points"`
	Timestamp  time.Time              `json:"timestamp"`
	Source     string                 `json:"source"`
}
