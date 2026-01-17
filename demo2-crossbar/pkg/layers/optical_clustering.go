// Package layers provides neural network layer implementations for CIM simulation.
// optical_clustering.go implements optical/photonic CIM interconnects and
// weight clustering quantization for ferroelectric CIM accelerators.
//
// Research basis:
// - Photonic tensor cores: 880 TOPS/mm², 5.1 TOPS/W
// - MZI-based matrix multiplication (SVD decomposition)
// - WDM: Wavelength division multiplexing for parallelism
// - Weight clustering: K-means, 5.27× compression
// - Column-wise quantization: 222% efficiency improvement
// - CIM²PQ: Arraywise mixed precision quantization
//
// Key concepts:
// - MZI: Mach-Zehnder Interferometer for coherent computing
// - WDM: Multiple wavelengths for spatial sharing
// - K-means: Cluster weights to codebook entries
// - Column-wise: Fine-grained quantization per crossbar column
package layers

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// =============================================================================
// OPTICAL/PHOTONIC CIM INTERCONNECTS
// =============================================================================

// OpticalConfig configures optical/photonic system
type OpticalConfig struct {
	// MZI parameters
	NumMZI         int     // Number of MZI stages
	PhaseResolution int    // Phase shifter bits
	InsertionLoss  float64 // dB per MZI

	// WDM parameters
	NumWavelengths int     // Wavelength channels
	ChannelSpacing float64 // nm between channels
	CenterWavelength float64 // nm (e.g., 1550nm C-band)

	// Performance
	ClockSpeedGHz  float64 // Modulation rate
	LaserPowerMW   float64 // Input laser power
	DetectorSensitivity float64 // A/W

	// Architecture
	CrossbarSize   int     // Logical crossbar size
}

// DefaultOpticalConfig returns typical optical configuration
func DefaultOpticalConfig() *OpticalConfig {
	return &OpticalConfig{
		NumMZI:           6,
		PhaseResolution:  8,
		InsertionLoss:    0.3,
		NumWavelengths:   64,
		ChannelSpacing:   0.8,
		CenterWavelength: 1550,
		ClockSpeedGHz:    25.0,
		LaserPowerMW:     10.0,
		DetectorSensitivity: 1.0,
		CrossbarSize:     64,
	}
}

// MZI represents a Mach-Zehnder Interferometer
type MZI struct {
	theta float64 // Internal phase
	phi   float64 // External phase
}

// NewMZI creates a new MZI with given phases
func NewMZI(theta, phi float64) *MZI {
	return &MZI{
		theta: theta,
		phi:   phi,
	}
}

// Transfer computes 2x2 transfer matrix for MZI
func (m *MZI) Transfer() [][]complex128 {
	// MZI transfer: T(θ,φ) = R(φ) × BS × P(θ) × BS
	// Where BS is 50:50 beam splitter, P is phase shift, R is rotation

	cosTheta := math.Cos(m.theta / 2)
	sinTheta := math.Sin(m.theta / 2)
	expPhi := complex(math.Cos(m.phi), math.Sin(m.phi))

	return [][]complex128{
		{complex(cosTheta, 0), complex(0, sinTheta) * expPhi},
		{complex(0, sinTheta) * expPhi, complex(cosTheta, 0)},
	}
}

// Forward passes two inputs through the MZI
func (m *MZI) Forward(in1, in2 complex128) (complex128, complex128) {
	T := m.Transfer()
	out1 := T[0][0]*in1 + T[0][1]*in2
	out2 := T[1][0]*in1 + T[1][1]*in2
	return out1, out2
}

// MZIMesh represents a mesh of MZIs for unitary transformation
type MZIMesh struct {
	config *OpticalConfig
	size   int
	mzis   [][]*MZI // Triangular arrangement
}

// NewMZIMesh creates a new MZI mesh for N×N unitary
func NewMZIMesh(size int, config *OpticalConfig) *MZIMesh {
	if config == nil {
		config = DefaultOpticalConfig()
	}

	mesh := &MZIMesh{
		config: config,
		size:   size,
		mzis:   make([][]*MZI, size-1),
	}

	// Initialize MZIs with random phases
	rng := rand.New(rand.NewSource(42))
	for layer := 0; layer < size-1; layer++ {
		numMZIs := size - layer - 1
		mesh.mzis[layer] = make([]*MZI, numMZIs)
		for i := 0; i < numMZIs; i++ {
			mesh.mzis[layer][i] = NewMZI(
				rng.Float64()*2*math.Pi,
				rng.Float64()*2*math.Pi,
			)
		}
	}

	return mesh
}

// SetFromSVD configures MZIs to implement a unitary matrix via SVD
func (m *MZIMesh) SetFromSVD(U [][]float64) {
	// Simplified: decompose U into MZI phases using Givens rotations
	// In practice, this requires iterative decomposition

	for layer := 0; layer < len(m.mzis); layer++ {
		for i := 0; i < len(m.mzis[layer]); i++ {
			// Extract rotation angles from U
			row := layer
			col := i + layer + 1
			if row < len(U) && col < len(U[row]) {
				angle := math.Atan2(U[row][col], U[row][row])
				m.mzis[layer][i].theta = angle
				m.mzis[layer][i].phi = 0
			}
		}
	}
}

// Forward applies the unitary transformation
func (m *MZIMesh) Forward(input []complex128) []complex128 {
	if len(input) != m.size {
		return nil
	}

	output := make([]complex128, m.size)
	copy(output, input)

	// Apply MZI layers
	for layer := 0; layer < len(m.mzis); layer++ {
		for i := 0; i < len(m.mzis[layer]); i++ {
			idx1 := layer
			idx2 := i + layer + 1
			if idx2 < m.size {
				output[idx1], output[idx2] = m.mzis[layer][i].Forward(output[idx1], output[idx2])
			}
		}
	}

	return output
}

// PhotonicTensorCore implements optical matrix multiplication
type PhotonicTensorCore struct {
	config    *OpticalConfig
	size      int
	meshU     *MZIMesh       // Left unitary
	meshV     *MZIMesh       // Right unitary
	sigmas    []float64      // Singular values (attenuators)
	rng       *rand.Rand

	// Statistics
	totalOps   int64
	energyPJ   float64
}

// NewPhotonicTensorCore creates a new PTC
func NewPhotonicTensorCore(size int, config *OpticalConfig) *PhotonicTensorCore {
	if config == nil {
		config = DefaultOpticalConfig()
	}

	ptc := &PhotonicTensorCore{
		config: config,
		size:   size,
		meshU:  NewMZIMesh(size, config),
		meshV:  NewMZIMesh(size, config),
		sigmas: make([]float64, size),
		rng:    rand.New(rand.NewSource(42)),
	}

	// Initialize singular values
	for i := 0; i < size; i++ {
		ptc.sigmas[i] = 1.0
	}

	return ptc
}

// ProgramWeight programs a weight matrix using SVD decomposition
func (ptc *PhotonicTensorCore) ProgramWeight(W [][]float64) {
	// SVD: W = U × Σ × V^T
	// Simplified SVD (in practice use proper SVD algorithm)

	n := len(W)
	if n == 0 {
		return
	}

	// Extract approximate singular values (diagonal)
	for i := 0; i < n && i < ptc.size; i++ {
		if i < len(W[i]) {
			ptc.sigmas[i] = math.Abs(W[i][i])
		}
	}

	// Configure U and V meshes (simplified)
	ptc.meshU.SetFromSVD(W)
	ptc.meshV.SetFromSVD(W)
}

// Forward performs optical matrix-vector multiplication
func (ptc *PhotonicTensorCore) Forward(input []float64) []float64 {
	// Convert to complex (optical amplitude)
	complexInput := make([]complex128, ptc.size)
	for i := 0; i < len(input) && i < ptc.size; i++ {
		complexInput[i] = complex(input[i], 0)
	}

	// Apply V^T
	afterV := ptc.meshV.Forward(complexInput)

	// Apply Σ (diagonal scaling)
	for i := range afterV {
		afterV[i] = complex(real(afterV[i])*ptc.sigmas[i], imag(afterV[i])*ptc.sigmas[i])
	}

	// Apply U
	result := ptc.meshU.Forward(afterV)

	// Convert back to real (photodetection gives intensity)
	output := make([]float64, ptc.size)
	for i := range result {
		// Photodetection: |E|² but we use amplitude for linear response
		output[i] = real(result[i])
	}

	// Update statistics
	ptc.totalOps += int64(ptc.size * ptc.size)
	ptc.energyPJ += float64(ptc.size) * 0.1 // ~0.1 pJ per operation

	return output
}

// GetEfficiency returns compute efficiency in TOPS/W
func (ptc *PhotonicTensorCore) GetEfficiency() float64 {
	// Based on research: ~5.1 TOPS/W for 64×64 at 25 GHz
	opsPerSecond := float64(ptc.size*ptc.size) * ptc.config.ClockSpeedGHz * 1e9
	powerW := ptc.config.LaserPowerMW / 1000.0 // Laser is main power consumer
	return opsPerSecond / powerW / 1e12 // TOPS/W
}

// WDMChannel represents a wavelength channel
type WDMChannel struct {
	wavelength float64 // nm
	power      float64 // mW
	data       []float64
}

// WDMSystem implements wavelength division multiplexing
type WDMSystem struct {
	config   *OpticalConfig
	channels []*WDMChannel
}

// NewWDMSystem creates a new WDM system
func NewWDMSystem(config *OpticalConfig) *WDMSystem {
	if config == nil {
		config = DefaultOpticalConfig()
	}

	wdm := &WDMSystem{
		config:   config,
		channels: make([]*WDMChannel, config.NumWavelengths),
	}

	// Initialize wavelength channels
	startWL := config.CenterWavelength - float64(config.NumWavelengths/2)*config.ChannelSpacing
	for i := 0; i < config.NumWavelengths; i++ {
		wdm.channels[i] = &WDMChannel{
			wavelength: startWL + float64(i)*config.ChannelSpacing,
			power:      config.LaserPowerMW / float64(config.NumWavelengths),
		}
	}

	return wdm
}

// Modulate modulates data onto wavelength channels
func (wdm *WDMSystem) Modulate(data [][]float64) {
	for i := 0; i < len(data) && i < len(wdm.channels); i++ {
		wdm.channels[i].data = data[i]
	}
}

// GetThroughput returns total throughput in Gbps
func (wdm *WDMSystem) GetThroughput() float64 {
	// Each channel at clock speed
	return float64(wdm.config.NumWavelengths) * wdm.config.ClockSpeedGHz
}

// =============================================================================
// WEIGHT CLUSTERING QUANTIZATION
// =============================================================================

// ClusteringConfig configures weight clustering
type ClusteringConfig struct {
	NumClusters    int     // K for k-means
	MaxIterations  int     // K-means iterations
	Tolerance      float64 // Convergence threshold
	InitMethod     string  // "random", "kmeans++", "uniform"
}

// DefaultClusteringConfig returns typical clustering configuration
func DefaultClusteringConfig() *ClusteringConfig {
	return &ClusteringConfig{
		NumClusters:   16,
		MaxIterations: 100,
		Tolerance:     1e-6,
		InitMethod:    "kmeans++",
	}
}

// WeightClusterer performs k-means weight clustering
type WeightClusterer struct {
	config    *ClusteringConfig
	centroids []float64  // Codebook
	labels    []int      // Cluster assignments
	rng       *rand.Rand
}

// NewWeightClusterer creates a new weight clusterer
func NewWeightClusterer(config *ClusteringConfig) *WeightClusterer {
	if config == nil {
		config = DefaultClusteringConfig()
	}

	return &WeightClusterer{
		config:    config,
		centroids: make([]float64, config.NumClusters),
		rng:       rand.New(rand.NewSource(42)),
	}
}

// Fit performs k-means clustering on weights
func (wc *WeightClusterer) Fit(weights []float64) {
	if len(weights) == 0 {
		return
	}

	// Initialize centroids
	wc.initCentroids(weights)

	// K-means iterations
	wc.labels = make([]int, len(weights))
	prevCentroids := make([]float64, len(wc.centroids))

	for iter := 0; iter < wc.config.MaxIterations; iter++ {
		// Assign labels
		for i, w := range weights {
			wc.labels[i] = wc.findNearestCentroid(w)
		}

		// Update centroids
		copy(prevCentroids, wc.centroids)
		wc.updateCentroids(weights)

		// Check convergence
		maxDiff := 0.0
		for i := range wc.centroids {
			diff := math.Abs(wc.centroids[i] - prevCentroids[i])
			if diff > maxDiff {
				maxDiff = diff
			}
		}
		if maxDiff < wc.config.Tolerance {
			break
		}
	}
}

// initCentroids initializes centroids using specified method
func (wc *WeightClusterer) initCentroids(weights []float64) {
	switch wc.config.InitMethod {
	case "kmeans++":
		wc.initKMeansPlusPlus(weights)
	case "uniform":
		wc.initUniform(weights)
	default:
		wc.initRandom(weights)
	}
}

// initRandom randomly selects initial centroids
func (wc *WeightClusterer) initRandom(weights []float64) {
	indices := wc.rng.Perm(len(weights))
	for i := 0; i < wc.config.NumClusters && i < len(indices); i++ {
		wc.centroids[i] = weights[indices[i]]
	}
}

// initUniform distributes centroids uniformly over weight range
func (wc *WeightClusterer) initUniform(weights []float64) {
	minW, maxW := weights[0], weights[0]
	for _, w := range weights {
		if w < minW {
			minW = w
		}
		if w > maxW {
			maxW = w
		}
	}

	step := (maxW - minW) / float64(wc.config.NumClusters-1)
	for i := 0; i < wc.config.NumClusters; i++ {
		wc.centroids[i] = minW + float64(i)*step
	}
}

// initKMeansPlusPlus uses k-means++ initialization
func (wc *WeightClusterer) initKMeansPlusPlus(weights []float64) {
	// First centroid: random
	wc.centroids[0] = weights[wc.rng.Intn(len(weights))]

	// Subsequent centroids: probability proportional to distance²
	distances := make([]float64, len(weights))

	for k := 1; k < wc.config.NumClusters; k++ {
		// Compute distances to nearest centroid
		totalDist := 0.0
		for i, w := range weights {
			minDist := math.MaxFloat64
			for j := 0; j < k; j++ {
				d := (w - wc.centroids[j]) * (w - wc.centroids[j])
				if d < minDist {
					minDist = d
				}
			}
			distances[i] = minDist
			totalDist += minDist
		}

		// Sample proportional to distance²
		r := wc.rng.Float64() * totalDist
		cumSum := 0.0
		for i, d := range distances {
			cumSum += d
			if cumSum >= r {
				wc.centroids[k] = weights[i]
				break
			}
		}
	}

	// Sort centroids
	sort.Float64s(wc.centroids)
}

// findNearestCentroid finds the nearest centroid index
func (wc *WeightClusterer) findNearestCentroid(weight float64) int {
	minDist := math.MaxFloat64
	minIdx := 0

	for i, c := range wc.centroids {
		d := (weight - c) * (weight - c)
		if d < minDist {
			minDist = d
			minIdx = i
		}
	}

	return minIdx
}

// updateCentroids updates centroids based on current assignments
func (wc *WeightClusterer) updateCentroids(weights []float64) {
	sums := make([]float64, wc.config.NumClusters)
	counts := make([]int, wc.config.NumClusters)

	for i, w := range weights {
		label := wc.labels[i]
		sums[label] += w
		counts[label]++
	}

	for i := range wc.centroids {
		if counts[i] > 0 {
			wc.centroids[i] = sums[i] / float64(counts[i])
		}
	}
}

// Quantize quantizes weights using learned codebook
func (wc *WeightClusterer) Quantize(weights []float64) []float64 {
	quantized := make([]float64, len(weights))
	for i, w := range weights {
		idx := wc.findNearestCentroid(w)
		quantized[i] = wc.centroids[idx]
	}
	return quantized
}

// GetCompressionRatio returns compression ratio
func (wc *WeightClusterer) GetCompressionRatio(originalBits, indexBits int) float64 {
	// Original: originalBits per weight
	// Compressed: indexBits per weight + codebook overhead
	// Codebook: NumClusters × originalBits
	return float64(originalBits) / float64(indexBits)
}

// GetMSE returns mean squared error after quantization
func (wc *WeightClusterer) GetMSE(original, quantized []float64) float64 {
	if len(original) != len(quantized) || len(original) == 0 {
		return 0
	}

	mse := 0.0
	for i := range original {
		diff := original[i] - quantized[i]
		mse += diff * diff
	}
	return mse / float64(len(original))
}

// ColumnWiseQuantizer implements column-wise quantization for CIM
type ColumnWiseQuantizer struct {
	numBits      int
	scaleFactors []float64 // Per-column scale factors
}

// NewColumnWiseQuantizer creates a new column-wise quantizer
func NewColumnWiseQuantizer(numColumns, numBits int) *ColumnWiseQuantizer {
	return &ColumnWiseQuantizer{
		numBits:      numBits,
		scaleFactors: make([]float64, numColumns),
	}
}

// CalibrateColumn calibrates scale factor for a column
func (cw *ColumnWiseQuantizer) CalibrateColumn(col int, weights []float64) {
	if col < 0 || col >= len(cw.scaleFactors) || len(weights) == 0 {
		return
	}

	// Find max absolute value in column
	maxAbs := 0.0
	for _, w := range weights {
		abs := math.Abs(w)
		if abs > maxAbs {
			maxAbs = abs
		}
	}

	// Scale factor: max_value / (2^(bits-1) - 1)
	maxQuant := math.Pow(2, float64(cw.numBits-1)) - 1
	if maxAbs > 0 {
		cw.scaleFactors[col] = maxAbs / maxQuant
	} else {
		cw.scaleFactors[col] = 1.0
	}
}

// QuantizeColumn quantizes a column's weights
func (cw *ColumnWiseQuantizer) QuantizeColumn(col int, weights []float64) []float64 {
	if col < 0 || col >= len(cw.scaleFactors) {
		return weights
	}

	scale := cw.scaleFactors[col]
	maxQuant := math.Pow(2, float64(cw.numBits-1)) - 1
	minQuant := -maxQuant - 1

	quantized := make([]float64, len(weights))
	for i, w := range weights {
		// Quantize
		q := math.Round(w / scale)
		// Clip
		if q > maxQuant {
			q = maxQuant
		} else if q < minQuant {
			q = minQuant
		}
		quantized[i] = q * scale
	}

	return quantized
}

// QuantizePartialSum quantizes partial sums with column-aligned scale
func (cw *ColumnWiseQuantizer) QuantizePartialSum(col int, partialSum float64, psScale float64) float64 {
	if col < 0 || col >= len(cw.scaleFactors) {
		return partialSum
	}

	// Align partial sum quantization with weight quantization
	wScale := cw.scaleFactors[col]
	combinedScale := wScale * psScale

	// Quantize and dequantize
	q := math.Round(partialSum / combinedScale)
	return q * combinedScale
}

// MixedPrecisionConfig configures mixed precision quantization
type MixedPrecisionConfig struct {
	BitOptions    []int   // Available bit widths
	TargetSize    int64   // Target model size (bytes)
	SensitivityThreshold float64 // Layer sensitivity threshold
}

// DefaultMixedPrecisionConfig returns typical MPQ configuration
func DefaultMixedPrecisionConfig() *MixedPrecisionConfig {
	return &MixedPrecisionConfig{
		BitOptions:           []int{2, 4, 6, 8},
		TargetSize:           1024 * 1024, // 1 MB
		SensitivityThreshold: 0.01,
	}
}

// MixedPrecisionQuantizer implements layer-wise mixed precision
type MixedPrecisionQuantizer struct {
	config      *MixedPrecisionConfig
	layerBits   []int     // Bits per layer
	sensitivities []float64 // Layer sensitivities
}

// NewMixedPrecisionQuantizer creates a new MPQ
func NewMixedPrecisionQuantizer(numLayers int, config *MixedPrecisionConfig) *MixedPrecisionQuantizer {
	if config == nil {
		config = DefaultMixedPrecisionConfig()
	}

	mpq := &MixedPrecisionQuantizer{
		config:       config,
		layerBits:    make([]int, numLayers),
		sensitivities: make([]float64, numLayers),
	}

	// Default to max bits
	maxBits := config.BitOptions[len(config.BitOptions)-1]
	for i := range mpq.layerBits {
		mpq.layerBits[i] = maxBits
	}

	return mpq
}

// ComputeSensitivity estimates layer sensitivity to quantization
func (mpq *MixedPrecisionQuantizer) ComputeSensitivity(layer int, weights []float64) float64 {
	if layer < 0 || layer >= len(mpq.sensitivities) || len(weights) == 0 {
		return 0
	}

	// Sensitivity based on weight variance
	mean := 0.0
	for _, w := range weights {
		mean += w
	}
	mean /= float64(len(weights))

	variance := 0.0
	for _, w := range weights {
		diff := w - mean
		variance += diff * diff
	}
	variance /= float64(len(weights))

	mpq.sensitivities[layer] = math.Sqrt(variance)
	return mpq.sensitivities[layer]
}

// AssignBits assigns bit widths based on sensitivities
func (mpq *MixedPrecisionQuantizer) AssignBits() {
	// Sort layers by sensitivity
	type layerSens struct {
		idx  int
		sens float64
	}

	layers := make([]layerSens, len(mpq.sensitivities))
	for i, s := range mpq.sensitivities {
		layers[i] = layerSens{i, s}
	}

	sort.Slice(layers, func(i, j int) bool {
		return layers[i].sens > layers[j].sens
	})

	// Assign higher bits to more sensitive layers
	for rank, ls := range layers {
		// Map rank to bit width
		bitIdx := rank * len(mpq.config.BitOptions) / len(layers)
		if bitIdx >= len(mpq.config.BitOptions) {
			bitIdx = len(mpq.config.BitOptions) - 1
		}
		mpq.layerBits[ls.idx] = mpq.config.BitOptions[len(mpq.config.BitOptions)-1-bitIdx]
	}
}

// GetModelSize estimates total model size
func (mpq *MixedPrecisionQuantizer) GetModelSize(layerSizes []int64) int64 {
	var totalBits int64
	for i, size := range layerSizes {
		if i < len(mpq.layerBits) {
			totalBits += size * int64(mpq.layerBits[i])
		}
	}
	return totalBits / 8 // Convert to bytes
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// CreateTestWeightMatrix creates a test weight matrix
func CreateTestWeightMatrix(rows, cols int, seed int64) [][]float64 {
	rng := rand.New(rand.NewSource(seed))
	W := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		W[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			W[i][j] = rng.NormFloat64()
		}
	}
	return W
}

// FlattenMatrix converts 2D matrix to 1D array
func FlattenMatrix(W [][]float64) []float64 {
	if len(W) == 0 {
		return nil
	}

	flat := make([]float64, 0, len(W)*len(W[0]))
	for _, row := range W {
		flat = append(flat, row...)
	}
	return flat
}

// FormatOpticalReport generates optical system report
func FormatOpticalReport(ptc *PhotonicTensorCore) string {
	report := "=== Photonic Tensor Core Report ===\n\n"

	report += fmt.Sprintf("Configuration:\n")
	report += fmt.Sprintf("  Size: %d × %d\n", ptc.size, ptc.size)
	report += fmt.Sprintf("  Clock speed: %.1f GHz\n", ptc.config.ClockSpeedGHz)
	report += fmt.Sprintf("  Wavelengths: %d\n", ptc.config.NumWavelengths)
	report += fmt.Sprintf("  MZI stages: %d\n", ptc.config.NumMZI)

	report += fmt.Sprintf("\nPerformance:\n")
	report += fmt.Sprintf("  Efficiency: %.2f TOPS/W\n", ptc.GetEfficiency())
	report += fmt.Sprintf("  Total operations: %d\n", ptc.totalOps)
	report += fmt.Sprintf("  Est. energy: %.2e pJ\n", ptc.energyPJ)

	return report
}

// FormatClusteringReport generates clustering report
func FormatClusteringReport(wc *WeightClusterer, original, quantized []float64) string {
	report := "=== Weight Clustering Report ===\n\n"

	report += fmt.Sprintf("Configuration:\n")
	report += fmt.Sprintf("  Clusters: %d\n", wc.config.NumClusters)
	report += fmt.Sprintf("  Init method: %s\n", wc.config.InitMethod)

	report += fmt.Sprintf("\nCodebook:\n")
	for i, c := range wc.centroids {
		report += fmt.Sprintf("  [%d] %.4f\n", i, c)
	}

	if len(original) > 0 && len(quantized) > 0 {
		mse := wc.GetMSE(original, quantized)
		report += fmt.Sprintf("\nQuantization Error:\n")
		report += fmt.Sprintf("  MSE: %.6f\n", mse)
		report += fmt.Sprintf("  RMSE: %.6f\n", math.Sqrt(mse))
	}

	return report
}
