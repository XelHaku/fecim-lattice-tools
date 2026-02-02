// Package gui provides Fyne-based GUI components for MNIST visualization.
// network_controller.go provides network state management and operations.
package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"fecim-lattice-tools/module3-mnist/pkg/core"
	"fecim-lattice-tools/module3-mnist/pkg/mnist"
	"fecim-lattice-tools/shared/utils"
)

// NetworkController manages the neural network state and operations.
// Extracted from DualModeApp to separate concerns (ARCH-001).
type NetworkController struct {
	// Network
	network    *core.DualModeNetwork
	dataDir    string // MNIST dataset directory
	weightsDir string // Quantized weight files directory

	// Test data cache
	testImages [][]float64
	testLabels []int

	// QAT (Quantization-Aware Training) weight tracking
	currentQATLevel int
	qatLevelMu      sync.RWMutex

	// Track which missing weight levels have already shown a warning
	warnedMissingLevels   map[int]bool
	warnedMissingLevelsMu sync.RWMutex

	// Callbacks for UI notifications
	onWeightsLoaded func(level int)
	onError         func(err error)
}

// NewNetworkController creates a new network controller.
func NewNetworkController(inputSize, hiddenSize, outputSize int) *NetworkController {
	dataDir := utils.FindModuleDataDir("module3-mnist", "train-images-idx3-ubyte.gz")
	if dataDir == "" {
		dataDir = "module3-mnist/data" // Default fallback
	}

	weightsDir := ""
	weightCandidates := []string{
		filepath.Join("data", "pretrained-weigths"),
		filepath.Join("data", "pretrained-weights"),
		filepath.Join("module3-mnist", "data"),
	}
	for _, candidate := range weightCandidates {
		if _, err := os.Stat(filepath.Join(candidate, "pretrained_weights.json")); err == nil {
			weightsDir = candidate
			break
		}
	}
	if weightsDir == "" {
		weightsDir = dataDir
	}

	nc := &NetworkController{
		dataDir:             dataDir,
		weightsDir:          weightsDir,
		currentQATLevel:     FeCIMDefaultLevels,
		warnedMissingLevels: make(map[int]bool),
	}

	// Create network
	nc.network = core.NewDualModeNetwork(inputSize, hiddenSize, outputSize)

	// Load pretrained weights
	weightsPath := filepath.Join(nc.weightsDir, "pretrained_weights.json")
	if _, err := os.Stat(weightsPath); err == nil {
		if err := nc.network.LoadWeights(weightsPath); err != nil {
			mnistLog.Printf("Warning: Failed to load weights from %s: %v", weightsPath, err)
		}
	} else {
		mnistLog.Printf("Note: No pretrained weights found at %s, using random initialization", weightsPath)
	}

	return nc
}

// Network returns the underlying DualModeNetwork.
func (nc *NetworkController) Network() *core.DualModeNetwork {
	return nc.network
}

// DataDir returns the data directory path.
func (nc *NetworkController) DataDir() string {
	return nc.dataDir
}

// WeightsDir returns the weights directory path.
func (nc *NetworkController) WeightsDir() string {
	return nc.weightsDir
}

// Infer runs dual-path inference on the input.
func (nc *NetworkController) Infer(input []float64) *core.InferenceResult {
	return nc.network.Infer(input)
}

// GetQuantWeights returns the quantized weights.
func (nc *NetworkController) GetQuantWeights() (w1, w2 [][]float64, b1, b2 []float64) {
	return nc.network.GetQuantWeights()
}

// GetFPWeights returns the full-precision weights.
func (nc *NetworkController) GetFPWeights() (w1, w2 [][]float64, b1, b2 []float64) {
	return nc.network.GetFPWeights()
}

// GetNumLevels returns the current quantization level.
func (nc *NetworkController) GetNumLevels() int {
	return nc.network.GetNumLevels()
}

// SetNumLevels sets the quantization level.
func (nc *NetworkController) SetNumLevels(levels int) {
	nc.network.SetNumLevels(levels)
}

// SetNoiseLevel sets the noise level.
func (nc *NetworkController) SetNoiseLevel(noise float64) {
	nc.network.SetNoiseLevel(noise)
}

// SetADCBits sets the ADC resolution.
func (nc *NetworkController) SetADCBits(bits int) {
	nc.network.SetADCBits(bits)
}

// SetDACBits sets the DAC resolution.
func (nc *NetworkController) SetDACBits(bits int) {
	nc.network.SetDACBits(bits)
}

// SetSingleLayer enables/disables single-layer mode.
func (nc *NetworkController) SetSingleLayer(enabled bool) {
	nc.network.SetSingleLayer(enabled)
}

// IsSingleLayer returns whether single-layer mode is enabled.
func (nc *NetworkController) IsSingleLayer() bool {
	return nc.network.IsSingleLayer()
}

// CurrentQATLevel returns the currently loaded QAT weight level.
func (nc *NetworkController) CurrentQATLevel() int {
	nc.qatLevelMu.RLock()
	defer nc.qatLevelMu.RUnlock()
	return nc.currentQATLevel
}

// SetCurrentQATLevel sets the current QAT level.
func (nc *NetworkController) SetCurrentQATLevel(level int) {
	nc.qatLevelMu.Lock()
	defer nc.qatLevelMu.Unlock()
	nc.currentQATLevel = level
}

// LoadWeightsForLevel loads QAT weights for the specified level.
func (nc *NetworkController) LoadWeightsForLevel(levels int) error {
	return nc.network.LoadWeightsForLevel(nc.dataDir, levels)
}

// HasWarnedMissingLevel returns true if a warning was shown for this level.
func (nc *NetworkController) HasWarnedMissingLevel(level int) bool {
	nc.warnedMissingLevelsMu.RLock()
	defer nc.warnedMissingLevelsMu.RUnlock()
	return nc.warnedMissingLevels[level]
}

// SetWarnedMissingLevel marks a level as warned.
func (nc *NetworkController) SetWarnedMissingLevel(level int) {
	nc.warnedMissingLevelsMu.Lock()
	defer nc.warnedMissingLevelsMu.Unlock()
	nc.warnedMissingLevels[level] = true
}

// ClearWarnedMissingLevel clears the warning flag for a level.
func (nc *NetworkController) ClearWarnedMissingLevel(level int) {
	nc.warnedMissingLevelsMu.Lock()
	defer nc.warnedMissingLevelsMu.Unlock()
	nc.warnedMissingLevels[level] = false
}

// ChangeHiddenSize creates a new network with a different hidden size.
func (nc *NetworkController) ChangeHiddenSize(hiddenSize int) error {
	// Create new network
	newNetwork := core.NewDualModeNetwork(MNISTInputSize, hiddenSize, MNISTOutputSize)

	// Try to load QAT weights for current level
	if err := newNetwork.LoadWeightsForLevel(nc.dataDir, nc.CurrentQATLevel()); err != nil {
		// Try default level
		if err := newNetwork.LoadWeightsForLevel(nc.dataDir, FeCIMDefaultLevels); err != nil {
			return fmt.Errorf("failed to load weights for hidden size %d: %w", hiddenSize, err)
		}
		nc.SetCurrentQATLevel(FeCIMDefaultLevels)
	}

	// Replace network
	nc.network = newNetwork
	return nil
}

// LoadTestData loads MNIST test data.
func (nc *NetworkController) LoadTestData() error {
	if len(nc.testImages) > 0 {
		return nil // Already loaded
	}

	images, labels, err := mnist.LoadMNIST(nc.dataDir, false)
	if err != nil {
		return err
	}

	nc.testImages = images
	nc.testLabels = labels
	return nil
}

// GetTestSample returns a test sample at the given index.
func (nc *NetworkController) GetTestSample(index int) ([]float64, int, error) {
	if err := nc.LoadTestData(); err != nil {
		return nil, 0, err
	}

	if index < 0 || index >= len(nc.testImages) {
		return nil, 0, fmt.Errorf("index %d out of range [0, %d)", index, len(nc.testImages))
	}

	return nc.testImages[index], nc.testLabels[index], nil
}

// TestDataSize returns the number of test samples.
func (nc *NetworkController) TestDataSize() int {
	return len(nc.testImages)
}

// GetTestData returns all test data (images and labels).
func (nc *NetworkController) GetTestData() ([][]float64, []int) {
	return nc.testImages, nc.testLabels
}

// SetOnWeightsLoaded sets the callback for weight loading events.
func (nc *NetworkController) SetOnWeightsLoaded(callback func(level int)) {
	nc.onWeightsLoaded = callback
}

// SetOnError sets the callback for error events.
func (nc *NetworkController) SetOnError(callback func(err error)) {
	nc.onError = callback
}

// QATLoadResult represents the result of a QAT weight loading attempt.
type QATLoadResult int

const (
	// QATAlreadyLoaded indicates weights for this level are already loaded.
	QATAlreadyLoaded QATLoadResult = iota
	// QATLoaded indicates weights were successfully loaded.
	QATLoaded
	// QATNotFound indicates no weights file exists for this level.
	QATNotFound
	// QATNotFoundFirstWarning indicates first warning for missing weights.
	QATNotFoundFirstWarning
	// QATLoadError indicates an error occurred loading weights.
	QATLoadError
)

// TryLoadQATWeights attempts to load QAT weights for the given level.
// Returns the result and any error that occurred.
func (nc *NetworkController) TryLoadQATWeights(targetLevel int) (QATLoadResult, error) {
	// Check if we already have optimal weights loaded
	if nc.CurrentQATLevel() == targetLevel {
		return QATAlreadyLoaded, nil
	}

	// Find the weights file for this level
	weightsPath := core.GetWeightsFilename(nc.weightsDir, targetLevel)

	// Check if the file exists
	if _, err := os.Stat(weightsPath); os.IsNotExist(err) {
		// No level-specific weights
		alreadyWarned := nc.HasWarnedMissingLevel(targetLevel)
		if !alreadyWarned {
			nc.SetWarnedMissingLevel(targetLevel)
			return QATNotFoundFirstWarning, nil
		}
		return QATNotFound, nil
	}

	// Load the new weights
	if err := nc.network.LoadWeights(weightsPath); err != nil {
		return QATLoadError, err
	}

	// Update tracking
	nc.SetCurrentQATLevel(targetLevel)
	return QATLoaded, nil
}
