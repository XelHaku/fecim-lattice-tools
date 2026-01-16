// hil_multichip.go - Hardware-in-the-loop testing and multi-chip CIM orchestration
// Implements CiMLoop-style full-stack modeling and NeuronLink-style chiplet interconnects
// Based on research: ISPASS 2024 Best Paper, DAC 2024/2025 chiplet architectures

package layers

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// ============================================================================
// Hardware-in-the-Loop Testing Framework
// ============================================================================

// HILMode defines the hardware-in-the-loop simulation mode
type HILMode int

const (
	HILModeSimulation  HILMode = iota // Pure software simulation
	HILModeEmulation                  // FPGA emulation
	HILModeHybrid                     // Mixed simulation/hardware
	HILModeHardware                   // Real hardware testing
)

// StackLevel defines levels of the CIM design stack
type StackLevel int

const (
	StackLevelDevice    StackLevel = iota // Device physics
	StackLevelCircuit                     // Circuit design
	StackLevelArray                       // Crossbar array
	StackLevelTile                        // Tile organization
	StackLevelChip                        // Chip architecture
	StackLevelSystem                      // Multi-chip system
	StackLevelWorkload                    // DNN workload
)

// DeviceModel represents a device-level model
type DeviceModel struct {
	Type           string  // "RRAM", "PCM", "FeFET", "SRAM"
	ROn            float64 // On-resistance (Ohms)
	ROff           float64 // Off-resistance (Ohms)
	ReadVoltage    float64 // Read voltage (V)
	WriteVoltage   float64 // Write/program voltage (V)
	SwitchingTime  float64 // Switching time (ns)
	Endurance      float64 // Cycles before failure
	Retention      float64 // Retention time (years)
	EnergyPerWrite float64 // Energy per write (pJ)
	EnergyPerRead  float64 // Energy per read (fJ)
	Variability    float64 // Device-to-device variation (%)
}

// HZODeviceModel returns IronLattice HZO FeFET parameters
func HZODeviceModel() DeviceModel {
	return DeviceModel{
		Type:           "HZO-FeFET",
		ROn:            10e3,        // 10 kΩ
		ROff:           10e6,        // 10 MΩ
		ReadVoltage:    0.5,         // V
		WriteVoltage:   3.0,         // V
		SwitchingTime:  10,          // ns
		Endurance:      1e12,        // cycles
		Retention:      10,          // years
		EnergyPerWrite: 0.5,         // pJ
		EnergyPerRead:  10,          // fJ
		Variability:    3.0,         // %
	}
}

// CircuitModel represents circuit-level parameters
type CircuitModel struct {
	ADCBits         int     // ADC resolution
	DACBits         int     // DAC resolution
	ADCLatency      float64 // ADC conversion time (ns)
	ADCEnergy       float64 // Energy per ADC conversion (pJ)
	SenseAmpGain    float64 // Sense amplifier gain
	WireResistance  float64 // Wire resistance (Ohms/um)
	LineCapacitance float64 // Line capacitance (fF/um)
}

// DefaultCircuitModel returns typical circuit parameters
func DefaultCircuitModel() CircuitModel {
	return CircuitModel{
		ADCBits:         6,
		DACBits:         8,
		ADCLatency:      10,  // ns
		ADCEnergy:       0.5, // pJ
		SenseAmpGain:    100,
		WireResistance:  0.1,  // Ohms/um
		LineCapacitance: 0.2, // fF/um
	}
}

// ArrayModel represents crossbar array parameters
type ArrayModel struct {
	Rows           int
	Cols           int
	CellPitch      float64 // um
	DeviceModel    DeviceModel
	CircuitModel   CircuitModel
	IRDropEnabled  bool
	SnowballEffect bool // Read disturb accumulation
}

// NewArrayModel creates a new array model
func NewArrayModel(rows, cols int, device DeviceModel) *ArrayModel {
	return &ArrayModel{
		Rows:          rows,
		Cols:          cols,
		CellPitch:     0.1, // 100nm
		DeviceModel:   device,
		CircuitModel:  DefaultCircuitModel(),
		IRDropEnabled: true,
	}
}

// HILTestConfig configures hardware-in-the-loop testing
type HILTestConfig struct {
	Mode              HILMode
	StackLevels       []StackLevel
	StatisticalModel  bool    // Use statistical approximation
	NumMonteCarlo     int     // Monte Carlo iterations
	ConfidenceLevel   float64 // Statistical confidence level
	TargetAccuracy    float64 // Target DNN accuracy
	MaxLatencyMs      float64 // Maximum latency constraint
	MaxEnergyMJ       float64 // Maximum energy constraint
}

// DefaultHILConfig returns default HIL configuration
func DefaultHILConfig() HILTestConfig {
	return HILTestConfig{
		Mode:             HILModeSimulation,
		StackLevels:      []StackLevel{StackLevelDevice, StackLevelCircuit, StackLevelArray},
		StatisticalModel: true,
		NumMonteCarlo:    1000,
		ConfidenceLevel:  0.95,
		TargetAccuracy:   0.95,
		MaxLatencyMs:     10,
		MaxEnergyMJ:      1,
	}
}

// HILTestResult holds test results
type HILTestResult struct {
	TestName       string
	Passed         bool
	Accuracy       float64
	AccuracyStdDev float64
	Latency        float64 // ms
	LatencyStdDev  float64
	Energy         float64 // mJ
	EnergyStdDev   float64
	Throughput     float64 // TOPS
	AreaMm2        float64
	Efficiency     float64 // TOPS/W
	ErrorSources   map[string]float64
	Timestamp      time.Time
}

// HILTestFramework implements hardware-in-the-loop testing
type HILTestFramework struct {
	Config         HILTestConfig
	DeviceModel    DeviceModel
	CircuitModel   CircuitModel
	ArrayModels    []*ArrayModel
	Results        []HILTestResult
	DesignPoints   []DesignPoint
	mu             sync.Mutex
}

// DesignPoint represents a point in the design space
type DesignPoint struct {
	ArraySize      int
	ADCBits        int
	WeightBits     int
	InputBits      int
	NoiseLevel     float64
	EstimatedEnergy float64
	EstimatedLatency float64
	EstimatedAccuracy float64
}

// NewHILTestFramework creates a new HIL test framework
func NewHILTestFramework(config HILTestConfig) *HILTestFramework {
	return &HILTestFramework{
		Config:       config,
		DeviceModel:  HZODeviceModel(),
		CircuitModel: DefaultCircuitModel(),
		ArrayModels:  make([]*ArrayModel, 0),
		Results:      make([]HILTestResult, 0),
		DesignPoints: make([]DesignPoint, 0),
	}
}

// AddArrayModel adds an array model to the framework
func (htf *HILTestFramework) AddArrayModel(model *ArrayModel) {
	htf.mu.Lock()
	defer htf.mu.Unlock()
	htf.ArrayModels = append(htf.ArrayModels, model)
}

// RunStatisticalTest runs Monte Carlo simulation
func (htf *HILTestFramework) RunStatisticalTest(testName string, weights []float64) HILTestResult {
	htf.mu.Lock()
	defer htf.mu.Unlock()

	n := htf.Config.NumMonteCarlo
	accuracies := make([]float64, n)
	latencies := make([]float64, n)
	energies := make([]float64, n)

	for i := 0; i < n; i++ {
		// Inject device variability
		noisyWeights := htf.injectVariability(weights)

		// Simulate inference with noise
		acc := htf.simulateInference(noisyWeights)
		lat := htf.estimateLatency()
		eng := htf.estimateEnergy(len(weights))

		accuracies[i] = acc
		latencies[i] = lat
		energies[i] = eng
	}

	// Compute statistics
	accMean, accStd := computeStats(accuracies)
	latMean, latStd := computeStats(latencies)
	engMean, engStd := computeStats(energies)

	result := HILTestResult{
		TestName:       testName,
		Passed:         accMean >= htf.Config.TargetAccuracy,
		Accuracy:       accMean,
		AccuracyStdDev: accStd,
		Latency:        latMean,
		LatencyStdDev:  latStd,
		Energy:         engMean,
		EnergyStdDev:   engStd,
		Throughput:     1000 / latMean, // ops/sec
		ErrorSources: map[string]float64{
			"device_variation": htf.DeviceModel.Variability,
			"adc_quantization": 100 / math.Pow(2, float64(htf.CircuitModel.ADCBits)),
			"thermal_noise":    0.5,
		},
		Timestamp: time.Now(),
	}

	htf.Results = append(htf.Results, result)
	return result
}

func (htf *HILTestFramework) injectVariability(weights []float64) []float64 {
	noisy := make([]float64, len(weights))
	variability := htf.DeviceModel.Variability / 100

	for i, w := range weights {
		noise := 1 + (rand.Float64()*2-1)*variability
		noisy[i] = w * noise
	}
	return noisy
}

func (htf *HILTestFramework) simulateInference(weights []float64) float64 {
	// Simplified inference simulation
	// Returns simulated accuracy based on noise level
	noiseImpact := htf.DeviceModel.Variability / 100
	baseAccuracy := 0.98
	return baseAccuracy * (1 - noiseImpact*0.5)
}

func (htf *HILTestFramework) estimateLatency() float64 {
	// Estimate latency based on array and ADC parameters
	adcLatency := htf.CircuitModel.ADCLatency // ns
	if len(htf.ArrayModels) > 0 {
		rows := htf.ArrayModels[0].Rows
		return float64(rows) * adcLatency / 1e6 // ms
	}
	return 1.0
}

func (htf *HILTestFramework) estimateEnergy(numWeights int) float64 {
	// Estimate energy based on device and circuit parameters
	readEnergy := htf.DeviceModel.EnergyPerRead // fJ
	adcEnergy := htf.CircuitModel.ADCEnergy     // pJ
	totalEnergy := float64(numWeights)*readEnergy/1e6 + adcEnergy/1e3
	return totalEnergy // mJ
}

// ExploreDesignSpace explores the CIM design space
func (htf *HILTestFramework) ExploreDesignSpace() []DesignPoint {
	htf.mu.Lock()
	defer htf.mu.Unlock()

	htf.DesignPoints = make([]DesignPoint, 0)

	// Sweep design parameters
	arraySizes := []int{32, 64, 128, 256}
	adcBits := []int{4, 6, 8}
	weightBits := []int{4, 6, 8}
	noiseLevels := []float64{1, 2, 5, 10}

	for _, size := range arraySizes {
		for _, adc := range adcBits {
			for _, wbits := range weightBits {
				for _, noise := range noiseLevels {
					dp := DesignPoint{
						ArraySize:  size,
						ADCBits:    adc,
						WeightBits: wbits,
						InputBits:  8,
						NoiseLevel: noise,
					}

					// Estimate metrics
					dp.EstimatedEnergy = htf.estimateDesignPointEnergy(dp)
					dp.EstimatedLatency = htf.estimateDesignPointLatency(dp)
					dp.EstimatedAccuracy = htf.estimateDesignPointAccuracy(dp)

					htf.DesignPoints = append(htf.DesignPoints, dp)
				}
			}
		}
	}

	return htf.DesignPoints
}

func (htf *HILTestFramework) estimateDesignPointEnergy(dp DesignPoint) float64 {
	// Energy scales with array size and ADC bits
	baseEnergy := 0.1 // mJ
	arrayFactor := float64(dp.ArraySize*dp.ArraySize) / (64 * 64)
	adcFactor := math.Pow(2, float64(dp.ADCBits-6))
	return baseEnergy * arrayFactor * adcFactor
}

func (htf *HILTestFramework) estimateDesignPointLatency(dp DesignPoint) float64 {
	// Latency scales with array size
	baseLatency := 1.0 // ms
	arrayFactor := float64(dp.ArraySize) / 64
	return baseLatency * arrayFactor
}

func (htf *HILTestFramework) estimateDesignPointAccuracy(dp DesignPoint) float64 {
	// Accuracy decreases with noise and quantization
	baseAcc := 0.98
	noisePenalty := dp.NoiseLevel / 100
	quantPenalty := (8 - float64(dp.WeightBits)) * 0.01
	return baseAcc - noisePenalty - quantPenalty
}

// FindParetoOptimal finds Pareto-optimal design points
func (htf *HILTestFramework) FindParetoOptimal() []DesignPoint {
	htf.mu.Lock()
	defer htf.mu.Unlock()

	pareto := make([]DesignPoint, 0)

	for _, dp := range htf.DesignPoints {
		dominated := false

		for _, other := range htf.DesignPoints {
			if other.EstimatedEnergy < dp.EstimatedEnergy &&
				other.EstimatedLatency < dp.EstimatedLatency &&
				other.EstimatedAccuracy > dp.EstimatedAccuracy {
				dominated = true
				break
			}
		}

		if !dominated {
			pareto = append(pareto, dp)
		}
	}

	return pareto
}

// computeStats computes mean and standard deviation
func computeStats(data []float64) (mean, stdDev float64) {
	n := float64(len(data))
	if n == 0 {
		return 0, 0
	}

	sum := 0.0
	for _, v := range data {
		sum += v
	}
	mean = sum / n

	sumSq := 0.0
	for _, v := range data {
		sumSq += (v - mean) * (v - mean)
	}
	stdDev = math.Sqrt(sumSq / n)

	return mean, stdDev
}

// ============================================================================
// Multi-Chip CIM Orchestration
// ============================================================================

// ChipletType defines types of chiplets
type ChipletType int

const (
	ChipletTypeCIM     ChipletType = iota // Compute-in-memory chiplet
	ChipletTypeDigital                    // Digital accelerator chiplet
	ChipletTypeMemory                     // Memory chiplet (HBM, etc.)
	ChipletTypeIO                         // I/O chiplet
	ChipletTypePhotonic                   // Photonic interconnect chiplet
)

// InterconnectType defines chip-to-chip interconnect types
type InterconnectType int

const (
	InterconnectTypeD2D    InterconnectType = iota // Die-to-die (on-interposer)
	InterconnectTypePCIe                           // PCIe link
	InterconnectTypeUCIe                           // UCIe standard
	InterconnectTypeNVLink                         // NVIDIA NVLink-style
	InterconnectTypePhotonic                       // Optical interconnect
	InterconnectTypeWireless                       // mm-Wave wireless
)

// ChipletSpec defines chiplet specifications
type ChipletSpec struct {
	ID              int
	Type            ChipletType
	Name            string
	ArrayRows       int
	ArrayCols       int
	NumArrays       int
	ProcessNode     int     // nm
	AreaMm2         float64
	PowerW          float64
	PeakTOPS        float64
	MemoryMB        int
	InterconnectBW  float64 // GB/s
}

// NewCIMChiplet creates a CIM chiplet specification
func NewCIMChiplet(id int, rows, cols, numArrays int) *ChipletSpec {
	return &ChipletSpec{
		ID:             id,
		Type:           ChipletTypeCIM,
		Name:           fmt.Sprintf("CIM-%d", id),
		ArrayRows:      rows,
		ArrayCols:      cols,
		NumArrays:      numArrays,
		ProcessNode:    7,
		AreaMm2:        float64(numArrays) * 0.1,
		PowerW:         float64(numArrays) * 0.5,
		PeakTOPS:       float64(numArrays*rows*cols) * 2 / 1e12,
		MemoryMB:       numArrays * rows * cols * 8 / 8 / 1024 / 1024,
		InterconnectBW: 100, // GB/s
	}
}

// InterconnectLink represents a chip-to-chip link
type InterconnectLink struct {
	ID           int
	Type         InterconnectType
	SourceChipID int
	DestChipID   int
	BandwidthGBs float64
	LatencyNs    float64
	EnergyPJPerBit float64
}

// NoCRouter represents a Network-on-Chip router
type NoCRouter struct {
	ID          int
	ChipletID   int
	Position    [2]int // (row, col) in mesh
	InputPorts  []chan *NoCPacket
	OutputPorts []chan *NoCPacket
	RoutingTable map[int]int // dest -> output port
	BufferSize  int
	mu          sync.Mutex
}

// NoCPacket represents a packet in the NoC
type NoCPacket struct {
	ID          int
	SourceID    int
	DestID      int
	Payload     []float32
	PayloadSize int
	Hops        int
	Timestamp   time.Time
	Priority    int
}

// NewNoCRouter creates a new NoC router
func NewNoCRouter(id, chipletID int, pos [2]int) *NoCRouter {
	return &NoCRouter{
		ID:           id,
		ChipletID:    chipletID,
		Position:     pos,
		InputPorts:   make([]chan *NoCPacket, 5),  // N, S, E, W, Local
		OutputPorts:  make([]chan *NoCPacket, 5),
		RoutingTable: make(map[int]int),
		BufferSize:   16,
	}
}

// XYRoute implements XY dimension-order routing
func (r *NoCRouter) XYRoute(packet *NoCPacket, destPos [2]int) int {
	// X-first, then Y routing
	if r.Position[0] < destPos[0] {
		return 2 // East
	} else if r.Position[0] > destPos[0] {
		return 3 // West
	} else if r.Position[1] < destPos[1] {
		return 1 // South
	} else if r.Position[1] > destPos[1] {
		return 0 // North
	}
	return 4 // Local (arrived)
}

// MeshTopology represents a 2D mesh interconnect topology
type MeshTopology struct {
	Rows       int
	Cols       int
	Routers    [][]*NoCRouter
	Chiplets   []*ChipletSpec
	Links      []*InterconnectLink
}

// NewMeshTopology creates a new mesh topology
func NewMeshTopology(rows, cols int) *MeshTopology {
	mt := &MeshTopology{
		Rows:     rows,
		Cols:     cols,
		Routers:  make([][]*NoCRouter, rows),
		Chiplets: make([]*ChipletSpec, 0),
		Links:    make([]*InterconnectLink, 0),
	}

	// Create router mesh
	routerID := 0
	for r := 0; r < rows; r++ {
		mt.Routers[r] = make([]*NoCRouter, cols)
		for c := 0; c < cols; c++ {
			mt.Routers[r][c] = NewNoCRouter(routerID, r*cols+c, [2]int{c, r})
			routerID++
		}
	}

	return mt
}

// AddChiplet adds a chiplet at a position
func (mt *MeshTopology) AddChiplet(chiplet *ChipletSpec, row, col int) {
	mt.Chiplets = append(mt.Chiplets, chiplet)
	if row < mt.Rows && col < mt.Cols {
		mt.Routers[row][col].ChipletID = chiplet.ID
	}
}

// ConnectNeighbors creates links between adjacent routers
func (mt *MeshTopology) ConnectNeighbors(bwGBs, latNs float64) {
	linkID := 0
	for r := 0; r < mt.Rows; r++ {
		for c := 0; c < mt.Cols; c++ {
			// Connect to East neighbor
			if c < mt.Cols-1 {
				mt.Links = append(mt.Links, &InterconnectLink{
					ID:             linkID,
					Type:           InterconnectTypeD2D,
					SourceChipID:   r*mt.Cols + c,
					DestChipID:     r*mt.Cols + c + 1,
					BandwidthGBs:   bwGBs,
					LatencyNs:      latNs,
					EnergyPJPerBit: 0.5,
				})
				linkID++
			}
			// Connect to South neighbor
			if r < mt.Rows-1 {
				mt.Links = append(mt.Links, &InterconnectLink{
					ID:             linkID,
					Type:           InterconnectTypeD2D,
					SourceChipID:   r*mt.Cols + c,
					DestChipID:     (r+1)*mt.Cols + c,
					BandwidthGBs:   bwGBs,
					LatencyNs:      latNs,
					EnergyPJPerBit: 0.5,
				})
				linkID++
			}
		}
	}
}

// MultiChipCIMSystem represents a multi-chip CIM system
type MultiChipCIMSystem struct {
	Name          string
	Topology      *MeshTopology
	Chiplets      []*ChipletSpec
	TotalTOPS     float64
	TotalPowerW   float64
	TotalAreaMm2  float64
	InterconnectBW float64
	Scheduler     *WorkloadScheduler
}

// WorkloadScheduler schedules DNN workloads across chiplets
type WorkloadScheduler struct {
	System       *MultiChipCIMSystem
	LayerQueue   []*LayerTask
	ChipletLoads []float64
	Mappings     map[string][]int // layer -> chiplet IDs
	mu           sync.Mutex
}

// LayerTask represents a DNN layer to execute
type LayerTask struct {
	ID          int
	Name        string
	Type        string // "conv", "fc", "attention"
	InputSize   int
	OutputSize  int
	WeightSize  int
	MACs        int64
	Dependencies []int // IDs of dependent layers
	AssignedChips []int
	Status      string // "pending", "running", "complete"
}

// NewMultiChipCIMSystem creates a new multi-chip system
func NewMultiChipCIMSystem(name string, meshRows, meshCols int) *MultiChipCIMSystem {
	topology := NewMeshTopology(meshRows, meshCols)

	system := &MultiChipCIMSystem{
		Name:     name,
		Topology: topology,
		Chiplets: make([]*ChipletSpec, 0),
	}

	system.Scheduler = &WorkloadScheduler{
		System:       system,
		LayerQueue:   make([]*LayerTask, 0),
		ChipletLoads: make([]float64, meshRows*meshCols),
		Mappings:     make(map[string][]int),
	}

	return system
}

// AddCIMChiplets adds CIM chiplets to the system
func (mcs *MultiChipCIMSystem) AddCIMChiplets(numChiplets, arrayRows, arrayCols, arraysPerChip int) {
	gridSize := mcs.Topology.Rows * mcs.Topology.Cols
	if numChiplets > gridSize {
		numChiplets = gridSize
	}

	for i := 0; i < numChiplets; i++ {
		chiplet := NewCIMChiplet(i, arrayRows, arrayCols, arraysPerChip)
		mcs.Chiplets = append(mcs.Chiplets, chiplet)

		row := i / mcs.Topology.Cols
		col := i % mcs.Topology.Cols
		mcs.Topology.AddChiplet(chiplet, row, col)

		mcs.TotalTOPS += chiplet.PeakTOPS
		mcs.TotalPowerW += chiplet.PowerW
		mcs.TotalAreaMm2 += chiplet.AreaMm2
	}

	// Connect chiplets
	mcs.Topology.ConnectNeighbors(100, 10) // 100 GB/s, 10ns latency
	mcs.InterconnectBW = float64(len(mcs.Topology.Links)) * 100
}

// ScheduleWorkload schedules a DNN workload
func (ws *WorkloadScheduler) ScheduleWorkload(layers []*LayerTask) map[string][]int {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.LayerQueue = layers
	ws.Mappings = make(map[string][]int)

	numChiplets := len(ws.System.Chiplets)
	if numChiplets == 0 {
		return ws.Mappings
	}

	// Reset loads
	ws.ChipletLoads = make([]float64, numChiplets)

	// Simple load-balancing scheduler
	for _, layer := range layers {
		// Find least-loaded chiplets
		requiredChips := ws.estimateChipsNeeded(layer)
		assignedChips := ws.selectLeastLoaded(requiredChips)

		layer.AssignedChips = assignedChips
		ws.Mappings[layer.Name] = assignedChips

		// Update loads
		loadPerChip := float64(layer.MACs) / float64(len(assignedChips))
		for _, chipID := range assignedChips {
			ws.ChipletLoads[chipID] += loadPerChip
		}
	}

	return ws.Mappings
}

func (ws *WorkloadScheduler) estimateChipsNeeded(layer *LayerTask) int {
	if len(ws.System.Chiplets) == 0 {
		return 1
	}

	// Estimate based on weight size and chiplet capacity
	chipCapacity := ws.System.Chiplets[0].ArrayRows * ws.System.Chiplets[0].ArrayCols *
		ws.System.Chiplets[0].NumArrays

	chipsNeeded := (layer.WeightSize + chipCapacity - 1) / chipCapacity
	if chipsNeeded < 1 {
		chipsNeeded = 1
	}
	if chipsNeeded > len(ws.System.Chiplets) {
		chipsNeeded = len(ws.System.Chiplets)
	}

	return chipsNeeded
}

func (ws *WorkloadScheduler) selectLeastLoaded(n int) []int {
	type chipLoad struct {
		id   int
		load float64
	}

	chips := make([]chipLoad, len(ws.ChipletLoads))
	for i, load := range ws.ChipletLoads {
		chips[i] = chipLoad{i, load}
	}

	sort.Slice(chips, func(i, j int) bool {
		return chips[i].load < chips[j].load
	})

	selected := make([]int, n)
	for i := 0; i < n; i++ {
		selected[i] = chips[i].id
	}

	return selected
}

// SimulateExecution simulates workload execution
func (mcs *MultiChipCIMSystem) SimulateExecution(layers []*LayerTask) ExecutionResult {
	mappings := mcs.Scheduler.ScheduleWorkload(layers)

	totalMACs := int64(0)
	totalDataMovement := int64(0)

	for _, layer := range layers {
		totalMACs += layer.MACs

		// Estimate inter-chip data movement
		if len(layer.AssignedChips) > 1 {
			// Data needs to be distributed and results gathered
			totalDataMovement += int64(layer.InputSize + layer.OutputSize)
		}
	}

	// Calculate execution time
	execTimeMs := float64(totalMACs) / (mcs.TotalTOPS * 1e9) // seconds to ms

	// Add communication overhead
	commOverheadMs := float64(totalDataMovement) / (mcs.InterconnectBW * 1e9) * 1000

	// Calculate energy
	computeEnergy := float64(totalMACs) * 0.1e-12 * 1000 // pJ to mJ
	commEnergy := float64(totalDataMovement) * 8 * 0.5e-12 * 1000 // bits * energy/bit

	return ExecutionResult{
		TotalMACs:          totalMACs,
		ExecutionTimeMs:    execTimeMs + commOverheadMs,
		ComputeTimeMs:      execTimeMs,
		CommunicationTimeMs: commOverheadMs,
		ComputeEnergyMJ:    computeEnergy,
		CommunicationEnergyMJ: commEnergy,
		TotalEnergyMJ:      computeEnergy + commEnergy,
		ChipletMappings:    mappings,
		Efficiency:         float64(totalMACs) / (execTimeMs + commOverheadMs) / 1e9, // TOPS
		Utilization:        mcs.calculateUtilization(layers),
	}
}

func (mcs *MultiChipCIMSystem) calculateUtilization(layers []*LayerTask) float64 {
	if len(mcs.Chiplets) == 0 {
		return 0
	}

	chipUsed := make(map[int]bool)
	for _, layer := range layers {
		for _, chipID := range layer.AssignedChips {
			chipUsed[chipID] = true
		}
	}

	return float64(len(chipUsed)) / float64(len(mcs.Chiplets)) * 100
}

// ExecutionResult holds execution simulation results
type ExecutionResult struct {
	TotalMACs           int64
	ExecutionTimeMs     float64
	ComputeTimeMs       float64
	CommunicationTimeMs float64
	ComputeEnergyMJ     float64
	CommunicationEnergyMJ float64
	TotalEnergyMJ       float64
	ChipletMappings     map[string][]int
	Efficiency          float64 // TOPS
	Utilization         float64 // %
}

// ============================================================================
// Photonic Interconnect for Multi-Chip CIM
// ============================================================================

// PhotonicLinkSpec specifies a photonic interconnect
type PhotonicLinkSpec struct {
	ID              int
	Wavelengths     int     // Number of WDM channels
	DataRateGbps    float64 // Per wavelength
	LaserPowerMW    float64
	ModulatorLoss   float64 // dB
	PropagationLoss float64 // dB/cm
	DetectorResp    float64 // A/W
	TotalBWGbps     float64
	EnergyPJPerBit  float64
}

// NewPhotonicLink creates a photonic link specification
func NewPhotonicLink(id, wavelengths int, dataRate float64) *PhotonicLinkSpec {
	return &PhotonicLinkSpec{
		ID:              id,
		Wavelengths:     wavelengths,
		DataRateGbps:    dataRate,
		LaserPowerMW:    1.0, // 1mW laser
		ModulatorLoss:   3.0, // dB
		PropagationLoss: 0.5, // dB/cm
		DetectorResp:    0.8, // A/W
		TotalBWGbps:     float64(wavelengths) * dataRate,
		EnergyPJPerBit:  0.1, // 100 fJ/bit target
	}
}

// PhotonicMeshNetwork represents a photonic mesh interconnect
type PhotonicMeshNetwork struct {
	Rows            int
	Cols            int
	Links           [][]*PhotonicLinkSpec
	TotalBandwidth  float64 // Tbps
	AverageLatency  float64 // ns
}

// NewPhotonicMeshNetwork creates a photonic mesh network
func NewPhotonicMeshNetwork(rows, cols, wavelengths int, dataRate float64) *PhotonicMeshNetwork {
	pmn := &PhotonicMeshNetwork{
		Rows: rows,
		Cols: cols,
		Links: make([][]*PhotonicLinkSpec, rows*cols),
	}

	linkID := 0
	for i := 0; i < rows*cols; i++ {
		// Each node has up to 4 photonic links (N, S, E, W)
		pmn.Links[i] = make([]*PhotonicLinkSpec, 0)

		if i%cols < cols-1 { // East link
			link := NewPhotonicLink(linkID, wavelengths, dataRate)
			pmn.Links[i] = append(pmn.Links[i], link)
			pmn.TotalBandwidth += link.TotalBWGbps / 1000 // Tbps
			linkID++
		}
		if i/cols < rows-1 { // South link
			link := NewPhotonicLink(linkID, wavelengths, dataRate)
			pmn.Links[i] = append(pmn.Links[i], link)
			pmn.TotalBandwidth += link.TotalBWGbps / 1000
			linkID++
		}
	}

	pmn.AverageLatency = 1.0 // ~1ns for photonic

	return pmn
}

// CalculateEnergyReduction computes energy savings vs electrical
func (pmn *PhotonicMeshNetwork) CalculateEnergyReduction() float64 {
	// Photonic typically 0.1 pJ/bit vs electrical 0.5-1 pJ/bit
	photonicEnergy := 0.1  // pJ/bit
	electricalEnergy := 0.5 // pJ/bit
	return (electricalEnergy - photonicEnergy) / electricalEnergy * 100
}

// ============================================================================
// CIM System Benchmark Suite
// ============================================================================

// SystemBenchmark benchmarks multi-chip CIM systems
type SystemBenchmark struct {
	System        *MultiChipCIMSystem
	Workloads     []BenchmarkWorkload
	Results       []SystemBenchmarkResult
}

// BenchmarkWorkload defines a benchmark workload
type BenchmarkWorkload struct {
	Name        string
	Model       string // "ResNet50", "BERT", "GPT-2"
	BatchSize   int
	TotalMACs   int64
	TotalParams int
	Layers      []*LayerTask
}

// SystemBenchmarkResult holds benchmark results
type SystemBenchmarkResult struct {
	WorkloadName     string
	ExecutionTimeMs  float64
	ThroughputTOPS   float64
	EfficiencyTOPSW  float64
	EnergyMJ         float64
	Utilization      float64
	InterconnectUtil float64
	BottleneckType   string // "compute", "memory", "interconnect"
}

// NewSystemBenchmark creates a new system benchmark
func NewSystemBenchmark(system *MultiChipCIMSystem) *SystemBenchmark {
	return &SystemBenchmark{
		System:    system,
		Workloads: make([]BenchmarkWorkload, 0),
		Results:   make([]SystemBenchmarkResult, 0),
	}
}

// AddStandardWorkloads adds standard DNN workloads
func (sb *SystemBenchmark) AddStandardWorkloads() {
	// ResNet-50
	resnet50 := BenchmarkWorkload{
		Name:        "ResNet-50",
		Model:       "ResNet50",
		BatchSize:   1,
		TotalMACs:   4100000000, // 4.1 GMACs
		TotalParams: 25600000,   // 25.6M
		Layers:      generateResNet50Layers(),
	}

	// BERT-Base
	bertBase := BenchmarkWorkload{
		Name:        "BERT-Base",
		Model:       "BERT",
		BatchSize:   1,
		TotalMACs:   22000000000, // 22 GMACs for sequence length 128
		TotalParams: 110000000,   // 110M
		Layers:      generateBERTLayers(),
	}

	// GPT-2 Small
	gpt2Small := BenchmarkWorkload{
		Name:        "GPT-2-Small",
		Model:       "GPT-2",
		BatchSize:   1,
		TotalMACs:   35000000000, // 35 GMACs
		TotalParams: 117000000,   // 117M
		Layers:      generateGPT2Layers(),
	}

	sb.Workloads = append(sb.Workloads, resnet50, bertBase, gpt2Small)
}

func generateResNet50Layers() []*LayerTask {
	layers := make([]*LayerTask, 0)
	layerID := 0

	// Simplified ResNet-50 layers
	layerConfigs := []struct {
		name       string
		ltype      string
		inSize     int
		outSize    int
		weightSize int
		macs       int64
	}{
		{"conv1", "conv", 150528, 802816, 9408, 118013952},
		{"layer1", "conv", 802816, 802816, 147456, 231211008},
		{"layer2", "conv", 802816, 401408, 524288, 412876800},
		{"layer3", "conv", 401408, 100352, 1179648, 412876800},
		{"layer4", "conv", 100352, 25088, 2359296, 231211008},
		{"fc", "fc", 2048, 1000, 2048000, 2048000},
	}

	for _, cfg := range layerConfigs {
		layers = append(layers, &LayerTask{
			ID:         layerID,
			Name:       cfg.name,
			Type:       cfg.ltype,
			InputSize:  cfg.inSize,
			OutputSize: cfg.outSize,
			WeightSize: cfg.weightSize,
			MACs:       cfg.macs,
			Status:     "pending",
		})
		layerID++
	}

	return layers
}

func generateBERTLayers() []*LayerTask {
	layers := make([]*LayerTask, 0)
	layerID := 0

	// 12 transformer layers
	for i := 0; i < 12; i++ {
		// Self-attention
		layers = append(layers, &LayerTask{
			ID:         layerID,
			Name:       fmt.Sprintf("layer%d_attention", i),
			Type:       "attention",
			InputSize:  98304,  // 128 * 768
			OutputSize: 98304,
			WeightSize: 2359296, // 768 * 768 * 4 (Q, K, V, O)
			MACs:       603979776,
			Status:     "pending",
		})
		layerID++

		// FFN
		layers = append(layers, &LayerTask{
			ID:         layerID,
			Name:       fmt.Sprintf("layer%d_ffn", i),
			Type:       "fc",
			InputSize:  98304,
			OutputSize: 98304,
			WeightSize: 4718592, // 768 * 3072 * 2
			MACs:       1207959552,
			Status:     "pending",
		})
		layerID++
	}

	return layers
}

func generateGPT2Layers() []*LayerTask {
	layers := make([]*LayerTask, 0)
	layerID := 0

	// 12 transformer decoder layers
	for i := 0; i < 12; i++ {
		// Masked self-attention
		layers = append(layers, &LayerTask{
			ID:         layerID,
			Name:       fmt.Sprintf("layer%d_attention", i),
			Type:       "attention",
			InputSize:  98304,
			OutputSize: 98304,
			WeightSize: 2359296,
			MACs:       603979776,
			Status:     "pending",
		})
		layerID++

		// FFN
		layers = append(layers, &LayerTask{
			ID:         layerID,
			Name:       fmt.Sprintf("layer%d_ffn", i),
			Type:       "fc",
			InputSize:  98304,
			OutputSize: 98304,
			WeightSize: 4718592,
			MACs:       1207959552,
			Status:     "pending",
		})
		layerID++
	}

	return layers
}

// RunBenchmarks runs all benchmark workloads
func (sb *SystemBenchmark) RunBenchmarks() []SystemBenchmarkResult {
	sb.Results = make([]SystemBenchmarkResult, 0)

	for _, workload := range sb.Workloads {
		result := sb.System.SimulateExecution(workload.Layers)

		throughput := float64(workload.TotalMACs) / result.ExecutionTimeMs / 1e9 // TOPS
		efficiency := throughput / sb.System.TotalPowerW                         // TOPS/W

		// Determine bottleneck
		bottleneck := "compute"
		if result.CommunicationTimeMs > result.ComputeTimeMs {
			bottleneck = "interconnect"
		}

		sbr := SystemBenchmarkResult{
			WorkloadName:     workload.Name,
			ExecutionTimeMs:  result.ExecutionTimeMs,
			ThroughputTOPS:   throughput,
			EfficiencyTOPSW:  efficiency,
			EnergyMJ:         result.TotalEnergyMJ,
			Utilization:      result.Utilization,
			InterconnectUtil: result.CommunicationTimeMs / result.ExecutionTimeMs * 100,
			BottleneckType:   bottleneck,
		}

		sb.Results = append(sb.Results, sbr)
	}

	return sb.Results
}

// PrintResults prints benchmark results
func (sb *SystemBenchmark) PrintResults() string {
	output := "Multi-Chip CIM System Benchmark Results\n"
	output += "=========================================\n\n"
	output += fmt.Sprintf("System: %s\n", sb.System.Name)
	output += fmt.Sprintf("Chiplets: %d\n", len(sb.System.Chiplets))
	output += fmt.Sprintf("Total TOPS: %.2f\n", sb.System.TotalTOPS)
	output += fmt.Sprintf("Total Power: %.2f W\n\n", sb.System.TotalPowerW)

	output += "Workload Results:\n"
	output += "-----------------\n"

	for _, r := range sb.Results {
		output += fmt.Sprintf("\n%s:\n", r.WorkloadName)
		output += fmt.Sprintf("  Execution Time: %.3f ms\n", r.ExecutionTimeMs)
		output += fmt.Sprintf("  Throughput: %.2f TOPS\n", r.ThroughputTOPS)
		output += fmt.Sprintf("  Efficiency: %.2f TOPS/W\n", r.EfficiencyTOPSW)
		output += fmt.Sprintf("  Energy: %.4f mJ\n", r.EnergyMJ)
		output += fmt.Sprintf("  Utilization: %.1f%%\n", r.Utilization)
		output += fmt.Sprintf("  Bottleneck: %s\n", r.BottleneckType)
	}

	return output
}
