package external_test

import (
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"fecim-lattice-tools/validation/external/internal/testsupport"
)

// TestBrevitasQATComparison validates Go PTQ quantization against Brevitas QAT.
//
// The Python script (scripts/brevitas_qat_compare.py) trains an MNIST MLP and
// compares three quantization approaches:
//   - Full-precision (FP32) baseline
//   - Post-training quantization (PTQ) at 30 levels (same algorithm as Go)
//   - Quantization-aware training (QAT) at 5-bit weights (32 levels ~ 30)
//
// Expected result: QAT accuracy >= PTQ accuracy (QAT adapts weights during training).
// This test skips gracefully if python3, torch, or brevitas are unavailable.
func TestBrevitasQATComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Brevitas QAT comparison in short mode (requires training)")
	}

	testsupport.RequireCommand(t, "python3", "python3 not installed")

	// Check torch
	check := exec.Command("python3", "-c", "import torch; print(torch.__version__)")
	if out, err := check.CombinedOutput(); err != nil {
		t.Skipf("PyTorch not installed: %v (run: pip install torch torchvision)", err)
	} else {
		t.Logf("PyTorch version: %s", string(out))
	}

	// Check brevitas
	check = exec.Command("python3", "-c", "import brevitas; print(brevitas.__version__)")
	if out, err := check.CombinedOutput(); err != nil {
		t.Skipf("Brevitas not installed: %v (run: pip install brevitas)", err)
	} else {
		t.Logf("Brevitas version: %s", string(out))
	}

	projectRoot := testsupport.ProjectRoot(t)
	scriptPath := filepath.Join(projectRoot, "scripts", "brevitas_qat_compare.py")

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Fatalf("script not found: %s", scriptPath)
	}

	t.Logf("Running Brevitas QAT comparison script: %s", scriptPath)
	t.Log("This may take a few minutes (MNIST training)...")

	cmd := exec.Command("python3", scriptPath)
	cmd.Dir = projectRoot
	outBytes, err := cmd.Output()
	if err != nil {
		// Try to get stderr for diagnostics
		if exitErr, ok := err.(*exec.ExitError); ok {
			t.Fatalf("script failed: %v\nstderr: %s", err, string(exitErr.Stderr))
		}
		t.Fatalf("script failed: %v", err)
	}

	// Parse JSON output
	var result struct {
		FPAccuracy    float64 `json:"fp_accuracy"`
		PTQAccuracy   float64 `json:"ptq_accuracy"`
		QATAccuracy   float64 `json:"qat_accuracy"`
		PTQLevels     int     `json:"ptq_levels"`
		QATBitWidth   int     `json:"qat_bit_width"`
		QATLevels     int     `json:"qat_levels"`
		DeltaQATvsPTQ float64 `json:"accuracy_delta_qat_vs_ptq"`
		DeltaPTQvsFP  float64 `json:"accuracy_delta_ptq_vs_fp"`
		WeightStats   map[string]struct {
			Mean      float64 `json:"mean"`
			Std       float64 `json:"std"`
			Min       float64 `json:"min"`
			Max       float64 `json:"max"`
			NumUnique int     `json:"num_unique"`
		} `json:"weight_stats"`
	}
	if err := json.Unmarshal(outBytes, &result); err != nil {
		t.Fatalf("failed to parse script output: %v\nraw output: %s", err, string(outBytes))
	}

	// Log results
	t.Logf("Full-precision accuracy: %.2f%%", result.FPAccuracy*100)
	t.Logf("PTQ accuracy (30 levels): %.2f%%", result.PTQAccuracy*100)
	t.Logf("QAT accuracy (5-bit):     %.2f%%", result.QATAccuracy*100)
	t.Logf("Delta QAT vs PTQ: %+.2f%%", result.DeltaQATvsPTQ*100)
	t.Logf("Delta PTQ vs FP:  %+.2f%%", result.DeltaPTQvsFP*100)

	if stats, ok := result.WeightStats["ptq_layer1"]; ok {
		t.Logf("PTQ layer1 weight stats: %d unique levels, range [%.4f, %.4f]",
			stats.NumUnique, stats.Min, stats.Max)
	}

	// Sanity checks
	if result.FPAccuracy < 0.90 {
		t.Errorf("FP accuracy %.2f%% below 90%% — model training likely failed",
			result.FPAccuracy*100)
	}

	if result.PTQAccuracy < 0.85 {
		t.Errorf("PTQ accuracy %.2f%% below 85%% — quantization too aggressive",
			result.PTQAccuracy*100)
	}

	// PTQ degradation from FP should be modest (< 5%)
	ptqDrop := result.FPAccuracy - result.PTQAccuracy
	if ptqDrop > 0.05 {
		t.Errorf("PTQ accuracy drop %.2f%% from FP exceeds 5%% threshold",
			ptqDrop*100)
	}

	// QAT should recover some or all of the PTQ accuracy loss.
	// We don't require QAT > PTQ (training is stochastic), but log the comparison.
	if result.QATAccuracy < result.PTQAccuracy {
		t.Logf("NOTE: QAT accuracy (%.2f%%) < PTQ accuracy (%.2f%%) — "+
			"QAT training may need more epochs or tuning",
			result.QATAccuracy*100, result.PTQAccuracy*100)
	}

	// Verify PTQ produces the expected number of unique weight levels
	if stats, ok := result.WeightStats["ptq_layer1"]; ok {
		if stats.NumUnique > result.PTQLevels+1 { // +1 for rounding edge cases
			t.Errorf("PTQ produced %d unique levels, expected <= %d",
				stats.NumUnique, result.PTQLevels)
		}
	}

	// Emit artifact JSON for CI
	artifact := map[string]interface{}{
		"test":             "brevitas_qat_comparison",
		"fp_accuracy":      result.FPAccuracy,
		"ptq_accuracy":     result.PTQAccuracy,
		"qat_accuracy":     result.QATAccuracy,
		"ptq_levels":       result.PTQLevels,
		"qat_bit_width":    result.QATBitWidth,
		"delta_qat_vs_ptq": result.DeltaQATvsPTQ,
		"delta_ptq_vs_fp":  result.DeltaPTQvsFP,
		"ptq_drop_ok":      ptqDrop <= 0.05,
		"qat_ge_ptq":       result.QATAccuracy >= result.PTQAccuracy,
		"fp_above_90":      result.FPAccuracy >= 0.90,
		"max_ptq_drop_pct": math.Round(ptqDrop*10000) / 100,
	}
	artifactJSON, _ := json.MarshalIndent(artifact, "", "  ")
	artifactPath := filepath.Join(testsupport.ExternalArtifactDir(t), "brevitas_qat_comparison.json")
	os.WriteFile(artifactPath, artifactJSON, 0644)
	t.Logf("Artifact written to %s", artifactPath)
}
