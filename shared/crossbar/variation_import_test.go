package crossbar

import (
	"math"
	"testing"
)

func TestImportCrossSimVariation_DefaultHZO(t *testing.T) {
	csConfig := DefaultCrossSimHZO()
	pvConfig := ImportCrossSimVariation(csConfig)

	if pvConfig == nil {
		t.Fatal("ImportCrossSimVariation returned nil")
	}

	// DeviceSigma should be non-zero and in a reasonable range (0.01-0.15).
	if pvConfig.DeviceSigma <= 0 {
		t.Errorf("DeviceSigma should be positive, got %.6f", pvConfig.DeviceSigma)
	}
	if pvConfig.DeviceSigma > 0.15 {
		t.Errorf("DeviceSigma=%.4f exceeds reasonable range for HZO (expected <0.15)", pvConfig.DeviceSigma)
	}

	// DeviceSigma should be >= max(D2D_HRS, D2D_LRS) since it combines D2D + programming noise.
	maxD2D := math.Max(csConfig.D2DSigmaHRS, csConfig.D2DSigmaLRS)
	if pvConfig.DeviceSigma < maxD2D*0.9 {
		t.Errorf("DeviceSigma=%.4f should be >= ~D2D sigma (%.4f)", pvConfig.DeviceSigma, maxD2D)
	}

	// Gradients should be small positive values.
	if pvConfig.GradientX <= 0 || pvConfig.GradientY <= 0 {
		t.Errorf("gradients should be positive: GradientX=%.6f, GradientY=%.6f", pvConfig.GradientX, pvConfig.GradientY)
	}
	if pvConfig.GradientX > 0.01 {
		t.Errorf("GradientX=%.4f too large (expected <0.01)", pvConfig.GradientX)
	}

	// Edge effect should be positive and bounded.
	if pvConfig.EdgeEffect <= 0 {
		t.Errorf("EdgeEffect should be positive, got %.6f", pvConfig.EdgeEffect)
	}
	if pvConfig.EdgeEffect > 0.20 {
		t.Errorf("EdgeEffect=%.4f exceeds 20%% cap", pvConfig.EdgeEffect)
	}
}

func TestImportCrossSimVariation_RoundTrip(t *testing.T) {
	original := DefaultCrossSimHZO()
	pvConfig := ImportCrossSimVariation(original)
	exported := ExportToCrossSimFormat(pvConfig)

	// Round-trip cannot be exact since the mapping is lossy, but the total
	// variation magnitude should be preserved within a reasonable tolerance.

	// Total imported sigma should approximately match total exported sigma.
	importedTotal := pvConfig.DeviceSigma
	exportedD2D := math.Sqrt(exported.D2DSigmaHRS*exported.D2DSigmaHRS+exported.D2DSigmaLRS*exported.D2DSigmaLRS) / math.Sqrt(2)
	exportedTotal := math.Sqrt(exportedD2D*exportedD2D + exported.ProgramNoiseSigma*exported.ProgramNoiseSigma)

	relDiff := math.Abs(importedTotal-exportedTotal) / importedTotal
	if relDiff > 0.30 {
		t.Errorf("round-trip total sigma mismatch: imported=%.4f exported=%.4f relDiff=%.1f%%",
			importedTotal, exportedTotal, relDiff*100)
	}

	// All exported values should be non-negative.
	if exported.ProgramNoiseSigma < 0 {
		t.Errorf("exported ProgramNoiseSigma negative: %.6f", exported.ProgramNoiseSigma)
	}
	if exported.ReadNoiseSigma < 0 {
		t.Errorf("exported ReadNoiseSigma negative: %.6f", exported.ReadNoiseSigma)
	}
	if exported.D2DSigmaHRS < 0 {
		t.Errorf("exported D2DSigmaHRS negative: %.6f", exported.D2DSigmaHRS)
	}
	if exported.D2DSigmaLRS < 0 {
		t.Errorf("exported D2DSigmaLRS negative: %.6f", exported.D2DSigmaLRS)
	}
	if exported.C2CSigma < 0 {
		t.Errorf("exported C2CSigma negative: %.6f", exported.C2CSigma)
	}
	if exported.DriftCoeff < 0 {
		t.Errorf("exported DriftCoeff negative: %.6f", exported.DriftCoeff)
	}
	if exported.DriftT0 <= 0 {
		t.Errorf("exported DriftT0 must be positive: %.6f", exported.DriftT0)
	}

	// HRS should have >= LRS variation (asymmetric split).
	if exported.D2DSigmaHRS < exported.D2DSigmaLRS {
		t.Errorf("exported D2D_HRS=%.4f should be >= D2D_LRS=%.4f", exported.D2DSigmaHRS, exported.D2DSigmaLRS)
	}
}

func TestImportCrossSimVariation_AppliesToArray(t *testing.T) {
	csConfig := DefaultCrossSimHZO()
	pvConfig := ImportCrossSimVariation(csConfig)

	// Create two arrays: one with no variation, one with imported CrossSim variation.
	cfgClean := &Config{
		Rows:    4,
		Cols:    4,
		ADCBits: 8,
		DACBits: 8,
	}
	cfgVaried := &Config{
		Rows:             4,
		Cols:             4,
		ADCBits:          8,
		DACBits:          8,
		ProcessVariation: pvConfig,
	}

	arrClean, err := NewArray(cfgClean)
	if err != nil {
		t.Fatalf("failed to create clean array: %v", err)
	}
	arrVaried, err := NewArray(cfgVaried)
	if err != nil {
		t.Fatalf("failed to create varied array: %v", err)
	}

	// Program identical weights.
	weights := [][]float64{
		{0.5, 0.5, 0.5, 0.5},
		{0.5, 0.5, 0.5, 0.5},
		{0.5, 0.5, 0.5, 0.5},
		{0.5, 0.5, 0.5, 0.5},
	}
	if err := arrClean.ProgramWeightMatrix(weights); err != nil {
		t.Fatalf("clean ProgramWeightMatrix: %v", err)
	}
	if err := arrVaried.ProgramWeightMatrix(weights); err != nil {
		t.Fatalf("varied ProgramWeightMatrix: %v", err)
	}

	// Run MVM with identical inputs.
	input := []float64{1.0, 1.0, 1.0, 1.0}
	outClean, err := arrClean.MVM(input)
	if err != nil {
		t.Fatalf("clean MVM: %v", err)
	}
	outVaried, err := arrVaried.MVM(input)
	if err != nil {
		t.Fatalf("varied MVM: %v", err)
	}

	if len(outClean) != len(outVaried) {
		t.Fatalf("output length mismatch: clean=%d varied=%d", len(outClean), len(outVaried))
	}

	// With variation applied, at least some outputs should differ.
	// (If all outputs are identical, the variation is not being applied.)
	anyDiff := false
	for i := range outClean {
		if math.Abs(outClean[i]-outVaried[i]) > 1e-12 {
			anyDiff = true
			break
		}
	}
	if !anyDiff {
		t.Error("MVM outputs are identical with and without CrossSim variation; expected some difference")
	}
}

func TestDefaultCrossSimHZO_ReasonableValues(t *testing.T) {
	cs := DefaultCrossSimHZO()

	// All sigma values should be in the range [0.01, 0.10] for HZO FeFETs.
	checks := []struct {
		name string
		val  float64
	}{
		{"ProgramNoiseSigma", cs.ProgramNoiseSigma},
		{"ReadNoiseSigma", cs.ReadNoiseSigma},
		{"D2DSigmaHRS", cs.D2DSigmaHRS},
		{"D2DSigmaLRS", cs.D2DSigmaLRS},
		{"C2CSigma", cs.C2CSigma},
	}

	for _, c := range checks {
		if c.val < 0.005 || c.val > 0.15 {
			t.Errorf("%s=%.4f outside expected range [0.005, 0.15]", c.name, c.val)
		}
	}

	// Drift coefficient should be small positive.
	if cs.DriftCoeff <= 0 || cs.DriftCoeff > 0.20 {
		t.Errorf("DriftCoeff=%.4f outside expected range (0, 0.20]", cs.DriftCoeff)
	}

	// Reference time should be positive.
	if cs.DriftT0 <= 0 {
		t.Errorf("DriftT0 must be positive, got %.4f", cs.DriftT0)
	}

	// D2D at HRS should be >= D2D at LRS (wider distribution at high resistance).
	if cs.D2DSigmaHRS < cs.D2DSigmaLRS {
		t.Errorf("D2D_HRS=%.4f should be >= D2D_LRS=%.4f for HZO", cs.D2DSigmaHRS, cs.D2DSigmaLRS)
	}
}

func TestImportCrossSimVariation_ZeroConfig(t *testing.T) {
	// Zero config should produce zero/minimal variation.
	pvConfig := ImportCrossSimVariation(CrossSimVariationConfig{})

	if pvConfig == nil {
		t.Fatal("ImportCrossSimVariation returned nil for zero config")
	}
	if pvConfig.DeviceSigma != 0 {
		t.Errorf("DeviceSigma should be 0 for zero input, got %.6f", pvConfig.DeviceSigma)
	}
	if pvConfig.GradientX != 0 {
		t.Errorf("GradientX should be 0 for zero input, got %.6f", pvConfig.GradientX)
	}
	if pvConfig.GradientY != 0 {
		t.Errorf("GradientY should be 0 for zero input, got %.6f", pvConfig.GradientY)
	}
	if pvConfig.EdgeEffect != 0 {
		t.Errorf("EdgeEffect should be 0 for zero input, got %.6f", pvConfig.EdgeEffect)
	}
}

func TestExportToCrossSimFormat_NilConfig(t *testing.T) {
	exported := ExportToCrossSimFormat(nil)

	if exported.ProgramNoiseSigma != 0 {
		t.Errorf("expected zero ProgramNoiseSigma for nil input, got %.6f", exported.ProgramNoiseSigma)
	}
	if exported.D2DSigmaHRS != 0 {
		t.Errorf("expected zero D2DSigmaHRS for nil input, got %.6f", exported.D2DSigmaHRS)
	}
}
