// Package layers provides neuromorphic vision sensors and multi-chip CIM interconnect simulations.
// Based on ferroelectric event cameras, DVS technology, UCIe 3.0, and CIM tiling research.
package layers

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// =============================================================================
// NEUROMORPHIC VISION SENSOR (EVENT CAMERA / DVS)
// =============================================================================

// EventCameraConfig holds configuration for dynamic vision sensor.
type EventCameraConfig struct {
	ResolutionX      int       // Horizontal resolution
	ResolutionY      int       // Vertical resolution
	ThresholdPos     float64   // Positive change threshold (log intensity)
	ThresholdNeg     float64   // Negative change threshold
	TemporalResUS    float64   // Temporal resolution (microseconds)
	DynamicRangeDB   float64   // Dynamic range in dB
	PowerMW          float64   // Power consumption (mW)
	LatencyUS        float64   // Pixel latency (μs)
}

// DefaultEventCameraConfig returns typical DVS configuration.
func DefaultEventCameraConfig() *EventCameraConfig {
	return &EventCameraConfig{
		ResolutionX:    320,
		ResolutionY:    320,
		ThresholdPos:   0.15,    // 15% log intensity change
		ThresholdNeg:   0.15,
		TemporalResUS:  1.0,     // 1 μs resolution
		DynamicRangeDB: 120.0,   // 120 dB
		PowerMW:        10.0,    // ~10 mW
		LatencyUS:      100.0,   // ~100 μs typical
	}
}

// HD720EventCameraConfig returns high-resolution DVS configuration.
func HD720EventCameraConfig() *EventCameraConfig {
	return &EventCameraConfig{
		ResolutionX:    1280,
		ResolutionY:    720,
		ThresholdPos:   0.12,
		ThresholdNeg:   0.12,
		TemporalResUS:  1.0,
		DynamicRangeDB: 120.0,
		PowerMW:        50.0,
		LatencyUS:      200.0,
	}
}

// Event represents a single DVS event.
type Event struct {
	X         int       // Pixel X coordinate
	Y         int       // Pixel Y coordinate
	Timestamp float64   // Timestamp in microseconds
	Polarity  int       // +1 for ON (brighter), -1 for OFF (darker)
}

// EventPixel represents a single pixel in the event camera.
type EventPixel struct {
	Config        *EventCameraConfig
	X, Y          int
	LogIntensity  float64   // Current log intensity
	RefIntensity  float64   // Reference intensity
	LastEventTime float64
	EventCount    int64
}

// NewEventPixel creates a new event pixel.
func NewEventPixel(cfg *EventCameraConfig, x, y int) *EventPixel {
	return &EventPixel{
		Config:       cfg,
		X:            x,
		Y:            y,
		LogIntensity: 0.5, // Mid-range
		RefIntensity: 0.5,
	}
}

// Update processes new intensity and generates events if threshold crossed.
func (p *EventPixel) Update(intensity float64, timestamp float64) *Event {
	// Convert to log intensity
	newLog := math.Log(intensity + 1e-6)
	p.LogIntensity = newLog

	// Check threshold crossing
	diff := newLog - p.RefIntensity

	if diff > p.Config.ThresholdPos {
		// ON event (brighter)
		p.RefIntensity = newLog
		p.LastEventTime = timestamp
		p.EventCount++
		return &Event{
			X:         p.X,
			Y:         p.Y,
			Timestamp: timestamp,
			Polarity:  1,
		}
	} else if diff < -p.Config.ThresholdNeg {
		// OFF event (darker)
		p.RefIntensity = newLog
		p.LastEventTime = timestamp
		p.EventCount++
		return &Event{
			X:         p.X,
			Y:         p.Y,
			Timestamp: timestamp,
			Polarity:  -1,
		}
	}

	return nil
}

// EventCamera implements a full event-based vision sensor.
type EventCamera struct {
	Config       *EventCameraConfig
	Pixels       [][]*EventPixel
	EventBuffer  []*Event
	CurrentTime  float64
	TotalEvents  int64
}

// NewEventCamera creates a new event camera.
func NewEventCamera(cfg *EventCameraConfig) *EventCamera {
	pixels := make([][]*EventPixel, cfg.ResolutionY)
	for y := range pixels {
		pixels[y] = make([]*EventPixel, cfg.ResolutionX)
		for x := range pixels[y] {
			pixels[y][x] = NewEventPixel(cfg, x, y)
		}
	}

	return &EventCamera{
		Config:      cfg,
		Pixels:      pixels,
		EventBuffer: make([]*Event, 0, 10000),
	}
}

// ProcessFrame processes a full intensity frame and generates events.
func (c *EventCamera) ProcessFrame(frame [][]float64, timestamp float64) []*Event {
	events := make([]*Event, 0)
	c.CurrentTime = timestamp

	for y := 0; y < c.Config.ResolutionY && y < len(frame); y++ {
		for x := 0; x < c.Config.ResolutionX && x < len(frame[y]); x++ {
			event := c.Pixels[y][x].Update(frame[y][x], timestamp)
			if event != nil {
				events = append(events, event)
				c.TotalEvents++
			}
		}
	}

	c.EventBuffer = append(c.EventBuffer, events...)
	return events
}

// GetSparsity returns data reduction ratio vs frame-based.
func (c *EventCamera) GetSparsity(frameCount int) float64 {
	totalPixels := int64(c.Config.ResolutionX * c.Config.ResolutionY * frameCount)
	if totalPixels == 0 {
		return 0
	}
	return 1.0 - float64(c.TotalEvents)/float64(totalPixels)
}

// =============================================================================
// FERROELECTRIC NEUROMORPHIC PHOTODETECTOR
// =============================================================================

// FEPhotodetectorConfig holds configuration for ferroelectric photodetector.
type FEPhotodetectorConfig struct {
	PixelArrayX     int       // Array X dimension
	PixelArrayY     int       // Array Y dimension
	HZOThickness    float64   // HZO layer thickness (nm)
	EventSpikeUS    float64   // Event spike duration (μs)
	STMDecayMS      float64   // Short-term memory decay (ms)
	StaticMode      bool      // Enable static sensing
	EventMode       bool      // Enable event detection
	MemoryMode      bool      // Enable STM
}

// DefaultFEPhotodetectorConfig returns default HZO photodetector config.
// Based on 2025 Nature Energy paper.
func DefaultFEPhotodetectorConfig() *FEPhotodetectorConfig {
	return &FEPhotodetectorConfig{
		PixelArrayX:  40,
		PixelArrayY:  40,
		HZOThickness: 6.0,      // ~6 nm HZO
		EventSpikeUS: 73.0,     // ~73 μs self-powered spikes
		STMDecayMS:   10.0,     // 10 ms STM decay
		StaticMode:   true,
		EventMode:    true,
		MemoryMode:   true,
	}
}

// FEPixel represents a ferroelectric photodetector pixel.
type FEPixel struct {
	Config          *FEPhotodetectorConfig
	X, Y            int
	Polarization    float64   // Ferroelectric polarization state
	Photocurrent    float64   // Static photocurrent
	MemoryTrace     float64   // Short-term memory trace
	LastEventTime   float64
	OperatingMode   string    // "static", "event", "memory"
}

// NewFEPixel creates a new FE photodetector pixel.
func NewFEPixel(cfg *FEPhotodetectorConfig, x, y int) *FEPixel {
	return &FEPixel{
		Config:        cfg,
		X:             x,
		Y:             y,
		OperatingMode: "event",
	}
}

// SetBias sets the bias voltage to control operating mode.
func (p *FEPixel) SetBias(bias float64) {
	// Bias-dependent mode transitions
	if bias < -2.0 {
		p.OperatingMode = "event"      // Self-powered event spikes
	} else if bias > 2.0 {
		p.OperatingMode = "static"     // Steady-state photocurrent
	} else {
		p.OperatingMode = "memory"     // Programmable STM
	}
}

// Process processes light input and returns response.
func (p *FEPixel) Process(intensity float64, timestamp float64) *FEResponse {
	response := &FEResponse{
		X:         p.X,
		Y:         p.Y,
		Timestamp: timestamp,
		Mode:      p.OperatingMode,
	}

	switch p.OperatingMode {
	case "static":
		// Steady-state photocurrent
		p.Photocurrent = intensity * p.Polarization
		response.Value = p.Photocurrent
		response.IsSpike = false

	case "event":
		// Self-powered event spike (~73 μs)
		if intensity > 0.5 { // Threshold
			response.IsSpike = true
			response.SpikeDuration = p.Config.EventSpikeUS
			p.LastEventTime = timestamp
		}
		response.Value = 0

	case "memory":
		// Short-term memory with decay
		// Update memory trace with input
		if intensity > 0.3 {
			p.MemoryTrace = math.Min(1.0, p.MemoryTrace+0.2)
		}
		// Exponential decay
		dt := timestamp - p.LastEventTime
		if dt > 0 {
			decay := math.Exp(-dt / (p.Config.STMDecayMS * 1000))
			p.MemoryTrace *= decay
		}
		p.LastEventTime = timestamp
		response.Value = p.MemoryTrace
	}

	return response
}

// FEResponse holds photodetector response.
type FEResponse struct {
	X             int
	Y             int
	Timestamp     float64
	Mode          string
	Value         float64
	IsSpike       bool
	SpikeDuration float64
}

// FEPhotodetector implements ferroelectric neuromorphic photodetector array.
type FEPhotodetector struct {
	Config        *FEPhotodetectorConfig
	Pixels        [][]*FEPixel
	Responses     []*FEResponse
	CurrentTime   float64
}

// NewFEPhotodetector creates a new FE photodetector array.
func NewFEPhotodetector(cfg *FEPhotodetectorConfig) *FEPhotodetector {
	pixels := make([][]*FEPixel, cfg.PixelArrayY)
	for y := range pixels {
		pixels[y] = make([]*FEPixel, cfg.PixelArrayX)
		for x := range pixels[y] {
			pixels[y][x] = NewFEPixel(cfg, x, y)
		}
	}

	return &FEPhotodetector{
		Config: cfg,
		Pixels: pixels,
	}
}

// SetGlobalBias sets bias for all pixels.
func (d *FEPhotodetector) SetGlobalBias(bias float64) {
	for y := range d.Pixels {
		for x := range d.Pixels[y] {
			d.Pixels[y][x].SetBias(bias)
		}
	}
}

// ProcessFrame processes light input array.
func (d *FEPhotodetector) ProcessFrame(frame [][]float64, timestamp float64) []*FEResponse {
	responses := make([]*FEResponse, 0)
	d.CurrentTime = timestamp

	for y := 0; y < d.Config.PixelArrayY && y < len(frame); y++ {
		for x := 0; x < d.Config.PixelArrayX && x < len(frame[y]); x++ {
			resp := d.Pixels[y][x].Process(frame[y][x], timestamp)
			responses = append(responses, resp)
		}
	}

	d.Responses = append(d.Responses, responses...)
	return responses
}

// ClassifySpatiotemporal performs on-chip classification.
// Based on 93% accuracy demonstrated in 2025 paper.
func (d *FEPhotodetector) ClassifySpatiotemporal(patterns [][][]float64) (int, float64) {
	// Simple pattern matching using memory traces
	bestMatch := 0
	bestScore := 0.0

	for i, pattern := range patterns {
		score := 0.0
		for y := range pattern {
			for x := range pattern[y] {
				if y < d.Config.PixelArrayY && x < d.Config.PixelArrayX {
					diff := math.Abs(d.Pixels[y][x].MemoryTrace - pattern[y][x])
					score += 1.0 - diff
				}
			}
		}
		score /= float64(d.Config.PixelArrayX * d.Config.PixelArrayY)

		if score > bestScore {
			bestScore = score
			bestMatch = i
		}
	}

	return bestMatch, bestScore
}

// =============================================================================
// MULTI-CHIP CIM INTERCONNECT
// =============================================================================

// UCIeConfig holds UCIe 3.0 interconnect configuration.
type UCIeConfig struct {
	Version         string    // "1.0", "2.0", "3.0"
	DataRateGTs     float64   // GT/s (32, 48, 64)
	BandwidthDensity float64  // Tbps/mm
	SidebandReachMM float64   // Sideband reach in mm
	Lanes           int       // Number of lanes
	LinkLatencyNS   float64   // Link latency (ns)
	PowerPerLane    float64   // mW per lane
}

// UCIe30Config returns UCIe 3.0 configuration.
func UCIe30Config() *UCIeConfig {
	return &UCIeConfig{
		Version:          "3.0",
		DataRateGTs:      64.0,     // 64 GT/s max
		BandwidthDensity: 20.0,     // >20 Tbps/mm
		SidebandReachMM:  100.0,    // 100 mm extended sideband
		Lanes:            16,       // Standard 16 lanes
		LinkLatencyNS:    2.0,      // ~2 ns link latency
		PowerPerLane:     50.0,     // ~50 mW/lane at 64 GT/s
	}
}

// UCIe20Config returns UCIe 2.0 configuration.
func UCIe20Config() *UCIeConfig {
	return &UCIeConfig{
		Version:          "2.0",
		DataRateGTs:      32.0,
		BandwidthDensity: 10.0,
		SidebandReachMM:  50.0,
		Lanes:            16,
		LinkLatencyNS:    3.0,
		PowerPerLane:     30.0,
	}
}

// ChipletLink represents a die-to-die link.
type ChipletLink struct {
	Config        *UCIeConfig
	SourceID      int
	DestID        int
	Bandwidth     float64   // Current bandwidth (Gbps)
	Utilization   float64   // Link utilization (0-1)
	PacketCount   int64
	TotalBytes    int64
}

// NewChipletLink creates a new chiplet link.
func NewChipletLink(cfg *UCIeConfig, src, dst int) *ChipletLink {
	bandwidth := cfg.DataRateGTs * float64(cfg.Lanes) // Gbps
	return &ChipletLink{
		Config:    cfg,
		SourceID:  src,
		DestID:    dst,
		Bandwidth: bandwidth,
	}
}

// Transfer transfers data over the link.
func (l *ChipletLink) Transfer(bytes int64) float64 {
	// Calculate transfer time in nanoseconds
	bits := bytes * 8
	transferTimeNS := float64(bits) / l.Bandwidth // Gbps to ns
	totalTimeNS := transferTimeNS + l.Config.LinkLatencyNS

	l.PacketCount++
	l.TotalBytes += bytes
	l.Utilization = 0.9*l.Utilization + 0.1*math.Min(1.0, float64(bits)/l.Bandwidth)

	return totalTimeNS
}

// GetPower returns current link power consumption.
func (l *ChipletLink) GetPower() float64 {
	return l.Config.PowerPerLane * float64(l.Config.Lanes) * l.Utilization
}

// =============================================================================
// NETWORK-ON-PACKAGE FOR CIM CHIPLETS
// =============================================================================

// NoPConfig holds Network-on-Package configuration.
type NoPConfig struct {
	NumChiplets     int       // Number of CIM chiplets
	Topology        string    // "mesh", "ring", "crossbar"
	UCIeVersion     string    // UCIe version
	ArraysPerChip   int       // CIM arrays per chiplet
	ArraySize       int       // Size of each CIM array
}

// DefaultNoPConfig returns default 3x3 mesh configuration.
func DefaultNoPConfig() *NoPConfig {
	return &NoPConfig{
		NumChiplets:   9,        // 3x3 mesh
		Topology:      "mesh",
		UCIeVersion:   "3.0",
		ArraysPerChip: 64,       // 64 CIM arrays per chiplet
		ArraySize:     64,       // 64x64 arrays
	}
}

// CIMChiplet represents a single CIM chiplet in the network.
type CIMChiplet struct {
	Config         *NoPConfig
	ID             int
	Position       [2]int      // Grid position
	Arrays         [][]*CIMArrayUnit
	LocalMemoryKB  int64
	ComputeTOPS    float64
	Links          []*ChipletLink
	TotalOps       int64
	EnergyPJ       float64
}

// CIMArrayUnit represents a CIM array within a chiplet.
type CIMArrayUnit struct {
	Size        int
	Weights     [][]float64
	Busy        bool
	UtilTime    float64
}

// NewCIMChiplet creates a new CIM chiplet.
func NewCIMChiplet(cfg *NoPConfig, id int, pos [2]int) *CIMChiplet {
	arrays := make([][]*CIMArrayUnit, 8)
	for i := range arrays {
		arrays[i] = make([]*CIMArrayUnit, 8)
		for j := range arrays[i] {
			weights := make([][]float64, cfg.ArraySize)
			for k := range weights {
				weights[k] = make([]float64, cfg.ArraySize)
			}
			arrays[i][j] = &CIMArrayUnit{
				Size:    cfg.ArraySize,
				Weights: weights,
			}
		}
	}

	return &CIMChiplet{
		Config:        cfg,
		ID:            id,
		Position:      pos,
		Arrays:        arrays,
		LocalMemoryKB: 256,                            // 256 KB local SRAM
		ComputeTOPS:   float64(cfg.ArraysPerChip) * 0.25, // ~0.25 TOPS per array
		Links:         make([]*ChipletLink, 0),
	}
}

// AddLink adds an interconnect link.
func (c *CIMChiplet) AddLink(link *ChipletLink) {
	c.Links = append(c.Links, link)
}

// ExecuteMVM executes matrix-vector multiply on available arrays.
func (c *CIMChiplet) ExecuteMVM(input []float64, weightRows, weightCols int) ([]float64, float64) {
	output := make([]float64, weightRows)

	// Find available arrays
	arraysNeeded := (weightRows + c.Config.ArraySize - 1) / c.Config.ArraySize
	if arraysNeeded > len(c.Arrays)*len(c.Arrays[0]) {
		arraysNeeded = len(c.Arrays) * len(c.Arrays[0])
	}

	// Execute on arrays
	energyPJ := 0.0
	for i := 0; i < weightRows; i++ {
		sum := 0.0
		for j := 0; j < weightCols && j < len(input); j++ {
			// Simplified - would use actual array weights
			sum += rand.Float64() * input[j]
		}
		output[i] = sum
		energyPJ += 0.5 // ~0.5 pJ per MAC
	}

	c.TotalOps += int64(weightRows * weightCols)
	c.EnergyPJ += energyPJ

	return output, energyPJ
}

// NetworkOnPackage implements a multi-chiplet CIM system.
type NetworkOnPackage struct {
	Config         *NoPConfig
	Chiplets       []*CIMChiplet
	Links          [][]*ChipletLink  // Adjacency matrix of links
	UCIeConfig     *UCIeConfig
	TotalBandwidth float64           // Aggregate bandwidth (Tbps)
	TotalOps       int64
	TotalEnergy    float64
}

// NewNetworkOnPackage creates a new NoP system.
func NewNetworkOnPackage(cfg *NoPConfig) *NetworkOnPackage {
	// Get UCIe config
	var ucieConfig *UCIeConfig
	if cfg.UCIeVersion == "3.0" {
		ucieConfig = UCIe30Config()
	} else {
		ucieConfig = UCIe20Config()
	}

	// Create chiplets
	gridSize := int(math.Sqrt(float64(cfg.NumChiplets)))
	chiplets := make([]*CIMChiplet, cfg.NumChiplets)
	for i := 0; i < cfg.NumChiplets; i++ {
		pos := [2]int{i / gridSize, i % gridSize}
		chiplets[i] = NewCIMChiplet(cfg, i, pos)
	}

	// Create link adjacency matrix
	links := make([][]*ChipletLink, cfg.NumChiplets)
	for i := range links {
		links[i] = make([]*ChipletLink, cfg.NumChiplets)
	}

	nop := &NetworkOnPackage{
		Config:     cfg,
		Chiplets:   chiplets,
		Links:      links,
		UCIeConfig: ucieConfig,
	}

	// Setup topology
	switch cfg.Topology {
	case "mesh":
		nop.setupMeshTopology(gridSize)
	case "ring":
		nop.setupRingTopology()
	case "crossbar":
		nop.setupCrossbarTopology()
	}

	// Calculate total bandwidth
	linkCount := 0
	for i := range links {
		for j := range links[i] {
			if links[i][j] != nil {
				linkCount++
			}
		}
	}
	nop.TotalBandwidth = float64(linkCount) * ucieConfig.DataRateGTs * float64(ucieConfig.Lanes) / 1000.0

	return nop
}

// setupMeshTopology creates mesh connections.
func (nop *NetworkOnPackage) setupMeshTopology(gridSize int) {
	for i, chip := range nop.Chiplets {
		row, col := chip.Position[0], chip.Position[1]

		// Connect to neighbors
		neighbors := [][2]int{
			{row - 1, col}, // North
			{row + 1, col}, // South
			{row, col - 1}, // West
			{row, col + 1}, // East
		}

		for _, n := range neighbors {
			if n[0] >= 0 && n[0] < gridSize && n[1] >= 0 && n[1] < gridSize {
				neighborID := n[0]*gridSize + n[1]
				if neighborID < len(nop.Chiplets) && nop.Links[i][neighborID] == nil {
					link := NewChipletLink(nop.UCIeConfig, i, neighborID)
					nop.Links[i][neighborID] = link
					chip.AddLink(link)
				}
			}
		}
	}
}

// setupRingTopology creates ring connections.
func (nop *NetworkOnPackage) setupRingTopology() {
	n := len(nop.Chiplets)
	for i := range nop.Chiplets {
		next := (i + 1) % n
		link := NewChipletLink(nop.UCIeConfig, i, next)
		nop.Links[i][next] = link
		nop.Chiplets[i].AddLink(link)
	}
}

// setupCrossbarTopology creates full crossbar connections.
func (nop *NetworkOnPackage) setupCrossbarTopology() {
	for i := range nop.Chiplets {
		for j := range nop.Chiplets {
			if i != j && nop.Links[i][j] == nil {
				link := NewChipletLink(nop.UCIeConfig, i, j)
				nop.Links[i][j] = link
				nop.Chiplets[i].AddLink(link)
			}
		}
	}
}

// RouteData routes data between chiplets.
func (nop *NetworkOnPackage) RouteData(srcID, dstID int, bytes int64) (float64, int) {
	if srcID == dstID {
		return 0, 0
	}

	// Simple shortest path routing
	path := nop.findShortestPath(srcID, dstID)
	if len(path) < 2 {
		return -1, 0 // No path
	}

	totalLatency := 0.0
	for i := 0; i < len(path)-1; i++ {
		link := nop.Links[path[i]][path[i+1]]
		if link != nil {
			totalLatency += link.Transfer(bytes)
		}
	}

	return totalLatency, len(path) - 1
}

// findShortestPath uses BFS to find shortest path.
func (nop *NetworkOnPackage) findShortestPath(src, dst int) []int {
	n := len(nop.Chiplets)
	visited := make([]bool, n)
	parent := make([]int, n)
	for i := range parent {
		parent[i] = -1
	}

	queue := []int{src}
	visited[src] = true

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr == dst {
			// Reconstruct path
			path := []int{}
			for c := dst; c != -1; c = parent[c] {
				path = append([]int{c}, path...)
			}
			return path
		}

		for next := 0; next < n; next++ {
			if nop.Links[curr][next] != nil && !visited[next] {
				visited[next] = true
				parent[next] = curr
				queue = append(queue, next)
			}
		}
	}

	return []int{}
}

// ExecuteDistributedMVM executes MVM across multiple chiplets with tiling.
func (nop *NetworkOnPackage) ExecuteDistributedMVM(input []float64, rows, cols int) (*DistributedMVMResult, error) {
	result := &DistributedMVMResult{
		Output:      make([]float64, rows),
		ChipletOps:  make([]int64, len(nop.Chiplets)),
		TotalHops:   0,
	}

	// Tile the computation across chiplets
	rowsPerChip := (rows + len(nop.Chiplets) - 1) / len(nop.Chiplets)

	for i, chip := range nop.Chiplets {
		startRow := i * rowsPerChip
		endRow := startRow + rowsPerChip
		if endRow > rows {
			endRow = rows
		}
		if startRow >= rows {
			break
		}

		tileRows := endRow - startRow

		// Execute local MVM
		output, energy := chip.ExecuteMVM(input, tileRows, cols)

		// Copy to result
		for j := 0; j < tileRows; j++ {
			result.Output[startRow+j] = output[j]
		}

		result.ChipletOps[i] = int64(tileRows * cols)
		result.TotalEnergy += energy
	}

	// Gather results (simulate data movement)
	for i := 1; i < len(nop.Chiplets); i++ {
		latency, hops := nop.RouteData(i, 0, int64(rowsPerChip*8))
		result.TotalLatency += latency
		result.TotalHops += hops
	}

	nop.TotalOps += int64(rows * cols)
	nop.TotalEnergy += result.TotalEnergy

	return result, nil
}

// DistributedMVMResult holds results from distributed MVM.
type DistributedMVMResult struct {
	Output       []float64
	ChipletOps   []int64
	TotalEnergy  float64
	TotalLatency float64
	TotalHops    int
}

// =============================================================================
// CIM DATAFLOW OPTIMIZER
// =============================================================================

// DataflowConfig holds CIM dataflow configuration.
type DataflowConfig struct {
	Strategy       string    // "weight_stationary", "output_stationary", "row_stationary"
	TileSize       int       // Tile dimension
	BufferSizeKB   int       // On-chip buffer size
	DoubleBuffer   bool      // Enable double buffering
	PipelineDepth  int       // Pipeline stages
}

// DefaultDataflowConfig returns weight-stationary config.
func DefaultDataflowConfig() *DataflowConfig {
	return &DataflowConfig{
		Strategy:      "weight_stationary",
		TileSize:      64,
		BufferSizeKB:  256,
		DoubleBuffer:  true,
		PipelineDepth: 4,
	}
}

// DataflowOptimizer optimizes CIM dataflow for multi-chip systems.
type DataflowOptimizer struct {
	Config         *DataflowConfig
	NoP            *NetworkOnPackage
}

// NewDataflowOptimizer creates a new optimizer.
func NewDataflowOptimizer(cfg *DataflowConfig, nop *NetworkOnPackage) *DataflowOptimizer {
	return &DataflowOptimizer{
		Config: cfg,
		NoP:    nop,
	}
}

// OptimizeMapping optimizes layer mapping to chiplets.
func (o *DataflowOptimizer) OptimizeMapping(layerShape LayerShape) *MappingResult {
	result := &MappingResult{
		ChipletAssignments: make([]int, 0),
		TileAssignments:    make([]TileAssignment, 0),
	}

	// Calculate tiles needed
	tilesX := (layerShape.OutputWidth + o.Config.TileSize - 1) / o.Config.TileSize
	tilesY := (layerShape.OutputHeight + o.Config.TileSize - 1) / o.Config.TileSize
	totalTiles := tilesX * tilesY

	// Distribute tiles across chiplets
	tilesPerChip := (totalTiles + len(o.NoP.Chiplets) - 1) / len(o.NoP.Chiplets)

	for i := 0; i < totalTiles; i++ {
		chiplet := i / tilesPerChip
		if chiplet >= len(o.NoP.Chiplets) {
			chiplet = len(o.NoP.Chiplets) - 1
		}

		result.ChipletAssignments = append(result.ChipletAssignments, chiplet)
		result.TileAssignments = append(result.TileAssignments, TileAssignment{
			TileID:    i,
			ChipletID: chiplet,
			StartX:    (i % tilesX) * o.Config.TileSize,
			StartY:    (i / tilesX) * o.Config.TileSize,
		})
	}

	// Estimate performance
	result.EstimatedLatency = o.estimateLatency(layerShape, totalTiles)
	result.EstimatedEnergy = o.estimateEnergy(layerShape, totalTiles)
	result.MemoryTraffic = o.estimateMemoryTraffic(layerShape)

	return result
}

// estimateLatency estimates execution latency.
func (o *DataflowOptimizer) estimateLatency(shape LayerShape, tiles int) float64 {
	// Compute time per tile
	opsPerTile := int64(o.Config.TileSize * o.Config.TileSize * shape.InputChannels)
	computeTimeNS := float64(opsPerTile) / 1000.0 // ~1 TOPS per chiplet

	// Communication overhead
	commOverhead := 0.0
	if tiles > len(o.NoP.Chiplets) {
		// Need to pipeline
		commOverhead = float64(tiles-len(o.NoP.Chiplets)) * 10.0 // 10 ns per extra tile
	}

	pipelineStages := float64(o.Config.PipelineDepth)
	return computeTimeNS*float64(tiles)/pipelineStages + commOverhead
}

// estimateEnergy estimates energy consumption.
func (o *DataflowOptimizer) estimateEnergy(shape LayerShape, tiles int) float64 {
	totalOps := int64(shape.OutputWidth * shape.OutputHeight * shape.InputChannels * shape.OutputChannels)
	energyPerOp := 0.5 // pJ per MAC

	// Add memory access energy
	memoryEnergy := float64(o.estimateMemoryTraffic(shape)) * 10.0 // 10 pJ per byte

	return float64(totalOps)*energyPerOp + memoryEnergy
}

// estimateMemoryTraffic estimates memory traffic in bytes.
func (o *DataflowOptimizer) estimateMemoryTraffic(shape LayerShape) int64 {
	// Weight traffic (once per layer)
	weightBytes := int64(shape.InputChannels * shape.OutputChannels * 4)

	// Activation traffic (tiled, depends on strategy)
	var activationBytes int64
	switch o.Config.Strategy {
	case "weight_stationary":
		// Weights stay, activations stream
		activationBytes = int64(shape.OutputWidth * shape.OutputHeight * shape.InputChannels * 4)
	case "output_stationary":
		// Output stays, weights and inputs stream
		activationBytes = int64(shape.OutputWidth*shape.OutputHeight*shape.InputChannels*4) + weightBytes
	default:
		activationBytes = int64(shape.OutputWidth * shape.OutputHeight * shape.InputChannels * 4)
	}

	// Reduction with tiling and buffering
	reuseFactor := float64(o.Config.TileSize) / 4.0
	if o.Config.DoubleBuffer {
		reuseFactor *= 1.5
	}

	return int64(float64(weightBytes+activationBytes) / reuseFactor)
}

// LayerShape describes a layer's dimensions.
type LayerShape struct {
	InputWidth     int
	InputHeight    int
	InputChannels  int
	OutputWidth    int
	OutputHeight   int
	OutputChannels int
	KernelSize     int
}

// TileAssignment holds tile-to-chiplet mapping.
type TileAssignment struct {
	TileID    int
	ChipletID int
	StartX    int
	StartY    int
}

// MappingResult holds optimization results.
type MappingResult struct {
	ChipletAssignments []int
	TileAssignments    []TileAssignment
	EstimatedLatency   float64   // ns
	EstimatedEnergy    float64   // pJ
	MemoryTraffic      int64     // bytes
}

// =============================================================================
// EVENT-BASED PROCESSING PIPELINE
// =============================================================================

// EventProcessingPipeline combines event camera with CIM accelerator.
type EventProcessingPipeline struct {
	Camera       *EventCamera
	Detector     *FEPhotodetector
	Accelerator  *NetworkOnPackage
	UseFE        bool    // Use ferroelectric detector
}

// NewEventProcessingPipeline creates a complete vision pipeline.
func NewEventProcessingPipeline(useFE bool) *EventProcessingPipeline {
	var camera *EventCamera
	var detector *FEPhotodetector

	if useFE {
		detector = NewFEPhotodetector(DefaultFEPhotodetectorConfig())
	} else {
		camera = NewEventCamera(DefaultEventCameraConfig())
	}

	nopCfg := DefaultNoPConfig()
	accelerator := NewNetworkOnPackage(nopCfg)

	return &EventProcessingPipeline{
		Camera:      camera,
		Detector:    detector,
		Accelerator: accelerator,
		UseFE:       useFE,
	}
}

// ProcessScene processes a scene and runs inference.
func (p *EventProcessingPipeline) ProcessScene(frames [][][]float64) (*PipelineResult, error) {
	result := &PipelineResult{
		EventCount:    0,
		InferenceOps:  0,
	}

	// Collect events from frames
	var events []*Event
	var feResponses []*FEResponse

	for t, frame := range frames {
		timestamp := float64(t) * 1000.0 // 1 ms between frames

		if p.UseFE {
			responses := p.Detector.ProcessFrame(frame, timestamp)
			feResponses = append(feResponses, responses...)
			for _, r := range responses {
				if r.IsSpike {
					result.EventCount++
				}
			}
		} else {
			frameEvents := p.Camera.ProcessFrame(frame, timestamp)
			events = append(events, frameEvents...)
			result.EventCount += int64(len(frameEvents))
		}
	}

	// Convert events to feature vector for inference
	featureSize := 64 * 64
	features := make([]float64, featureSize)

	if p.UseFE && len(feResponses) > 0 {
		// Use memory traces
		for _, r := range feResponses {
			idx := r.Y*64 + r.X
			if idx < featureSize {
				features[idx] = r.Value
			}
		}
	} else if len(events) > 0 {
		// Event histogram
		for _, e := range events {
			idx := e.Y*64 + e.X
			if idx < featureSize {
				features[idx] += float64(e.Polarity)
			}
		}
	}

	// Run inference on accelerator
	mvmResult, err := p.Accelerator.ExecuteDistributedMVM(features, 10, featureSize)
	if err != nil {
		return nil, err
	}

	result.InferenceOps = p.Accelerator.TotalOps
	result.TotalEnergy = p.Accelerator.TotalEnergy + mvmResult.TotalEnergy
	result.LatencyNS = mvmResult.TotalLatency
	result.Classification = argmax(mvmResult.Output)

	// Calculate data efficiency
	totalPixels := int64(len(frames)) * 320 * 320
	result.DataReduction = 1.0 - float64(result.EventCount)/float64(totalPixels)

	return result, nil
}

// argmax returns index of maximum value.
func argmax(arr []float64) int {
	maxIdx := 0
	maxVal := arr[0]
	for i, v := range arr {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}
	return maxIdx
}

// PipelineResult holds processing results.
type PipelineResult struct {
	EventCount     int64
	InferenceOps   int64
	TotalEnergy    float64   // pJ
	LatencyNS      float64
	Classification int
	DataReduction  float64   // Sparsity ratio
}

// =============================================================================
// BENCHMARK COMPARISON
// =============================================================================

// VisionBenchmark compares vision processing approaches.
type VisionBenchmark struct {
	Results map[string]*VisionMetrics
}

// VisionMetrics holds performance metrics.
type VisionMetrics struct {
	Name            string
	Resolution      string
	DataRate        float64   // Events/s or frames/s
	PowerMW         float64
	LatencyUS       float64
	DynamicRangeDB  float64
	DataEfficiency  float64   // Reduction vs frame-based
}

// NewVisionBenchmark creates a benchmark comparison.
func NewVisionBenchmark() *VisionBenchmark {
	return &VisionBenchmark{
		Results: make(map[string]*VisionMetrics),
	}
}

// AddDVS adds DVS metrics.
func (b *VisionBenchmark) AddDVS() {
	b.Results["DVS_320"] = &VisionMetrics{
		Name:           "DVS 320x320",
		Resolution:     "320x320",
		DataRate:       1e6,      // 1 Mev/s typical
		PowerMW:        10,
		LatencyUS:      100,
		DynamicRangeDB: 120,
		DataEfficiency: 0.99,     // 99% reduction
	}
}

// AddFEPhotodetector adds FE photodetector metrics.
func (b *VisionBenchmark) AddFEPhotodetector() {
	b.Results["FE_HZO"] = &VisionMetrics{
		Name:           "FE HZO Photodetector",
		Resolution:     "40x40",
		DataRate:       1e6,
		PowerMW:        1,        // Very low power
		LatencyUS:      73,       // 73 μs spikes
		DynamicRangeDB: 100,
		DataEfficiency: 0.95,
	}
}

// AddFrameCamera adds traditional camera metrics.
func (b *VisionBenchmark) AddFrameCamera() {
	b.Results["Frame_HD"] = &VisionMetrics{
		Name:           "HD Frame Camera",
		Resolution:     "1280x720",
		DataRate:       60,       // 60 fps
		PowerMW:        500,
		LatencyUS:      16000,    // 16 ms frame time
		DynamicRangeDB: 60,
		DataEfficiency: 0,        // Baseline
	}
}

// CompareVision generates comparison report.
func (b *VisionBenchmark) CompareVision() string {
	report := "Vision Sensor Comparison\n"
	report += "========================\n\n"

	// Sort by name for consistent output
	names := make([]string, 0, len(b.Results))
	for name := range b.Results {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		m := b.Results[name]
		report += fmt.Sprintf("%s:\n", m.Name)
		report += fmt.Sprintf("  Resolution: %s\n", m.Resolution)
		report += fmt.Sprintf("  Power: %.0f mW\n", m.PowerMW)
		report += fmt.Sprintf("  Latency: %.0f μs\n", m.LatencyUS)
		report += fmt.Sprintf("  Dynamic range: %.0f dB\n", m.DynamicRangeDB)
		report += fmt.Sprintf("  Data efficiency: %.0f%%\n\n", m.DataEfficiency*100)
	}

	return report
}
