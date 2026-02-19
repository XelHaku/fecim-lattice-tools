package neural

import (
	"math"
	"sort"
	"testing"
)

// TestCIMEndToEndMNISTDigit0Pipeline validates the full CIM inference path for
// a 2-layer 784→30→10 network using a synthetic "digit 0" pattern.
func TestCIMEndToEndMNISTDigit0Pipeline(t *testing.T) {
	const (
		inputSize  = 28 * 28
		hiddenSize = 30
		outputSize = 10
	)

	net := NewDualModeNetwork(inputSize, hiddenSize, outputSize)
	net.SetNoiseLevel(0.0)
	net.SetNumLevels(30)
	net.SetADCBits(8)
	net.SetDACBits(8)

	// (1) Construct known weights for a center-bright / edge-dark detector.
	for h := 0; h < hiddenSize; h++ {
		centerW := 0.045 + 0.001*float64(h%3)
		edgeW := -0.025 - 0.001*float64(h%2)
		for idx := 0; idx < inputSize; idx++ {
			row, col := idx/28, idx%28
			if isCenterPixel(row, col) {
				net.FPWeights1[h][idx] = centerW
			} else {
				net.FPWeights1[h][idx] = edgeW
			}
		}
		net.FPBias1[h] = 0.02
	}

	for out := 0; out < outputSize; out++ {
		for h := 0; h < hiddenSize; h++ {
			if out == 0 {
				net.FPWeights2[out][h] = 0.20 + 0.002*float64(h%4)
			} else {
				net.FPWeights2[out][h] = -0.06
			}
		}
		if out == 0 {
			net.FPBias2[out] = 0.8
		} else {
			net.FPBias2[out] = -0.2 - 0.02*float64(out)
		}
	}

	// (2) Program weights into crossbar representation via quantization.
	net.RequantizeWeights()
	if len(net.QuantWeights1) != hiddenSize || len(net.QuantWeights2) != outputSize {
		t.Fatalf("quantized crossbar dimensions invalid: got (%d,%d)", len(net.QuantWeights1), len(net.QuantWeights2))
	}

	// (3) Create synthetic digit-0 pattern: center high, edges low.
	input := make([]float64, inputSize)
	for idx := 0; idx < inputSize; idx++ {
		row, col := idx/28, idx%28
		if isCenterPixel(row, col) {
			input[idx] = 0.95
		} else {
			input[idx] = 0.05
		}
	}

	result := net.Infer(input)
	if result == nil {
		t.Fatal("Infer returned nil")
	}
	if len(result.CIMProbabilities) != outputSize {
		t.Fatalf("expected %d CIM probabilities, got %d", outputSize, len(result.CIMProbabilities))
	}

	// (4) Verify digit 0 is argmax or at least in top-3.
	if result.CIMPrediction != 0 && !isInTopK(result.CIMProbabilities, 0, 3) {
		t.Fatalf("digit 0 not argmax/top-3: pred=%d probs=%v", result.CIMPrediction, result.CIMProbabilities)
	}

	softmaxSum := 0.0
	for _, p := range result.CIMProbabilities {
		softmaxSum += p
	}
	if math.Abs(softmaxSum-1.0) > 1e-6 {
		t.Fatalf("CIM softmax probabilities do not sum to 1: sum=%.9f", softmaxSum)
	}

	// (5) Verify positive and reasonable energy estimate.
	if result.EnergyUsed <= 0 {
		t.Fatalf("energy must be positive, got %.9f uJ", result.EnergyUsed)
	}
	expectedEnergyUj := EstimateInferenceEnergyMicroJ(net.Config, inputSize, hiddenSize, outputSize)
	if result.EnergyUsed < expectedEnergyUj*0.95 || result.EnergyUsed > expectedEnergyUj*1.05 {
		t.Fatalf("unexpected energy estimate: got %.9f uJ, expected %.9f uJ (±5%%)", result.EnergyUsed, expectedEnergyUj)
	}
	if result.EnergyUsed > 0.05 {
		t.Fatalf("energy estimate appears unreasonable for 784→30→10 FeCIM inference: %.9f uJ", result.EnergyUsed)
	}
}

func isCenterPixel(row, col int) bool {
	return row >= 8 && row <= 19 && col >= 8 && col <= 19
}

func isInTopK(probs []float64, class, k int) bool {
	if class < 0 || class >= len(probs) || k <= 0 {
		return false
	}
	idx := make([]int, len(probs))
	for i := range probs {
		idx[i] = i
	}
	sort.Slice(idx, func(i, j int) bool {
		return probs[idx[i]] > probs[idx[j]]
	})
	if k > len(idx) {
		k = len(idx)
	}
	for i := 0; i < k; i++ {
		if idx[i] == class {
			return true
		}
	}
	return false
}
