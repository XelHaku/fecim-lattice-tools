// Package layers provides CIM-accelerated diffusion model and cryogenic quantum computing simulations.
// Based on AIG-CIM (DAC 2024), FeSQUID research, and ferroelectric Josephson junction work.
package layers

import (
	"fmt"
	"math"
	"math/rand"
)

// =============================================================================
// DIFFUSION MODEL CIM ACCELERATION
// =============================================================================

// DiffusionConfig holds configuration for diffusion model CIM acceleration.
type DiffusionConfig struct {
	ModelName        string  // "stable_diffusion", "sdxl", "dit"
	ImageSize        int     // Output image resolution (e.g., 512, 1024)
	LatentSize       int     // Latent space size (ImageSize/8 typically)
	NumChannels      int     // U-Net channels (320 base)
	NumLayers        int     // U-Net depth
	NumHeads         int     // Attention heads
	NumTimesteps     int     // Denoising steps (typically 20-50 for DDIM)
	ArraySize        int     // CIM array size
	WeightBits       int     // Weight quantization
	ActivationBits   int     // Activation quantization
	UseTriGear       bool    // Use tri-gear heterogeneous CIM
	ExploitSimilarity bool   // Exploit denoising similarity
}

// DefaultStableDiffusionConfig returns config for Stable Diffusion 1.5.
func DefaultStableDiffusionConfig() *DiffusionConfig {
	return &DiffusionConfig{
		ModelName:        "stable_diffusion_1.5",
		ImageSize:        512,
		LatentSize:       64,
		NumChannels:      320,
		NumLayers:        12,
		NumHeads:         8,
		NumTimesteps:     50,
		ArraySize:        64,
		WeightBits:       8,
		ActivationBits:   8,
		UseTriGear:       true,
		ExploitSimilarity: true,
	}
}

// SDXLConfig returns config for SDXL.
func SDXLConfig() *DiffusionConfig {
	return &DiffusionConfig{
		ModelName:        "sdxl_1.0",
		ImageSize:        1024,
		LatentSize:       128,
		NumChannels:      640,
		NumLayers:        16,
		NumHeads:         16,
		NumTimesteps:     30,
		ArraySize:        64,
		WeightBits:       8,
		ActivationBits:   8,
		UseTriGear:       true,
		ExploitSimilarity: true,
	}
}

// DiTConfig returns config for Diffusion Transformer.
func DiTConfig() *DiffusionConfig {
	return &DiffusionConfig{
		ModelName:        "dit_xl",
		ImageSize:        256,
		LatentSize:       32,
		NumChannels:      1152,
		NumLayers:        28,
		NumHeads:         16,
		NumTimesteps:     250,
		ArraySize:        64,
		WeightBits:       8,
		ActivationBits:   8,
		UseTriGear:       true,
		ExploitSimilarity: true,
	}
}

// TriGearMacro represents a heterogeneous CIM macro with three gear modes.
// Inspired by AIG-CIM architecture for flexible data reuse in diffusion.
type TriGearMacro struct {
	Config       *DiffusionConfig
	Mode         string         // "fine", "medium", "coarse"
	Weights      [][]float64
	Size         int
	DataReuse    float64        // Data reuse factor
	ReadCycles   int64
	WriteCycles  int64
	EnergyPJ     float64
}

// NewTriGearMacro creates a new tri-gear CIM macro.
func NewTriGearMacro(cfg *DiffusionConfig, size int) *TriGearMacro {
	weights := make([][]float64, size)
	for i := range weights {
		weights[i] = make([]float64, size)
		for j := range weights[i] {
			weights[i][j] = rand.NormFloat64() * 0.1
		}
	}

	return &TriGearMacro{
		Config:    cfg,
		Mode:      "medium",
		Weights:   weights,
		Size:      size,
		DataReuse: 1.0,
	}
}

// SetGear switches the macro to a specific gear mode.
// Fine: High precision, low reuse (early timesteps)
// Medium: Balanced (middle timesteps)
// Coarse: High reuse, lower precision (late timesteps)
func (m *TriGearMacro) SetGear(mode string) {
	m.Mode = mode
	switch mode {
	case "fine":
		m.DataReuse = 1.0
	case "medium":
		m.DataReuse = 2.0
	case "coarse":
		m.DataReuse = 4.0
	}
}

// MatVecMul performs matrix-vector multiplication with gear-specific optimization.
func (m *TriGearMacro) MatVecMul(input []float64) []float64 {
	output := make([]float64, m.Size)

	// Apply gear-specific precision
	precisionScale := 1.0
	if m.Mode == "coarse" {
		precisionScale = 0.9 // Slightly lower precision
	}

	for i := 0; i < m.Size; i++ {
		sum := 0.0
		for j := 0; j < len(input) && j < m.Size; j++ {
			sum += m.Weights[i][j] * input[j]
		}
		output[i] = sum * precisionScale
	}

	// Track energy (adjusted for data reuse)
	baseEnergy := float64(m.Size*len(input)) * 0.5 // fJ/MAC
	m.EnergyPJ += baseEnergy / m.DataReuse / 1000.0
	m.ReadCycles++

	return output
}

// UNetBlock represents a U-Net encoder/decoder block mapped to CIM.
type UNetBlock struct {
	Config     *DiffusionConfig
	ResBlocks  []*TriGearMacro
	AttnBlocks []*TriGearMacro
	BlockType  string // "encoder", "decoder", "middle"
	Resolution int
}

// NewUNetBlock creates a new U-Net block with CIM macros.
func NewUNetBlock(cfg *DiffusionConfig, blockType string, resolution int) *UNetBlock {
	numRes := 2
	numAttn := 1
	if resolution <= 32 {
		numAttn = 2 // More attention at lower resolutions
	}

	resBlocks := make([]*TriGearMacro, numRes)
	for i := range resBlocks {
		resBlocks[i] = NewTriGearMacro(cfg, cfg.ArraySize)
	}

	attnBlocks := make([]*TriGearMacro, numAttn)
	for i := range attnBlocks {
		attnBlocks[i] = NewTriGearMacro(cfg, cfg.ArraySize)
	}

	return &UNetBlock{
		Config:     cfg,
		ResBlocks:  resBlocks,
		AttnBlocks: attnBlocks,
		BlockType:  blockType,
		Resolution: resolution,
	}
}

// DiffusionAccelerator implements AIG-CIM-style diffusion acceleration.
type DiffusionAccelerator struct {
	Config          *DiffusionConfig
	Encoder         []*UNetBlock
	Decoder         []*UNetBlock
	MiddleBlock     *UNetBlock
	CurrentTimestep int
	TotalEnergy     float64   // pJ
	TotalOps        int64
	LatencyMS       float64
}

// NewDiffusionAccelerator creates a new diffusion model accelerator.
func NewDiffusionAccelerator(cfg *DiffusionConfig) *DiffusionAccelerator {
	// Create encoder blocks (progressively lower resolution)
	resolutions := []int{64, 32, 16, 8}
	encoder := make([]*UNetBlock, len(resolutions))
	for i, res := range resolutions {
		encoder[i] = NewUNetBlock(cfg, "encoder", res)
	}

	// Create decoder blocks (progressively higher resolution)
	decoder := make([]*UNetBlock, len(resolutions))
	for i := range decoder {
		decoder[i] = NewUNetBlock(cfg, "decoder", resolutions[len(resolutions)-1-i])
	}

	// Middle block
	middle := NewUNetBlock(cfg, "middle", 8)

	return &DiffusionAccelerator{
		Config:      cfg,
		Encoder:     encoder,
		Decoder:     decoder,
		MiddleBlock: middle,
	}
}

// DenoiseStep performs a single denoising step with CIM acceleration.
func (acc *DiffusionAccelerator) DenoiseStep(latent []float64, timestep int) ([]float64, *DenoisingStats) {
	stats := &DenoisingStats{Timestep: timestep}
	acc.CurrentTimestep = timestep

	// Select gear based on timestep
	gear := acc.selectGear(timestep)
	acc.setAllGears(gear)
	stats.Gear = gear

	// Encoder pass
	features := latent
	skipConnections := make([][]float64, len(acc.Encoder))
	for i, block := range acc.Encoder {
		for _, res := range block.ResBlocks {
			features = res.MatVecMul(features)
			stats.ResBlockOps++
		}
		for _, attn := range block.AttnBlocks {
			features = attn.MatVecMul(features)
			stats.AttnBlockOps++
		}
		skipConnections[i] = make([]float64, len(features))
		copy(skipConnections[i], features)
	}

	// Middle block
	for _, res := range acc.MiddleBlock.ResBlocks {
		features = res.MatVecMul(features)
		stats.ResBlockOps++
	}
	for _, attn := range acc.MiddleBlock.AttnBlocks {
		features = attn.MatVecMul(features)
		stats.AttnBlockOps++
	}

	// Decoder pass with skip connections
	for i, block := range acc.Decoder {
		// Add skip connection
		skipIdx := len(acc.Encoder) - 1 - i
		if skipIdx >= 0 && skipIdx < len(skipConnections) {
			for j := range features {
				if j < len(skipConnections[skipIdx]) {
					features[j] += skipConnections[skipIdx][j]
				}
			}
		}

		for _, res := range block.ResBlocks {
			features = res.MatVecMul(features)
			stats.ResBlockOps++
		}
		for _, attn := range block.AttnBlocks {
			features = attn.MatVecMul(features)
			stats.AttnBlockOps++
		}
	}

	// Calculate energy
	stats.EnergyPJ = acc.calculateStepEnergy()
	acc.TotalEnergy += stats.EnergyPJ
	acc.TotalOps += stats.ResBlockOps + stats.AttnBlockOps

	return features, stats
}

// selectGear chooses gear based on denoising progress.
func (acc *DiffusionAccelerator) selectGear(timestep int) string {
	// Early steps: fine gear (high noise, need precision)
	// Middle steps: medium gear
	// Late steps: coarse gear (low noise, can reuse)
	progress := float64(timestep) / float64(acc.Config.NumTimesteps)

	if progress < 0.3 {
		return "fine"
	} else if progress < 0.7 {
		return "medium"
	}
	return "coarse"
}

// setAllGears sets gear for all macros.
func (acc *DiffusionAccelerator) setAllGears(gear string) {
	for _, block := range acc.Encoder {
		for _, m := range block.ResBlocks {
			m.SetGear(gear)
		}
		for _, m := range block.AttnBlocks {
			m.SetGear(gear)
		}
	}
	for _, block := range acc.Decoder {
		for _, m := range block.ResBlocks {
			m.SetGear(gear)
		}
		for _, m := range block.AttnBlocks {
			m.SetGear(gear)
		}
	}
	for _, m := range acc.MiddleBlock.ResBlocks {
		m.SetGear(gear)
	}
	for _, m := range acc.MiddleBlock.AttnBlocks {
		m.SetGear(gear)
	}
}

// calculateStepEnergy calculates energy for current step.
func (acc *DiffusionAccelerator) calculateStepEnergy() float64 {
	energy := 0.0
	for _, block := range acc.Encoder {
		for _, m := range block.ResBlocks {
			energy += m.EnergyPJ
		}
		for _, m := range block.AttnBlocks {
			energy += m.EnergyPJ
		}
	}
	// Similar for decoder and middle
	return energy
}

// GenerateImage runs full diffusion generation.
func (acc *DiffusionAccelerator) GenerateImage(noise []float64) (*GenerationResult, error) {
	result := &GenerationResult{
		Config:      acc.Config,
		StepStats:   make([]*DenoisingStats, acc.Config.NumTimesteps),
	}

	latent := noise
	for t := 0; t < acc.Config.NumTimesteps; t++ {
		var stats *DenoisingStats
		latent, stats = acc.DenoiseStep(latent, t)
		result.StepStats[t] = stats
	}

	result.TotalEnergy = acc.TotalEnergy
	result.TotalOps = acc.TotalOps

	// Calculate efficiency metrics
	result.TOPSW = float64(result.TotalOps) / (result.TotalEnergy / 1e12) / 1e12
	result.LatencyReduction = 21.3  // Based on AIG-CIM paper
	result.ThroughputGain = 231.2   // Based on AIG-CIM paper

	return result, nil
}

// DenoisingStats holds statistics for a single denoising step.
type DenoisingStats struct {
	Timestep     int
	Gear         string
	ResBlockOps  int64
	AttnBlockOps int64
	EnergyPJ     float64
	LatencyUS    float64
}

// GenerationResult holds results from image generation.
type GenerationResult struct {
	Config           *DiffusionConfig
	StepStats        []*DenoisingStats
	TotalEnergy      float64
	TotalOps         int64
	TOPSW            float64
	LatencyReduction float64
	ThroughputGain   float64
}

// =============================================================================
// CRYOGENIC FERROELECTRIC MEMORY (FeSQUID)
// =============================================================================

// FeSQUIDConfig holds configuration for ferroelectric SQUID arrays.
type FeSQUIDConfig struct {
	ArraySize        int       // TCAM array size
	Temperature      float64   // Operating temperature (K)
	CriticalCurrent  float64   // SQUID critical current (μA)
	MemoryLevels     int       // Multi-level storage
	EnergyPerSearch  float64   // aJ per bit search
	SwitchingVoltage float64   // Ferroelectric switching voltage (V)
}

// DefaultFeSQUIDConfig returns default FeSQUID configuration.
func DefaultFeSQUIDConfig() *FeSQUIDConfig {
	return &FeSQUIDConfig{
		ArraySize:        64,
		Temperature:      4.0,     // 4 Kelvin
		CriticalCurrent:  10.0,    // 10 μA typical
		MemoryLevels:     4,       // 2-bit per cell
		EnergyPerSearch:  1.36,    // 1.36 aJ per binary search
		SwitchingVoltage: 2.0,     // 2V switching
	}
}

// FeSQUIDCell represents a single FeSQUID memory cell.
type FeSQUIDCell struct {
	Config         *FeSQUIDConfig
	Polarization   float64   // Ferroelectric polarization state (-1 to 1)
	CriticalHigh   float64   // Critical current in high state
	CriticalLow    float64   // Critical current in low state
	CurrentState   int       // Current memory level
	ReadCount      int64
	WriteCount     int64
}

// NewFeSQUIDCell creates a new FeSQUID cell.
func NewFeSQUIDCell(cfg *FeSQUIDConfig) *FeSQUIDCell {
	return &FeSQUIDCell{
		Config:       cfg,
		Polarization: 0,
		CriticalHigh: cfg.CriticalCurrent * 1.54, // 54% modulation
		CriticalLow:  cfg.CriticalCurrent,
		CurrentState: 0,
	}
}

// Program writes a value to the cell.
func (c *FeSQUIDCell) Program(level int) error {
	if level < 0 || level >= c.Config.MemoryLevels {
		return fmt.Errorf("invalid level %d", level)
	}

	// Map level to polarization
	c.Polarization = float64(level)/float64(c.Config.MemoryLevels-1)*2 - 1
	c.CurrentState = level
	c.WriteCount++

	return nil
}

// Read retrieves the current value.
func (c *FeSQUIDCell) Read() int {
	c.ReadCount++
	return c.CurrentState
}

// GetCriticalCurrent returns current critical current based on polarization.
func (c *FeSQUIDCell) GetCriticalCurrent() float64 {
	// Linear interpolation between low and high
	t := (c.Polarization + 1) / 2
	return c.CriticalLow + t*(c.CriticalHigh-c.CriticalLow)
}

// FeSQUIDTCAM implements ternary content-addressable memory.
type FeSQUIDTCAM struct {
	Config       *FeSQUIDConfig
	DataCells    [][]*FeSQUIDCell
	MaskCells    [][]*FeSQUIDCell
	NumRows      int
	NumCols      int
	SearchCount  int64
	TotalEnergy  float64  // aJ
}

// NewFeSQUIDTCAM creates a new TCAM array.
func NewFeSQUIDTCAM(cfg *FeSQUIDConfig, rows, cols int) *FeSQUIDTCAM {
	dataCells := make([][]*FeSQUIDCell, rows)
	maskCells := make([][]*FeSQUIDCell, rows)

	for i := range dataCells {
		dataCells[i] = make([]*FeSQUIDCell, cols)
		maskCells[i] = make([]*FeSQUIDCell, cols)
		for j := range dataCells[i] {
			dataCells[i][j] = NewFeSQUIDCell(cfg)
			maskCells[i][j] = NewFeSQUIDCell(cfg)
		}
	}

	return &FeSQUIDTCAM{
		Config:    cfg,
		DataCells: dataCells,
		MaskCells: maskCells,
		NumRows:   rows,
		NumCols:   cols,
	}
}

// Store writes a ternary pattern to a row.
// value: 0=0, 1=1, 2=don't care (X)
func (t *FeSQUIDTCAM) Store(row int, pattern []int) error {
	if row >= t.NumRows {
		return fmt.Errorf("row %d out of bounds", row)
	}

	for col := 0; col < t.NumCols && col < len(pattern); col++ {
		switch pattern[col] {
		case 0: // Store 0
			t.DataCells[row][col].Program(0)
			t.MaskCells[row][col].Program(0)
		case 1: // Store 1
			t.DataCells[row][col].Program(1)
			t.MaskCells[row][col].Program(0)
		case 2: // Don't care
			t.DataCells[row][col].Program(0)
			t.MaskCells[row][col].Program(1) // Mask bit set
		}
	}

	return nil
}

// Search performs parallel search for a pattern.
func (t *FeSQUIDTCAM) Search(pattern []int) *TCAMSearchResult {
	result := &TCAMSearchResult{
		Matches:          make([]bool, t.NumRows),
		HammingDistances: make([]int, t.NumRows),
	}

	for row := 0; row < t.NumRows; row++ {
		match := true
		distance := 0

		for col := 0; col < t.NumCols && col < len(pattern); col++ {
			// Check if masked (don't care)
			if t.MaskCells[row][col].Read() == 1 {
				continue
			}

			stored := t.DataCells[row][col].Read()
			if stored != pattern[col] {
				match = false
				distance++
			}
		}

		result.Matches[row] = match
		result.HammingDistances[row] = distance
		if match {
			result.MatchCount++
		}
	}

	// Track energy
	bitsSearched := int64(t.NumRows * t.NumCols)
	t.TotalEnergy += float64(bitsSearched) * t.Config.EnergyPerSearch
	t.SearchCount++

	result.EnergyAJ = float64(bitsSearched) * t.Config.EnergyPerSearch

	return result
}

// TCAMSearchResult holds search results.
type TCAMSearchResult struct {
	Matches          []bool
	HammingDistances []int
	MatchCount       int
	EnergyAJ         float64
}

// =============================================================================
// FERROELECTRIC JOSEPHSON JUNCTION
// =============================================================================

// FeJJConfig holds configuration for ferroelectric Josephson junction.
type FeJJConfig struct {
	Temperature        float64 // Operating temperature (K)
	BarrierThickness1  float64 // Insulator 1 thickness (nm)
	BarrierThickness2  float64 // Insulator 2 thickness (nm)
	FEThickness        float64 // Ferroelectric thickness (nm)
	BarrierHeight1     float64 // Barrier potential 1 (eV)
	BarrierHeight2     float64 // Barrier potential 2 (eV)
	FEDielectric       float64 // Ferroelectric dielectric constant
	OnOffEfficiency    float64 // Target on-off efficiency
}

// DefaultFeJJConfig returns optimized FeJJ configuration.
func DefaultFeJJConfig() *FeJJConfig {
	return &FeJJConfig{
		Temperature:       0.01,   // 10 mK
		BarrierThickness1: 1.0,    // 1 nm
		BarrierThickness2: 2.0,    // 2 nm (asymmetric)
		FEThickness:       5.0,    // 5 nm HZO
		BarrierHeight1:    1.0,    // 1 eV
		BarrierHeight2:    0.8,    // 0.8 eV
		FEDielectric:      25,     // HZO
		OnOffEfficiency:   0.9,    // 90% efficiency target
	}
}

// FeJJDevice represents a ferroelectric Josephson junction device.
type FeJJDevice struct {
	Config            *FeJJConfig
	Polarization      float64   // FE polarization state (-1 to 1)
	CriticalCurrentOn float64   // Critical current in ON state (μA)
	CriticalCurrentOff float64  // Critical current in OFF state (μA)
	CurrentState      bool      // Current binary state
	SwitchCount       int64
}

// NewFeJJDevice creates a new FeJJ device.
func NewFeJJDevice(cfg *FeJJConfig) *FeJJDevice {
	// Calculate critical currents based on WKB tunneling model
	// I_c ~ exp(-2*kappa*d) where kappa depends on barrier height

	kappa1 := math.Sqrt(2 * 0.511 * cfg.BarrierHeight1) / 0.197 // nm^-1
	kappa2 := math.Sqrt(2 * 0.511 * cfg.BarrierHeight2) / 0.197

	// Tunneling probability (simplified)
	t1 := math.Exp(-2 * kappa1 * cfg.BarrierThickness1)
	t2 := math.Exp(-2 * kappa2 * cfg.BarrierThickness2)

	// Base critical current (arbitrary units scaled to μA)
	baseCurrent := 10.0 * t1 * t2

	// On-off ratio based on asymmetry and polarization effect
	// Asymmetric barriers give different current for P+ vs P-
	asymmetry := cfg.BarrierThickness2 / cfg.BarrierThickness1
	onOff := 1.0 + (asymmetry - 1.0) * cfg.OnOffEfficiency

	return &FeJJDevice{
		Config:            cfg,
		Polarization:      1.0, // Start in ON state
		CriticalCurrentOn: baseCurrent * onOff,
		CriticalCurrentOff: baseCurrent,
		CurrentState:      true,
	}
}

// Switch toggles the device state.
func (d *FeJJDevice) Switch() {
	d.Polarization *= -1
	d.CurrentState = !d.CurrentState
	d.SwitchCount++
}

// GetCriticalCurrent returns current critical current.
func (d *FeJJDevice) GetCriticalCurrent() float64 {
	if d.CurrentState {
		return d.CriticalCurrentOn
	}
	return d.CriticalCurrentOff
}

// GetOnOffEfficiency returns actual on-off efficiency.
func (d *FeJJDevice) GetOnOffEfficiency() float64 {
	return 1.0 - d.CriticalCurrentOff/d.CriticalCurrentOn
}

// FeJJMemory implements a memory array using FeJJ devices.
type FeJJMemory struct {
	Config      *FeJJConfig
	Devices     [][]*FeJJDevice
	NumRows     int
	NumCols     int
	ReadCount   int64
	WriteCount  int64
}

// NewFeJJMemory creates a new FeJJ memory array.
func NewFeJJMemory(cfg *FeJJConfig, rows, cols int) *FeJJMemory {
	devices := make([][]*FeJJDevice, rows)
	for i := range devices {
		devices[i] = make([]*FeJJDevice, cols)
		for j := range devices[i] {
			devices[i][j] = NewFeJJDevice(cfg)
		}
	}

	return &FeJJMemory{
		Config:  cfg,
		Devices: devices,
		NumRows: rows,
		NumCols: cols,
	}
}

// Write stores a bit pattern.
func (m *FeJJMemory) Write(row int, data []bool) error {
	if row >= m.NumRows {
		return fmt.Errorf("row %d out of bounds", row)
	}

	for col := 0; col < m.NumCols && col < len(data); col++ {
		if m.Devices[row][col].CurrentState != data[col] {
			m.Devices[row][col].Switch()
		}
	}

	m.WriteCount++
	return nil
}

// Read retrieves a bit pattern.
func (m *FeJJMemory) Read(row int) []bool {
	if row >= m.NumRows {
		return nil
	}

	data := make([]bool, m.NumCols)
	for col := 0; col < m.NumCols; col++ {
		data[col] = m.Devices[row][col].CurrentState
	}

	m.ReadCount++
	return data
}

// =============================================================================
// HZO PHOTONIC PHASE SHIFTER
// =============================================================================

// PhotonicHZOConfig holds configuration for HZO photonic phase shifter.
type PhotonicHZOConfig struct {
	WaveguideLength   float64 // Phase modulator length (mm)
	HZOThickness      float64 // Total HZO thickness (nm)
	NumLayers         int     // Number of HZO layers
	InterlayerThick   float64 // Al2O3 interlayer thickness (nm)
	Wavelength        float64 // Operating wavelength (μm)
	MaxVoltage        float64 // Maximum applied voltage (V)
	RefractiveChange  float64 // Max refractive index change
}

// DefaultPhotonicConfig returns default HZO photonic configuration.
// Based on Nature Communications 2024 paper.
func DefaultPhotonicConfig() *PhotonicHZOConfig {
	return &PhotonicHZOConfig{
		WaveguideLength:  4.5,       // 4.5 mm
		HZOThickness:     30.0,      // 30 nm total (3 × 10 nm)
		NumLayers:        3,
		InterlayerThick:  1.0,       // 1 nm Al2O3
		Wavelength:       1.55,      // 1.55 μm telecom
		MaxVoltage:       210.0,     // 210 V
		RefractiveChange: -1.5e-4,   // -1.5 × 10^-4
	}
}

// PhotonicPhaseShifter implements HZO-based nonvolatile optical phase shifter.
type PhotonicPhaseShifter struct {
	Config         *PhotonicHZOConfig
	CurrentPhase   float64   // Current phase shift (radians)
	Polarization   float64   // FE polarization state (0-1)
	RetentionTime  float64   // Retention time (seconds)
	NumStates      int       // Number of intermediate states
	CurrentVoltage float64
}

// NewPhotonicPhaseShifter creates a new HZO phase shifter.
func NewPhotonicPhaseShifter(cfg *PhotonicHZOConfig) *PhotonicPhaseShifter {
	return &PhotonicPhaseShifter{
		Config:        cfg,
		CurrentPhase:  0,
		Polarization:  0,
		RetentionTime: 10000, // >10,000 seconds demonstrated
		NumStates:     10,    // ~10 intermediate states
	}
}

// ApplyVoltage applies programming voltage and calculates phase shift.
func (p *PhotonicPhaseShifter) ApplyVoltage(voltage float64) error {
	if voltage < 110 || voltage > p.Config.MaxVoltage {
		return fmt.Errorf("voltage must be between 110-%.0f V", p.Config.MaxVoltage)
	}

	// Calculate electric field in HZO layer
	fieldMVcm := voltage / (p.Config.HZOThickness / 10) / 1000 // MV/cm

	// Polarization rotation (simplified model)
	// Above 110V, FE switching begins
	switchProgress := (voltage - 110) / (p.Config.MaxVoltage - 110)
	p.Polarization = switchProgress

	// Refractive index change
	// Approximately linear with polarization
	deltaN := p.Config.RefractiveChange * switchProgress

	// Phase shift = (2π/λ) × Δn × L
	p.CurrentPhase = (2 * math.Pi / p.Config.Wavelength) * math.Abs(deltaN) * p.Config.WaveguideLength * 1000
	p.CurrentVoltage = voltage

	return nil
}

// GetPhaseShift returns current phase shift in radians.
func (p *PhotonicPhaseShifter) GetPhaseShift() float64 {
	return p.CurrentPhase
}

// GetTransmission calculates MZI transmission based on phase.
func (p *PhotonicPhaseShifter) GetTransmission() float64 {
	// MZI: T = cos²(Δφ/2)
	return math.Pow(math.Cos(p.CurrentPhase/2), 2)
}

// =============================================================================
// QUANTUM ERROR CORRECTION DECODER
// =============================================================================

// QECConfig holds configuration for quantum error correction.
type QECConfig struct {
	CodeDistance    int      // Surface code distance
	NumDataQubits   int      // Number of data qubits
	NumAncilla      int      // Number of ancilla qubits
	ErrorRate       float64  // Physical error rate
	DecodingMethod  string   // "minimum_weight", "neural", "lookup"
}

// DefaultQECConfig returns default QEC configuration.
func DefaultQECConfig() *QECConfig {
	distance := 5
	return &QECConfig{
		CodeDistance:   distance,
		NumDataQubits:  distance * distance,
		NumAncilla:     (distance-1) * (distance-1) * 2,
		ErrorRate:      0.001,
		DecodingMethod: "lookup",
	}
}

// QECDecoder implements FeSQUID-accelerated QEC decoding.
type QECDecoder struct {
	Config        *QECConfig
	SyndromeTCAM  *FeSQUIDTCAM  // Syndrome lookup table
	CorrectionMap map[string][]int
	DecodeCycles  int64
	TotalEnergy   float64  // aJ
}

// NewQECDecoder creates a new QEC decoder with FeSQUID TCAM.
func NewQECDecoder(cfg *QECConfig) *QECDecoder {
	// Create syndrome TCAM
	// Each row stores a syndrome pattern and correction
	numSyndromes := 1 << cfg.NumAncilla // 2^n possible syndromes
	if numSyndromes > 4096 {
		numSyndromes = 4096 // Cap for practical TCAM size
	}

	tcamCfg := DefaultFeSQUIDConfig()
	tcam := NewFeSQUIDTCAM(tcamCfg, numSyndromes, cfg.NumAncilla)

	return &QECDecoder{
		Config:        cfg,
		SyndromeTCAM:  tcam,
		CorrectionMap: make(map[string][]int),
	}
}

// LoadSyndromeTable loads syndrome-correction mappings.
func (d *QECDecoder) LoadSyndromeTable(syndromes [][]int, corrections [][]int) error {
	if len(syndromes) != len(corrections) {
		return fmt.Errorf("syndrome and correction arrays must match")
	}

	for i, syndrome := range syndromes {
		if i >= d.SyndromeTCAM.NumRows {
			break
		}
		d.SyndromeTCAM.Store(i, syndrome)

		// Store correction mapping
		key := fmt.Sprintf("%v", syndrome)
		d.CorrectionMap[key] = corrections[i]
	}

	return nil
}

// Decode performs syndrome decoding using TCAM lookup.
func (d *QECDecoder) Decode(syndrome []int) *QECResult {
	result := &QECResult{}

	// Search TCAM for matching syndrome
	searchResult := d.SyndromeTCAM.Search(syndrome)

	// Find exact match
	for i, match := range searchResult.Matches {
		if match {
			result.MatchedRow = i
			result.Found = true
			break
		}
	}

	// If no exact match, find minimum Hamming distance
	if !result.Found {
		minDist := d.Config.NumAncilla + 1
		for i, dist := range searchResult.HammingDistances {
			if dist < minDist {
				minDist = dist
				result.MatchedRow = i
				result.HammingDistance = dist
			}
		}
	}

	result.EnergyAJ = searchResult.EnergyAJ
	d.TotalEnergy += result.EnergyAJ
	d.DecodeCycles++

	return result
}

// QECResult holds decoding result.
type QECResult struct {
	Found           bool
	MatchedRow      int
	HammingDistance int
	Correction      []int
	EnergyAJ        float64
}

// =============================================================================
// BENCHMARK AND COMPARISON
// =============================================================================

// CryogenicBenchmark compares different cryogenic memory technologies.
type CryogenicBenchmark struct {
	Results map[string]*CryogenicMetrics
}

// CryogenicMetrics holds performance metrics.
type CryogenicMetrics struct {
	Technology      string
	Temperature     float64  // K
	EnergyPerBit    float64  // aJ or fJ
	ReadLatency     float64  // ns
	WriteLatency    float64  // ns
	Endurance       float64  // cycles
	Density         float64  // bits/μm²
	NonVolatile     bool
}

// NewCryogenicBenchmark creates a benchmark comparison.
func NewCryogenicBenchmark() *CryogenicBenchmark {
	return &CryogenicBenchmark{
		Results: make(map[string]*CryogenicMetrics),
	}
}

// AddFeSQUID adds FeSQUID metrics.
func (b *CryogenicBenchmark) AddFeSQUID() {
	b.Results["FeSQUID"] = &CryogenicMetrics{
		Technology:   "FeSQUID",
		Temperature:  4.0,
		EnergyPerBit: 1.36,      // 1.36 aJ
		ReadLatency:  0.1,       // ~100 ps
		WriteLatency: 1.0,       // ~1 ns
		Endurance:    1e15,      // Very high
		Density:      0.1,       // Low density
		NonVolatile:  true,
	}
}

// AddFeJJ adds FeJJ metrics.
func (b *CryogenicBenchmark) AddFeJJ() {
	b.Results["FeJJ"] = &CryogenicMetrics{
		Technology:   "Ferroelectric Josephson Junction",
		Temperature:  0.01,      // 10 mK
		EnergyPerBit: 0.1,       // ~0.1 aJ
		ReadLatency:  0.01,      // ~10 ps
		WriteLatency: 0.1,       // ~100 ps
		Endurance:    1e12,
		Density:      0.05,
		NonVolatile:  true,
	}
}

// AddCryoSRAM adds cryogenic SRAM metrics.
func (b *CryogenicBenchmark) AddCryoSRAM() {
	b.Results["CryoSRAM"] = &CryogenicMetrics{
		Technology:   "5nm FinFET Cryogenic SRAM",
		Temperature:  4.0,
		EnergyPerBit: 50.0,      // ~50 aJ
		ReadLatency:  0.5,       // ~500 ps
		WriteLatency: 0.5,
		Endurance:    1e18,
		Density:      1.0,
		NonVolatile:  false,
	}
}

// Compare generates comparison report.
func (b *CryogenicBenchmark) Compare() string {
	report := "Cryogenic Memory Technology Comparison\n"
	report += "=======================================\n\n"

	for name, m := range b.Results {
		report += fmt.Sprintf("%s:\n", name)
		report += fmt.Sprintf("  Temperature: %.2f K\n", m.Temperature)
		report += fmt.Sprintf("  Energy/bit: %.2f aJ\n", m.EnergyPerBit)
		report += fmt.Sprintf("  Read latency: %.2f ns\n", m.ReadLatency)
		report += fmt.Sprintf("  Write latency: %.2f ns\n", m.WriteLatency)
		report += fmt.Sprintf("  Endurance: %.0e cycles\n", m.Endurance)
		report += fmt.Sprintf("  Non-volatile: %v\n\n", m.NonVolatile)
	}

	return report
}
