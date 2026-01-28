package ferroelectric

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// GoldenLoopData represents the stored golden reference for hysteresis loops.
type GoldenLoopData struct {
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Generated   string                 `json:"generated"`
	Material    string                 `json:"material"`
	Parameters  map[string]interface{} `json:"parameters"`
	Data        struct {
		E []float64 `json:"E"`
		P []float64 `json:"P"`
	} `json:"data"`
}

// GoldenTempSweep represents the stored temperature sweep data.
type GoldenTempSweep struct {
	Version       string    `json:"version"`
	Description   string    `json:"description"`
	TemperaturesK []float64 `json:"temperatures_K"`
	EcValues      []float64 `json:"Ec_values"`
	PrValues      []float64 `json:"pr_values"`
}

// Golden30States represents the stored 30 discrete analog states.
type Golden30States struct {
	Version     string    `json:"version"`
	Description string    `json:"description"`
	States      []float64 `json:"states"`
}

// calculateRMS computes the root-mean-square error between two slices.
func calculateRMS(a, b []float64) float64 {
	if len(a) != len(b) {
		return math.Inf(1)
	}

	sum := 0.0
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}

	return math.Sqrt(sum / float64(len(a)))
}

// maxAbsError computes the maximum absolute error between two slices.
func maxAbsError(a, b []float64) float64 {
	if len(a) != len(b) {
		return math.Inf(1)
	}

	maxErr := 0.0
	for i := range a {
		err := math.Abs(a[i] - b[i])
		if err > maxErr {
			maxErr = err
		}
	}

	return maxErr
}

// TestGoldenHysteresisLoopData compares generated loops against stored golden reference.
// This test ensures numerical stability across platforms and Go versions.
func TestGoldenHysteresisLoopData(t *testing.T) {
	// Load golden reference
	goldenPath := filepath.Join("testdata", "golden_loop_default_hzo.json")
	f, err := os.Open(goldenPath)
	if err != nil {
		t.Fatalf("Failed to open golden reference: %v", err)
	}
	defer f.Close()

	var golden GoldenLoopData
	if err := json.NewDecoder(f).Decode(&golden); err != nil {
		t.Fatalf("Failed to decode golden reference: %v", err)
	}

	// Generate fresh loop with same parameters
	mat := DefaultHZO()
	model := NewPreisachModel(mat)

	Emax := 2.0 * mat.Ec
	points := 100
	E, P := model.GetHysteresisLoop(Emax, points)

	// Verify array lengths match
	if len(E) != len(golden.Data.E) {
		t.Errorf("E array length mismatch: got %d, expected %d", len(E), len(golden.Data.E))
	}
	if len(P) != len(golden.Data.P) {
		t.Errorf("P array length mismatch: got %d, expected %d", len(P), len(golden.Data.P))
	}

	// Calculate RMS error for E field
	rmsE := calculateRMS(E, golden.Data.E)
	toleranceE := 0.02 * Emax // 2% of max field
	if rmsE > toleranceE {
		t.Errorf("E field RMS error too large: %e (tolerance: %e)", rmsE, toleranceE)
	}

	// Calculate RMS error for polarization
	rmsP := calculateRMS(P, golden.Data.P)
	toleranceP := 0.02 * mat.Ps // 2% of saturation polarization
	if rmsP > toleranceP {
		t.Errorf("Polarization RMS error too large: %e (tolerance: %e)", rmsP, toleranceP)
	}

	t.Logf("Golden loop validation passed: RMS_E=%e, RMS_P=%e", rmsE, rmsP)
}

// TestGoldenTemperatureSweepData compares Ec(T) and Pr(T) against stored golden reference.
// This ensures temperature-dependent physics models remain stable.
func TestGoldenTemperatureSweepData(t *testing.T) {
	// Load golden reference
	goldenPath := filepath.Join("testdata", "golden_temp_sweep.json")
	f, err := os.Open(goldenPath)
	if err != nil {
		t.Fatalf("Failed to open golden reference: %v", err)
	}
	defer f.Close()

	var golden GoldenTempSweep
	if err := json.NewDecoder(f).Decode(&golden); err != nil {
		t.Fatalf("Failed to decode golden reference: %v", err)
	}

	// Generate fresh temperature sweep
	mat := DefaultHZO()

	ecVals := make([]float64, len(golden.TemperaturesK))
	prVals := make([]float64, len(golden.TemperaturesK))

	for i, T := range golden.TemperaturesK {
		ecVals[i] = mat.CoerciveFieldAtTemp(T)
		prVals[i] = mat.PolarizationAtTemp(T)
	}

	// Calculate maximum error for Ec(T)
	maxEcErr := maxAbsError(ecVals, golden.EcValues)
	toleranceEc := 0.02 * mat.Ec // 2% of room temperature Ec

	// Calculate maximum error for Pr(T)
	maxPrErr := maxAbsError(prVals, golden.PrValues)
	tolerancePr := 0.02 * mat.Pr // 2% of room temperature Pr

	// Check automotive temperature range (233K to 423K) more strictly
	automotiveStart := -1
	automotiveEnd := -1
	for i, T := range golden.TemperaturesK {
		if T >= 233 && automotiveStart == -1 {
			automotiveStart = i
		}
		if T <= 423 {
			automotiveEnd = i
		}
	}

	if automotiveStart >= 0 && automotiveEnd >= 0 {
		automotiveEcErr := maxAbsError(
			ecVals[automotiveStart:automotiveEnd+1],
			golden.EcValues[automotiveStart:automotiveEnd+1],
		)
		automotivePrErr := maxAbsError(
			prVals[automotiveStart:automotiveEnd+1],
			golden.PrValues[automotiveStart:automotiveEnd+1],
		)

		// Automotive range should be even more accurate
		automotiveTolEc := 0.01 * mat.Ec // 1% for automotive
		automotiveTolPr := 0.01 * mat.Pr

		if automotiveEcErr > automotiveTolEc {
			t.Errorf("Automotive Ec(T) error too large: %e (tolerance: %e)", automotiveEcErr, automotiveTolEc)
		}
		if automotivePrErr > automotiveTolPr {
			t.Errorf("Automotive Pr(T) error too large: %e (tolerance: %e)", automotivePrErr, automotiveTolPr)
		}

		t.Logf("Automotive range validation passed: Ec_err=%e, Pr_err=%e", automotiveEcErr, automotivePrErr)
	}

	// Full range validation
	if maxEcErr > toleranceEc {
		t.Errorf("Ec(T) max error too large: %e (tolerance: %e)", maxEcErr, toleranceEc)
	}
	if maxPrErr > tolerancePr {
		t.Errorf("Pr(T) max error too large: %e (tolerance: %e)", maxPrErr, tolerancePr)
	}

	t.Logf("Full range validation passed: Ec_max_err=%e, Pr_max_err=%e", maxEcErr, maxPrErr)
}

// TestGolden30StateQuantization compares discrete state values against stored golden reference.
// This ensures the 30-state FeCIM capability remains accurate.
func TestGolden30StateQuantization(t *testing.T) {
	// Load golden reference
	goldenPath := filepath.Join("testdata", "golden_30_states.json")
	f, err := os.Open(goldenPath)
	if err != nil {
		t.Fatalf("Failed to open golden reference: %v", err)
	}
	defer f.Close()

	var golden Golden30States
	if err := json.NewDecoder(f).Decode(&golden); err != nil {
		t.Fatalf("Failed to decode golden reference: %v", err)
	}

	// Generate fresh 30 states
	mat := DefaultHZO()
	model := NewPreisachModel(mat)
	states := model.DiscreteStates(30)

	// Verify we have 30 states
	if len(states) != 30 {
		t.Errorf("Expected 30 states, got %d", len(states))
	}

	if len(golden.States) != 30 {
		t.Errorf("Golden reference has %d states, expected 30", len(golden.States))
	}

	// All states should match exactly (within floating point precision)
	tolerance := 1e-10 // Very tight tolerance for discrete states
	for i := 0; i < len(states) && i < len(golden.States); i++ {
		diff := math.Abs(states[i] - golden.States[i])
		if diff > tolerance {
			t.Errorf("State %d mismatch: got %e, expected %e (diff: %e)",
				i, states[i], golden.States[i], diff)
		}
	}

	// Verify states span from -Ps to +Ps
	if math.Abs(states[0]-(-mat.Ps)) > tolerance {
		t.Errorf("First state should be -Ps: got %e, expected %e", states[0], -mat.Ps)
	}
	if math.Abs(states[29]-mat.Ps) > tolerance {
		t.Errorf("Last state should be +Ps: got %e, expected %e", states[29], mat.Ps)
	}

	// Verify uniform spacing
	expectedSpacing := 2 * mat.Ps / 29 // (Ps - (-Ps)) / (30 - 1)
	for i := 0; i < len(states)-1; i++ {
		spacing := states[i+1] - states[i]
		if math.Abs(spacing-expectedSpacing) > tolerance {
			t.Errorf("Non-uniform spacing at index %d: got %e, expected %e",
				i, spacing, expectedSpacing)
		}
	}

	t.Logf("30-state quantization validation passed: all states match within %e", tolerance)
}

// TestGoldenDataVersioning ensures golden reference files have proper version metadata.
func TestGoldenDataVersioning(t *testing.T) {
	tests := []struct {
		name          string
		file          string
		expectedVer   string
		checkFields   func(interface{}) error
	}{
		{
			name:        "loop data",
			file:        "golden_loop_default_hzo.json",
			expectedVer: "1.0.0",
			checkFields: func(data interface{}) error {
				d := data.(*GoldenLoopData)
				if d.Material != "DefaultHZO" {
					t.Errorf("Expected material DefaultHZO, got %s", d.Material)
				}
				if len(d.Data.E) == 0 || len(d.Data.P) == 0 {
					t.Error("Empty E or P data")
				}
				return nil
			},
		},
		{
			name:        "temp sweep",
			file:        "golden_temp_sweep.json",
			expectedVer: "1.0.0",
			checkFields: func(data interface{}) error {
				d := data.(*GoldenTempSweep)
				if len(d.TemperaturesK) == 0 {
					t.Error("Empty temperature data")
				}
				if len(d.EcValues) != len(d.TemperaturesK) {
					t.Error("Ec values length mismatch")
				}
				if len(d.PrValues) != len(d.TemperaturesK) {
					t.Error("Pr values length mismatch")
				}
				return nil
			},
		},
		{
			name:        "30 states",
			file:        "golden_30_states.json",
			expectedVer: "1.0.0",
			checkFields: func(data interface{}) error {
				d := data.(*Golden30States)
				if len(d.States) != 30 {
					t.Errorf("Expected 30 states, got %d", len(d.States))
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("testdata", tt.file)
			f, err := os.Open(path)
			if err != nil {
				t.Fatalf("Failed to open %s: %v", tt.file, err)
			}
			defer f.Close()

			// Decode based on file type
			var version string
			switch tt.name {
			case "loop data":
				var d GoldenLoopData
				if err := json.NewDecoder(f).Decode(&d); err != nil {
					t.Fatalf("Failed to decode: %v", err)
				}
				version = d.Version
				tt.checkFields(&d)
			case "temp sweep":
				var d GoldenTempSweep
				if err := json.NewDecoder(f).Decode(&d); err != nil {
					t.Fatalf("Failed to decode: %v", err)
				}
				version = d.Version
				tt.checkFields(&d)
			case "30 states":
				var d Golden30States
				if err := json.NewDecoder(f).Decode(&d); err != nil {
					t.Fatalf("Failed to decode: %v", err)
				}
				version = d.Version
				tt.checkFields(&d)
			}

			if version != tt.expectedVer {
				t.Errorf("Version mismatch: got %s, expected %s", version, tt.expectedVer)
			}
		})
	}
}
