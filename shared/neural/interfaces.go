// Package core provides interfaces for MNIST network operations.
// interfaces.go defines abstractions for testability and modularity.
package neural

// Inferer defines the interface for neural network inference operations.
// This allows mocking for tests and enables alternative implementations.
type Inferer interface {
	// Infer performs dual-path inference returning both FP and CIM predictions.
	Infer(input []float64) *InferenceResult

	// InferFPOnly performs floating-point only inference.
	// Returns: prediction index, confidence (0-1), probability distribution
	InferFPOnly(input []float64) (prediction int, confidence float64, probs []float64)

	// InferCIMOnly performs CIM (quantized) only inference.
	// Returns: prediction index, confidence (0-1), probability distribution
	InferCIMOnly(input []float64) (prediction int, confidence float64, probs []float64)
}

// WeightLoader defines the interface for loading network weights.
type WeightLoader interface {
	// LoadWeights loads weights from a JSON file.
	LoadWeights(filename string) error

	// LoadWeightsForLevel loads QAT weights optimized for a specific quantization level.
	LoadWeightsForLevel(dataDir string, levels int) error
}

// WeightProvider defines the interface for accessing network weights.
type WeightProvider interface {
	// GetFPWeights returns the full-precision weights.
	GetFPWeights() (w1, w2 [][]float64, b1, b2 []float64)

	// GetQuantWeights returns the quantized weights.
	GetQuantWeights() (w1, w2 [][]float64, b1, b2 []float64)
}

// NetworkConfigurer defines the interface for configuring network parameters.
type NetworkConfigurer interface {
	// GetNumLevels returns the current quantization level.
	GetNumLevels() int

	// SetNumLevels sets the quantization level (2-30).
	SetNumLevels(levels int)

	// SetNoiseLevel sets the noise level as coefficient of variation.
	SetNoiseLevel(noise float64)

	// SetADCBits sets the ADC resolution in bits.
	SetADCBits(bits int)

	// SetDACBits sets the DAC resolution in bits.
	SetDACBits(bits int)

	// SetSingleLayer enables/disables single-layer (calibration) mode.
	SetSingleLayer(enabled bool)

	// IsSingleLayer returns whether single-layer mode is enabled.
	IsSingleLayer() bool
}

// DataLoader defines the interface for loading MNIST data.
type DataLoader interface {
	// LoadMNIST loads MNIST dataset from the specified directory.
	// If training is true, loads training set; otherwise loads test set.
	LoadMNIST(dataDir string, training bool) (images [][]float64, labels []int, err error)
}

// Network combines all network-related interfaces for a complete implementation.
type Network interface {
	Inferer
	WeightLoader
	WeightProvider
	NetworkConfigurer
}

// Compile-time verification that DualModeNetwork implements Network interface.
var _ Network = (*DualModeNetwork)(nil)
