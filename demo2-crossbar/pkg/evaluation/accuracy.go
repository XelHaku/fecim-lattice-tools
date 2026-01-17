// Package evaluation provides accuracy and performance metrics for neural network inference.
package evaluation

import (
	"fmt"
	"math"
	"time"
)

// Metrics holds evaluation metrics for a neural network.
type Metrics struct {
	Accuracy    float64 // Classification accuracy (0-1)
	TopKAcc     float64 // Top-K accuracy
	Loss        float64 // Cross-entropy loss
	Correct     int     // Number of correct predictions
	Total       int     // Total predictions
	ConfMatrix  [][]int // Confusion matrix [actual][predicted]
	NumClasses  int
	InferenceNs int64 // Average inference time in nanoseconds
}

// NewMetrics creates a new metrics tracker.
func NewMetrics(numClasses int) *Metrics {
	confMatrix := make([][]int, numClasses)
	for i := range confMatrix {
		confMatrix[i] = make([]int, numClasses)
	}
	return &Metrics{
		NumClasses: numClasses,
		ConfMatrix: confMatrix,
	}
}

// Update adds a new prediction to the metrics.
func (m *Metrics) Update(predicted, actual int, probabilities []float64) {
	m.Total++

	if predicted == actual {
		m.Correct++
	}

	// Update confusion matrix
	if actual >= 0 && actual < m.NumClasses && predicted >= 0 && predicted < m.NumClasses {
		m.ConfMatrix[actual][predicted]++
	}

	// Update cross-entropy loss
	if actual >= 0 && actual < len(probabilities) {
		p := probabilities[actual]
		if p > 0 {
			m.Loss -= math.Log(p)
		} else {
			m.Loss -= math.Log(1e-10) // Avoid log(0)
		}
	}

	// Recalculate accuracy
	if m.Total > 0 {
		m.Accuracy = float64(m.Correct) / float64(m.Total)
		m.Loss /= float64(m.Total)
	}
}

// UpdateWithTiming adds a prediction with timing information.
func (m *Metrics) UpdateWithTiming(predicted, actual int, probabilities []float64, inferenceTime time.Duration) {
	m.Update(predicted, actual, probabilities)
	m.InferenceNs = (m.InferenceNs*int64(m.Total-1) + inferenceTime.Nanoseconds()) / int64(m.Total)
}

// Reset clears all metrics.
func (m *Metrics) Reset() {
	m.Accuracy = 0
	m.TopKAcc = 0
	m.Loss = 0
	m.Correct = 0
	m.Total = 0
	m.InferenceNs = 0
	for i := range m.ConfMatrix {
		for j := range m.ConfMatrix[i] {
			m.ConfMatrix[i][j] = 0
		}
	}
}

// GetPrecision returns precision for a specific class.
func (m *Metrics) GetPrecision(class int) float64 {
	if class < 0 || class >= m.NumClasses {
		return 0
	}

	truePositives := m.ConfMatrix[class][class]
	predictedPositives := 0
	for i := 0; i < m.NumClasses; i++ {
		predictedPositives += m.ConfMatrix[i][class]
	}

	if predictedPositives == 0 {
		return 0
	}
	return float64(truePositives) / float64(predictedPositives)
}

// GetRecall returns recall (sensitivity) for a specific class.
func (m *Metrics) GetRecall(class int) float64 {
	if class < 0 || class >= m.NumClasses {
		return 0
	}

	truePositives := m.ConfMatrix[class][class]
	actualPositives := 0
	for j := 0; j < m.NumClasses; j++ {
		actualPositives += m.ConfMatrix[class][j]
	}

	if actualPositives == 0 {
		return 0
	}
	return float64(truePositives) / float64(actualPositives)
}

// GetF1Score returns F1 score for a specific class.
func (m *Metrics) GetF1Score(class int) float64 {
	precision := m.GetPrecision(class)
	recall := m.GetRecall(class)

	if precision+recall == 0 {
		return 0
	}
	return 2 * precision * recall / (precision + recall)
}

// GetMacroF1 returns the macro-averaged F1 score.
func (m *Metrics) GetMacroF1() float64 {
	var sum float64
	for c := 0; c < m.NumClasses; c++ {
		sum += m.GetF1Score(c)
	}
	return sum / float64(m.NumClasses)
}

// String returns a formatted summary of the metrics.
func (m *Metrics) String() string {
	return fmt.Sprintf(
		"Accuracy: %.2f%% (%d/%d)\n"+
			"Loss: %.4f\n"+
			"Macro F1: %.4f\n"+
			"Avg Inference: %.2f µs",
		m.Accuracy*100, m.Correct, m.Total,
		m.Loss,
		m.GetMacroF1(),
		float64(m.InferenceNs)/1000.0,
	)
}

// PrintConfusionMatrix prints the confusion matrix.
func (m *Metrics) PrintConfusionMatrix() {
	fmt.Println("\nConfusion Matrix:")
	fmt.Print("     ")
	for j := 0; j < m.NumClasses; j++ {
		fmt.Printf("%4d ", j)
	}
	fmt.Println()
	fmt.Println("     " + string(make([]byte, m.NumClasses*5)))

	for i := 0; i < m.NumClasses; i++ {
		fmt.Printf("%3d |", i)
		for j := 0; j < m.NumClasses; j++ {
			fmt.Printf("%4d ", m.ConfMatrix[i][j])
		}
		fmt.Println()
	}
}

// PrintPerClassMetrics prints precision, recall, and F1 for each class.
func (m *Metrics) PrintPerClassMetrics() {
	fmt.Println("\nPer-Class Metrics:")
	fmt.Println("Class  Precision  Recall  F1-Score")
	fmt.Println("-----  ---------  ------  --------")
	for c := 0; c < m.NumClasses; c++ {
		fmt.Printf("%5d  %9.4f  %6.4f  %8.4f\n",
			c, m.GetPrecision(c), m.GetRecall(c), m.GetF1Score(c))
	}
}

// ArgMax returns the index of the maximum value.
func ArgMax(values []float64) int {
	maxIdx := 0
	maxVal := values[0]
	for i, v := range values {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}
	return maxIdx
}

// TopK returns indices of the top K values.
func TopK(values []float64, k int) []int {
	if k > len(values) {
		k = len(values)
	}

	// Create index-value pairs
	type pair struct {
		idx int
		val float64
	}
	pairs := make([]pair, len(values))
	for i, v := range values {
		pairs[i] = pair{i, v}
	}

	// Simple selection sort for top K
	for i := 0; i < k; i++ {
		maxIdx := i
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].val > pairs[maxIdx].val {
				maxIdx = j
			}
		}
		pairs[i], pairs[maxIdx] = pairs[maxIdx], pairs[i]
	}

	result := make([]int, k)
	for i := 0; i < k; i++ {
		result[i] = pairs[i].idx
	}
	return result
}

// IsInTopK checks if the actual class is in the top K predictions.
func IsInTopK(values []float64, actual, k int) bool {
	topK := TopK(values, k)
	for _, idx := range topK {
		if idx == actual {
			return true
		}
	}
	return false
}
