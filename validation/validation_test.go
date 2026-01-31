package validation

import (
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLiteratureDatasetSaveLoad(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_dataset.json")

	// Create a test dataset
	originalDataset := &LiteratureDataset{
		Reference: LiteratureReference{
			DOI:         "10.1038/example",
			Authors:     "Smith et al.",
			Year:        2024,
			Title:       "Test Paper",
			Journal:     "Nature Communications",
			Figure:      "Fig. 1a",
			ValidatedAt: time.Now().UTC(),
		},
		DataPoints: []LiteratureDataPoint{
			{
				Value:       15.5,
				Unit:        "µC/cm²",
				Uncertainty: 0.5,
				Conditions: map[string]float64{
					"temperature": 300.0,
					"frequency":   1000.0,
				},
			},
			{
				Value:       34.2,
				Unit:        "µC/cm²",
				Uncertainty: 1.2,
				Conditions: map[string]float64{
					"temperature": 4.0,
					"frequency":   1000.0,
				},
			},
		},
		Metadata: map[string]interface{}{
			"material": "HfO2-ZrO2",
			"verified": true,
		},
	}

	// Save the dataset
	if err := SaveLiteratureDataset(originalDataset, filePath); err != nil {
		t.Fatalf("Failed to save dataset: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("Dataset file was not created")
	}

	// Load the dataset
	loadedDataset, err := LoadLiteratureDataset(filePath)
	if err != nil {
		t.Fatalf("Failed to load dataset: %v", err)
	}

	// Verify basic fields
	if loadedDataset.Reference.DOI != originalDataset.Reference.DOI {
		t.Errorf("DOI mismatch: got %s, want %s", loadedDataset.Reference.DOI, originalDataset.Reference.DOI)
	}

	if len(loadedDataset.DataPoints) != len(originalDataset.DataPoints) {
		t.Errorf("DataPoints count mismatch: got %d, want %d", len(loadedDataset.DataPoints), len(originalDataset.DataPoints))
	}

	if loadedDataset.DataPoints[0].Value != originalDataset.DataPoints[0].Value {
		t.Errorf("First data point value mismatch: got %f, want %f", loadedDataset.DataPoints[0].Value, originalDataset.DataPoints[0].Value)
	}
}

func TestExtractValues(t *testing.T) {
	dataset := &LiteratureDataset{
		DataPoints: []LiteratureDataPoint{
			{Value: 1.5, Unit: "test"},
			{Value: 2.5, Unit: "test"},
			{Value: 3.5, Unit: "test"},
		},
	}

	values := ExtractValues(dataset)
	expected := []float64{1.5, 2.5, 3.5}

	if len(values) != len(expected) {
		t.Fatalf("Length mismatch: got %d, want %d", len(values), len(expected))
	}

	for i := range values {
		if values[i] != expected[i] {
			t.Errorf("Value[%d] mismatch: got %f, want %f", i, values[i], expected[i])
		}
	}
}

func TestFilterByCondition(t *testing.T) {
	dataset := &LiteratureDataset{
		DataPoints: []LiteratureDataPoint{
			{Value: 1.0, Conditions: map[string]float64{"temp": 300.0}},
			{Value: 2.0, Conditions: map[string]float64{"temp": 310.0}},
			{Value: 3.0, Conditions: map[string]float64{"temp": 4.0}},
			{Value: 4.0, Conditions: map[string]float64{"temp": 305.0}},
		},
	}

	filtered := FilterByCondition(dataset, "temp", 300.0, 10.0)
	if len(filtered) != 3 {
		t.Errorf("Expected 3 filtered points, got %d", len(filtered))
	}
}

func TestMean(t *testing.T) {
	values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	mean := Mean(values)
	expected := 3.0

	if math.Abs(mean-expected) > 1e-10 {
		t.Errorf("Mean calculation error: got %f, want %f", mean, expected)
	}
}

func TestStandardDeviation(t *testing.T) {
	values := []float64{2.0, 4.0, 4.0, 4.0, 5.0, 5.0, 7.0, 9.0}
	stdDev := StandardDeviation(values)
	expected := 2.138 // Actual expected value

	if math.Abs(stdDev-expected) > 0.01 {
		t.Errorf("StdDev calculation error: got %f, want ~%f", stdDev, expected)
	}
}

func TestRelativeError(t *testing.T) {
	tests := []struct {
		measured float64
		expected float64
		relError float64
	}{
		{100.0, 100.0, 0.0},
		{105.0, 100.0, 0.05},
		{95.0, 100.0, 0.05},
		{110.0, 100.0, 0.10},
	}

	for _, tt := range tests {
		result := RelativeError(tt.measured, tt.expected)
		if math.Abs(result-tt.relError) > 1e-10 {
			t.Errorf("RelativeError(%f, %f) = %f, want %f", tt.measured, tt.expected, result, tt.relError)
		}
	}
}

func TestWithinTolerance(t *testing.T) {
	tests := []struct {
		measured    float64
		expected    float64
		tolerancePct float64
		within      bool
	}{
		{100.0, 100.0, 5.0, true},
		{105.0, 100.0, 5.0, true},
		{95.0, 100.0, 5.0, true},
		{106.0, 100.0, 5.0, false},
		{94.0, 100.0, 5.0, false},
	}

	for _, tt := range tests {
		result := WithinTolerance(tt.measured, tt.expected, tt.tolerancePct)
		if result != tt.within {
			t.Errorf("WithinTolerance(%f, %f, %f) = %v, want %v",
				tt.measured, tt.expected, tt.tolerancePct, result, tt.within)
		}
	}
}

func TestKolmogorovSmirnovTest(t *testing.T) {
	// Test with identical distributions
	sample1 := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	sample2 := []float64{1.0, 2.0, 3.0, 4.0, 5.0}

	statistic, pValue := KolmogorovSmirnovTest(sample1, sample2)

	if statistic > 0.3 {
		t.Errorf("KS statistic for identical samples too high: %f", statistic)
	}

	if pValue < 0.01 {
		t.Errorf("KS p-value for identical samples too low: %f", pValue)
	}

	// Test with different distributions
	sample3 := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	sample4 := []float64{10.0, 20.0, 30.0, 40.0, 50.0}

	statistic2, pValue2 := KolmogorovSmirnovTest(sample3, sample4)

	if statistic2 < 0.7 {
		t.Errorf("KS statistic for different samples too low: %f", statistic2)
	}

	if pValue2 > 0.1 {
		t.Errorf("KS p-value for different samples too high: %f", pValue2)
	}
}

func TestChiSquaredTest(t *testing.T) {
	observed := []float64{10, 15, 20, 25}
	expected := []float64{12, 13, 21, 24}

	chiSq, df := ChiSquaredTest(observed, expected)

	if df != 3 {
		t.Errorf("Degrees of freedom incorrect: got %d, want 3", df)
	}

	if chiSq <= 0 {
		t.Errorf("Chi-squared statistic should be positive, got %f", chiSq)
	}
}

func TestMeanAbsoluteError(t *testing.T) {
	measured := []float64{1.0, 2.0, 3.0}
	expected := []float64{1.5, 2.5, 2.5}

	mae := MeanAbsoluteError(measured, expected)
	expectedMAE := 0.5

	if math.Abs(mae-expectedMAE) > 1e-10 {
		t.Errorf("MAE calculation error: got %f, want %f", mae, expectedMAE)
	}
}

func TestRootMeanSquaredError(t *testing.T) {
	measured := []float64{1.0, 2.0, 3.0}
	expected := []float64{1.0, 2.0, 3.0}

	rmse := RootMeanSquaredError(measured, expected)

	if rmse > 1e-10 {
		t.Errorf("RMSE for identical samples should be ~0, got %f", rmse)
	}
}
