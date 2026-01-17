// Package layers provides federated learning and neuromorphic vision sensor
// implementations for IronLattice ferroelectric CIM systems.
//
// Based on research findings:
// - Ferroelectric-memristor hybrid memory for training AND inference (Nature Electronics 2025)
// - CMOS-compatible ferroelectric hybrid CIM achieving 96.36% yield
// - DVS/event cameras with asynchronous sparse output
// - SCIMITAR architecture for sparse CIM processing
// - SPIDR: SNN accelerator for event-based perception
// - AI-native robotic vision with in-sensor computing
//
// Reference: IronLattice Research Log Sections 278-279
package layers

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
)

// ============================================================================
// FEDERATED LEARNING FOR FERROELECTRIC CIM
// ============================================================================

// FederatedConfig configures the federated learning system
type FederatedConfig struct {
	// Communication settings
	NumClients           int     // Number of federated clients
	LocalEpochs          int     // Epochs per round on each client
	BatchSize            int     // Local training batch size
	CommunicationRounds  int     // Total federated rounds

	// Aggregation strategy
	AggregationMethod    string  // "fedavg", "fedprox", "scaffold", "fednova"
	FedProxMu            float64 // Proximal term coefficient for FedProx
	ClientFraction       float64 // Fraction of clients per round (0-1)

	// Gradient compression
	CompressionEnabled   bool    // Enable gradient compression
	CompressionRatio     float64 // Target compression ratio
	TopKSparsity         float64 // Top-K sparsification ratio
	QuantizationBits     int     // Bits for gradient quantization

	// FeFET-specific settings
	OnDeviceTraining     bool    // Enable on-device training with FeFET
	WeightUpdateVoltage  float64 // Programming voltage for updates (V)
	PulseWidth           float64 // Programming pulse width (ns)
	MaxEndurance         int     // FeFET endurance cycles

	// Privacy
	DifferentialPrivacy  bool    // Enable differential privacy
	DPEpsilon            float64 // Privacy budget epsilon
	DPDelta              float64 // Privacy parameter delta
	DPClipNorm           float64 // Gradient clipping norm
}

// DefaultFederatedConfig returns default federated learning configuration
func DefaultFederatedConfig() *FederatedConfig {
	return &FederatedConfig{
		NumClients:          100,
		LocalEpochs:         5,
		BatchSize:           32,
		CommunicationRounds: 100,
		AggregationMethod:   "fedavg",
		FedProxMu:           0.01,
		ClientFraction:      0.1,
		CompressionEnabled:  true,
		CompressionRatio:    0.1,
		TopKSparsity:        0.01,
		QuantizationBits:    8,
		OnDeviceTraining:    true,
		WeightUpdateVoltage: 3.3,
		PulseWidth:          100.0,
		MaxEndurance:        1e10,
		DifferentialPrivacy: false,
		DPEpsilon:           1.0,
		DPDelta:             1e-5,
		DPClipNorm:          1.0,
	}
}

// FeFETWeight represents a ferroelectric weight cell for on-device training
type FeFETWeight struct {
	Polarization     float64 // Current polarization state (-1 to 1)
	Conductance      float64 // Effective conductance (μS)
	WriteCycles      int     // Total write operations
	MaxEndurance     int     // Maximum endurance
	RetentionLoss    float64 // Cumulative retention degradation

	// Multi-domain state for analog weight
	DomainStates     []float64 // Individual domain polarizations
	NumDomains       int       // Number of ferroelectric domains
}

// NewFeFETWeight creates a new FeFET weight cell
func NewFeFETWeight(numDomains int, maxEndurance int) *FeFETWeight {
	domains := make([]float64, numDomains)
	for i := range domains {
		domains[i] = 0.0 // Initialize to zero polarization
	}

	return &FeFETWeight{
		Polarization:  0.0,
		Conductance:   1.0,
		WriteCycles:   0,
		MaxEndurance:  maxEndurance,
		RetentionLoss: 0.0,
		DomainStates:  domains,
		NumDomains:    numDomains,
	}
}

// Update applies a weight update using FeFET programming
func (w *FeFETWeight) Update(delta float64, voltage float64, pulseWidth float64) error {
	if w.WriteCycles >= w.MaxEndurance {
		return fmt.Errorf("FeFET endurance exceeded: %d >= %d", w.WriteCycles, w.MaxEndurance)
	}

	// Calculate polarization change based on applied field
	// Landau-Khalatnikov dynamics: dP/dt = -γ(∂F/∂P)
	coerciveField := 1.0 // Normalized coercive field
	effectiveField := voltage / 3.3 // Normalize to typical FeFET voltage

	// Sigmoid-like switching characteristic
	switchingProb := 1.0 / (1.0 + math.Exp(-5.0*(math.Abs(effectiveField)-coerciveField)))

	// Update individual domains stochastically
	sign := 1.0
	if delta < 0 {
		sign = -1.0
	}

	for i := range w.DomainStates {
		if rand.Float64() < switchingProb*math.Abs(delta) {
			// Domain switches toward target state
			target := sign * 1.0
			w.DomainStates[i] = w.DomainStates[i] + 0.1*(target-w.DomainStates[i])
			w.DomainStates[i] = math.Max(-1.0, math.Min(1.0, w.DomainStates[i]))
		}
	}

	// Average polarization across domains
	sum := 0.0
	for _, d := range w.DomainStates {
		sum += d
	}
	w.Polarization = sum / float64(w.NumDomains)

	// Update conductance (linear mapping)
	// Conductance range: 1-100 μS typical for FeFET
	w.Conductance = 50.0 + 49.0*w.Polarization

	w.WriteCycles++

	// Apply fatigue model
	fatigueFactor := 1.0 - 0.1*math.Log10(float64(w.WriteCycles+1))
	w.Conductance *= fatigueFactor

	return nil
}

// FederatedClient represents a federated learning client with FeFET CIM
type FederatedClient struct {
	Config       *FederatedConfig
	ID           int

	// Local model weights (FeFET-based)
	Weights      [][]*FeFETWeight // [layer][weight]
	Gradients    [][]float64      // Accumulated gradients

	// Local data statistics
	NumSamples   int
	DataQuality  float64 // Measure of local data quality

	// Training state
	LocalLoss    float64
	LocalAccuracy float64

	// Communication buffer
	CompressedGradients []float64
	GradientMask        []bool // For Top-K sparsification
}

// NewFederatedClient creates a new federated client
func NewFederatedClient(config *FederatedConfig, clientID int, layerSizes []int) *FederatedClient {
	// Initialize FeFET weights
	weights := make([][]*FeFETWeight, len(layerSizes)-1)
	gradients := make([][]float64, len(layerSizes)-1)

	for l := 0; l < len(layerSizes)-1; l++ {
		size := layerSizes[l] * layerSizes[l+1]
		weights[l] = make([]*FeFETWeight, size)
		gradients[l] = make([]float64, size)

		for i := 0; i < size; i++ {
			weights[l][i] = NewFeFETWeight(16, config.MaxEndurance)
			// Xavier initialization
			stddev := math.Sqrt(2.0 / float64(layerSizes[l]+layerSizes[l+1]))
			weights[l][i].Polarization = rand.NormFloat64() * stddev
			weights[l][i].Conductance = 50.0 + 49.0*weights[l][i].Polarization
		}
	}

	return &FederatedClient{
		Config:     config,
		ID:         clientID,
		Weights:    weights,
		Gradients:  gradients,
		NumSamples: 0,
		DataQuality: 1.0,
	}
}

// LocalTrain performs local training on the client
func (c *FederatedClient) LocalTrain(data [][]float64, labels []int) error {
	c.NumSamples = len(data)

	for epoch := 0; epoch < c.Config.LocalEpochs; epoch++ {
		// Mini-batch SGD
		numBatches := (len(data) + c.Config.BatchSize - 1) / c.Config.BatchSize
		epochLoss := 0.0

		for batch := 0; batch < numBatches; batch++ {
			start := batch * c.Config.BatchSize
			end := start + c.Config.BatchSize
			if end > len(data) {
				end = len(data)
			}

			// Forward pass and gradient computation (simplified)
			batchLoss := c.computeBatchGradients(data[start:end], labels[start:end])
			epochLoss += batchLoss

			// Apply gradients using FeFET programming
			if c.Config.OnDeviceTraining {
				c.applyFeFETUpdates()
			}
		}

		c.LocalLoss = epochLoss / float64(numBatches)
	}

	return nil
}

// computeBatchGradients computes gradients for a batch
func (c *FederatedClient) computeBatchGradients(batch [][]float64, labels []int) float64 {
	// Simplified gradient computation
	// In practice, this would be full backpropagation

	totalLoss := 0.0

	for i, sample := range batch {
		// Forward pass through layers
		activations := sample
		for l, layerWeights := range c.Weights {
			nextSize := len(c.Gradients[l]) / len(activations)
			if nextSize == 0 {
				nextSize = 1
			}
			nextActivations := make([]float64, nextSize)

			for j := 0; j < nextSize; j++ {
				sum := 0.0
				for k, a := range activations {
					idx := k*nextSize + j
					if idx < len(layerWeights) {
						sum += a * layerWeights[idx].Conductance / 50.0
					}
				}
				nextActivations[j] = math.Tanh(sum) // Activation
			}
			activations = nextActivations
		}

		// Compute loss (cross-entropy approximation)
		if len(activations) > 0 {
			target := labels[i]
			if target < len(activations) {
				loss := -math.Log(math.Max(0.01, (activations[target]+1.0)/2.0))
				totalLoss += loss
			}
		}

		// Backward pass (simplified - random gradients for simulation)
		for l := range c.Gradients {
			for j := range c.Gradients[l] {
				// Simulated gradient with noise
				c.Gradients[l][j] += 0.01 * (rand.Float64() - 0.5)
			}
		}
	}

	// Normalize gradients
	for l := range c.Gradients {
		for j := range c.Gradients[l] {
			c.Gradients[l][j] /= float64(len(batch))
		}
	}

	return totalLoss / float64(len(batch))
}

// applyFeFETUpdates applies gradient updates using FeFET programming
func (c *FederatedClient) applyFeFETUpdates() {
	learningRate := 0.01

	for l := range c.Weights {
		for j, w := range c.Weights[l] {
			delta := -learningRate * c.Gradients[l][j]
			w.Update(delta, c.Config.WeightUpdateVoltage, c.Config.PulseWidth)
		}
	}
}

// CompressGradients compresses gradients for communication
func (c *FederatedClient) CompressGradients() []float64 {
	// Flatten all gradients
	totalSize := 0
	for _, g := range c.Gradients {
		totalSize += len(g)
	}

	flat := make([]float64, totalSize)
	idx := 0
	for _, g := range c.Gradients {
		copy(flat[idx:], g)
		idx += len(g)
	}

	if !c.Config.CompressionEnabled {
		c.CompressedGradients = flat
		return flat
	}

	// Top-K sparsification
	k := int(float64(len(flat)) * c.Config.TopKSparsity)
	if k < 1 {
		k = 1
	}

	// Find top-K by magnitude
	type gradIdx struct {
		value float64
		idx   int
	}

	sortable := make([]gradIdx, len(flat))
	for i, v := range flat {
		sortable[i] = gradIdx{math.Abs(v), i}
	}

	sort.Slice(sortable, func(i, j int) bool {
		return sortable[i].value > sortable[j].value
	})

	// Create sparse representation
	c.GradientMask = make([]bool, len(flat))
	compressed := make([]float64, len(flat))

	for i := 0; i < k; i++ {
		idx := sortable[i].idx
		c.GradientMask[idx] = true
		compressed[idx] = flat[idx]
	}

	// Quantize remaining values
	if c.Config.QuantizationBits > 0 {
		levels := 1 << c.Config.QuantizationBits
		for i := range compressed {
			if c.GradientMask[i] {
				// Find max for scaling
				maxVal := 0.0
				for _, v := range compressed {
					if math.Abs(v) > maxVal {
						maxVal = math.Abs(v)
					}
				}
				if maxVal > 0 {
					// Quantize
					scaled := compressed[i] / maxVal
					quantized := math.Round(scaled * float64(levels/2))
					compressed[i] = quantized * maxVal / float64(levels/2)
				}
			}
		}
	}

	c.CompressedGradients = compressed
	return compressed
}

// FederatedServer manages federated aggregation
type FederatedServer struct {
	Config       *FederatedConfig

	// Global model
	GlobalWeights [][]float64

	// Aggregation state
	ClientUpdates map[int][]float64
	ClientSamples map[int]int

	// SCAFFOLD control variates
	ServerControl []float64
	ClientControl map[int][]float64

	// Statistics
	RoundNumber   int
	GlobalLoss    float64
	GlobalAccuracy float64
}

// NewFederatedServer creates a federated server
func NewFederatedServer(config *FederatedConfig, layerSizes []int) *FederatedServer {
	// Initialize global weights
	totalWeights := 0
	for i := 0; i < len(layerSizes)-1; i++ {
		totalWeights += layerSizes[i] * layerSizes[i+1]
	}

	globalWeights := make([][]float64, len(layerSizes)-1)
	idx := 0
	for l := 0; l < len(layerSizes)-1; l++ {
		size := layerSizes[l] * layerSizes[l+1]
		globalWeights[l] = make([]float64, size)
		for i := 0; i < size; i++ {
			// Xavier initialization
			stddev := math.Sqrt(2.0 / float64(layerSizes[l]+layerSizes[l+1]))
			globalWeights[l][i] = rand.NormFloat64() * stddev
		}
		idx += size
	}

	return &FederatedServer{
		Config:        config,
		GlobalWeights: globalWeights,
		ClientUpdates: make(map[int][]float64),
		ClientSamples: make(map[int]int),
		ServerControl: make([]float64, totalWeights),
		ClientControl: make(map[int][]float64),
		RoundNumber:   0,
	}
}

// AggregateUpdates aggregates client updates using configured method
func (s *FederatedServer) AggregateUpdates() error {
	switch s.Config.AggregationMethod {
	case "fedavg":
		return s.fedAvgAggregate()
	case "fedprox":
		return s.fedProxAggregate()
	case "scaffold":
		return s.scaffoldAggregate()
	case "fednova":
		return s.fedNovaAggregate()
	default:
		return s.fedAvgAggregate()
	}
}

// fedAvgAggregate implements FedAvg aggregation
func (s *FederatedServer) fedAvgAggregate() error {
	totalSamples := 0
	for _, n := range s.ClientSamples {
		totalSamples += n
	}

	if totalSamples == 0 {
		return fmt.Errorf("no samples from clients")
	}

	// Flatten global weights
	flat := s.flattenWeights()

	// Weighted average of client updates
	for clientID, update := range s.ClientUpdates {
		weight := float64(s.ClientSamples[clientID]) / float64(totalSamples)
		for i := range flat {
			if i < len(update) {
				flat[i] += weight * (update[i] - flat[i])
			}
		}
	}

	// Unflatten back to layer structure
	s.unflattenWeights(flat)

	s.RoundNumber++
	return nil
}

// fedProxAggregate implements FedProx aggregation
func (s *FederatedServer) fedProxAggregate() error {
	// Similar to FedAvg but clients use proximal term
	// The proximal term is applied during local training
	return s.fedAvgAggregate()
}

// scaffoldAggregate implements SCAFFOLD aggregation
func (s *FederatedServer) scaffoldAggregate() error {
	// SCAFFOLD uses control variates for variance reduction
	totalSamples := 0
	for _, n := range s.ClientSamples {
		totalSamples += n
	}

	if totalSamples == 0 {
		return fmt.Errorf("no samples from clients")
	}

	flat := s.flattenWeights()
	newControl := make([]float64, len(s.ServerControl))

	// Aggregate model updates and control variates
	for clientID, update := range s.ClientUpdates {
		weight := float64(s.ClientSamples[clientID]) / float64(totalSamples)
		clientCtrl := s.ClientControl[clientID]

		for i := range flat {
			if i < len(update) {
				flat[i] += weight * (update[i] - flat[i])
			}
			if clientCtrl != nil && i < len(clientCtrl) {
				newControl[i] += weight * (clientCtrl[i] - s.ServerControl[i])
			}
		}
	}

	// Update server control variate
	for i := range s.ServerControl {
		s.ServerControl[i] += newControl[i]
	}

	s.unflattenWeights(flat)
	s.RoundNumber++
	return nil
}

// fedNovaAggregate implements FedNova aggregation
func (s *FederatedServer) fedNovaAggregate() error {
	// FedNova normalizes by local steps
	totalNormalized := 0.0
	for _, n := range s.ClientSamples {
		// Assume each sample corresponds to one local step
		totalNormalized += float64(n) / float64(s.Config.LocalEpochs)
	}

	if totalNormalized == 0 {
		return fmt.Errorf("no normalized samples from clients")
	}

	flat := s.flattenWeights()

	for clientID, update := range s.ClientUpdates {
		normalizedWeight := (float64(s.ClientSamples[clientID]) / float64(s.Config.LocalEpochs)) / totalNormalized
		for i := range flat {
			if i < len(update) {
				flat[i] += normalizedWeight * (update[i] - flat[i])
			}
		}
	}

	s.unflattenWeights(flat)
	s.RoundNumber++
	return nil
}

func (s *FederatedServer) flattenWeights() []float64 {
	size := 0
	for _, w := range s.GlobalWeights {
		size += len(w)
	}

	flat := make([]float64, size)
	idx := 0
	for _, w := range s.GlobalWeights {
		copy(flat[idx:], w)
		idx += len(w)
	}
	return flat
}

func (s *FederatedServer) unflattenWeights(flat []float64) {
	idx := 0
	for l := range s.GlobalWeights {
		copy(s.GlobalWeights[l], flat[idx:idx+len(s.GlobalWeights[l])])
		idx += len(s.GlobalWeights[l])
	}
}

// GradientCompressor handles gradient compression strategies
type GradientCompressor struct {
	Config *FederatedConfig

	// Error feedback buffer (for error accumulation)
	ErrorBuffer map[int][]float64

	// Statistics
	CompressionRatios []float64
	BandwidthSaved    float64
}

// NewGradientCompressor creates a gradient compressor
func NewGradientCompressor(config *FederatedConfig) *GradientCompressor {
	return &GradientCompressor{
		Config:            config,
		ErrorBuffer:       make(map[int][]float64),
		CompressionRatios: make([]float64, 0),
	}
}

// Compress applies compression to gradients
func (gc *GradientCompressor) Compress(clientID int, gradients []float64) ([]float64, []int) {
	// Add error feedback from previous round
	if errorBuf, exists := gc.ErrorBuffer[clientID]; exists {
		for i := range gradients {
			if i < len(errorBuf) {
				gradients[i] += errorBuf[i]
			}
		}
	}

	// Top-K sparsification
	k := int(float64(len(gradients)) * gc.Config.TopKSparsity)
	if k < 1 {
		k = 1
	}

	// Find indices of top-K values
	type absIdx struct {
		abs float64
		idx int
	}

	sorted := make([]absIdx, len(gradients))
	for i, g := range gradients {
		sorted[i] = absIdx{math.Abs(g), i}
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].abs > sorted[j].abs
	})

	// Select top-K
	indices := make([]int, k)
	values := make([]float64, k)

	for i := 0; i < k; i++ {
		indices[i] = sorted[i].idx
		values[i] = gradients[sorted[i].idx]
	}

	// Compute error for feedback
	newError := make([]float64, len(gradients))
	copy(newError, gradients)
	for i := 0; i < k; i++ {
		newError[indices[i]] = 0
	}
	gc.ErrorBuffer[clientID] = newError

	// Record compression ratio
	ratio := float64(k) / float64(len(gradients))
	gc.CompressionRatios = append(gc.CompressionRatios, ratio)
	gc.BandwidthSaved += float64(len(gradients)-k) * 4 // 4 bytes per float32

	return values, indices
}

// ============================================================================
// NEUROMORPHIC VISION SENSORS FOR CIM
// ============================================================================

// DVSConfig configures the Dynamic Vision Sensor
type DVSConfig struct {
	// Sensor dimensions
	Width           int     // Pixel width
	Height          int     // Pixel height

	// Temporal parameters
	TemporalResolution float64 // Microseconds
	RefractoryPeriod   float64 // Minimum time between events (μs)

	// Threshold parameters
	ONThreshold     float64 // Positive change threshold (log units)
	OFFThreshold    float64 // Negative change threshold (log units)
	ThresholdNoise  float64 // Threshold variation (%)

	// Latency
	PixelLatency    float64 // Per-pixel processing latency (μs)

	// Power
	PowerPerPixel   float64 // Power consumption per active pixel (μW)

	// Integration with CIM
	DirectToCIM     bool    // Direct spike injection to CIM
}

// DefaultDVSConfig returns default DVS configuration
func DefaultDVSConfig() *DVSConfig {
	return &DVSConfig{
		Width:              346,
		Height:             260,
		TemporalResolution: 1.0,
		RefractoryPeriod:   1.0,
		ONThreshold:        0.15,
		OFFThreshold:       0.15,
		ThresholdNoise:     0.03,
		PixelLatency:       15.0,
		PowerPerPixel:      0.1,
		DirectToCIM:        true,
	}
}

// DVSEvent represents a single DVS event
type DVSEvent struct {
	X         int     // X coordinate
	Y         int     // Y coordinate
	Timestamp float64 // Event timestamp (μs)
	Polarity  int     // +1 for ON, -1 for OFF
}

// DVSEventCamera simulates a Dynamic Vision Sensor
type DVSEventCamera struct {
	Config       *DVSConfig

	// Pixel state
	LogIntensity [][]float64 // Log intensity at each pixel
	LastEventTime [][]float64 // Time of last event per pixel

	// Thresholds with per-pixel noise
	PixelThresholds [][]float64

	// Event buffer
	EventBuffer []*DVSEvent
	BufferMutex sync.Mutex

	// Statistics
	TotalEvents   int
	EventRate     float64 // Events per second
	PowerConsumption float64
}

// NewDVSEventCamera creates a new DVS camera
func NewDVSEventCamera(config *DVSConfig) *DVSEventCamera {
	logInt := make([][]float64, config.Height)
	lastEvent := make([][]float64, config.Height)
	thresholds := make([][]float64, config.Height)

	for y := 0; y < config.Height; y++ {
		logInt[y] = make([]float64, config.Width)
		lastEvent[y] = make([]float64, config.Width)
		thresholds[y] = make([]float64, config.Width)

		for x := 0; x < config.Width; x++ {
			logInt[y][x] = 0.5 // Initialize to mid-gray
			lastEvent[y][x] = -1000.0 // No recent events
			// Per-pixel threshold with noise
			noise := 1.0 + config.ThresholdNoise*(rand.Float64()*2-1)
			thresholds[y][x] = config.ONThreshold * noise
		}
	}

	return &DVSEventCamera{
		Config:          config,
		LogIntensity:    logInt,
		LastEventTime:   lastEvent,
		PixelThresholds: thresholds,
		EventBuffer:     make([]*DVSEvent, 0),
	}
}

// ProcessFrame processes a frame and generates events
func (dvs *DVSEventCamera) ProcessFrame(frame [][]float64, timestamp float64) []*DVSEvent {
	events := make([]*DVSEvent, 0)

	for y := 0; y < dvs.Config.Height && y < len(frame); y++ {
		for x := 0; x < dvs.Config.Width && x < len(frame[y]); x++ {
			// Convert to log intensity
			intensity := frame[y][x]
			if intensity < 0.001 {
				intensity = 0.001
			}
			newLogInt := math.Log(intensity)

			// Check if refractory period has passed
			if timestamp-dvs.LastEventTime[y][x] < dvs.Config.RefractoryPeriod {
				continue
			}

			// Compute change
			change := newLogInt - dvs.LogIntensity[y][x]

			// Check for events
			if change >= dvs.PixelThresholds[y][x] {
				// ON event
				event := &DVSEvent{
					X:         x,
					Y:         y,
					Timestamp: timestamp + dvs.Config.PixelLatency,
					Polarity:  1,
				}
				events = append(events, event)
				dvs.LogIntensity[y][x] = newLogInt
				dvs.LastEventTime[y][x] = timestamp
				dvs.TotalEvents++
			} else if -change >= dvs.PixelThresholds[y][x] {
				// OFF event
				event := &DVSEvent{
					X:         x,
					Y:         y,
					Timestamp: timestamp + dvs.Config.PixelLatency,
					Polarity:  -1,
				}
				events = append(events, event)
				dvs.LogIntensity[y][x] = newLogInt
				dvs.LastEventTime[y][x] = timestamp
				dvs.TotalEvents++
			}
		}
	}

	// Update statistics
	dvs.PowerConsumption = float64(len(events)) * dvs.Config.PowerPerPixel

	dvs.BufferMutex.Lock()
	dvs.EventBuffer = append(dvs.EventBuffer, events...)
	dvs.BufferMutex.Unlock()

	return events
}

// GetEventStream returns accumulated events and clears buffer
func (dvs *DVSEventCamera) GetEventStream() []*DVSEvent {
	dvs.BufferMutex.Lock()
	defer dvs.BufferMutex.Unlock()

	events := dvs.EventBuffer
	dvs.EventBuffer = make([]*DVSEvent, 0)
	return events
}

// SCIMITARConfig configures the SCIMITAR architecture
type SCIMITARConfig struct {
	// Array dimensions
	NumRows       int     // Crossbar rows
	NumCols       int     // Crossbar columns

	// Sparsity exploitation
	InputSparsity  float64 // Expected input sparsity (0-1)
	WeightSparsity float64 // Weight sparsity (0-1)

	// Processing
	SkipZeroRows  bool    // Skip zero input rows
	SkipZeroCols  bool    // Skip zero weight columns

	// Timing
	ClockFreqMHz  float64 // Operating frequency

	// Energy
	EnergyPerMAC  float64 // Energy per MAC operation (fJ)
}

// DefaultSCIMITARConfig returns default SCIMITAR configuration
func DefaultSCIMITARConfig() *SCIMITARConfig {
	return &SCIMITARConfig{
		NumRows:        64,
		NumCols:        64,
		InputSparsity:  0.9,
		WeightSparsity: 0.5,
		SkipZeroRows:   true,
		SkipZeroCols:   true,
		ClockFreqMHz:   200.0,
		EnergyPerMAC:   50.0,
	}
}

// SCIMITARCore implements sparse CIM processing
type SCIMITARCore struct {
	Config      *SCIMITARConfig

	// Weight storage (sparse representation)
	WeightCSR   *SparseMatrix // Compressed Sparse Row

	// Processing statistics
	SkippedMACs  int
	TotalMACs    int
	EnergyUsed   float64 // fJ
	Latency      float64 // ns
}

// SparseMatrix represents weights in CSR format
type SparseMatrix struct {
	Values    []float64
	ColIdx    []int
	RowPtr    []int
	NumRows   int
	NumCols   int
	NNZ       int // Number of non-zeros
}

// NewSparseMatrix creates CSR matrix from dense
func NewSparseMatrix(dense [][]float64, threshold float64) *SparseMatrix {
	numRows := len(dense)
	numCols := 0
	if numRows > 0 {
		numCols = len(dense[0])
	}

	values := make([]float64, 0)
	colIdx := make([]int, 0)
	rowPtr := make([]int, numRows+1)

	rowPtr[0] = 0
	for i, row := range dense {
		for j, val := range row {
			if math.Abs(val) > threshold {
				values = append(values, val)
				colIdx = append(colIdx, j)
			}
		}
		rowPtr[i+1] = len(values)
	}

	return &SparseMatrix{
		Values:  values,
		ColIdx:  colIdx,
		RowPtr:  rowPtr,
		NumRows: numRows,
		NumCols: numCols,
		NNZ:     len(values),
	}
}

// NewSCIMITARCore creates a SCIMITAR core
func NewSCIMITARCore(config *SCIMITARConfig) *SCIMITARCore {
	return &SCIMITARCore{
		Config: config,
	}
}

// LoadWeights loads sparse weights
func (sc *SCIMITARCore) LoadWeights(dense [][]float64) {
	sc.WeightCSR = NewSparseMatrix(dense, 1e-6)
}

// Compute performs sparse MVM
func (sc *SCIMITARCore) Compute(input []float64) []float64 {
	if sc.WeightCSR == nil {
		return nil
	}

	output := make([]float64, sc.WeightCSR.NumCols)
	sc.SkippedMACs = 0
	sc.TotalMACs = 0

	// Sparse input × sparse weight computation
	for i := 0; i < len(input) && i < sc.WeightCSR.NumRows; i++ {
		// Skip zero inputs if enabled
		if sc.Config.SkipZeroRows && math.Abs(input[i]) < 1e-6 {
			// Count skipped MACs
			sc.SkippedMACs += sc.WeightCSR.RowPtr[i+1] - sc.WeightCSR.RowPtr[i]
			continue
		}

		// Process non-zero weights in this row
		for k := sc.WeightCSR.RowPtr[i]; k < sc.WeightCSR.RowPtr[i+1]; k++ {
			j := sc.WeightCSR.ColIdx[k]
			w := sc.WeightCSR.Values[k]

			output[j] += input[i] * w
			sc.TotalMACs++
		}
	}

	// Calculate energy and latency
	sc.EnergyUsed = float64(sc.TotalMACs) * sc.Config.EnergyPerMAC
	cyclesPerMAC := 1.0 // Assuming 1 cycle per MAC
	sc.Latency = float64(sc.TotalMACs) * cyclesPerMAC * (1000.0 / sc.Config.ClockFreqMHz)

	return output
}

// InSensorConfig configures in-sensor computing
type InSensorConfig struct {
	// Sensor array
	SensorRows    int     // Photodetector rows
	SensorCols    int     // Photodetector columns

	// Integration
	IntegrationTime float64 // Exposure time (ms)

	// CIM integration
	DirectCompute bool    // Compute in photocurrent domain
	NumConvKernels int    // Convolution kernels in sensor
	KernelSize    int     // Kernel dimension

	// Power
	SensorPower   float64 // Sensor array power (mW)
	ComputePower  float64 // In-sensor compute power (mW)
}

// DefaultInSensorConfig returns default in-sensor configuration
func DefaultInSensorConfig() *InSensorConfig {
	return &InSensorConfig{
		SensorRows:      256,
		SensorCols:      256,
		IntegrationTime: 33.0,
		DirectCompute:   true,
		NumConvKernels:  8,
		KernelSize:      3,
		SensorPower:     10.0,
		ComputePower:    5.0,
	}
}

// InSensorProcessor performs computation at the sensor level
type InSensorProcessor struct {
	Config *InSensorConfig

	// Convolution kernels stored in analog domain
	Kernels [][][]float64 // [numKernels][height][width]

	// Photocurrent accumulation
	PhotoCurrents [][]float64

	// Output feature maps
	FeatureMaps [][][]float64
}

// NewInSensorProcessor creates an in-sensor processor
func NewInSensorProcessor(config *InSensorConfig) *InSensorProcessor {
	// Initialize random kernels (would be trained)
	kernels := make([][][]float64, config.NumConvKernels)
	for k := 0; k < config.NumConvKernels; k++ {
		kernels[k] = make([][]float64, config.KernelSize)
		for i := 0; i < config.KernelSize; i++ {
			kernels[k][i] = make([]float64, config.KernelSize)
			for j := 0; j < config.KernelSize; j++ {
				kernels[k][i][j] = rand.NormFloat64() * 0.1
			}
		}
	}

	return &InSensorProcessor{
		Config:  config,
		Kernels: kernels,
	}
}

// ProcessImage performs in-sensor convolution
func (isp *InSensorProcessor) ProcessImage(image [][]float64) [][][]float64 {
	outH := len(image) - isp.Config.KernelSize + 1
	outW := len(image[0]) - isp.Config.KernelSize + 1

	// Initialize output feature maps
	isp.FeatureMaps = make([][][]float64, isp.Config.NumConvKernels)

	for k := 0; k < isp.Config.NumConvKernels; k++ {
		isp.FeatureMaps[k] = make([][]float64, outH)
		for i := 0; i < outH; i++ {
			isp.FeatureMaps[k][i] = make([]float64, outW)
		}
	}

	// Convolution in photocurrent domain
	for k := 0; k < isp.Config.NumConvKernels; k++ {
		for i := 0; i < outH; i++ {
			for j := 0; j < outW; j++ {
				sum := 0.0
				for ki := 0; ki < isp.Config.KernelSize; ki++ {
					for kj := 0; kj < isp.Config.KernelSize; kj++ {
						sum += image[i+ki][j+kj] * isp.Kernels[k][ki][kj]
					}
				}
				// Photocurrent accumulation with noise
				noise := rand.NormFloat64() * 0.01
				isp.FeatureMaps[k][i][j] = sum + noise
			}
		}
	}

	return isp.FeatureMaps
}

// NeuromorphicVisionPipeline integrates DVS with CIM processing
type NeuromorphicVisionPipeline struct {
	DVS         *DVSEventCamera
	SCIMITAR    *SCIMITARCore
	InSensor    *InSensorProcessor

	// Event processing
	EventWindow  float64        // Time window for event accumulation (ms)
	EventFrame   [][]float64    // Accumulated event frame

	// Processing chain
	LayerWeights [][][]float64  // Conv/FC weights

	// Statistics
	FrameRate    float64
	TotalLatency float64
	TotalEnergy  float64
}

// NewNeuromorphicVisionPipeline creates the pipeline
func NewNeuromorphicVisionPipeline(dvsConfig *DVSConfig, scimitarConfig *SCIMITARConfig) *NeuromorphicVisionPipeline {
	return &NeuromorphicVisionPipeline{
		DVS:         NewDVSEventCamera(dvsConfig),
		SCIMITAR:    NewSCIMITARCore(scimitarConfig),
		EventWindow: 10.0,
		EventFrame:  make([][]float64, dvsConfig.Height),
	}
}

// ProcessEventWindow accumulates events into a frame
func (nvp *NeuromorphicVisionPipeline) ProcessEventWindow(events []*DVSEvent) [][]float64 {
	// Reset event frame
	for y := range nvp.EventFrame {
		if nvp.EventFrame[y] == nil {
			nvp.EventFrame[y] = make([]float64, nvp.DVS.Config.Width)
		}
		for x := range nvp.EventFrame[y] {
			nvp.EventFrame[y][x] = 0
		}
	}

	// Accumulate events
	for _, event := range events {
		if event.Y < len(nvp.EventFrame) && event.X < len(nvp.EventFrame[event.Y]) {
			nvp.EventFrame[event.Y][event.X] += float64(event.Polarity)
		}
	}

	return nvp.EventFrame
}

// ============================================================================
// INTEGRATED IRONLATTICE FEDERATED VISION SYSTEM
// ============================================================================

// IronLatticeFederatedVision integrates federated learning with neuromorphic vision
type IronLatticeFederatedVision struct {
	// Configuration
	FedConfig    *FederatedConfig
	DVSConfig    *DVSConfig
	SCIMConfig   *SCIMITARConfig

	// Components
	Server       *FederatedServer
	Clients      []*FederatedClient
	VisionPipelines []*NeuromorphicVisionPipeline
	Compressor   *GradientCompressor

	// Integration
	FeFETWeightStorage [][]*FeFETWeight

	// Statistics
	GlobalRound     int
	AverageAccuracy float64
	TotalEnergy     float64
	CommunicationCost float64
}

// NewIronLatticeFederatedVision creates the integrated system
func NewIronLatticeFederatedVision(
	fedConfig *FederatedConfig,
	dvsConfig *DVSConfig,
	scimitarConfig *SCIMITARConfig,
	layerSizes []int,
) *IronLatticeFederatedVision {

	// Create server
	server := NewFederatedServer(fedConfig, layerSizes)

	// Create clients with vision pipelines
	clients := make([]*FederatedClient, fedConfig.NumClients)
	pipelines := make([]*NeuromorphicVisionPipeline, fedConfig.NumClients)

	for i := 0; i < fedConfig.NumClients; i++ {
		clients[i] = NewFederatedClient(fedConfig, i, layerSizes)
		pipelines[i] = NewNeuromorphicVisionPipeline(dvsConfig, scimitarConfig)
	}

	// Create gradient compressor
	compressor := NewGradientCompressor(fedConfig)

	return &IronLatticeFederatedVision{
		FedConfig:       fedConfig,
		DVSConfig:       dvsConfig,
		SCIMConfig:      scimitarConfig,
		Server:          server,
		Clients:         clients,
		VisionPipelines: pipelines,
		Compressor:      compressor,
	}
}

// RunFederatedRound executes one round of federated learning
func (ilfv *IronLatticeFederatedVision) RunFederatedRound(
	clientData map[int][][]float64,
	clientLabels map[int][]int,
) error {
	// Select subset of clients
	numSelected := int(float64(len(ilfv.Clients)) * ilfv.FedConfig.ClientFraction)
	if numSelected < 1 {
		numSelected = 1
	}

	selectedClients := rand.Perm(len(ilfv.Clients))[:numSelected]

	// Clear previous updates
	ilfv.Server.ClientUpdates = make(map[int][]float64)
	ilfv.Server.ClientSamples = make(map[int]int)

	// Local training on selected clients (parallel in practice)
	for _, clientIdx := range selectedClients {
		client := ilfv.Clients[clientIdx]

		// Get client's data
		data := clientData[clientIdx]
		labels := clientLabels[clientIdx]

		if len(data) == 0 {
			continue
		}

		// Local training
		if err := client.LocalTrain(data, labels); err != nil {
			continue
		}

		// Compress and send gradients
		compressed := client.CompressGradients()

		// Apply differential privacy if enabled
		if ilfv.FedConfig.DifferentialPrivacy {
			compressed = ilfv.applyDifferentialPrivacy(compressed)
		}

		// Store update at server
		ilfv.Server.ClientUpdates[clientIdx] = compressed
		ilfv.Server.ClientSamples[clientIdx] = client.NumSamples

		// Track communication cost
		ilfv.CommunicationCost += float64(len(compressed)) * 4 // 4 bytes per float32
	}

	// Aggregate at server
	if err := ilfv.Server.AggregateUpdates(); err != nil {
		return err
	}

	// Distribute updated model to all clients
	ilfv.distributeGlobalModel()

	ilfv.GlobalRound++
	return nil
}

// applyDifferentialPrivacy adds DP noise to gradients
func (ilfv *IronLatticeFederatedVision) applyDifferentialPrivacy(gradients []float64) []float64 {
	// Clip gradients
	norm := 0.0
	for _, g := range gradients {
		norm += g * g
	}
	norm = math.Sqrt(norm)

	clipFactor := 1.0
	if norm > ilfv.FedConfig.DPClipNorm {
		clipFactor = ilfv.FedConfig.DPClipNorm / norm
	}

	// Add Gaussian noise
	sensitivity := ilfv.FedConfig.DPClipNorm
	sigma := sensitivity * math.Sqrt(2*math.Log(1.25/ilfv.FedConfig.DPDelta)) / ilfv.FedConfig.DPEpsilon

	result := make([]float64, len(gradients))
	for i, g := range gradients {
		result[i] = g*clipFactor + rand.NormFloat64()*sigma
	}

	return result
}

// distributeGlobalModel sends global model to all clients
func (ilfv *IronLatticeFederatedVision) distributeGlobalModel() {
	globalFlat := ilfv.Server.flattenWeights()

	for _, client := range ilfv.Clients {
		// Update FeFET weights
		idx := 0
		for l := range client.Weights {
			for j := range client.Weights[l] {
				if idx < len(globalFlat) {
					// Program FeFET to new weight value
					targetPol := globalFlat[idx]
					client.Weights[l][j].Polarization = targetPol
					client.Weights[l][j].Conductance = 50.0 + 49.0*targetPol
				}
				idx++
			}
		}
	}
}

// ProcessVisionInput processes DVS events and runs inference
func (ilfv *IronLatticeFederatedVision) ProcessVisionInput(
	clientIdx int,
	frame [][]float64,
	timestamp float64,
) ([]float64, error) {
	if clientIdx >= len(ilfv.VisionPipelines) {
		return nil, fmt.Errorf("invalid client index: %d", clientIdx)
	}

	pipeline := ilfv.VisionPipelines[clientIdx]

	// Generate DVS events from frame
	events := pipeline.DVS.ProcessFrame(frame, timestamp)

	// Accumulate events into frame
	eventFrame := pipeline.ProcessEventWindow(events)

	// Flatten for SCIMITAR processing
	flat := make([]float64, 0)
	for _, row := range eventFrame {
		flat = append(flat, row...)
	}

	// Load client weights into SCIMITAR
	client := ilfv.Clients[clientIdx]
	if len(client.Weights) > 0 {
		denseWeights := make([][]float64, len(flat))
		for i := range denseWeights {
			denseWeights[i] = make([]float64, len(client.Weights[0]))
			for j := range denseWeights[i] {
				if i*len(denseWeights[i])+j < len(client.Weights[0]) {
					denseWeights[i][j] = client.Weights[0][i*len(denseWeights[i])+j].Conductance / 50.0
				}
			}
		}
		pipeline.SCIMITAR.LoadWeights(denseWeights)
	}

	// Run sparse inference
	output := pipeline.SCIMITAR.Compute(flat)

	// Update energy statistics
	ilfv.TotalEnergy += pipeline.SCIMITAR.EnergyUsed
	ilfv.TotalEnergy += pipeline.DVS.PowerConsumption * 1000.0 // Convert mW to fJ equivalent

	return output, nil
}

// GetStatistics returns system statistics
func (ilfv *IronLatticeFederatedVision) GetStatistics() map[string]float64 {
	avgCompressionRatio := 0.0
	if len(ilfv.Compressor.CompressionRatios) > 0 {
		for _, r := range ilfv.Compressor.CompressionRatios {
			avgCompressionRatio += r
		}
		avgCompressionRatio /= float64(len(ilfv.Compressor.CompressionRatios))
	}

	return map[string]float64{
		"global_round":            float64(ilfv.GlobalRound),
		"total_energy_fJ":         ilfv.TotalEnergy,
		"communication_cost_KB":   ilfv.CommunicationCost / 1024.0,
		"avg_compression_ratio":   avgCompressionRatio,
		"bandwidth_saved_KB":      ilfv.Compressor.BandwidthSaved / 1024.0,
		"num_clients":             float64(len(ilfv.Clients)),
		"dvs_total_events":        float64(ilfv.VisionPipelines[0].DVS.TotalEvents),
	}
}

// Preset configurations for common scenarios
func IronLatticeFederatedVisionPreset(scenario string) (*FederatedConfig, *DVSConfig, *SCIMITARConfig) {
	switch scenario {
	case "edge_surveillance":
		// Low-power edge deployment for surveillance
		return &FederatedConfig{
				NumClients:          10,
				LocalEpochs:         3,
				BatchSize:           16,
				CommunicationRounds: 50,
				AggregationMethod:   "fedavg",
				ClientFraction:      0.3,
				CompressionEnabled:  true,
				TopKSparsity:        0.01,
				OnDeviceTraining:    true,
				MaxEndurance:        1e9,
			},
			&DVSConfig{
				Width:              128,
				Height:             128,
				TemporalResolution: 10.0,
				ONThreshold:        0.2,
				DirectToCIM:        true,
			},
			&SCIMITARConfig{
				NumRows:       32,
				NumCols:       32,
				SkipZeroRows:  true,
				EnergyPerMAC:  30.0,
			}

	case "autonomous_vehicle":
		// High-performance for autonomous driving
		return &FederatedConfig{
				NumClients:          100,
				LocalEpochs:         10,
				BatchSize:           64,
				CommunicationRounds: 200,
				AggregationMethod:   "scaffold",
				ClientFraction:      0.2,
				CompressionEnabled:  true,
				TopKSparsity:        0.05,
				DifferentialPrivacy: true,
				DPEpsilon:           2.0,
				OnDeviceTraining:    true,
				MaxEndurance:        1e10,
			},
			&DVSConfig{
				Width:              346,
				Height:             260,
				TemporalResolution: 1.0,
				ONThreshold:        0.15,
				DirectToCIM:        true,
			},
			&SCIMITARConfig{
				NumRows:       128,
				NumCols:       128,
				SkipZeroRows:  true,
				EnergyPerMAC:  50.0,
			}

	case "medical_imaging":
		// Privacy-critical medical imaging
		return &FederatedConfig{
				NumClients:          50,
				LocalEpochs:         5,
				BatchSize:           8,
				CommunicationRounds: 100,
				AggregationMethod:   "fedprox",
				FedProxMu:           0.1,
				ClientFraction:      0.1,
				CompressionEnabled:  false, // Full precision for accuracy
				DifferentialPrivacy: true,
				DPEpsilon:           0.5, // Strong privacy
				DPClipNorm:          0.5,
				OnDeviceTraining:    true,
				MaxEndurance:        1e10,
			},
			DefaultDVSConfig(),
			DefaultSCIMITARConfig()

	default:
		return DefaultFederatedConfig(), DefaultDVSConfig(), DefaultSCIMITARConfig()
	}
}
