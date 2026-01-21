// Package training provides neural network training for MNIST classification.
package training

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"

	"multilayer-ferroelectric-cim-visualizer/demo2-crossbar/pkg/crossbar"
)

// MNISTNetwork represents a 2-layer neural network for MNIST classification.
// Architecture: 784 -> hidden -> 10 (with ReLU and softmax)
type MNISTNetwork struct {
	layer1  *crossbar.Array // 784 -> hidden
	layer2  *crossbar.Array // hidden -> 10
	biases1 []float64       // Hidden layer biases
	biases2 []float64       // Output layer biases

	hiddenSize int
}

// NewMNISTNetwork creates a new MNIST network using crossbar arrays.
func NewMNISTNetwork(layer1, layer2 *crossbar.Array) *MNISTNetwork {
	hiddenSize := layer1.Rows()

	net := &MNISTNetwork{
		layer1:     layer1,
		layer2:     layer2,
		biases1:    make([]float64, hiddenSize),
		biases2:    make([]float64, 10),
		hiddenSize: hiddenSize,
	}

	// Initialize with Xavier weights
	net.initializeWeights()

	return net
}

func (n *MNISTNetwork) initializeWeights() {
	// Xavier initialization for weights
	// Layer 1: 784 inputs
	scale1 := math.Sqrt(2.0 / float64(784+n.hiddenSize))
	for i := 0; i < n.layer1.Rows(); i++ {
		for j := 0; j < n.layer1.Cols(); j++ {
			// Map to 30 levels (0-1 range)
			w := rand.NormFloat64()*scale1*0.5 + 0.5
			w = math.Max(0, math.Min(1, w))
			n.layer1.ProgramWeight(i, j, w)
		}
	}

	// Layer 2: hidden -> 10
	scale2 := math.Sqrt(2.0 / float64(n.hiddenSize+10))
	for i := 0; i < n.layer2.Rows(); i++ {
		for j := 0; j < n.layer2.Cols(); j++ {
			w := rand.NormFloat64()*scale2*0.5 + 0.5
			w = math.Max(0, math.Min(1, w))
			n.layer2.ProgramWeight(i, j, w)
		}
	}

	// Initialize biases to small random values
	for i := range n.biases1 {
		n.biases1[i] = rand.Float64()*0.1 - 0.05
	}
	for i := range n.biases2 {
		n.biases2[i] = rand.Float64()*0.1 - 0.05
	}
}

// Forward performs forward pass through the network.
// Uses dual-weight encoding: W_effective = W_pos - W_neg (simulated with offset)
func (n *MNISTNetwork) Forward(input []float64) []float64 {
	// Layer 1: MVM + bias + ReLU
	hidden, _ := n.layer1.MVM(input)
	for i := range hidden {
		// Scale up from [0,1] range and center around 0
		// Effective weight = (conductance - 0.5) * 2, giving [-1, 1] range
		hidden[i] = (hidden[i]-0.5)*4.0 + n.biases1[i]
		// ReLU
		if hidden[i] < 0 {
			hidden[i] = 0
		}
	}

	// Normalize hidden activations for stability
	maxHidden := 0.0
	for _, h := range hidden {
		if h > maxHidden {
			maxHidden = h
		}
	}
	if maxHidden > 1.0 {
		for i := range hidden {
			hidden[i] /= maxHidden
		}
	}

	// Layer 2: MVM + bias
	output, _ := n.layer2.MVM(hidden)
	for i := range output {
		output[i] = (output[i]-0.5)*4.0 + n.biases2[i]
	}

	// Softmax
	return softmax(output)
}

// Predict returns the predicted digit and confidence.
func (n *MNISTNetwork) Predict(input []float64) (int, float64) {
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

// GetOutputProbabilities returns the softmax probabilities for all classes.
func (n *MNISTNetwork) GetOutputProbabilities(input []float64) []float64 {
	return n.Forward(input)
}

// TrainEpoch runs one epoch of training using stochastic gradient descent.
func (n *MNISTNetwork) TrainEpoch(images [][]float64, labels []int, learningRate float64) float64 {
	totalLoss := 0.0

	// Shuffle indices
	indices := rand.Perm(len(images))

	for _, idx := range indices {
		input := images[idx]
		target := labels[idx]

		// Forward pass with gradient tracking
		hidden := n.forwardHidden(input)
		output := n.forwardOutput(hidden)
		probs := softmax(output)

		// Compute cross-entropy loss
		loss := -math.Log(probs[target] + 1e-10)
		totalLoss += loss

		// Backward pass (simplified gradient descent)
		// Compute output gradients
		outputGrad := make([]float64, 10)
		for i := range outputGrad {
			outputGrad[i] = probs[i]
			if i == target {
				outputGrad[i] -= 1.0
			}
		}

		// Update layer 2 weights and biases
		n.updateLayer2(hidden, outputGrad, learningRate)

		// Compute hidden gradients
		// Note: layer2Weights already fetched in updateLayer2, but we need it before update
		// This is acceptable since it's O(1) per sample, not O(n²)
		hiddenGrad := make([]float64, n.hiddenSize)
		layer2Weights := n.layer2.GetConductanceMatrix()
		for j := 0; j < n.hiddenSize; j++ {
			for i := 0; i < 10; i++ {
				// Map conductance [0,1] to effective weight [-2,2] (matches forward pass scaling)
				effectiveWeight := (layer2Weights[i][j] - 0.5) * 4.0
				hiddenGrad[j] += outputGrad[i] * effectiveWeight
			}
			// ReLU derivative (straight-through estimator)
			if hidden[j] <= 0 {
				hiddenGrad[j] = 0
			}
		}

		// Update layer 1 weights and biases
		n.updateLayer1(input, hiddenGrad, learningRate)
	}

	return totalLoss / float64(len(images))
}

func (n *MNISTNetwork) forwardHidden(input []float64) []float64 {
	hidden, _ := n.layer1.MVM(input)
	for i := range hidden {
		hidden[i] = (hidden[i]-0.5)*4.0 + n.biases1[i]
		if hidden[i] < 0 {
			hidden[i] = 0
		}
	}

	// Normalize for stability
	maxHidden := 0.0
	for _, h := range hidden {
		if h > maxHidden {
			maxHidden = h
		}
	}
	if maxHidden > 1.0 {
		for i := range hidden {
			hidden[i] /= maxHidden
		}
	}

	return hidden
}

func (n *MNISTNetwork) forwardOutput(hidden []float64) []float64 {
	output, _ := n.layer2.MVM(hidden)
	for i := range output {
		output[i] = (output[i]-0.5)*4.0 + n.biases2[i]
	}
	return output
}

func (n *MNISTNetwork) updateLayer2(hidden, grad []float64, lr float64) {
	// Fetch matrix once (was O(n³) when called inside loop)
	weights := n.layer2.GetConductanceMatrix()

	// Update weights
	for i := 0; i < 10; i++ {
		for j := 0; j < n.hiddenSize; j++ {
			w := weights[i][j]

			// Compute gradient and update
			// Use straight-through estimator: gradient flows through quantization
			dw := grad[i] * hidden[j] * lr
			newW := w - dw*0.5 // Scale and convert back to conductance range

			// ProgramWeight handles quantization to 30 levels
			n.layer2.ProgramWeight(i, j, newW)
		}

		// Update bias
		n.biases2[i] -= grad[i] * lr
	}
}

func (n *MNISTNetwork) updateLayer1(input, grad []float64, lr float64) {
	// Fetch matrix once (was O(n³) when called inside loop)
	weights := n.layer1.GetConductanceMatrix()

	// Update weights
	for i := 0; i < n.hiddenSize; i++ {
		for j := 0; j < len(input); j++ {
			w := weights[i][j]

			// Compute gradient and update
			// Use straight-through estimator: gradient flows through quantization
			dw := grad[i] * input[j] * lr
			newW := w - dw*0.5

			// ProgramWeight handles quantization to 30 levels
			n.layer1.ProgramWeight(i, j, newW)
		}

		// Update bias
		n.biases1[i] -= grad[i] * lr
	}
}

// Evaluate computes accuracy on a dataset.
func (n *MNISTNetwork) Evaluate(images [][]float64, labels []int) float64 {
	correct := 0
	for i, img := range images {
		pred, _ := n.Predict(img)
		if pred == labels[i] {
			correct++
		}
	}
	return float64(correct) / float64(len(images))
}

// SaveWeights saves the network weights to a JSON file.
func (n *MNISTNetwork) SaveWeights(filename string) error {
	data := struct {
		Layer1Weights [][]float64 `json:"layer1_weights"`
		Layer2Weights [][]float64 `json:"layer2_weights"`
		Biases1       []float64   `json:"biases1"`
		Biases2       []float64   `json:"biases2"`
	}{
		Layer1Weights: n.layer1.GetConductanceMatrix(),
		Layer2Weights: n.layer2.GetConductanceMatrix(),
		Biases1:       n.biases1,
		Biases2:       n.biases2,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, jsonData, 0644)
}

// LoadWeights loads network weights from a JSON file.
func (n *MNISTNetwork) LoadWeights(filename string) error {
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var data struct {
		Layer1Weights [][]float64 `json:"layer1_weights"`
		Layer2Weights [][]float64 `json:"layer2_weights"`
		Biases1       []float64   `json:"biases1"`
		Biases2       []float64   `json:"biases2"`
	}

	if err := json.Unmarshal(jsonData, &data); err != nil {
		return err
	}

	// Program weights to crossbar arrays
	for i, row := range data.Layer1Weights {
		for j, w := range row {
			if i < n.layer1.Rows() && j < n.layer1.Cols() {
				n.layer1.ProgramWeight(i, j, w)
			}
		}
	}

	for i, row := range data.Layer2Weights {
		for j, w := range row {
			if i < n.layer2.Rows() && j < n.layer2.Cols() {
				n.layer2.ProgramWeight(i, j, w)
			}
		}
	}

	// Copy biases
	if len(data.Biases1) > 0 {
		copy(n.biases1, data.Biases1)
	}
	if len(data.Biases2) > 0 {
		copy(n.biases2, data.Biases2)
	}

	return nil
}

// GetBiases1 returns the hidden layer biases.
func (n *MNISTNetwork) GetBiases1() []float64 {
	return n.biases1
}

// GetBiases2 returns the output layer biases.
func (n *MNISTNetwork) GetBiases2() []float64 {
	return n.biases2
}

// SetBias1 sets a hidden layer bias.
func (n *MNISTNetwork) SetBias1(i int, value float64) {
	if i >= 0 && i < len(n.biases1) {
		n.biases1[i] = value
	}
}

// SetBias2 sets an output layer bias.
func (n *MNISTNetwork) SetBias2(i int, value float64) {
	if i >= 0 && i < len(n.biases2) {
		n.biases2[i] = value
	}
}

func softmax(x []float64) []float64 {
	result := make([]float64, len(x))

	// Find max for numerical stability
	max := x[0]
	for _, v := range x {
		if v > max {
			max = v
		}
	}

	// Compute exp and sum
	var sum float64
	for i, v := range x {
		result[i] = math.Exp(v - max)
		sum += result[i]
	}

	// Normalize
	for i := range result {
		result[i] /= sum
	}

	return result
}

// GetHiddenActivations returns the hidden layer activations for an input.
func (n *MNISTNetwork) GetHiddenActivations(input []float64) []float64 {
	return n.forwardHidden(input)
}

// GetLayerActivations returns activations from all layers for visualization.
func (n *MNISTNetwork) GetLayerActivations(input []float64) ([]float64, []float64, []float64) {
	hidden := n.forwardHidden(input)
	output := n.forwardOutput(hidden)
	probs := softmax(output)
	return input, hidden, probs
}

// ComputeConfusionMatrix computes a confusion matrix for the given dataset.
func (n *MNISTNetwork) ComputeConfusionMatrix(images [][]float64, labels []int) [][]int {
	matrix := make([][]int, 10)
	for i := range matrix {
		matrix[i] = make([]int, 10)
	}

	for i, img := range images {
		pred, _ := n.Predict(img)
		actual := labels[i]
		if actual >= 0 && actual < 10 && pred >= 0 && pred < 10 {
			matrix[actual][pred]++
		}
	}

	return matrix
}

// GetPerClassMetrics returns precision, recall, F1 for each class.
func (n *MNISTNetwork) GetPerClassMetrics(confMatrix [][]int) ([]float64, []float64, []float64) {
	precision := make([]float64, 10)
	recall := make([]float64, 10)
	f1 := make([]float64, 10)

	for i := 0; i < 10; i++ {
		// True positives
		tp := float64(confMatrix[i][i])

		// False positives (predicted as i but was something else)
		var fp float64
		for j := 0; j < 10; j++ {
			if j != i {
				fp += float64(confMatrix[j][i])
			}
		}

		// False negatives (was i but predicted as something else)
		var fn float64
		for j := 0; j < 10; j++ {
			if j != i {
				fn += float64(confMatrix[i][j])
			}
		}

		// Calculate metrics
		if tp+fp > 0 {
			precision[i] = tp / (tp + fp)
		}
		if tp+fn > 0 {
			recall[i] = tp / (tp + fn)
		}
		if precision[i]+recall[i] > 0 {
			f1[i] = 2 * precision[i] * recall[i] / (precision[i] + recall[i])
		}
	}

	return precision, recall, f1
}

// QuantizeWeightsTo30Levels requantizes all weights to exactly 30 levels.
// Note: ProgramWeight now automatically quantizes, so this is mainly for explicit re-quantization.
func (n *MNISTNetwork) QuantizeWeightsTo30Levels() {
	// Layer 1 - fetch matrix once
	weights1 := n.layer1.GetConductanceMatrix()
	for i := 0; i < n.layer1.Rows(); i++ {
		for j := 0; j < n.layer1.Cols(); j++ {
			// ProgramWeight handles 30-level quantization via crossbar.QuantizeTo30Levels
			n.layer1.ProgramWeight(i, j, weights1[i][j])
		}
	}

	// Layer 2 - fetch matrix once
	weights2 := n.layer2.GetConductanceMatrix()
	for i := 0; i < n.layer2.Rows(); i++ {
		for j := 0; j < n.layer2.Cols(); j++ {
			n.layer2.ProgramWeight(i, j, weights2[i][j])
		}
	}

	fmt.Println("Weights quantized to 30 discrete levels (FeCIM format)")
}
