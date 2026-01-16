// emerging_analog_cim.go - Emerging Memory Technologies and Analog Gradient Descent for CIM
// Part of the IronLattice CIM simulation framework
// Iteration 129: PCM, MRAM, FTJ comparison + Analog in-memory training

package layers

import (
	"fmt"
	"math"
	"math/rand"
)

// =============================================================================
// EMERGING MEMORY TECHNOLOGY COMPARISON
// =============================================================================

// MemoryTechnology represents different NVM technologies
type MemoryTechnology int

const (
	MemoryReRAM MemoryTechnology = iota
	MemoryPCM
	MemoryMRAM
	MemoryFTJ
	MemoryFeFET
	Memory2DFTJ // 2D materials FTJ (MoS2-based)
)

// MemoryTechSpec defines specifications for a memory technology
type MemoryTechSpec struct {
	Technology      MemoryTechnology
	Name            string
	WriteEnergyFJ   float64 // Femtojoules per write
	ReadEnergyFJ    float64 // Femtojoules per read
	SwitchingTimeNs float64 // Nanoseconds
	EnduranceCycles float64 // Write cycles before degradation
	RetentionYears  float64 // Data retention at 85°C
	OnOffRatio      float64 // Resistance ratio
	CellSizeF2      int     // Cell size in F² units
	ConductLevels   int     // MLC levels
	DriftRate       float64 // Resistance drift per decade (for PCM)
	SNR             float64 // Signal-to-noise ratio (dB)
}

// GetMemorySpecs returns specifications for each memory technology
func GetMemorySpecs() map[MemoryTechnology]*MemoryTechSpec {
	return map[MemoryTechnology]*MemoryTechSpec{
		MemoryReRAM: {
			Technology:      MemoryReRAM,
			Name:            "ReRAM (HfOx)",
			WriteEnergyFJ:   100,   // 100 fJ typical
			ReadEnergyFJ:    10,    // 10 fJ
			SwitchingTimeNs: 10,    // 10 ns
			EnduranceCycles: 1e6,   // 10^6 cycles
			RetentionYears:  10,    // 10 years
			OnOffRatio:      100,   // 100:1
			CellSizeF2:      4,     // 4F²
			ConductLevels:   16,    // 4-bit MLC
			DriftRate:       0.01,  // 1% per decade
			SNR:             30,    // 30 dB
		},
		MemoryPCM: {
			Technology:      MemoryPCM,
			Name:            "PCM (GST)",
			WriteEnergyFJ:   1000,  // 1 pJ (high due to heating)
			ReadEnergyFJ:    50,    // 50 fJ
			SwitchingTimeNs: 50,    // 50 ns
			EnduranceCycles: 1e9,   // 10^9 cycles
			RetentionYears:  10,    // 10 years
			OnOffRatio:      1000,  // 1000:1
			CellSizeF2:      4,     // 4F²
			ConductLevels:   16,    // 4-bit MLC
			DriftRate:       0.05,  // 5% per decade (major issue)
			SNR:             35,    // 35 dB
		},
		MemoryMRAM: {
			Technology:      MemoryMRAM,
			Name:            "STT-MRAM",
			WriteEnergyFJ:   500,   // 500 fJ
			ReadEnergyFJ:    20,    // 20 fJ
			SwitchingTimeNs: 5,     // 5 ns (fast!)
			EnduranceCycles: 1e15,  // Virtually unlimited
			RetentionYears:  10,    // 10 years
			OnOffRatio:      3,     // 3:1 (small - major limitation)
			CellSizeF2:      20,    // 20F² (large due to transistor)
			ConductLevels:   2,     // Binary only
			DriftRate:       0,     // No drift
			SNR:             20,    // 20 dB (limited by low TMR)
		},
		MemoryFTJ: {
			Technology:      MemoryFTJ,
			Name:            "FTJ (HZO)",
			WriteEnergyFJ:   50,    // 10-500 fJ, use 50 fJ typical
			ReadEnergyFJ:    5,     // 5 fJ
			SwitchingTimeNs: 10,    // 10 ns
			EnduranceCycles: 1e10,  // 10^10 cycles
			RetentionYears:  10,    // 10 years
			OnOffRatio:      100,   // 100:1
			CellSizeF2:      4,     // 4F²
			ConductLevels:   64,    // 6-bit (60 states demonstrated)
			DriftRate:       0.005, // 0.5% per decade (low)
			SNR:             40,    // 40 dB
		},
		MemoryFeFET: {
			Technology:      MemoryFeFET,
			Name:            "FeFET (HZO gate)",
			WriteEnergyFJ:   30,    // 30 fJ
			ReadEnergyFJ:    10,    // 10 fJ
			SwitchingTimeNs: 20,    // 20 ns
			EnduranceCycles: 1e8,   // 10^8 cycles
			RetentionYears:  10,    // 10 years
			OnOffRatio:      1e4,   // 10000:1 (large window)
			CellSizeF2:      6,     // 6F²
			ConductLevels:   32,    // 5-bit MLC
			DriftRate:       0.01,  // 1% per decade
			SNR:             38,    // 38 dB
		},
		Memory2DFTJ: {
			Technology:      Memory2DFTJ,
			Name:            "2D-FTJ (MoS2)",
			WriteEnergyFJ:   20,    // 20 fJ (very low)
			ReadEnergyFJ:    2,     // 2 fJ
			SwitchingTimeNs: 5,     // 5 ns
			EnduranceCycles: 1e12,  // 10^12 cycles
			RetentionYears:  5,     // 5 years (still improving)
			OnOffRatio:      1e6,   // 10^6:1 (excellent)
			CellSizeF2:      4,     // 4F²
			ConductLevels:   128,   // 7-bit potential
			DriftRate:       0.002, // 0.2% per decade
			SNR:             45,    // 45 dB
		},
	}
}

// MemoryCell represents a single memory cell with technology-specific behavior
type MemoryCell struct {
	Technology   MemoryTechnology
	Spec         *MemoryTechSpec
	Conductance  float64 // Current conductance state
	TargetCond   float64 // Target conductance
	WriteCount   int64   // Total write operations
	AgeDecades   float64 // Time since last write (in decades for drift)
	IsDegraded   bool    // Endurance limit reached
}

// NewMemoryCell creates a new memory cell
func NewMemoryCell(tech MemoryTechnology) *MemoryCell {
	specs := GetMemorySpecs()
	return &MemoryCell{
		Technology:  tech,
		Spec:        specs[tech],
		Conductance: 0.5, // Mid-range initial
		WriteCount:  0,
	}
}

// Write programs the cell to a target conductance
func (mc *MemoryCell) Write(target float64) error {
	if mc.IsDegraded {
		return fmt.Errorf("cell degraded after %d writes", mc.WriteCount)
	}

	// Check endurance
	mc.WriteCount++
	if float64(mc.WriteCount) > mc.Spec.EnduranceCycles {
		mc.IsDegraded = true
	}

	// Quantize to available levels
	levels := float64(mc.Spec.ConductLevels)
	quantized := math.Round(target*levels) / levels

	// Add programming noise
	noiseStd := 1.0 / mc.Spec.SNR
	noise := rand.NormFloat64() * noiseStd
	mc.Conductance = math.Max(0, math.Min(1, quantized+noise))
	mc.TargetCond = target
	mc.AgeDecades = 0 // Reset drift counter

	return nil
}

// Read returns the current conductance with drift and noise
func (mc *MemoryCell) Read() float64 {
	// Apply drift (especially for PCM)
	driftedCond := mc.Conductance * (1 - mc.Spec.DriftRate*mc.AgeDecades)

	// Add read noise
	noiseStd := 1.0 / mc.Spec.SNR * 0.5 // Read noise is lower
	noise := rand.NormFloat64() * noiseStd

	return math.Max(0, math.Min(1, driftedCond+noise))
}

// Age simulates time passage for drift
func (mc *MemoryCell) Age(decades float64) {
	mc.AgeDecades += decades
}

// =============================================================================
// FERROELECTRIC TUNNEL JUNCTION (FTJ) CROSSBAR
// =============================================================================

// FTJCrossbarConfig configures FTJ crossbar array
type FTJCrossbarConfig struct {
	Rows           int
	Cols           int
	ConductLevels  int     // MLC levels (up to 64 for HZO FTJ)
	Nonlinearity   float64 // I-V nonlinearity (>1100 for advanced FTJ)
	DynamicRange   float64 // On/off ratio (10 demonstrated)
	HalfBiasSelect bool    // Half-bias selection scheme
	ReadVoltage    float64 // Read voltage (0.3V typical)
}

// FTJCrossbar implements FTJ-based crossbar array
type FTJCrossbar struct {
	Config       *FTJCrossbarConfig
	Cells        [][]*MemoryCell
	SneakCurrent [][]float64 // Sneak current map
	Stats        *FTJStats
}

// FTJStats tracks FTJ crossbar statistics
type FTJStats struct {
	TotalReads     int64
	TotalWrites    int64
	SneakErrors    int
	ProgramErrors  int
	ArrayYield     float64
}

// NewFTJCrossbar creates an FTJ crossbar array
func NewFTJCrossbar(config *FTJCrossbarConfig) *FTJCrossbar {
	cb := &FTJCrossbar{
		Config:       config,
		Cells:        make([][]*MemoryCell, config.Rows),
		SneakCurrent: make([][]float64, config.Rows),
		Stats:        &FTJStats{ArrayYield: 1.0},
	}

	for i := 0; i < config.Rows; i++ {
		cb.Cells[i] = make([]*MemoryCell, config.Cols)
		cb.SneakCurrent[i] = make([]float64, config.Cols)
		for j := 0; j < config.Cols; j++ {
			cb.Cells[i][j] = NewMemoryCell(MemoryFTJ)
		}
	}

	return cb
}

// ProgramWeights programs weights into the FTJ array
func (cb *FTJCrossbar) ProgramWeights(weights [][]float64) error {
	for i := 0; i < cb.Config.Rows && i < len(weights); i++ {
		for j := 0; j < cb.Config.Cols && j < len(weights[i]); j++ {
			// Normalize weight to [0, 1] conductance
			normalized := (weights[i][j] + 1) / 2 // Assume [-1, 1] weights

			err := cb.Cells[i][j].Write(normalized)
			if err != nil {
				cb.Stats.ProgramErrors++
			}
			cb.Stats.TotalWrites++
		}
	}
	return nil
}

// MVM performs matrix-vector multiply with FTJ characteristics
func (cb *FTJCrossbar) MVM(input []float64) []float64 {
	output := make([]float64, cb.Config.Cols)

	for j := 0; j < cb.Config.Cols; j++ {
		sum := 0.0
		sneakSum := 0.0

		for i := 0; i < cb.Config.Rows && i < len(input); i++ {
			// Read cell conductance
			cond := cb.Cells[i][j].Read()
			cb.Stats.TotalReads++

			// Apply I-V nonlinearity for sneak current suppression
			effectiveCond := cond
			if cb.Config.HalfBiasSelect {
				// Half-bias reduces sneak current by nonlinearity factor
				sneakFactor := 1.0 / cb.Config.Nonlinearity
				sneakSum += cond * sneakFactor * 0.5 // Unselected cells contribute less
			}

			// Convert conductance back to weight
			weight := effectiveCond*2 - 1
			sum += input[i] * weight
		}

		// Subtract estimated sneak current
		output[j] = sum - sneakSum*0.1 // Residual sneak current

		if sneakSum > 0.1*math.Abs(sum) {
			cb.Stats.SneakErrors++
		}
	}

	return output
}

// =============================================================================
// PHASE CHANGE MEMORY (PCM) WITH DRIFT COMPENSATION
// =============================================================================

// PCMCrossbarConfig configures PCM crossbar
type PCMCrossbarConfig struct {
	Rows              int
	Cols              int
	DriftCompensation bool    // Enable drift compensation
	DriftModel        string  // "log" or "power"
	RefreshInterval   float64 // Hours between refresh
}

// PCMCrossbar implements PCM-based crossbar with drift handling
type PCMCrossbar struct {
	Config         *PCMCrossbarConfig
	Cells          [][]*MemoryCell
	ReferenceCell  *MemoryCell // For drift tracking
	DriftEstimate  float64
	LastRefresh    float64
	Stats          *PCMStats
}

// PCMStats tracks PCM statistics
type PCMStats struct {
	TotalReads      int64
	TotalWrites     int64
	DriftErrors     int
	RefreshCount    int
	AccuracyLoss    float64
}

// NewPCMCrossbar creates a PCM crossbar array
func NewPCMCrossbar(config *PCMCrossbarConfig) *PCMCrossbar {
	cb := &PCMCrossbar{
		Config:        config,
		Cells:         make([][]*MemoryCell, config.Rows),
		ReferenceCell: NewMemoryCell(MemoryPCM),
		Stats:         &PCMStats{},
	}

	for i := 0; i < config.Rows; i++ {
		cb.Cells[i] = make([]*MemoryCell, config.Cols)
		for j := 0; j < config.Cols; j++ {
			cb.Cells[i][j] = NewMemoryCell(MemoryPCM)
		}
	}

	// Initialize reference cell to known state
	cb.ReferenceCell.Write(0.5)

	return cb
}

// MVM with drift compensation
func (cb *PCMCrossbar) MVM(input []float64, elapsedHours float64) []float64 {
	output := make([]float64, cb.Config.Cols)

	// Estimate drift from reference cell
	if cb.Config.DriftCompensation {
		cb.estimateDrift(elapsedHours)
	}

	for j := 0; j < cb.Config.Cols; j++ {
		sum := 0.0
		for i := 0; i < cb.Config.Rows && i < len(input); i++ {
			cond := cb.Cells[i][j].Read()

			// Apply drift compensation
			if cb.Config.DriftCompensation {
				cond = cb.compensateDrift(cond)
			}

			weight := cond*2 - 1
			sum += input[i] * weight
			cb.Stats.TotalReads++
		}
		output[j] = sum
	}

	return output
}

func (cb *PCMCrossbar) estimateDrift(elapsedHours float64) {
	// PCM drift follows: R(t) = R0 * (t/t0)^v where v ≈ 0.05-0.1
	decades := elapsedHours / (24 * 365 * 10) // Convert to decades
	driftExponent := 0.07                      // Typical drift coefficient

	// Read reference cell
	refCond := cb.ReferenceCell.Read()
	expectedCond := 0.5 * math.Pow(1+decades, -driftExponent)

	cb.DriftEstimate = refCond / expectedCond
}

func (cb *PCMCrossbar) compensateDrift(cond float64) float64 {
	if cb.DriftEstimate > 0 {
		return cond / cb.DriftEstimate
	}
	return cond
}

// Refresh reprograms cells to counteract drift
func (cb *PCMCrossbar) Refresh() {
	for i := 0; i < cb.Config.Rows; i++ {
		for j := 0; j < cb.Config.Cols; j++ {
			target := cb.Cells[i][j].TargetCond
			cb.Cells[i][j].Write(target)
		}
	}
	cb.ReferenceCell.Write(0.5)
	cb.Stats.RefreshCount++
}

// =============================================================================
// ANALOG GRADIENT DESCENT HARDWARE
// =============================================================================

// AnalogSGDConfig configures analog stochastic gradient descent
type AnalogSGDConfig struct {
	LearningRate      float64
	BatchSize         int
	ProgressiveUpdate bool    // Layer-by-layer progressive update
	ParallelUpdate    bool    // Outer product parallel update
	PulseWidth        float64 // Programming pulse width (ns)
	PulseAmplitude    float64 // Programming pulse amplitude (V)
	UpdatePrecision   int     // Bits for update
}

// AnalogSGD implements hardware analog gradient descent
type AnalogSGD struct {
	Config      *AnalogSGDConfig
	WeightArray *FTJCrossbar // Or any NVM crossbar
	GradBuffer  [][]float64  // Gradient accumulator
	Stats       *AnalogSGDStats
}

// AnalogSGDStats tracks training statistics
type AnalogSGDStats struct {
	Iterations      int
	TotalUpdates    int64
	UpdateEnergy    float64 // Total energy in pJ
	ConvergenceRate float64
	WeightNoise     float64
}

// NewAnalogSGD creates an analog SGD trainer
func NewAnalogSGD(config *AnalogSGDConfig, array *FTJCrossbar) *AnalogSGD {
	return &AnalogSGD{
		Config:      config,
		WeightArray: array,
		GradBuffer:  make([][]float64, array.Config.Rows),
		Stats:       &AnalogSGDStats{},
	}
}

// OuterProductUpdate performs parallel weight update
func (sgd *AnalogSGD) OuterProductUpdate(activation []float64, error []float64) {
	// Outer product: deltaW = -lr * activation * error^T
	rows := len(activation)
	cols := len(error)

	if rows > sgd.WeightArray.Config.Rows {
		rows = sgd.WeightArray.Config.Rows
	}
	if cols > sgd.WeightArray.Config.Cols {
		cols = sgd.WeightArray.Config.Cols
	}

	// Parallel update using voltage pulses
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			// Compute gradient
			grad := -sgd.Config.LearningRate * activation[i] * error[j]

			// Quantize to pulse
			pulseLevel := sgd.quantizeGradient(grad)

			// Apply update pulse
			currentCond := sgd.WeightArray.Cells[i][j].Conductance
			newCond := currentCond + pulseLevel

			sgd.WeightArray.Cells[i][j].Write(newCond)
			sgd.Stats.TotalUpdates++

			// Track energy
			sgd.Stats.UpdateEnergy += sgd.WeightArray.Cells[i][j].Spec.WriteEnergyFJ / 1000 // Convert to pJ
		}
	}

	sgd.Stats.Iterations++
}

func (sgd *AnalogSGD) quantizeGradient(grad float64) float64 {
	// Quantize gradient to available pulse levels
	levels := math.Pow(2, float64(sgd.Config.UpdatePrecision))
	maxStep := 1.0 / float64(sgd.WeightArray.Config.Config.ConductLevels)

	// Scale and quantize
	scaled := grad / maxStep
	quantized := math.Round(scaled*levels) / levels * maxStep

	// Add stochasticity for better convergence
	if rand.Float64() < math.Abs(scaled-math.Round(scaled)) {
		if scaled > 0 {
			quantized += maxStep / levels
		} else {
			quantized -= maxStep / levels
		}
	}

	return quantized
}

// =============================================================================
// PROGRESSIVE GRADIENT DESCENT
// =============================================================================

// ProgressiveSGDConfig configures progressive (layer-by-layer) SGD
type ProgressiveSGDConfig struct {
	NumLayers         int
	LayerSizes        []int
	LearningRates     []float64 // Per-layer learning rates
	LocalLossEnabled  bool      // Use local loss functions
	BufferSize        int       // Activation buffer size
}

// ProgressiveSGD implements progressive layer-by-layer training
type ProgressiveSGD struct {
	Config      *ProgressiveSGDConfig
	Layers      []*AnalogLayer
	Activations [][]float64 // Per-layer activations
	Errors      [][]float64 // Per-layer error signals
	Stats       *ProgressiveSGDStats
}

// AnalogLayer represents a single analog CIM layer
type AnalogLayer struct {
	Index       int
	InputDim    int
	OutputDim   int
	WeightArray *FTJCrossbar
	SGD         *AnalogSGD
	BiasArray   []float64
	Activation  string // "relu", "sigmoid", "tanh"
}

// ProgressiveSGDStats tracks progressive training
type ProgressiveSGDStats struct {
	LayerUpdates  []int
	LayerLosses   []float64
	TotalEnergy   float64
	Convergence   []float64
}

// NewProgressiveSGD creates progressive SGD trainer
func NewProgressiveSGD(config *ProgressiveSGDConfig) *ProgressiveSGD {
	psgd := &ProgressiveSGD{
		Config:      config,
		Layers:      make([]*AnalogLayer, config.NumLayers),
		Activations: make([][]float64, config.NumLayers+1),
		Errors:      make([][]float64, config.NumLayers),
		Stats: &ProgressiveSGDStats{
			LayerUpdates: make([]int, config.NumLayers),
			LayerLosses:  make([]float64, config.NumLayers),
			Convergence:  make([]float64, 0),
		},
	}

	// Initialize layers
	for l := 0; l < config.NumLayers; l++ {
		inputDim := config.LayerSizes[l]
		outputDim := config.LayerSizes[l+1]

		arrayConfig := &FTJCrossbarConfig{
			Rows:           inputDim,
			Cols:           outputDim,
			ConductLevels:  64,
			Nonlinearity:   1100,
			DynamicRange:   10,
			HalfBiasSelect: true,
			ReadVoltage:    0.3,
		}

		sgdConfig := &AnalogSGDConfig{
			LearningRate:      config.LearningRates[l],
			ProgressiveUpdate: true,
			ParallelUpdate:    true,
			PulseWidth:        10,
			PulseAmplitude:    1.0,
			UpdatePrecision:   6,
		}

		array := NewFTJCrossbar(arrayConfig)
		sgd := NewAnalogSGD(sgdConfig, array)

		psgd.Layers[l] = &AnalogLayer{
			Index:       l,
			InputDim:    inputDim,
			OutputDim:   outputDim,
			WeightArray: array,
			SGD:         sgd,
			BiasArray:   make([]float64, outputDim),
			Activation:  "relu",
		}
	}

	return psgd
}

// ForwardPass performs forward propagation
func (psgd *ProgressiveSGD) ForwardPass(input []float64) []float64 {
	psgd.Activations[0] = input

	for l := 0; l < psgd.Config.NumLayers; l++ {
		layer := psgd.Layers[l]

		// MVM
		preAct := layer.WeightArray.MVM(psgd.Activations[l])

		// Add bias
		for j := range preAct {
			preAct[j] += layer.BiasArray[j]
		}

		// Activation
		psgd.Activations[l+1] = applyActivation(preAct, layer.Activation)
	}

	return psgd.Activations[psgd.Config.NumLayers]
}

// ProgressiveBackward performs progressive layer-by-layer backprop
func (psgd *ProgressiveSGD) ProgressiveBackward(target []float64) {
	numLayers := psgd.Config.NumLayers

	// Compute output error
	output := psgd.Activations[numLayers]
	outputError := make([]float64, len(output))
	for i := range output {
		if i < len(target) {
			outputError[i] = output[i] - target[i]
		}
	}
	psgd.Errors[numLayers-1] = outputError

	// Progressive update: process each layer immediately
	for l := numLayers - 1; l >= 0; l-- {
		layer := psgd.Layers[l]

		// Update weights using outer product
		layer.SGD.OuterProductUpdate(psgd.Activations[l], psgd.Errors[l])
		psgd.Stats.LayerUpdates[l]++

		// Compute error for previous layer (if not first layer)
		if l > 0 {
			// Backpropagate through weights
			psgd.Errors[l-1] = psgd.backpropError(l, psgd.Errors[l])
		}

		// Track local loss if enabled
		if psgd.Config.LocalLossEnabled {
			psgd.Stats.LayerLosses[l] = psgd.computeLocalLoss(l)
		}
	}
}

func (psgd *ProgressiveSGD) backpropError(layer int, error []float64) []float64 {
	// Transpose MVM for backprop
	l := psgd.Layers[layer]
	prevError := make([]float64, l.InputDim)

	for i := 0; i < l.InputDim; i++ {
		sum := 0.0
		for j := 0; j < l.OutputDim && j < len(error); j++ {
			// Read weight (transposed access)
			cond := l.WeightArray.Cells[i][j].Read()
			weight := cond*2 - 1
			sum += error[j] * weight
		}

		// Apply activation derivative
		act := psgd.Activations[layer][i]
		prevError[i] = sum * activationDerivative(act, l.Activation)
	}

	return prevError
}

func (psgd *ProgressiveSGD) computeLocalLoss(layer int) float64 {
	// Compute local reconstruction loss
	loss := 0.0
	for _, err := range psgd.Errors[layer] {
		loss += err * err
	}
	return loss / float64(len(psgd.Errors[layer]))
}

// =============================================================================
// MULTI-TILE RESIDUAL LEARNING
// =============================================================================

// MultiTileConfig configures multi-tile residual learning
type MultiTileConfig struct {
	NumTiles       int
	TileSize       int
	ConductLevels  int     // Limited levels per tile
	ResidualWeight float64 // Weight for residual tiles
}

// MultiTileArray implements multi-tile residual weight storage
type MultiTileArray struct {
	Config       *MultiTileConfig
	MainTile     *FTJCrossbar
	ResidualTile []*FTJCrossbar // Additional tiles for residual
	Stats        *MultiTileStats
}

// MultiTileStats tracks multi-tile statistics
type MultiTileStats struct {
	MainTileUpdates     int64
	ResidualTileUpdates int64
	EffectivePrecision  int
	StorageOverhead     float64
}

// NewMultiTileArray creates multi-tile residual array
func NewMultiTileArray(config *MultiTileConfig) *MultiTileArray {
	mta := &MultiTileArray{
		Config:       config,
		ResidualTile: make([]*FTJCrossbar, config.NumTiles-1),
		Stats:        &MultiTileStats{},
	}

	// Main tile with limited precision
	mainConfig := &FTJCrossbarConfig{
		Rows:           config.TileSize,
		Cols:           config.TileSize,
		ConductLevels:  config.ConductLevels,
		Nonlinearity:   1100,
		DynamicRange:   10,
		HalfBiasSelect: true,
	}
	mta.MainTile = NewFTJCrossbar(mainConfig)

	// Residual tiles for high precision
	for t := 0; t < config.NumTiles-1; t++ {
		mta.ResidualTile[t] = NewFTJCrossbar(mainConfig)
	}

	// Calculate effective precision
	mta.Stats.EffectivePrecision = int(math.Log2(float64(config.ConductLevels))) * config.NumTiles
	mta.Stats.StorageOverhead = float64(config.NumTiles)

	return mta
}

// ProgramWeights programs weights across multiple tiles
func (mta *MultiTileArray) ProgramWeights(weights [][]float64) {
	// Quantize to main tile
	residual := make([][]float64, len(weights))

	for i := range weights {
		residual[i] = make([]float64, len(weights[i]))
		for j := range weights[i] {
			// Quantize to main tile precision
			mainWeight := quantizeToLevels(weights[i][j], mta.Config.ConductLevels)
			mta.MainTile.Cells[i][j].Write((mainWeight + 1) / 2)

			// Compute residual
			residual[i][j] = weights[i][j] - mainWeight
		}
	}

	// Program residual into additional tiles
	for t := 0; t < len(mta.ResidualTile); t++ {
		scale := math.Pow(mta.Config.ResidualWeight, float64(t+1))
		for i := range residual {
			for j := range residual[i] {
				scaledResidual := residual[i][j] / scale
				quantized := quantizeToLevels(scaledResidual, mta.Config.ConductLevels)
				mta.ResidualTile[t].Cells[i][j].Write((quantized + 1) / 2)

				// Update residual for next tile
				residual[i][j] = scaledResidual - quantized
			}
		}
	}
}

// MVM performs MVM across all tiles
func (mta *MultiTileArray) MVM(input []float64) []float64 {
	// Main tile contribution
	output := mta.MainTile.MVM(input)

	// Add residual contributions
	for t := 0; t < len(mta.ResidualTile); t++ {
		scale := math.Pow(mta.Config.ResidualWeight, float64(t+1))
		residualOut := mta.ResidualTile[t].MVM(input)

		for j := range output {
			output[j] += residualOut[j] * scale
		}
	}

	return output
}

func quantizeToLevels(val float64, levels int) float64 {
	step := 2.0 / float64(levels-1)
	quantized := math.Round(val/step) * step
	return math.Max(-1, math.Min(1, quantized))
}

// =============================================================================
// BENCHMARK AND COMPARISON
// =============================================================================

// EmergingMemoryBenchmark compares different memory technologies
type EmergingMemoryBenchmark struct {
	Technologies map[MemoryTechnology]*MemoryTechSpec
	TestArrays   map[MemoryTechnology]interface{}
	Results      *BenchmarkResult
}

// BenchmarkResult stores comparison results
type BenchmarkResult struct {
	EnergyEfficiency map[MemoryTechnology]float64 // TOPS/W
	AreaEfficiency   map[MemoryTechnology]float64 // TOPS/mm²
	Accuracy         map[MemoryTechnology]float64 // Inference accuracy
	TrainingSupport  map[MemoryTechnology]bool
	DriftTolerance   map[MemoryTechnology]float64
}

// NewEmergingMemoryBenchmark creates benchmark suite
func NewEmergingMemoryBenchmark() *EmergingMemoryBenchmark {
	return &EmergingMemoryBenchmark{
		Technologies: GetMemorySpecs(),
		TestArrays:   make(map[MemoryTechnology]interface{}),
		Results: &BenchmarkResult{
			EnergyEfficiency: make(map[MemoryTechnology]float64),
			AreaEfficiency:   make(map[MemoryTechnology]float64),
			Accuracy:         make(map[MemoryTechnology]float64),
			TrainingSupport:  make(map[MemoryTechnology]bool),
			DriftTolerance:   make(map[MemoryTechnology]float64),
		},
	}
}

// RunComparison runs technology comparison
func (emb *EmergingMemoryBenchmark) RunComparison() {
	for tech, spec := range emb.Technologies {
		// Energy efficiency (TOPS/W)
		// TOPS = 2 * N² operations / latency, W = energy/latency
		opsPerCycle := 2.0 * 128 * 128 // 128x128 array
		latency := spec.SwitchingTimeNs * 1e-9
		energy := (spec.ReadEnergyFJ * 128 * 128) * 1e-15 // Total read energy

		emb.Results.EnergyEfficiency[tech] = (opsPerCycle / latency) / (energy / latency) / 1e12

		// Area efficiency
		cellArea := float64(spec.CellSizeF2) * 100 // Assume 10nm node
		arrayArea := 128 * 128 * cellArea * 1e-12  // mm²
		emb.Results.AreaEfficiency[tech] = (opsPerCycle / latency) / arrayArea / 1e12

		// Accuracy based on precision
		bits := math.Log2(float64(spec.ConductLevels))
		emb.Results.Accuracy[tech] = 1.0 - 0.01*(8-bits) // Approximate accuracy loss

		// Training support
		emb.Results.TrainingSupport[tech] = spec.EnduranceCycles >= 1e8

		// Drift tolerance
		emb.Results.DriftTolerance[tech] = 1.0 - spec.DriftRate*10 // Over 10 decades
	}
}

// GenerateReport creates comparison report
func (emb *EmergingMemoryBenchmark) GenerateReport() string {
	report := "=== Emerging Memory Technology Comparison ===\n\n"

	report += "Technology Specifications:\n"
	report += fmt.Sprintf("%-15s %10s %10s %10s %10s %10s\n",
		"Technology", "Write(fJ)", "Switch(ns)", "Endurance", "On/Off", "MLC Bits")
	report += "-------------------------------------------------------------------\n"

	for _, spec := range emb.Technologies {
		bits := int(math.Log2(float64(spec.ConductLevels)))
		report += fmt.Sprintf("%-15s %10.0f %10.0f %10.0e %10.0f %10d\n",
			spec.Name, spec.WriteEnergyFJ, spec.SwitchingTimeNs,
			spec.EnduranceCycles, spec.OnOffRatio, bits)
	}

	report += "\nPerformance Metrics:\n"
	report += fmt.Sprintf("%-15s %12s %12s %10s %10s\n",
		"Technology", "TOPS/W", "TOPS/mm²", "Accuracy", "Training")
	report += "-------------------------------------------------------------------\n"

	for tech, spec := range emb.Technologies {
		trainStr := "No"
		if emb.Results.TrainingSupport[tech] {
			trainStr = "Yes"
		}
		report += fmt.Sprintf("%-15s %12.1f %12.1f %10.2f%% %10s\n",
			spec.Name, emb.Results.EnergyEfficiency[tech],
			emb.Results.AreaEfficiency[tech],
			emb.Results.Accuracy[tech]*100, trainStr)
	}

	return report
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func applyActivation(x []float64, activation string) []float64 {
	result := make([]float64, len(x))
	for i, v := range x {
		switch activation {
		case "relu":
			result[i] = math.Max(0, v)
		case "sigmoid":
			result[i] = 1.0 / (1.0 + math.Exp(-v))
		case "tanh":
			result[i] = math.Tanh(v)
		default:
			result[i] = v
		}
	}
	return result
}

func activationDerivative(x float64, activation string) float64 {
	switch activation {
	case "relu":
		if x > 0 {
			return 1
		}
		return 0
	case "sigmoid":
		s := 1.0 / (1.0 + math.Exp(-x))
		return s * (1 - s)
	case "tanh":
		t := math.Tanh(x)
		return 1 - t*t
	default:
		return 1
	}
}
