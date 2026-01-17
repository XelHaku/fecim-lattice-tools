// Package layers provides neural network layer implementations for CIM deployment.
// This file implements neuromorphic tactile sensing and federated learning for CIM.
// Based on research: PNAS (NRE-skin), Nature Communications (RePACK CIM),
// Science Robotics (neuro-inspired e-skin)
package layers

import (
	"math"
	"math/rand"
)

// =============================================================================
// NEUROMORPHIC TACTILE SENSING
// Bio-inspired e-skin with ferroelectric/piezoelectric transducers
// =============================================================================

// MechanoreceptorType defines types of tactile receptors.
type MechanoreceptorType int

const (
	MechanoreceptorSA1   MechanoreceptorType = iota // Slowly adapting type I (Merkel)
	MechanoreceptorSA2                               // Slowly adapting type II (Ruffini)
	MechanoreceptorFA1                               // Fast adapting type I (Meissner)
	MechanoreceptorFA2                               // Fast adapting type II (Pacinian)
	MechanoreceptorNoci                              // Nociceptor (pain)
	MechanoreceptorThermo                            // Thermoreceptor
)

// TactileSensorConfig configures a tactile sensor unit.
type TactileSensorConfig struct {
	Type            MechanoreceptorType
	SensitivityPa   float64 // Pressure sensitivity (Pa^-1)
	ThresholdPa     float64 // Activation threshold (Pa)
	SaturationPa    float64 // Saturation pressure (Pa)

	// Temporal dynamics
	AdaptationTau   float64 // Adaptation time constant (ms)
	ResponseTimeMs  float64 // Response time (ms)

	// Ferroelectric parameters
	UseFerroelectric bool
	PolarizationMax  float64 // µC/cm²
	CoerciveFieldMV  float64 // MV/cm
}

// DefaultMechanoreceptorConfigs returns biologically-inspired configs.
func DefaultMechanoreceptorConfigs() map[MechanoreceptorType]*TactileSensorConfig {
	return map[MechanoreceptorType]*TactileSensorConfig{
		MechanoreceptorSA1: {
			Type:             MechanoreceptorSA1,
			SensitivityPa:    0.01,
			ThresholdPa:      10.0,
			SaturationPa:     100000.0,
			AdaptationTau:    5000.0, // Slow adaptation
			ResponseTimeMs:   10.0,
			UseFerroelectric: true,
		},
		MechanoreceptorFA1: {
			Type:             MechanoreceptorFA1,
			SensitivityPa:    0.1,
			ThresholdPa:      5.0,
			SaturationPa:     50000.0,
			AdaptationTau:    50.0, // Fast adaptation
			ResponseTimeMs:   2.0,
			UseFerroelectric: true,
		},
		MechanoreceptorFA2: {
			Type:             MechanoreceptorFA2,
			SensitivityPa:    1.0,
			ThresholdPa:      1.0,
			SaturationPa:     10000.0,
			AdaptationTau:    10.0, // Very fast adaptation
			ResponseTimeMs:   0.5,
			UseFerroelectric: true,
		},
		MechanoreceptorNoci: {
			Type:             MechanoreceptorNoci,
			SensitivityPa:    0.001,
			ThresholdPa:      50000.0, // High threshold for pain
			SaturationPa:     500000.0,
			AdaptationTau:    10000.0, // Very slow adaptation
			ResponseTimeMs:   50.0,
			UseFerroelectric: false,
		},
	}
}

// TactileSensor implements a single tactile sensing unit.
type TactileSensor struct {
	Config *TactileSensorConfig

	// State
	CurrentPressure   float64
	AdaptedResponse   float64
	RawResponse       float64
	Polarization      float64

	// Spike generation
	MembranePotential float64
	SpikeThreshold    float64
	LastSpikeTime     float64
	RefractoryMs      float64

	// Temporal state
	LastUpdateTime    float64
	PressureHistory   []float64
}

// NewTactileSensor creates a new tactile sensor.
func NewTactileSensor(config *TactileSensorConfig) *TactileSensor {
	if config == nil {
		config = DefaultMechanoreceptorConfigs()[MechanoreceptorSA1]
	}
	return &TactileSensor{
		Config:          config,
		SpikeThreshold:  1.0,
		RefractoryMs:    2.0,
		PressureHistory: make([]float64, 0),
	}
}

// Sense processes pressure input and generates response.
func (ts *TactileSensor) Sense(pressure float64, timeMs float64) *TactileResponse {
	dt := timeMs - ts.LastUpdateTime
	if dt <= 0 {
		dt = 1
	}

	ts.CurrentPressure = pressure
	ts.PressureHistory = append(ts.PressureHistory, pressure)
	if len(ts.PressureHistory) > 100 {
		ts.PressureHistory = ts.PressureHistory[1:]
	}

	// Compute pressure derivative for FA receptors
	pressureDerivative := 0.0
	if len(ts.PressureHistory) >= 2 {
		pressureDerivative = (ts.PressureHistory[len(ts.PressureHistory)-1] -
			ts.PressureHistory[len(ts.PressureHistory)-2]) / dt
	}

	// Raw response based on receptor type
	var rawResp float64
	switch ts.Config.Type {
	case MechanoreceptorSA1, MechanoreceptorSA2:
		// Sustained response to static pressure
		if pressure > ts.Config.ThresholdPa {
			rawResp = (pressure - ts.Config.ThresholdPa) * ts.Config.SensitivityPa
		}
	case MechanoreceptorFA1, MechanoreceptorFA2:
		// Response to pressure changes
		rawResp = math.Abs(pressureDerivative) * ts.Config.SensitivityPa
	case MechanoreceptorNoci:
		// Pain response above threshold
		if pressure > ts.Config.ThresholdPa {
			rawResp = math.Pow((pressure-ts.Config.ThresholdPa)/ts.Config.ThresholdPa, 2) *
				ts.Config.SensitivityPa
		}
	}

	// Saturation
	maxResponse := ts.Config.SaturationPa * ts.Config.SensitivityPa
	if rawResp > maxResponse {
		rawResp = maxResponse
	}
	ts.RawResponse = rawResp

	// Adaptation
	adaptRate := dt / ts.Config.AdaptationTau
	if adaptRate > 1 {
		adaptRate = 1
	}
	ts.AdaptedResponse = ts.AdaptedResponse + adaptRate*(rawResp-ts.AdaptedResponse)

	// Final response (raw - adapted baseline)
	finalResponse := rawResp
	if ts.Config.Type != MechanoreceptorNoci {
		finalResponse = rawResp - ts.AdaptedResponse*0.5
		if finalResponse < 0 {
			finalResponse = 0
		}
	}

	// Ferroelectric modulation
	if ts.Config.UseFerroelectric {
		// Polarization follows pressure with hysteresis
		targetPol := finalResponse * ts.Config.PolarizationMax
		ts.Polarization = ts.Polarization + 0.1*(targetPol-ts.Polarization)
	}

	// Spike generation (integrate-and-fire)
	ts.MembranePotential += finalResponse * dt * 0.01
	spiked := false
	if ts.MembranePotential > ts.SpikeThreshold &&
		(timeMs-ts.LastSpikeTime) > ts.RefractoryMs {
		spiked = true
		ts.MembranePotential = 0
		ts.LastSpikeTime = timeMs
	}

	// Decay membrane potential
	ts.MembranePotential *= math.Exp(-dt / 20.0)

	ts.LastUpdateTime = timeMs

	return &TactileResponse{
		Pressure:      pressure,
		RawResponse:   rawResp,
		AdaptedResp:   finalResponse,
		Polarization:  ts.Polarization,
		Spiked:        spiked,
		MembranePot:   ts.MembranePotential,
		IsPainful:     ts.Config.Type == MechanoreceptorNoci && rawResp > 0,
	}
}

// TactileResponse contains tactile sensor output.
type TactileResponse struct {
	Pressure     float64
	RawResponse  float64
	AdaptedResp  float64
	Polarization float64
	Spiked       bool
	MembranePot  float64
	IsPainful    bool
}

// =============================================================================
// NEUROMORPHIC E-SKIN ARRAY
// =============================================================================

// ESkinConfig configures the electronic skin array.
type ESkinConfig struct {
	Width, Height int
	PixelSizeMm   float64

	// Receptor distribution
	SA1Density    float64 // Proportion of SA1 receptors
	FA1Density    float64
	FA2Density    float64
	NociDensity   float64

	// Pain detection
	EnablePainDetection bool
	PainThresholdPa     float64

	// Injury detection
	EnableInjuryDetection bool
}

// ESkinPixel represents a single e-skin pixel.
type ESkinPixel struct {
	X, Y      int
	Receptors []*TactileSensor
	Damaged   bool
}

// NeuromorphicESkin implements bio-inspired electronic skin.
type NeuromorphicESkin struct {
	Config *ESkinConfig
	Pixels [][]*ESkinPixel

	// Global state
	CurrentTime     float64
	PainLocations   [][]int
	InjuryLocations [][]int

	// Processing outputs
	PressureMap     [][]float64
	SpikeMap        [][]bool
	PainMap         [][]bool
}

// NewNeuromorphicESkin creates a new e-skin array.
func NewNeuromorphicESkin(config *ESkinConfig) *NeuromorphicESkin {
	if config == nil {
		config = &ESkinConfig{
			Width:                 16,
			Height:                16,
			PixelSizeMm:           2.0,
			SA1Density:            0.4,
			FA1Density:            0.3,
			FA2Density:            0.2,
			NociDensity:           0.1,
			EnablePainDetection:   true,
			PainThresholdPa:       50000.0,
			EnableInjuryDetection: true,
		}
	}

	receptorConfigs := DefaultMechanoreceptorConfigs()

	pixels := make([][]*ESkinPixel, config.Height)
	for y := 0; y < config.Height; y++ {
		pixels[y] = make([]*ESkinPixel, config.Width)
		for x := 0; x < config.Width; x++ {
			// Create receptors based on density
			receptors := make([]*TactileSensor, 0)

			if rand.Float64() < config.SA1Density {
				receptors = append(receptors, NewTactileSensor(receptorConfigs[MechanoreceptorSA1]))
			}
			if rand.Float64() < config.FA1Density {
				receptors = append(receptors, NewTactileSensor(receptorConfigs[MechanoreceptorFA1]))
			}
			if rand.Float64() < config.FA2Density {
				receptors = append(receptors, NewTactileSensor(receptorConfigs[MechanoreceptorFA2]))
			}
			if rand.Float64() < config.NociDensity {
				receptors = append(receptors, NewTactileSensor(receptorConfigs[MechanoreceptorNoci]))
			}

			// Ensure at least one receptor
			if len(receptors) == 0 {
				receptors = append(receptors, NewTactileSensor(receptorConfigs[MechanoreceptorSA1]))
			}

			pixels[y][x] = &ESkinPixel{
				X:         x,
				Y:         y,
				Receptors: receptors,
			}
		}
	}

	return &NeuromorphicESkin{
		Config:          config,
		Pixels:          pixels,
		PainLocations:   make([][]int, 0),
		InjuryLocations: make([][]int, 0),
		PressureMap:     make([][]float64, config.Height),
		SpikeMap:        make([][]bool, config.Height),
		PainMap:         make([][]bool, config.Height),
	}
}

// ProcessPressureMap processes a 2D pressure input.
func (es *NeuromorphicESkin) ProcessPressureMap(pressures [][]float64, timeMs float64) *ESkinOutput {
	es.CurrentTime = timeMs

	// Initialize output maps
	for y := 0; y < es.Config.Height; y++ {
		es.PressureMap[y] = make([]float64, es.Config.Width)
		es.SpikeMap[y] = make([]bool, es.Config.Width)
		es.PainMap[y] = make([]bool, es.Config.Width)
	}

	es.PainLocations = make([][]int, 0)

	totalSpikes := 0
	painDetected := false
	maxPressure := 0.0

	// Process each pixel
	for y := 0; y < es.Config.Height && y < len(pressures); y++ {
		for x := 0; x < es.Config.Width && x < len(pressures[y]); x++ {
			pixel := es.Pixels[y][x]
			if pixel.Damaged {
				continue
			}

			pressure := pressures[y][x]
			es.PressureMap[y][x] = pressure

			if pressure > maxPressure {
				maxPressure = pressure
			}

			// Process through all receptors
			pixelSpiked := false
			pixelPain := false

			for _, receptor := range pixel.Receptors {
				resp := receptor.Sense(pressure, timeMs)
				if resp.Spiked {
					pixelSpiked = true
					totalSpikes++
				}
				if resp.IsPainful {
					pixelPain = true
				}
			}

			es.SpikeMap[y][x] = pixelSpiked
			es.PainMap[y][x] = pixelPain

			if pixelPain {
				es.PainLocations = append(es.PainLocations, []int{x, y})
				painDetected = true
			}
		}
	}

	return &ESkinOutput{
		PressureMap:   es.PressureMap,
		SpikeMap:      es.SpikeMap,
		PainMap:       es.PainMap,
		PainLocations: es.PainLocations,
		PainDetected:  painDetected,
		TotalSpikes:   totalSpikes,
		MaxPressure:   maxPressure,
	}
}

// MarkDamaged marks a pixel as damaged/injured.
func (es *NeuromorphicESkin) MarkDamaged(x, y int) {
	if y >= 0 && y < es.Config.Height && x >= 0 && x < es.Config.Width {
		es.Pixels[y][x].Damaged = true
		es.InjuryLocations = append(es.InjuryLocations, []int{x, y})
	}
}

// GetDamagedPixels returns locations of damaged pixels.
func (es *NeuromorphicESkin) GetDamagedPixels() [][]int {
	return es.InjuryLocations
}

// ESkinOutput contains e-skin processing output.
type ESkinOutput struct {
	PressureMap   [][]float64
	SpikeMap      [][]bool
	PainMap       [][]bool
	PainLocations [][]int
	PainDetected  bool
	TotalSpikes   int
	MaxPressure   float64
}

// =============================================================================
// TACTILE CLASSIFICATION WITH CIM
// =============================================================================

// TactileClassifier implements tactile pattern classification.
type TactileClassifier struct {
	InputSize    int
	HiddenSize   int
	OutputSize   int

	// Weights (CIM-mapped)
	Weights1     [][]float64
	Weights2     [][]float64
	Biases1      []float64
	Biases2      []float64

	// Quantization
	WeightBits   int
}

// NewTactileClassifier creates a tactile classifier.
func NewTactileClassifier(inputSize, hiddenSize, outputSize int) *TactileClassifier {
	// Initialize weights
	weights1 := make([][]float64, inputSize)
	for i := range weights1 {
		weights1[i] = make([]float64, hiddenSize)
		scale := math.Sqrt(2.0 / float64(inputSize))
		for j := range weights1[i] {
			weights1[i][j] = rand.NormFloat64() * scale
		}
	}

	weights2 := make([][]float64, hiddenSize)
	for i := range weights2 {
		weights2[i] = make([]float64, outputSize)
		scale := math.Sqrt(2.0 / float64(hiddenSize))
		for j := range weights2[i] {
			weights2[i][j] = rand.NormFloat64() * scale
		}
	}

	return &TactileClassifier{
		InputSize:  inputSize,
		HiddenSize: hiddenSize,
		OutputSize: outputSize,
		Weights1:   weights1,
		Weights2:   weights2,
		Biases1:    make([]float64, hiddenSize),
		Biases2:    make([]float64, outputSize),
		WeightBits: 6,
	}
}

// Classify performs tactile pattern classification.
func (tc *TactileClassifier) Classify(input []float64) (int, []float64) {
	// Hidden layer
	hidden := make([]float64, tc.HiddenSize)
	for j := 0; j < tc.HiddenSize; j++ {
		sum := tc.Biases1[j]
		for i := 0; i < tc.InputSize && i < len(input); i++ {
			sum += input[i] * tc.Weights1[i][j]
		}
		// ReLU
		if sum > 0 {
			hidden[j] = sum
		}
	}

	// Output layer
	output := make([]float64, tc.OutputSize)
	for j := 0; j < tc.OutputSize; j++ {
		sum := tc.Biases2[j]
		for i := 0; i < tc.HiddenSize; i++ {
			sum += hidden[i] * tc.Weights2[i][j]
		}
		output[j] = sum
	}

	// Softmax
	maxVal := output[0]
	for _, v := range output {
		if v > maxVal {
			maxVal = v
		}
	}
	sumExp := 0.0
	for i := range output {
		output[i] = math.Exp(output[i] - maxVal)
		sumExp += output[i]
	}
	for i := range output {
		output[i] /= sumExp
	}

	// Find max class
	maxClass := 0
	maxProb := output[0]
	for i, p := range output {
		if p > maxProb {
			maxProb = p
			maxClass = i
		}
	}

	return maxClass, output
}

// =============================================================================
// FEDERATED LEARNING FOR CIM
// Privacy-preserving distributed learning on edge CIM devices
// =============================================================================

// FederatedConfig configures federated learning.
type FederatedConfig struct {
	NumClients       int
	LocalEpochs      int
	LocalBatchSize   int
	GlobalRounds     int

	// Aggregation
	AggregationMethod string // "fedavg", "fedprox", "scaffold"
	FedProxMu         float64

	// Privacy
	UseDifferentialPrivacy bool
	DPEpsilon              float64
	DPDelta                float64
	ClipNorm               float64

	// Communication
	CompressionBits  int
	TopKSparsity     float64

	// CIM-specific
	QuantizeBits     int
	UseSecureAgg     bool
}

// DefaultFederatedConfig returns default federated config.
func DefaultFederatedConfig() *FederatedConfig {
	return &FederatedConfig{
		NumClients:            10,
		LocalEpochs:           5,
		LocalBatchSize:        32,
		GlobalRounds:          100,
		AggregationMethod:     "fedavg",
		FedProxMu:             0.01,
		UseDifferentialPrivacy: true,
		DPEpsilon:             1.0,
		DPDelta:               1e-5,
		ClipNorm:              1.0,
		CompressionBits:       8,
		TopKSparsity:          0.1,
		QuantizeBits:          6,
		UseSecureAgg:          true,
	}
}

// FederatedClient represents a federated learning client.
type FederatedClient struct {
	ID           int
	LocalWeights [][][]float64
	LocalData    [][]float64
	LocalLabels  []int
	DataSize     int

	// Training state
	Gradients    [][][]float64
	LocalLoss    float64

	// Privacy state
	NoiseScale   float64
}

// FederatedServer manages federated learning.
type FederatedServer struct {
	Config       *FederatedConfig
	Clients      []*FederatedClient
	GlobalWeights [][][]float64

	// Training state
	CurrentRound int
	GlobalLoss   float64
	Accuracy     float64

	// History
	LossHistory  []float64
	AccHistory   []float64
}

// NewFederatedServer creates a federated learning server.
func NewFederatedServer(layerSizes []int, config *FederatedConfig) *FederatedServer {
	if config == nil {
		config = DefaultFederatedConfig()
	}

	// Initialize global weights
	globalWeights := make([][][]float64, len(layerSizes)-1)
	for l := 0; l < len(layerSizes)-1; l++ {
		inSize, outSize := layerSizes[l], layerSizes[l+1]
		globalWeights[l] = make([][]float64, inSize)
		scale := math.Sqrt(2.0 / float64(inSize))
		for i := range globalWeights[l] {
			globalWeights[l][i] = make([]float64, outSize)
			for j := range globalWeights[l][i] {
				globalWeights[l][i][j] = rand.NormFloat64() * scale
			}
		}
	}

	// Initialize clients
	clients := make([]*FederatedClient, config.NumClients)
	for i := 0; i < config.NumClients; i++ {
		// Deep copy global weights
		localWeights := make([][][]float64, len(globalWeights))
		for l := range globalWeights {
			localWeights[l] = make([][]float64, len(globalWeights[l]))
			for i2 := range globalWeights[l] {
				localWeights[l][i2] = make([]float64, len(globalWeights[l][i2]))
				copy(localWeights[l][i2], globalWeights[l][i2])
			}
		}

		clients[i] = &FederatedClient{
			ID:           i,
			LocalWeights: localWeights,
			NoiseScale:   config.DPEpsilon,
		}
	}

	return &FederatedServer{
		Config:        config,
		Clients:       clients,
		GlobalWeights: globalWeights,
		LossHistory:   make([]float64, 0),
		AccHistory:    make([]float64, 0),
	}
}

// DistributeGlobalWeights sends global weights to all clients.
func (fs *FederatedServer) DistributeGlobalWeights() {
	for _, client := range fs.Clients {
		// Deep copy global weights to client
		for l := range fs.GlobalWeights {
			for i := range fs.GlobalWeights[l] {
				copy(client.LocalWeights[l][i], fs.GlobalWeights[l][i])
			}
		}
	}
}

// ClientLocalTrain performs local training on a client.
func (fs *FederatedServer) ClientLocalTrain(clientID int) {
	client := fs.Clients[clientID]
	if client.DataSize == 0 {
		return
	}

	// Simplified local training (gradient computation)
	lr := 0.01

	for epoch := 0; epoch < fs.Config.LocalEpochs; epoch++ {
		// Mini-batch SGD simulation
		for batch := 0; batch < client.DataSize/fs.Config.LocalBatchSize; batch++ {
			// Compute gradients (simplified)
			for l := range client.LocalWeights {
				for i := range client.LocalWeights[l] {
					for j := range client.LocalWeights[l][i] {
						// Simulated gradient
						grad := rand.NormFloat64() * 0.01

						// FedProx regularization
						if fs.Config.AggregationMethod == "fedprox" {
							proxTerm := fs.Config.FedProxMu *
								(client.LocalWeights[l][i][j] - fs.GlobalWeights[l][i][j])
							grad += proxTerm
						}

						// Gradient clipping
						if math.Abs(grad) > fs.Config.ClipNorm {
							grad = fs.Config.ClipNorm * grad / math.Abs(grad)
						}

						// Add DP noise
						if fs.Config.UseDifferentialPrivacy {
							noise := rand.NormFloat64() * fs.Config.ClipNorm *
								math.Sqrt(2*math.Log(1.25/fs.Config.DPDelta)) / fs.Config.DPEpsilon
							grad += noise
						}

						// Update
						client.LocalWeights[l][i][j] -= lr * grad
					}
				}
			}
		}
	}
}

// AggregateWeights aggregates client weights using FedAvg.
func (fs *FederatedServer) AggregateWeights() {
	totalData := 0
	for _, client := range fs.Clients {
		totalData += client.DataSize
	}
	if totalData == 0 {
		totalData = len(fs.Clients) // Equal weighting if no data
	}

	// Reset global weights
	for l := range fs.GlobalWeights {
		for i := range fs.GlobalWeights[l] {
			for j := range fs.GlobalWeights[l][i] {
				fs.GlobalWeights[l][i][j] = 0
			}
		}
	}

	// Weighted average
	for _, client := range fs.Clients {
		weight := float64(client.DataSize) / float64(totalData)
		if client.DataSize == 0 {
			weight = 1.0 / float64(len(fs.Clients))
		}

		for l := range client.LocalWeights {
			for i := range client.LocalWeights[l] {
				for j := range client.LocalWeights[l][i] {
					fs.GlobalWeights[l][i][j] += weight * client.LocalWeights[l][i][j]
				}
			}
		}
	}

	// Quantize for CIM deployment
	if fs.Config.QuantizeBits > 0 {
		fs.QuantizeGlobalWeights()
	}
}

// QuantizeGlobalWeights quantizes weights for CIM.
func (fs *FederatedServer) QuantizeGlobalWeights() {
	levels := float64(int(1) << fs.Config.QuantizeBits)

	for l := range fs.GlobalWeights {
		// Find min/max
		minW, maxW := fs.GlobalWeights[l][0][0], fs.GlobalWeights[l][0][0]
		for i := range fs.GlobalWeights[l] {
			for j := range fs.GlobalWeights[l][i] {
				if fs.GlobalWeights[l][i][j] < minW {
					minW = fs.GlobalWeights[l][i][j]
				}
				if fs.GlobalWeights[l][i][j] > maxW {
					maxW = fs.GlobalWeights[l][i][j]
				}
			}
		}

		// Quantize
		scale := (maxW - minW) / levels
		if scale == 0 {
			scale = 1
		}
		for i := range fs.GlobalWeights[l] {
			for j := range fs.GlobalWeights[l][i] {
				normalized := (fs.GlobalWeights[l][i][j] - minW) / scale
				quantized := math.Round(normalized)
				fs.GlobalWeights[l][i][j] = quantized*scale + minW
			}
		}
	}
}

// RunRound executes one federated learning round.
func (fs *FederatedServer) RunRound() {
	fs.CurrentRound++

	// Distribute global weights
	fs.DistributeGlobalWeights()

	// Client local training
	for i := range fs.Clients {
		fs.ClientLocalTrain(i)
	}

	// Aggregate
	fs.AggregateWeights()

	// Record history
	fs.LossHistory = append(fs.LossHistory, fs.GlobalLoss)
	fs.AccHistory = append(fs.AccHistory, fs.Accuracy)
}

// SetClientData sets training data for a client.
func (fs *FederatedServer) SetClientData(clientID int, data [][]float64, labels []int) {
	if clientID >= 0 && clientID < len(fs.Clients) {
		fs.Clients[clientID].LocalData = data
		fs.Clients[clientID].LocalLabels = labels
		fs.Clients[clientID].DataSize = len(data)
	}
}

// =============================================================================
// REPACK-STYLE CIM ENCRYPTION
// Privacy protection via inherent device variation
// =============================================================================

// RePACKConfig configures RePACK encryption.
type RePACKConfig struct {
	UseInherentVariation bool
	VariationSigma       float64
	EncryptionKey        []float64

	// PUF parameters
	ChallengeBits        int
	ResponseBits         int
}

// RePACKEncryptor implements RePACK-style CIM encryption.
type RePACKEncryptor struct {
	Config           *RePACKConfig
	VariationPattern [][]float64 // Unique per-chip variation
	PUFResponses     map[string][]float64
}

// NewRePACKEncryptor creates a RePACK encryptor.
func NewRePACKEncryptor(rows, cols int, config *RePACKConfig) *RePACKEncryptor {
	if config == nil {
		config = &RePACKConfig{
			UseInherentVariation: true,
			VariationSigma:       0.05,
			ChallengeBits:        64,
			ResponseBits:         128,
		}
	}

	// Generate unique variation pattern (simulates manufacturing variation)
	variation := make([][]float64, rows)
	for i := range variation {
		variation[i] = make([]float64, cols)
		for j := range variation[i] {
			variation[i][j] = 1.0 + config.VariationSigma*rand.NormFloat64()
		}
	}

	return &RePACKEncryptor{
		Config:           config,
		VariationPattern: variation,
		PUFResponses:     make(map[string][]float64),
	}
}

// EncryptWeights encrypts weights using variation pattern.
func (re *RePACKEncryptor) EncryptWeights(weights [][]float64) [][]float64 {
	rows := len(weights)
	cols := len(weights[0])

	encrypted := make([][]float64, rows)
	for i := range encrypted {
		encrypted[i] = make([]float64, cols)
		for j := range encrypted[i] {
			// XOR-like encryption using variation
			varI := i % len(re.VariationPattern)
			varJ := j % len(re.VariationPattern[0])
			encrypted[i][j] = weights[i][j] * re.VariationPattern[varI][varJ]
		}
	}

	return encrypted
}

// DecryptWeights decrypts weights (requires same chip).
func (re *RePACKEncryptor) DecryptWeights(encrypted [][]float64) [][]float64 {
	rows := len(encrypted)
	cols := len(encrypted[0])

	decrypted := make([][]float64, rows)
	for i := range decrypted {
		decrypted[i] = make([]float64, cols)
		for j := range decrypted[i] {
			varI := i % len(re.VariationPattern)
			varJ := j % len(re.VariationPattern[0])
			decrypted[i][j] = encrypted[i][j] / re.VariationPattern[varI][varJ]
		}
	}

	return decrypted
}

// GeneratePUFResponse generates PUF response for challenge.
func (re *RePACKEncryptor) GeneratePUFResponse(challenge []float64) []float64 {
	response := make([]float64, re.Config.ResponseBits)

	for i := 0; i < re.Config.ResponseBits; i++ {
		// Use variation pattern as PUF
		row := i % len(re.VariationPattern)
		col := (i / len(re.VariationPattern)) % len(re.VariationPattern[0])

		challengeIdx := i % len(challenge)
		response[i] = re.VariationPattern[row][col] * challenge[challengeIdx]

		// Threshold to binary
		if response[i] > 1.0 {
			response[i] = 1.0
		} else {
			response[i] = 0.0
		}
	}

	return response
}

// =============================================================================
// FEDERATED CIM METRICS
// =============================================================================

// FederatedCIMMetrics contains federated learning metrics.
type FederatedCIMMetrics struct {
	GlobalAccuracy    float64
	AverageLocalLoss  float64
	CommunicationCost int64   // bits
	PrivacyBudget     float64 // ε spent
	NumRounds         int
	NumClients        int
	Convergence       bool
}

// ComputeFederatedMetrics computes FL metrics.
func ComputeFederatedMetrics(server *FederatedServer) *FederatedCIMMetrics {
	// Compute average local loss
	avgLoss := 0.0
	for _, client := range server.Clients {
		avgLoss += client.LocalLoss
	}
	avgLoss /= float64(len(server.Clients))

	// Compute communication cost
	var commCost int64
	for l := range server.GlobalWeights {
		for i := range server.GlobalWeights[l] {
			commCost += int64(len(server.GlobalWeights[l][i]) * server.Config.CompressionBits)
		}
	}
	commCost *= int64(len(server.Clients)) * 2 // Up and down

	// Check convergence
	converged := false
	if len(server.LossHistory) > 10 {
		recent := server.LossHistory[len(server.LossHistory)-10:]
		variance := 0.0
		mean := 0.0
		for _, l := range recent {
			mean += l
		}
		mean /= float64(len(recent))
		for _, l := range recent {
			variance += (l - mean) * (l - mean)
		}
		variance /= float64(len(recent))
		converged = variance < 0.001
	}

	return &FederatedCIMMetrics{
		GlobalAccuracy:    server.Accuracy,
		AverageLocalLoss:  avgLoss,
		CommunicationCost: commCost,
		PrivacyBudget:     float64(server.CurrentRound) * server.Config.DPEpsilon,
		NumRounds:         server.CurrentRound,
		NumClients:        len(server.Clients),
		Convergence:       converged,
	}
}
