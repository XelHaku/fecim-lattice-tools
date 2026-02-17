package network

import (
	"math"
	"testing"

	"fecim-lattice-tools/shared/crossbar"
)

func TestNewNetwork_ValidConfig(t *testing.T) {
	cfg := &Config{
		InputSize:  784,
		HiddenSize: 128,
		OutputSize: 10,
		NumLayers:  3,
	}

	// Create a base array (not actually used in NewNetwork, but required by API)
	arrayCfg := &crossbar.Config{
		Rows: 128, Cols: 784, NoiseLevel: 0.01, ADCBits: 8, DACBits: 8,
	}
	baseArray, _ := crossbar.NewArray(arrayCfg)

	net, err := NewNetwork(cfg, baseArray)
	if err != nil {
		t.Fatalf("NewNetwork failed: %v", err)
	}

	if net.GetLayerCount() != 2 {
		t.Errorf("Expected 2 weight layers, got %d", net.GetLayerCount())
	}
}

func TestNewNetwork_InvalidLayerCount(t *testing.T) {
	cfg := &Config{
		InputSize:  10,
		HiddenSize: 5,
		OutputSize: 2,
		NumLayers:  1, // Invalid: need at least 2
	}

	_, err := NewNetwork(cfg, nil)
	if err == nil {
		t.Error("Expected error for network with < 2 layers")
	}
}

func TestNetwork_Forward(t *testing.T) {
	cfg := &Config{
		InputSize:  4,
		HiddenSize: 3,
		OutputSize: 2,
		NumLayers:  3,
	}

	arrayCfg := &crossbar.Config{
		Rows: 3, Cols: 4, NoiseLevel: 0.0, ADCBits: 8, DACBits: 8,
	}
	baseArray, _ := crossbar.NewArray(arrayCfg)

	net, err := NewNetwork(cfg, baseArray)
	if err != nil {
		t.Fatalf("NewNetwork failed: %v", err)
	}

	input := []float64{0.5, 0.5, 0.5, 0.5}
	output, err := net.Forward(input)
	if err != nil {
		t.Fatalf("Forward failed: %v", err)
	}

	// Check output is valid probability distribution
	if len(output) != 2 {
		t.Errorf("Expected 2 outputs, got %d", len(output))
	}

	sum := 0.0
	for _, p := range output {
		if p < 0 || p > 1 {
			t.Errorf("Output probability out of range: %f", p)
		}
		sum += p
	}

	if math.Abs(sum-1.0) > 0.001 {
		t.Errorf("Output probabilities don't sum to 1: %f", sum)
	}
}

func TestNetwork_OpsCount(t *testing.T) {
	cfg := &Config{
		InputSize:  4,
		HiddenSize: 3,
		OutputSize: 2,
		NumLayers:  3,
	}

	arrayCfg := &crossbar.Config{
		Rows: 3, Cols: 4, NoiseLevel: 0.0, ADCBits: 8, DACBits: 8,
	}
	baseArray, _ := crossbar.NewArray(arrayCfg)

	net, _ := NewNetwork(cfg, baseArray)

	// Before inference
	if net.GetOpsCount() != 0 {
		t.Error("Ops count should be 0 before inference")
	}

	// After inference
	input := []float64{0.5, 0.5, 0.5, 0.5}
	net.Forward(input)

	ops := net.GetOpsCount()
	// Layer 1: 3×4 = 12 ops, Layer 2: 2×3 = 6 ops
	expected := int64(12 + 6)
	if ops != expected {
		t.Errorf("Expected %d ops, got %d", expected, ops)
	}
}

func TestNetwork_GetLayerDimensions(t *testing.T) {
	cfg := &Config{
		InputSize:  784,
		HiddenSize: 128,
		OutputSize: 10,
		NumLayers:  3,
	}

	arrayCfg := &crossbar.Config{
		Rows: 128, Cols: 784, NoiseLevel: 0.0, ADCBits: 8, DACBits: 8,
	}
	baseArray, _ := crossbar.NewArray(arrayCfg)

	net, _ := NewNetwork(cfg, baseArray)

	// Check first layer (input -> hidden)
	in, out, err := net.GetLayerDimensions(0)
	if err != nil {
		t.Fatalf("GetLayerDimensions failed: %v", err)
	}
	if in != 784 || out != 128 {
		t.Errorf("Layer 0: expected 784x128, got %dx%d", in, out)
	}

	// Check second layer (hidden -> output)
	in, out, err = net.GetLayerDimensions(1)
	if err != nil {
		t.Fatalf("GetLayerDimensions failed: %v", err)
	}
	if in != 128 || out != 10 {
		t.Errorf("Layer 1: expected 128x10, got %dx%d", in, out)
	}

	// Check out of range
	_, _, err = net.GetLayerDimensions(5)
	if err == nil {
		t.Error("Expected error for out-of-range layer index")
	}
}

func TestNetwork_LoadWeights(t *testing.T) {
	cfg := &Config{
		InputSize:  4,
		HiddenSize: 3,
		OutputSize: 2,
		NumLayers:  3,
	}

	arrayCfg := &crossbar.Config{
		Rows: 3, Cols: 4, NoiseLevel: 0.0, ADCBits: 8, DACBits: 8,
	}
	baseArray, _ := crossbar.NewArray(arrayCfg)

	net, _ := NewNetwork(cfg, baseArray)

	// Create test weights
	weights := [][][]float64{
		{ // Layer 1: 3x4
			{0.1, 0.2, 0.3, 0.4},
			{0.5, 0.6, 0.7, 0.8},
			{0.9, 0.8, 0.7, 0.6},
		},
		{ // Layer 2: 2x3
			{0.1, 0.2, 0.3},
			{0.4, 0.5, 0.6},
		},
	}
	biases := [][]float64{
		{0.01, 0.02, 0.03},
		{0.04, 0.05},
	}

	err := net.LoadWeights(weights, biases)
	if err != nil {
		t.Fatalf("LoadWeights failed: %v", err)
	}
}

func TestNetwork_LoadWeights_SizeMismatch(t *testing.T) {
	cfg := &Config{
		InputSize:  4,
		HiddenSize: 3,
		OutputSize: 2,
		NumLayers:  3,
	}

	arrayCfg := &crossbar.Config{
		Rows: 3, Cols: 4, NoiseLevel: 0.0, ADCBits: 8, DACBits: 8,
	}
	baseArray, _ := crossbar.NewArray(arrayCfg)

	net, _ := NewNetwork(cfg, baseArray)

	// Wrong number of layers
	weights := [][][]float64{
		{{0.1, 0.2}}, // Only one layer
	}

	err := net.LoadWeights(weights, nil)
	if err == nil {
		t.Error("Expected error for weight count mismatch")
	}
}

func TestRelu(t *testing.T) {
	input := []float64{-1.0, 0.0, 1.0, -0.5, 0.5}
	expected := []float64{0.0, 0.0, 1.0, 0.0, 0.5}

	output := relu(input)

	for i := range output {
		if output[i] != expected[i] {
			t.Errorf("ReLU(%f) = %f, expected %f", input[i], output[i], expected[i])
		}
	}
}

func TestSoftmax(t *testing.T) {
	input := []float64{1.0, 2.0, 3.0}
	output := softmax(input)

	// Check sum to 1
	sum := 0.0
	for _, p := range output {
		if p < 0 || p > 1 {
			t.Errorf("Softmax output out of range: %f", p)
		}
		sum += p
	}

	if math.Abs(sum-1.0) > 0.001 {
		t.Errorf("Softmax doesn't sum to 1: %f", sum)
	}

	// Check ordering (higher input -> higher probability)
	if output[2] <= output[1] || output[1] <= output[0] {
		t.Error("Softmax should preserve ordering")
	}
}

func TestSoftmax_Stability(t *testing.T) {
	// Test numerical stability with large values
	input := []float64{1000.0, 1001.0, 1002.0}
	output := softmax(input)

	sum := 0.0
	for _, p := range output {
		if math.IsNaN(p) || math.IsInf(p, 0) {
			t.Error("Softmax produced NaN or Inf")
		}
		sum += p
	}

	if math.Abs(sum-1.0) > 0.001 {
		t.Errorf("Softmax doesn't sum to 1 for large inputs: %f", sum)
	}
}

func BenchmarkNetwork_Forward(b *testing.B) {
	cfg := &Config{
		InputSize:  784,
		HiddenSize: 128,
		OutputSize: 10,
		NumLayers:  3,
	}

	arrayCfg := &crossbar.Config{
		Rows: 128, Cols: 784, NoiseLevel: 0.0, ADCBits: 8, DACBits: 8,
	}
	baseArray, _ := crossbar.NewArray(arrayCfg)

	net, _ := NewNetwork(cfg, baseArray)

	input := make([]float64, 784)
	for i := range input {
		input[i] = 0.5
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		net.Forward(input)
	}
}
