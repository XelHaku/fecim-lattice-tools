// Package layers provides optical/photonic CIM accelerator and HBM4 memory simulations.
// Based on Lightmatter Envise, MZI/MRR architectures, and HBM4 JEDEC specification.
package layers

import (
	"fmt"
	"math"
	"math/rand"
)

// =============================================================================
// PHOTONIC TENSOR CORE ARCHITECTURES
// =============================================================================

// PhotonicConfig holds configuration for photonic accelerators.
type PhotonicConfig struct {
	Architecture     string    // "mzi_mesh", "mrr_weight", "pcm_crossbar"
	ArraySize        int       // Tensor core size (e.g., 128)
	Wavelength       float64   // Operating wavelength (nm)
	ClockSpeedGHz    float64   // Modulation frequency
	OpticalPowerW    float64   // Laser power (W)
	ElectricalPowerW float64   // Control electronics power (W)
	BitPrecision     int       // Effective bit precision
	InsertionLossDB  float64   // Optical insertion loss per stage (dB)
}

// LightmatterEnviseConfig returns Lightmatter Envise-like configuration.
func LightmatterEnviseConfig() *PhotonicConfig {
	return &PhotonicConfig{
		Architecture:     "mzi_mesh",
		ArraySize:        128,          // 128×128 tensor cores
		Wavelength:       1550,         // C-band telecom
		ClockSpeedGHz:    25,           // 25 GHz modulation
		OpticalPowerW:    1.6,          // 1.6 W optical
		ElectricalPowerW: 78,           // 78 W electrical
		BitPrecision:     16,           // ABFP16
		InsertionLossDB:  0.05,         // Per MZI
	}
}

// MRRTensorCoreConfig returns microring resonator tensor core configuration.
func MRRTensorCoreConfig() *PhotonicConfig {
	return &PhotonicConfig{
		Architecture:     "mrr_weight",
		ArraySize:        64,           // 64×64 core
		Wavelength:       1550,
		ClockSpeedGHz:    25,
		OpticalPowerW:    0.5,
		ElectricalPowerW: 20,
		BitPrecision:     8,
		InsertionLossDB:  0.1,
	}
}

// PCMCrossbarConfig returns phase-change material crossbar configuration.
func PCMCrossbarConfig() *PhotonicConfig {
	return &PhotonicConfig{
		Architecture:     "pcm_crossbar",
		ArraySize:        64,
		Wavelength:       1550,
		ClockSpeedGHz:    10,
		OpticalPowerW:    0.3,
		ElectricalPowerW: 15,
		BitPrecision:     4,            // Multi-level PCM
		InsertionLossDB:  0.2,
	}
}

// =============================================================================
// MACH-ZEHNDER INTERFEROMETER (MZI) MESH
// =============================================================================

// MZIElement represents a single Mach-Zehnder interferometer.
type MZIElement struct {
	PhaseShift1    float64   // Internal phase (θ)
	PhaseShift2    float64   // External phase (φ)
	InsertionLoss  float64   // dB loss per element
	Transmission   float64   // Output power ratio
}

// NewMZIElement creates a new MZI with random initialization.
func NewMZIElement(loss float64) *MZIElement {
	return &MZIElement{
		PhaseShift1:   rand.Float64() * 2 * math.Pi,
		PhaseShift2:   rand.Float64() * 2 * math.Pi,
		InsertionLoss: loss,
		Transmission:  1.0,
	}
}

// SetWeight programs the MZI to represent a weight value.
func (m *MZIElement) SetWeight(weight float64) {
	// Map weight [-1, 1] to phase shifts
	// Using SVD decomposition scheme
	m.PhaseShift1 = math.Acos(math.Abs(weight)) * 2
	if weight < 0 {
		m.PhaseShift2 = math.Pi
	} else {
		m.PhaseShift2 = 0
	}
	m.updateTransmission()
}

// updateTransmission calculates transmission based on phases.
func (m *MZIElement) updateTransmission() {
	// MZI transfer function: T = cos²(θ/2) * exp(iφ)
	m.Transmission = math.Pow(math.Cos(m.PhaseShift1/2), 2)
	// Apply insertion loss
	m.Transmission *= math.Pow(10, -m.InsertionLoss/10)
}

// Compute performs optical interference computation.
func (m *MZIElement) Compute(input1, input2 float64) (float64, float64) {
	// Beam splitter + phase shift model
	cos_t := math.Cos(m.PhaseShift1 / 2)
	sin_t := math.Sin(m.PhaseShift1 / 2)

	out1 := cos_t*input1 + sin_t*input2*math.Cos(m.PhaseShift2)
	out2 := -sin_t*input1 + cos_t*input2*math.Cos(m.PhaseShift2)

	// Apply loss
	lossFactor := math.Pow(10, -m.InsertionLoss/20)
	return out1 * lossFactor, out2 * lossFactor
}

// MZIMesh implements a rectangular MZI mesh for unitary transformation.
type MZIMesh struct {
	Config    *PhotonicConfig
	Elements  [][]*MZIElement
	Size      int
	TotalLoss float64   // Accumulated insertion loss
}

// NewMZIMesh creates a new MZI mesh.
func NewMZIMesh(cfg *PhotonicConfig) *MZIMesh {
	size := cfg.ArraySize
	// Clements architecture: N(N-1)/2 MZIs for N×N unitary
	numLayers := size - 1

	elements := make([][]*MZIElement, numLayers)
	for layer := 0; layer < numLayers; layer++ {
		numMZIs := size / 2
		if layer%2 == 1 {
			numMZIs = (size - 1) / 2
		}
		elements[layer] = make([]*MZIElement, numMZIs)
		for i := range elements[layer] {
			elements[layer][i] = NewMZIElement(cfg.InsertionLossDB)
		}
	}

	return &MZIMesh{
		Config:   cfg,
		Elements: elements,
		Size:     size,
	}
}

// SetUnitary programs the mesh to implement a unitary matrix.
func (m *MZIMesh) SetUnitary(matrix [][]float64) {
	// Simplified: directly set weights (full SVD decomposition complex)
	for layer, row := range m.Elements {
		for i, mzi := range row {
			if layer < len(matrix) && i < len(matrix[layer]) {
				mzi.SetWeight(matrix[layer][i])
			}
		}
	}
	m.calculateTotalLoss()
}

// calculateTotalLoss computes accumulated loss through mesh.
func (m *MZIMesh) calculateTotalLoss() {
	totalMZIs := 0
	for _, layer := range m.Elements {
		totalMZIs += len(layer)
	}
	m.TotalLoss = m.Config.InsertionLossDB * float64(totalMZIs) / float64(m.Size)
}

// Forward passes input through the MZI mesh.
func (m *MZIMesh) Forward(input []float64) []float64 {
	if len(input) < m.Size {
		padded := make([]float64, m.Size)
		copy(padded, input)
		input = padded
	}

	signal := make([]float64, m.Size)
	copy(signal, input[:m.Size])

	// Pass through each layer
	for layerIdx, layer := range m.Elements {
		offset := 0
		if layerIdx%2 == 1 {
			offset = 1
		}

		for i, mzi := range layer {
			idx1 := offset + i*2
			idx2 := offset + i*2 + 1
			if idx2 < m.Size {
				signal[idx1], signal[idx2] = mzi.Compute(signal[idx1], signal[idx2])
			}
		}
	}

	return signal
}

// =============================================================================
// MICRORING RESONATOR (MRR) WEIGHT BANK
// =============================================================================

// MRRElement represents a microring resonator for weight multiplication.
type MRRElement struct {
	Radius        float64   // Ring radius (μm)
	Coupling      float64   // Coupling coefficient
	Resonance     float64   // Resonance wavelength (nm)
	Detuning      float64   // Wavelength detuning
	Transmission  float64   // Weight value
	ThermalTuning float64   // Thermal phase shift
}

// NewMRRElement creates a new microring resonator.
func NewMRRElement(radius float64) *MRRElement {
	return &MRRElement{
		Radius:       radius,
		Coupling:     0.2,          // 20% coupling
		Resonance:    1550,         // nm
		Detuning:     0,
		Transmission: 0.5,
	}
}

// SetWeight programs the MRR transmission.
func (m *MRRElement) SetWeight(weight float64) {
	// Weight encoded in thermal detuning
	m.Detuning = weight * 2.0 // nm detuning range
	m.updateTransmission()
}

// updateTransmission calculates Lorentzian transmission.
func (m *MRRElement) updateTransmission() {
	// Lorentzian response: T = 1 - a²/(1 + (2Δλ/FWHM)²)
	fwhm := 0.5 // nm
	lorentzian := 1.0 / (1 + math.Pow(2*m.Detuning/fwhm, 2))
	m.Transmission = 1 - 0.9*lorentzian // Partial coupling
}

// Multiply performs weighted multiplication.
func (m *MRRElement) Multiply(input float64) float64 {
	return input * m.Transmission
}

// MRRWeightBank implements a bank of MRRs for dot product.
type MRRWeightBank struct {
	Config     *PhotonicConfig
	Resonators []*MRRElement
	Size       int
}

// NewMRRWeightBank creates a new MRR weight bank.
func NewMRRWeightBank(cfg *PhotonicConfig) *MRRWeightBank {
	resonators := make([]*MRRElement, cfg.ArraySize)
	for i := range resonators {
		resonators[i] = NewMRRElement(5.0) // 5 μm radius
	}

	return &MRRWeightBank{
		Config:     cfg,
		Resonators: resonators,
		Size:       cfg.ArraySize,
	}
}

// SetWeights programs all weights.
func (wb *MRRWeightBank) SetWeights(weights []float64) {
	for i, w := range weights {
		if i < len(wb.Resonators) {
			wb.Resonators[i].SetWeight(w)
		}
	}
}

// DotProduct computes weighted sum.
func (wb *MRRWeightBank) DotProduct(input []float64) float64 {
	sum := 0.0
	for i, r := range wb.Resonators {
		if i < len(input) {
			sum += r.Multiply(input[i])
		}
	}
	return sum
}

// =============================================================================
// PHOTONIC TENSOR CORE
// =============================================================================

// PhotonicTensorCore implements optical matrix-vector multiply.
type PhotonicTensorCore struct {
	Config        *PhotonicConfig
	MZIMeshU      *MZIMesh          // SVD: U matrix
	MZIMeshV      *MZIMesh          // SVD: V^T matrix
	SingularMRRs  *MRRWeightBank    // SVD: Σ diagonal
	WeightMatrix  [][]float64
	ComputeCycles int64
	EnergyPJ      float64
}

// NewPhotonicTensorCore creates a new photonic tensor core.
func NewPhotonicTensorCore(cfg *PhotonicConfig) *PhotonicTensorCore {
	return &PhotonicTensorCore{
		Config:       cfg,
		MZIMeshU:     NewMZIMesh(cfg),
		MZIMeshV:     NewMZIMesh(cfg),
		SingularMRRs: NewMRRWeightBank(cfg),
		WeightMatrix: make([][]float64, cfg.ArraySize),
	}
}

// SetWeights programs weight matrix using SVD decomposition.
func (tc *PhotonicTensorCore) SetWeights(weights [][]float64) {
	tc.WeightMatrix = weights

	// Simplified SVD simulation (real implementation uses actual SVD)
	// U * Σ * V^T decomposition
	diagonal := make([]float64, tc.Config.ArraySize)
	for i := 0; i < tc.Config.ArraySize && i < len(weights); i++ {
		// Extract diagonal-like values
		if i < len(weights[i]) {
			diagonal[i] = weights[i][i]
		}
	}

	tc.SingularMRRs.SetWeights(diagonal)
	tc.MZIMeshU.SetUnitary(weights)
	tc.MZIMeshV.SetUnitary(weights)
}

// MatVecMul performs optical matrix-vector multiplication.
func (tc *PhotonicTensorCore) MatVecMul(input []float64) []float64 {
	// SVD: y = U * Σ * V^T * x

	// Step 1: V^T * x (right unitary)
	vOut := tc.MZIMeshV.Forward(input)

	// Step 2: Σ * (V^T * x) (scaling by singular values)
	scaled := make([]float64, len(vOut))
	for i, v := range vOut {
		if i < len(tc.SingularMRRs.Resonators) {
			scaled[i] = tc.SingularMRRs.Resonators[i].Multiply(v)
		}
	}

	// Step 3: U * (Σ * V^T * x) (left unitary)
	output := tc.MZIMeshU.Forward(scaled)

	// Track energy
	tc.ComputeCycles++
	tc.EnergyPJ += tc.calculateEnergy()

	return output
}

// calculateEnergy estimates energy per operation.
func (tc *PhotonicTensorCore) calculateEnergy() float64 {
	// Energy dominated by modulators and photodetectors
	opsPerCycle := float64(tc.Config.ArraySize * tc.Config.ArraySize)
	totalPower := tc.Config.OpticalPowerW + tc.Config.ElectricalPowerW
	timePerOp := 1.0 / (tc.Config.ClockSpeedGHz * 1e9)

	return totalPower * timePerOp * 1e12 // pJ
}

// GetTOPS returns compute throughput.
func (tc *PhotonicTensorCore) GetTOPS() float64 {
	opsPerCycle := float64(tc.Config.ArraySize * tc.Config.ArraySize * 2) // MAC = 2 ops
	return opsPerCycle * tc.Config.ClockSpeedGHz / 1000.0 // TOPS
}

// GetTOPSW returns energy efficiency.
func (tc *PhotonicTensorCore) GetTOPSW() float64 {
	totalPower := tc.Config.OpticalPowerW + tc.Config.ElectricalPowerW
	return tc.GetTOPS() / totalPower
}

// =============================================================================
// PHOTONIC ACCELERATOR SYSTEM
// =============================================================================

// PhotonicAccelerator implements multi-core photonic processor.
type PhotonicAccelerator struct {
	Config       *PhotonicConfig
	TensorCores  []*PhotonicTensorCore
	NumCores     int
	TotalOps     int64
	TotalEnergy  float64
}

// NewPhotonicAccelerator creates a new photonic accelerator.
// Based on Lightmatter Envise: 4 tensor cores.
func NewPhotonicAccelerator(cfg *PhotonicConfig, numCores int) *PhotonicAccelerator {
	cores := make([]*PhotonicTensorCore, numCores)
	for i := range cores {
		cores[i] = NewPhotonicTensorCore(cfg)
	}

	return &PhotonicAccelerator{
		Config:      cfg,
		TensorCores: cores,
		NumCores:    numCores,
	}
}

// EnviseAccelerator creates Lightmatter Envise-like accelerator.
func EnviseAccelerator() *PhotonicAccelerator {
	cfg := LightmatterEnviseConfig()
	return NewPhotonicAccelerator(cfg, 4) // 4 × 128×128 cores
}

// ExecuteLayer executes a neural network layer.
func (acc *PhotonicAccelerator) ExecuteLayer(input []float64, weights [][]float64) ([]float64, *PhotonicLayerStats) {
	stats := &PhotonicLayerStats{}

	// Tile computation across cores
	rowsPerCore := (len(weights) + acc.NumCores - 1) / acc.NumCores

	output := make([]float64, len(weights))

	for i, core := range acc.TensorCores {
		startRow := i * rowsPerCore
		endRow := startRow + rowsPerCore
		if endRow > len(weights) {
			endRow = len(weights)
		}
		if startRow >= len(weights) {
			break
		}

		// Set weights for this tile
		tileWeights := weights[startRow:endRow]
		core.SetWeights(tileWeights)

		// Execute MVM
		tileOutput := core.MatVecMul(input)

		// Copy to output
		for j := 0; j < endRow-startRow && j < len(tileOutput); j++ {
			output[startRow+j] = tileOutput[j]
		}

		stats.CoreOps = append(stats.CoreOps, int64(len(tileWeights)*len(input)))
	}

	// Aggregate stats
	for _, core := range acc.TensorCores {
		acc.TotalOps += core.ComputeCycles * int64(acc.Config.ArraySize*acc.Config.ArraySize)
		acc.TotalEnergy += core.EnergyPJ
	}

	stats.TotalOps = acc.TotalOps
	stats.TotalEnergy = acc.TotalEnergy
	stats.TOPS = acc.TensorCores[0].GetTOPS() * float64(acc.NumCores)
	stats.TOPSW = stats.TOPS / (acc.Config.OpticalPowerW + acc.Config.ElectricalPowerW)

	return output, stats
}

// PhotonicLayerStats holds layer execution statistics.
type PhotonicLayerStats struct {
	CoreOps     []int64
	TotalOps    int64
	TotalEnergy float64
	TOPS        float64
	TOPSW       float64
}

// =============================================================================
// HBM4 MEMORY ARCHITECTURE
// =============================================================================

// HBM4Config holds HBM4 memory configuration.
type HBM4Config struct {
	Generation       string    // "HBM4", "HBM4E"
	NumStacks        int       // Number of HBM stacks
	LayersPerStack   int       // DRAM layers (12, 16)
	CapacityGB       int       // Capacity per stack (GB)
	InterfaceWidth   int       // 2048 bits
	NumChannels      int       // 32 channels per stack
	DataRateGbps     float64   // Per-pin data rate
	BandwidthTBs     float64   // Per-stack bandwidth (TB/s)
	CoreVoltage      float64   // Core voltage (V)
	IOVoltage        float64   // I/O voltage (V)
	CustomLogicNode  string    // Logic die process node
}

// DefaultHBM4Config returns HBM4 specification per JEDEC.
func DefaultHBM4Config() *HBM4Config {
	return &HBM4Config{
		Generation:      "HBM4",
		NumStacks:       8,           // Typical AI accelerator
		LayersPerStack:  12,          // 12-Hi
		CapacityGB:      48,          // 48 GB per stack
		InterfaceWidth:  2048,        // 2048-bit interface
		NumChannels:     32,          // 32 × 64-bit channels
		DataRateGbps:    8.0,         // 8 Gb/s per JEDEC
		BandwidthTBs:    2.0,         // 2 TB/s per stack
		CoreVoltage:     1.0,         // 1.0V core
		IOVoltage:       0.8,         // 0.8V I/O
		CustomLogicNode: "12nm",      // TSMC 12nm logic die
	}
}

// HBM4EConfig returns HBM4E enhanced specification.
func HBM4EConfig() *HBM4Config {
	return &HBM4Config{
		Generation:      "HBM4E",
		NumStacks:       8,
		LayersPerStack:  16,          // 16-Hi
		CapacityGB:      64,          // 64 GB per stack
		InterfaceWidth:  2048,
		NumChannels:     32,
		DataRateGbps:    12.0,        // 12 GT/s
		BandwidthTBs:    3.0,         // ~3 TB/s
		CoreVoltage:     1.0,
		IOVoltage:       0.8,
		CustomLogicNode: "4nm",       // Samsung 4nm
	}
}

// HBM4Channel represents a single HBM4 channel.
type HBM4Channel struct {
	Config        *HBM4Config
	ChannelID     int
	Width         int         // 64 or 32 bits
	ReadQueue     int64
	WriteQueue    int64
	BandwidthUsed float64
	Utilization   float64
}

// NewHBM4Channel creates a new HBM4 channel.
func NewHBM4Channel(cfg *HBM4Config, id int) *HBM4Channel {
	return &HBM4Channel{
		Config:    cfg,
		ChannelID: id,
		Width:     64, // 64-bit channel
	}
}

// Read performs a read operation.
func (c *HBM4Channel) Read(bytes int64) (float64, float64) {
	// Calculate latency and bandwidth
	bits := bytes * 8
	transferTime := float64(bits) / (c.Config.DataRateGbps * 1e9) * 1e9 // ns

	channelBandwidth := c.Config.DataRateGbps * float64(c.Width) / 8 // GB/s
	c.BandwidthUsed += float64(bytes) / 1e9

	c.ReadQueue++

	return transferTime, channelBandwidth
}

// Write performs a write operation.
func (c *HBM4Channel) Write(bytes int64) (float64, float64) {
	bits := bytes * 8
	transferTime := float64(bits) / (c.Config.DataRateGbps * 1e9) * 1e9

	channelBandwidth := c.Config.DataRateGbps * float64(c.Width) / 8
	c.BandwidthUsed += float64(bytes) / 1e9

	c.WriteQueue++

	return transferTime, channelBandwidth
}

// HBM4Stack represents a single HBM4 memory stack.
type HBM4Stack struct {
	Config         *HBM4Config
	StackID        int
	Channels       []*HBM4Channel
	LogicDie       *HBM4LogicDie
	TotalReads     int64
	TotalWrites    int64
	EnergyPJ       float64
}

// NewHBM4Stack creates a new HBM4 stack.
func NewHBM4Stack(cfg *HBM4Config, id int) *HBM4Stack {
	channels := make([]*HBM4Channel, cfg.NumChannels)
	for i := range channels {
		channels[i] = NewHBM4Channel(cfg, i)
	}

	return &HBM4Stack{
		Config:   cfg,
		StackID:  id,
		Channels: channels,
		LogicDie: NewHBM4LogicDie(cfg),
	}
}

// Read performs read across all channels.
func (s *HBM4Stack) Read(address int64, bytes int64) *HBM4AccessResult {
	result := &HBM4AccessResult{
		Type: "read",
	}

	// Distribute across channels (interleaved)
	bytesPerChannel := bytes / int64(s.Config.NumChannels)
	if bytesPerChannel == 0 {
		bytesPerChannel = bytes
	}

	totalLatency := 0.0
	for i, ch := range s.Channels {
		if int64(i)*bytesPerChannel >= bytes {
			break
		}
		latency, _ := ch.Read(bytesPerChannel)
		if latency > totalLatency {
			totalLatency = latency // Parallel - max latency
		}
	}

	result.LatencyNS = totalLatency
	result.Bandwidth = s.Config.BandwidthTBs
	result.EnergyPJ = s.calculateReadEnergy(bytes)

	s.TotalReads++
	s.EnergyPJ += result.EnergyPJ

	// Process through logic die if custom compute enabled
	if s.LogicDie.ComputeEnabled {
		result.ComputeLatencyNS = s.LogicDie.ProcessData(bytes)
	}

	return result
}

// Write performs write across all channels.
func (s *HBM4Stack) Write(address int64, data []byte) *HBM4AccessResult {
	result := &HBM4AccessResult{
		Type: "write",
	}

	bytes := int64(len(data))
	bytesPerChannel := bytes / int64(s.Config.NumChannels)
	if bytesPerChannel == 0 {
		bytesPerChannel = bytes
	}

	totalLatency := 0.0
	for i, ch := range s.Channels {
		if int64(i)*bytesPerChannel >= bytes {
			break
		}
		latency, _ := ch.Write(bytesPerChannel)
		if latency > totalLatency {
			totalLatency = latency
		}
	}

	result.LatencyNS = totalLatency
	result.Bandwidth = s.Config.BandwidthTBs
	result.EnergyPJ = s.calculateWriteEnergy(bytes)

	s.TotalWrites++
	s.EnergyPJ += result.EnergyPJ

	return result
}

// calculateReadEnergy estimates read energy.
func (s *HBM4Stack) calculateReadEnergy(bytes int64) float64 {
	// ~3-4 pJ/bit for HBM4
	return float64(bytes*8) * 3.5
}

// calculateWriteEnergy estimates write energy.
func (s *HBM4Stack) calculateWriteEnergy(bytes int64) float64 {
	// Writes slightly more expensive
	return float64(bytes*8) * 4.0
}

// HBM4AccessResult holds memory access result.
type HBM4AccessResult struct {
	Type             string
	LatencyNS        float64
	Bandwidth        float64
	EnergyPJ         float64
	ComputeLatencyNS float64
}

// =============================================================================
// HBM4 CUSTOM LOGIC DIE
// =============================================================================

// HBM4LogicDie represents the custom logic die at base of HBM4 stack.
type HBM4LogicDie struct {
	Config         *HBM4Config
	ProcessNode    string      // "12nm", "4nm"
	ComputeEnabled bool        // Near-memory compute
	CacheKB        int         // On-die cache
	PreprocessOps  []string    // Supported operations
	ComputeCycles  int64
}

// NewHBM4LogicDie creates a new logic die.
func NewHBM4LogicDie(cfg *HBM4Config) *HBM4LogicDie {
	return &HBM4LogicDie{
		Config:         cfg,
		ProcessNode:    cfg.CustomLogicNode,
		ComputeEnabled: true,
		CacheKB:        256,   // 256 KB on-die cache
		PreprocessOps:  []string{"gather", "scatter", "reduce", "prefetch"},
	}
}

// ProcessData performs near-memory processing.
func (ld *HBM4LogicDie) ProcessData(bytes int64) float64 {
	// Simple preprocessing latency model
	// 12nm: ~1ns/64B, 4nm: ~0.5ns/64B
	latencyPerBlock := 1.0
	if ld.ProcessNode == "4nm" {
		latencyPerBlock = 0.5
	}

	blocks := (bytes + 63) / 64
	ld.ComputeCycles += blocks

	return float64(blocks) * latencyPerBlock
}

// EnableCompute enables near-memory compute features.
func (ld *HBM4LogicDie) EnableCompute(ops []string) {
	ld.ComputeEnabled = true
	ld.PreprocessOps = ops
}

// =============================================================================
// HBM4 MEMORY SYSTEM
// =============================================================================

// HBM4System implements complete HBM4 memory subsystem.
type HBM4System struct {
	Config           *HBM4Config
	Stacks           []*HBM4Stack
	TotalCapacityGB  int
	TotalBandwidthTB float64
	TotalReads       int64
	TotalWrites      int64
	TotalEnergy      float64
}

// NewHBM4System creates a new HBM4 memory system.
func NewHBM4System(cfg *HBM4Config) *HBM4System {
	stacks := make([]*HBM4Stack, cfg.NumStacks)
	for i := range stacks {
		stacks[i] = NewHBM4Stack(cfg, i)
	}

	return &HBM4System{
		Config:           cfg,
		Stacks:           stacks,
		TotalCapacityGB:  cfg.NumStacks * cfg.CapacityGB,
		TotalBandwidthTB: float64(cfg.NumStacks) * cfg.BandwidthTBs,
	}
}

// DefaultHBM4System creates 8-stack HBM4 system (~16 TB/s).
func DefaultHBM4System() *HBM4System {
	return NewHBM4System(DefaultHBM4Config())
}

// Read performs distributed read across stacks.
func (sys *HBM4System) Read(address int64, bytes int64) *HBM4SystemAccess {
	result := &HBM4SystemAccess{}

	// Address-based stack selection (interleaved)
	stackID := int(address / int64(sys.Config.CapacityGB*1024*1024*1024)) % len(sys.Stacks)

	accessResult := sys.Stacks[stackID].Read(address, bytes)

	result.StackID = stackID
	result.LatencyNS = accessResult.LatencyNS
	result.Bandwidth = accessResult.Bandwidth
	result.EnergyPJ = accessResult.EnergyPJ

	sys.TotalReads++
	sys.TotalEnergy += accessResult.EnergyPJ

	return result
}

// Write performs write to appropriate stack.
func (sys *HBM4System) Write(address int64, data []byte) *HBM4SystemAccess {
	result := &HBM4SystemAccess{}

	stackID := int(address / int64(sys.Config.CapacityGB*1024*1024*1024)) % len(sys.Stacks)

	accessResult := sys.Stacks[stackID].Write(address, data)

	result.StackID = stackID
	result.LatencyNS = accessResult.LatencyNS
	result.Bandwidth = accessResult.Bandwidth
	result.EnergyPJ = accessResult.EnergyPJ

	sys.TotalWrites++
	sys.TotalEnergy += accessResult.EnergyPJ

	return result
}

// GetStats returns system statistics.
func (sys *HBM4System) GetStats() *HBM4Stats {
	return &HBM4Stats{
		TotalCapacityGB:  sys.TotalCapacityGB,
		TotalBandwidthTB: sys.TotalBandwidthTB,
		TotalReads:       sys.TotalReads,
		TotalWrites:      sys.TotalWrites,
		TotalEnergyPJ:    sys.TotalEnergy,
	}
}

// HBM4SystemAccess holds system access result.
type HBM4SystemAccess struct {
	StackID   int
	LatencyNS float64
	Bandwidth float64
	EnergyPJ  float64
}

// HBM4Stats holds system statistics.
type HBM4Stats struct {
	TotalCapacityGB  int
	TotalBandwidthTB float64
	TotalReads       int64
	TotalWrites      int64
	TotalEnergyPJ    float64
}

// =============================================================================
// PHOTONIC + HBM4 INTEGRATED SYSTEM
// =============================================================================

// PhotonicHBM4System integrates photonic accelerator with HBM4.
type PhotonicHBM4System struct {
	Accelerator  *PhotonicAccelerator
	Memory       *HBM4System
	TotalOps     int64
	TotalEnergy  float64
}

// NewPhotonicHBM4System creates integrated photonic + HBM4 system.
func NewPhotonicHBM4System() *PhotonicHBM4System {
	return &PhotonicHBM4System{
		Accelerator: EnviseAccelerator(),
		Memory:      DefaultHBM4System(),
	}
}

// ExecuteInference runs inference workload.
func (sys *PhotonicHBM4System) ExecuteInference(
	inputs [][]float64,
	weights [][][]float64,
) (*InferenceResult, error) {
	result := &InferenceResult{
		LayerStats: make([]*LayerInferenceStats, len(weights)),
	}

	currentInput := inputs[0]

	for layer, w := range weights {
		layerStats := &LayerInferenceStats{LayerID: layer}

		// Load weights from HBM4
		weightBytes := int64(len(w) * len(w[0]) * 4)
		memResult := sys.Memory.Read(int64(layer)*weightBytes, weightBytes)
		layerStats.MemoryLatencyNS = memResult.LatencyNS
		layerStats.MemoryEnergy = memResult.EnergyPJ

		// Execute on photonic accelerator
		output, computeStats := sys.Accelerator.ExecuteLayer(currentInput, w)
		layerStats.ComputeOps = computeStats.TotalOps
		layerStats.ComputeEnergy = computeStats.TotalEnergy
		layerStats.TOPS = computeStats.TOPS
		layerStats.TOPSW = computeStats.TOPSW

		currentInput = output
		result.LayerStats[layer] = layerStats
	}

	// Aggregate results
	for _, ls := range result.LayerStats {
		result.TotalOps += ls.ComputeOps
		result.TotalEnergy += ls.ComputeEnergy + ls.MemoryEnergy
		result.TotalLatency += ls.MemoryLatencyNS
	}

	result.Output = currentInput
	result.EffectiveTOPS = float64(result.TotalOps) / (result.TotalLatency / 1e9) / 1e12
	result.EffectiveTOPSW = result.EffectiveTOPS / ((result.TotalEnergy / 1e12) / (result.TotalLatency / 1e9))

	sys.TotalOps = result.TotalOps
	sys.TotalEnergy = result.TotalEnergy

	return result, nil
}

// InferenceResult holds inference execution results.
type InferenceResult struct {
	Output         []float64
	LayerStats     []*LayerInferenceStats
	TotalOps       int64
	TotalEnergy    float64
	TotalLatency   float64
	EffectiveTOPS  float64
	EffectiveTOPSW float64
}

// LayerInferenceStats holds per-layer statistics.
type LayerInferenceStats struct {
	LayerID         int
	ComputeOps      int64
	ComputeEnergy   float64
	MemoryLatencyNS float64
	MemoryEnergy    float64
	TOPS            float64
	TOPSW           float64
}

// =============================================================================
// BENCHMARK COMPARISON
// =============================================================================

// AcceleratorBenchmark compares accelerator technologies.
type AcceleratorBenchmark struct {
	Results map[string]*AcceleratorMetrics
}

// AcceleratorMetrics holds performance metrics.
type AcceleratorMetrics struct {
	Name           string
	Technology     string
	TOPS           float64
	TOPSW          float64
	TOPSmm2        float64   // Compute density
	PowerW         float64
	MemoryBandwidth float64  // TB/s
	Precision      string
}

// NewAcceleratorBenchmark creates a benchmark comparison.
func NewAcceleratorBenchmark() *AcceleratorBenchmark {
	return &AcceleratorBenchmark{
		Results: make(map[string]*AcceleratorMetrics),
	}
}

// AddLightmatterEnvise adds Lightmatter metrics.
func (b *AcceleratorBenchmark) AddLightmatterEnvise() {
	b.Results["Lightmatter_Envise"] = &AcceleratorMetrics{
		Name:            "Lightmatter Envise",
		Technology:      "Silicon Photonics (MZI)",
		TOPS:            65.5,
		TOPSW:           0.82,     // 65.5/79.6W
		TOPSmm2:         0.5,      // Estimated
		PowerW:          79.6,     // 78W + 1.6W optical
		MemoryBandwidth: 0,        // External
		Precision:       "ABFP16",
	}
}

// AddMRRTensorCore adds MRR tensor core metrics.
func (b *AcceleratorBenchmark) AddMRRTensorCore() {
	b.Results["MRR_TensorCore"] = &AcceleratorMetrics{
		Name:            "MRR In-Memory Tensor Core",
		Technology:      "Silicon Photonics (MRR)",
		TOPS:            204.8,     // 64×64 @ 25GHz
		TOPSW:           5.1,       // Predicted
		TOPSmm2:         880,       // Predicted
		PowerW:          40,
		MemoryBandwidth: 0,
		Precision:       "INT8",
	}
}

// AddNvidiaH100 adds NVIDIA H100 metrics for comparison.
func (b *AcceleratorBenchmark) AddNvidiaH100() {
	b.Results["NVIDIA_H100"] = &AcceleratorMetrics{
		Name:            "NVIDIA H100 SXM5",
		Technology:      "Digital CMOS (4nm)",
		TOPS:            1979,      // FP16 Tensor Core
		TOPSW:           2.8,       // 1979/700W
		TOPSmm2:         2.4,       // 1979/814mm²
		PowerW:          700,
		MemoryBandwidth: 3.35,      // HBM3
		Precision:       "FP16",
	}
}

// AddHBM4System adds HBM4 memory metrics.
func (b *AcceleratorBenchmark) AddHBM4System() {
	b.Results["HBM4_8Stack"] = &AcceleratorMetrics{
		Name:            "HBM4 8-Stack System",
		Technology:      "3D DRAM (12nm logic)",
		TOPS:            0,         // Memory, not compute
		TOPSW:           0,
		TOPSmm2:         0,
		PowerW:          100,       // Estimated
		MemoryBandwidth: 16,        // 8 × 2 TB/s
		Precision:       "N/A",
	}
}

// Compare generates comparison report.
func (b *AcceleratorBenchmark) Compare() string {
	report := "AI Accelerator Technology Comparison\n"
	report += "=====================================\n\n"

	for name, m := range b.Results {
		report += fmt.Sprintf("%s (%s):\n", name, m.Technology)
		if m.TOPS > 0 {
			report += fmt.Sprintf("  Throughput: %.1f TOPS\n", m.TOPS)
			report += fmt.Sprintf("  Efficiency: %.2f TOPS/W\n", m.TOPSW)
		}
		if m.TOPSmm2 > 0 {
			report += fmt.Sprintf("  Density: %.1f TOPS/mm²\n", m.TOPSmm2)
		}
		report += fmt.Sprintf("  Power: %.0f W\n", m.PowerW)
		if m.MemoryBandwidth > 0 {
			report += fmt.Sprintf("  Memory BW: %.2f TB/s\n", m.MemoryBandwidth)
		}
		report += fmt.Sprintf("  Precision: %s\n\n", m.Precision)
	}

	return report
}
