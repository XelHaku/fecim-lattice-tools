package validation

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	sharedval "fecim-lattice-tools/shared/validation"
)

const (
	senseRegressionDir   = "output/regression/module4"
	senseGoldenFile      = "sense_chain_4x4.json"
	updateSenseGoldenEnv = "FECIM_UPDATE_SENSE_GOLDEN"
)

type senseChainGolden struct {
	Version     string                 `json:"version"`
	Scenario    string                 `json:"scenario"`
	Description string                 `json:"description"`
	Generated   string                 `json:"generated_utc"`
	Parameters  map[string]interface{} `json:"parameters"`
	Results     senseChainResults      `json:"results"`
}

type senseChainResults struct {
	RowCurrents     []float64 `json:"row_currents"`
	TIAVoltages     []float64 `json:"tia_voltages"`
	ADCCodes        []int     `json:"adc_codes"`
	CurrentRangeMin float64   `json:"current_range_min"`
	CurrentRangeMax float64   `json:"current_range_max"`
	CurrentLSB      float64   `json:"current_lsb"`
}

// TestSenseChainRegression_4x4 is a golden regression test for the sense chain
// (TIA + ADC) using a fixed 4×4 conductance matrix and fixed input voltages.
// Regenerate golden file with: FECIM_UPDATE_SENSE_GOLDEN=1 go test ./validation/... -run TestSenseChainRegression_4x4
func TestSenseChainRegression_4x4(t *testing.T) {
	// Fixed conductance matrix (Siemens).
	G := [4][4]float64{
		{1e-6, 5e-6, 10e-6, 20e-6},
		{2e-6, 8e-6, 15e-6, 25e-6},
		{3e-6, 12e-6, 18e-6, 30e-6},
		{4e-6, 6e-6, 22e-6, 35e-6},
	}
	inputV := [4]float64{0.1, 0.3, 0.5, 0.7}

	// Compute MVM: I[r] = sum_c(inputV[c] * G[r][c]).
	rowCurrents := make([]float64, 4)
	for r := 0; r < 4; r++ {
		sum := 0.0
		for c := 0; c < 4; c++ {
			sum += inputV[c] * G[r][c]
		}
		rowCurrents[r] = sum
	}

	// Standard sense chain config.
	sense := arraysim.SenseChain{
		TIA: arraysim.TIAConfig{Rf: 10e3, Vref: 0.5, Vmin: 0, Vmax: 1.0},
		ADC: arraysim.ADCConfig{Bits: 4, Vmin: 0.0, Vmax: 1.0},
	}

	// Process each row current through the sense chain.
	tiaVoltages := make([]float64, 4)
	adcCodes := make([]int, 4)
	for r, I := range rowCurrents {
		result := sense.ConvertCurrent(I)
		tiaVoltages[r] = result.Vout
		adcCodes[r] = result.Code
	}

	iMin, iMax := sense.CurrentRange()
	iLSB := sense.CurrentLSB()

	results := senseChainResults{
		RowCurrents:     rowCurrents,
		TIAVoltages:     tiaVoltages,
		ADCCodes:        adcCodes,
		CurrentRangeMin: iMin,
		CurrentRangeMax: iMax,
		CurrentLSB:      iLSB,
	}

	params := map[string]interface{}{
		"rows":    4,
		"cols":    4,
		"tia_rf":  sense.TIA.Rf,
		"tia_vref": sense.TIA.Vref,
		"tia_vmin": sense.TIA.Vmin,
		"tia_vmax": sense.TIA.Vmax,
		"adc_bits": sense.ADC.Bits,
		"adc_vmin": sense.ADC.Vmin,
		"adc_vmax": sense.ADC.Vmax,
	}

	goldenPath := filepath.Join(senseRegressionDir, senseGoldenFile)

	if os.Getenv(updateSenseGoldenEnv) != "" {
		if err := os.MkdirAll(senseRegressionDir, 0o755); err != nil {
			t.Fatalf("create regression dir: %v", err)
		}
		golden := senseChainGolden{
			Version:     "v1",
			Scenario:    "sense_chain_4x4",
			Description: "Golden sense chain regression: 4x4 MVM through TIA+ADC with fixed conductance matrix and input voltages.",
			Generated:   sharedval.NewEnvelope("", "", true).TimestampUTC,
			Parameters:  params,
			Results:     results,
		}
		b, err := json.MarshalIndent(golden, "", "  ")
		if err != nil {
			t.Fatalf("marshal golden: %v", err)
		}
		b = append(b, '\n')
		if err := os.WriteFile(goldenPath, b, 0o644); err != nil {
			t.Fatalf("write golden file: %v", err)
		}
		t.Logf("updated golden: %s (unset %s to run regression)", goldenPath, updateSenseGoldenEnv)
		return
	}

	// Read and compare against golden.
	f, err := os.Open(goldenPath)
	if err != nil {
		t.Fatalf("open golden file %s: %v (run with %s=1 to generate)", goldenPath, err, updateSenseGoldenEnv)
	}
	defer f.Close()

	var golden senseChainGolden
	if err := json.NewDecoder(f).Decode(&golden); err != nil {
		t.Fatalf("decode golden file: %v", err)
	}

	const tol = 1e-12

	// Compare row currents.
	if len(golden.Results.RowCurrents) != len(rowCurrents) {
		t.Fatalf("row currents length mismatch: got %d want %d", len(rowCurrents), len(golden.Results.RowCurrents))
	}
	for i, got := range rowCurrents {
		want := golden.Results.RowCurrents[i]
		if math.Abs(got-want) > tol {
			t.Errorf("row_currents[%d]: got %.15e want %.15e (diff %.2e)", i, got, want, math.Abs(got-want))
		}
	}

	// Compare TIA voltages.
	if len(golden.Results.TIAVoltages) != len(tiaVoltages) {
		t.Fatalf("tia_voltages length mismatch: got %d want %d", len(tiaVoltages), len(golden.Results.TIAVoltages))
	}
	for i, got := range tiaVoltages {
		want := golden.Results.TIAVoltages[i]
		if math.Abs(got-want) > tol {
			t.Errorf("tia_voltages[%d]: got %.15e want %.15e (diff %.2e)", i, got, want, math.Abs(got-want))
		}
	}

	// Compare ADC codes (exact).
	if len(golden.Results.ADCCodes) != len(adcCodes) {
		t.Fatalf("adc_codes length mismatch: got %d want %d", len(adcCodes), len(golden.Results.ADCCodes))
	}
	for i, got := range adcCodes {
		want := golden.Results.ADCCodes[i]
		if got != want {
			t.Errorf("adc_codes[%d]: got %d want %d", i, got, want)
		}
	}

	// Compare current range and LSB.
	if math.Abs(iMin-golden.Results.CurrentRangeMin) > tol {
		t.Errorf("current_range_min: got %.15e want %.15e", iMin, golden.Results.CurrentRangeMin)
	}
	if math.Abs(iMax-golden.Results.CurrentRangeMax) > tol {
		t.Errorf("current_range_max: got %.15e want %.15e", iMax, golden.Results.CurrentRangeMax)
	}
	if math.Abs(iLSB-golden.Results.CurrentLSB) > tol {
		t.Errorf("current_lsb: got %.15e want %.15e", iLSB, golden.Results.CurrentLSB)
	}

	t.Logf("sense chain regression PASS: rowCurrents=%v adcCodes=%v iMin=%.3e iMax=%.3e iLSB=%.3e",
		rowCurrents, adcCodes, iMin, iMax, iLSB)
}
