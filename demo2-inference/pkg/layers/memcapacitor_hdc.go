// Package layers provides ferroelectric memcapacitor and hyperdimensional computing
// (HDC) simulation for ultra-low power CIM applications.
//
// Based on research:
// - HZO memcapacitor arrays: 29,600 TOPS/W, 31 fJ/inference
// - Capacitive crossbar: 20-200× lower energy than resistive
// - FeFET-based HDC encoders: 826× energy improvement
// - Vector symbolic architectures: 10,000+ dimensional hypervectors
//
// References:
// - Nature Electronics 2021, "Energy-efficient memcapacitor devices"
// - Scientific Reports 2022, "Software-equivalent accuracy for HDC"
// - Nano Energy 2025, "HZO memcapacitor array for neuromorphic computing"
// - IEEE JSSC 2024, "Low-power charge-domain CIM with memcapacitor"
package layers

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
)

// ============================================================================
// FERROELECTRIC MEMCAPACITOR DEVICE MODELING
// ============================================================================

// MemcapacitorConfig configures ferroelectric memcapacitor device
type MemcapacitorConfig struct {
	// Material parameters
	FerroThickness   float64 // nm (typically 10nm for HZO)
	DielectricConst  float64 // Relative permittivity
	RemanentPol      float64 // μC/cm² (Pr)
	CoerciveVoltage  float64 // V (Vc)

	// Device parameters
	Area             float64 // μm²
	InterfaceLayer   float64 // nm (IL thickness for low-voltage operation)

	// Performance targets
	MemoryWindow     float64 // fF/μm² (capacitance modulation)
	ONOFFRatio       float64 // Capacitance ON/OFF ratio
	Endurance        int64   // Cycles
	RetentionYears   float64 // Years at elevated temperature
}

// DefaultMemcapacitorConfig returns typical HZO memcapacitor configuration
func DefaultMemcapacitorConfig() *MemcapacitorConfig {
	return &MemcapacitorConfig{
		FerroThickness:  10.0,
		DielectricConst: 25.0,
		RemanentPol:     25.0,
		CoerciveVoltage: 2.0,
		Area:            1.0,
		InterfaceLayer:  0.0, // No IL for standard device
		MemoryWindow:    7.8,
		ONOFFRatio:      1000.0,
		Endurance:       1e9,
		RetentionYears:  10.0,
	}
}

// LowVoltageMemcapacitorConfig returns 1.2V operation configuration
func LowVoltageMemcapacitorConfig() *MemcapacitorConfig {
	return &MemcapacitorConfig{
		FerroThickness:  5.0,  // Thinner HZO
		DielectricConst: 25.0,
		RemanentPol:     20.0,
		CoerciveVoltage: 1.2,  // Low voltage operation
		Area:            1.0,
		InterfaceLayer:  2.0,  // IL engineering
		MemoryWindow:    5.0,
		ONOFFRatio:      500.0,
		Endurance:       1e8,
		RetentionYears:  10.0,
	}
}

// MemcapacitorState represents the polarization state
type MemcapacitorState struct {
	Polarization float64 // Normalized -1 to +1
	Capacitance  float64 // fF
	Charge       float64 // fC
}

// Memcapacitor represents a single ferroelectric memcapacitor device
type Memcapacitor struct {
	Config *MemcapacitorConfig
	State  MemcapacitorState

	// Calculated parameters
	BaseCapacitance float64 // fF (unpolarized)
	MaxCapacitance  float64 // fF (fully polarized)
	MinCapacitance  float64 // fF (opposite polarization)

	// Variation parameters
	DeviceVariation float64 // % D2D variation
	CycleVariation  float64 // % C2C variation
}

// NewMemcapacitor creates a new ferroelectric memcapacitor
func NewMemcapacitor(config *MemcapacitorConfig) *Memcapacitor {
	if config == nil {
		config = DefaultMemcapacitorConfig()
	}

	// Calculate capacitance from geometry
	// C = ε₀ × εr × A / d
	epsilon0 := 8.854e-3 // fF/μm
	totalThickness := config.FerroThickness + config.InterfaceLayer
	baseC := epsilon0 * config.DielectricConst * config.Area / (totalThickness * 1e-3)

	mc := &Memcapacitor{
		Config:          config,
		BaseCapacitance: baseC,
		MaxCapacitance:  baseC * (1 + config.MemoryWindow/baseC),
		MinCapacitance:  baseC / config.ONOFFRatio,
		DeviceVariation: 3.0, // 3% typical
		CycleVariation:  1.0, // 1% typical
	}

	// Initialize to mid-state
	mc.State = MemcapacitorState{
		Polarization: 0,
		Capacitance:  baseC,
		Charge:       0,
	}

	return mc
}

// Program sets the memcapacitor to a target polarization state
func (mc *Memcapacitor) Program(targetPol float64) {
	// Clamp to valid range
	if targetPol > 1.0 {
		targetPol = 1.0
	}
	if targetPol < -1.0 {
		targetPol = -1.0
	}

	mc.State.Polarization = targetPol

	// Calculate capacitance based on polarization
	// Capacitance varies with ferroelectric domain configuration
	polFraction := (targetPol + 1.0) / 2.0 // 0 to 1
	mc.State.Capacitance = mc.MinCapacitance +
		(mc.MaxCapacitance-mc.MinCapacitance)*polFraction

	// Add variation
	variation := 1.0 + (rand.Float64()-0.5)*mc.CycleVariation/100.0
	mc.State.Capacitance *= variation
}

// Read returns the stored charge for a given read voltage
func (mc *Memcapacitor) Read(readVoltage float64) float64 {
	// Q = C × V (non-destructive read)
	mc.State.Charge = mc.State.Capacitance * readVoltage
	return mc.State.Charge
}

// GetWeight returns normalized weight for CIM operation
func (mc *Memcapacitor) GetWeight() float64 {
	// Map capacitance to weight [-1, 1]
	normalized := (mc.State.Capacitance - mc.MinCapacitance) /
		(mc.MaxCapacitance - mc.MinCapacitance)
	return 2.0*normalized - 1.0
}

// ============================================================================
// MEMCAPACITOR CROSSBAR ARRAY
// ============================================================================

// MemcapacitorCrossbar implements capacitive crossbar array
type MemcapacitorCrossbar struct {
	Rows         int
	Cols         int
	Cells        [][]*Memcapacitor
	Config       *MemcapacitorConfig

	// Performance metrics
	EnergyPerMVM float64 // pJ per vector-matrix multiplication
	Throughput   float64 // GOPS
}

// NewMemcapacitorCrossbar creates a capacitive crossbar array
func NewMemcapacitorCrossbar(rows, cols int, config *MemcapacitorConfig) *MemcapacitorCrossbar {
	if config == nil {
		config = DefaultMemcapacitorConfig()
	}

	cb := &MemcapacitorCrossbar{
		Rows:   rows,
		Cols:   cols,
		Cells:  make([][]*Memcapacitor, rows),
		Config: config,
	}

	// Initialize cells with device-to-device variation
	for i := 0; i < rows; i++ {
		cb.Cells[i] = make([]*Memcapacitor, cols)
		for j := 0; j < cols; j++ {
			cb.Cells[i][j] = NewMemcapacitor(config)
			// Add D2D variation
			cb.Cells[i][j].MaxCapacitance *= (1 + (rand.Float64()-0.5)*cb.Cells[i][j].DeviceVariation/100.0)
			cb.Cells[i][j].MinCapacitance *= (1 + (rand.Float64()-0.5)*cb.Cells[i][j].DeviceVariation/100.0)
		}
	}

	// Calculate energy per MVM (capacitive, no static power)
	// E = 0.5 × C × V² × N
	avgC := (config.MemoryWindow + config.Area) * 1e-15 // F
	cb.EnergyPerMVM = 0.5 * avgC * config.CoerciveVoltage * config.CoerciveVoltage *
		float64(rows*cols) * 1e12 // pJ

	return cb
}

// ProgramWeights programs the crossbar with weight matrix
func (cb *MemcapacitorCrossbar) ProgramWeights(weights [][]float64) error {
	if len(weights) != cb.Rows {
		return fmt.Errorf("weight matrix row count mismatch: got %d, expected %d", len(weights), cb.Rows)
	}

	for i, row := range weights {
		if len(row) != cb.Cols {
			return fmt.Errorf("weight matrix column count mismatch at row %d", i)
		}
		for j, w := range row {
			// Map weight [-1, 1] to polarization
			cb.Cells[i][j].Program(w)
		}
	}

	return nil
}

// MatrixVectorMultiply performs capacitive MVM
func (cb *MemcapacitorCrossbar) MatrixVectorMultiply(input []float64) ([]float64, error) {
	if len(input) != cb.Cols {
		return nil, fmt.Errorf("input size mismatch: got %d, expected %d", len(input), cb.Cols)
	}

	output := make([]float64, cb.Rows)

	for i := 0; i < cb.Rows; i++ {
		var chargeSum float64
		for j := 0; j < cb.Cols; j++ {
			// Q_ij = C_ij × V_j
			charge := cb.Cells[i][j].Read(input[j])
			chargeSum += charge
		}
		// Output is total charge on row line
		output[i] = chargeSum
	}

	return output, nil
}

// GetEnergyEfficiency returns TOPS/W efficiency
func (cb *MemcapacitorCrossbar) GetEnergyEfficiency() float64 {
	// Operations per MVM = 2 × rows × cols (multiply-accumulate)
	ops := 2.0 * float64(cb.Rows*cb.Cols)

	// Energy in Joules
	energyJ := cb.EnergyPerMVM * 1e-12

	// TOPS/W = (ops / 1e12) / (energy / 1)
	return ops / (energyJ * 1e12)
}

// GetStatistics returns crossbar statistics
func (cb *MemcapacitorCrossbar) GetStatistics() map[string]interface{} {
	var totalC, minC, maxC float64
	maxC = 0
	minC = math.MaxFloat64

	for i := 0; i < cb.Rows; i++ {
		for j := 0; j < cb.Cols; j++ {
			c := cb.Cells[i][j].State.Capacitance
			totalC += c
			if c < minC {
				minC = c
			}
			if c > maxC {
				maxC = c
			}
		}
	}

	numCells := float64(cb.Rows * cb.Cols)

	return map[string]interface{}{
		"rows":              cb.Rows,
		"cols":              cb.Cols,
		"avg_capacitance_fF": totalC / numCells,
		"min_capacitance_fF": minC,
		"max_capacitance_fF": maxC,
		"energy_per_mvm_pJ": cb.EnergyPerMVM,
		"efficiency_TOPS_W": cb.GetEnergyEfficiency(),
		"static_power_W":    0.0, // Zero static power!
	}
}

// ============================================================================
// HYPERDIMENSIONAL COMPUTING (HDC) FUNDAMENTALS
// ============================================================================

// HDCModel defines the hyperdimensional computing model type
type HDCModel int

const (
	BinarySpatterCode HDCModel = iota // BSC: binary {0, 1}
	HolographicReduced               // HRR: complex unit phasors
	MultiplyAddPermute               // MAP: integer vectors
	SparseDistributed                // SBDR: sparse binary
)

// HypervectorConfig configures hypervector parameters
type HypervectorConfig struct {
	Dimensions   int      // D (typically 10,000)
	Model        HDCModel // HDC model type
	Sparsity     float64  // For sparse models (0-1)
	Quantization int      // Bits for quantized vectors
}

// DefaultHypervectorConfig returns typical HDC configuration
func DefaultHypervectorConfig() *HypervectorConfig {
	return &HypervectorConfig{
		Dimensions:   10000,
		Model:        BinarySpatterCode,
		Sparsity:     0.5,
		Quantization: 1, // Binary
	}
}

// Hypervector represents a high-dimensional distributed representation
type Hypervector struct {
	Config   *HypervectorConfig
	Elements []float64 // Vector elements
}

// NewHypervector creates a new random hypervector
func NewHypervector(config *HypervectorConfig) *Hypervector {
	if config == nil {
		config = DefaultHypervectorConfig()
	}

	hv := &Hypervector{
		Config:   config,
		Elements: make([]float64, config.Dimensions),
	}

	// Initialize based on model type
	switch config.Model {
	case BinarySpatterCode:
		// Random binary {-1, +1}
		for i := range hv.Elements {
			if rand.Float64() < 0.5 {
				hv.Elements[i] = -1.0
			} else {
				hv.Elements[i] = 1.0
			}
		}

	case HolographicReduced:
		// Random unit phasors (real part)
		for i := range hv.Elements {
			angle := rand.Float64() * 2 * math.Pi
			hv.Elements[i] = math.Cos(angle)
		}

	case MultiplyAddPermute:
		// Random integers in range
		maxVal := float64(1 << config.Quantization)
		for i := range hv.Elements {
			hv.Elements[i] = math.Floor(rand.Float64() * maxVal)
		}

	case SparseDistributed:
		// Sparse binary
		for i := range hv.Elements {
			if rand.Float64() < config.Sparsity {
				hv.Elements[i] = 1.0
			} else {
				hv.Elements[i] = 0.0
			}
		}
	}

	return hv
}

// NewZeroHypervector creates a zero-initialized hypervector
func NewZeroHypervector(config *HypervectorConfig) *Hypervector {
	if config == nil {
		config = DefaultHypervectorConfig()
	}

	return &Hypervector{
		Config:   config,
		Elements: make([]float64, config.Dimensions),
	}
}

// Clone creates a copy of the hypervector
func (hv *Hypervector) Clone() *Hypervector {
	clone := &Hypervector{
		Config:   hv.Config,
		Elements: make([]float64, len(hv.Elements)),
	}
	copy(clone.Elements, hv.Elements)
	return clone
}

// ============================================================================
// HDC OPERATIONS
// ============================================================================

// Bind performs element-wise binding (multiplication for BSC)
// Used to associate two concepts (e.g., role-filler binding)
func Bind(a, b *Hypervector) *Hypervector {
	if a.Config.Dimensions != b.Config.Dimensions {
		return nil
	}

	result := NewZeroHypervector(a.Config)

	switch a.Config.Model {
	case BinarySpatterCode:
		// XOR equivalent: multiply bipolar
		for i := range result.Elements {
			result.Elements[i] = a.Elements[i] * b.Elements[i]
		}

	case HolographicReduced:
		// Circular convolution (simplified as element-wise for speed)
		for i := range result.Elements {
			result.Elements[i] = a.Elements[i] * b.Elements[i]
		}

	default:
		// Element-wise multiplication
		for i := range result.Elements {
			result.Elements[i] = a.Elements[i] * b.Elements[i]
		}
	}

	return result
}

// Bundle performs element-wise bundling (addition)
// Used to create sets/bags of concepts
func Bundle(vectors []*Hypervector) *Hypervector {
	if len(vectors) == 0 {
		return nil
	}

	result := NewZeroHypervector(vectors[0].Config)

	for _, v := range vectors {
		for i := range result.Elements {
			result.Elements[i] += v.Elements[i]
		}
	}

	return result
}

// BundleWithThreshold bundles and applies threshold (for binary models)
func BundleWithThreshold(vectors []*Hypervector) *Hypervector {
	bundled := Bundle(vectors)
	if bundled == nil {
		return nil
	}

	threshold := float64(len(vectors)) / 2.0

	for i := range bundled.Elements {
		if bundled.Elements[i] >= threshold {
			bundled.Elements[i] = 1.0
		} else {
			bundled.Elements[i] = -1.0
		}
	}

	return bundled
}

// Permute performs cyclic permutation (rotation)
// Used for encoding sequence/position information
func Permute(hv *Hypervector, positions int) *Hypervector {
	result := NewZeroHypervector(hv.Config)
	d := hv.Config.Dimensions

	for i := range result.Elements {
		srcIdx := (i - positions + d) % d
		result.Elements[i] = hv.Elements[srcIdx]
	}

	return result
}

// Similarity computes cosine similarity between hypervectors
func Similarity(a, b *Hypervector) float64 {
	if a.Config.Dimensions != b.Config.Dimensions {
		return 0
	}

	var dotProduct, normA, normB float64

	for i := range a.Elements {
		dotProduct += a.Elements[i] * b.Elements[i]
		normA += a.Elements[i] * a.Elements[i]
		normB += b.Elements[i] * b.Elements[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// HammingDistance computes Hamming distance for binary hypervectors
func HammingDistance(a, b *Hypervector) int {
	if a.Config.Dimensions != b.Config.Dimensions {
		return -1
	}

	distance := 0
	for i := range a.Elements {
		if a.Elements[i] != b.Elements[i] {
			distance++
		}
	}

	return distance
}

// ============================================================================
// HDC ENCODER
// ============================================================================

// HDCEncoder encodes data into hypervectors
type HDCEncoder struct {
	Config       *HypervectorConfig
	ItemMemory   map[string]*Hypervector // Base hypervectors for items
	PositionVecs []*Hypervector          // Position encoding vectors
}

// NewHDCEncoder creates a new HDC encoder
func NewHDCEncoder(config *HypervectorConfig, numPositions int) *HDCEncoder {
	if config == nil {
		config = DefaultHypervectorConfig()
	}

	enc := &HDCEncoder{
		Config:       config,
		ItemMemory:   make(map[string]*Hypervector),
		PositionVecs: make([]*Hypervector, numPositions),
	}

	// Generate position vectors using permutation
	basePos := NewHypervector(config)
	for i := 0; i < numPositions; i++ {
		enc.PositionVecs[i] = Permute(basePos, i)
	}

	return enc
}

// GetOrCreateItem gets or creates an item hypervector
func (enc *HDCEncoder) GetOrCreateItem(item string) *Hypervector {
	if hv, exists := enc.ItemMemory[item]; exists {
		return hv
	}

	hv := NewHypervector(enc.Config)
	enc.ItemMemory[item] = hv
	return hv
}

// EncodeNGram encodes an n-gram sequence into a hypervector
func (enc *HDCEncoder) EncodeNGram(items []string) *Hypervector {
	if len(items) == 0 {
		return NewZeroHypervector(enc.Config)
	}

	// Bind items with position vectors
	var boundVecs []*Hypervector
	for i, item := range items {
		itemHV := enc.GetOrCreateItem(item)
		if i < len(enc.PositionVecs) {
			bound := Bind(itemHV, enc.PositionVecs[i])
			boundVecs = append(boundVecs, bound)
		} else {
			boundVecs = append(boundVecs, itemHV)
		}
	}

	// Bundle all bound vectors
	return BundleWithThreshold(boundVecs)
}

// EncodeSequence encodes a full sequence using sliding n-gram window
func (enc *HDCEncoder) EncodeSequence(items []string, ngramSize int) *Hypervector {
	if len(items) < ngramSize {
		return enc.EncodeNGram(items)
	}

	var ngrams []*Hypervector

	for i := 0; i <= len(items)-ngramSize; i++ {
		ngram := enc.EncodeNGram(items[i : i+ngramSize])
		ngrams = append(ngrams, ngram)
	}

	return BundleWithThreshold(ngrams)
}

// ============================================================================
// HDC CLASSIFIER
// ============================================================================

// HDCClassifier implements associative memory classification
type HDCClassifier struct {
	Config        *HypervectorConfig
	ClassVectors  map[string]*Hypervector // Class prototype vectors
	TrainingCount map[string]int          // Number of training samples per class
}

// NewHDCClassifier creates a new HDC classifier
func NewHDCClassifier(config *HypervectorConfig) *HDCClassifier {
	if config == nil {
		config = DefaultHypervectorConfig()
	}

	return &HDCClassifier{
		Config:        config,
		ClassVectors:  make(map[string]*Hypervector),
		TrainingCount: make(map[string]int),
	}
}

// Train adds a training sample to a class
func (clf *HDCClassifier) Train(classLabel string, sample *Hypervector) {
	if _, exists := clf.ClassVectors[classLabel]; !exists {
		clf.ClassVectors[classLabel] = NewZeroHypervector(clf.Config)
		clf.TrainingCount[classLabel] = 0
	}

	// Accumulate into class vector
	for i := range clf.ClassVectors[classLabel].Elements {
		clf.ClassVectors[classLabel].Elements[i] += sample.Elements[i]
	}
	clf.TrainingCount[classLabel]++
}

// Retrain applies majority threshold to all class vectors
func (clf *HDCClassifier) Retrain() {
	for label, classVec := range clf.ClassVectors {
		count := clf.TrainingCount[label]
		threshold := float64(count) / 2.0

		for i := range classVec.Elements {
			if classVec.Elements[i] >= threshold {
				classVec.Elements[i] = 1.0
			} else {
				classVec.Elements[i] = -1.0
			}
		}
	}
}

// Predict classifies a query hypervector
func (clf *HDCClassifier) Predict(query *Hypervector) (string, float64) {
	bestLabel := ""
	bestSimilarity := -2.0

	for label, classVec := range clf.ClassVectors {
		sim := Similarity(query, classVec)
		if sim > bestSimilarity {
			bestSimilarity = sim
			bestLabel = label
		}
	}

	return bestLabel, bestSimilarity
}

// GetAccuracy evaluates classifier on test set
func (clf *HDCClassifier) GetAccuracy(testSamples []*Hypervector, testLabels []string) float64 {
	if len(testSamples) != len(testLabels) {
		return 0
	}

	correct := 0
	for i, sample := range testSamples {
		predicted, _ := clf.Predict(sample)
		if predicted == testLabels[i] {
			correct++
		}
	}

	return float64(correct) / float64(len(testSamples))
}

// ============================================================================
// FEFET-BASED HDC IN-MEMORY COMPUTING
// ============================================================================

// FeFETHDCConfig configures FeFET-based HDC accelerator
type FeFETHDCConfig struct {
	Dimensions     int     // Hypervector dimensions
	NumClasses     int     // Number of classes to store
	Precision      int     // Bits per element
	ArraySize      int     // Crossbar array size
	EnergyPerOp    float64 // fJ per operation
}

// DefaultFeFETHDCConfig returns typical FeFET HDC configuration
func DefaultFeFETHDCConfig() *FeFETHDCConfig {
	return &FeFETHDCConfig{
		Dimensions:  10000,
		NumClasses:  26,    // e.g., 26 letters
		Precision:   4,     // 4-bit
		ArraySize:   128,
		EnergyPerOp: 0.5,   // fJ
	}
}

// FeFETHDCAccelerator implements FeFET-based HDC hardware
type FeFETHDCAccelerator struct {
	Config          *FeFETHDCConfig
	EncoderArray    *MemcapacitorCrossbar // For encoding operations
	ClassifierArray *MemcapacitorCrossbar // For associative search (CAM)
	ClassVectors    [][]*Memcapacitor     // Stored class prototypes

	// Performance metrics
	EncodingEnergy   float64 // fJ per encoding
	SearchEnergy     float64 // fJ per search
	Latency          float64 // ns per inference
	Throughput       float64 // Inferences/sec
}

// NewFeFETHDCAccelerator creates a FeFET-based HDC accelerator
func NewFeFETHDCAccelerator(config *FeFETHDCConfig) *FeFETHDCAccelerator {
	if config == nil {
		config = DefaultFeFETHDCConfig()
	}

	// Calculate number of arrays needed
	arraysForDim := (config.Dimensions + config.ArraySize - 1) / config.ArraySize

	acc := &FeFETHDCAccelerator{
		Config:       config,
		EncoderArray: NewMemcapacitorCrossbar(config.ArraySize, config.ArraySize, nil),
		ClassifierArray: NewMemcapacitorCrossbar(config.NumClasses, config.Dimensions,
			LowVoltageMemcapacitorConfig()),
		ClassVectors: make([][]*Memcapacitor, config.NumClasses),
	}

	// Initialize class vector storage
	for i := 0; i < config.NumClasses; i++ {
		acc.ClassVectors[i] = make([]*Memcapacitor, config.Dimensions)
		for j := 0; j < config.Dimensions; j++ {
			acc.ClassVectors[i][j] = NewMemcapacitor(LowVoltageMemcapacitorConfig())
		}
	}

	// Calculate energy metrics
	acc.EncodingEnergy = float64(config.Dimensions) * config.EnergyPerOp
	acc.SearchEnergy = float64(config.NumClasses*config.Dimensions) * config.EnergyPerOp * 0.5 // CAM is more efficient
	acc.Latency = float64(arraysForDim) * 10.0 // 10 ns per array operation
	acc.Throughput = 1e9 / acc.Latency

	return acc
}

// StoreClassVector stores a class prototype in the CAM array
func (acc *FeFETHDCAccelerator) StoreClassVector(classIdx int, hv *Hypervector) error {
	if classIdx >= acc.Config.NumClasses {
		return fmt.Errorf("class index out of range")
	}

	if len(hv.Elements) != acc.Config.Dimensions {
		return fmt.Errorf("hypervector dimension mismatch")
	}

	for j, val := range hv.Elements {
		// Map hypervector element to memcapacitor state
		acc.ClassVectors[classIdx][j].Program(val)
	}

	return nil
}

// AssociativeSearch performs CAM-based associative search
func (acc *FeFETHDCAccelerator) AssociativeSearch(query *Hypervector) (int, float64) {
	if len(query.Elements) != acc.Config.Dimensions {
		return -1, 0
	}

	bestClass := -1
	bestSimilarity := -2.0

	for i := 0; i < acc.Config.NumClasses; i++ {
		// Compute similarity using charge-based comparison
		var dotProduct float64
		for j := 0; j < acc.Config.Dimensions; j++ {
			// Read stored value
			stored := acc.ClassVectors[i][j].GetWeight()
			dotProduct += stored * query.Elements[j]
		}

		// Normalize
		similarity := dotProduct / float64(acc.Config.Dimensions)

		if similarity > bestSimilarity {
			bestSimilarity = similarity
			bestClass = i
		}
	}

	return bestClass, bestSimilarity
}

// GetPerformanceMetrics returns accelerator performance
func (acc *FeFETHDCAccelerator) GetPerformanceMetrics() map[string]interface{} {
	// Calculate energy improvement over GPU baseline
	gpuEnergyPerInference := 500e6 // fJ (0.5 mJ)
	totalEnergy := acc.EncodingEnergy + acc.SearchEnergy
	energyImprovement := gpuEnergyPerInference / totalEnergy

	// Calculate latency improvement
	gpuLatency := 200e3 // ns (200 μs)
	latencyImprovement := gpuLatency / acc.Latency

	return map[string]interface{}{
		"dimensions":              acc.Config.Dimensions,
		"num_classes":             acc.Config.NumClasses,
		"encoding_energy_fJ":      acc.EncodingEnergy,
		"search_energy_fJ":        acc.SearchEnergy,
		"total_energy_fJ":         totalEnergy,
		"latency_ns":              acc.Latency,
		"throughput_inf_per_sec":  acc.Throughput,
		"energy_improvement_vs_gpu": energyImprovement,
		"latency_improvement_vs_gpu": latencyImprovement,
	}
}

// ============================================================================
// CONTENT-ADDRESSABLE MEMORY (CAM) FOR HDC
// ============================================================================

// TCAMCell represents a ternary CAM cell (0, 1, X)
type TCAMCell struct {
	Value    int // 0, 1, or 2 (don't care)
	Memcap   *Memcapacitor
}

// TernaryCAM implements ternary content-addressable memory
type TernaryCAM struct {
	Rows     int // Number of stored patterns
	Cols     int // Pattern width (dimensions)
	Cells    [][]*TCAMCell
	Config   *MemcapacitorConfig
}

// NewTernaryCAM creates a ternary CAM array
func NewTernaryCAM(rows, cols int, config *MemcapacitorConfig) *TernaryCAM {
	if config == nil {
		config = LowVoltageMemcapacitorConfig()
	}

	tcam := &TernaryCAM{
		Rows:   rows,
		Cols:   cols,
		Cells:  make([][]*TCAMCell, rows),
		Config: config,
	}

	for i := 0; i < rows; i++ {
		tcam.Cells[i] = make([]*TCAMCell, cols)
		for j := 0; j < cols; j++ {
			tcam.Cells[i][j] = &TCAMCell{
				Value:  2, // Initialize as don't care
				Memcap: NewMemcapacitor(config),
			}
		}
	}

	return tcam
}

// StorePattern stores a pattern in a TCAM row
func (tcam *TernaryCAM) StorePattern(row int, pattern []int) error {
	if row >= tcam.Rows {
		return fmt.Errorf("row index out of range")
	}

	if len(pattern) > tcam.Cols {
		return fmt.Errorf("pattern too long")
	}

	for j, val := range pattern {
		tcam.Cells[row][j].Value = val
		// Map to memcapacitor: 0 → -1, 1 → +1, X → 0
		switch val {
		case 0:
			tcam.Cells[row][j].Memcap.Program(-1.0)
		case 1:
			tcam.Cells[row][j].Memcap.Program(1.0)
		default: // Don't care
			tcam.Cells[row][j].Memcap.Program(0.0)
		}
	}

	return nil
}

// Search performs parallel pattern matching
func (tcam *TernaryCAM) Search(query []int) []int {
	var matches []int

	for i := 0; i < tcam.Rows; i++ {
		match := true
		for j := 0; j < tcam.Cols && j < len(query); j++ {
			stored := tcam.Cells[i][j].Value
			if stored != 2 && stored != query[j] { // Not don't care and not matching
				match = false
				break
			}
		}
		if match {
			matches = append(matches, i)
		}
	}

	return matches
}

// GetSearchEnergy returns energy for one search operation
func (tcam *TernaryCAM) GetSearchEnergy() float64 {
	// Ternary search: 26.5 aJ per bit (from cryogenic TCAM research)
	// At room temperature, approximately 100× higher
	energyPerBit := 2.65e-3 // fJ (26.5 aJ × 100)
	return energyPerBit * float64(tcam.Cols)
}

// ============================================================================
// IRONLATTICE MEMCAPACITOR-HDC INTEGRATION
// ============================================================================

// IronLatticeHDCConfig configures IronLattice HDC system
type IronLatticeHDCConfig struct {
	// Memcapacitor parameters
	MemcapConfig    *MemcapacitorConfig

	// HDC parameters
	Dimensions      int
	NumClasses      int
	NGramSize       int

	// System parameters
	ArraySize       int
	NumArrays       int
	TargetAccuracy  float64
}

// DefaultIronLatticeHDCConfig returns IronLattice-optimized configuration
func DefaultIronLatticeHDCConfig() *IronLatticeHDCConfig {
	return &IronLatticeHDCConfig{
		MemcapConfig:   LowVoltageMemcapacitorConfig(),
		Dimensions:     10000,
		NumClasses:     26,
		NGramSize:      3,
		ArraySize:      128,
		NumArrays:      80, // 10000/128 ≈ 79
		TargetAccuracy: 0.95,
	}
}

// IronLatticeHDCSystem implements complete IronLattice HDC system
type IronLatticeHDCSystem struct {
	Config       *IronLatticeHDCConfig
	Encoder      *HDCEncoder
	Classifier   *HDCClassifier
	Accelerator  *FeFETHDCAccelerator
	CrossbarBank []*MemcapacitorCrossbar

	// Performance tracking
	InferenceCount   int64
	TotalEnergy      float64 // fJ
	AccuracyHistory  []float64
}

// NewIronLatticeHDCSystem creates an IronLattice HDC system
func NewIronLatticeHDCSystem(config *IronLatticeHDCConfig) *IronLatticeHDCSystem {
	if config == nil {
		config = DefaultIronLatticeHDCConfig()
	}

	hvConfig := &HypervectorConfig{
		Dimensions:   config.Dimensions,
		Model:        BinarySpatterCode,
		Sparsity:     0.5,
		Quantization: 1,
	}

	fefetConfig := &FeFETHDCConfig{
		Dimensions:  config.Dimensions,
		NumClasses:  config.NumClasses,
		Precision:   4,
		ArraySize:   config.ArraySize,
		EnergyPerOp: 0.5,
	}

	sys := &IronLatticeHDCSystem{
		Config:      config,
		Encoder:    NewHDCEncoder(hvConfig, config.NGramSize),
		Classifier: NewHDCClassifier(hvConfig),
		Accelerator: NewFeFETHDCAccelerator(fefetConfig),
		CrossbarBank: make([]*MemcapacitorCrossbar, config.NumArrays),
	}

	// Initialize crossbar bank
	for i := 0; i < config.NumArrays; i++ {
		sys.CrossbarBank[i] = NewMemcapacitorCrossbar(
			config.ArraySize, config.ArraySize, config.MemcapConfig)
	}

	return sys
}

// TrainOnSequence trains the system on labeled sequences
func (sys *IronLatticeHDCSystem) TrainOnSequence(sequences [][]string, labels []string) {
	for i, seq := range sequences {
		encoded := sys.Encoder.EncodeSequence(seq, sys.Config.NGramSize)
		sys.Classifier.Train(labels[i], encoded)
	}

	sys.Classifier.Retrain()

	// Store class vectors in accelerator
	for i, label := range sys.getUniqueLabels(labels) {
		if classVec, exists := sys.Classifier.ClassVectors[label]; exists {
			sys.Accelerator.StoreClassVector(i, classVec)
		}
	}
}

// getUniqueLabels returns unique labels from slice
func (sys *IronLatticeHDCSystem) getUniqueLabels(labels []string) []string {
	seen := make(map[string]bool)
	var unique []string

	for _, label := range labels {
		if !seen[label] {
			seen[label] = true
			unique = append(unique, label)
		}
	}

	return unique
}

// Inference performs HDC inference on a sequence
func (sys *IronLatticeHDCSystem) Inference(sequence []string) (string, float64) {
	// Encode sequence
	encoded := sys.Encoder.EncodeSequence(sequence, sys.Config.NGramSize)

	// Perform associative search
	classIdx, similarity := sys.Accelerator.AssociativeSearch(encoded)

	// Map class index back to label
	labels := sys.getUniqueLabels(nil)
	if classIdx >= 0 && classIdx < len(labels) {
		return labels[classIdx], similarity
	}

	// Fallback to software classifier
	return sys.Classifier.Predict(encoded)
}

// GetSystemMetrics returns comprehensive system metrics
func (sys *IronLatticeHDCSystem) GetSystemMetrics() map[string]interface{} {
	accMetrics := sys.Accelerator.GetPerformanceMetrics()

	// Calculate system-level efficiency
	ops := 2.0 * float64(sys.Config.Dimensions*sys.Config.NumClasses)
	energyJ := accMetrics["total_energy_fJ"].(float64) * 1e-15
	topsPerW := ops / (energyJ * 1e12)

	return map[string]interface{}{
		"config":              sys.Config,
		"accelerator_metrics": accMetrics,
		"tops_per_watt":       topsPerW,
		"inference_count":     sys.InferenceCount,
		"num_crossbars":       len(sys.CrossbarBank),
		"total_cells":         len(sys.CrossbarBank) * sys.Config.ArraySize * sys.Config.ArraySize,
		"static_power_W":      0.0, // Memcapacitor: zero static power
	}
}

// ExportConfiguration exports system configuration
func (sys *IronLatticeHDCSystem) ExportConfiguration() ([]byte, error) {
	export := map[string]interface{}{
		"config":  sys.Config,
		"metrics": sys.GetSystemMetrics(),
		"encoder": map[string]interface{}{
			"item_memory_size":  len(sys.Encoder.ItemMemory),
			"position_vec_count": len(sys.Encoder.PositionVecs),
		},
		"classifier": map[string]interface{}{
			"num_classes": len(sys.Classifier.ClassVectors),
		},
	}

	return json.MarshalIndent(export, "", "  ")
}

// ============================================================================
// BENCHMARKS
// ============================================================================

// MemcapacitorBenchmark runs memcapacitor performance benchmarks
type MemcapacitorBenchmark struct {
	Results []map[string]interface{}
}

// RunEnergyBenchmark compares memcapacitor vs resistive crossbar
func (mb *MemcapacitorBenchmark) RunEnergyBenchmark(sizes []int) {
	for _, size := range sizes {
		// Memcapacitor crossbar
		memcapCB := NewMemcapacitorCrossbar(size, size, nil)

		// Simulated resistive crossbar energy
		resistiveEnergy := memcapCB.EnergyPerMVM * 50.0 // 20-200× higher

		result := map[string]interface{}{
			"array_size":           size,
			"memcap_energy_pJ":     memcapCB.EnergyPerMVM,
			"resistive_energy_pJ":  resistiveEnergy,
			"energy_improvement":   resistiveEnergy / memcapCB.EnergyPerMVM,
			"memcap_efficiency":    memcapCB.GetEnergyEfficiency(),
		}

		mb.Results = append(mb.Results, result)
	}
}

// HDCBenchmark runs HDC accuracy benchmarks
type HDCBenchmark struct {
	Config  *HypervectorConfig
	Results []map[string]interface{}
}

// RunDimensionalityBenchmark tests accuracy vs dimensions
func (hb *HDCBenchmark) RunDimensionalityBenchmark(dimensions []int, numClasses int) {
	for _, dim := range dimensions {
		config := &HypervectorConfig{
			Dimensions:   dim,
			Model:        BinarySpatterCode,
			Sparsity:     0.5,
			Quantization: 1,
		}

		clf := NewHDCClassifier(config)

		// Generate synthetic training data
		for c := 0; c < numClasses; c++ {
			label := fmt.Sprintf("class_%d", c)
			for s := 0; s < 100; s++ { // 100 samples per class
				sample := NewHypervector(config)
				clf.Train(label, sample)
			}
		}
		clf.Retrain()

		// Generate test data and evaluate
		correct := 0
		total := numClasses * 20 // 20 test samples per class
		for c := 0; c < numClasses; c++ {
			label := fmt.Sprintf("class_%d", c)
			for t := 0; t < 20; t++ {
				// Test samples are similar to training (with noise)
				testSample := clf.ClassVectors[label].Clone()
				// Add noise
				for i := range testSample.Elements {
					if rand.Float64() < 0.1 { // 10% noise
						testSample.Elements[i] *= -1
					}
				}

				predicted, _ := clf.Predict(testSample)
				if predicted == label {
					correct++
				}
			}
		}

		accuracy := float64(correct) / float64(total)

		result := map[string]interface{}{
			"dimensions":  dim,
			"num_classes": numClasses,
			"accuracy":    accuracy,
			"noise_tolerance": accuracy,
		}

		hb.Results = append(hb.Results, result)
	}
}

// GenerateReport generates benchmark report
func (hb *HDCBenchmark) GenerateReport() string {
	report := "# HDC Benchmark Report\n\n"
	report += "## Dimensionality vs Accuracy\n\n"
	report += "| Dimensions | Classes | Accuracy |\n"
	report += "|------------|---------|----------|\n"

	for _, r := range hb.Results {
		report += fmt.Sprintf("| %d | %d | %.2f%% |\n",
			r["dimensions"].(int),
			r["num_classes"].(int),
			r["accuracy"].(float64)*100)
	}

	return report
}
