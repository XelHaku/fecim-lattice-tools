// visualization_edge.go - Real-time visualization and edge deployment optimization for CIM
// Implements WebGPU-style compute abstractions and TinyML deployment pipelines
// Based on research: WebGPU 10× faster than WebGL, sub-30ms inference
// Edge targets: MAX78000 (microjoule), Mythic (100× efficiency), Coral (512 GOPS)

package layers

import (
	"fmt"
	"math"
	"sort"
	"sync"
)

// ============================================================================
// WebGPU Compute Abstraction
// ============================================================================

// WGSLDataType represents WGSL shader data types
type WGSLDataType int

const (
	WGSL_F32 WGSLDataType = iota
	WGSL_F16
	WGSL_I32
	WGSL_U32
	WGSL_VEC2F
	WGSL_VEC4F
	WGSL_MAT4X4F
)

// GPUBufferUsage defines WebGPU buffer usage flags
type GPUBufferUsage int

const (
	BufferUsageStorage GPUBufferUsage = 1 << iota
	BufferUsageUniform
	BufferUsageCopySrc
	BufferUsageCopyDst
	BufferUsageMapRead
	BufferUsageMapWrite
)

// GPUBuffer represents a WebGPU buffer abstraction
type GPUBuffer struct {
	ID       string
	Size     int
	Usage    GPUBufferUsage
	Data     []float32
	Mapped   bool
	ReadOnly bool
}

// NewGPUBuffer creates a new GPU buffer
func NewGPUBuffer(id string, size int, usage GPUBufferUsage) *GPUBuffer {
	return &GPUBuffer{
		ID:    id,
		Size:  size,
		Usage: usage,
		Data:  make([]float32, size),
	}
}

// WriteData writes data to the buffer
func (b *GPUBuffer) WriteData(data []float32) error {
	if b.ReadOnly {
		return fmt.Errorf("cannot write to read-only buffer")
	}
	if len(data) > b.Size {
		return fmt.Errorf("data size %d exceeds buffer size %d", len(data), b.Size)
	}
	copy(b.Data, data)
	return nil
}

// ReadData reads data from the buffer
func (b *GPUBuffer) ReadData() []float32 {
	result := make([]float32, len(b.Data))
	copy(result, b.Data)
	return result
}

// WGSLShader represents a WGSL compute shader
type WGSLShader struct {
	Name           string
	EntryPoint     string
	WorkgroupSize  [3]int // x, y, z
	Code           string
	BindGroupIndex int
	Bindings       []ShaderBinding
}

// ShaderBinding defines a resource binding in the shader
type ShaderBinding struct {
	Group      int
	Binding    int
	BufferType string // "storage", "uniform", "read-only-storage"
	DataType   WGSLDataType
}

// ComputePipeline represents a WebGPU compute pipeline
type ComputePipeline struct {
	Shader       *WGSLShader
	BindGroups   map[int]*BindGroup
	Dispatches   int
	WorkgroupsX  int
	WorkgroupsY  int
	WorkgroupsZ  int
}

// BindGroup represents a group of buffer bindings
type BindGroup struct {
	Index   int
	Entries map[int]*GPUBuffer
}

// NewComputePipeline creates a new compute pipeline
func NewComputePipeline(shader *WGSLShader) *ComputePipeline {
	return &ComputePipeline{
		Shader:     shader,
		BindGroups: make(map[int]*BindGroup),
	}
}

// SetBindGroup sets a bind group for the pipeline
func (p *ComputePipeline) SetBindGroup(group int, bindGroup *BindGroup) {
	p.BindGroups[group] = bindGroup
}

// Dispatch executes the compute pipeline
func (p *ComputePipeline) Dispatch(workgroupsX, workgroupsY, workgroupsZ int) {
	p.WorkgroupsX = workgroupsX
	p.WorkgroupsY = workgroupsY
	p.WorkgroupsZ = workgroupsZ
	p.Dispatches++
}

// ============================================================================
// Matrix Multiply Compute Shader (WGSL-style)
// ============================================================================

// MatMulConfig configures matrix multiplication shader
type MatMulConfig struct {
	TileSize       int // Typically 16 or 32
	WorkgroupSizeX int
	WorkgroupSizeY int
	UseSharedMem   bool
	UseTensorCores bool // Simulated tensor core acceleration
}

// DefaultMatMulConfig returns optimized defaults
func DefaultMatMulConfig() MatMulConfig {
	return MatMulConfig{
		TileSize:       16,
		WorkgroupSizeX: 16,
		WorkgroupSizeY: 16,
		UseSharedMem:   true,
		UseTensorCores: false,
	}
}

// MatMulShaderGen generates WGSL-style matrix multiply shader code
type MatMulShaderGen struct {
	Config MatMulConfig
}

// GenerateShader generates the WGSL shader code
func (g *MatMulShaderGen) GenerateShader(M, K, N int) *WGSLShader {
	tileSize := g.Config.TileSize

	code := fmt.Sprintf(`
// Matrix multiplication: C[M×N] = A[M×K] × B[K×N]
// Optimized with shared memory tiling for 1+ TFLOPS

struct Uniforms {
    M: u32,
    K: u32,
    N: u32,
    padding: u32,
}

@group(0) @binding(0) var<uniform> uniforms: Uniforms;
@group(0) @binding(1) var<storage, read> matrixA: array<f32>;
@group(0) @binding(2) var<storage, read> matrixB: array<f32>;
@group(0) @binding(3) var<storage, read_write> matrixC: array<f32>;

const TILE_SIZE: u32 = %du;
var<workgroup> tileA: array<f32, %d>;
var<workgroup> tileB: array<f32, %d>;

@compute @workgroup_size(%d, %d, 1)
fn main(
    @builtin(global_invocation_id) global_id: vec3<u32>,
    @builtin(local_invocation_id) local_id: vec3<u32>,
    @builtin(workgroup_id) workgroup_id: vec3<u32>
) {
    let row = workgroup_id.y * TILE_SIZE + local_id.y;
    let col = workgroup_id.x * TILE_SIZE + local_id.x;

    var sum: f32 = 0.0;

    let numTiles = (uniforms.K + TILE_SIZE - 1u) / TILE_SIZE;

    for (var t: u32 = 0u; t < numTiles; t = t + 1u) {
        // Load tile from A into shared memory
        let aCol = t * TILE_SIZE + local_id.x;
        if (row < uniforms.M && aCol < uniforms.K) {
            tileA[local_id.y * TILE_SIZE + local_id.x] = matrixA[row * uniforms.K + aCol];
        } else {
            tileA[local_id.y * TILE_SIZE + local_id.x] = 0.0;
        }

        // Load tile from B into shared memory
        let bRow = t * TILE_SIZE + local_id.y;
        if (bRow < uniforms.K && col < uniforms.N) {
            tileB[local_id.y * TILE_SIZE + local_id.x] = matrixB[bRow * uniforms.N + col];
        } else {
            tileB[local_id.y * TILE_SIZE + local_id.x] = 0.0;
        }

        workgroupBarrier();

        // Compute partial dot product
        for (var k: u32 = 0u; k < TILE_SIZE; k = k + 1u) {
            sum = sum + tileA[local_id.y * TILE_SIZE + k] * tileB[k * TILE_SIZE + local_id.x];
        }

        workgroupBarrier();
    }

    if (row < uniforms.M && col < uniforms.N) {
        matrixC[row * uniforms.N + col] = sum;
    }
}
`, tileSize, tileSize*tileSize, tileSize*tileSize,
		g.Config.WorkgroupSizeX, g.Config.WorkgroupSizeY)

	return &WGSLShader{
		Name:          "matmul_tiled",
		EntryPoint:    "main",
		WorkgroupSize: [3]int{g.Config.WorkgroupSizeX, g.Config.WorkgroupSizeY, 1},
		Code:          code,
		Bindings: []ShaderBinding{
			{0, 0, "uniform", WGSL_U32},
			{0, 1, "read-only-storage", WGSL_F32},
			{0, 2, "read-only-storage", WGSL_F32},
			{0, 3, "storage", WGSL_F32},
		},
	}
}

// SimulateMatMul simulates the GPU matrix multiply
func (g *MatMulShaderGen) SimulateMatMul(A, B []float32, M, K, N int) []float32 {
	C := make([]float32, M*N)

	// Simulate tiled multiplication
	tileSize := g.Config.TileSize

	for tileRow := 0; tileRow < M; tileRow += tileSize {
		for tileCol := 0; tileCol < N; tileCol += tileSize {
			for tileK := 0; tileK < K; tileK += tileSize {
				// Process tile
				for i := 0; i < tileSize && tileRow+i < M; i++ {
					for j := 0; j < tileSize && tileCol+j < N; j++ {
						sum := float32(0)
						for k := 0; k < tileSize && tileK+k < K; k++ {
							aIdx := (tileRow+i)*K + (tileK+k)
							bIdx := (tileK+k)*N + (tileCol+j)
							sum += A[aIdx] * B[bIdx]
						}
						cIdx := (tileRow+i)*N + (tileCol+j)
						C[cIdx] += sum
					}
				}
			}
		}
	}

	return C
}

// ============================================================================
// Real-Time Visualization Pipeline
// ============================================================================

// VisualizationMode defines different visualization modes
type VisualizationMode int

const (
	VizModeActivation VisualizationMode = iota
	VizModeGradient
	VizModeCrossbarState
	VizModeAttentionMap
	VizModeFeatureMap
	VizModeSpikeRaster
)

// ColorMap defines color mapping schemes
type ColorMap int

const (
	ColorMapJet ColorMap = iota
	ColorMapViridis
	ColorMapPlasma
	ColorMapInferno
	ColorMapMagma
	ColorMapCoolwarm
	ColorMapGrayscale
)

// Color represents RGB color
type Color struct {
	R, G, B float32
}

// ColorMapper maps values to colors
type ColorMapper struct {
	Scheme ColorMap
	MinVal float32
	MaxVal float32
}

// NewColorMapper creates a new color mapper
func NewColorMapper(scheme ColorMap) *ColorMapper {
	return &ColorMapper{
		Scheme: scheme,
		MinVal: 0,
		MaxVal: 1,
	}
}

// SetRange sets the value range
func (cm *ColorMapper) SetRange(minVal, maxVal float32) {
	cm.MinVal = minVal
	cm.MaxVal = maxVal
}

// Map maps a value to a color
func (cm *ColorMapper) Map(value float32) Color {
	// Normalize value
	t := (value - cm.MinVal) / (cm.MaxVal - cm.MinVal)
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}

	switch cm.Scheme {
	case ColorMapJet:
		return cm.jetColor(t)
	case ColorMapViridis:
		return cm.viridisColor(t)
	case ColorMapGrayscale:
		return Color{t, t, t}
	default:
		return cm.viridisColor(t)
	}
}

func (cm *ColorMapper) jetColor(t float32) Color {
	// Jet colormap approximation
	r := float32(math.Min(1, math.Max(0, 1.5-math.Abs(float64(4*t-3)))))
	g := float32(math.Min(1, math.Max(0, 1.5-math.Abs(float64(4*t-2)))))
	b := float32(math.Min(1, math.Max(0, 1.5-math.Abs(float64(4*t-1)))))
	return Color{r, g, b}
}

func (cm *ColorMapper) viridisColor(t float32) Color {
	// Viridis colormap approximation
	r := float32(0.267004 + t*(0.282327+t*(-0.897035+t*(1.868753+t*(-1.593541+t*0.449324)))))
	g := float32(0.004874 + t*(1.345424+t*(-0.619648+t*(-0.178768+t*(0.641738+t*-0.302223)))))
	b := float32(0.329415 + t*(0.120367+t*(1.102575+t*(-2.251035+t*(1.877047+t*-0.520159)))))
	return Color{
		R: float32(math.Max(0, math.Min(1, float64(r)))),
		G: float32(math.Max(0, math.Min(1, float64(g)))),
		B: float32(math.Max(0, math.Min(1, float64(b)))),
	}
}

// FrameBuffer represents a visualization frame buffer
type FrameBuffer struct {
	Width    int
	Height   int
	Pixels   []Color
	DepthBuf []float32
}

// NewFrameBuffer creates a new frame buffer
func NewFrameBuffer(width, height int) *FrameBuffer {
	return &FrameBuffer{
		Width:    width,
		Height:   height,
		Pixels:   make([]Color, width*height),
		DepthBuf: make([]float32, width*height),
	}
}

// Clear clears the frame buffer
func (fb *FrameBuffer) Clear(color Color) {
	for i := range fb.Pixels {
		fb.Pixels[i] = color
		fb.DepthBuf[i] = 1.0
	}
}

// SetPixel sets a pixel color
func (fb *FrameBuffer) SetPixel(x, y int, color Color) {
	if x >= 0 && x < fb.Width && y >= 0 && y < fb.Height {
		fb.Pixels[y*fb.Width+x] = color
	}
}

// GetPixel gets a pixel color
func (fb *FrameBuffer) GetPixel(x, y int) Color {
	if x >= 0 && x < fb.Width && y >= 0 && y < fb.Height {
		return fb.Pixels[y*fb.Width+x]
	}
	return Color{0, 0, 0}
}

// VisualizationPipeline handles real-time visualization
type VisualizationPipeline struct {
	Mode           VisualizationMode
	ColorMapper    *ColorMapper
	FrameBuffer    *FrameBuffer
	FPS            float64
	FrameCount     int
	LastFrameTime  float64
	RenderQueue    chan *RenderTask
	mu             sync.Mutex
}

// RenderTask represents a rendering task
type RenderTask struct {
	Type      string
	Data      []float32
	Rows      int
	Cols      int
	Timestamp float64
}

// NewVisualizationPipeline creates a new visualization pipeline
func NewVisualizationPipeline(width, height int) *VisualizationPipeline {
	return &VisualizationPipeline{
		Mode:        VizModeActivation,
		ColorMapper: NewColorMapper(ColorMapViridis),
		FrameBuffer: NewFrameBuffer(width, height),
		RenderQueue: make(chan *RenderTask, 100),
	}
}

// SetMode sets the visualization mode
func (vp *VisualizationPipeline) SetMode(mode VisualizationMode) {
	vp.mu.Lock()
	defer vp.mu.Unlock()
	vp.Mode = mode
}

// RenderActivationMap renders activation values as heatmap
func (vp *VisualizationPipeline) RenderActivationMap(activations []float32, rows, cols int) {
	vp.mu.Lock()
	defer vp.mu.Unlock()

	// Find min/max for normalization
	minVal, maxVal := activations[0], activations[0]
	for _, v := range activations {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	vp.ColorMapper.SetRange(minVal, maxVal)

	// Calculate cell size
	cellW := vp.FrameBuffer.Width / cols
	cellH := vp.FrameBuffer.Height / rows

	// Render cells
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			idx := r*cols + c
			if idx >= len(activations) {
				continue
			}

			color := vp.ColorMapper.Map(activations[idx])

			// Fill cell
			for y := r * cellH; y < (r+1)*cellH && y < vp.FrameBuffer.Height; y++ {
				for x := c * cellW; x < (c+1)*cellW && x < vp.FrameBuffer.Width; x++ {
					vp.FrameBuffer.SetPixel(x, y, color)
				}
			}
		}
	}

	vp.FrameCount++
}

// RenderCrossbarState renders crossbar conductance state
func (vp *VisualizationPipeline) RenderCrossbarState(conductances []float32, rows, cols int) {
	vp.RenderActivationMap(conductances, rows, cols)
}

// RenderSpikeRaster renders spike timing raster plot
func (vp *VisualizationPipeline) RenderSpikeRaster(spikeTimes [][]float32, neurons int, timeWindow float64) {
	vp.mu.Lock()
	defer vp.mu.Unlock()

	vp.FrameBuffer.Clear(Color{0, 0, 0})

	neuronHeight := vp.FrameBuffer.Height / neurons
	timeScale := float64(vp.FrameBuffer.Width) / timeWindow

	spikeColor := Color{0, 1, 0} // Green spikes

	for n := 0; n < neurons && n < len(spikeTimes); n++ {
		y := n * neuronHeight + neuronHeight/2
		for _, t := range spikeTimes[n] {
			x := int(float64(t) * timeScale)
			if x >= 0 && x < vp.FrameBuffer.Width {
				// Draw vertical line for spike
				for dy := -neuronHeight/3; dy <= neuronHeight/3; dy++ {
					vp.FrameBuffer.SetPixel(x, y+dy, spikeColor)
				}
			}
		}
	}

	vp.FrameCount++
}

// ExportFrame exports frame as raw RGB data
func (vp *VisualizationPipeline) ExportFrame() []byte {
	vp.mu.Lock()
	defer vp.mu.Unlock()

	data := make([]byte, vp.FrameBuffer.Width*vp.FrameBuffer.Height*3)
	for i, color := range vp.FrameBuffer.Pixels {
		data[i*3] = byte(color.R * 255)
		data[i*3+1] = byte(color.G * 255)
		data[i*3+2] = byte(color.B * 255)
	}
	return data
}

// ============================================================================
// Edge Deployment Optimization
// ============================================================================

// EdgeTarget represents target edge hardware
type EdgeTarget int

const (
	EdgeTargetMAX78000 EdgeTarget = iota // Maxim MAX78000, microjoule inference
	EdgeTargetMythic                     // Mythic M1076, analog 100× efficiency
	EdgeTargetCoral                      // Google Coral, 4 TOPS @ 2W
	EdgeTargetKendryteK210               // Kendryte K210, 0.3W
	EdgeTargetESP32S3                    // ESP32-S3, neural accelerator
	EdgeTargetGreenWaves                 // GreenWaves GAP9, ultra-low power
	EdgeTargetGenericMCU                 // Generic MCU without accelerator
)

// EdgeHardwareProfile defines hardware constraints
type EdgeHardwareProfile struct {
	Target           EdgeTarget
	Name             string
	MaxOps           int64   // Max ops/second
	MaxPower         float64 // Watts
	SRAM             int     // KB
	FlashSize        int     // KB
	HasAccelerator   bool
	AcceleratorType  string
	BitWidth         int     // Native bit width
	VectorWidth      int     // SIMD vector width
	EnergyPerMAC     float64 // Joules per MAC operation
}

// GetEdgeProfile returns profile for target hardware
func GetEdgeProfile(target EdgeTarget) EdgeHardwareProfile {
	profiles := map[EdgeTarget]EdgeHardwareProfile{
		EdgeTargetMAX78000: {
			Target:          EdgeTargetMAX78000,
			Name:            "MAX78000",
			MaxOps:          442000000, // 442 GOPS peak
			MaxPower:        0.001,     // ~1mW typical
			SRAM:            512,
			FlashSize:       512,
			HasAccelerator:  true,
			AcceleratorType: "CNN",
			BitWidth:        8,
			VectorWidth:     64,
			EnergyPerMAC:    1e-12, // ~1pJ/MAC
		},
		EdgeTargetMythic: {
			Target:          EdgeTargetMythic,
			Name:            "Mythic M1076",
			MaxOps:          25000000000, // 25 TOPS
			MaxPower:        4.0,
			SRAM:            8192,
			FlashSize:       0, // In-memory compute
			HasAccelerator:  true,
			AcceleratorType: "Analog CIM",
			BitWidth:        8,
			VectorWidth:     1024,
			EnergyPerMAC:    0.5e-12, // ~0.5pJ/MAC
		},
		EdgeTargetCoral: {
			Target:          EdgeTargetCoral,
			Name:            "Google Coral TPU",
			MaxOps:          4000000000, // 4 TOPS
			MaxPower:        2.0,
			SRAM:            8192,
			FlashSize:       0,
			HasAccelerator:  true,
			AcceleratorType: "TPU",
			BitWidth:        8,
			VectorWidth:     256,
			EnergyPerMAC:    0.5e-12,
		},
		EdgeTargetKendryteK210: {
			Target:          EdgeTargetKendryteK210,
			Name:            "Kendryte K210",
			MaxOps:          230000000, // 230 GOPS
			MaxPower:        0.3,
			SRAM:            8192,
			FlashSize:       16384,
			HasAccelerator:  true,
			AcceleratorType: "KPU",
			BitWidth:        8,
			VectorWidth:     64,
			EnergyPerMAC:    1.3e-12,
		},
		EdgeTargetESP32S3: {
			Target:          EdgeTargetESP32S3,
			Name:            "ESP32-S3",
			MaxOps:          240000000, // 240MHz dual-core
			MaxPower:        0.5,
			SRAM:            512,
			FlashSize:       16384,
			HasAccelerator:  false,
			AcceleratorType: "",
			BitWidth:        32,
			VectorWidth:     1,
			EnergyPerMAC:    10e-12,
		},
		EdgeTargetGreenWaves: {
			Target:          EdgeTargetGreenWaves,
			Name:            "GreenWaves GAP9",
			MaxOps:          50000000000, // 50 GOPS
			MaxPower:        0.075,       // 75mW
			SRAM:            1536,
			FlashSize:       2048,
			HasAccelerator:  true,
			AcceleratorType: "NE16",
			BitWidth:        8,
			VectorWidth:     16,
			EnergyPerMAC:    0.33e-12,
		},
		EdgeTargetGenericMCU: {
			Target:          EdgeTargetGenericMCU,
			Name:            "Generic MCU",
			MaxOps:          100000000,
			MaxPower:        0.1,
			SRAM:            256,
			FlashSize:       1024,
			HasAccelerator:  false,
			AcceleratorType: "",
			BitWidth:        32,
			VectorWidth:     1,
			EnergyPerMAC:    50e-12,
		},
	}

	if profile, ok := profiles[target]; ok {
		return profile
	}
	return profiles[EdgeTargetGenericMCU]
}

// EdgeQuantizer quantizes models for edge deployment
type EdgeQuantizer struct {
	TargetBits      int
	CalibrationData [][]float32
	PerChannel      bool
	Symmetric       bool
}

// QuantizationParams holds quantization parameters
type QuantizationParams struct {
	Scale      float32
	ZeroPoint  int32
	MinVal     float32
	MaxVal     float32
	BitWidth   int
	PerChannel bool
}

// NewEdgeQuantizer creates a new edge quantizer
func NewEdgeQuantizer(bits int) *EdgeQuantizer {
	return &EdgeQuantizer{
		TargetBits:      bits,
		CalibrationData: make([][]float32, 0),
		PerChannel:      false,
		Symmetric:       true,
	}
}

// Calibrate calibrates quantization using sample data
func (eq *EdgeQuantizer) Calibrate(data [][]float32) {
	eq.CalibrationData = append(eq.CalibrationData, data...)
}

// ComputeParams computes quantization parameters
func (eq *EdgeQuantizer) ComputeParams(weights []float32) QuantizationParams {
	minVal, maxVal := weights[0], weights[0]
	for _, w := range weights {
		if w < minVal {
			minVal = w
		}
		if w > maxVal {
			maxVal = w
		}
	}

	qmin := int32(0)
	qmax := int32((1 << eq.TargetBits) - 1)

	var scale float32
	var zeroPoint int32

	if eq.Symmetric {
		// Symmetric quantization
		absMax := float32(math.Max(math.Abs(float64(minVal)), math.Abs(float64(maxVal))))
		scale = absMax / float32(qmax/2)
		zeroPoint = int32(qmax/2 + 1)
	} else {
		// Asymmetric quantization
		scale = (maxVal - minVal) / float32(qmax-qmin)
		zeroPoint = qmin - int32(minVal/scale)
	}

	if scale == 0 {
		scale = 1e-8
	}

	return QuantizationParams{
		Scale:      scale,
		ZeroPoint:  zeroPoint,
		MinVal:     minVal,
		MaxVal:     maxVal,
		BitWidth:   eq.TargetBits,
		PerChannel: eq.PerChannel,
	}
}

// Quantize quantizes float32 weights to int8
func (eq *EdgeQuantizer) Quantize(weights []float32) ([]int8, QuantizationParams) {
	params := eq.ComputeParams(weights)

	qmax := int32((1 << eq.TargetBits) - 1)
	quantized := make([]int8, len(weights))

	for i, w := range weights {
		q := int32(w/params.Scale) + params.ZeroPoint
		if q < 0 {
			q = 0
		}
		if q > qmax {
			q = qmax
		}
		quantized[i] = int8(q - 128) // Center around 0
	}

	return quantized, params
}

// Dequantize converts quantized values back to float32
func (eq *EdgeQuantizer) Dequantize(quantized []int8, params QuantizationParams) []float32 {
	result := make([]float32, len(quantized))
	for i, q := range quantized {
		result[i] = float32(int32(q)+128-params.ZeroPoint) * params.Scale
	}
	return result
}

// EdgePruner prunes models for edge constraints
type EdgePruner struct {
	TargetSparsity float64
	Method         string // "magnitude", "structured", "block"
	BlockSize      int
}

// NewEdgePruner creates a new edge pruner
func NewEdgePruner(sparsity float64, method string) *EdgePruner {
	return &EdgePruner{
		TargetSparsity: sparsity,
		Method:         method,
		BlockSize:      4,
	}
}

// PruneWeights prunes weights to target sparsity
func (ep *EdgePruner) PruneWeights(weights []float32) ([]float32, []bool) {
	n := len(weights)
	mask := make([]bool, n) // true = keep, false = prune
	pruned := make([]float32, n)
	copy(pruned, weights)

	switch ep.Method {
	case "magnitude":
		ep.magnitudePrune(pruned, mask)
	case "structured":
		ep.structuredPrune(pruned, mask)
	case "block":
		ep.blockPrune(pruned, mask)
	default:
		ep.magnitudePrune(pruned, mask)
	}

	return pruned, mask
}

func (ep *EdgePruner) magnitudePrune(weights []float32, mask []bool) {
	// Sort by absolute magnitude
	type weightIdx struct {
		absVal float64
		idx    int
	}

	sorted := make([]weightIdx, len(weights))
	for i, w := range weights {
		sorted[i] = weightIdx{math.Abs(float64(w)), i}
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].absVal < sorted[j].absVal
	})

	// Prune smallest weights
	numPrune := int(float64(len(weights)) * ep.TargetSparsity)
	for i := 0; i < len(weights); i++ {
		if i < numPrune {
			mask[sorted[i].idx] = false
			weights[sorted[i].idx] = 0
		} else {
			mask[sorted[i].idx] = true
		}
	}
}

func (ep *EdgePruner) structuredPrune(weights []float32, mask []bool) {
	// Prune entire rows/columns based on L2 norm
	// Simplified: prune individual weights by magnitude
	ep.magnitudePrune(weights, mask)
}

func (ep *EdgePruner) blockPrune(weights []float32, mask []bool) {
	// Block-wise pruning
	blockSize := ep.BlockSize
	numBlocks := (len(weights) + blockSize - 1) / blockSize

	// Compute block magnitudes
	type blockInfo struct {
		magnitude float64
		startIdx  int
	}

	blocks := make([]blockInfo, numBlocks)
	for b := 0; b < numBlocks; b++ {
		startIdx := b * blockSize
		endIdx := startIdx + blockSize
		if endIdx > len(weights) {
			endIdx = len(weights)
		}

		sumSq := float64(0)
		for i := startIdx; i < endIdx; i++ {
			sumSq += float64(weights[i] * weights[i])
		}
		blocks[b] = blockInfo{math.Sqrt(sumSq), startIdx}
	}

	// Sort blocks by magnitude
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].magnitude < blocks[j].magnitude
	})

	// Prune smallest blocks
	numPruneBlocks := int(float64(numBlocks) * ep.TargetSparsity)

	// Initialize all as keep
	for i := range mask {
		mask[i] = true
	}

	for b := 0; b < numPruneBlocks; b++ {
		startIdx := blocks[b].startIdx
		endIdx := startIdx + blockSize
		if endIdx > len(weights) {
			endIdx = len(weights)
		}

		for i := startIdx; i < endIdx; i++ {
			mask[i] = false
			weights[i] = 0
		}
	}
}

// EdgeDeploymentPipeline handles end-to-end edge deployment
type EdgeDeploymentPipeline struct {
	Target      EdgeHardwareProfile
	Quantizer   *EdgeQuantizer
	Pruner      *EdgePruner
	Optimizer   *MemoryOptimizer
}

// MemoryOptimizer optimizes memory layout for edge
type MemoryOptimizer struct {
	MaxSRAM        int // KB
	TilingEnabled  bool
	BufferReuse    bool
	InPlaceOps     bool
}

// NewMemoryOptimizer creates a new memory optimizer
func NewMemoryOptimizer(maxSRAM int) *MemoryOptimizer {
	return &MemoryOptimizer{
		MaxSRAM:       maxSRAM,
		TilingEnabled: true,
		BufferReuse:   true,
		InPlaceOps:    true,
	}
}

// OptimizeBuffers optimizes activation buffer allocation
func (mo *MemoryOptimizer) OptimizeBuffers(layerSizes []int) ([]int, int) {
	// Compute buffer offsets with reuse
	offsets := make([]int, len(layerSizes))
	peakMemory := 0

	if mo.BufferReuse {
		// Ping-pong buffer strategy
		bufferA := 0
		bufferB := 0

		for i, size := range layerSizes {
			if i%2 == 0 {
				offsets[i] = 0
				bufferA = size
			} else {
				offsets[i] = bufferA
				bufferB = size
			}
			if bufferA+bufferB > peakMemory {
				peakMemory = bufferA + bufferB
			}
		}
	} else {
		// Sequential allocation
		offset := 0
		for i, size := range layerSizes {
			offsets[i] = offset
			offset += size
		}
		peakMemory = offset
	}

	return offsets, peakMemory
}

// NewEdgeDeploymentPipeline creates a new deployment pipeline
func NewEdgeDeploymentPipeline(target EdgeTarget) *EdgeDeploymentPipeline {
	profile := GetEdgeProfile(target)
	return &EdgeDeploymentPipeline{
		Target:    profile,
		Quantizer: NewEdgeQuantizer(profile.BitWidth),
		Pruner:    NewEdgePruner(0.5, "magnitude"),
		Optimizer: NewMemoryOptimizer(profile.SRAM),
	}
}

// DeploymentResult contains deployment analysis results
type DeploymentResult struct {
	FitsInMemory     bool
	RequiredSRAM     int // KB
	RequiredFlash    int // KB
	EstimatedLatency float64 // ms
	EstimatedEnergy  float64 // mJ
	EstimatedOpsPerSec int64
	Warnings         []string
	Optimizations    []string
}

// Analyze analyzes model for deployment feasibility
func (edp *EdgeDeploymentPipeline) Analyze(modelParams int, totalMACs int64) DeploymentResult {
	result := DeploymentResult{
		Warnings:      make([]string, 0),
		Optimizations: make([]string, 0),
	}

	// Calculate memory requirements
	bytesPerParam := edp.Target.BitWidth / 8
	if bytesPerParam < 1 {
		bytesPerParam = 1
	}

	result.RequiredFlash = (modelParams * bytesPerParam) / 1024 // KB
	result.RequiredSRAM = (modelParams * bytesPerParam) / 4 / 1024 // Estimate 25% for activations

	result.FitsInMemory = result.RequiredSRAM <= edp.Target.SRAM &&
		(edp.Target.FlashSize == 0 || result.RequiredFlash <= edp.Target.FlashSize)

	// Calculate performance
	if edp.Target.MaxOps > 0 {
		result.EstimatedOpsPerSec = edp.Target.MaxOps
		result.EstimatedLatency = float64(totalMACs) / float64(edp.Target.MaxOps) * 1000 // ms
	}

	// Calculate energy
	result.EstimatedEnergy = float64(totalMACs) * edp.Target.EnergyPerMAC * 1000 // mJ

	// Generate warnings
	if !result.FitsInMemory {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Model requires %d KB SRAM, target has %d KB",
				result.RequiredSRAM, edp.Target.SRAM))
	}

	if result.EstimatedLatency > 100 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Estimated latency %.1f ms may be too high for real-time",
				result.EstimatedLatency))
	}

	// Suggest optimizations
	if result.RequiredSRAM > edp.Target.SRAM {
		result.Optimizations = append(result.Optimizations,
			"Apply tiling to reduce peak SRAM usage")
	}

	if edp.Target.BitWidth == 8 {
		result.Optimizations = append(result.Optimizations,
			"Model already quantized to INT8")
	} else {
		result.Optimizations = append(result.Optimizations,
			fmt.Sprintf("Quantize to %d-bit for %s", edp.Target.BitWidth, edp.Target.Name))
	}

	return result
}

// ============================================================================
// TinyML Inference Engine
// ============================================================================

// TinyMLConfig configures TinyML inference
type TinyMLConfig struct {
	UseFixedPoint    bool
	FixedPointBits   int
	EnableSIMD       bool
	EnableUnrolling  bool
	LoopUnrollFactor int
}

// DefaultTinyMLConfig returns default TinyML configuration
func DefaultTinyMLConfig() TinyMLConfig {
	return TinyMLConfig{
		UseFixedPoint:    true,
		FixedPointBits:   15, // Q15 format
		EnableSIMD:       true,
		EnableUnrolling:  true,
		LoopUnrollFactor: 4,
	}
}

// TinyMLLayer represents a layer in TinyML format
type TinyMLLayer struct {
	Type       string
	Name       string
	InputSize  int
	OutputSize int
	Weights    []int8
	Biases     []int32
	Scale      float32
	ZeroPoint  int32
}

// TinyMLModel represents a quantized TinyML model
type TinyMLModel struct {
	Config     TinyMLConfig
	Layers     []*TinyMLLayer
	InputScale float32
	InputZP    int32
}

// NewTinyMLModel creates a new TinyML model
func NewTinyMLModel(config TinyMLConfig) *TinyMLModel {
	return &TinyMLModel{
		Config:     config,
		Layers:     make([]*TinyMLLayer, 0),
		InputScale: 1.0 / 255.0,
		InputZP:    0,
	}
}

// AddLayer adds a layer to the model
func (tm *TinyMLModel) AddLayer(layer *TinyMLLayer) {
	tm.Layers = append(tm.Layers, layer)
}

// QuantizeInput quantizes float input to int8
func (tm *TinyMLModel) QuantizeInput(input []float32) []int8 {
	result := make([]int8, len(input))
	for i, v := range input {
		q := int32(v/tm.InputScale) + tm.InputZP
		if q < -128 {
			q = -128
		}
		if q > 127 {
			q = 127
		}
		result[i] = int8(q)
	}
	return result
}

// Inference runs quantized inference
func (tm *TinyMLModel) Inference(input []int8) []int8 {
	current := input

	for _, layer := range tm.Layers {
		switch layer.Type {
		case "fc":
			current = tm.fcLayerInt8(current, layer)
		case "conv":
			current = tm.convLayerInt8(current, layer)
		case "relu":
			current = tm.reluInt8(current)
		}
	}

	return current
}

// fcLayerInt8 performs fully-connected layer in int8
func (tm *TinyMLModel) fcLayerInt8(input []int8, layer *TinyMLLayer) []int8 {
	output := make([]int8, layer.OutputSize)

	for o := 0; o < layer.OutputSize; o++ {
		var acc int32 = layer.Biases[o]

		// MAC operations in int32 accumulator
		for i := 0; i < layer.InputSize; i++ {
			acc += int32(input[i]) * int32(layer.Weights[o*layer.InputSize+i])
		}

		// Requantize to int8
		scaled := float32(acc) * layer.Scale
		q := int32(scaled) + layer.ZeroPoint
		if q < -128 {
			q = -128
		}
		if q > 127 {
			q = 127
		}
		output[o] = int8(q)
	}

	return output
}

// convLayerInt8 performs convolution in int8
func (tm *TinyMLModel) convLayerInt8(input []int8, layer *TinyMLLayer) []int8 {
	// Simplified 1D conv
	return tm.fcLayerInt8(input, layer)
}

// reluInt8 performs ReLU activation in int8
func (tm *TinyMLModel) reluInt8(input []int8) []int8 {
	output := make([]int8, len(input))
	for i, v := range input {
		if v < 0 {
			output[i] = 0
		} else {
			output[i] = v
		}
	}
	return output
}

// ============================================================================
// WebGPU CIM Visualization Demo
// ============================================================================

// CIMVisualizerConfig configures CIM visualization
type CIMVisualizerConfig struct {
	CrossbarRows    int
	CrossbarCols    int
	UpdateRateHz    float64
	ShowVoltages    bool
	ShowCurrents    bool
	ShowWeights     bool
	AnimateSpikes   bool
}

// DefaultCIMVisualizerConfig returns default configuration
func DefaultCIMVisualizerConfig() CIMVisualizerConfig {
	return CIMVisualizerConfig{
		CrossbarRows:  64,
		CrossbarCols:  64,
		UpdateRateHz:  30,
		ShowVoltages:  true,
		ShowCurrents:  true,
		ShowWeights:   true,
		AnimateSpikes: false,
	}
}

// CIMVisualizer provides real-time CIM visualization
type CIMVisualizer struct {
	Config           CIMVisualizerConfig
	Pipeline         *VisualizationPipeline
	MatMulShader     *WGSLShader
	WeightBuffer     *GPUBuffer
	InputBuffer      *GPUBuffer
	OutputBuffer     *GPUBuffer
	CurrentConductances []float32
	CurrentVoltages  []float32
	CurrentOutputs   []float32
	FrameTime        float64
	mu               sync.Mutex
}

// NewCIMVisualizer creates a new CIM visualizer
func NewCIMVisualizer(config CIMVisualizerConfig) *CIMVisualizer {
	totalCells := config.CrossbarRows * config.CrossbarCols

	// Initialize shader generator
	shaderGen := &MatMulShaderGen{Config: DefaultMatMulConfig()}

	return &CIMVisualizer{
		Config:            config,
		Pipeline:          NewVisualizationPipeline(512, 512),
		MatMulShader:      shaderGen.GenerateShader(config.CrossbarRows, config.CrossbarCols, 1),
		WeightBuffer:      NewGPUBuffer("weights", totalCells, BufferUsageStorage),
		InputBuffer:       NewGPUBuffer("input", config.CrossbarCols, BufferUsageStorage),
		OutputBuffer:      NewGPUBuffer("output", config.CrossbarRows, BufferUsageStorage),
		CurrentConductances: make([]float32, totalCells),
		CurrentVoltages:  make([]float32, config.CrossbarCols),
		CurrentOutputs:   make([]float32, config.CrossbarRows),
	}
}

// UpdateConductances updates crossbar conductance values
func (cv *CIMVisualizer) UpdateConductances(conductances []float32) {
	cv.mu.Lock()
	defer cv.mu.Unlock()

	if len(conductances) <= len(cv.CurrentConductances) {
		copy(cv.CurrentConductances, conductances)
	}
	cv.WeightBuffer.WriteData(conductances)
}

// UpdateVoltages updates input voltages
func (cv *CIMVisualizer) UpdateVoltages(voltages []float32) {
	cv.mu.Lock()
	defer cv.mu.Unlock()

	if len(voltages) <= len(cv.CurrentVoltages) {
		copy(cv.CurrentVoltages, voltages)
	}
	cv.InputBuffer.WriteData(voltages)
}

// ComputeMVM computes matrix-vector multiply and updates visualization
func (cv *CIMVisualizer) ComputeMVM() []float32 {
	cv.mu.Lock()
	defer cv.mu.Unlock()

	// Simulate MVM: output = weights * input
	rows := cv.Config.CrossbarRows
	cols := cv.Config.CrossbarCols

	for r := 0; r < rows; r++ {
		sum := float32(0)
		for c := 0; c < cols; c++ {
			sum += cv.CurrentConductances[r*cols+c] * cv.CurrentVoltages[c]
		}
		cv.CurrentOutputs[r] = sum
	}

	return cv.CurrentOutputs
}

// RenderFrame renders a visualization frame
func (cv *CIMVisualizer) RenderFrame() {
	cv.mu.Lock()
	defer cv.mu.Unlock()

	if cv.Config.ShowWeights {
		cv.Pipeline.RenderActivationMap(
			cv.CurrentConductances,
			cv.Config.CrossbarRows,
			cv.Config.CrossbarCols,
		)
	}
}

// GetFrameData returns current frame as RGB data
func (cv *CIMVisualizer) GetFrameData() []byte {
	return cv.Pipeline.ExportFrame()
}

// GetVisualizationStats returns current visualization statistics
func (cv *CIMVisualizer) GetVisualizationStats() map[string]interface{} {
	cv.mu.Lock()
	defer cv.mu.Unlock()

	// Compute statistics
	var minG, maxG, sumG float32 = cv.CurrentConductances[0], cv.CurrentConductances[0], 0
	for _, g := range cv.CurrentConductances {
		if g < minG {
			minG = g
		}
		if g > maxG {
			maxG = g
		}
		sumG += g
	}
	avgG := sumG / float32(len(cv.CurrentConductances))

	var minOut, maxOut float32 = cv.CurrentOutputs[0], cv.CurrentOutputs[0]
	for _, o := range cv.CurrentOutputs {
		if o < minOut {
			minOut = o
		}
		if o > maxOut {
			maxOut = o
		}
	}

	return map[string]interface{}{
		"conductance_min":     minG,
		"conductance_max":     maxG,
		"conductance_avg":     avgG,
		"output_min":          minOut,
		"output_max":          maxOut,
		"frame_count":         cv.Pipeline.FrameCount,
		"crossbar_rows":       cv.Config.CrossbarRows,
		"crossbar_cols":       cv.Config.CrossbarCols,
	}
}

// ============================================================================
// Performance Benchmarking
// ============================================================================

// EdgeBenchmark benchmarks edge deployment
type EdgeBenchmark struct {
	Target       EdgeHardwareProfile
	ModelMACs    int64
	ModelParams  int
	Iterations   int
	Results      []BenchmarkRun
}

// BenchmarkRun represents a single benchmark run
type BenchmarkRun struct {
	Iteration    int
	LatencyMs    float64
	EnergyMJ     float64
	Throughput   float64 // inferences/sec
	MemoryKB     int
}

// NewEdgeBenchmark creates a new edge benchmark
func NewEdgeBenchmark(target EdgeTarget, macs int64, params int) *EdgeBenchmark {
	return &EdgeBenchmark{
		Target:      GetEdgeProfile(target),
		ModelMACs:   macs,
		ModelParams: params,
		Iterations:  100,
		Results:     make([]BenchmarkRun, 0),
	}
}

// Run executes the benchmark
func (eb *EdgeBenchmark) Run() {
	eb.Results = make([]BenchmarkRun, eb.Iterations)

	for i := 0; i < eb.Iterations; i++ {
		// Simulate inference latency
		baseLatency := float64(eb.ModelMACs) / float64(eb.Target.MaxOps) * 1000
		// Add some variance
		variance := baseLatency * 0.1 * (0.5 - float64(i%10)/10)
		latency := baseLatency + variance

		// Calculate energy
		energy := float64(eb.ModelMACs) * eb.Target.EnergyPerMAC * 1000 // mJ

		// Calculate throughput
		throughput := 1000.0 / latency // inferences/sec

		// Memory estimate
		memoryKB := (eb.ModelParams * eb.Target.BitWidth / 8) / 1024

		eb.Results[i] = BenchmarkRun{
			Iteration:  i,
			LatencyMs:  latency,
			EnergyMJ:   energy,
			Throughput: throughput,
			MemoryKB:   memoryKB,
		}
	}
}

// GetSummary returns benchmark summary statistics
func (eb *EdgeBenchmark) GetSummary() map[string]float64 {
	if len(eb.Results) == 0 {
		return nil
	}

	var sumLatency, sumEnergy, sumThroughput float64
	minLatency := eb.Results[0].LatencyMs
	maxLatency := eb.Results[0].LatencyMs

	for _, r := range eb.Results {
		sumLatency += r.LatencyMs
		sumEnergy += r.EnergyMJ
		sumThroughput += r.Throughput
		if r.LatencyMs < minLatency {
			minLatency = r.LatencyMs
		}
		if r.LatencyMs > maxLatency {
			maxLatency = r.LatencyMs
		}
	}

	n := float64(len(eb.Results))

	return map[string]float64{
		"avg_latency_ms":     sumLatency / n,
		"min_latency_ms":     minLatency,
		"max_latency_ms":     maxLatency,
		"avg_energy_mj":      sumEnergy / n,
		"avg_throughput_ips": sumThroughput / n,
		"memory_kb":          float64(eb.Results[0].MemoryKB),
		"tops_w":             float64(eb.ModelMACs) / (sumEnergy/n) / 1e9, // TOPS/W
	}
}
