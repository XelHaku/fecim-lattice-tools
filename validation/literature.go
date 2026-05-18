package validation

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"

	sharedio "fecim-lattice-tools/shared/io"
)

// LoadLiteratureDataset loads a literature dataset from a JSON file.
// Uses shared/io for consistent file handling across the codebase.
func LoadLiteratureDataset(path string) (*LiteratureDataset, error) {
	var dataset LiteratureDataset
	if err := sharedio.LoadJSON(path, &dataset); err != nil {
		return nil, fmt.Errorf("failed to load literature dataset: %w", err)
	}
	return &dataset, nil
}

// SaveLiteratureDataset saves a literature dataset to a JSON file.
// Uses shared/io for consistent file handling across the codebase.
func SaveLiteratureDataset(dataset *LiteratureDataset, path string) error {
	if dataset == nil {
		return fmt.Errorf("dataset must not be nil")
	}
	if err := sharedio.SaveJSON(path, dataset); err != nil {
		return fmt.Errorf("failed to save literature dataset: %w", err)
	}
	return nil
}

// LoadMultipleDatasets loads multiple literature datasets from a directory.
// Both dir and pattern are validated: dir must not be empty and pattern
// must not contain path-traversal sequences.
func LoadMultipleDatasets(dir string, pattern string) ([]*LiteratureDataset, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, fmt.Errorf("directory must not be empty")
	}
	if strings.TrimSpace(pattern) == "" {
		return nil, fmt.Errorf("pattern must not be empty")
	}
	// Reject patterns that could escape the directory.
	cleanPattern := filepath.Clean(pattern)
	if cleanPattern == ".." || strings.HasPrefix(cleanPattern, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("pattern contains path traversal: %q", pattern)
	}

	matches, err := filepath.Glob(filepath.Join(dir, cleanPattern))
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
