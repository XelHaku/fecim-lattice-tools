// Package peripherals provides GPU-accelerated peripheral circuit simulation.
package peripherals

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"unsafe"

	"fecim-lattice-tools/shared/compute"
)

// GPUPeripherals provides GPU-accelerated peripheral simulation using Vulkan compute shaders.
// This enables batch processing of DAC, ADC, and TIA conversions at high throughput.
type GPUPeripherals struct {
	ctx         *compute.VulkanContext
	dacPipeline *compute.ComputePipeline
	adcPipeline *compute.ComputePipeline
	tiaPipeline *compute.ComputePipeline
}

// DACParams matches dac.comp uniform buffer layout (std140).
// All fields must align to 4-byte boundaries as per std140 rules.
type DACParams struct {
	Bits    int32   // DAC resolution (e.g., 5 bits = 32 levels)
	VrefP   float32 // Positive reference voltage
	VrefN   float32 // Negative reference voltage
	INLMax  float32 // Maximum INL in LSBs
	DNLMax  float32 // Maximum DNL in LSBs
	Size    int32   // Array size
	Seed    float32 // Random seed for variation
	Padding float32 // Padding to align to std140 rules
}

// ADCParams matches adc.comp uniform buffer layout (std140).
type ADCParams struct {
	Bits     int32   // ADC resolution
	VrefP    float32 // Positive reference voltage
	VrefN    float32 // Negative reference voltage
	INLMax   float32 // Maximum INL in LSBs
	DNLMax   float32 // Maximum DNL in LSBs
	NoiseRMS float32 // Input-referred noise (V)
	Size     int32   // Array size
	Seed     float32 // Random seed
}

// TIAParams matches tia.comp uniform buffer layout (std140).
type TIAParams struct {
	Gain       float32 // Transimpedance gain (Ohms)
	Bandwidth  float32 // -3dB bandwidth (Hz)
	InputNoise float32 // Input-referred current noise (A/sqrt(Hz))
	Vmax       float32 // Output saturation voltage
	Size       int32   // Array size
	Seed       float32 // Random seed
	Padding1   float32 // Padding for std140 alignment
	Padding2   float32 // Padding for std140 alignment
}

// NewGPUPeripherals creates and initializes GPU peripheral pipelines.
// Returns a GPUPeripherals instance even if GPU is unavailable (check with IsAvailable).
func NewGPUPeripherals() (*GPUPeripherals, error) {
	// Create Vulkan context (won't error if unavailable)
	ctx, err := compute.NewVulkanContext()
	if err != nil {
		return nil, fmt.Errorf("failed to create Vulkan context: %w", err)
	}

	g := &GPUPeripherals{
		ctx: ctx,
	}

	// If Vulkan not available, return early with available=false
	if !ctx.IsAvailable() {
		return g, nil
	}

	// Find repo root by looking for go.mod
	repoRoot, err := findRepoRoot()
	if err != nil {
		g.Destroy()
		return nil, fmt.Errorf("failed to find repo root: %w", err)
	}

	// Create DAC pipeline
	dacConfig := compute.PipelineConfig{
		ShaderPath: filepath.Join(repoRoot, "module4-circuits/shaders/dac.comp.spv"),
		Bindings: []compute.BindingInfo{
			{Binding: 0, Type: compute.BindingTypeUniform, Size: uint64(unsafe.Sizeof(DACParams{}))},
			{Binding: 1, Type: compute.BindingTypeStorage, Size: 0}, // Input: int32[]
			{Binding: 2, Type: compute.BindingTypeStorage, Size: 0}, // Output: float32[]
		},
	}
	dacPipeline, err := compute.NewComputePipeline(ctx, dacConfig)
	if err != nil {
		g.Destroy()
		return nil, fmt.Errorf("failed to create DAC pipeline: %w", err)
	}
	g.dacPipeline = dacPipeline

	// Create ADC pipeline
	adcConfig := compute.PipelineConfig{
		ShaderPath: filepath.Join(repoRoot, "module4-circuits/shaders/adc.comp.spv"),
		Bindings: []compute.BindingInfo{
			{Binding: 0, Type: compute.BindingTypeUniform, Size: uint64(unsafe.Sizeof(ADCParams{}))},
			{Binding: 1, Type: compute.BindingTypeStorage, Size: 0}, // Input: float32[]
			{Binding: 2, Type: compute.BindingTypeStorage, Size: 0}, // Output: int32[]
			{Binding: 3, Type: compute.BindingTypeStorage, Size: 0}, // Output: float32[] quantized
		},
	}
	adcPipeline, err := compute.NewComputePipeline(ctx, adcConfig)
	if err != nil {
		g.Destroy()
		return nil, fmt.Errorf("failed to create ADC pipeline: %w", err)
	}
	g.adcPipeline = adcPipeline

	// Create TIA pipeline
	tiaConfig := compute.PipelineConfig{
		ShaderPath: filepath.Join(repoRoot, "module4-circuits/shaders/tia.comp.spv"),
		Bindings: []compute.BindingInfo{
			{Binding: 0, Type: compute.BindingTypeUniform, Size: uint64(unsafe.Sizeof(TIAParams{}))},
			{Binding: 1, Type: compute.BindingTypeStorage, Size: 0}, // Input: float32[]
			{Binding: 2, Type: compute.BindingTypeStorage, Size: 0}, // Output: float32[]
		},
	}
	tiaPipeline, err := compute.NewComputePipeline(ctx, tiaConfig)
	if err != nil {
		g.Destroy()
		return nil, fmt.Errorf("failed to create TIA pipeline: %w", err)
	}
	g.tiaPipeline = tiaPipeline

	return g, nil
}

// IsAvailable returns true if GPU compute is available for peripheral operations.
func (g *GPUPeripherals) IsAvailable() bool {
	return g.ctx != nil && g.ctx.IsAvailable()
}

// BatchDAC performs GPU-accelerated DAC conversion on a batch of digital codes.
// Returns analog voltages with INL/DNL nonlinearity applied.
func (g *GPUPeripherals) BatchDAC(codes []int32, params DACParams) ([]float32, error) {
	if !g.IsAvailable() {
		return nil, fmt.Errorf("GPU compute not available")
	}
	if len(codes) == 0 {
		return []float32{}, nil
	}

	params.Size = int32(len(codes))

	// Create GPU buffers (host-visible for easy upload/download)
	inputSize := uint64(len(codes)) * uint64(unsafe.Sizeof(int32(0)))
	outputSize := uint64(len(codes)) * uint64(unsafe.Sizeof(float32(0)))

	inputBuf, err := g.ctx.CreateBuffer(inputSize, compute.BufferUsageStorage, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create input buffer: %w", err)
	}
	defer inputBuf.Destroy()

	outputBuf, err := g.ctx.CreateBuffer(outputSize, compute.BufferUsageStorage, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create output buffer: %w", err)
	}
	defer outputBuf.Destroy()

	// Upload input codes
	if err := uploadInt32Slice(inputBuf, codes); err != nil {
		return nil, fmt.Errorf("failed to upload input codes: %w", err)
	}

	// Set uniform parameters
	paramsBytes := structToBytes(&params)
	if err := g.dacPipeline.SetUniformRaw(0, paramsBytes); err != nil {
		return nil, fmt.Errorf("failed to set DAC parameters: %w", err)
	}

	// Bind buffers
	if err := g.dacPipeline.BindBuffer(1, inputBuf); err != nil {
		return nil, fmt.Errorf("failed to bind input buffer: %w", err)
	}
	if err := g.dacPipeline.BindBuffer(2, outputBuf); err != nil {
		return nil, fmt.Errorf("failed to bind output buffer: %w", err)
	}

	// Dispatch compute shader (256 threads per workgroup)
	numWorkgroups := uint32((len(codes) + 255) / 256)
	if err := g.dacPipeline.Dispatch(numWorkgroups, 1, 1); err != nil {
		return nil, fmt.Errorf("failed to dispatch DAC compute: %w", err)
	}

	// Download results
	results := make([]float32, len(codes))
	if err := outputBuf.DownloadFloat32(results); err != nil {
		return nil, fmt.Errorf("failed to download results: %w", err)
	}

	return results, nil
}

// BatchADC performs GPU-accelerated ADC conversion on a batch of analog voltages.
// Returns digital codes and reconstructed quantized voltages.
func (g *GPUPeripherals) BatchADC(voltages []float32, params ADCParams) ([]int32, []float32, error) {
	if !g.IsAvailable() {
		return nil, nil, fmt.Errorf("GPU compute not available")
	}
	if len(voltages) == 0 {
		return []int32{}, []float32{}, nil
	}

	params.Size = int32(len(voltages))

	// Create GPU buffers
	inputSize := uint64(len(voltages)) * uint64(unsafe.Sizeof(float32(0)))
	codesSize := uint64(len(voltages)) * uint64(unsafe.Sizeof(int32(0)))
	quantSize := uint64(len(voltages)) * uint64(unsafe.Sizeof(float32(0)))

	inputBuf, err := g.ctx.CreateBuffer(inputSize, compute.BufferUsageStorage, false)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create input buffer: %w", err)
	}
	defer inputBuf.Destroy()

	codesBuf, err := g.ctx.CreateBuffer(codesSize, compute.BufferUsageStorage, false)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create codes buffer: %w", err)
	}
	defer codesBuf.Destroy()

	quantBuf, err := g.ctx.CreateBuffer(quantSize, compute.BufferUsageStorage, false)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create quantized buffer: %w", err)
	}
	defer quantBuf.Destroy()

	// Upload input voltages
	if err := inputBuf.UploadFloat32(voltages); err != nil {
		return nil, nil, fmt.Errorf("failed to upload input voltages: %w", err)
	}

	// Set uniform parameters
	paramsBytes := structToBytes(&params)
	if err := g.adcPipeline.SetUniformRaw(0, paramsBytes); err != nil {
		return nil, nil, fmt.Errorf("failed to set ADC parameters: %w", err)
	}

	// Bind buffers
	if err := g.adcPipeline.BindBuffer(1, inputBuf); err != nil {
		return nil, nil, fmt.Errorf("failed to bind input buffer: %w", err)
	}
	if err := g.adcPipeline.BindBuffer(2, codesBuf); err != nil {
		return nil, nil, fmt.Errorf("failed to bind codes buffer: %w", err)
	}
	if err := g.adcPipeline.BindBuffer(3, quantBuf); err != nil {
		return nil, nil, fmt.Errorf("failed to bind quantized buffer: %w", err)
	}

	// Dispatch compute shader
	numWorkgroups := uint32((len(voltages) + 255) / 256)
	if err := g.adcPipeline.Dispatch(numWorkgroups, 1, 1); err != nil {
		return nil, nil, fmt.Errorf("failed to dispatch ADC compute: %w", err)
	}

	// Download results
	codes := make([]int32, len(voltages))
	if err := downloadInt32Slice(codesBuf, codes); err != nil {
		return nil, nil, fmt.Errorf("failed to download codes: %w", err)
	}

	quantized := make([]float32, len(voltages))
	if err := quantBuf.DownloadFloat32(quantized); err != nil {
		return nil, nil, fmt.Errorf("failed to download quantized voltages: %w", err)
	}

	return codes, quantized, nil
}

// BatchTIA performs GPU-accelerated TIA (transimpedance amplifier) conversion.
// Converts currents to voltages with noise and saturation modeling.
func (g *GPUPeripherals) BatchTIA(currents []float32, params TIAParams) ([]float32, error) {
	if !g.IsAvailable() {
		return nil, fmt.Errorf("GPU compute not available")
	}
	if len(currents) == 0 {
		return []float32{}, nil
	}

	params.Size = int32(len(currents))

	// Create GPU buffers
	bufferSize := uint64(len(currents)) * uint64(unsafe.Sizeof(float32(0)))

	inputBuf, err := g.ctx.CreateBuffer(bufferSize, compute.BufferUsageStorage, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create input buffer: %w", err)
	}
	defer inputBuf.Destroy()

	outputBuf, err := g.ctx.CreateBuffer(bufferSize, compute.BufferUsageStorage, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create output buffer: %w", err)
	}
	defer outputBuf.Destroy()

	// Upload input currents
	if err := inputBuf.UploadFloat32(currents); err != nil {
		return nil, fmt.Errorf("failed to upload input currents: %w", err)
	}

	// Set uniform parameters
	paramsBytes := structToBytes(&params)
	if err := g.tiaPipeline.SetUniformRaw(0, paramsBytes); err != nil {
		return nil, fmt.Errorf("failed to set TIA parameters: %w", err)
	}

	// Bind buffers
	if err := g.tiaPipeline.BindBuffer(1, inputBuf); err != nil {
		return nil, fmt.Errorf("failed to bind input buffer: %w", err)
	}
	if err := g.tiaPipeline.BindBuffer(2, outputBuf); err != nil {
		return nil, fmt.Errorf("failed to bind output buffer: %w", err)
	}

	// Dispatch compute shader
	numWorkgroups := uint32((len(currents) + 255) / 256)
	if err := g.tiaPipeline.Dispatch(numWorkgroups, 1, 1); err != nil {
		return nil, fmt.Errorf("failed to dispatch TIA compute: %w", err)
	}

	// Download results
	results := make([]float32, len(currents))
	if err := outputBuf.DownloadFloat32(results); err != nil {
		return nil, fmt.Errorf("failed to download results: %w", err)
	}

	return results, nil
}

// Destroy releases all GPU resources.
// Must be called when done with GPU peripherals to avoid memory leaks.
func (g *GPUPeripherals) Destroy() {
	if g.tiaPipeline != nil {
		g.tiaPipeline.Destroy()
		g.tiaPipeline = nil
	}
	if g.adcPipeline != nil {
		g.adcPipeline.Destroy()
		g.adcPipeline = nil
	}
	if g.dacPipeline != nil {
		g.dacPipeline.Destroy()
		g.dacPipeline = nil
	}
	if g.ctx != nil {
		g.ctx.Destroy()
		g.ctx = nil
	}
}

// structToBytes converts a struct to a byte slice for uniform buffer upload.
// Uses unsafe to directly interpret struct memory layout.
func structToBytes(s interface{}) []byte {
	switch v := s.(type) {
	case *DACParams:
		size := unsafe.Sizeof(*v)
		return unsafe.Slice((*byte)(unsafe.Pointer(v)), size)
	case *ADCParams:
		size := unsafe.Sizeof(*v)
		return unsafe.Slice((*byte)(unsafe.Pointer(v)), size)
	case *TIAParams:
		size := unsafe.Sizeof(*v)
		return unsafe.Slice((*byte)(unsafe.Pointer(v)), size)
	default:
		panic(fmt.Sprintf("unsupported struct type: %T", s))
	}
}

// uploadInt32Slice uploads an int32 slice to a GPU buffer.
func uploadInt32Slice(buf *compute.GPUBuffer, data []int32) error {
	if len(data) == 0 {
		return nil
	}

	byteSize := len(data) * int(unsafe.Sizeof(int32(0)))
	if uint64(byteSize) > buf.Size() {
		return fmt.Errorf("data size (%d bytes) exceeds buffer size (%d bytes)", byteSize, buf.Size())
	}

	// Convert to bytes
	bytes := make([]byte, byteSize)
	for i, v := range data {
		binary.LittleEndian.PutUint32(bytes[i*4:], uint32(v))
	}

	return buf.Upload(bytes)
}

// downloadInt32Slice downloads an int32 slice from a GPU buffer.
func downloadInt32Slice(buf *compute.GPUBuffer, data []int32) error {
	if len(data) == 0 {
		return nil
	}

	byteSize := len(data) * int(unsafe.Sizeof(int32(0)))
	if uint64(byteSize) > buf.Size() {
		return fmt.Errorf("data size (%d bytes) exceeds buffer size (%d bytes)", byteSize, buf.Size())
	}

	// Download bytes
	bytes := make([]byte, byteSize)
	if err := buf.Download(bytes); err != nil {
		return err
	}

	// Convert from bytes
	for i := range data {
		data[i] = int32(binary.LittleEndian.Uint32(bytes[i*4:]))
	}

	return nil
}

// Verify constants match shader expectations at compile time
var (
	_ = [1]struct{}{}[unsafe.Sizeof(DACParams{})-32] // Must be 32 bytes (8 x 4-byte fields)
	_ = [1]struct{}{}[unsafe.Sizeof(ADCParams{})-32] // Must be 32 bytes (8 x 4-byte fields)
	_ = [1]struct{}{}[unsafe.Sizeof(TIAParams{})-32] // Must be 32 bytes (8 x 4-byte fields)
)

// Helper to validate std140 alignment at runtime
func init() {
	// Verify struct sizes match expected uniform buffer layout
	if unsafe.Sizeof(DACParams{}) != 32 {
		panic(fmt.Sprintf("DACParams size mismatch: got %d, expected 32", unsafe.Sizeof(DACParams{})))
	}
	if unsafe.Sizeof(ADCParams{}) != 32 {
		panic(fmt.Sprintf("ADCParams size mismatch: got %d, expected 32", unsafe.Sizeof(ADCParams{})))
	}
	if unsafe.Sizeof(TIAParams{}) != 32 {
		panic(fmt.Sprintf("TIAParams size mismatch: got %d, expected 32", unsafe.Sizeof(TIAParams{})))
	}

	// Verify field alignment for std140 (all fields 4-byte aligned)
	verifyAlignment := func(name string, offset, expected uintptr) {
		if offset != expected {
			panic(fmt.Sprintf("%s field alignment mismatch: got %d, expected %d", name, offset, expected))
		}
	}

	// DACParams layout checks
	dacParams := DACParams{}
	verifyAlignment("DACParams.Bits", unsafe.Offsetof(dacParams.Bits), 0)
	verifyAlignment("DACParams.VrefP", unsafe.Offsetof(dacParams.VrefP), 4)
	verifyAlignment("DACParams.VrefN", unsafe.Offsetof(dacParams.VrefN), 8)
	verifyAlignment("DACParams.INLMax", unsafe.Offsetof(dacParams.INLMax), 12)
	verifyAlignment("DACParams.DNLMax", unsafe.Offsetof(dacParams.DNLMax), 16)
	verifyAlignment("DACParams.Size", unsafe.Offsetof(dacParams.Size), 20)
	verifyAlignment("DACParams.Seed", unsafe.Offsetof(dacParams.Seed), 24)
	verifyAlignment("DACParams.Padding", unsafe.Offsetof(dacParams.Padding), 28)

	// ADCParams layout checks
	adcParams := ADCParams{}
	verifyAlignment("ADCParams.Bits", unsafe.Offsetof(adcParams.Bits), 0)
	verifyAlignment("ADCParams.VrefP", unsafe.Offsetof(adcParams.VrefP), 4)
	verifyAlignment("ADCParams.VrefN", unsafe.Offsetof(adcParams.VrefN), 8)
	verifyAlignment("ADCParams.INLMax", unsafe.Offsetof(adcParams.INLMax), 12)
	verifyAlignment("ADCParams.DNLMax", unsafe.Offsetof(adcParams.DNLMax), 16)
	verifyAlignment("ADCParams.NoiseRMS", unsafe.Offsetof(adcParams.NoiseRMS), 20)
	verifyAlignment("ADCParams.Size", unsafe.Offsetof(adcParams.Size), 24)
	verifyAlignment("ADCParams.Seed", unsafe.Offsetof(adcParams.Seed), 28)

	// TIAParams layout checks
	tiaParams := TIAParams{}
	verifyAlignment("TIAParams.Gain", unsafe.Offsetof(tiaParams.Gain), 0)
	verifyAlignment("TIAParams.Bandwidth", unsafe.Offsetof(tiaParams.Bandwidth), 4)
	verifyAlignment("TIAParams.InputNoise", unsafe.Offsetof(tiaParams.InputNoise), 8)
	verifyAlignment("TIAParams.Vmax", unsafe.Offsetof(tiaParams.Vmax), 12)
	verifyAlignment("TIAParams.Size", unsafe.Offsetof(tiaParams.Size), 16)
	verifyAlignment("TIAParams.Seed", unsafe.Offsetof(tiaParams.Seed), 20)
	verifyAlignment("TIAParams.Padding1", unsafe.Offsetof(tiaParams.Padding1), 24)
	verifyAlignment("TIAParams.Padding2", unsafe.Offsetof(tiaParams.Padding2), 28)
}

// DefaultDACParams returns typical DAC parameters for 5-bit FeCIM.
// IMPORTANT: Must match DefaultDAC() in dac.go for CPU/GPU consistency.
func DefaultDACParams(size int) DACParams {
	return DACParams{
		Bits:   5,              // 5 bits = 32 levels
		VrefP:  1.5,            // +1.5V reference (matches dac.go DefaultDAC)
		VrefN:  -1.5,           // -1.5V reference (matches dac.go DefaultDAC)
		INLMax: 0.5,            // 0.5 LSB INL (typical for R-2R)
		DNLMax: 0.25,           // 0.25 LSB DNL
		Size:   int32(size),    // Number of conversions
		Seed:   float32(12345), // Random seed for reproducibility
	}
}

// DefaultADCParams returns typical ADC parameters for 5-bit FeCIM.
// IMPORTANT: Must match DefaultADC() in adc.go for CPU/GPU consistency.
func DefaultADCParams(size int) ADCParams {
	return ADCParams{
		Bits:     5,              // 5 bits = 32 levels
		VrefP:    1.0,            // +1V reference (matches adc.go DefaultADC)
		VrefN:    0.0,            // 0V reference (matches adc.go DefaultADC)
		INLMax:   0.5,            // 0.5 LSB INL
		DNLMax:   0.25,           // 0.25 LSB DNL
		NoiseRMS: 1e-3,           // 1mV input noise (typical SAR ADC)
		Size:     int32(size),    // Number of conversions
		Seed:     float32(54321), // Random seed
	}
}

// DefaultTIAParams returns typical TIA parameters for FeCIM readout.
func DefaultTIAParams(size int) TIAParams {
	return TIAParams{
		Gain:       1e6,           // 1MΩ transimpedance gain
		Bandwidth:  1e6,           // 1MHz bandwidth
		InputNoise: 1e-12,         // 1pA/sqrt(Hz) input noise
		Vmax:       1.0,           // ±1V output range
		Size:       int32(size),   // Number of conversions
		Seed:       float32(9876), // Random seed
	}
}

// Compile-time check that math package is imported (needed for future extensions)
var _ = math.Pi

// findRepoRoot searches for the repository root by looking for go.mod.
// Starts from the current working directory and walks up the directory tree.
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Walk up directory tree looking for go.mod
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding go.mod
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("could not find repository root (no go.mod found)")
}
