// Package layers provides neural network layer implementations for CIM simulation.
// init.go implements weight initialization methods for neural networks.
package layers

import (
	"math"
	"math/rand"
)

// =============================================================================
// Initialization Interface
// =============================================================================

// Initializer defines the interface for weight initialization
type Initializer interface {
	// Initialize fills weights with initialized values
	Initialize(weights [][]float64)

	// InitializeVec initializes a 1D vector (biases)
	InitializeVec(vec []float64)

	// Name returns the initializer name
	Name() string
}

// =============================================================================
// Xavier/Glorot Initialization
// =============================================================================

// XavierUniform implements Xavier uniform initialization
// Suitable for tanh and sigmoid activations
// W ~ U[-sqrt(6/(fan_in+fan_out)), sqrt(6/(fan_in+fan_out))]
type XavierUniform struct {
	Gain float64 // Scaling factor (default 1.0)
}

// NewXavierUniform creates Xavier uniform initializer
func NewXavierUniform() *XavierUniform {
	return &XavierUniform{Gain: 1.0}
}

// Initialize fills weights with Xavier uniform values
func (x *XavierUniform) Initialize(weights [][]float64) {
	if len(weights) == 0 || len(weights[0]) == 0 {
		return
	}

	fanIn := len(weights[0])
	fanOut := len(weights)
	bound := x.Gain * math.Sqrt(6.0/float64(fanIn+fanOut))

	for i := range weights {
		for j := range weights[i] {
			weights[i][j] = (rand.Float64()*2 - 1) * bound
		}
	}
}

// InitializeVec initializes bias to zeros (common practice)
func (x *XavierUniform) InitializeVec(vec []float64) {
	for i := range vec {
		vec[i] = 0
	}
}

func (x *XavierUniform) Name() string { return "xavier_uniform" }

// XavierNormal implements Xavier normal initialization
// W ~ N(0, sqrt(2/(fan_in+fan_out)))
type XavierNormal struct {
	Gain float64
}

// NewXavierNormal creates Xavier normal initializer
func NewXavierNormal() *XavierNormal {
	return &XavierNormal{Gain: 1.0}
}

// Initialize fills weights with Xavier normal values
func (x *XavierNormal) Initialize(weights [][]float64) {
	if len(weights) == 0 || len(weights[0]) == 0 {
		return
	}

	fanIn := len(weights[0])
	fanOut := len(weights)
	std := x.Gain * math.Sqrt(2.0/float64(fanIn+fanOut))

	for i := range weights {
		for j := range weights[i] {
			weights[i][j] = rand.NormFloat64() * std
		}
	}
}

func (x *XavierNormal) InitializeVec(vec []float64) {
	for i := range vec {
		vec[i] = 0
	}
}

func (x *XavierNormal) Name() string { return "xavier_normal" }

// =============================================================================
// Kaiming/He Initialization
// =============================================================================

// KaimingUniform implements Kaiming/He uniform initialization
// Suitable for ReLU and variants
// W ~ U[-bound, bound] where bound = sqrt(6/fan_in) for fan_in mode
type KaimingUniform struct {
	Mode         string  // "fan_in" or "fan_out"
	Nonlinearity string  // "relu", "leaky_relu", etc.
	A            float64 // Negative slope for leaky_relu
}

// NewKaimingUniform creates Kaiming uniform initializer
func NewKaimingUniform() *KaimingUniform {
	return &KaimingUniform{
		Mode:         "fan_in",
		Nonlinearity: "relu",
		A:            0,
	}
}

// Initialize fills weights with Kaiming uniform values
func (k *KaimingUniform) Initialize(weights [][]float64) {
	if len(weights) == 0 || len(weights[0]) == 0 {
		return
	}

	fanIn := len(weights[0])
	fanOut := len(weights)

	var fan int
	if k.Mode == "fan_out" {
		fan = fanOut
	} else {
		fan = fanIn
	}

	gain := k.calculateGain()
	std := gain / math.Sqrt(float64(fan))
	bound := math.Sqrt(3.0) * std

	for i := range weights {
		for j := range weights[i] {
			weights[i][j] = (rand.Float64()*2 - 1) * bound
		}
	}
}

func (k *KaimingUniform) calculateGain() float64 {
	switch k.Nonlinearity {
	case "relu":
		return math.Sqrt(2.0)
	case "leaky_relu":
		return math.Sqrt(2.0 / (1 + k.A*k.A))
	case "tanh":
		return 5.0 / 3.0
	case "sigmoid":
		return 1.0
	default:
		return 1.0
	}
}

func (k *KaimingUniform) InitializeVec(vec []float64) {
	for i := range vec {
		vec[i] = 0
	}
}

func (k *KaimingUniform) Name() string { return "kaiming_uniform" }

// KaimingNormal implements Kaiming/He normal initialization
// W ~ N(0, sqrt(2/fan_in)) for ReLU
type KaimingNormal struct {
	Mode         string
	Nonlinearity string
	A            float64
}

// NewKaimingNormal creates Kaiming normal initializer
func NewKaimingNormal() *KaimingNormal {
	return &KaimingNormal{
		Mode:         "fan_in",
		Nonlinearity: "relu",
		A:            0,
	}
}

// Initialize fills weights with Kaiming normal values
func (k *KaimingNormal) Initialize(weights [][]float64) {
	if len(weights) == 0 || len(weights[0]) == 0 {
		return
	}

	fanIn := len(weights[0])
	fanOut := len(weights)

	var fan int
	if k.Mode == "fan_out" {
		fan = fanOut
	} else {
		fan = fanIn
	}

	gain := k.calculateGain()
	std := gain / math.Sqrt(float64(fan))

	for i := range weights {
		for j := range weights[i] {
			weights[i][j] = rand.NormFloat64() * std
		}
	}
}

func (k *KaimingNormal) calculateGain() float64 {
	switch k.Nonlinearity {
	case "relu":
		return math.Sqrt(2.0)
	case "leaky_relu":
		return math.Sqrt(2.0 / (1 + k.A*k.A))
	case "tanh":
		return 5.0 / 3.0
	default:
		return 1.0
	}
}

func (k *KaimingNormal) InitializeVec(vec []float64) {
	for i := range vec {
		vec[i] = 0
	}
}

func (k *KaimingNormal) Name() string { return "kaiming_normal" }

// =============================================================================
// LeCun Initialization
// =============================================================================

// LeCunUniform implements LeCun uniform initialization
// W ~ U[-sqrt(3/fan_in), sqrt(3/fan_in)]
type LeCunUniform struct{}

// NewLeCunUniform creates LeCun uniform initializer
func NewLeCunUniform() *LeCunUniform {
	return &LeCunUniform{}
}

func (l *LeCunUniform) Initialize(weights [][]float64) {
	if len(weights) == 0 || len(weights[0]) == 0 {
		return
	}

	fanIn := len(weights[0])
	bound := math.Sqrt(3.0 / float64(fanIn))

	for i := range weights {
		for j := range weights[i] {
			weights[i][j] = (rand.Float64()*2 - 1) * bound
		}
	}
}

func (l *LeCunUniform) InitializeVec(vec []float64) {
	for i := range vec {
		vec[i] = 0
	}
}

func (l *LeCunUniform) Name() string { return "lecun_uniform" }

// LeCunNormal implements LeCun normal initialization
// W ~ N(0, sqrt(1/fan_in))
type LeCunNormal struct{}

// NewLeCunNormal creates LeCun normal initializer
func NewLeCunNormal() *LeCunNormal {
	return &LeCunNormal{}
}

func (l *LeCunNormal) Initialize(weights [][]float64) {
	if len(weights) == 0 || len(weights[0]) == 0 {
		return
	}

	fanIn := len(weights[0])
	std := math.Sqrt(1.0 / float64(fanIn))

	for i := range weights {
		for j := range weights[i] {
			weights[i][j] = rand.NormFloat64() * std
		}
	}
}

func (l *LeCunNormal) InitializeVec(vec []float64) {
	for i := range vec {
		vec[i] = 0
	}
}

func (l *LeCunNormal) Name() string { return "lecun_normal" }

// =============================================================================
// Simple Initializers
// =============================================================================

// UniformInit initializes weights uniformly in [low, high]
type UniformInit struct {
	Low  float64
	High float64
}

// NewUniformInit creates uniform initializer
func NewUniformInit(low, high float64) *UniformInit {
	return &UniformInit{Low: low, High: high}
}

func (u *UniformInit) Initialize(weights [][]float64) {
	for i := range weights {
		for j := range weights[i] {
			weights[i][j] = u.Low + rand.Float64()*(u.High-u.Low)
		}
	}
}

func (u *UniformInit) InitializeVec(vec []float64) {
	for i := range vec {
		vec[i] = u.Low + rand.Float64()*(u.High-u.Low)
	}
}

func (u *UniformInit) Name() string { return "uniform" }

// NormalInit initializes weights from normal distribution
type NormalInit struct {
	Mean float64
	Std  float64
}

// NewNormalInit creates normal initializer
func NewNormalInit(mean, std float64) *NormalInit {
	return &NormalInit{Mean: mean, Std: std}
}

func (n *NormalInit) Initialize(weights [][]float64) {
	for i := range weights {
		for j := range weights[i] {
			weights[i][j] = n.Mean + rand.NormFloat64()*n.Std
		}
	}
}

func (n *NormalInit) InitializeVec(vec []float64) {
	for i := range vec {
		vec[i] = n.Mean + rand.NormFloat64()*n.Std
	}
}

func (n *NormalInit) Name() string { return "normal" }

// ConstantInit initializes all weights to a constant value
type ConstantInit struct {
	Value float64
}

// NewConstantInit creates constant initializer
func NewConstantInit(value float64) *ConstantInit {
	return &ConstantInit{Value: value}
}

func (c *ConstantInit) Initialize(weights [][]float64) {
	for i := range weights {
		for j := range weights[i] {
			weights[i][j] = c.Value
		}
	}
}

func (c *ConstantInit) InitializeVec(vec []float64) {
	for i := range vec {
		vec[i] = c.Value
	}
}

func (c *ConstantInit) Name() string { return "constant" }

// ZerosInit initializes all weights to zero
type ZerosInit struct{}

func (z *ZerosInit) Initialize(weights [][]float64) {
	for i := range weights {
		for j := range weights[i] {
			weights[i][j] = 0
		}
	}
}

func (z *ZerosInit) InitializeVec(vec []float64) {
	for i := range vec {
		vec[i] = 0
	}
}

func (z *ZerosInit) Name() string { return "zeros" }

// OnesInit initializes all weights to one
type OnesInit struct{}

func (o *OnesInit) Initialize(weights [][]float64) {
	for i := range weights {
		for j := range weights[i] {
			weights[i][j] = 1
		}
	}
}

func (o *OnesInit) InitializeVec(vec []float64) {
	for i := range vec {
		vec[i] = 1
	}
}

func (o *OnesInit) Name() string { return "ones" }

// =============================================================================
// Orthogonal Initialization
// =============================================================================

// OrthogonalInit initializes with orthogonal matrix (good for RNNs)
type OrthogonalInit struct {
	Gain float64
}

// NewOrthogonalInit creates orthogonal initializer
func NewOrthogonalInit() *OrthogonalInit {
	return &OrthogonalInit{Gain: 1.0}
}

// Initialize creates orthogonal matrix using QR decomposition approximation
func (o *OrthogonalInit) Initialize(weights [][]float64) {
	if len(weights) == 0 || len(weights[0]) == 0 {
		return
	}

	rows := len(weights)
	cols := len(weights[0])

	// Create random matrix
	flat := make([]float64, rows*cols)
	for i := range flat {
		flat[i] = rand.NormFloat64()
	}

	// Simple Gram-Schmidt orthogonalization for small matrices
	if rows <= cols {
		// Orthogonalize rows
		for i := 0; i < rows; i++ {
			// Start with random vector
			for j := 0; j < cols; j++ {
				weights[i][j] = flat[i*cols+j]
			}

			// Subtract projections onto previous rows
			for k := 0; k < i; k++ {
				dot := 0.0
				for j := 0; j < cols; j++ {
					dot += weights[i][j] * weights[k][j]
				}
				for j := 0; j < cols; j++ {
					weights[i][j] -= dot * weights[k][j]
				}
			}

			// Normalize
			norm := 0.0
			for j := 0; j < cols; j++ {
				norm += weights[i][j] * weights[i][j]
			}
			norm = math.Sqrt(norm)
			if norm > 1e-8 {
				for j := 0; j < cols; j++ {
					weights[i][j] = o.Gain * weights[i][j] / norm
				}
			}
		}
	} else {
		// For tall matrices, orthogonalize columns
		for j := 0; j < cols; j++ {
			for i := 0; i < rows; i++ {
				weights[i][j] = flat[i*cols+j]
			}

			for k := 0; k < j; k++ {
				dot := 0.0
				for i := 0; i < rows; i++ {
					dot += weights[i][j] * weights[i][k]
				}
				for i := 0; i < rows; i++ {
					weights[i][j] -= dot * weights[i][k]
				}
			}

			norm := 0.0
			for i := 0; i < rows; i++ {
				norm += weights[i][j] * weights[i][j]
			}
			norm = math.Sqrt(norm)
			if norm > 1e-8 {
				for i := 0; i < rows; i++ {
					weights[i][j] = o.Gain * weights[i][j] / norm
				}
			}
		}
	}
}

func (o *OrthogonalInit) InitializeVec(vec []float64) {
	for i := range vec {
		vec[i] = 0
	}
}

func (o *OrthogonalInit) Name() string { return "orthogonal" }

// =============================================================================
// Sparse Initialization
// =============================================================================

// SparseInit initializes with sparse matrix (good for regularization)
type SparseInit struct {
	Sparsity float64 // Fraction of zeros (0.0 to 1.0)
	Std      float64 // Std of non-zero elements
}

// NewSparseInit creates sparse initializer
func NewSparseInit(sparsity float64) *SparseInit {
	return &SparseInit{Sparsity: sparsity, Std: 0.01}
}

func (s *SparseInit) Initialize(weights [][]float64) {
	for i := range weights {
		for j := range weights[i] {
			if rand.Float64() < s.Sparsity {
				weights[i][j] = 0
			} else {
				weights[i][j] = rand.NormFloat64() * s.Std
			}
		}
	}
}

func (s *SparseInit) InitializeVec(vec []float64) {
	for i := range vec {
		vec[i] = 0
	}
}

func (s *SparseInit) Name() string { return "sparse" }

// =============================================================================
// CIM-Specific Initializers
// =============================================================================

// QuantizedInit initializes weights to quantized values
// Useful for hardware-aware initialization
type QuantizedInit struct {
	Bits       int     // Number of bits
	Symmetric  bool    // Symmetric quantization
	BaseInit   Initializer // Base initializer to quantize
}

// NewQuantizedInit creates quantized initializer
func NewQuantizedInit(bits int, baseInit Initializer) *QuantizedInit {
	return &QuantizedInit{
		Bits:      bits,
		Symmetric: true,
		BaseInit:  baseInit,
	}
}

func (q *QuantizedInit) Initialize(weights [][]float64) {
	// First apply base initialization
	q.BaseInit.Initialize(weights)

	// Then quantize
	levels := math.Pow(2, float64(q.Bits))
	if q.Symmetric {
		levels = levels - 1 // Reserve one level for zero
	}

	// Find max absolute value
	maxAbs := 0.0
	for i := range weights {
		for j := range weights[i] {
			if math.Abs(weights[i][j]) > maxAbs {
				maxAbs = math.Abs(weights[i][j])
			}
		}
	}

	if maxAbs < 1e-8 {
		return
	}

	// Quantize
	scale := (levels / 2) / maxAbs
	for i := range weights {
		for j := range weights[i] {
			quantized := math.Round(weights[i][j] * scale)
			weights[i][j] = quantized / scale
		}
	}
}

func (q *QuantizedInit) InitializeVec(vec []float64) {
	q.BaseInit.InitializeVec(vec)
}

func (q *QuantizedInit) Name() string { return "quantized" }

// CrossbarAwareInit initializes weights considering crossbar constraints
type CrossbarAwareInit struct {
	TileSize   int     // Crossbar tile dimension
	WeightBits int     // Weight precision
	MaxCond    float64 // Maximum conductance (normalized)
	MinCond    float64 // Minimum conductance (normalized)
}

// NewCrossbarAwareInit creates crossbar-aware initializer
func NewCrossbarAwareInit(tileSize, weightBits int) *CrossbarAwareInit {
	return &CrossbarAwareInit{
		TileSize:   tileSize,
		WeightBits: weightBits,
		MaxCond:    1.0,
		MinCond:    0.0,
	}
}

func (c *CrossbarAwareInit) Initialize(weights [][]float64) {
	if len(weights) == 0 || len(weights[0]) == 0 {
		return
	}

	fanIn := len(weights[0])
	fanOut := len(weights)

	// Use Xavier-like scaling but constrained to conductance range
	std := math.Sqrt(2.0 / float64(fanIn+fanOut))

	// Scale to fit in [MinCond, MaxCond] range
	// Map [-3*std, 3*std] to [MinCond, MaxCond]
	scale := (c.MaxCond - c.MinCond) / (6 * std)
	offset := (c.MaxCond + c.MinCond) / 2

	levels := math.Pow(2, float64(c.WeightBits)) - 1

	for i := range weights {
		for j := range weights[i] {
			// Generate value in ~[-3std, 3std]
			val := rand.NormFloat64() * std

			// Map to conductance range
			cond := val*scale + offset
			cond = math.Max(c.MinCond, math.Min(c.MaxCond, cond))

			// Quantize to available levels
			quantized := math.Round(cond * levels) / levels
			weights[i][j] = quantized
		}
	}
}

func (c *CrossbarAwareInit) InitializeVec(vec []float64) {
	for i := range vec {
		vec[i] = 0
	}
}

func (c *CrossbarAwareInit) Name() string { return "crossbar_aware" }

// =============================================================================
// Initializer Registry
// =============================================================================

// InitializerRegistry provides name-based initializer lookup
var InitializerRegistry = map[string]func() Initializer{
	"xavier_uniform":  func() Initializer { return NewXavierUniform() },
	"xavier_normal":   func() Initializer { return NewXavierNormal() },
	"glorot_uniform":  func() Initializer { return NewXavierUniform() }, // Alias
	"glorot_normal":   func() Initializer { return NewXavierNormal() },  // Alias
	"kaiming_uniform": func() Initializer { return NewKaimingUniform() },
	"kaiming_normal":  func() Initializer { return NewKaimingNormal() },
	"he_uniform":      func() Initializer { return NewKaimingUniform() }, // Alias
	"he_normal":       func() Initializer { return NewKaimingNormal() },  // Alias
	"lecun_uniform":   func() Initializer { return NewLeCunUniform() },
	"lecun_normal":    func() Initializer { return NewLeCunNormal() },
	"orthogonal":      func() Initializer { return NewOrthogonalInit() },
	"zeros":           func() Initializer { return &ZerosInit{} },
	"ones":            func() Initializer { return &OnesInit{} },
}

// GetInitializer returns an initializer by name
func GetInitializer(name string) Initializer {
	if creator, ok := InitializerRegistry[name]; ok {
		return creator()
	}
	return NewXavierUniform() // Default
}

// =============================================================================
// Utility Functions
// =============================================================================

// CalculateFanInOut returns fan_in and fan_out for a weight matrix
func CalculateFanInOut(weights [][]float64) (fanIn, fanOut int) {
	if len(weights) == 0 {
		return 0, 0
	}
	fanOut = len(weights)
	fanIn = len(weights[0])
	return
}

// CalculateGain returns recommended gain for activation function
func CalculateGain(activation string) float64 {
	switch activation {
	case "relu":
		return math.Sqrt(2.0)
	case "leaky_relu":
		return math.Sqrt(2.0 / 1.01) // Assuming slope 0.1
	case "tanh":
		return 5.0 / 3.0
	case "sigmoid":
		return 1.0
	case "selu":
		return 0.75 // Empirical
	default:
		return 1.0
	}
}

// InitializeLayer is a convenience function for initializing a layer
func InitializeLayer(weights [][]float64, biases []float64, initName string) {
	init := GetInitializer(initName)
	init.Initialize(weights)
	if biases != nil {
		init.InitializeVec(biases)
	}
}
