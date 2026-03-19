// Package training provides neural network training for MNIST classification.
// single_layer.go implements the single-layer (784→10) network matching Dr. Tour's demo.
package training

import (
	"fmt"
	"math"
	"math/rand"

	"fecim-lattice-tools/shared/crossbar"
	"fecim-lattice-tools/shared/io"
	"fecim-lattice-tools/shared/mathutil"
)

// SingleLayerNetwork represents a single-layer neural network for MNIST.
// Architecture: 784 -> 10 (no hidden layer)
// Note: Peer-reviewed FeCIM achieves 96.6-98.24% MNIST accuracy; software baseline is 98-99%
type SingleLayerNetwork struct {
	layer  *crossbar.Array // 784 -> 10
	biases []float64       // Output biases
}

// NewSingleLayerNetwork creates a new single-layer MNIST network.
func NewSingleLayerNetwork() (*SingleLayerNetwork, error) {
	// Create crossbar array: 10 rows (outputs) x 784 cols (inputs)
	cfg := &crossbar.Config{
		Rows:       10,
		Cols:       784,
		NoiseLevel: 0.02,
		ADCBits:    6,
		DACBits:    8,
	}
	layer, err := crossbar.NewArray(cfg)
	if err != nil {
		return nil, fmt.Errorf("create single-layer crossbar: %w", err)
	}

	net := &SingleLayerNetwork{
		layer:  layer,
		biases: make([]float64, 10),
	}

	net.initializeWeights()
	return net, nil
}

func (n *SingleLayerNetwork) initializeWeights() {
	// Xavier initialization for single layer
	scale := math.Sqrt(2.0 / float64(784+10))
	for i := 0; i < 10; i++ {
		for j := 0; j < 784; j++ {
			// Map to 30 levels (0-1 range)
			w := rand.NormFloat64()*scale*0.5 + 0.5
			w = mathutil.Clamp01(w)
			n.layer.ProgramWeight(i, j, w)
		}
	}

	// Initialize biases to small random values
	for i := range n.biases {
		n.biases[i] = rand.Float64()*0.1 - 0.05
	}
}

// Forward performs forward pass through the network.
func (n *SingleLayerNetwork) Forward(input []float64) []float64 {
	weights := n.layer.GetConductanceMatrix() // [10][784]

	output := make([]float64, 10)
	for j := 0; j < 10; j++ {
		sum := n.biases[j]
		for i := 0; i < len(input); i++ {
			// effective = (conductance - 0.5) * 4
			effectiveWeight := (weights[j][i] - 0.5) * 4.0
			sum += input[i] * effectiveWeight
		}
		output[j] = sum
	}

	return softmax(output)
}

// Predict returns the predicted digit and confidence.
func (n *SingleLayerNetwork) Predict(input []float64) (int, float64) {
	probs := n.Forward(input)
	maxIdx := 0
	maxVal := probs[0]
	for i, v := range probs {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}
	return maxIdx, maxVal
}

// TrainEpoch runs one epoch of training using SGD.
func (n *SingleLayerNetwork) TrainEpoch(images [][]float64, labels []int, learningRate float64) float64 {
	totalLoss := 0.0

	// Shuffle indices
	indices := rand.Perm(len(images))

	for _, idx := range indices {
		input := images[idx]
		target := labels[idx]

		// Forward pass
		weights := n.layer.GetConductanceMatrix()
		output := make([]float64, 10)
		for j := 0; j < 10; j++ {
			sum := n.biases[j]
			for i := 0; i < len(input); i++ {
				effectiveWeight := (weights[j][i] - 0.5) * 4.0
				sum += input[i] * effectiveWeight
			}
			output[j] = sum
		}
		probs := softmax(output)

		// Compute cross-entropy loss
		loss := -math.Log(probs[target] + 1e-10)
		totalLoss += loss

		// Backward pass - compute gradients
		grad := make([]float64, 10)
		for i := range grad {
			grad[i] = probs[i]
			if i == target {
				grad[i] -= 1.0
			}
		}

		// Update weights
		for i := 0; i < 10; i++ {
			for j := 0; j < len(input); j++ {
				w := weights[i][j]
				dw := grad[i] * input[j] * learningRate * 0.25
				newW := w - dw
				n.layer.ProgramWeight(i, j, newW)
			}
			n.biases[i] -= grad[i] * learningRate
		}
	}

	return totalLoss / float64(len(images))
}

// Evaluate computes accuracy on a dataset.
func (n *SingleLayerNetwork) Evaluate(images [][]float64, labels []int) float64 {
	correct := 0
	for i, img := range images {
		pred, _ := n.Predict(img)
		if pred == labels[i] {
			correct++
		}
	}
	return float64(correct) / float64(len(images))
}

// Train trains the network for multiple epochs.
func (n *SingleLayerNetwork) Train(trainImages [][]float64, trainLabels []int,
	testImages [][]float64, testLabels []int,
	epochs int, learningRate float64) float64 {

	bestAcc := 0.0

	for epoch := 0; epoch < epochs; epoch++ {
		loss := n.TrainEpoch(trainImages, trainLabels, learningRate)
		trainAcc := n.Evaluate(trainImages[:1000], trainLabels[:1000]) // Sample for speed
		testAcc := n.Evaluate(testImages, testLabels)

		if testAcc > bestAcc {
			bestAcc = testAcc
		}

		fmt.Printf("Epoch %d: loss=%.4f, train_acc=%.2f%%, test_acc=%.2f%% (best=%.2f%%)\n",
			epoch+1, loss, trainAcc*100, testAcc*100, bestAcc*100)

		// Learning rate decay
		if epoch > 0 && epoch%5 == 0 {
			learningRate *= 0.9
		}
	}

	return bestAcc
}

// GetWeights returns the weight matrix.
func (n *SingleLayerNetwork) GetWeights() [][]float64 {
	return n.layer.GetConductanceMatrix()
}

// GetBiases returns the biases.
func (n *SingleLayerNetwork) GetBiases() []float64 {
	return n.biases
}

// singleLayerWeightsData is the JSON structure for single-layer weights.
type singleLayerWeightsData struct {
	SingleLayerWeights [][]float64 `json:"single_layer_weights"`
	SingleLayerBias    []float64   `json:"single_layer_bias"`
}

// SaveWeights saves the single-layer weights to a JSON file.
// Uses shared/io for consistent file handling across the codebase.
func (n *SingleLayerNetwork) SaveWeights(filename string) error {
	data := singleLayerWeightsData{
		SingleLayerWeights: n.layer.GetConductanceMatrix(),
		SingleLayerBias:    n.biases,
	}
	return io.SaveJSON(filename, data)
}

// LoadWeights loads single-layer weights from a JSON file.
// Uses shared/io for consistent file handling across the codebase.
func (n *SingleLayerNetwork) LoadWeights(filename string) error {
	var data singleLayerWeightsData
	if err := io.LoadJSON(filename, &data); err != nil {
		return err
	}

	// Program weights to crossbar array
	for i, row := range data.SingleLayerWeights {
		for j, w := range row {
			if i < 10 && j < 784 {
				n.layer.ProgramWeight(i, j, w)
			}
		}
	}

	// Copy biases
	if len(data.SingleLayerBias) >= 10 {
		copy(n.biases, data.SingleLayerBias)
	}

	return nil
}
