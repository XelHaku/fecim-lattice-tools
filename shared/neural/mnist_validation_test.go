package neural

import (
	"math"
	"testing"
)

func TestFPPathMath_LinearReLUSoftmaxNormalization(t *testing.T) {
	net := createTestNetwork()
	input := []float64{0.1, 0.2, 0.3, 0.4}
	got := net.Infer(input)
	if got == nil {
		t.Fatal("Infer returned nil")
	}

	wantHidden := []float64{0.4, 0.9, 1.4}
	wantLogits := []float64{0.74, 1.65}

	for i := range wantHidden {
		if math.Abs(got.FPHidden[i]-wantHidden[i]) > 1e-12 {
			t.Fatalf("FPHidden[%d]=%.12f, want %.12f", i, got.FPHidden[i], wantHidden[i])
		}
	}
	for i := range wantLogits {
		if math.Abs(got.FPLogits[i]-wantLogits[i]) > 1e-12 {
			t.Fatalf("FPLogits[%d]=%.12f, want %.12f", i, got.FPLogits[i], wantLogits[i])
		}
	}

	sum := 0.0
	for _, p := range got.FPProbabilities {
		if p < 0 || p > 1 {
			t.Fatalf("probability out of range: %.12f", p)
		}
		sum += p
	}
	if math.Abs(sum-1) > 1e-12 {
		t.Fatalf("softmax sum=%.15f, want 1.0", sum)
	}
}

func TestCIMPath_Order_DACThenMVMThenADC_NoNoise(t *testing.T) {
	net := NewDualModeNetwork(2, 2, 2)
	net.Config.SingleLayer = true
	net.Config.DACBits = 3
	net.Config.ADCBits = 3
	net.Config.NoiseLevel = 0

	net.QuantSingleLayerWeights = [][]float64{{0.9, 0.1}, {0.2, 0.8}}
	net.QuantSingleLayerBias = []float64{0.05, -0.02}
	net.SingleLayerWeights = net.QuantSingleLayerWeights
	net.SingleLayerBias = net.QuantSingleLayerBias

	input := []float64{0.23, 0.77}
	got := net.Infer(input)
	if got == nil {
		t.Fatal("Infer returned nil")
	}

	dac := quantizeDAC(input, net.Config.DACBits)
	mvm := net.forwardCIM(dac, net.QuantSingleLayerWeights, net.QuantSingleLayerBias)
	want := quantizeADC(mvm, net.Config.ADCBits)

	for i := range want {
		if math.Abs(got.CIMLogits[i]-want[i]) > 1e-12 {
			t.Fatalf("CIMLogits[%d]=%.15f, want %.15f", i, got.CIMLogits[i], want[i])
		}
	}
}

func TestDisagreementMetric_KLDivergenceKnownValue(t *testing.T) {
	p := []float64{0.5, 0.5}
	q := []float64{0.9, 0.1}
	got := klDivergence(p, q)
	eps := 1e-10
	want := 0.5*math.Log(0.5/(0.9+eps)) + 0.5*math.Log(0.5/(0.1+eps))
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("KL=%.15f, want %.15f", got, want)
	}
}

func TestEnergyModel_PerLayerQuantUsesLayerLevels(t *testing.T) {
	cfg := DefaultNetworkConfig()
	cfg.PerLayerQuant = true
	cfg.Layer1Levels = 8
	cfg.Layer2Levels = 16
	est := EstimateInferenceEnergyJ(cfg, 784, 128, 10)

	if est.Layer1Levels != 8 || est.Layer2Levels != 16 {
		t.Fatalf("levels mismatch: got L1=%d L2=%d", est.Layer1Levels, est.Layer2Levels)
	}

	wantCompute := float64(est.MACs1)*EnergyPerMACJ(8) + float64(est.MACs2)*EnergyPerMACJ(16)
	if math.Abs(est.ComputeJ-wantCompute) > wantCompute*1e-12 {
		t.Fatalf("ComputeJ=%.18g, want %.18g", est.ComputeJ, wantCompute)
	}
}
