package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LoadLiteratureDataset loads a literature dataset from a JSON file
func LoadLiteratureDataset(path string) (*LiteratureDataset, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read literature dataset: %w", err)
	}

	var dataset LiteratureDataset
	if err := json.Unmarshal(data, &dataset); err != nil {
		return nil, fmt.Errorf("failed to parse literature dataset: %w", err)
	}

	return &dataset, nil
}

// SaveLiteratureDataset saves a literature dataset to a JSON file
func SaveLiteratureDataset(dataset *LiteratureDataset, path string) error {
	data, err := json.MarshalIndent(dataset, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal dataset: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write dataset: %w", err)
	}

	return nil
}

// LoadMultipleDatasets loads multiple literature datasets from a directory
func LoadMultipleDatasets(dir string, pattern string) ([]*LiteratureDataset, error) {
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return nil, fmt.Errorf("failed to glob pattern: %w", err)
	}

	datasets := make([]*LiteratureDataset, 0, len(matches))
	for _, match := range matches {
		dataset, err := LoadLiteratureDataset(match)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", match, err)
		}
		datasets = append(datasets, dataset)
	}

	return datasets, nil
}

// ExtractValues extracts raw values from a dataset for statistical analysis
func ExtractValues(dataset *LiteratureDataset) []float64 {
	values := make([]float64, len(dataset.DataPoints))
	for i, dp := range dataset.DataPoints {
		values[i] = dp.Value
	}
	return values
}

// ExtractUncertainties extracts uncertainties from a dataset
func ExtractUncertainties(dataset *LiteratureDataset) []float64 {
	uncertainties := make([]float64, len(dataset.DataPoints))
	for i, dp := range dataset.DataPoints {
		uncertainties[i] = dp.Uncertainty
	}
	return uncertainties
}

// FilterByCondition filters data points by a specific condition
func FilterByCondition(dataset *LiteratureDataset, conditionKey string, conditionValue float64, tolerance float64) []LiteratureDataPoint {
	filtered := []LiteratureDataPoint{}
	for _, dp := range dataset.DataPoints {
		if val, ok := dp.Conditions[conditionKey]; ok {
			if abs(val-conditionValue) <= tolerance {
				filtered = append(filtered, dp)
			}
		}
	}
	return filtered
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
