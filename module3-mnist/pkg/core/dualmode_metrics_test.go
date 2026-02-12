package core

import (
	"math"
	"testing"
)

func TestEvaluateDualModeDataset_BasicAccounting(t *testing.T) {
	net := createTestNetwork()
	net.Config.ADCBits = 16
	net.Config.DACBits = 16
	net.Config.NoiseLevel = 0

	images := [][]float64{
		{0.1, 0.2, 0.3, 0.4},
		{0.5, 0.4, 0.3, 0.2},
		{0.0, 0.1, 0.0, 0.1},
	}
	labels := []int{0, 1, 1}

	m := EvaluateDualModeDataset(net, images, labels)
	if m.Samples != 3 {
		t.Fatalf("Samples=%d, want 3", m.Samples)
	}

	fpTotal := 0
	cimTotal := 0
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			fpTotal += m.FP.Confusion[i][j]
			cimTotal += m.CIM.Confusion[i][j]
		}
	}
	if fpTotal != m.Samples || cimTotal != m.Samples {
		t.Fatalf("confusion totals mismatch: FP=%d CIM=%d samples=%d", fpTotal, cimTotal, m.Samples)
	}
	if m.FP.Accuracy < 0 || m.FP.Accuracy > 1 || m.CIM.Accuracy < 0 || m.CIM.Accuracy > 1 {
		t.Fatalf("accuracy out of range: fp=%.4f cim=%.4f", m.FP.Accuracy, m.CIM.Accuracy)
	}
	if m.Agreement < 0 || m.Agreement > 1 {
		t.Fatalf("agreement out of range: %.4f", m.Agreement)
	}
}

func TestInfer_CIMOrder_ADCBeforeNoise(t *testing.T) {
	net := NewDualModeNetwork(2, 2, 2)
	net.Config.SingleLayer = true
	net.Config.DACBits = 3
	net.Config.ADCBits = 3
	net.Config.NoiseLevel = 0.15

	net.SingleLayerWeights = [][]float64{{0.9, 0.1}, {0.2, 0.8}}
	net.SingleLayerBias = []float64{0.05, -0.02}
	net.QuantSingleLayerWeights = [][]float64{{0.9, 0.1}, {0.2, 0.8}}
	net.QuantSingleLayerBias = []float64{0.05, -0.02}

	input := []float64{0.23, 0.77}
	got := net.Infer(input)
	if got == nil {
		t.Fatal("Infer returned nil")
	}

	// Expected order in implementation: DAC -> MVM -> ADC -> noise -> softmax
	dacIn := quantizeDAC(input, net.Config.DACBits)
	base := net.forwardCIM(dacIn, net.QuantSingleLayerWeights, net.QuantSingleLayerBias)
	components := defaultNoiseComponents(net.Config.NoiseLevel)
	expected := applyDecomposedNoise(quantizeADC(base, net.Config.ADCBits), components, NewRandomSource(42))
	for i := range expected {
		if math.Abs(got.CIMLogits[i]-expected[i]) > 1e-12 {
			t.Fatalf("logit[%d]=%.15f, want %.15f (ADC->noise order)", i, got.CIMLogits[i], expected[i])
		}
	}
}
