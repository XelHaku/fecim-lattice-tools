// cam_attention.go - Ferroelectric Content-Addressable Memory and Transformer Attention Mapping
//
// This module implements:
// 1. Ferroelectric TCAM cells (2FeFET, 2FeFET-1T, 2FeFET-2T)
// 2. Combination-encoding CAM (CECAM) for higher content density
// 3. Analog CAM (ACAM) for neural network search
// 4. Transformer attention mechanism mapping to crossbar arrays
// 5. KV cache optimization using ferroelectric memory
// 6. ReCAT cascaded crossbar architecture for attention
//
// References:
// - CECAM HZO FeFET: 65% power reduction (ACS AEM 2024)
// - FACAM: 60× energy, 2700× latency reduction vs GPU
// - AIMC Attention: 70,000× energy, 100× speed vs GPU (Nature Comp Sci 2025)
// - ReCAT: Cascaded crossbar for transformer (ACM TODAES 2024)

package layers

import (
	"math"
	"math/rand"
)

// =============================================================================
// FERROELECTRIC CONTENT-ADDRESSABLE MEMORY (CAM)
// =============================================================================

// CAMCellType identifies the CAM cell architecture
type CAMCellType int

const (
	CAMCell2FeFET   CAMCellType = iota // 2FeFET cell (most compact)
	CAMCell2FeFET1T                    // 2FeFET-1T cell
	CAMCell2FeFET2T                    // 2FeFET-2T cell
	CAMCellHFNN                        // HFNN TCAM design
	CAMCellCECAM                       // Combination-encoding CAM
	CAMCellACAM                        // Analog CAM (multi-bit)
)

// FeFETCAMConfig configures ferroelectric CAM cell
type FeFETCAMConfig struct {
	CellType         CAMCellType
	NumBits          int     // bits per cell (1 for binary, 2 for ternary)
	Threshold        float64 // V (match threshold voltage)
	FeFETOnOff       float64 // on/off ratio (>10^6 typical)
	MatchlineLength  int     // cells per matchline
	SearchParallelism int    // simultaneous searches

	// Power parameters
	WriteVoltage     float64 // V
	SearchVoltage    float64 // V
	StandbyPower     float64 // nW (near zero for ferroelectric)
}

// DefaultFeFETCAMConfig returns default configuration
func DefaultFeFETCAMConfig() *FeFETCAMConfig {
	return &FeFETCAMConfig{
		CellType:          CAMCell2FeFET,
		NumBits:           2, // ternary (0, 1, X)
		Threshold:         0.3,
		FeFETOnOff:        1e6,
		MatchlineLength:   128,
		SearchParallelism: 64,
		WriteVoltage:      3.0,
		SearchVoltage:     0.5,
		StandbyPower:      0.001, // nW (negligible)
	}
}

// FeFETCAMCell represents a single ferroelectric CAM cell
type FeFETCAMCell struct {
	Config *FeFETCAMConfig

	// Cell state (ternary: 0, 1, X)
	StoredValue int // 0, 1, or 2 (don't care)
	Polarization [2]float64 // for 2FeFET cells

	// Performance metrics
	SearchEnergy float64 // fJ per search
	WriteEnergy  float64 // fJ per write
	MatchDelay   float64 // ns
}

// NewFeFETCAMCell creates a new FeFET CAM cell
func NewFeFETCAMCell(config *FeFETCAMConfig) *FeFETCAMCell {
	if config == nil {
		config = DefaultFeFETCAMConfig()
	}

	cell := &FeFETCAMCell{
		Config:       config,
		StoredValue:  2, // don't care initially
		Polarization: [2]float64{0, 0},
	}

	cell.calculatePerformance()

	return cell
}

// calculatePerformance computes cell metrics
func (c *FeFETCAMCell) calculatePerformance() {
	config := c.Config

	// Energy based on cell type
	switch config.CellType {
	case CAMCell2FeFET:
		c.SearchEnergy = 0.5  // fJ (most efficient)
		c.WriteEnergy = 10.0  // fJ
		c.MatchDelay = 0.5    // ns
	case CAMCell2FeFET1T:
		c.SearchEnergy = 1.0  // fJ
		c.WriteEnergy = 12.0  // fJ
		c.MatchDelay = 0.8    // ns
	case CAMCell2FeFET2T:
		c.SearchEnergy = 1.5  // fJ
		c.WriteEnergy = 15.0  // fJ
		c.MatchDelay = 1.0    // ns
	case CAMCellHFNN:
		c.SearchEnergy = 0.1  // fJ (226× reduction)
		c.WriteEnergy = 8.0   // fJ
		c.MatchDelay = 0.3    // ns
	case CAMCellCECAM:
		c.SearchEnergy = 0.3  // fJ (65% reduction)
		c.WriteEnergy = 10.0  // fJ
		c.MatchDelay = 0.4    // ns
	case CAMCellACAM:
		c.SearchEnergy = 2.0  // fJ (multi-bit)
		c.WriteEnergy = 20.0  // fJ
		c.MatchDelay = 1.5    // ns
	}

	// Scale by voltage
	voltageScale := math.Pow(config.SearchVoltage/0.5, 2)
	c.SearchEnergy *= voltageScale
}

// Write programs the CAM cell with a ternary value
func (c *FeFETCAMCell) Write(value int) {
	c.StoredValue = value

	// Set polarization states for 2FeFET cell
	switch value {
	case 0: // Store 0
		c.Polarization[0] = 1.0  // FeFET1 ON
		c.Polarization[1] = -1.0 // FeFET2 OFF
	case 1: // Store 1
		c.Polarization[0] = -1.0 // FeFET1 OFF
		c.Polarization[1] = 1.0  // FeFET2 ON
	case 2: // Don't care (X)
		c.Polarization[0] = 1.0 // Both ON
		c.Polarization[1] = 1.0
	}
}

// Search compares input with stored value
func (c *FeFETCAMCell) Search(input int) bool {
	// Don't care always matches
	if c.StoredValue == 2 {
		return true
	}
	return c.StoredValue == input
}

// =============================================================================
// FERROELECTRIC TCAM ARRAY
// =============================================================================

// FeTCAMConfig configures ferroelectric TCAM array
type FeTCAMConfig struct {
	Rows         int // number of entries (words)
	Cols         int // bits per entry (word width)
	CellConfig   *FeFETCAMConfig
	MatchType    string // "exact", "hamming", "threshold"
}

// DefaultFeTCAMConfig returns default TCAM configuration
func DefaultFeTCAMConfig() *FeTCAMConfig {
	return &FeTCAMConfig{
		Rows:       1024,
		Cols:       256,
		CellConfig: DefaultFeFETCAMConfig(),
		MatchType:  "exact",
	}
}

// FeTCAM represents a ferroelectric TCAM array
type FeTCAM struct {
	Config *FeTCAMConfig

	// Cell array
	Cells [][]*FeFETCAMCell

	// Matchlines
	MatchResults []bool
	MatchCount   int

	// Performance metrics
	TotalSearchEnergy float64 // fJ
	SearchLatency     float64 // ns
	ThroughputMOPS    float64 // million searches per second
	EnergyVsCMOS      float64 // improvement factor
}

// NewFeTCAM creates a new ferroelectric TCAM array
func NewFeTCAM(config *FeTCAMConfig) *FeTCAM {
	if config == nil {
		config = DefaultFeTCAMConfig()
	}

	tcam := &FeTCAM{
		Config:       config,
		Cells:        make([][]*FeFETCAMCell, config.Rows),
		MatchResults: make([]bool, config.Rows),
	}

	// Initialize cells
	for i := 0; i < config.Rows; i++ {
		tcam.Cells[i] = make([]*FeFETCAMCell, config.Cols)
		for j := 0; j < config.Cols; j++ {
			tcam.Cells[i][j] = NewFeFETCAMCell(config.CellConfig)
		}
	}

	tcam.calculateMetrics()

	return tcam
}

// calculateMetrics computes array performance
func (t *FeTCAM) calculateMetrics() {
	config := t.Config
	cellConfig := config.CellConfig

	// Search energy = sum of all cell energies
	cellEnergy := t.Cells[0][0].SearchEnergy
	t.TotalSearchEnergy = cellEnergy * float64(config.Rows*config.Cols)

	// Latency = single cell delay (parallel search)
	t.SearchLatency = t.Cells[0][0].MatchDelay

	// Throughput
	t.ThroughputMOPS = 1000.0 / t.SearchLatency // MHz → MOPS

	// Energy comparison vs 16T CMOS TCAM
	cmosEnergyPerCell := 5.0 // fJ (baseline)
	switch cellConfig.CellType {
	case CAMCell2FeFET1T:
		t.EnergyVsCMOS = 3.03 // 3.03× reduction
	case CAMCell2FeFET2T:
		t.EnergyVsCMOS = 8.08 // 8.08× reduction
	case CAMCellHFNN:
		t.EnergyVsCMOS = 226.92 // 226× reduction
	default:
		t.EnergyVsCMOS = cmosEnergyPerCell / cellEnergy
	}
}

// ProgramEntry writes a word to a specific row
func (t *FeTCAM) ProgramEntry(row int, data []int) {
	if row >= t.Config.Rows {
		return
	}

	for j := 0; j < t.Config.Cols && j < len(data); j++ {
		t.Cells[row][j].Write(data[j])
	}
}

// Search performs parallel associative search
func (t *FeTCAM) Search(query []int) []int {
	matchingRows := make([]int, 0)

	// Reset match results
	t.MatchCount = 0
	for i := range t.MatchResults {
		t.MatchResults[i] = true
	}

	// Parallel search across all rows
	for i := 0; i < t.Config.Rows; i++ {
		match := true
		for j := 0; j < t.Config.Cols && j < len(query); j++ {
			if !t.Cells[i][j].Search(query[j]) {
				match = false
				break
			}
		}
		t.MatchResults[i] = match
		if match {
			matchingRows = append(matchingRows, i)
			t.MatchCount++
		}
	}

	return matchingRows
}

// HammingSearch returns rows within Hamming distance threshold
func (t *FeTCAM) HammingSearch(query []int, threshold int) []int {
	matchingRows := make([]int, 0)

	for i := 0; i < t.Config.Rows; i++ {
		distance := 0
		for j := 0; j < t.Config.Cols && j < len(query); j++ {
			if t.Cells[i][j].StoredValue != 2 && t.Cells[i][j].StoredValue != query[j] {
				distance++
			}
		}
		if distance <= threshold {
			matchingRows = append(matchingRows, i)
		}
	}

	return matchingRows
}

// =============================================================================
// ANALOG CONTENT-ADDRESSABLE MEMORY (ACAM)
// =============================================================================

// ACAMConfig configures analog CAM for nearest neighbor search
type ACAMConfig struct {
	Rows          int     // number of stored vectors
	Dimensions    int     // vector dimensionality
	BitsPerWeight int     // precision (4-8)
	DistanceType  string  // "L1", "L2", "cosine"
	SearchTopK    int     // return top-K matches
}

// DefaultACAMConfig returns default analog CAM configuration
func DefaultACAMConfig() *ACAMConfig {
	return &ACAMConfig{
		Rows:          1024,
		Dimensions:    512,
		BitsPerWeight: 6,
		DistanceType:  "L2",
		SearchTopK:    10,
	}
}

// ACAM represents an analog CAM for neural network search
type ACAM struct {
	Config *ACAMConfig

	// Stored vectors (normalized)
	Vectors [][]float64

	// Distance computation crossbar
	DistanceCrossbar [][]float64

	// Performance metrics
	SearchEnergy    float64 // fJ per search
	SearchLatency   float64 // ns
	EnergyVsGPU     float64 // improvement factor (60×)
	LatencyVsGPU    float64 // improvement factor (2700×)
}

// NewACAM creates a new analog CAM
func NewACAM(config *ACAMConfig) *ACAM {
	if config == nil {
		config = DefaultACAMConfig()
	}

	acam := &ACAM{
		Config:           config,
		Vectors:          make([][]float64, config.Rows),
		DistanceCrossbar: make([][]float64, config.Rows),
	}

	// Initialize storage
	for i := 0; i < config.Rows; i++ {
		acam.Vectors[i] = make([]float64, config.Dimensions)
		acam.DistanceCrossbar[i] = make([]float64, config.Dimensions)
	}

	acam.calculateMetrics()

	return acam
}

// calculateMetrics computes ACAM performance
func (a *ACAM) calculateMetrics() {
	config := a.Config

	// Energy and latency based on FACAM research
	// Reference: 60× energy, 2700× latency reduction vs GPU

	// Base energy for crossbar search
	baseEnergy := 10.0 * float64(config.Rows*config.Dimensions) // fJ

	// FeFET analog advantage
	a.SearchEnergy = baseEnergy / 60.0 // 60× reduction

	// Latency (single cycle parallel)
	a.SearchLatency = 10.0 // ns (vs 27000 ns on GPU)

	a.EnergyVsGPU = 60.0
	a.LatencyVsGPU = 2700.0
}

// ProgramVector stores a vector at specified index
func (a *ACAM) ProgramVector(index int, vector []float64) {
	if index >= a.Config.Rows {
		return
	}

	// Normalize and store
	norm := 0.0
	for _, v := range vector {
		norm += v * v
	}
	norm = math.Sqrt(norm)

	for j := 0; j < a.Config.Dimensions && j < len(vector); j++ {
		a.Vectors[index][j] = vector[j] / norm
		a.DistanceCrossbar[index][j] = a.Vectors[index][j]
	}
}

// NearestNeighborSearch finds closest vectors to query
func (a *ACAM) NearestNeighborSearch(query []float64) []int {
	config := a.Config

	// Normalize query
	norm := 0.0
	for _, v := range query {
		norm += v * v
	}
	norm = math.Sqrt(norm)

	normalizedQuery := make([]float64, len(query))
	for i, v := range query {
		normalizedQuery[i] = v / norm
	}

	// Compute distances (parallel in analog crossbar)
	distances := make([]float64, config.Rows)

	for i := 0; i < config.Rows; i++ {
		switch config.DistanceType {
		case "cosine":
			// Cosine similarity (higher = closer)
			dot := 0.0
			for j := 0; j < config.Dimensions && j < len(normalizedQuery); j++ {
				dot += a.Vectors[i][j] * normalizedQuery[j]
			}
			distances[i] = -dot // negate for sorting
		case "L2":
			// Euclidean distance
			sumSq := 0.0
			for j := 0; j < config.Dimensions && j < len(normalizedQuery); j++ {
				diff := a.Vectors[i][j] - normalizedQuery[j]
				sumSq += diff * diff
			}
			distances[i] = sumSq
		case "L1":
			// Manhattan distance
			sum := 0.0
			for j := 0; j < config.Dimensions && j < len(normalizedQuery); j++ {
				sum += math.Abs(a.Vectors[i][j] - normalizedQuery[j])
			}
			distances[i] = sum
		}
	}

	// Find top-K (simple selection)
	topK := make([]int, 0, config.SearchTopK)
	used := make([]bool, config.Rows)

	for k := 0; k < config.SearchTopK && k < config.Rows; k++ {
		minIdx := -1
		minDist := math.Inf(1)
		for i := 0; i < config.Rows; i++ {
			if !used[i] && distances[i] < minDist {
				minDist = distances[i]
				minIdx = i
			}
		}
		if minIdx >= 0 {
			topK = append(topK, minIdx)
			used[minIdx] = true
		}
	}

	return topK
}

// =============================================================================
// TRANSFORMER ATTENTION MAPPING TO CROSSBAR
// =============================================================================

// AttentionMappingConfig configures attention-to-crossbar mapping
type AttentionMappingConfig struct {
	NumHeads      int     // attention heads
	HeadDim       int     // dimension per head
	SeqLength     int     // sequence length
	HiddenDim     int     // model hidden dimension
	CrossbarSize  int     // physical crossbar size
	Precision     int     // weight precision (bits)
	UseCascade    bool    // use cascaded crossbars (ReCAT)
	UseKVCache    bool    // enable KV cache optimization
}

// DefaultAttentionMappingConfig returns default configuration
func DefaultAttentionMappingConfig() *AttentionMappingConfig {
	return &AttentionMappingConfig{
		NumHeads:     8,
		HeadDim:      64,
		SeqLength:    512,
		HiddenDim:    512,
		CrossbarSize: 128,
		Precision:    6,
		UseCascade:   true,
		UseKVCache:   true,
	}
}

// AttentionCrossbar represents attention mechanism mapped to crossbar
type AttentionCrossbar struct {
	Config *AttentionMappingConfig

	// Weight crossbars (Q, K, V projections)
	QueryWeights [][]float64
	KeyWeights   [][]float64
	ValueWeights [][]float64

	// Attention score crossbar
	AttentionScores [][]float64

	// KV Cache (ferroelectric storage)
	KeyCache   [][]float64
	ValueCache [][]float64
	CacheValid []bool

	// Tiling information
	NumTilesQ int
	NumTilesK int
	NumTilesV int

	// Performance metrics
	ProjectionLatency  float64 // ns
	AttentionLatency   float64 // ns
	TotalLatency       float64 // ns
	ProjectionEnergy   float64 // fJ
	AttentionEnergy    float64 // fJ
	TotalEnergy        float64 // fJ
	EnergyVsGPU        float64 // improvement factor
	SpeedVsGPU         float64 // improvement factor
}

// NewAttentionCrossbar creates attention mechanism on crossbar
func NewAttentionCrossbar(config *AttentionMappingConfig) *AttentionCrossbar {
	if config == nil {
		config = DefaultAttentionMappingConfig()
	}

	dim := config.HiddenDim

	ac := &AttentionCrossbar{
		Config:          config,
		QueryWeights:    make([][]float64, dim),
		KeyWeights:      make([][]float64, dim),
		ValueWeights:    make([][]float64, dim),
		AttentionScores: make([][]float64, config.SeqLength),
		KeyCache:        make([][]float64, config.SeqLength),
		ValueCache:      make([][]float64, config.SeqLength),
		CacheValid:      make([]bool, config.SeqLength),
	}

	// Initialize weight matrices
	for i := 0; i < dim; i++ {
		ac.QueryWeights[i] = make([]float64, dim)
		ac.KeyWeights[i] = make([]float64, dim)
		ac.ValueWeights[i] = make([]float64, dim)
	}

	// Initialize attention scores
	for i := 0; i < config.SeqLength; i++ {
		ac.AttentionScores[i] = make([]float64, config.SeqLength)
		ac.KeyCache[i] = make([]float64, dim)
		ac.ValueCache[i] = make([]float64, dim)
	}

	ac.calculateTiling()
	ac.calculateMetrics()

	return ac
}

// calculateTiling determines crossbar tiling for weights
func (ac *AttentionCrossbar) calculateTiling() {
	config := ac.Config
	crossbarSize := config.CrossbarSize

	// Tiling for weight matrices
	ac.NumTilesQ = (config.HiddenDim + crossbarSize - 1) / crossbarSize
	ac.NumTilesK = ac.NumTilesQ
	ac.NumTilesV = ac.NumTilesQ
}

// calculateMetrics computes performance
func (ac *AttentionCrossbar) calculateMetrics() {
	config := ac.Config

	// Projection latency (single crossbar MVM)
	mvmLatency := 10.0 // ns per crossbar
	ac.ProjectionLatency = mvmLatency * float64(ac.NumTilesQ)

	// Attention computation
	if config.UseCascade {
		// ReCAT: cascaded crossbars avoid ADC between Q*K and softmax*V
		ac.AttentionLatency = mvmLatency * 2 // two cascaded operations
	} else {
		// Standard: separate Q*K, softmax, V multiply
		seqTiles := (config.SeqLength + config.CrossbarSize - 1) / config.CrossbarSize
		ac.AttentionLatency = mvmLatency * float64(seqTiles) * 3
	}

	ac.TotalLatency = ac.ProjectionLatency*3 + ac.AttentionLatency

	// Energy calculation
	cellEnergy := 1.0 // fJ per MAC
	numMACs := float64(config.HiddenDim * config.HiddenDim * 3) // Q, K, V projections
	attentionMACs := float64(config.SeqLength * config.SeqLength * config.HeadDim)

	ac.ProjectionEnergy = cellEnergy * numMACs / 100.0 // analog efficiency
	ac.AttentionEnergy = cellEnergy * attentionMACs / 100.0
	ac.TotalEnergy = ac.ProjectionEnergy + ac.AttentionEnergy

	// AIMC attention paper: 70,000× energy reduction, 100× speed
	// More conservative estimate based on cascaded architecture
	if config.UseCascade {
		ac.EnergyVsGPU = 1000.0  // 1000× reduction (cascaded)
		ac.SpeedVsGPU = 85.0    // 85× speedup (X-Former result)
	} else {
		ac.EnergyVsGPU = 100.0
		ac.SpeedVsGPU = 10.0
	}
}

// ProgramWeights loads Q, K, V projection weights
func (ac *AttentionCrossbar) ProgramWeights(wq, wk, wv [][]float64) {
	dim := ac.Config.HiddenDim

	for i := 0; i < dim && i < len(wq); i++ {
		for j := 0; j < dim && j < len(wq[i]); j++ {
			ac.QueryWeights[i][j] = wq[i][j]
			ac.KeyWeights[i][j] = wk[i][j]
			ac.ValueWeights[i][j] = wv[i][j]
		}
	}
}

// ComputeProjection performs Q, K, V projections
func (ac *AttentionCrossbar) ComputeProjection(input []float64) (q, k, v []float64) {
	dim := ac.Config.HiddenDim

	q = make([]float64, dim)
	k = make([]float64, dim)
	v = make([]float64, dim)

	// MVM operations (parallel in crossbar)
	for i := 0; i < dim; i++ {
		for j := 0; j < dim && j < len(input); j++ {
			q[i] += ac.QueryWeights[i][j] * input[j]
			k[i] += ac.KeyWeights[i][j] * input[j]
			v[i] += ac.ValueWeights[i][j] * input[j]
		}
	}

	return q, k, v
}

// UpdateKVCache stores key and value for position
func (ac *AttentionCrossbar) UpdateKVCache(position int, key, value []float64) {
	if position >= ac.Config.SeqLength {
		return
	}

	copy(ac.KeyCache[position], key)
	copy(ac.ValueCache[position], value)
	ac.CacheValid[position] = true
}

// ComputeAttention performs attention mechanism
func (ac *AttentionCrossbar) ComputeAttention(query []float64, currentPos int) []float64 {
	config := ac.Config
	dim := config.HiddenDim
	scale := 1.0 / math.Sqrt(float64(config.HeadDim))

	// Compute attention scores: Q * K^T
	scores := make([]float64, currentPos+1)

	for i := 0; i <= currentPos; i++ {
		if !ac.CacheValid[i] {
			continue
		}
		dot := 0.0
		for j := 0; j < dim && j < len(query); j++ {
			dot += query[j] * ac.KeyCache[i][j]
		}
		scores[i] = dot * scale
	}

	// Softmax
	maxScore := scores[0]
	for _, s := range scores {
		if s > maxScore {
			maxScore = s
		}
	}

	expSum := 0.0
	expScores := make([]float64, len(scores))
	for i, s := range scores {
		expScores[i] = math.Exp(s - maxScore)
		expSum += expScores[i]
	}

	for i := range expScores {
		expScores[i] /= expSum
	}

	// Output: softmax(scores) * V
	output := make([]float64, dim)
	for i := 0; i <= currentPos; i++ {
		if !ac.CacheValid[i] {
			continue
		}
		for j := 0; j < dim; j++ {
			output[j] += expScores[i] * ac.ValueCache[i][j]
		}
	}

	return output
}

// =============================================================================
// RECOMBINED ATTENTION ENGINE (ReCAT-inspired)
// =============================================================================

// ReCATConfig configures cascaded crossbar attention
type ReCATConfig struct {
	CrossbarRows   int
	CrossbarCols   int
	NumCascades    int  // number of cascaded crossbar pairs
	UseTIA         bool // use transimpedance amplifiers
	AvoidADC       bool // skip ADC between cascades
}

// DefaultReCATConfig returns default ReCAT configuration
func DefaultReCATConfig() *ReCATConfig {
	return &ReCATConfig{
		CrossbarRows: 128,
		CrossbarCols: 128,
		NumCascades:  2,
		UseTIA:       true,
		AvoidADC:     true,
	}
}

// ReCATEngine implements cascaded crossbar attention
type ReCATEngine struct {
	Config *ReCATConfig

	// Cascaded crossbar pairs
	Crossbars [][][]float64 // [cascade][row][col]

	// Performance metrics
	ADCReduction    float64 // factor of ADC reduction
	LatencyReductionFloat64 // factor vs non-cascaded
	EnergyReduction float64 // factor vs non-cascaded
}

// NewReCATEngine creates a new ReCAT engine
func NewReCATEngine(config *ReCATConfig) *ReCATEngine {
	if config == nil {
		config = DefaultReCATConfig()
	}

	engine := &ReCATEngine{
		Config:    config,
		Crossbars: make([][][]float64, config.NumCascades*2),
	}

	// Initialize crossbar arrays
	for c := 0; c < config.NumCascades*2; c++ {
		engine.Crossbars[c] = make([][]float64, config.CrossbarRows)
		for i := 0; i < config.CrossbarRows; i++ {
			engine.Crossbars[c][i] = make([]float64, config.CrossbarCols)
		}
	}

	engine.calculateMetrics()

	return engine
}

// calculateMetrics computes ReCAT performance
func (r *ReCATEngine) calculateMetrics() {
	config := r.Config

	// ADC reduction from cascading
	if config.AvoidADC {
		// Each cascade avoids one set of ADC/DAC conversions
		r.ADCReduction = float64(config.NumCascades)
	} else {
		r.ADCReduction = 1.0
	}

	// Latency reduction from overlapped operations
	r.LatencyReductionFloat64 = 2.3 // 2.3× from ReCAT paper

	// Energy reduction (ADC dominates ~60% of energy)
	r.EnergyReduction = 1.7 // 1.7× from ReCAT paper
}

// CascadedMVM performs cascaded matrix-vector multiply
func (r *ReCATEngine) CascadedMVM(input []float64, cascadeIdx int) []float64 {
	config := r.Config

	if cascadeIdx*2+1 >= len(r.Crossbars) {
		return nil
	}

	// First crossbar MVM
	intermediate := make([]float64, config.CrossbarRows)
	crossbar1 := r.Crossbars[cascadeIdx*2]

	for i := 0; i < config.CrossbarRows; i++ {
		for j := 0; j < config.CrossbarCols && j < len(input); j++ {
			intermediate[i] += crossbar1[i][j] * input[j]
		}
	}

	// Cascade to second crossbar (via TIA, no ADC)
	output := make([]float64, config.CrossbarRows)
	crossbar2 := r.Crossbars[cascadeIdx*2+1]

	for i := 0; i < config.CrossbarRows; i++ {
		for j := 0; j < config.CrossbarCols && j < len(intermediate); j++ {
			output[i] += crossbar2[i][j] * intermediate[j]
		}
	}

	return output
}

// =============================================================================
// KV CACHE MEMORY FOR LLM INFERENCE
// =============================================================================

// KVCacheConfig configures ferroelectric KV cache
type KVCacheConfig struct {
	SeqLength    int     // maximum sequence length
	NumLayers    int     // transformer layers
	NumHeads     int     // attention heads
	HeadDim      int     // dimension per head
	UseGainCell  bool    // use gain cell devices (AIMC)
	Use3D        bool    // use 3D integration
}

// DefaultKVCacheConfig returns default KV cache configuration
func DefaultKVCacheConfig() *KVCacheConfig {
	return &KVCacheConfig{
		SeqLength:   2048,
		NumLayers:   12,
		NumHeads:    8,
		HeadDim:     64,
		UseGainCell: true,
		Use3D:       true,
	}
}

// FerroKVCache represents ferroelectric KV cache for LLMs
type FerroKVCache struct {
	Config *KVCacheConfig

	// Cache storage [layer][head][position][dim]
	KeyCache   [][][][]float64
	ValueCache [][][][]float64
	Valid      [][][]bool // [layer][head][position]

	// Cache metrics
	TotalSizeBytes   int64   // total cache size
	FootprintMM2     float64 // mm² with 3D integration
	ReadLatency      float64 // ns
	WriteLatency     float64 // ns
	ReadEnergy       float64 // fJ per read
	WriteEnergy      float64 // fJ per write
	EnergyVsHBM      float64 // improvement vs HBM
}

// NewFerroKVCache creates a new ferroelectric KV cache
func NewFerroKVCache(config *KVCacheConfig) *FerroKVCache {
	if config == nil {
		config = DefaultKVCacheConfig()
	}

	cache := &FerroKVCache{
		Config:     config,
		KeyCache:   make([][][][]float64, config.NumLayers),
		ValueCache: make([][][][]float64, config.NumLayers),
		Valid:      make([][][]bool, config.NumLayers),
	}

	// Initialize cache arrays
	for l := 0; l < config.NumLayers; l++ {
		cache.KeyCache[l] = make([][][]float64, config.NumHeads)
		cache.ValueCache[l] = make([][][]float64, config.NumHeads)
		cache.Valid[l] = make([][]bool, config.NumHeads)

		for h := 0; h < config.NumHeads; h++ {
			cache.KeyCache[l][h] = make([][]float64, config.SeqLength)
			cache.ValueCache[l][h] = make([][]float64, config.SeqLength)
			cache.Valid[l][h] = make([]bool, config.SeqLength)

			for p := 0; p < config.SeqLength; p++ {
				cache.KeyCache[l][h][p] = make([]float64, config.HeadDim)
				cache.ValueCache[l][h][p] = make([]float64, config.HeadDim)
			}
		}
	}

	cache.calculateMetrics()

	return cache
}

// calculateMetrics computes KV cache performance
func (c *FerroKVCache) calculateMetrics() {
	config := c.Config

	// Total size: 2 (K+V) × layers × heads × seq × dim × 4 bytes
	bytesPerElement := 4 // float32
	c.TotalSizeBytes = int64(2 * config.NumLayers * config.NumHeads *
		config.SeqLength * config.HeadDim * bytesPerElement)

	// Footprint with 3D integration
	// From AIMC paper: 3.1 × 10^-3 mm² per layer
	if config.Use3D {
		c.FootprintMM2 = 3.1e-3 * float64(config.NumLayers)
	} else {
		c.FootprintMM2 = 0.1 * float64(config.NumLayers) // 2D baseline
	}

	// Latency (gain cell vs SRAM)
	if config.UseGainCell {
		c.ReadLatency = 5.0   // ns
		c.WriteLatency = 10.0 // ns
	} else {
		c.ReadLatency = 1.0   // ns (SRAM)
		c.WriteLatency = 1.0  // ns
	}

	// Energy (ferroelectric advantage)
	// AIMC paper: 70,000× reduction vs GPU
	hbmReadEnergy := 1000.0 // fJ per access
	c.ReadEnergy = hbmReadEnergy / 100.0  // 100× reduction
	c.WriteEnergy = hbmReadEnergy / 50.0  // 50× reduction
	c.EnergyVsHBM = 100.0
}

// Store writes KV pair to cache
func (c *FerroKVCache) Store(layer, head, position int, key, value []float64) {
	if layer >= c.Config.NumLayers || head >= c.Config.NumHeads ||
		position >= c.Config.SeqLength {
		return
	}

	copy(c.KeyCache[layer][head][position], key)
	copy(c.ValueCache[layer][head][position], value)
	c.Valid[layer][head][position] = true
}

// Load retrieves KV pair from cache
func (c *FerroKVCache) Load(layer, head, position int) ([]float64, []float64, bool) {
	if layer >= c.Config.NumLayers || head >= c.Config.NumHeads ||
		position >= c.Config.SeqLength {
		return nil, nil, false
	}

	if !c.Valid[layer][head][position] {
		return nil, nil, false
	}

	return c.KeyCache[layer][head][position], c.ValueCache[layer][head][position], true
}

// GetAllKeys returns all valid keys for a layer/head up to position
func (c *FerroKVCache) GetAllKeys(layer, head, maxPosition int) [][]float64 {
	keys := make([][]float64, 0)

	for p := 0; p <= maxPosition && p < c.Config.SeqLength; p++ {
		if c.Valid[layer][head][p] {
			keys = append(keys, c.KeyCache[layer][head][p])
		}
	}

	return keys
}

// GetAllValues returns all valid values for a layer/head up to position
func (c *FerroKVCache) GetAllValues(layer, head, maxPosition int) [][]float64 {
	values := make([][]float64, 0)

	for p := 0; p <= maxPosition && p < c.Config.SeqLength; p++ {
		if c.Valid[layer][head][p] {
			values = append(values, c.ValueCache[layer][head][p])
		}
	}

	return values
}

// =============================================================================
// IRONLATTICE CAM-ATTENTION SYSTEM
// =============================================================================

// IronLatticeCAMAttentionConfig configures integrated system
type IronLatticeCAMAttentionConfig struct {
	// TCAM configuration
	TCAMEnabled  bool
	TCAMConfig   *FeTCAMConfig

	// ACAM configuration (for NN search)
	ACAMEnabled  bool
	ACAMConfig   *ACAMConfig

	// Attention configuration
	AttentionEnabled bool
	AttentionConfig  *AttentionMappingConfig

	// KV Cache configuration
	KVCacheEnabled bool
	KVCacheConfig  *KVCacheConfig
}

// DefaultIronLatticeCAMAttentionConfig returns default configuration
func DefaultIronLatticeCAMAttentionConfig() *IronLatticeCAMAttentionConfig {
	return &IronLatticeCAMAttentionConfig{
		TCAMEnabled:      true,
		TCAMConfig:       DefaultFeTCAMConfig(),
		ACAMEnabled:      true,
		ACAMConfig:       DefaultACAMConfig(),
		AttentionEnabled: true,
		AttentionConfig:  DefaultAttentionMappingConfig(),
		KVCacheEnabled:   true,
		KVCacheConfig:    DefaultKVCacheConfig(),
	}
}

// IronLatticeCAMAttention represents integrated CAM and attention system
type IronLatticeCAMAttention struct {
	Config *IronLatticeCAMAttentionConfig

	// Components
	TCAM       *FeTCAM
	ACAM       *ACAM
	Attention  *AttentionCrossbar
	KVCache    *FerroKVCache
	ReCAT      *ReCATEngine

	// System metrics
	TotalEnergy      float64 // fJ
	TotalLatency     float64 // ns
	MemoryFootprint  float64 // mm²
	EnergyVsBaseline float64 // improvement factor
}

// NewIronLatticeCAMAttention creates integrated system
func NewIronLatticeCAMAttention(config *IronLatticeCAMAttentionConfig) *IronLatticeCAMAttention {
	if config == nil {
		config = DefaultIronLatticeCAMAttentionConfig()
	}

	sys := &IronLatticeCAMAttention{
		Config: config,
	}

	// Initialize components
	if config.TCAMEnabled {
		sys.TCAM = NewFeTCAM(config.TCAMConfig)
	}

	if config.ACAMEnabled {
		sys.ACAM = NewACAM(config.ACAMConfig)
	}

	if config.AttentionEnabled {
		sys.Attention = NewAttentionCrossbar(config.AttentionConfig)
		sys.ReCAT = NewReCATEngine(DefaultReCATConfig())
	}

	if config.KVCacheEnabled {
		sys.KVCache = NewFerroKVCache(config.KVCacheConfig)
	}

	sys.calculateSystemMetrics()

	return sys
}

// calculateSystemMetrics computes overall performance
func (s *IronLatticeCAMAttention) calculateSystemMetrics() {
	s.TotalEnergy = 0
	s.TotalLatency = 0
	s.MemoryFootprint = 0

	if s.TCAM != nil {
		s.TotalEnergy += s.TCAM.TotalSearchEnergy
		s.TotalLatency += s.TCAM.SearchLatency
	}

	if s.ACAM != nil {
		s.TotalEnergy += s.ACAM.SearchEnergy
		s.TotalLatency += s.ACAM.SearchLatency
	}

	if s.Attention != nil {
		s.TotalEnergy += s.Attention.TotalEnergy
		s.TotalLatency += s.Attention.TotalLatency
	}

	if s.KVCache != nil {
		s.MemoryFootprint = s.KVCache.FootprintMM2
	}

	// Overall improvement estimate
	// Combining TCAM (226×), ACAM (60×), Attention (85×)
	s.EnergyVsBaseline = 100.0 // conservative estimate
}

// RunDatabaseSearch performs TCAM-based database search
func (s *IronLatticeCAMAttention) RunDatabaseSearch(queries [][]int) [][]int {
	if s.TCAM == nil {
		return nil
	}

	results := make([][]int, len(queries))

	for i, query := range queries {
		results[i] = s.TCAM.Search(query)
	}

	return results
}

// RunNearestNeighborSearch performs ACAM-based NN search
func (s *IronLatticeCAMAttention) RunNearestNeighborSearch(queries [][]float64) [][]int {
	if s.ACAM == nil {
		return nil
	}

	results := make([][]int, len(queries))

	for i, query := range queries {
		results[i] = s.ACAM.NearestNeighborSearch(query)
	}

	return results
}

// RunTransformerInference performs attention-based inference
func (s *IronLatticeCAMAttention) RunTransformerInference(input [][]float64) [][]float64 {
	if s.Attention == nil {
		return nil
	}

	outputs := make([][]float64, len(input))

	for pos, token := range input {
		// Compute projections
		q, k, v := s.Attention.ComputeProjection(token)

		// Update KV cache
		s.Attention.UpdateKVCache(pos, k, v)

		// Compute attention
		outputs[pos] = s.Attention.ComputeAttention(q, pos)
	}

	return outputs
}

// GetPerformanceSummary returns human-readable performance summary
func (s *IronLatticeCAMAttention) GetPerformanceSummary() map[string]interface{} {
	summary := make(map[string]interface{})

	summary["total_energy_fj"] = s.TotalEnergy
	summary["total_latency_ns"] = s.TotalLatency
	summary["memory_footprint_mm2"] = s.MemoryFootprint
	summary["energy_vs_baseline"] = s.EnergyVsBaseline

	if s.TCAM != nil {
		summary["tcam_rows"] = s.Config.TCAMConfig.Rows
		summary["tcam_cols"] = s.Config.TCAMConfig.Cols
		summary["tcam_energy_vs_cmos"] = s.TCAM.EnergyVsCMOS
		summary["tcam_throughput_mops"] = s.TCAM.ThroughputMOPS
	}

	if s.ACAM != nil {
		summary["acam_vectors"] = s.Config.ACAMConfig.Rows
		summary["acam_dimensions"] = s.Config.ACAMConfig.Dimensions
		summary["acam_energy_vs_gpu"] = s.ACAM.EnergyVsGPU
		summary["acam_latency_vs_gpu"] = s.ACAM.LatencyVsGPU
	}

	if s.Attention != nil {
		summary["attention_heads"] = s.Config.AttentionConfig.NumHeads
		summary["attention_seq_length"] = s.Config.AttentionConfig.SeqLength
		summary["attention_energy_vs_gpu"] = s.Attention.EnergyVsGPU
		summary["attention_speed_vs_gpu"] = s.Attention.SpeedVsGPU
	}

	if s.KVCache != nil {
		summary["kvcache_size_bytes"] = s.KVCache.TotalSizeBytes
		summary["kvcache_footprint_mm2"] = s.KVCache.FootprintMM2
		summary["kvcache_energy_vs_hbm"] = s.KVCache.EnergyVsHBM
	}

	return summary
}

// =============================================================================
// BENCHMARKING UTILITIES
// =============================================================================

// CAMBenchmark stores CAM benchmark results
type CAMBenchmark struct {
	CellType       string
	Year           int
	TransistorCount int // transistors per cell
	EnergyVsCMOS   float64 // improvement factor
	SearchLatency  float64 // ns
	AreaMM2        float64 // mm² per Mb
}

// GetCAMBenchmarks returns literature benchmark data
func GetCAMBenchmarks() []CAMBenchmark {
	return []CAMBenchmark{
		{
			CellType:        "16T CMOS TCAM",
			Year:            2020,
			TransistorCount: 16,
			EnergyVsCMOS:    1.0,
			SearchLatency:   1.0,
			AreaMM2:         0.5,
		},
		{
			CellType:        "2FeFET-1T TCAM",
			Year:            2022,
			TransistorCount: 3,
			EnergyVsCMOS:    3.03,
			SearchLatency:   0.8,
			AreaMM2:         0.15,
		},
		{
			CellType:        "2FeFET-2T TCAM",
			Year:            2022,
			TransistorCount: 4,
			EnergyVsCMOS:    8.08,
			SearchLatency:   1.0,
			AreaMM2:         0.18,
		},
		{
			CellType:        "HFNN TCAM",
			Year:            2022,
			TransistorCount: 2,
			EnergyVsCMOS:    226.92,
			SearchLatency:   0.5,
			AreaMM2:         0.1,
		},
		{
			CellType:        "CECAM (HZO)",
			Year:            2024,
			TransistorCount: 2,
			EnergyVsCMOS:    2.86, // 65% reduction
			SearchLatency:   0.4,
			AreaMM2:         0.08,
		},
		{
			CellType:        "TAP-CAM",
			Year:            2024,
			TransistorCount: 4, // 2FeFET-2R
			EnergyVsCMOS:    5.0,
			SearchLatency:   0.6,
			AreaMM2:         0.12,
		},
		{
			CellType:        "FeSQUID TCAM (cryo)",
			Year:            2025,
			TransistorCount: 1,
			EnergyVsCMOS:    1000000, // 1.36 aJ vs ~1 fJ
			SearchLatency:   0.1,
			AreaMM2:         0.05,
		},
	}
}

// AttentionBenchmark stores attention accelerator benchmarks
type AttentionBenchmark struct {
	Name           string
	Year           int
	EnergyVsGPU    float64 // improvement factor
	SpeedVsGPU     float64 // improvement factor
	Technology     string
	TargetModel    string
}

// GetAttentionBenchmarks returns attention accelerator benchmarks
func GetAttentionBenchmarks() []AttentionBenchmark {
	return []AttentionBenchmark{
		{
			Name:        "X-Former",
			Year:        2023,
			EnergyVsGPU: 7.5,
			SpeedVsGPU:  85.0,
			Technology:  "NVM + CMOS",
			TargetModel: "BERT",
		},
		{
			Name:        "ReCAT",
			Year:        2024,
			EnergyVsGPU: 1.7,
			SpeedVsGPU:  2.3,
			Technology:  "ReRAM cascaded",
			TargetModel: "Transformer",
		},
		{
			Name:        "HARDSEA",
			Year:        2024,
			EnergyVsGPU: 3.0,
			SpeedVsGPU:  5.0,
			Technology:  "ReRAM + SRAM hybrid",
			TargetModel: "Sparse attention",
		},
		{
			Name:        "3D FeFET-CIM",
			Year:        2024,
			EnergyVsGPU: 3.1,
			SpeedVsGPU:  2.6,
			Technology:  "22nm FeFET 3D",
			TargetModel: "BERT, GPT-2",
		},
		{
			Name:        "AIMC Attention",
			Year:        2025,
			EnergyVsGPU: 70000.0,
			SpeedVsGPU:  100.0,
			Technology:  "Gain cell AIMC",
			TargetModel: "1.5B LLM",
		},
		{
			Name:        "FACAM MANN",
			Year:        2024,
			EnergyVsGPU: 60.0,
			SpeedVsGPU:  2700.0,
			Technology:  "FeFET ACAM",
			TargetModel: "Memory-augmented NN",
		},
	}
}

// RunComprehensiveBenchmark runs full CAM and attention benchmark
func RunComprehensiveBenchmarkCAMAttention() map[string]interface{} {
	results := make(map[string]interface{})

	// Literature benchmarks
	results["cam_benchmarks"] = GetCAMBenchmarks()
	results["attention_benchmarks"] = GetAttentionBenchmarks()

	// Full system simulation
	sysConfig := DefaultIronLatticeCAMAttentionConfig()
	sys := NewIronLatticeCAMAttention(sysConfig)

	// Generate test data
	testQueries := make([][]int, 100)
	for i := range testQueries {
		testQueries[i] = make([]int, sysConfig.TCAMConfig.Cols)
		for j := range testQueries[i] {
			testQueries[i][j] = rand.Intn(2)
		}
	}

	// Run TCAM search
	_ = sys.RunDatabaseSearch(testQueries)

	// Generate NN test data
	testVectors := make([][]float64, 10)
	for i := range testVectors {
		testVectors[i] = make([]float64, sysConfig.ACAMConfig.Dimensions)
		for j := range testVectors[i] {
			testVectors[i][j] = rand.Float64()*2 - 1
		}
	}

	// Program ACAM
	for i := 0; i < sysConfig.ACAMConfig.Rows; i++ {
		vec := make([]float64, sysConfig.ACAMConfig.Dimensions)
		for j := range vec {
			vec[j] = rand.Float64()*2 - 1
		}
		sys.ACAM.ProgramVector(i, vec)
	}

	// Run NN search
	_ = sys.RunNearestNeighborSearch(testVectors)

	// Generate transformer input
	testInput := make([][]float64, 32) // 32 tokens
	for i := range testInput {
		testInput[i] = make([]float64, sysConfig.AttentionConfig.HiddenDim)
		for j := range testInput[i] {
			testInput[i][j] = rand.Float64()*2 - 1
		}
	}

	// Program attention weights
	dim := sysConfig.AttentionConfig.HiddenDim
	wq := make([][]float64, dim)
	wk := make([][]float64, dim)
	wv := make([][]float64, dim)
	for i := 0; i < dim; i++ {
		wq[i] = make([]float64, dim)
		wk[i] = make([]float64, dim)
		wv[i] = make([]float64, dim)
		for j := 0; j < dim; j++ {
			wq[i][j] = rand.NormFloat64() * 0.1
			wk[i][j] = rand.NormFloat64() * 0.1
			wv[i][j] = rand.NormFloat64() * 0.1
		}
	}
	sys.Attention.ProgramWeights(wq, wk, wv)

	// Run transformer inference
	_ = sys.RunTransformerInference(testInput)

	results["system_performance"] = sys.GetPerformanceSummary()

	return results
}
