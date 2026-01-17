// interconnect_ecc.go - Multi-Die CIM Interconnects and Error Correction Codes
// for Ferroelectric Compute-in-Memory (CIM) Systems
//
// This module implements:
// 1. UCIe-based chiplet interconnect simulation
// 2. Network-on-Chip (NoC) and Network-on-Package (NoP) modeling
// 3. SIAM-style mesh interconnect for CIM chiplets
// 4. Error correction codes (ECC) for CIM reliability
// 5. Residue Number System (RNS) for high-precision analog CIM
// 6. Successive correction and MAC-ECC schemes
//
// Based on research:
// - UCIe 3.0 specification (64 GT/s, <2ns latency)
// - SIAM: Chiplet-based Scalable In-Memory Acceleration
// - 3D-CIMlet framework for edge LLM inference
// - CIM-SECDED and MAC-ECC for reliability

package layers

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
)

// =============================================================================
// UCIe INTERCONNECT CONFIGURATION
// =============================================================================

// UCIeConfig configures Universal Chiplet Interconnect Express
type UCIeConfig struct {
	// Protocol version
	Version           string  // "1.0", "1.1", "2.0", "3.0"
	DataRate          float64 // GT/s (gigatransfers per second)
	BandwidthDensity  float64 // Tbps/mm

	// Physical parameters
	PackageType       string  // "Standard", "Advanced"
	TraceLength       float64 // mm (up to 25mm standard, 2mm advanced)
	NumLanes          int     // Number of lanes

	// Latency and power
	Latency           float64 // ns (<2ns typical)
	EnergyPerBit      float64 // pJ/bit

	// Features
	FECEnabled        bool    // Forward Error Correction
	RetrainSupport    bool    // Link retraining support
	CRCEnabled        bool    // Cyclic redundancy check
}

// DefaultUCIeConfig returns UCIe 3.0 configuration
func DefaultUCIeConfig() *UCIeConfig {
	return &UCIeConfig{
		Version:          "3.0",
		DataRate:         64.0,  // 64 GT/s
		BandwidthDensity: 20.0,  // 20 Tbps/mm
		PackageType:      "Advanced",
		TraceLength:      2.0,   // 2mm interposer
		NumLanes:         64,
		Latency:          1.5,   // 1.5 ns
		EnergyPerBit:     0.5,   // 0.5 pJ/bit
		FECEnabled:       true,
		RetrainSupport:   true,
		CRCEnabled:       true,
	}
}

// UCIeLink represents a UCIe die-to-die link
type UCIeLink struct {
	Config        *UCIeConfig
	SourceChiplet int
	DestChiplet   int
	Bandwidth     float64 // Gbps
	Utilization   float64 // 0-1
	ErrorRate     float64 // BER
}

// NewUCIeLink creates a UCIe link between chiplets
func NewUCIeLink(config *UCIeConfig, source, dest int) *UCIeLink {
	// Bandwidth = data rate * lanes * encoding efficiency
	encodingEff := 0.97 // 128b/130b encoding
	bandwidth := config.DataRate * float64(config.NumLanes) * encodingEff

	return &UCIeLink{
		Config:        config,
		SourceChiplet: source,
		DestChiplet:   dest,
		Bandwidth:     bandwidth,
		Utilization:   0.0,
		ErrorRate:     1e-15, // Very low BER with FEC
	}
}

// Transfer simulates data transfer over the link
func (link *UCIeLink) Transfer(dataSizeBits int) (float64, float64) {
	// Transfer time
	transferTime := float64(dataSizeBits) / link.Bandwidth // ns

	// Total latency = transfer time + protocol latency
	totalLatency := transferTime + link.Config.Latency

	// Energy
	energy := float64(dataSizeBits) * link.Config.EnergyPerBit // pJ

	// Update utilization
	link.Utilization = math.Min(1.0, link.Utilization+0.1)

	return totalLatency, energy
}

// =============================================================================
// NETWORK-ON-CHIP (NoC) MODEL
// =============================================================================

// NoCConfig configures Network-on-Chip
type NoCConfig struct {
	// Topology
	Topology      string // "Mesh", "Torus", "Ring", "Tree"
	Rows          int    // Grid rows
	Cols          int    // Grid cols

	// Router parameters
	BufferDepth   int     // Flits per buffer
	FlitWidth     int     // Bits per flit
	RouterLatency float64 // Cycles

	// Link parameters
	LinkBandwidth float64 // Gbps
	LinkLatency   float64 // ns

	// Power
	RouterPower   float64 // mW
	LinkPower     float64 // mW per mm
}

// DefaultNoCConfig returns default mesh NoC configuration
func DefaultNoCConfig(rows, cols int) *NoCConfig {
	return &NoCConfig{
		Topology:      "Mesh",
		Rows:          rows,
		Cols:          cols,
		BufferDepth:   4,
		FlitWidth:     64,
		RouterLatency: 1.0,
		LinkBandwidth: 128.0, // 128 Gbps
		LinkLatency:   0.5,   // 0.5 ns
		RouterPower:   10.0,  // 10 mW
		LinkPower:     5.0,   // 5 mW/mm
	}
}

// NoCRouter represents a router in the NoC
type NoCRouter struct {
	ID            int
	Row           int
	Col           int
	Neighbors     [4]*NoCRouter // N, S, E, W
	InputBuffers  [][]int       // Per-port input buffers
	OutputPorts   []bool        // Port availability
	PacketsRouted int
}

// NetworkOnChip represents the complete NoC
type NetworkOnChip struct {
	Config   *NoCConfig
	Routers  [][]*NoCRouter
	NumHops  int
	TotalLatency float64
	TotalEnergy  float64
}

// NewNetworkOnChip creates a mesh NoC
func NewNetworkOnChip(config *NoCConfig) *NetworkOnChip {
	noc := &NetworkOnChip{
		Config:  config,
		Routers: make([][]*NoCRouter, config.Rows),
	}

	// Create routers
	id := 0
	for r := 0; r < config.Rows; r++ {
		noc.Routers[r] = make([]*NoCRouter, config.Cols)
		for c := 0; c < config.Cols; c++ {
			noc.Routers[r][c] = &NoCRouter{
				ID:           id,
				Row:          r,
				Col:          c,
				InputBuffers: make([][]int, 5), // 4 ports + local
				OutputPorts:  make([]bool, 5),
			}
			for p := 0; p < 5; p++ {
				noc.Routers[r][c].InputBuffers[p] = make([]int, 0, config.BufferDepth)
				noc.Routers[r][c].OutputPorts[p] = true
			}
			id++
		}
	}

	// Connect neighbors (mesh topology)
	for r := 0; r < config.Rows; r++ {
		for c := 0; c < config.Cols; c++ {
			router := noc.Routers[r][c]
			// North
			if r > 0 {
				router.Neighbors[0] = noc.Routers[r-1][c]
			}
			// South
			if r < config.Rows-1 {
				router.Neighbors[1] = noc.Routers[r+1][c]
			}
			// East
			if c < config.Cols-1 {
				router.Neighbors[2] = noc.Routers[r][c+1]
			}
			// West
			if c > 0 {
				router.Neighbors[3] = noc.Routers[r][c-1]
			}
		}
	}

	return noc
}

// RouteXY performs XY routing between two routers
func (noc *NetworkOnChip) RouteXY(srcRow, srcCol, dstRow, dstCol int) ([]int, float64, float64) {
	path := []int{}
	hops := 0

	currentRow, currentCol := srcRow, srcCol

	// Route in X direction first
	for currentCol != dstCol {
		path = append(path, noc.Routers[currentRow][currentCol].ID)
		if currentCol < dstCol {
			currentCol++
		} else {
			currentCol--
		}
		hops++
	}

	// Then route in Y direction
	for currentRow != dstRow {
		path = append(path, noc.Routers[currentRow][currentCol].ID)
		if currentRow < dstRow {
			currentRow++
		} else {
			currentRow--
		}
		hops++
	}

	path = append(path, noc.Routers[dstRow][dstCol].ID)

	// Calculate latency and energy
	latency := float64(hops) * (noc.Config.RouterLatency + noc.Config.LinkLatency)
	energy := float64(hops) * (noc.Config.RouterPower*noc.Config.RouterLatency/1000.0 +
		noc.Config.LinkPower*noc.Config.LinkLatency/1000.0)

	noc.NumHops += hops
	noc.TotalLatency += latency
	noc.TotalEnergy += energy

	return path, latency, energy
}

// =============================================================================
// NETWORK-ON-PACKAGE (NoP) MODEL
// =============================================================================

// NoPConfig configures Network-on-Package
type NoPConfig struct {
	// Topology
	NumChiplets     int
	Topology        string // "Mesh", "Ring", "FullyConnected"

	// UCIe links
	UCIeConfig      *UCIeConfig

	// Interposer parameters
	InterposerType  string  // "Silicon", "Organic", "Glass"
	TraceLength     float64 // mm between chiplets
	TracePitch      float64 // μm

	// Power
	IOPower         float64 // mW per chiplet
}

// DefaultNoPConfig returns default NoP configuration
func DefaultNoPConfig(numChiplets int) *NoPConfig {
	return &NoPConfig{
		NumChiplets:    numChiplets,
		Topology:       "Mesh",
		UCIeConfig:     DefaultUCIeConfig(),
		InterposerType: "Silicon",
		TraceLength:    2.0,  // 2mm
		TracePitch:     25.0, // 25μm
		IOPower:        100.0, // 100mW
	}
}

// NetworkOnPackage represents chiplet interconnect
type NetworkOnPackage struct {
	Config     *NoPConfig
	Chiplets   []*CIMChiplet
	Links      [][]*UCIeLink
	Statistics *NoPStatistics
}

// NoPStatistics tracks NoP performance
type NoPStatistics struct {
	TotalTransfers    int
	TotalDataBits     int64
	TotalLatency      float64
	TotalEnergy       float64
	AverageBandwidth  float64
}

// CIMChiplet represents a single CIM chiplet
type CIMChiplet struct {
	ID            int
	Type          string // "SRAM-CIM", "RRAM-CIM", "FeFET-CIM"
	ArraySize     int    // Crossbar size
	NumArrays     int    // Number of crossbar arrays
	LocalNoC      *NetworkOnChip
	BufferSize    int    // KB
	Utilization   float64
}

// NewNetworkOnPackage creates a NoP with CIM chiplets
func NewNetworkOnPackage(config *NoPConfig) *NetworkOnPackage {
	nop := &NetworkOnPackage{
		Config:     config,
		Chiplets:   make([]*CIMChiplet, config.NumChiplets),
		Links:      make([][]*UCIeLink, config.NumChiplets),
		Statistics: &NoPStatistics{},
	}

	// Create chiplets
	for i := 0; i < config.NumChiplets; i++ {
		// Each chiplet has internal NoC (4x4 mesh)
		nocConfig := DefaultNoCConfig(4, 4)
		nop.Chiplets[i] = &CIMChiplet{
			ID:         i,
			Type:       "FeFET-CIM",
			ArraySize:  256,
			NumArrays:  16,
			LocalNoC:   NewNetworkOnChip(nocConfig),
			BufferSize: 256, // 256 KB
		}
	}

	// Create UCIe links based on topology
	for i := 0; i < config.NumChiplets; i++ {
		nop.Links[i] = make([]*UCIeLink, config.NumChiplets)
	}

	// Mesh connectivity (assume sqrt arrangement)
	gridSize := int(math.Ceil(math.Sqrt(float64(config.NumChiplets))))
	for i := 0; i < config.NumChiplets; i++ {
		row := i / gridSize
		col := i % gridSize

		// Right neighbor
		if col < gridSize-1 && i+1 < config.NumChiplets {
			nop.Links[i][i+1] = NewUCIeLink(config.UCIeConfig, i, i+1)
			nop.Links[i+1][i] = NewUCIeLink(config.UCIeConfig, i+1, i)
		}

		// Bottom neighbor
		if row < gridSize-1 && i+gridSize < config.NumChiplets {
			nop.Links[i][i+gridSize] = NewUCIeLink(config.UCIeConfig, i, i+gridSize)
			nop.Links[i+gridSize][i] = NewUCIeLink(config.UCIeConfig, i+gridSize, i)
		}
	}

	return nop
}

// TransferBetweenChiplets transfers data between chiplets
func (nop *NetworkOnPackage) TransferBetweenChiplets(src, dst, dataSizeBits int) (float64, float64) {
	if src == dst {
		return 0, 0
	}

	// Find path using mesh routing
	gridSize := int(math.Ceil(math.Sqrt(float64(nop.Config.NumChiplets))))
	srcRow, srcCol := src/gridSize, src%gridSize
	dstRow, dstCol := dst/gridSize, dst%gridSize

	// XY routing on package mesh
	totalLatency := 0.0
	totalEnergy := 0.0
	current := src

	// Route X first
	for srcCol != dstCol {
		next := current
		if srcCol < dstCol {
			next = current + 1
			srcCol++
		} else {
			next = current - 1
			srcCol--
		}

		if nop.Links[current][next] != nil {
			lat, en := nop.Links[current][next].Transfer(dataSizeBits)
			totalLatency += lat
			totalEnergy += en
		}
		current = next
	}

	// Then route Y
	for srcRow != dstRow {
		next := current
		if srcRow < dstRow {
			next = current + gridSize
			srcRow++
		} else {
			next = current - gridSize
			srcRow--
		}

		if nop.Links[current][next] != nil {
			lat, en := nop.Links[current][next].Transfer(dataSizeBits)
			totalLatency += lat
			totalEnergy += en
		}
		current = next
	}

	// Update statistics
	nop.Statistics.TotalTransfers++
	nop.Statistics.TotalDataBits += int64(dataSizeBits)
	nop.Statistics.TotalLatency += totalLatency
	nop.Statistics.TotalEnergy += totalEnergy

	return totalLatency, totalEnergy
}

// =============================================================================
// ERROR CORRECTION CODE CONFIGURATION
// =============================================================================

// ECCConfig configures error correction for CIM
type ECCConfig struct {
	// ECC type
	ECCType          string // "SECDED", "BCH", "RS", "LDPC", "MAC-ECC"
	DataBits         int    // Original data width
	ParityBits       int    // ECC parity bits
	CorrectionCap    int    // Number of correctable errors

	// CIM-specific
	SuccessiveCorrection bool    // Enable successive correction
	BitSlicing           bool    // Enable bit slicing
	RNSEnabled           bool    // Enable Residue Number System

	// Performance trade-offs
	AreaOverhead     float64 // Percentage
	PowerOverhead    float64 // Percentage
	LatencyOverhead  float64 // Cycles
}

// DefaultECCConfig returns CIM-optimized ECC configuration
func DefaultECCConfig() *ECCConfig {
	return &ECCConfig{
		ECCType:              "MAC-ECC",
		DataBits:             64,
		ParityBits:           8,
		CorrectionCap:        2,
		SuccessiveCorrection: true,
		BitSlicing:           true,
		RNSEnabled:           false,
		AreaOverhead:         29.1,
		PowerOverhead:        26.3,
		LatencyOverhead:      2.0,
	}
}

// =============================================================================
// SECDED IMPLEMENTATION
// =============================================================================

// SECDEDCodec implements Single Error Correction, Double Error Detection
type SECDEDCodec struct {
	Config         *ECCConfig
	HMatrix        [][]int // Parity check matrix
	GMatrix        [][]int // Generator matrix

	// Statistics
	SingleErrors   int
	DoubleErrors   int
	Corrections    int
}

// NewSECDEDCodec creates a SECDED codec
func NewSECDEDCodec(dataBits int) *SECDEDCodec {
	// Calculate required parity bits: 2^r >= m + r + 1
	parityBits := 0
	for (1 << parityBits) < dataBits+parityBits+1 {
		parityBits++
	}

	codec := &SECDEDCodec{
		Config: &ECCConfig{
			ECCType:       "SECDED",
			DataBits:      dataBits,
			ParityBits:    parityBits + 1, // +1 for overall parity
			CorrectionCap: 1,
		},
	}

	// Build H matrix (simplified)
	codec.HMatrix = make([][]int, parityBits+1)
	totalBits := dataBits + parityBits + 1
	for i := 0; i <= parityBits; i++ {
		codec.HMatrix[i] = make([]int, totalBits)
	}

	// Standard Hamming parity positions
	for col := 0; col < totalBits; col++ {
		for row := 0; row < parityBits; row++ {
			if ((col + 1) & (1 << row)) != 0 {
				codec.HMatrix[row][col] = 1
			}
		}
		// Overall parity (last row)
		codec.HMatrix[parityBits][col] = 1
	}

	return codec
}

// Encode adds SECDED parity to data
func (codec *SECDEDCodec) Encode(data []int) []int {
	totalBits := codec.Config.DataBits + codec.Config.ParityBits
	encoded := make([]int, totalBits)

	// Copy data bits to non-parity positions
	dataIdx := 0
	for i := 0; i < totalBits-1; i++ {
		// Skip power-of-2 positions (parity bits)
		if (i+1)&i == 0 {
			continue
		}
		if dataIdx < len(data) {
			encoded[i] = data[dataIdx]
			dataIdx++
		}
	}

	// Calculate parity bits
	for p := 0; p < codec.Config.ParityBits-1; p++ {
		parityPos := (1 << p) - 1
		parity := 0
		for i := 0; i < totalBits-1; i++ {
			if ((i + 1) & (1 << p)) != 0 {
				parity ^= encoded[i]
			}
		}
		encoded[parityPos] = parity
	}

	// Overall parity
	overallParity := 0
	for i := 0; i < totalBits-1; i++ {
		overallParity ^= encoded[i]
	}
	encoded[totalBits-1] = overallParity

	return encoded
}

// Decode performs SECDED decoding with error correction
func (codec *SECDEDCodec) Decode(received []int) ([]int, int, bool) {
	// Calculate syndrome
	syndrome := 0
	for p := 0; p < codec.Config.ParityBits-1; p++ {
		parity := 0
		for i := 0; i < len(received)-1; i++ {
			if ((i + 1) & (1 << p)) != 0 {
				parity ^= received[i]
			}
		}
		if parity != 0 {
			syndrome |= (1 << p)
		}
	}

	// Check overall parity
	overallParity := 0
	for _, bit := range received {
		overallParity ^= bit
	}

	// Determine error status
	corrected := make([]int, len(received))
	copy(corrected, received)
	errorsDetected := 0

	if syndrome == 0 && overallParity == 0 {
		// No errors
		errorsDetected = 0
	} else if syndrome != 0 && overallParity != 0 {
		// Single error - correctable
		errorPos := syndrome - 1
		if errorPos < len(corrected) {
			corrected[errorPos] ^= 1
			codec.SingleErrors++
			codec.Corrections++
		}
		errorsDetected = 1
	} else if syndrome != 0 && overallParity == 0 {
		// Double error - detectable but not correctable
		codec.DoubleErrors++
		errorsDetected = 2
		return corrected, errorsDetected, false
	}

	// Extract data bits
	data := make([]int, codec.Config.DataBits)
	dataIdx := 0
	for i := 0; i < len(corrected)-1; i++ {
		if (i+1)&i != 0 { // Not a power of 2
			if dataIdx < len(data) {
				data[dataIdx] = corrected[i]
				dataIdx++
			}
		}
	}

	return data, errorsDetected, true
}

// =============================================================================
// MAC-ECC FOR CIM
// =============================================================================

// MACECCCodec implements MAC-level error correction for CIM
type MACECCCodec struct {
	Config          *ECCConfig
	NumRows         int     // Parallel row accesses
	Precision       int     // ADC precision bits

	// Redundancy
	ChecksumRows    int     // Additional checksum rows
	ChecksumCols    int     // Additional checksum columns

	// Statistics
	MACErrors       int
	CellErrors      int
	Corrections     int
}

// NewMACECCCodec creates a MAC-ECC codec
func NewMACECCCodec(numRows, precision int) *MACECCCodec {
	// Checksum overhead based on precision
	checksumRows := (precision + 3) / 4 // ~25% overhead
	checksumCols := 2                    // Column checksums

	return &MACECCCodec{
		Config: &ECCConfig{
			ECCType:       "MAC-ECC",
			DataBits:      numRows * precision,
			ParityBits:    checksumRows * precision,
			CorrectionCap: 2,
			AreaOverhead:  29.1,
			PowerOverhead: 26.3,
		},
		NumRows:      numRows,
		Precision:    precision,
		ChecksumRows: checksumRows,
		ChecksumCols: checksumCols,
	}
}

// ComputeRowChecksum computes checksum for a row
func (codec *MACECCCodec) ComputeRowChecksum(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum
}

// ComputeColumnChecksum computes checksum for a column
func (codec *MACECCCodec) ComputeColumnChecksum(matrix [][]float64, col int) float64 {
	sum := 0.0
	for row := 0; row < len(matrix); row++ {
		if col < len(matrix[row]) {
			sum += matrix[row][col]
		}
	}
	return sum
}

// VerifyMACOutput verifies CIM MAC output
func (codec *MACECCCodec) VerifyMACOutput(weights [][]float64, inputs []float64, output []float64) ([]float64, bool) {
	// Compute expected checksums
	expectedRowChecksums := make([]float64, len(weights))
	for r := range weights {
		expectedRowChecksums[r] = codec.ComputeRowChecksum(weights[r])
	}

	// Compute output with checksums
	correctedOutput := make([]float64, len(output))
	copy(correctedOutput, output)

	// Verify each output
	hasError := false
	for col := 0; col < len(output); col++ {
		// Compute expected output from weights and inputs
		expected := 0.0
		for row := 0; row < len(inputs) && row < len(weights); row++ {
			if col < len(weights[row]) {
				expected += weights[row][col] * inputs[row]
			}
		}

		// Check for error
		error := math.Abs(output[col] - expected)
		threshold := 0.01 * math.Abs(expected) // 1% tolerance

		if error > threshold && expected != 0 {
			hasError = true
			codec.MACErrors++
			// Correct using expected value
			correctedOutput[col] = expected
			codec.Corrections++
		}
	}

	return correctedOutput, !hasError
}

// =============================================================================
// SUCCESSIVE CORRECTION
// =============================================================================

// SuccessiveCorrectionCodec implements iterative error correction
type SuccessiveCorrectionCodec struct {
	Config       *ECCConfig
	InnerCodec   *SECDEDCodec
	MaxIters     int
	Threshold    float64 // Confidence threshold

	// Statistics
	Iterations   int
	Corrections  int
	Failures     int
}

// NewSuccessiveCorrectionCodec creates a successive correction codec
func NewSuccessiveCorrectionCodec(dataBits, maxIters int) *SuccessiveCorrectionCodec {
	return &SuccessiveCorrectionCodec{
		Config: &ECCConfig{
			ECCType:              "Successive",
			DataBits:             dataBits,
			SuccessiveCorrection: true,
		},
		InnerCodec: NewSECDEDCodec(dataBits),
		MaxIters:   maxIters,
		Threshold:  0.9,
	}
}

// CorrectSuccessively performs iterative error correction
func (codec *SuccessiveCorrectionCodec) CorrectSuccessively(received []int, softInfo []float64) ([]int, bool) {
	current := make([]int, len(received))
	copy(current, received)

	for iter := 0; iter < codec.MaxIters; iter++ {
		codec.Iterations++

		// Decode with inner codec
		decoded, errors, success := codec.InnerCodec.Decode(current)

		if success && errors == 0 {
			return decoded, true
		}

		if errors > codec.InnerCodec.Config.CorrectionCap {
			// Use soft information to flip least confident bits
			if len(softInfo) > 0 {
				// Find least confident bit
				minConf := math.MaxFloat64
				minIdx := 0
				for i, conf := range softInfo {
					if math.Abs(conf-0.5) < minConf {
						minConf = math.Abs(conf - 0.5)
						minIdx = i
					}
				}
				// Flip it
				if minIdx < len(current) {
					current[minIdx] ^= 1
					codec.Corrections++
				}
			}
		} else {
			// Re-encode and continue
			encoded := codec.InnerCodec.Encode(decoded)
			copy(current, encoded)
		}
	}

	codec.Failures++
	return current[:codec.Config.DataBits], false
}

// =============================================================================
// RESIDUE NUMBER SYSTEM (RNS)
// =============================================================================

// RNSConfig configures Residue Number System
type RNSConfig struct {
	Moduli       []int   // Co-prime moduli
	DynamicRange int64   // Product of all moduli
	Precision    int     // Equivalent bit precision
}

// DefaultRNSConfig returns RNS configuration for 16-bit precision
func DefaultRNSConfig() *RNSConfig {
	// Use co-prime moduli: 17, 19, 23, 29 (product = 214,913)
	moduli := []int{17, 19, 23, 29}
	dynamicRange := int64(1)
	for _, m := range moduli {
		dynamicRange *= int64(m)
	}

	return &RNSConfig{
		Moduli:       moduli,
		DynamicRange: dynamicRange,
		Precision:    17, // log2(214913) ≈ 17 bits
	}
}

// RNSCodec implements RNS encoding/decoding
type RNSCodec struct {
	Config *RNSConfig
}

// NewRNSCodec creates an RNS codec
func NewRNSCodec(config *RNSConfig) *RNSCodec {
	return &RNSCodec{Config: config}
}

// Encode converts integer to RNS representation
func (codec *RNSCodec) Encode(value int64) []int {
	residues := make([]int, len(codec.Config.Moduli))
	for i, m := range codec.Config.Moduli {
		residues[i] = int(((value % int64(m)) + int64(m)) % int64(m))
	}
	return residues
}

// Decode converts RNS representation back to integer (CRT)
func (codec *RNSCodec) Decode(residues []int) int64 {
	// Chinese Remainder Theorem
	M := codec.Config.DynamicRange
	result := int64(0)

	for i, ri := range residues {
		mi := int64(codec.Config.Moduli[i])
		Mi := M / mi
		yi := codec.modInverse(Mi, mi)
		result += int64(ri) * Mi * yi
	}

	return ((result % M) + M) % M
}

// modInverse computes modular multiplicative inverse
func (codec *RNSCodec) modInverse(a, m int64) int64 {
	// Extended Euclidean algorithm
	if m == 1 {
		return 0
	}

	m0 := m
	x0, x1 := int64(0), int64(1)

	for a > 1 {
		q := a / m
		m, a = a%m, m
		x0, x1 = x1-q*x0, x0
	}

	if x1 < 0 {
		x1 += m0
	}

	return x1
}

// AddRNS adds two RNS values
func (codec *RNSCodec) AddRNS(a, b []int) []int {
	result := make([]int, len(codec.Config.Moduli))
	for i, m := range codec.Config.Moduli {
		result[i] = (a[i] + b[i]) % m
	}
	return result
}

// MulRNS multiplies two RNS values
func (codec *RNSCodec) MulRNS(a, b []int) []int {
	result := make([]int, len(codec.Config.Moduli))
	for i, m := range codec.Config.Moduli {
		result[i] = (a[i] * b[i]) % m
	}
	return result
}

// DotProductRNS computes dot product in RNS domain
func (codec *RNSCodec) DotProductRNS(weights [][]int, inputs [][]int) [][]int {
	if len(weights) == 0 || len(inputs) == 0 {
		return nil
	}

	numOutputs := len(weights[0]) / len(codec.Config.Moduli)
	results := make([][]int, numOutputs)

	for o := 0; o < numOutputs; o++ {
		results[o] = make([]int, len(codec.Config.Moduli))
		// Initialize to zero
		for m := range codec.Config.Moduli {
			results[o][m] = 0
		}

		// Accumulate
		for i := 0; i < len(inputs); i++ {
			// Multiply weight[i] * input[i]
			wStart := o * len(codec.Config.Moduli)
			for m, mod := range codec.Config.Moduli {
				if wStart+m < len(weights[i]) && m < len(inputs[i]) {
					results[o][m] = (results[o][m] + weights[i][wStart+m]*inputs[i][m]) % mod
				}
			}
		}
	}

	return results
}

// =============================================================================
// BIT SLICING FOR CIM
// =============================================================================

// BitSlicingConfig configures bit slicing
type BitSlicingConfig struct {
	InputBits     int // Original input precision
	WeightBits    int // Original weight precision
	SliceBits     int // Bits per slice
	NumSlices     int // Number of slices
}

// BitSlicer implements bit slicing for CIM
type BitSlicer struct {
	Config *BitSlicingConfig
}

// NewBitSlicer creates a bit slicer
func NewBitSlicer(inputBits, weightBits, sliceBits int) *BitSlicer {
	numSlices := (weightBits + sliceBits - 1) / sliceBits

	return &BitSlicer{
		Config: &BitSlicingConfig{
			InputBits:  inputBits,
			WeightBits: weightBits,
			SliceBits:  sliceBits,
			NumSlices:  numSlices,
		},
	}
}

// SliceWeights slices weight matrix into multiple low-precision slices
func (bs *BitSlicer) SliceWeights(weights [][]int) [][][]int {
	slices := make([][][]int, bs.Config.NumSlices)
	mask := (1 << bs.Config.SliceBits) - 1

	for s := 0; s < bs.Config.NumSlices; s++ {
		slices[s] = make([][]int, len(weights))
		shift := s * bs.Config.SliceBits

		for i := range weights {
			slices[s][i] = make([]int, len(weights[i]))
			for j := range weights[i] {
				slices[s][i][j] = (weights[i][j] >> shift) & mask
			}
		}
	}

	return slices
}

// CombineSlices combines partial results with proper scaling
func (bs *BitSlicer) CombineSlices(partialResults [][]float64) []float64 {
	if len(partialResults) == 0 {
		return nil
	}

	result := make([]float64, len(partialResults[0]))

	for s := 0; s < len(partialResults); s++ {
		scale := float64(int(1) << (s * bs.Config.SliceBits))
		for i := range result {
			result[i] += partialResults[s][i] * scale
		}
	}

	return result
}

// =============================================================================
// INTEGRATED CIM INTERCONNECT + ECC SYSTEM
// =============================================================================

// CIMInterconnectECCSystem integrates interconnect and ECC
type CIMInterconnectECCSystem struct {
	// Configuration
	NumChiplets  int
	ArraySize    int

	// Components
	NoP          *NetworkOnPackage
	ECCCodec     *MACECCCodec
	RNSCodec     *RNSCodec
	BitSlicer    *BitSlicer

	// Statistics
	TotalMACs        int64
	TotalTransfers   int64
	ErrorsCorrected  int
	TotalLatency     float64
	TotalEnergy      float64
}

// NewCIMInterconnectECCSystem creates an integrated system
func NewCIMInterconnectECCSystem(numChiplets, arraySize int) *CIMInterconnectECCSystem {
	nopConfig := DefaultNoPConfig(numChiplets)

	return &CIMInterconnectECCSystem{
		NumChiplets: numChiplets,
		ArraySize:   arraySize,
		NoP:         NewNetworkOnPackage(nopConfig),
		ECCCodec:    NewMACECCCodec(arraySize, 8),
		RNSCodec:    NewRNSCodec(DefaultRNSConfig()),
		BitSlicer:   NewBitSlicer(8, 8, 4),
	}
}

// DistributeLayer distributes a layer across chiplets
func (sys *CIMInterconnectECCSystem) DistributeLayer(weights [][]float64) map[int][][]float64 {
	distribution := make(map[int][][]float64)

	// Simple row-based distribution
	rowsPerChiplet := (len(weights) + sys.NumChiplets - 1) / sys.NumChiplets

	for c := 0; c < sys.NumChiplets; c++ {
		startRow := c * rowsPerChiplet
		endRow := startRow + rowsPerChiplet
		if endRow > len(weights) {
			endRow = len(weights)
		}
		if startRow < len(weights) {
			distribution[c] = weights[startRow:endRow]
		}
	}

	return distribution
}

// ExecuteDistributedMAC executes MAC across chiplets with ECC
func (sys *CIMInterconnectECCSystem) ExecuteDistributedMAC(distribution map[int][][]float64, inputs []float64) []float64 {
	partialResults := make(map[int][]float64)

	// Execute on each chiplet
	for chipletID, localWeights := range distribution {
		// Simulate local MAC
		localResult := make([]float64, len(localWeights[0]))
		for col := range localResult {
			sum := 0.0
			for row := range localWeights {
				if row < len(inputs) {
					sum += localWeights[row][col] * inputs[row]
				}
			}
			localResult[col] = sum
		}

		// Add noise (simulating CIM non-idealities)
		for i := range localResult {
			localResult[i] += rand.NormFloat64() * 0.01 * localResult[i]
		}

		// ECC verification
		corrected, _ := sys.ECCCodec.VerifyMACOutput(localWeights, inputs, localResult)
		partialResults[chipletID] = corrected

		sys.TotalMACs += int64(len(localWeights) * len(localWeights[0]))
	}

	// Transfer and accumulate results
	output := make([]float64, len(partialResults[0]))
	for chipletID, partial := range partialResults {
		if chipletID > 0 {
			// Transfer to chiplet 0 for accumulation
			dataBits := len(partial) * 32 // 32-bit floats
			lat, en := sys.NoP.TransferBetweenChiplets(chipletID, 0, dataBits)
			sys.TotalLatency += lat
			sys.TotalEnergy += en
			sys.TotalTransfers++
		}

		for i := range partial {
			output[i] += partial[i]
		}
	}

	return output
}

// GetPerformanceReport returns system performance summary
func (sys *CIMInterconnectECCSystem) GetPerformanceReport() string {
	return fmt.Sprintf(`
CIM Interconnect + ECC System Performance
==========================================
Chiplets: %d
Array Size: %dx%d

Interconnect Statistics:
  Total Transfers: %d
  Total Data: %.2f Mb
  Avg Latency: %.2f ns
  Avg Energy: %.2f pJ

Compute Statistics:
  Total MACs: %d
  Errors Corrected: %d

UCIe Configuration:
  Version: %s
  Data Rate: %.0f GT/s
  Bandwidth Density: %.1f Tbps/mm
`,
		sys.NumChiplets,
		sys.ArraySize, sys.ArraySize,
		sys.NoP.Statistics.TotalTransfers,
		float64(sys.NoP.Statistics.TotalDataBits)/1e6,
		sys.NoP.Statistics.TotalLatency/float64(max(1, sys.NoP.Statistics.TotalTransfers)),
		sys.NoP.Statistics.TotalEnergy/float64(max(1, sys.NoP.Statistics.TotalTransfers)),
		sys.TotalMACs,
		sys.ECCCodec.Corrections,
		sys.NoP.Config.UCIeConfig.Version,
		sys.NoP.Config.UCIeConfig.DataRate,
		sys.NoP.Config.UCIeConfig.BandwidthDensity,
	)
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// =============================================================================
// IRONLATTICE INTEGRATED SYSTEM
// =============================================================================

// IronLatticeInterconnectECC represents the complete IronLattice multi-die system
type IronLatticeInterconnectECC struct {
	// Core system
	CIMSystem *CIMInterconnectECCSystem

	// HZO ferroelectric parameters
	HZOParameters struct {
		EnduranceCycles float64
		Retention       float64 // Years
		SwitchingEnergy float64 // fJ
	}

	// Configuration
	ECCScheme         string
	InterconnectType  string

	// Performance targets
	TargetThroughput float64 // TOPS
	TargetEfficiency float64 // TOPS/W
}

// NewIronLatticeInterconnectECC creates an IronLattice multi-die system
func NewIronLatticeInterconnectECC(numChiplets, arraySize int) *IronLatticeInterconnectECC {
	system := &IronLatticeInterconnectECC{
		CIMSystem:        NewCIMInterconnectECCSystem(numChiplets, arraySize),
		ECCScheme:        "MAC-ECC",
		InterconnectType: "UCIe-3.0",
	}

	// HZO parameters
	system.HZOParameters.EnduranceCycles = 1e10
	system.HZOParameters.Retention = 10.0
	system.HZOParameters.SwitchingEnergy = 10.0 // 10 fJ

	// Performance targets based on 3D-CIMlet results
	system.TargetThroughput = 100.0  // 100 TOPS
	system.TargetEfficiency = 300.0  // 300 TOPS/W

	return system
}

// ExportJSON exports system configuration
func (ils *IronLatticeInterconnectECC) ExportJSON() ([]byte, error) {
	export := map[string]interface{}{
		"num_chiplets":      ils.CIMSystem.NumChiplets,
		"array_size":        ils.CIMSystem.ArraySize,
		"ecc_scheme":        ils.ECCScheme,
		"interconnect_type": ils.InterconnectType,
		"hzo_parameters":    ils.HZOParameters,
		"performance_targets": map[string]float64{
			"throughput_tops":    ils.TargetThroughput,
			"efficiency_tops_w":  ils.TargetEfficiency,
		},
		"nop_statistics": map[string]interface{}{
			"total_transfers": ils.CIMSystem.NoP.Statistics.TotalTransfers,
			"total_data_bits": ils.CIMSystem.NoP.Statistics.TotalDataBits,
			"total_latency":   ils.CIMSystem.NoP.Statistics.TotalLatency,
			"total_energy":    ils.CIMSystem.NoP.Statistics.TotalEnergy,
		},
	}

	return json.MarshalIndent(export, "", "  ")
}
