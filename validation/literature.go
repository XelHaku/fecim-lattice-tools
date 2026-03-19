package validation

import (
	"fmt"
	"math"
	"path/filepath"

	"fecim-lattice-tools/shared/io"
)

// LoadLiteratureDataset loads a literature dataset from a JSON file.
// Uses shared/io for consistent file handling across the codebase.
func LoadLiteratureDataset(path string) (*LiteratureDataset, error) {
	var dataset LiteratureDataset
	if err := io.LoadJSON(path, &dataset); err != nil {
		return nil, fmt.Errorf("failed to load literature dataset: %w", err)
	}
	return &dataset, nil
}

// SaveLiteratureDataset saves a literature dataset to a JSON file.
// Uses shared/io for consistent file handling across the codebase.
func SaveLiteratureDataset(dataset *LiteratureDataset, path string) error {
	if err := io.SaveJSON(path, dataset); err != nil {
		return fmt.Errorf("failed to save literature dataset: %w", err)
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
			if math.Abs(val-conditionValue) <= tolerance {
				filtered = append(filtered, dp)
			}
		}
	}
	return filtered
}

