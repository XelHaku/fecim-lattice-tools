package validation

import "time"

// LiteratureReference documents the source of validation data
type LiteratureReference struct {
	DOI         string    `json:"doi"`
	Authors     string    `json:"authors"`
	Year        int       `json:"year"`
	Title       string    `json:"title"`
	Journal     string    `json:"journal"`
	Figure      string    `json:"figure,omitempty"`
	Table       string    `json:"table,omitempty"`
	ValidatedAt time.Time `json:"validated_at"`
}

// LiteratureDataPoint represents a single measurement from literature
type LiteratureDataPoint struct {
	Value       float64            `json:"value"`
	Unit        string             `json:"unit"`
	Uncertainty float64            `json:"uncertainty"`
	Conditions  map[string]float64 `json:"conditions,omitempty"`
}

// ValidationResult captures the outcome of a physics validation test
type ValidationResult struct {
	TestName        string
	Reference       LiteratureReference
	Expected        float64
	Measured        float64
	Uncertainty     float64
	RelativeError   float64
	WithinTolerance bool
	TolerancePct    float64
}

// StatisticalTestResult for stochastic validation
type StatisticalTestResult struct {
	TestName       string
	Statistic      float64
	PValue         float64
	NullHypothesis string
	Rejected       bool
	SampleSize     int
	Alpha          float64
}

// LiteratureDataset represents a collection of data points from a single source
type LiteratureDataset struct {
	Reference  LiteratureReference    `json:"reference"`
	DataPoints []LiteratureDataPoint  `json:"data_points"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}
