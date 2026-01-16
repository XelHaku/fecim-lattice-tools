// Package layers provides neural network layer implementations for CIM simulation.
// loss.go implements common loss functions for training and evaluation.
package layers

import (
	"math"
)

// =============================================================================
// Loss Function Interface
// =============================================================================

// Loss defines the interface for loss functions
type Loss interface {
	// Forward computes the loss value
	Forward(predictions, targets []float64) float64

	// Backward computes the gradient w.r.t. predictions
	Backward(predictions, targets []float64) []float64

	// Name returns the loss function name
	Name() string
}

// =============================================================================
// Mean Squared Error (MSE)
// =============================================================================

// MSELoss implements mean squared error loss
// L = (1/n) * Σ(y_pred - y_true)²
type MSELoss struct {
	Reduction string // "mean", "sum", "none"
}

// NewMSELoss creates a new MSE loss
func NewMSELoss() *MSELoss {
	return &MSELoss{Reduction: "mean"}
}

// Forward computes MSE loss
func (l *MSELoss) Forward(predictions, targets []float64) float64 {
	n := len(predictions)
	if n == 0 {
		return 0
	}

	sum := 0.0
	for i := 0; i < n; i++ {
		diff := predictions[i] - targets[i]
		sum += diff * diff
	}

	switch l.Reduction {
	case "sum":
		return sum
	case "none":
		return sum // Return total for consistency
	default: // "mean"
		return sum / float64(n)
	}
}

// Backward computes gradient of MSE loss
func (l *MSELoss) Backward(predictions, targets []float64) []float64 {
	n := len(predictions)
	grad := make([]float64, n)

	for i := 0; i < n; i++ {
		grad[i] = 2 * (predictions[i] - targets[i])
		if l.Reduction == "mean" {
			grad[i] /= float64(n)
		}
	}

	return grad
}

func (l *MSELoss) Name() string { return "mse" }

// =============================================================================
// Mean Absolute Error (MAE)
// =============================================================================

// MAELoss implements mean absolute error loss (L1 loss)
// L = (1/n) * Σ|y_pred - y_true|
type MAELoss struct {
	Reduction string
}

// NewMAELoss creates a new MAE loss
func NewMAELoss() *MAELoss {
	return &MAELoss{Reduction: "mean"}
}

// Forward computes MAE loss
func (l *MAELoss) Forward(predictions, targets []float64) float64 {
	n := len(predictions)
	if n == 0 {
		return 0
	}

	sum := 0.0
	for i := 0; i < n; i++ {
		sum += math.Abs(predictions[i] - targets[i])
	}

	switch l.Reduction {
	case "sum":
		return sum
	default:
		return sum / float64(n)
	}
}

// Backward computes gradient of MAE loss
func (l *MAELoss) Backward(predictions, targets []float64) []float64 {
	n := len(predictions)
	grad := make([]float64, n)

	for i := 0; i < n; i++ {
		if predictions[i] > targets[i] {
			grad[i] = 1.0
		} else if predictions[i] < targets[i] {
			grad[i] = -1.0
		} else {
			grad[i] = 0.0 // Subgradient at 0
		}
		if l.Reduction == "mean" {
			grad[i] /= float64(n)
		}
	}

	return grad
}

func (l *MAELoss) Name() string { return "mae" }

// =============================================================================
// Huber Loss (Smooth L1)
// =============================================================================

// HuberLoss implements Huber loss (smooth combination of MSE and MAE)
// L = 0.5*(y-t)² for |y-t| < delta
// L = delta*(|y-t| - 0.5*delta) otherwise
type HuberLoss struct {
	Delta     float64
	Reduction string
}

// NewHuberLoss creates a new Huber loss
func NewHuberLoss(delta float64) *HuberLoss {
	return &HuberLoss{Delta: delta, Reduction: "mean"}
}

// Forward computes Huber loss
func (l *HuberLoss) Forward(predictions, targets []float64) float64 {
	n := len(predictions)
	if n == 0 {
		return 0
	}

	sum := 0.0
	for i := 0; i < n; i++ {
		diff := math.Abs(predictions[i] - targets[i])
		if diff < l.Delta {
			sum += 0.5 * diff * diff
		} else {
			sum += l.Delta * (diff - 0.5*l.Delta)
		}
	}

	switch l.Reduction {
	case "sum":
		return sum
	default:
		return sum / float64(n)
	}
}

// Backward computes gradient of Huber loss
func (l *HuberLoss) Backward(predictions, targets []float64) []float64 {
	n := len(predictions)
	grad := make([]float64, n)

	for i := 0; i < n; i++ {
		diff := predictions[i] - targets[i]
		if math.Abs(diff) < l.Delta {
			grad[i] = diff
		} else {
			if diff > 0 {
				grad[i] = l.Delta
			} else {
				grad[i] = -l.Delta
			}
		}
		if l.Reduction == "mean" {
			grad[i] /= float64(n)
		}
	}

	return grad
}

func (l *HuberLoss) Name() string { return "huber" }

// =============================================================================
// Cross-Entropy Loss
// =============================================================================

// CrossEntropyLoss implements cross-entropy loss for classification
// L = -Σ t_i * log(p_i) where p = softmax(predictions)
type CrossEntropyLoss struct {
	Reduction string
}

// NewCrossEntropyLoss creates a new cross-entropy loss
func NewCrossEntropyLoss() *CrossEntropyLoss {
	return &CrossEntropyLoss{Reduction: "mean"}
}

// Forward computes cross-entropy loss
// predictions: logits (pre-softmax), targets: one-hot or class indices
func (l *CrossEntropyLoss) Forward(predictions, targets []float64) float64 {
	// Apply softmax to predictions
	probs := softmaxVec(predictions)

	// Compute cross-entropy
	loss := 0.0
	for i := range probs {
		if targets[i] > 0 {
			// Add small epsilon for numerical stability
			loss -= targets[i] * math.Log(math.Max(probs[i], 1e-15))
		}
	}

	return loss
}

// Backward computes gradient of cross-entropy w.r.t. logits
func (l *CrossEntropyLoss) Backward(predictions, targets []float64) []float64 {
	probs := softmaxVec(predictions)
	grad := make([]float64, len(predictions))

	for i := range grad {
		grad[i] = probs[i] - targets[i]
	}

	return grad
}

func (l *CrossEntropyLoss) Name() string { return "cross_entropy" }

// softmaxVec computes softmax of a vector
func softmaxVec(x []float64) []float64 {
	n := len(x)
	result := make([]float64, n)

	// Find max for numerical stability
	maxVal := x[0]
	for _, v := range x {
		if v > maxVal {
			maxVal = v
		}
	}

	// Compute exp and sum
	sum := 0.0
	for i := 0; i < n; i++ {
		result[i] = math.Exp(x[i] - maxVal)
		sum += result[i]
	}

	// Normalize
	for i := 0; i < n; i++ {
		result[i] /= sum
	}

	return result
}

// =============================================================================
// Binary Cross-Entropy Loss
// =============================================================================

// BCELoss implements binary cross-entropy loss
// L = -[t*log(p) + (1-t)*log(1-p)]
type BCELoss struct {
	Reduction string
}

// NewBCELoss creates a new binary cross-entropy loss
func NewBCELoss() *BCELoss {
	return &BCELoss{Reduction: "mean"}
}

// Forward computes BCE loss (predictions should be in [0,1])
func (l *BCELoss) Forward(predictions, targets []float64) float64 {
	n := len(predictions)
	if n == 0 {
		return 0
	}

	sum := 0.0
	eps := 1e-15
	for i := 0; i < n; i++ {
		p := math.Max(math.Min(predictions[i], 1-eps), eps)
		sum -= targets[i]*math.Log(p) + (1-targets[i])*math.Log(1-p)
	}

	switch l.Reduction {
	case "sum":
		return sum
	default:
		return sum / float64(n)
	}
}

// Backward computes gradient of BCE loss
func (l *BCELoss) Backward(predictions, targets []float64) []float64 {
	n := len(predictions)
	grad := make([]float64, n)
	eps := 1e-15

	for i := 0; i < n; i++ {
		p := math.Max(math.Min(predictions[i], 1-eps), eps)
		grad[i] = (p - targets[i]) / (p * (1 - p))
		if l.Reduction == "mean" {
			grad[i] /= float64(n)
		}
	}

	return grad
}

func (l *BCELoss) Name() string { return "bce" }

// =============================================================================
// BCE with Logits Loss
// =============================================================================

// BCEWithLogitsLoss combines sigmoid + BCE for numerical stability
// L = -[t*log(sigmoid(x)) + (1-t)*log(1-sigmoid(x))]
//   = max(x, 0) - x*t + log(1 + exp(-|x|))
type BCEWithLogitsLoss struct {
	Reduction string
}

// NewBCEWithLogitsLoss creates a new BCE with logits loss
func NewBCEWithLogitsLoss() *BCEWithLogitsLoss {
	return &BCEWithLogitsLoss{Reduction: "mean"}
}

// Forward computes BCE with logits loss
func (l *BCEWithLogitsLoss) Forward(logits, targets []float64) float64 {
	n := len(logits)
	if n == 0 {
		return 0
	}

	sum := 0.0
	for i := 0; i < n; i++ {
		x := logits[i]
		t := targets[i]
		// Numerically stable formula
		sum += math.Max(x, 0) - x*t + math.Log(1+math.Exp(-math.Abs(x)))
	}

	switch l.Reduction {
	case "sum":
		return sum
	default:
		return sum / float64(n)
	}
}

// Backward computes gradient
func (l *BCEWithLogitsLoss) Backward(logits, targets []float64) []float64 {
	n := len(logits)
	grad := make([]float64, n)

	for i := 0; i < n; i++ {
		// Gradient is sigmoid(x) - t
		p := 1.0 / (1.0 + math.Exp(-logits[i]))
		grad[i] = p - targets[i]
		if l.Reduction == "mean" {
			grad[i] /= float64(n)
		}
	}

	return grad
}

func (l *BCEWithLogitsLoss) Name() string { return "bce_with_logits" }

// =============================================================================
// Hinge Loss (SVM)
// =============================================================================

// HingeLoss implements hinge loss for SVM
// L = max(0, 1 - t*y) where t ∈ {-1, +1}
type HingeLoss struct {
	Reduction string
}

// NewHingeLoss creates a new hinge loss
func NewHingeLoss() *HingeLoss {
	return &HingeLoss{Reduction: "mean"}
}

// Forward computes hinge loss
func (l *HingeLoss) Forward(predictions, targets []float64) float64 {
	n := len(predictions)
	if n == 0 {
		return 0
	}

	sum := 0.0
	for i := 0; i < n; i++ {
		sum += math.Max(0, 1-targets[i]*predictions[i])
	}

	switch l.Reduction {
	case "sum":
		return sum
	default:
		return sum / float64(n)
	}
}

// Backward computes gradient of hinge loss
func (l *HingeLoss) Backward(predictions, targets []float64) []float64 {
	n := len(predictions)
	grad := make([]float64, n)

	for i := 0; i < n; i++ {
		if targets[i]*predictions[i] < 1 {
			grad[i] = -targets[i]
		} else {
			grad[i] = 0
		}
		if l.Reduction == "mean" {
			grad[i] /= float64(n)
		}
	}

	return grad
}

func (l *HingeLoss) Name() string { return "hinge" }

// =============================================================================
// Cosine Embedding Loss
// =============================================================================

// CosineEmbeddingLoss implements loss for similarity learning
// L = 1 - cos(x1, x2) if y = 1
// L = max(0, cos(x1, x2) - margin) if y = -1
type CosineEmbeddingLoss struct {
	Margin    float64
	Reduction string
}

// NewCosineEmbeddingLoss creates a new cosine embedding loss
func NewCosineEmbeddingLoss(margin float64) *CosineEmbeddingLoss {
	return &CosineEmbeddingLoss{Margin: margin, Reduction: "mean"}
}

// CosineSimilarity computes cosine similarity between two vectors
func CosineSimilarity(x1, x2 []float64) float64 {
	dot := 0.0
	norm1 := 0.0
	norm2 := 0.0

	for i := range x1 {
		dot += x1[i] * x2[i]
		norm1 += x1[i] * x1[i]
		norm2 += x2[i] * x2[i]
	}

	norm1 = math.Sqrt(norm1)
	norm2 = math.Sqrt(norm2)

	if norm1 < 1e-8 || norm2 < 1e-8 {
		return 0
	}

	return dot / (norm1 * norm2)
}

// =============================================================================
// KL Divergence Loss
// =============================================================================

// KLDivLoss implements Kullback-Leibler divergence loss
// L = Σ t * (log(t) - log(p))
type KLDivLoss struct {
	Reduction string
	LogTarget bool // If true, targets are already log probabilities
}

// NewKLDivLoss creates a new KL divergence loss
func NewKLDivLoss() *KLDivLoss {
	return &KLDivLoss{Reduction: "mean", LogTarget: false}
}

// Forward computes KL divergence (predictions are log probabilities)
func (l *KLDivLoss) Forward(logPredictions, targets []float64) float64 {
	n := len(logPredictions)
	if n == 0 {
		return 0
	}

	sum := 0.0
	for i := 0; i < n; i++ {
		if targets[i] > 0 {
			sum += targets[i] * (math.Log(targets[i]) - logPredictions[i])
		}
	}

	switch l.Reduction {
	case "sum":
		return sum
	case "batchmean":
		return sum // Return sum for batch mean (divide externally)
	default:
		return sum / float64(n)
	}
}

// Backward computes gradient of KL divergence
func (l *KLDivLoss) Backward(logPredictions, targets []float64) []float64 {
	n := len(logPredictions)
	grad := make([]float64, n)

	for i := 0; i < n; i++ {
		grad[i] = -targets[i]
		if l.Reduction == "mean" {
			grad[i] /= float64(n)
		}
	}

	return grad
}

func (l *KLDivLoss) Name() string { return "kl_div" }

// =============================================================================
// Contrastive Loss
// =============================================================================

// ContrastiveLoss implements contrastive loss for metric learning
// L = (1-y) * 0.5 * D² + y * 0.5 * max(0, margin - D)²
// where D is the distance between embeddings
type ContrastiveLoss struct {
	Margin    float64
	Reduction string
}

// NewContrastiveLoss creates a new contrastive loss
func NewContrastiveLoss(margin float64) *ContrastiveLoss {
	return &ContrastiveLoss{Margin: margin, Reduction: "mean"}
}

// ForwardPair computes contrastive loss for a pair of embeddings
// y=0 means similar pair, y=1 means dissimilar pair
func (l *ContrastiveLoss) ForwardPair(x1, x2 []float64, y float64) float64 {
	// Compute Euclidean distance
	distSq := 0.0
	for i := range x1 {
		diff := x1[i] - x2[i]
		distSq += diff * diff
	}
	dist := math.Sqrt(distSq)

	// Contrastive loss
	loss := (1-y)*0.5*distSq + y*0.5*math.Pow(math.Max(0, l.Margin-dist), 2)
	return loss
}

// =============================================================================
// Triplet Loss
// =============================================================================

// TripletLoss implements triplet loss for metric learning
// L = max(0, D(anchor, positive) - D(anchor, negative) + margin)
type TripletLoss struct {
	Margin    float64
	P         float64 // Norm power (1 for L1, 2 for L2)
	Reduction string
}

// NewTripletLoss creates a new triplet loss
func NewTripletLoss(margin float64) *TripletLoss {
	return &TripletLoss{Margin: margin, P: 2, Reduction: "mean"}
}

// ForwardTriplet computes triplet loss
func (l *TripletLoss) ForwardTriplet(anchor, positive, negative []float64) float64 {
	// Compute distances
	distPos := l.distance(anchor, positive)
	distNeg := l.distance(anchor, negative)

	// Triplet loss
	return math.Max(0, distPos-distNeg+l.Margin)
}

func (l *TripletLoss) distance(x1, x2 []float64) float64 {
	sum := 0.0
	for i := range x1 {
		diff := math.Abs(x1[i] - x2[i])
		sum += math.Pow(diff, l.P)
	}
	return math.Pow(sum, 1.0/l.P)
}

// =============================================================================
// Loss Registry
// =============================================================================

// LossRegistry provides name-based loss lookup
var LossRegistry = map[string]func() Loss{
	"mse":              func() Loss { return NewMSELoss() },
	"mae":              func() Loss { return NewMAELoss() },
	"l1":               func() Loss { return NewMAELoss() },
	"l2":               func() Loss { return NewMSELoss() },
	"huber":            func() Loss { return NewHuberLoss(1.0) },
	"smooth_l1":        func() Loss { return NewHuberLoss(1.0) },
	"cross_entropy":    func() Loss { return NewCrossEntropyLoss() },
	"bce":              func() Loss { return NewBCELoss() },
	"bce_with_logits":  func() Loss { return NewBCEWithLogitsLoss() },
	"hinge":            func() Loss { return NewHingeLoss() },
	"kl_div":           func() Loss { return NewKLDivLoss() },
}

// GetLoss returns a loss function by name
func GetLoss(name string) Loss {
	if creator, ok := LossRegistry[name]; ok {
		return creator()
	}
	return NewMSELoss() // Default to MSE
}

// =============================================================================
// Accuracy Metrics
// =============================================================================

// Accuracy computes classification accuracy
func Accuracy(predictions []int, targets []int) float64 {
	if len(predictions) != len(targets) || len(predictions) == 0 {
		return 0
	}

	correct := 0
	for i := range predictions {
		if predictions[i] == targets[i] {
			correct++
		}
	}

	return float64(correct) / float64(len(predictions))
}

// TopKAccuracy computes top-k accuracy
func TopKAccuracy(scores [][]float64, targets []int, k int) float64 {
	if len(scores) != len(targets) || len(scores) == 0 {
		return 0
	}

	correct := 0
	for i, score := range scores {
		topK := topKIndices(score, k)
		for _, idx := range topK {
			if idx == targets[i] {
				correct++
				break
			}
		}
	}

	return float64(correct) / float64(len(targets))
}

// topKIndices returns indices of top-k values
func topKIndices(arr []float64, k int) []int {
	n := len(arr)
	if k > n {
		k = n
	}

	// Simple selection (for small k)
	indices := make([]int, k)
	selected := make(map[int]bool)

	for j := 0; j < k; j++ {
		maxIdx := -1
		maxVal := math.Inf(-1)
		for i := 0; i < n; i++ {
			if !selected[i] && arr[i] > maxVal {
				maxVal = arr[i]
				maxIdx = i
			}
		}
		indices[j] = maxIdx
		selected[maxIdx] = true
	}

	return indices
}

// ArgMax returns the index of maximum value
func ArgMax(arr []float64) int {
	maxIdx := 0
	maxVal := arr[0]
	for i, v := range arr {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}
	return maxIdx
}

// ArgMin returns the index of minimum value
func ArgMin(arr []float64) int {
	minIdx := 0
	minVal := arr[0]
	for i, v := range arr {
		if v < minVal {
			minVal = v
			minIdx = i
		}
	}
	return minIdx
}
