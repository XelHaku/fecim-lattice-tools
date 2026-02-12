package training

import (
	"math"
	"testing"

	"fecim-lattice-tools/module2-crossbar/pkg/crossbar"
)

func newDeterministicTinyNet(t *testing.T) *MNISTNetwork {
	t.Helper()

	layer1, err := crossbar.NewArray(&crossbar.Config{Rows: 2, Cols: 2, NoiseLevel: 0, ADCBits: 8, DACBits: 8})
	if err != nil {
		t.Fatalf("create layer1: %v", err)
	}
	layer2, err := crossbar.NewArray(&crossbar.Config{Rows: 10, Cols: 2, NoiseLevel: 0, ADCBits: 8, DACBits: 8})
	if err != nil {
		t.Fatalf("create layer2: %v", err)
	}

	net := NewMNISTNetworkWithWeights(layer1, layer2)

	// Layer 1 (effective weight = (g-0.5)*4):
	// h0 = 1*x0 + 0*x1 + 0.1, h1 = 0*x0 + 1*x1 - 0.2
	layer1.ProgramWeight(0, 0, 0.75)
	layer1.ProgramWeight(0, 1, 0.50)
	layer1.ProgramWeight(1, 0, 0.50)
	layer1.ProgramWeight(1, 1, 0.75)
	net.biases1[0] = 0.1
	net.biases1[1] = -0.2

	// Layer 2: class-0 reads h0, class-1 reads h1, others zero weights.
	for o := 0; o < 10; o++ {
		layer2.ProgramWeight(o, 0, 0.5)
		layer2.ProgramWeight(o, 1, 0.5)
	}
	layer2.ProgramWeight(0, 0, 0.75)
	layer2.ProgramWeight(1, 1, 0.75)
	for i := range net.biases2 {
		net.biases2[i] = 0
	}

	return net
}

func makeZeroGrad(net *MNISTNetwork) Gradients {
	g := Gradients{
		dW1: make([][]float64, net.hiddenSize),
		dB1: make([]float64, net.hiddenSize),
		dW2: make([][]float64, 10),
		dB2: make([]float64, 10),
	}
	for h := 0; h < net.hiddenSize; h++ {
		g.dW1[h] = make([]float64, net.layer1.Cols())
	}
	for o := 0; o < 10; o++ {
		g.dW2[o] = make([]float64, net.hiddenSize)
	}
	return g
}

func TestTraining_AnalyticalGradientSingleStep(t *testing.T) {
	net := newDeterministicTinyNet(t)
	input := []float64{1.0, 0.5}
	target := 0

	cache := net.forwardWithCache(input)
	_, dLogits := CrossEntropyLoss{}.Forward(cache.Logits, target)
	grads := net.backward(cache, dLogits)

	probs := softmax(cache.Logits)
	hidden := cache.HiddenAct

	// dW2[o][h] = (p_o - y_o) * h_h
	for o := 0; o < 10; o++ {
		y := 0.0
		if o == target {
			y = 1.0
		}
		for h := 0; h < 2; h++ {
			expected := (probs[o] - y) * hidden[h]
			if math.Abs(grads.dW2[o][h]-expected) > 1e-12 {
				t.Fatalf("dW2[%d][%d]=%.12f, expected %.12f", o, h, grads.dW2[o][h], expected)
			}
		}
	}

	// dHidden = W2^T * dLogits (with effective W2 and ReLU derivative)
	effW2 := net.layer2.GetConductanceMatrix()
	dHidden := make([]float64, 2)
	for h := 0; h < 2; h++ {
		sum := 0.0
		for o := 0; o < 10; o++ {
			sum += dLogits[o] * ((effW2[o][h] - 0.5) * 4.0)
		}
		if cache.HiddenPre[h] <= 0 {
			sum = 0
		}
		dHidden[h] = sum
	}

	for h := 0; h < 2; h++ {
		if math.Abs(grads.dB1[h]-dHidden[h]) > 1e-12 {
			t.Fatalf("dB1[%d]=%.12f, expected %.12f", h, grads.dB1[h], dHidden[h])
		}
		for i := 0; i < 2; i++ {
			expected := dHidden[h] * input[i]
			if math.Abs(grads.dW1[h][i]-expected) > 1e-12 {
				t.Fatalf("dW1[%d][%d]=%.12f, expected %.12f", h, i, grads.dW1[h][i], expected)
			}
		}
	}
}

func TestTraining_LossMonotonicDecrease10Steps(t *testing.T) {
	net := newDeterministicTinyNet(t)
	input := []float64{1.0, 1.0}
	target := 0
	cfg := TrainingConfig{LearningRate: 0.2, Loss: CrossEntropyLoss{}, Optimizer: NewSGDOptimizer(0.2)}

	prevLoss := math.Inf(1)
	for step := 0; step < 10; step++ {
		cache := net.forwardWithCache(input)
		loss, dLogits := cfg.Loss.Forward(cache.Logits, target)
		if loss > prevLoss+1e-12 {
			t.Fatalf("loss not monotonic at step %d: prev=%.12f curr=%.12f", step, prevLoss, loss)
		}
		prevLoss = loss

		grads := net.backward(cache, dLogits)
		net.applyGradients(grads, cfg)
	}
}

func TestTraining_WeightUpdateScalesWithLearningRate(t *testing.T) {
	netSmallLR := newDeterministicTinyNet(t)
	netLargeLR := newDeterministicTinyNet(t)

	g1 := makeZeroGrad(netSmallLR)
	g2 := makeZeroGrad(netLargeLR)
	g1.dW2[0][0], g2.dW2[0][0] = 4.0, 4.0 // conductance-space delta = lr*(4*0.25)=lr

	w0 := netSmallLR.layer2.GetConductanceMatrix()[0][0]

	cfgSmall := TrainingConfig{LearningRate: 0.1, Loss: CrossEntropyLoss{}, Optimizer: NewSGDOptimizer(0.1)}
	cfgLarge := TrainingConfig{LearningRate: 0.2, Loss: CrossEntropyLoss{}, Optimizer: NewSGDOptimizer(0.2)}

	netSmallLR.applyGradients(g1, cfgSmall)
	netLargeLR.applyGradients(g2, cfgLarge)

	deltaSmall := math.Abs(w0 - netSmallLR.layer2.GetConductanceMatrix()[0][0])
	deltaLarge := math.Abs(w0 - netLargeLR.layer2.GetConductanceMatrix()[0][0])
	if deltaSmall == 0 {
		t.Fatalf("small-lr update vanished due to quantization")
	}
	ratio := deltaLarge / deltaSmall
	if math.Abs(ratio-2.0) > 0.35 {
		t.Fatalf("learning-rate scaling mismatch: deltaSmall=%.6f deltaLarge=%.6f ratio=%.3f (expected ~2)", deltaSmall, deltaLarge, ratio)
	}
}

type clippedSGD struct {
	lr   float64
	clip float64
}

func (o *clippedSGD) Step(_ string, value, grad float64) float64 {
	if grad > o.clip {
		grad = o.clip
	}
	if grad < -o.clip {
		grad = -o.clip
	}
	return value - o.lr*grad
}
func (o *clippedSGD) Reset() {}

func TestTraining_GradientClippingPreventsExplodingUpdates(t *testing.T) {
	netNoClip := newDeterministicTinyNet(t)
	netClip := newDeterministicTinyNet(t)

	gNoClip := makeZeroGrad(netNoClip)
	gClip := makeZeroGrad(netClip)

	// Extremely large gradients.
	gNoClip.dW2[0][0], gClip.dW2[0][0] = 400.0, 400.0
	gNoClip.dB2[0], gClip.dB2[0] = 400.0, 400.0

	w0 := netNoClip.layer2.GetConductanceMatrix()[0][0]

	netNoClip.applyGradients(gNoClip, TrainingConfig{LearningRate: 0.1, Optimizer: NewSGDOptimizer(0.1)})
	netClip.applyGradients(gClip, TrainingConfig{LearningRate: 0.1, Optimizer: &clippedSGD{lr: 0.1, clip: 1.0}})

	deltaNoClip := math.Abs(w0 - netNoClip.layer2.GetConductanceMatrix()[0][0])
	deltaClip := math.Abs(w0 - netClip.layer2.GetConductanceMatrix()[0][0])

	if deltaClip >= deltaNoClip {
		t.Fatalf("clipping ineffective: clipped delta %.6f >= unclipped %.6f", deltaClip, deltaNoClip)
	}
	if deltaClip > 0.2 {
		t.Fatalf("clipped update too large: %.6f", deltaClip)
	}
}
