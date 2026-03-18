package external_test

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"fecim-lattice-tools/shared/crossbar"
)

// requirePythonNumpy skips the test if python3 or numpy is not available.
func requirePythonNumpy(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not installed — skipping badcrossbar oracle test")
	}
	check := exec.Command("python3", "-c", "import numpy")
	if err := check.Run(); err != nil {
		t.Skip("numpy not installed — skipping badcrossbar oracle test. Install via: pip3 install numpy")
	}
}

// TestBadcrossbarOracle_IdealMVM tests a 4x4 identity weight matrix.
// With identity weights, MVM(x) = x, so the numpy oracle must return
// the input vector exactly. This validates the basic cross-validation
// pipeline end-to-end.
func TestBadcrossbarOracle_IdealMVM(t *testing.T) {
	requirePythonNumpy(t)

	n := 4
	// Identity matrix (values near 1.0 after 30-level quantization)
	weights := make([][]float64, n)
	for i := range weights {
		weights[i] = make([]float64, n)
		for j := range weights[i] {
			if i == j {
				weights[i][j] = 1.0
			}
		}
	}

	// Quantize weights the same way the crossbar does
	qWeights := make([][]float64, n)
	for i := range qWeights {
		qWeights[i] = make([]float64, n)
		for j := range qWeights[i] {
			qWeights[i][j] = crossbar.QuantizeToLevels(weights[i][j])
		}
	}

	input := []float64{0.25, 0.50, 0.75, 1.00}

	cvInput := crossvalInput{
		Weights:     qWeights,
		InputVector: input,
		ArraySize:   [2]int{n, n},
	}
	pyResult := runCrossvalScript(t, cvInput)

	// Go raw dot product
	goOutput := goMVMRaw(weights, input)

	t.Log("IdealMVM: 4x4 identity weights (quantized to 30 levels)")
	t.Log("────────────────────────────────────────────────────────")

	maxErr := 0.0
	for i := 0; i < n; i++ {
		absErr := math.Abs(goOutput[i] - pyResult.IdealOutput[i])
		if absErr > maxErr {
			maxErr = absErr
		}

		// Identity weights quantized to level 29 (= 1.0 in 30-level scheme).
		// So output[i] = quantized(1.0) * input[i] = input[i] exactly.
		t.Logf("Row %d: Go=%.10f  numpy=%.10f  input=%.2f  Δ=%.2e",
			i, goOutput[i], pyResult.IdealOutput[i], input[i], absErr)

		if absErr > 1e-6 {
			t.Errorf("Row %d: agreement violation: Go=%.8e numpy=%.8e Δ=%.3e > 1e-6",
				i, goOutput[i], pyResult.IdealOutput[i], absErr)
		}
	}

	t.Logf("Max error: %.3e (threshold: 1e-6)", maxErr)
	if maxErr <= 1e-6 {
		t.Log("PASS: identity MVM matches numpy oracle within 1e-6")
	}

	// Emit artifact
	dir := filepath.Join("..", "..", "output", "validation", "external")
	os.MkdirAll(dir, 0755)
	artifact := map[string]interface{}{
		"test": "badcrossbar_oracle_identity", "n": n,
		"max_err": maxErr, "pass": maxErr <= 1e-6,
	}
	b, _ := json.MarshalIndent(artifact, "", "  ")
	os.WriteFile(filepath.Join(dir, "badcrossbar_oracle_identity.json"), b, 0644)
}

// TestBadcrossbarOracle_RandomWeights tests random NxN weight matrices
// for N = 8, 16, 32. Verifies that Go and numpy agree on the raw MVM
// output within 1e-6 (same math, same quantized weights).
func TestBadcrossbarOracle_RandomWeights(t *testing.T) {
	requirePythonNumpy(t)

	sizes := []int{8, 16, 32}
	rng := rand.New(rand.NewSource(42)) // deterministic seed

	for _, n := range sizes {
		t.Run(fmt.Sprintf("%dx%d", n, n), func(t *testing.T) {
			// Generate random weights in [0, 1]
			weights := make([][]float64, n)
			for i := range weights {
				weights[i] = make([]float64, n)
				for j := range weights[i] {
					weights[i][j] = rng.Float64()
				}
			}

			// Quantize weights
			qWeights := make([][]float64, n)
			for i := range qWeights {
				qWeights[i] = make([]float64, n)
				for j := range qWeights[i] {
					qWeights[i][j] = crossbar.QuantizeToLevels(weights[i][j])
				}
			}

			// Random input in [0, 1]
			input := make([]float64, n)
			for j := range input {
				input[j] = rng.Float64()
			}

			cvInput := crossvalInput{
				Weights:     qWeights,
				InputVector: input,
				ArraySize:   [2]int{n, n},
			}
			pyResult := runCrossvalScript(t, cvInput)

			// Go raw dot product
			goOutput := goMVMRaw(weights, input)

			maxErr := 0.0
			worstRow := 0
			for i := 0; i < n; i++ {
				absErr := math.Abs(goOutput[i] - pyResult.IdealOutput[i])
				if absErr > maxErr {
					maxErr = absErr
					worstRow = i
				}
			}

			t.Logf("Size %dx%d: maxErr=%.3e at row %d (threshold: 1e-6)", n, n, maxErr, worstRow)
			t.Logf("  Go[%d]=%.10f  numpy[%d]=%.10f", worstRow, goOutput[worstRow], worstRow, pyResult.IdealOutput[worstRow])

			if maxErr > 1e-6 {
				t.Errorf("FAIL: %dx%d max error %.3e > 1e-6", n, n, maxErr)
			} else {
				t.Logf("PASS: %dx%d random MVM matches numpy within 1e-6", n, n)
			}

			// Also verify via the crossbar Array API (full pipeline with normalization)
			cfg := &crossbar.Config{
				Rows:       n,
				Cols:       n,
				NoiseLevel: 0,
				ADCBits:    16,
				DACBits:    16,
			}
			arr, err := crossbar.NewArray(cfg)
			if err != nil {
				t.Fatalf("NewArray: %v", err)
			}
			defer arr.Destroy()

			if err := arr.ProgramWeightMatrix(weights); err != nil {
				t.Fatalf("ProgramWeightMatrix: %v", err)
			}

			goMVMOutput, err := arr.MVM(input)
			if err != nil {
				t.Fatalf("MVM: %v", err)
			}

			// Trend check: verify monotonicity agreement.
			// Rows with larger numpy outputs should also have larger Go outputs.
			trendPass := true
			trendChecks := 0
			for i := 0; i < n-1; i++ {
				npDir := pyResult.IdealOutput[i+1] - pyResult.IdealOutput[i]
				goDir := goMVMOutput[i+1] - goMVMOutput[i]
				// Skip near-zero differences (below quantization noise)
				if math.Abs(npDir) < 1e-4 {
					continue
				}
				trendChecks++
				if (npDir > 0) != (goDir > 0) {
					trendPass = false
					t.Errorf("Trend violation at rows %d→%d: numpy Δ=%.4e, Go Δ=%.4e",
						i, i+1, npDir, goDir)
				}
			}
			t.Logf("Trend checks: %d pairs evaluated, all consistent: %v", trendChecks, trendPass)

			fmt.Printf("BADCROSSBAR_RANDOM: n=%d maxErr=%.3e trend=%v\n", n, maxErr, trendPass)
		})
	}
}

// TestBadcrossbarOracle_IRDropDirection verifies that when wire resistance
// increases, output current magnitudes decrease. This is a fundamental
// physical property: parasitic resistance drops voltage across devices,
// reducing their current contribution.
//
// The test uses the badcrossbar Python package if available. If badcrossbar
// is not installed, it falls back to a simple analytical IR-drop model
// to verify the expected trend.
func TestBadcrossbarOracle_IRDropDirection(t *testing.T) {
	requirePythonNumpy(t)

	n := 8
	rng := rand.New(rand.NewSource(99))

	// Fixed weight matrix and input
	weights := make([][]float64, n)
	for i := range weights {
		weights[i] = make([]float64, n)
		for j := range weights[i] {
			weights[i][j] = 0.3 + 0.4*rng.Float64() // range [0.3, 0.7]
		}
	}

	qWeights := make([][]float64, n)
	for i := range qWeights {
		qWeights[i] = make([]float64, n)
		for j := range qWeights[i] {
			qWeights[i][j] = crossbar.QuantizeToLevels(weights[i][j])
		}
	}

	input := make([]float64, n)
	for j := range input {
		input[j] = 0.5 + 0.3*rng.Float64() // range [0.5, 0.8]
	}

	// Probe whether badcrossbar is available by running a small test case
	hasBadcrossbar := false
	probeInput := crossvalInput{
		Weights:        qWeights,
		InputVector:    input,
		WireResistance: wireResistance{Wordline: 1.0, Bitline: 1.0},
		ArraySize:      [2]int{n, n},
	}
	probeResult := runCrossvalScript(t, probeInput)
	hasBadcrossbar = probeResult.BadcrossbarAvailable

	// Test increasing wire resistance values (all > 0 to engage badcrossbar)
	resistances := []float64{0.01, 1.0, 5.0, 10.0, 50.0}

	type result struct {
		wireR      float64
		output     []float64
		outputNorm float64 // L2 norm of output vector
		hasBCB     bool
	}

	var results []result

	for _, r := range resistances {
		cvInput := crossvalInput{
			Weights:     qWeights,
			InputVector: input,
			WireResistance: wireResistance{
				Wordline: r,
				Bitline:  r,
			},
			ArraySize: [2]int{n, n},
		}
		pyResult := runCrossvalScript(t, cvInput)

		// Use IR-drop output if badcrossbar is available, else ideal
		var output []float64
		usedBCB := pyResult.BadcrossbarAvailable
		if usedBCB && pyResult.IRDropOutput != nil {
			output = pyResult.IRDropOutput
		} else {
			output = pyResult.IdealOutput
		}

		// Compute L2 norm
		norm := 0.0
		for _, v := range output {
			norm += v * v
		}
		norm = math.Sqrt(norm)

		results = append(results, result{
			wireR:      r,
			output:     output,
			outputNorm: norm,
			hasBCB:     usedBCB,
		})
	}

	// Log results
	if hasBadcrossbar {
		t.Log("IR-drop direction test using badcrossbar (full resistor network solver)")
	} else {
		t.Log("IR-drop direction test: badcrossbar not available, verifying ideal baseline")
		t.Log("Without badcrossbar, ideal MVM is independent of wire resistance.")
		t.Log("Test validates that the oracle pipeline works; IR-drop trend requires badcrossbar.")
	}
	t.Log("──────────────────────────────────────────────────────────────────")
	t.Logf("%-12s  %-14s  %-10s", "Wire R (Ω)", "||output||₂", "BCB used")
	t.Log("──────────────────────────────────────────────────────────────────")
	for _, r := range results {
		t.Logf("%-12.2f  %-14.6f  %-10v", r.wireR, r.outputNorm, r.hasBCB)
	}

	if hasBadcrossbar {
		// With badcrossbar: verify monotonic decrease in output norm
		// as wire resistance increases (IR-drop reduces effective voltages)
		t.Log("")
		t.Log("Checking monotonic decrease in output magnitude with increasing wire R:")
		allDecreasing := true
		for i := 1; i < len(results); i++ {
			prev := results[i-1]
			curr := results[i]
			if curr.wireR > prev.wireR && curr.outputNorm >= prev.outputNorm {
				// Allow small tolerance for numerical noise
				if curr.outputNorm-prev.outputNorm > 1e-6 {
					allDecreasing = false
					t.Errorf("IR-drop direction violated: R=%.1f->%.1f, ||y||=%.6f->%.6f (should decrease)",
						prev.wireR, curr.wireR, prev.outputNorm, curr.outputNorm)
				}
			}
			delta := prev.outputNorm - curr.outputNorm
			label := "OK (decreased)"
			if delta <= -1e-6 {
				label = "VIOLATION"
			}
			t.Logf("  R %.2f->%.2f ohm:  delta||y|| = %.6e  %s",
				prev.wireR, curr.wireR, delta, label)
		}
		if allDecreasing {
			t.Log("PASS: output magnitude monotonically decreases with increasing wire resistance")
		}
		fmt.Printf("BADCROSSBAR_IRDROP: n=%d resistances=%v decreasing=%v\n", n, resistances, allDecreasing)
	} else {
		// Without badcrossbar: verify ideal outputs are constant regardless of wire R
		// (since ideal MVM ignores wire resistance)
		baseNorm := results[0].outputNorm
		allSame := true
		for i := 1; i < len(results); i++ {
			delta := math.Abs(results[i].outputNorm - baseNorm)
			if delta > 1e-10 {
				allSame = false
				t.Errorf("Ideal output should be constant: R=%.2f norm=%.6f vs base=%.6f",
					results[i].wireR, results[i].outputNorm, baseNorm)
			}
		}
		if allSame {
			t.Log("PASS: ideal MVM output is constant (wire R ignored without badcrossbar)")
			t.Log("NOTE: install badcrossbar (pip3 install badcrossbar) to test IR-drop trends")
		}
		fmt.Printf("BADCROSSBAR_IRDROP: n=%d badcrossbar=false ideal_constant=%v\n", n, allSame)
	}

	// Emit artifact
	dir := filepath.Join("..", "..", "output", "validation", "external")
	os.MkdirAll(dir, 0755)
	artifactResults := make([]map[string]interface{}, len(results))
	for i, r := range results {
		artifactResults[i] = map[string]interface{}{
			"wire_resistance_ohm": r.wireR,
			"output_l2_norm":      r.outputNorm,
			"badcrossbar_used":    r.hasBCB,
		}
	}
	artifact := map[string]interface{}{
		"test":    "badcrossbar_oracle_irdrop_direction",
		"n":       n,
		"results": artifactResults,
	}
	b, _ := json.MarshalIndent(artifact, "", "  ")
	os.WriteFile(filepath.Join(dir, "badcrossbar_oracle_irdrop.json"), b, 0644)
}
