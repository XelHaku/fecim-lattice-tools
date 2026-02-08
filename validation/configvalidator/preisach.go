package configvalidator

import (
	"fmt"
	"math"
)

// PreisachStateConfig represents a Preisach hysteron state configuration.
type PreisachStateConfig struct {
	Version          int       `json:"version"`
	Material         string    `json:"material"`
	TemperatureK     int       `json:"temperature_k"`
	GridSize         int       `json:"grid_size"`
	DistributionType string    `json:"distribution_type"`
	HysteronStates   []int     `json:"hysteron_states"`
	AlphaMean        float64   `json:"alpha_mean"`
	AlphaSigma       float64   `json:"alpha_sigma"`
	BetaMean         float64   `json:"beta_mean"`
	BetaSigma        float64   `json:"beta_sigma"`
	Correlation      float64   `json:"correlation,omitempty"`
	LorentzAlphaC    float64   `json:"lorentz_alpha_c,omitempty"`
	LorentzAlphaW    float64   `json:"lorentz_alpha_w,omitempty"`
	LorentzBetaC     float64   `json:"lorentz_beta_c,omitempty"`
	LorentzBetaW     float64   `json:"lorentz_beta_w,omitempty"`
	CycleCount       int       `json:"cycle_count,omitempty"`
	CurrentWakeup    float64   `json:"current_wakeup,omitempty"`
	Timestamp        string    `json:"timestamp,omitempty"`
	NumStates        int       `json:"num_states,omitempty"`
}

// Preisach config constraints
const (
	MinPreisachVersion   = 1
	MaxPreisachVersion   = 2
	MinGridSize          = 1
	MaxGridSize          = 1000
	MinCorrelation       = -1.0
	MaxCorrelation       = 1.0
	MinWakeup            = 0.0
	MaxWakeup            = 1.0
)

// Valid distribution types
var validDistributionTypes = map[string]bool{
	"gaussian":    true,
	"lorentzian":  true,
	"bimodal":     true,
	"uniform":     true,
}

// validatePreisachConfig validates a Preisach state configuration.
func validatePreisachConfig(data map[string]any, result *ValidationResult) {
	// Validate version
	version, ok := getInt(data, "version")
	if !ok {
		result.AddError("version", "required field missing or invalid type", data["version"])
	} else {
		if version < MinPreisachVersion || version > MaxPreisachVersion {
			result.AddError("version", fmt.Sprintf("must be between %d and %d", MinPreisachVersion, MaxPreisachVersion), version)
		}
	}
	
	// Validate material
	material, ok := getString(data, "material")
	if !ok {
		result.AddError("material", "required field missing or invalid type", data["material"])
	} else if material == "" {
		result.AddError("material", "must not be empty", material)
	}
	
	// Validate temperature_k
	tempK, ok := getInt(data, "temperature_k")
	if !ok {
		result.AddError("temperature_k", "required field missing or invalid type", data["temperature_k"])
	} else {
		if tempK < MinTemperatureK || tempK > MaxTemperatureK {
			result.AddError("temperature_k", fmt.Sprintf("must be between %d and %d K", MinTemperatureK, MaxTemperatureK), tempK)
		}
	}
	
	// Validate grid_size
	gridSize, ok := getInt(data, "grid_size")
	if !ok {
		result.AddError("grid_size", "required field missing or invalid type", data["grid_size"])
		gridSize = 0
	} else {
		if gridSize < MinGridSize {
			result.AddError("grid_size", fmt.Sprintf("must be at least %d", MinGridSize), gridSize)
		}
		if gridSize > MaxGridSize {
			result.AddError("grid_size", fmt.Sprintf("must be at most %d", MaxGridSize), gridSize)
		}
	}
	
	// Validate distribution_type
	distType, ok := getString(data, "distribution_type")
	if !ok {
		result.AddError("distribution_type", "required field missing or invalid type", data["distribution_type"])
	} else {
		if !validDistributionTypes[distType] {
			result.AddError("distribution_type", "must be one of: gaussian, lorentzian, bimodal, uniform", distType)
		}
	}
	
	// Validate hysteron_states
	hysteronStates, ok := getIntArray(data, "hysteron_states")
	if !ok {
		result.AddError("hysteron_states", "required field missing or invalid type", nil)
	} else {
		// Calculate expected size (triangular number for Preisach grid)
		expectedSize := gridSize * (gridSize + 1) / 2
		if gridSize > 0 && len(hysteronStates) != expectedSize {
			result.AddWarning("hysteron_states", fmt.Sprintf("length (%d) does not match expected triangular number for grid_size %d (%d)", len(hysteronStates), gridSize, expectedSize), len(hysteronStates))
		}
		
		// Validate each hysteron state is -1 or 1
		for i, state := range hysteronStates {
			if state != -1 && state != 1 {
				result.AddError(fmt.Sprintf("hysteron_states[%d]", i), "must be -1 or 1", state)
			}
		}
	}
	
	// Validate distribution parameters
	validatePreisachDistributionParams(data, result)
	
	// Validate optional fields
	validatePreisachOptionalFields(data, result)
}

// validatePreisachDistributionParams validates the distribution parameters.
func validatePreisachDistributionParams(data map[string]any, result *ValidationResult) {
	// Validate alpha_mean (required)
	alphaMean, ok := getFloat(data, "alpha_mean")
	if !ok {
		result.AddError("alpha_mean", "required field missing or invalid type", data["alpha_mean"])
	} else {
		if math.IsNaN(alphaMean) || math.IsInf(alphaMean, 0) {
			result.AddError("alpha_mean", "must be a finite number", alphaMean)
		}
	}
	
	// Validate alpha_sigma (required, must be positive)
	alphaSigma, ok := getFloat(data, "alpha_sigma")
	if !ok {
		result.AddError("alpha_sigma", "required field missing or invalid type", data["alpha_sigma"])
	} else {
		if alphaSigma <= 0 {
			result.AddError("alpha_sigma", "must be positive", alphaSigma)
		}
		if math.IsNaN(alphaSigma) || math.IsInf(alphaSigma, 0) {
			result.AddError("alpha_sigma", "must be a finite number", alphaSigma)
		}
	}
	
	// Validate beta_mean (required)
	betaMean, ok := getFloat(data, "beta_mean")
	if !ok {
		result.AddError("beta_mean", "required field missing or invalid type", data["beta_mean"])
	} else {
		if math.IsNaN(betaMean) || math.IsInf(betaMean, 0) {
			result.AddError("beta_mean", "must be a finite number", betaMean)
		}
	}
	
	// Validate beta_sigma (required, must be positive)
	betaSigma, ok := getFloat(data, "beta_sigma")
	if !ok {
		result.AddError("beta_sigma", "required field missing or invalid type", data["beta_sigma"])
	} else {
		if betaSigma <= 0 {
			result.AddError("beta_sigma", "must be positive", betaSigma)
		}
		if math.IsNaN(betaSigma) || math.IsInf(betaSigma, 0) {
			result.AddError("beta_sigma", "must be a finite number", betaSigma)
		}
	}
	
	// Physical constraint: alpha should be >= beta (switching up >= switching down)
	if ok1, ok2 := (alphaMean > 0), (betaMean < 0); ok1 && ok2 {
		// This is the expected case for ferroelectrics
	} else {
		result.AddWarning("alpha_mean/beta_mean", "typically alpha_mean > 0 and beta_mean < 0 for ferroelectric materials", nil)
	}
}

// validatePreisachOptionalFields validates optional Preisach config fields.
func validatePreisachOptionalFields(data map[string]any, result *ValidationResult) {
	// Validate correlation if present
	if corr, ok := getFloat(data, "correlation"); ok {
		if corr < MinCorrelation || corr > MaxCorrelation {
			result.AddError("correlation", fmt.Sprintf("must be between %.1f and %.1f", MinCorrelation, MaxCorrelation), corr)
		}
	}
	
	// Validate Lorentzian parameters if present
	lorentzParams := []string{"lorentz_alpha_c", "lorentz_alpha_w", "lorentz_beta_c", "lorentz_beta_w"}
	for _, param := range lorentzParams {
		if v, ok := getFloat(data, param); ok {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				result.AddError(param, "must be a finite number", v)
			}
			// Width parameters should be positive
			if (param == "lorentz_alpha_w" || param == "lorentz_beta_w") && v <= 0 {
				result.AddError(param, "width must be positive", v)
			}
		}
	}
	
	// Validate cycle_count if present
	if cycleCount, ok := getInt(data, "cycle_count"); ok {
		if cycleCount < 0 {
			result.AddError("cycle_count", "must be non-negative", cycleCount)
		}
	}
	
	// Validate current_wakeup if present
	if wakeup, ok := getFloat(data, "current_wakeup"); ok {
		if wakeup < MinWakeup || wakeup > MaxWakeup {
			result.AddError("current_wakeup", fmt.Sprintf("must be between %.1f and %.1f", MinWakeup, MaxWakeup), wakeup)
		}
	}
	
	// Validate timestamp if present
	if ts, ok := getString(data, "timestamp"); ok && ts != "" {
		if !validateTimestamp(ts) {
			result.AddWarning("timestamp", "invalid ISO 8601 timestamp format", ts)
		}
	}
	
	// Validate num_states if present
	if numStates, ok := getInt(data, "num_states"); ok {
		if numStates < 0 {
			result.AddError("num_states", "must be non-negative", numStates)
		}
	}
}
