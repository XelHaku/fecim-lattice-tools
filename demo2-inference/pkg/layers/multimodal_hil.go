// Package layers provides neural network layer implementations for CIM deployment.
// This file implements multimodal sensor fusion and hardware-in-the-loop (HIL)
// simulation for CIM accelerator validation.
// Based on research: Nano Letters (multimodal FeFET), Frontiers (NeuroSim),
// Nature (cross-modal neuromorphic), PMC (multimodal synapses)
package layers

import (
	"math"
	"math/rand"
)

// =============================================================================
// MULTIMODAL SENSORY FUSION
// Unified processing of visual, auditory, tactile, and olfactory signals
// =============================================================================

// ModalityType defines types of sensory modalities.
type ModalityType int

const (
	ModalityVisual   ModalityType = iota // Image/video
	ModalityAuditory                     // Audio/speech
	ModalityTactile                      // Touch/pressure
	ModalityOlfactory                    // Smell/gas
	ModalityThermal                      // Temperature
	ModalityProprioception               // Body position
)

// SensorInput represents input from a single sensor modality.
type SensorInput struct {
	Modality  ModalityType
	Data      []float64
	Timestamp float64 // ms
	Spikes    []float64 // Spike times if event-based
	Confidence float64
}

// MultimodalSynapseConfig configures a multimodal artificial synapse.
type MultimodalSynapseConfig struct {
	// Supported modalities
	EnabledModalities []ModalityType

	// Synaptic parameters
	UseFerroelectric  bool
	PolarizationMax   float64
	STPTimeConstant   float64 // Short-term plasticity (ms)
	LTPThreshold      float64 // Long-term plasticity threshold

	// Fusion parameters
	FusionMethod     string // "early", "late", "hierarchical"
	AttentionWeights map[ModalityType]float64

	// Energy constraints
	MaxEnergyPJ float64
}

// DefaultMultimodalConfig returns default multimodal config.
func DefaultMultimodalConfig() *MultimodalSynapseConfig {
	return &MultimodalSynapseConfig{
		EnabledModalities: []ModalityType{
			ModalityVisual, ModalityAuditory, ModalityTactile,
		},
		UseFerroelectric: true,
		PolarizationMax:  50.0,
		STPTimeConstant:  100.0,
		LTPThreshold:     0.5,
		FusionMethod:     "hierarchical",
		AttentionWeights: map[ModalityType]float64{
			ModalityVisual:   0.4,
			ModalityAuditory: 0.3,
			ModalityTactile:  0.2,
			ModalityOlfactory: 0.1,
		},
		MaxEnergyPJ: 100.0,
	}
}

// MultimodalSynapse implements a multimodal artificial synapse.
type MultimodalSynapse struct {
	Config *MultimodalSynapseConfig

	// Per-modality state
	ModalityWeights map[ModalityType][]float64
	ModalitySTM     map[ModalityType]float64 // Short-term memory
	ModalityLTM     map[ModalityType]float64 // Long-term memory

	// Ferroelectric state
	Polarization    float64
	PolarizationHistory []float64

	// Plasticity state
	LastSpikeTime   map[ModalityType]float64
	PPFState        map[ModalityType]float64 // Paired-pulse facilitation

	// Statistics
	TotalEnergy     float64
	ProcessedEvents int64
}

// NewMultimodalSynapse creates a new multimodal synapse.
func NewMultimodalSynapse(config *MultimodalSynapseConfig) *MultimodalSynapse {
	if config == nil {
		config = DefaultMultimodalConfig()
	}

	syn := &MultimodalSynapse{
		Config:          config,
		ModalityWeights: make(map[ModalityType][]float64),
		ModalitySTM:     make(map[ModalityType]float64),
		ModalityLTM:     make(map[ModalityType]float64),
		LastSpikeTime:   make(map[ModalityType]float64),
		PPFState:        make(map[ModalityType]float64),
		PolarizationHistory: make([]float64, 0),
	}

	// Initialize weights for each modality
	for _, mod := range config.EnabledModalities {
		syn.ModalityWeights[mod] = make([]float64, 64) // Default 64 features
		for i := range syn.ModalityWeights[mod] {
			syn.ModalityWeights[mod][i] = rand.Float64()*0.2 - 0.1
		}
	}

	return syn
}

// ProcessInput processes input from a single modality.
func (ms *MultimodalSynapse) ProcessInput(input *SensorInput) *SynapticResponse {
	if _, enabled := ms.ModalityWeights[input.Modality]; !enabled {
		return nil
	}

	// Compute weighted response
	weights := ms.ModalityWeights[input.Modality]
	response := 0.0
	for i := 0; i < len(input.Data) && i < len(weights); i++ {
		response += input.Data[i] * weights[i]
	}

	// Apply short-term plasticity
	dt := input.Timestamp - ms.LastSpikeTime[input.Modality]
	if dt > 0 && dt < ms.Config.STPTimeConstant*5 {
		// Paired-pulse facilitation
		ppf := math.Exp(-dt / ms.Config.STPTimeConstant)
		ms.PPFState[input.Modality] = ppf
		response *= (1 + ppf)
	}

	// Update STM (exponential decay toward LTM)
	stmDecay := math.Exp(-dt / ms.Config.STPTimeConstant)
	ms.ModalitySTM[input.Modality] = ms.ModalitySTM[input.Modality]*stmDecay + response*(1-stmDecay)

	// Check for LTP transition
	if math.Abs(response) > ms.Config.LTPThreshold {
		// Update LTM
		ltpRate := 0.01
		ms.ModalityLTM[input.Modality] += ltpRate * (response - ms.ModalityLTM[input.Modality])
	}

	// Ferroelectric modulation
	if ms.Config.UseFerroelectric {
		// Polarization follows strong inputs
		if math.Abs(response) > ms.Config.LTPThreshold {
			targetPol := response * ms.Config.PolarizationMax / 10.0
			ms.Polarization += 0.1 * (targetPol - ms.Polarization)
		}
		ms.PolarizationHistory = append(ms.PolarizationHistory, ms.Polarization)
	}

	// Update timing
	ms.LastSpikeTime[input.Modality] = input.Timestamp
	ms.ProcessedEvents++

	// Estimate energy (simplified)
	energy := 1.0 // pJ base
	if ms.Config.UseFerroelectric && math.Abs(response) > ms.Config.LTPThreshold {
		energy += 10.0 // Ferroelectric switching energy
	}
	ms.TotalEnergy += energy

	return &SynapticResponse{
		Modality:     input.Modality,
		Response:     response,
		STMState:     ms.ModalitySTM[input.Modality],
		LTMState:     ms.ModalityLTM[input.Modality],
		PPFState:     ms.PPFState[input.Modality],
		Polarization: ms.Polarization,
		EnergyPJ:     energy,
	}
}

// SynapticResponse contains synapse output.
type SynapticResponse struct {
	Modality     ModalityType
	Response     float64
	STMState     float64
	LTMState     float64
	PPFState     float64
	Polarization float64
	EnergyPJ     float64
}

// =============================================================================
// MULTIMODAL FUSION NETWORK
// =============================================================================

// FusionNetworkConfig configures multimodal fusion network.
type FusionNetworkConfig struct {
	// Architecture
	ModalityEncoderSize int
	FusionLayerSize     int
	OutputSize          int

	// Fusion strategy
	FusionMethod string // "early", "late", "attention", "reservoir"

	// Reservoir computing
	UseReservoir     bool
	ReservoirSize    int
	ReservoirSparsity float64

	// CIM deployment
	CrossbarSize int
	WeightBits   int
}

// DefaultFusionConfig returns default fusion config.
func DefaultFusionConfig() *FusionNetworkConfig {
	return &FusionNetworkConfig{
		ModalityEncoderSize: 64,
		FusionLayerSize:     128,
		OutputSize:          10,
		FusionMethod:        "attention",
		UseReservoir:        true,
		ReservoirSize:       256,
		ReservoirSparsity:   0.1,
		CrossbarSize:        64,
		WeightBits:          6,
	}
}

// MultimodalFusionNetwork implements multimodal fusion.
type MultimodalFusionNetwork struct {
	Config *FusionNetworkConfig

	// Per-modality encoders
	Encoders map[ModalityType]*ModalityEncoder

	// Fusion layer
	FusionWeights [][]float64
	FusionBias    []float64

	// Attention mechanism
	AttentionQuery  [][]float64
	AttentionKey    [][]float64
	AttentionValue  [][]float64

	// Reservoir (if enabled)
	ReservoirWeights [][]float64
	ReservoirState   []float64

	// Output layer
	OutputWeights [][]float64
	OutputBias    []float64

	// Statistics
	InferenceCount int64
	TotalLatencyUs float64
}

// ModalityEncoder encodes single modality.
type ModalityEncoder struct {
	Modality    ModalityType
	InputSize   int
	OutputSize  int
	Weights     [][]float64
	Bias        []float64
}

// NewMultimodalFusionNetwork creates a fusion network.
func NewMultimodalFusionNetwork(modalities []ModalityType, inputSizes map[ModalityType]int, config *FusionNetworkConfig) *MultimodalFusionNetwork {
	if config == nil {
		config = DefaultFusionConfig()
	}

	net := &MultimodalFusionNetwork{
		Config:   config,
		Encoders: make(map[ModalityType]*ModalityEncoder),
	}

	// Create per-modality encoders
	for _, mod := range modalities {
		inSize := inputSizes[mod]
		if inSize == 0 {
			inSize = 64 // Default
		}

		encoder := &ModalityEncoder{
			Modality:   mod,
			InputSize:  inSize,
			OutputSize: config.ModalityEncoderSize,
			Weights:    make([][]float64, inSize),
			Bias:       make([]float64, config.ModalityEncoderSize),
		}

		scale := math.Sqrt(2.0 / float64(inSize))
		for i := range encoder.Weights {
			encoder.Weights[i] = make([]float64, config.ModalityEncoderSize)
			for j := range encoder.Weights[i] {
				encoder.Weights[i][j] = rand.NormFloat64() * scale
			}
		}
		net.Encoders[mod] = encoder
	}

	// Fusion layer
	fusionInputSize := len(modalities) * config.ModalityEncoderSize
	net.FusionWeights = make([][]float64, fusionInputSize)
	scale := math.Sqrt(2.0 / float64(fusionInputSize))
	for i := range net.FusionWeights {
		net.FusionWeights[i] = make([]float64, config.FusionLayerSize)
		for j := range net.FusionWeights[i] {
			net.FusionWeights[i][j] = rand.NormFloat64() * scale
		}
	}
	net.FusionBias = make([]float64, config.FusionLayerSize)

	// Attention weights
	if config.FusionMethod == "attention" {
		attDim := config.ModalityEncoderSize
		net.AttentionQuery = makeRandomMatrix(attDim, attDim, scale)
		net.AttentionKey = makeRandomMatrix(attDim, attDim, scale)
		net.AttentionValue = makeRandomMatrix(attDim, attDim, scale)
	}

	// Reservoir
	if config.UseReservoir {
		net.ReservoirWeights = make([][]float64, config.ReservoirSize)
		for i := range net.ReservoirWeights {
			net.ReservoirWeights[i] = make([]float64, config.ReservoirSize)
			for j := range net.ReservoirWeights[i] {
				if rand.Float64() < config.ReservoirSparsity {
					net.ReservoirWeights[i][j] = rand.NormFloat64() * 0.1
				}
			}
		}
		net.ReservoirState = make([]float64, config.ReservoirSize)
	}

	// Output layer
	outInSize := config.FusionLayerSize
	if config.UseReservoir {
		outInSize = config.ReservoirSize
	}
	net.OutputWeights = make([][]float64, outInSize)
	scale = math.Sqrt(2.0 / float64(outInSize))
	for i := range net.OutputWeights {
		net.OutputWeights[i] = make([]float64, config.OutputSize)
		for j := range net.OutputWeights[i] {
			net.OutputWeights[i][j] = rand.NormFloat64() * scale
		}
	}
	net.OutputBias = make([]float64, config.OutputSize)

	return net
}

// makeRandomMatrix creates a random matrix.
func makeRandomMatrix(rows, cols int, scale float64) [][]float64 {
	m := make([][]float64, rows)
	for i := range m {
		m[i] = make([]float64, cols)
		for j := range m[i] {
			m[i][j] = rand.NormFloat64() * scale
		}
	}
	return m
}

// Forward performs multimodal forward pass.
func (net *MultimodalFusionNetwork) Forward(inputs map[ModalityType][]float64) *FusionOutput {
	net.InferenceCount++

	// Encode each modality
	encodings := make(map[ModalityType][]float64)
	for mod, encoder := range net.Encoders {
		if data, ok := inputs[mod]; ok {
			encoding := make([]float64, encoder.OutputSize)
			for j := 0; j < encoder.OutputSize; j++ {
				sum := encoder.Bias[j]
				for i := 0; i < len(data) && i < encoder.InputSize; i++ {
					sum += data[i] * encoder.Weights[i][j]
				}
				// ReLU
				if sum > 0 {
					encoding[j] = sum
				}
			}
			encodings[mod] = encoding
		}
	}

	// Fuse modalities
	var fused []float64
	switch net.Config.FusionMethod {
	case "early":
		fused = net.earlyFusion(encodings)
	case "attention":
		fused = net.attentionFusion(encodings)
	default:
		fused = net.earlyFusion(encodings)
	}

	// Apply fusion layer
	hidden := make([]float64, net.Config.FusionLayerSize)
	for j := 0; j < net.Config.FusionLayerSize; j++ {
		sum := net.FusionBias[j]
		for i := 0; i < len(fused) && i < len(net.FusionWeights); i++ {
			sum += fused[i] * net.FusionWeights[i][j]
		}
		if sum > 0 {
			hidden[j] = sum
		}
	}

	// Reservoir processing
	if net.Config.UseReservoir {
		hidden = net.reservoirProcess(hidden)
	}

	// Output layer
	output := make([]float64, net.Config.OutputSize)
	for j := 0; j < net.Config.OutputSize; j++ {
		sum := net.OutputBias[j]
		for i := 0; i < len(hidden) && i < len(net.OutputWeights); i++ {
			sum += hidden[i] * net.OutputWeights[i][j]
		}
		output[j] = sum
	}

	// Softmax
	output = softmax(output)

	// Find prediction
	predClass := 0
	maxProb := output[0]
	for i, p := range output {
		if p > maxProb {
			maxProb = p
			predClass = i
		}
	}

	return &FusionOutput{
		ModalityEncodings: encodings,
		FusedRepresentation: fused,
		Output:            output,
		PredictedClass:    predClass,
		Confidence:        maxProb,
	}
}

// earlyFusion concatenates all modality encodings.
func (net *MultimodalFusionNetwork) earlyFusion(encodings map[ModalityType][]float64) []float64 {
	totalSize := 0
	for _, enc := range encodings {
		totalSize += len(enc)
	}

	fused := make([]float64, 0, totalSize)
	for _, enc := range encodings {
		fused = append(fused, enc...)
	}
	return fused
}

// attentionFusion uses attention mechanism.
func (net *MultimodalFusionNetwork) attentionFusion(encodings map[ModalityType][]float64) []float64 {
	// Stack encodings
	encList := make([][]float64, 0, len(encodings))
	for _, enc := range encodings {
		encList = append(encList, enc)
	}

	if len(encList) == 0 {
		return nil
	}

	dim := len(encList[0])
	numMod := len(encList)

	// Compute queries, keys, values
	queries := make([][]float64, numMod)
	keys := make([][]float64, numMod)
	values := make([][]float64, numMod)

	for m := 0; m < numMod; m++ {
		queries[m] = make([]float64, dim)
		keys[m] = make([]float64, dim)
		values[m] = make([]float64, dim)

		for d := 0; d < dim; d++ {
			for i := 0; i < dim && i < len(encList[m]); i++ {
				if d < len(net.AttentionQuery) && i < len(net.AttentionQuery[d]) {
					queries[m][d] += encList[m][i] * net.AttentionQuery[i][d]
				}
				if d < len(net.AttentionKey) && i < len(net.AttentionKey[d]) {
					keys[m][d] += encList[m][i] * net.AttentionKey[i][d]
				}
				if d < len(net.AttentionValue) && i < len(net.AttentionValue[d]) {
					values[m][d] += encList[m][i] * net.AttentionValue[i][d]
				}
			}
		}
	}

	// Compute attention scores
	scores := make([][]float64, numMod)
	scale := math.Sqrt(float64(dim))
	for i := 0; i < numMod; i++ {
		scores[i] = make([]float64, numMod)
		for j := 0; j < numMod; j++ {
			dot := 0.0
			for d := 0; d < dim; d++ {
				dot += queries[i][d] * keys[j][d]
			}
			scores[i][j] = dot / scale
		}
		// Softmax over j
		scores[i] = softmax(scores[i])
	}

	// Weighted sum of values
	output := make([]float64, dim*numMod)
	for i := 0; i < numMod; i++ {
		for d := 0; d < dim; d++ {
			sum := 0.0
			for j := 0; j < numMod; j++ {
				sum += scores[i][j] * values[j][d]
			}
			output[i*dim+d] = sum
		}
	}

	return output
}

// reservoirProcess applies reservoir dynamics.
func (net *MultimodalFusionNetwork) reservoirProcess(input []float64) []float64 {
	// Input to reservoir
	inputSize := len(input)
	resSize := net.Config.ReservoirSize

	// Update reservoir state
	newState := make([]float64, resSize)
	for i := 0; i < resSize; i++ {
		// Input contribution
		if i < inputSize {
			newState[i] += input[i] * 0.1
		}

		// Recurrent contribution
		for j := 0; j < resSize; j++ {
			newState[i] += net.ReservoirState[j] * net.ReservoirWeights[j][i]
		}

		// Tanh activation
		newState[i] = math.Tanh(newState[i])
	}

	net.ReservoirState = newState
	return newState
}

// softmax applies softmax normalization.
func softmax(x []float64) []float64 {
	if len(x) == 0 {
		return x
	}

	// Find max for numerical stability
	maxVal := x[0]
	for _, v := range x {
		if v > maxVal {
			maxVal = v
		}
	}

	// Compute exp and sum
	result := make([]float64, len(x))
	sum := 0.0
	for i, v := range x {
		result[i] = math.Exp(v - maxVal)
		sum += result[i]
	}

	// Normalize
	for i := range result {
		result[i] /= sum
	}

	return result
}

// FusionOutput contains fusion network output.
type FusionOutput struct {
	ModalityEncodings   map[ModalityType][]float64
	FusedRepresentation []float64
	Output              []float64
	PredictedClass      int
	Confidence          float64
}

// =============================================================================
// HARDWARE-IN-THE-LOOP (HIL) SIMULATION
// Validation framework for CIM accelerators
// =============================================================================

// HILConfig configures hardware-in-the-loop simulation.
type HILConfig struct {
	// Simulation parameters
	SimulationStepUs float64 // Time step (µs)
	RealTimeFactor   float64 // 1.0 = real-time, <1 = slower

	// Hardware model
	CrossbarRows     int
	CrossbarCols     int
	ADCBits          int
	DACBits          int

	// Non-idealities
	EnableNoise      bool
	NoiseSigma       float64
	EnableVariation  bool
	VariationSigma   float64
	EnableDrift      bool
	DriftRate        float64

	// Validation
	ValidationMode   string // "functional", "timing", "power"
	GoldenModel      bool   // Compare against ideal
}

// DefaultHILConfig returns default HIL config.
func DefaultHILConfig() *HILConfig {
	return &HILConfig{
		SimulationStepUs: 1.0,
		RealTimeFactor:   1.0,
		CrossbarRows:     64,
		CrossbarCols:     64,
		ADCBits:          6,
		DACBits:          8,
		EnableNoise:      true,
		NoiseSigma:       0.02,
		EnableVariation:  true,
		VariationSigma:   0.05,
		EnableDrift:      true,
		DriftRate:        0.001,
		ValidationMode:   "functional",
		GoldenModel:      true,
	}
}

// CIMHardwareModel models CIM hardware for HIL.
type CIMHardwareModel struct {
	Config *HILConfig

	// Crossbar state
	Conductances     [][]float64
	IdealConductances [][]float64 // For golden comparison
	Variation        [][]float64  // Device variation

	// ADC/DAC models
	DACLevels int
	ADCLevels int

	// Timing model
	MVMLatencyNs     float64
	ADCLatencyNs     float64
	ProgramLatencyNs float64

	// Power model
	ReadPowerUW      float64
	WritePowerUW     float64
	LeakagePowerUW   float64

	// Statistics
	TotalMVMOps      int64
	TotalReadOps     int64
	TotalWriteOps    int64
	CumulativeError  float64
}

// NewCIMHardwareModel creates a CIM hardware model.
func NewCIMHardwareModel(config *HILConfig) *CIMHardwareModel {
	if config == nil {
		config = DefaultHILConfig()
	}

	model := &CIMHardwareModel{
		Config:      config,
		DACLevels:   1 << config.DACBits,
		ADCLevels:   1 << config.ADCBits,
		MVMLatencyNs: 100.0,
		ADCLatencyNs: 50.0,
		ProgramLatencyNs: 1000.0,
		ReadPowerUW:  10.0,
		WritePowerUW: 100.0,
		LeakagePowerUW: 1.0,
	}

	// Initialize conductances
	model.Conductances = make([][]float64, config.CrossbarRows)
	model.IdealConductances = make([][]float64, config.CrossbarRows)
	model.Variation = make([][]float64, config.CrossbarRows)

	for i := 0; i < config.CrossbarRows; i++ {
		model.Conductances[i] = make([]float64, config.CrossbarCols)
		model.IdealConductances[i] = make([]float64, config.CrossbarCols)
		model.Variation[i] = make([]float64, config.CrossbarCols)

		for j := 0; j < config.CrossbarCols; j++ {
			// Initialize variation
			if config.EnableVariation {
				model.Variation[i][j] = config.VariationSigma * rand.NormFloat64()
			}
		}
	}

	return model
}

// ProgramWeights programs weights to crossbar with non-idealities.
func (hw *CIMHardwareModel) ProgramWeights(weights [][]float64) {
	for i := 0; i < len(weights) && i < hw.Config.CrossbarRows; i++ {
		for j := 0; j < len(weights[i]) && j < hw.Config.CrossbarCols; j++ {
			// Store ideal
			hw.IdealConductances[i][j] = weights[i][j]

			// Apply variation
			programmed := weights[i][j]
			if hw.Config.EnableVariation {
				programmed *= (1 + hw.Variation[i][j])
			}

			// Add programming noise
			if hw.Config.EnableNoise {
				programmed += hw.Config.NoiseSigma * rand.NormFloat64()
			}

			hw.Conductances[i][j] = programmed
			hw.TotalWriteOps++
		}
	}
}

// MVM performs matrix-vector multiplication with non-idealities.
func (hw *CIMHardwareModel) MVM(input []float64) ([]float64, *MVMResult) {
	hw.TotalMVMOps++
	hw.TotalReadOps += int64(len(input))

	// DAC quantization
	dacQuantized := make([]float64, len(input))
	for i := range input {
		level := int(input[i] * float64(hw.DACLevels-1))
		if level < 0 {
			level = 0
		}
		if level >= hw.DACLevels {
			level = hw.DACLevels - 1
		}
		dacQuantized[i] = float64(level) / float64(hw.DACLevels-1)
	}

	// Ideal MVM (golden model)
	idealOutput := make([]float64, hw.Config.CrossbarCols)
	for j := 0; j < hw.Config.CrossbarCols; j++ {
		for i := 0; i < len(dacQuantized) && i < hw.Config.CrossbarRows; i++ {
			idealOutput[j] += dacQuantized[i] * hw.IdealConductances[i][j]
		}
	}

	// Actual MVM with non-idealities
	actualOutput := make([]float64, hw.Config.CrossbarCols)
	for j := 0; j < hw.Config.CrossbarCols; j++ {
		sum := 0.0
		for i := 0; i < len(dacQuantized) && i < hw.Config.CrossbarRows; i++ {
			// Use actual (non-ideal) conductances
			sum += dacQuantized[i] * hw.Conductances[i][j]
		}

		// Add read noise
		if hw.Config.EnableNoise {
			sum += hw.Config.NoiseSigma * rand.NormFloat64() * 0.1
		}

		// ADC quantization
		adcLevel := int(sum * float64(hw.ADCLevels-1))
		if adcLevel < 0 {
			adcLevel = 0
		}
		if adcLevel >= hw.ADCLevels {
			adcLevel = hw.ADCLevels - 1
		}
		actualOutput[j] = float64(adcLevel) / float64(hw.ADCLevels-1)
	}

	// Apply drift if enabled
	if hw.Config.EnableDrift {
		hw.applyDrift()
	}

	// Compute error metrics
	mse := 0.0
	maxError := 0.0
	for j := range actualOutput {
		err := math.Abs(actualOutput[j] - idealOutput[j])
		mse += err * err
		if err > maxError {
			maxError = err
		}
	}
	mse /= float64(len(actualOutput))
	hw.CumulativeError += mse

	return actualOutput, &MVMResult{
		IdealOutput:  idealOutput,
		ActualOutput: actualOutput,
		MSE:          mse,
		MaxError:     maxError,
		LatencyNs:    hw.MVMLatencyNs + hw.ADCLatencyNs,
		EnergyPJ:     hw.ReadPowerUW * (hw.MVMLatencyNs + hw.ADCLatencyNs) / 1000,
	}
}

// applyDrift applies conductance drift.
func (hw *CIMHardwareModel) applyDrift() {
	for i := 0; i < hw.Config.CrossbarRows; i++ {
		for j := 0; j < hw.Config.CrossbarCols; j++ {
			drift := hw.Config.DriftRate * rand.NormFloat64()
			hw.Conductances[i][j] *= (1 + drift)
		}
	}
}

// MVMResult contains MVM operation results.
type MVMResult struct {
	IdealOutput  []float64
	ActualOutput []float64
	MSE          float64
	MaxError     float64
	LatencyNs    float64
	EnergyPJ     float64
}

// =============================================================================
// HIL SIMULATION FRAMEWORK
// =============================================================================

// HILSimulator manages HIL simulation.
type HILSimulator struct {
	Config       *HILConfig
	HardwareModel *CIMHardwareModel

	// Simulation state
	CurrentTimeUs float64
	StepCount     int64

	// Test vectors
	TestInputs   [][]float64
	ExpectedOutputs [][]float64

	// Results
	Results      []*HILTestResult
	PassedTests  int
	FailedTests  int

	// Thresholds
	MSEThreshold float64
	MaxErrorThreshold float64
}

// HILTestResult contains single test result.
type HILTestResult struct {
	TestID       int
	Input        []float64
	Expected     []float64
	Actual       []float64
	MSE          float64
	MaxError     float64
	Passed       bool
	LatencyNs    float64
	EnergyPJ     float64
}

// NewHILSimulator creates a new HIL simulator.
func NewHILSimulator(config *HILConfig) *HILSimulator {
	if config == nil {
		config = DefaultHILConfig()
	}

	return &HILSimulator{
		Config:        config,
		HardwareModel: NewCIMHardwareModel(config),
		Results:       make([]*HILTestResult, 0),
		MSEThreshold:  0.01,
		MaxErrorThreshold: 0.1,
	}
}

// LoadWeights loads weights to hardware model.
func (sim *HILSimulator) LoadWeights(weights [][]float64) {
	sim.HardwareModel.ProgramWeights(weights)
}

// AddTestVector adds a test vector.
func (sim *HILSimulator) AddTestVector(input []float64, expected []float64) {
	sim.TestInputs = append(sim.TestInputs, input)
	sim.ExpectedOutputs = append(sim.ExpectedOutputs, expected)
}

// RunTest runs a single test.
func (sim *HILSimulator) RunTest(testID int) *HILTestResult {
	if testID >= len(sim.TestInputs) {
		return nil
	}

	input := sim.TestInputs[testID]
	expected := sim.ExpectedOutputs[testID]

	// Run MVM
	actual, mvmResult := sim.HardwareModel.MVM(input)

	// Compare with expected
	mse := 0.0
	maxError := 0.0
	for i := 0; i < len(expected) && i < len(actual); i++ {
		err := math.Abs(actual[i] - expected[i])
		mse += err * err
		if err > maxError {
			maxError = err
		}
	}
	if len(expected) > 0 {
		mse /= float64(len(expected))
	}

	passed := mse < sim.MSEThreshold && maxError < sim.MaxErrorThreshold

	result := &HILTestResult{
		TestID:    testID,
		Input:     input,
		Expected:  expected,
		Actual:    actual,
		MSE:       mse,
		MaxError:  maxError,
		Passed:    passed,
		LatencyNs: mvmResult.LatencyNs,
		EnergyPJ:  mvmResult.EnergyPJ,
	}

	sim.Results = append(sim.Results, result)
	if passed {
		sim.PassedTests++
	} else {
		sim.FailedTests++
	}

	sim.StepCount++
	sim.CurrentTimeUs += sim.Config.SimulationStepUs

	return result
}

// RunAllTests runs all test vectors.
func (sim *HILSimulator) RunAllTests() *HILSummary {
	for i := range sim.TestInputs {
		sim.RunTest(i)
	}

	return sim.GetSummary()
}

// GetSummary returns simulation summary.
func (sim *HILSimulator) GetSummary() *HILSummary {
	totalMSE := 0.0
	totalLatency := 0.0
	totalEnergy := 0.0
	maxError := 0.0

	for _, r := range sim.Results {
		totalMSE += r.MSE
		totalLatency += r.LatencyNs
		totalEnergy += r.EnergyPJ
		if r.MaxError > maxError {
			maxError = r.MaxError
		}
	}

	numTests := len(sim.Results)
	if numTests == 0 {
		numTests = 1
	}

	return &HILSummary{
		TotalTests:    len(sim.Results),
		PassedTests:   sim.PassedTests,
		FailedTests:   sim.FailedTests,
		PassRate:      float64(sim.PassedTests) / float64(numTests) * 100,
		AverageMSE:    totalMSE / float64(numTests),
		MaxError:      maxError,
		TotalLatencyNs: totalLatency,
		TotalEnergyPJ:  totalEnergy,
		TotalMVMOps:    sim.HardwareModel.TotalMVMOps,
		CumulativeError: sim.HardwareModel.CumulativeError,
	}
}

// HILSummary contains HIL simulation summary.
type HILSummary struct {
	TotalTests      int
	PassedTests     int
	FailedTests     int
	PassRate        float64
	AverageMSE      float64
	MaxError        float64
	TotalLatencyNs  float64
	TotalEnergyPJ   float64
	TotalMVMOps     int64
	CumulativeError float64
}

// =============================================================================
// NEUROSIM-STYLE VALIDATION
// =============================================================================

// NeuroSimConfig configures NeuroSim-style validation.
type NeuroSimConfig struct {
	// Technology parameters
	TechNode        int     // nm
	CellArea        float64 // F²
	MemoryType      string  // "RRAM", "FeFET", "PCM"

	// Array parameters
	ArrayRows       int
	ArrayCols       int
	SubarrayRows    int
	SubarrayCols    int

	// ADC/DAC
	ADCPrecision    int
	DACPrecision    int
	ADCAreaF2       float64
	ADCEnergyFJ     float64

	// Performance targets
	TargetAccuracy  float64
	TargetThroughput float64 // TOPS
	TargetEfficiency float64 // TOPS/W
}

// DefaultNeuroSimConfig returns default NeuroSim config.
func DefaultNeuroSimConfig() *NeuroSimConfig {
	return &NeuroSimConfig{
		TechNode:        28,
		CellArea:        50.0,
		MemoryType:      "FeFET",
		ArrayRows:       256,
		ArrayCols:       256,
		SubarrayRows:    64,
		SubarrayCols:    64,
		ADCPrecision:    6,
		DACPrecision:    8,
		ADCAreaF2:       1000.0,
		ADCEnergyFJ:     100.0,
		TargetAccuracy:  95.0,
		TargetThroughput: 10.0,
		TargetEfficiency: 100.0,
	}
}

// NeuroSimValidator validates CIM designs.
type NeuroSimValidator struct {
	Config *NeuroSimConfig

	// Computed metrics
	TotalArea       float64 // mm²
	TotalPower      float64 // mW
	Throughput      float64 // TOPS
	Efficiency      float64 // TOPS/W
	Latency         float64 // ns

	// Accuracy tracking
	AccuracyIdeal   float64
	AccuracyActual  float64
	AccuracyDrop    float64
}

// NewNeuroSimValidator creates a NeuroSim-style validator.
func NewNeuroSimValidator(config *NeuroSimConfig) *NeuroSimValidator {
	if config == nil {
		config = DefaultNeuroSimConfig()
	}

	validator := &NeuroSimValidator{
		Config: config,
	}

	// Compute metrics
	validator.computeMetrics()

	return validator
}

// computeMetrics computes area, power, throughput.
func (ns *NeuroSimValidator) computeMetrics() {
	cfg := ns.Config

	// Area calculation
	cellAreaUM2 := cfg.CellArea * float64(cfg.TechNode*cfg.TechNode) / 1e6
	arrayAreaMM2 := cellAreaUM2 * float64(cfg.ArrayRows*cfg.ArrayCols) / 1e6

	// ADC area
	numADCs := cfg.ArrayCols / cfg.SubarrayCols
	adcAreaUM2 := cfg.ADCAreaF2 * float64(cfg.TechNode*cfg.TechNode) / 1e6
	adcTotalMM2 := adcAreaUM2 * float64(numADCs) / 1e6

	ns.TotalArea = arrayAreaMM2 + adcTotalMM2

	// Latency (simplified)
	mvmLatencyNs := 50.0 // Base MVM latency
	adcLatencyNs := float64(1<<cfg.ADCPrecision) * 0.5 // SAR ADC
	ns.Latency = mvmLatencyNs + adcLatencyNs

	// Throughput
	opsPerMVM := float64(cfg.ArrayRows * cfg.ArrayCols * 2) // MACs
	mvmPerSecond := 1e9 / ns.Latency
	ns.Throughput = opsPerMVM * mvmPerSecond / 1e12 // TOPS

	// Power (simplified)
	readPowerUW := 10.0 * float64(cfg.ArrayRows) // µW per row
	adcPowerUW := cfg.ADCEnergyFJ * 1e3 / ns.Latency * float64(numADCs)
	ns.TotalPower = (readPowerUW + adcPowerUW) / 1000 // mW

	// Efficiency
	ns.Efficiency = ns.Throughput * 1000 / ns.TotalPower // TOPS/W
}

// ValidateAccuracy validates accuracy degradation.
func (ns *NeuroSimValidator) ValidateAccuracy(idealAcc, actualAcc float64) bool {
	ns.AccuracyIdeal = idealAcc
	ns.AccuracyActual = actualAcc
	ns.AccuracyDrop = idealAcc - actualAcc

	return actualAcc >= ns.Config.TargetAccuracy
}

// ValidateDesign performs full design validation.
func (ns *NeuroSimValidator) ValidateDesign() *DesignValidationResult {
	return &DesignValidationResult{
		// Area
		TotalAreaMM2:     ns.TotalArea,
		AreaMeetsTarget:  ns.TotalArea < 10.0, // 10 mm² target

		// Performance
		ThroughputTOPS:   ns.Throughput,
		ThroughputMeets:  ns.Throughput >= ns.Config.TargetThroughput,

		// Efficiency
		EfficiencyTOPSW:  ns.Efficiency,
		EfficiencyMeets:  ns.Efficiency >= ns.Config.TargetEfficiency,

		// Latency
		LatencyNs:        ns.Latency,

		// Power
		PowerMW:          ns.TotalPower,

		// Accuracy
		AccuracyDrop:     ns.AccuracyDrop,
		AccuracyMeets:    ns.AccuracyActual >= ns.Config.TargetAccuracy,

		// Overall
		OverallPass:      ns.Throughput >= ns.Config.TargetThroughput &&
		                  ns.Efficiency >= ns.Config.TargetEfficiency &&
		                  ns.AccuracyActual >= ns.Config.TargetAccuracy,
	}
}

// DesignValidationResult contains validation results.
type DesignValidationResult struct {
	TotalAreaMM2    float64
	AreaMeetsTarget bool

	ThroughputTOPS  float64
	ThroughputMeets bool

	EfficiencyTOPSW float64
	EfficiencyMeets bool

	LatencyNs       float64
	PowerMW         float64

	AccuracyDrop    float64
	AccuracyMeets   bool

	OverallPass     bool
}
