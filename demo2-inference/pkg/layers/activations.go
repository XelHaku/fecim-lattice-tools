// Package layers provides neural network layer implementations for crossbar arrays.
package layers

import (
	"math"
)

// Activation represents an activation function interface.
type Activation interface {
	Forward(x float64) float64
	ForwardVec(x []float64) []float64
	Derivative(x float64) float64 // For backpropagation
	Name() string
}

// ReLU implements Rectified Linear Unit activation.
// ReLU(x) = max(0, x)
type ReLU struct{}

func NewReLU() *ReLU { return &ReLU{} }

func (r *ReLU) Forward(x float64) float64 {
	if x > 0 {
		return x
	}
	return 0
}

func (r *ReLU) ForwardVec(x []float64) []float64 {
	result := make([]float64, len(x))
	for i, v := range x {
		result[i] = r.Forward(v)
	}
	return result
}

func (r *ReLU) Derivative(x float64) float64 {
	if x > 0 {
		return 1
	}
	return 0
}

func (r *ReLU) Name() string { return "relu" }

// LeakyReLU implements Leaky ReLU activation.
// LeakyReLU(x) = x if x > 0 else alpha * x
type LeakyReLU struct {
	Alpha float64 // Typically 0.01
}

func NewLeakyReLU(alpha float64) *LeakyReLU {
	if alpha == 0 {
		alpha = 0.01
	}
	return &LeakyReLU{Alpha: alpha}
}

func (l *LeakyReLU) Forward(x float64) float64 {
	if x > 0 {
		return x
	}
	return l.Alpha * x
}

func (l *LeakyReLU) ForwardVec(x []float64) []float64 {
	result := make([]float64, len(x))
	for i, v := range x {
		result[i] = l.Forward(v)
	}
	return result
}

func (l *LeakyReLU) Derivative(x float64) float64 {
	if x > 0 {
		return 1
	}
	return l.Alpha
}

func (l *LeakyReLU) Name() string { return "leaky_relu" }

// PReLU implements Parametric ReLU (learnable alpha).
type PReLU struct {
	Alpha []float64 // Per-channel learnable parameters
}

func NewPReLU(channels int) *PReLU {
	alpha := make([]float64, channels)
	for i := range alpha {
		alpha[i] = 0.25 // Default initialization
	}
	return &PReLU{Alpha: alpha}
}

func (p *PReLU) Forward(x float64) float64 {
	if x > 0 {
		return x
	}
	if len(p.Alpha) > 0 {
		return p.Alpha[0] * x
	}
	return 0.25 * x
}

func (p *PReLU) ForwardVec(x []float64) []float64 {
	result := make([]float64, len(x))
	for i, v := range x {
		if v > 0 {
			result[i] = v
		} else {
			alpha := 0.25
			if i < len(p.Alpha) {
				alpha = p.Alpha[i]
			} else if len(p.Alpha) > 0 {
				alpha = p.Alpha[0]
			}
			result[i] = alpha * v
		}
	}
	return result
}

func (p *PReLU) Derivative(x float64) float64 {
	if x > 0 {
		return 1
	}
	if len(p.Alpha) > 0 {
		return p.Alpha[0]
	}
	return 0.25
}

func (p *PReLU) Name() string { return "prelu" }

// ELU implements Exponential Linear Unit.
// ELU(x) = x if x > 0 else alpha * (exp(x) - 1)
type ELU struct {
	Alpha float64
}

func NewELU(alpha float64) *ELU {
	if alpha == 0 {
		alpha = 1.0
	}
	return &ELU{Alpha: alpha}
}

func (e *ELU) Forward(x float64) float64 {
	if x > 0 {
		return x
	}
	return e.Alpha * (math.Exp(x) - 1)
}

func (e *ELU) ForwardVec(x []float64) []float64 {
	result := make([]float64, len(x))
	for i, v := range x {
		result[i] = e.Forward(v)
	}
	return result
}

func (e *ELU) Derivative(x float64) float64 {
	if x > 0 {
		return 1
	}
	return e.Forward(x) + e.Alpha
}

func (e *ELU) Name() string { return "elu" }

// SELU implements Scaled Exponential Linear Unit.
// Self-normalizing activation for deep networks.
type SELU struct {
	Alpha  float64
	Lambda float64
}

func NewSELU() *SELU {
	return &SELU{
		Alpha:  1.6732632423543772848170429916717,
		Lambda: 1.0507009873554804934193349852946,
	}
}

func (s *SELU) Forward(x float64) float64 {
	if x > 0 {
		return s.Lambda * x
	}
	return s.Lambda * s.Alpha * (math.Exp(x) - 1)
}

func (s *SELU) ForwardVec(x []float64) []float64 {
	result := make([]float64, len(x))
	for i, v := range x {
		result[i] = s.Forward(v)
	}
	return result
}

func (s *SELU) Derivative(x float64) float64 {
	if x > 0 {
		return s.Lambda
	}
	return s.Lambda * s.Alpha * math.Exp(x)
}

func (s *SELU) Name() string { return "selu" }

// Sigmoid implements logistic sigmoid activation.
// Sigmoid(x) = 1 / (1 + exp(-x))
type Sigmoid struct{}

func NewSigmoid() *Sigmoid { return &Sigmoid{} }

func (s *Sigmoid) Forward(x float64) float64 {
	return 1 / (1 + math.Exp(-x))
}

func (s *Sigmoid) ForwardVec(x []float64) []float64 {
	result := make([]float64, len(x))
	for i, v := range x {
		result[i] = s.Forward(v)
	}
	return result
}

func (s *Sigmoid) Derivative(x float64) float64 {
	sig := s.Forward(x)
	return sig * (1 - sig)
}

func (s *Sigmoid) Name() string { return "sigmoid" }

// Tanh implements hyperbolic tangent activation.
type Tanh struct{}

func NewTanh() *Tanh { return &Tanh{} }

func (t *Tanh) Forward(x float64) float64 {
	return math.Tanh(x)
}

func (t *Tanh) ForwardVec(x []float64) []float64 {
	result := make([]float64, len(x))
	for i, v := range x {
		result[i] = t.Forward(v)
	}
	return result
}

func (t *Tanh) Derivative(x float64) float64 {
	tanh := math.Tanh(x)
	return 1 - tanh*tanh
}

func (t *Tanh) Name() string { return "tanh" }

// HardTanh implements hard (clamped) tanh.
// HardTanh(x) = clip(x, -1, 1)
type HardTanh struct {
	MinVal float64
	MaxVal float64
}

func NewHardTanh() *HardTanh {
	return &HardTanh{MinVal: -1, MaxVal: 1}
}

func (h *HardTanh) Forward(x float64) float64 {
	if x < h.MinVal {
		return h.MinVal
	}
	if x > h.MaxVal {
		return h.MaxVal
	}
	return x
}

func (h *HardTanh) ForwardVec(x []float64) []float64 {
	result := make([]float64, len(x))
	for i, v := range x {
		result[i] = h.Forward(v)
	}
	return result
}

func (h *HardTanh) Derivative(x float64) float64 {
	if x >= h.MinVal && x <= h.MaxVal {
		return 1
	}
	return 0
}

func (h *HardTanh) Name() string { return "hardtanh" }

// HardSigmoidAct implements hard (piecewise linear) sigmoid.
// HardSigmoid(x) = clip(0.2*x + 0.5, 0, 1)
// Hardware-friendly approximation to sigmoid.
type HardSigmoidAct struct{}

func NewHardSigmoidAct() *HardSigmoidAct { return &HardSigmoidAct{} }

func (h *HardSigmoidAct) Forward(x float64) float64 {
	result := 0.2*x + 0.5
	if result < 0 {
		return 0
	}
	if result > 1 {
		return 1
	}
	return result
}

func (h *HardSigmoidAct) ForwardVec(x []float64) []float64 {
	result := make([]float64, len(x))
	for i, v := range x {
		result[i] = h.Forward(v)
	}
	return result
}

func (h *HardSigmoidAct) Derivative(x float64) float64 {
	y := 0.2*x + 0.5
	if y > 0 && y < 1 {
		return 0.2
	}
	return 0
}

func (h *HardSigmoidAct) Name() string { return "hardsigmoid" }

// HardSwish implements hard swish activation.
// HardSwish(x) = x * HardSigmoid(x)
// Used in MobileNetV3, hardware-friendly.
type HardSwish struct {
	hardsig *HardSigmoidAct
}

func NewHardSwish() *HardSwish {
	return &HardSwish{hardsig: NewHardSigmoidAct()}
}

func (h *HardSwish) Forward(x float64) float64 {
	return x * h.hardsig.Forward(x)
}

func (h *HardSwish) ForwardVec(x []float64) []float64 {
	result := make([]float64, len(x))
	for i, v := range x {
		result[i] = h.Forward(v)
	}
	return result
}

func (h *HardSwish) Derivative(x float64) float64 {
	if x <= -3 {
		return 0
	}
	if x >= 3 {
		return 1
	}
	return (2*x + 3) / 6
}

func (h *HardSwish) Name() string { return "hardswish" }

// Softplus implements softplus activation.
// Softplus(x) = log(1 + exp(x))
// Smooth approximation to ReLU.
type Softplus struct {
	Beta      float64 // Sharpness parameter
	Threshold float64 // Above this, return linear
}

func NewSoftplus() *Softplus {
	return &Softplus{Beta: 1, Threshold: 20}
}

func (s *Softplus) Forward(x float64) float64 {
	if x*s.Beta > s.Threshold {
		return x
	}
	return math.Log(1+math.Exp(s.Beta*x)) / s.Beta
}

func (s *Softplus) ForwardVec(x []float64) []float64 {
	result := make([]float64, len(x))
	for i, v := range x {
		result[i] = s.Forward(v)
	}
	return result
}

func (s *Softplus) Derivative(x float64) float64 {
	return 1 / (1 + math.Exp(-s.Beta*x))
}

func (s *Softplus) Name() string { return "softplus" }

// Mish implements Mish activation.
// Mish(x) = x * tanh(softplus(x))
type Mish struct {
	softplus *Softplus
}

func NewMish() *Mish {
	return &Mish{softplus: NewSoftplus()}
}

func (m *Mish) Forward(x float64) float64 {
	return x * math.Tanh(m.softplus.Forward(x))
}

func (m *Mish) ForwardVec(x []float64) []float64 {
	result := make([]float64, len(x))
	for i, v := range x {
		result[i] = m.Forward(v)
	}
	return result
}

func (m *Mish) Derivative(x float64) float64 {
	sp := m.softplus.Forward(x)
	tanhSp := math.Tanh(sp)
	sigmoid := 1 / (1 + math.Exp(-x))
	sech2 := 1 - tanhSp*tanhSp
	return tanhSp + x*sech2*sigmoid
}

func (m *Mish) Name() string { return "mish" }

// GELUAct implements GELU activation as Activation interface.
type GELUAct struct{}

func NewGELUAct() *GELUAct { return &GELUAct{} }

func (g *GELUAct) Forward(x float64) float64 {
	return GELU(x)
}

func (g *GELUAct) ForwardVec(x []float64) []float64 {
	return GELUVec(x)
}

func (g *GELUAct) Derivative(x float64) float64 {
	// Approximate derivative
	cdf := 0.5 * (1 + math.Erf(x/math.Sqrt(2)))
	pdf := math.Exp(-x*x/2) / math.Sqrt(2*math.Pi)
	return cdf + x*pdf
}

func (g *GELUAct) Name() string { return "gelu" }

// SiLUAct implements SiLU/Swish as Activation interface.
type SiLUAct struct{}

func NewSiLUAct() *SiLUAct { return &SiLUAct{} }

func (s *SiLUAct) Forward(x float64) float64 {
	return SiLU(x)
}

func (s *SiLUAct) ForwardVec(x []float64) []float64 {
	return SiLUVec(x)
}

func (s *SiLUAct) Derivative(x float64) float64 {
	sigmoid := 1 / (1 + math.Exp(-x))
	return sigmoid + x*sigmoid*(1-sigmoid)
}

func (s *SiLUAct) Name() string { return "silu" }

// Softmax implements softmax activation for classification.
type Softmax struct {
	Temperature float64 // Temperature scaling
}

func NewSoftmax() *Softmax {
	return &Softmax{Temperature: 1.0}
}

func (s *Softmax) Forward(x float64) float64 {
	return math.Exp(x / s.Temperature)
}

func (s *Softmax) ForwardVec(x []float64) []float64 {
	// Numerical stability: subtract max
	maxVal := x[0]
	for _, v := range x {
		if v > maxVal {
			maxVal = v
		}
	}

	result := make([]float64, len(x))
	sum := 0.0
	for i, v := range x {
		result[i] = math.Exp((v - maxVal) / s.Temperature)
		sum += result[i]
	}

	for i := range result {
		result[i] /= sum
	}

	return result
}

func (s *Softmax) Derivative(x float64) float64 {
	// Softmax derivative is complex (Jacobian), return simplified
	return 1.0
}

func (s *Softmax) Name() string { return "softmax" }

// BinaryStep implements binary step activation.
// BinaryStep(x) = 1 if x >= 0 else 0
type BinaryStep struct {
	Threshold float64
}

func NewBinaryStep() *BinaryStep {
	return &BinaryStep{Threshold: 0}
}

func (b *BinaryStep) Forward(x float64) float64 {
	if x >= b.Threshold {
		return 1
	}
	return 0
}

func (b *BinaryStep) ForwardVec(x []float64) []float64 {
	result := make([]float64, len(x))
	for i, v := range x {
		result[i] = b.Forward(v)
	}
	return result
}

func (b *BinaryStep) Derivative(x float64) float64 {
	return 0 // Derivative is 0 everywhere except undefined at threshold
}

func (b *BinaryStep) Name() string { return "binary_step" }

// QuantizedReLU implements quantized ReLU for hardware deployment.
// Clips output to [0, max_val] and quantizes to n bits.
type QuantizedReLU struct {
	MaxVal float64
	Bits   int
}

func NewQuantizedReLU(maxVal float64, bits int) *QuantizedReLU {
	return &QuantizedReLU{MaxVal: maxVal, Bits: bits}
}

func (q *QuantizedReLU) Forward(x float64) float64 {
	if x <= 0 {
		return 0
	}
	if x >= q.MaxVal {
		return q.MaxVal
	}

	// Quantize
	levels := float64(int(1) << q.Bits)
	step := q.MaxVal / levels
	quantized := math.Round(x/step) * step
	return quantized
}

func (q *QuantizedReLU) ForwardVec(x []float64) []float64 {
	result := make([]float64, len(x))
	for i, v := range x {
		result[i] = q.Forward(v)
	}
	return result
}

func (q *QuantizedReLU) Derivative(x float64) float64 {
	if x > 0 && x < q.MaxVal {
		return 1
	}
	return 0
}

func (q *QuantizedReLU) Name() string { return "quantized_relu" }

// ActivationRegistry provides a registry of available activations.
var ActivationRegistry = map[string]func() Activation{
	"relu":           func() Activation { return NewReLU() },
	"leaky_relu":     func() Activation { return NewLeakyReLU(0.01) },
	"elu":            func() Activation { return NewELU(1.0) },
	"selu":           func() Activation { return NewSELU() },
	"sigmoid":        func() Activation { return NewSigmoid() },
	"tanh":           func() Activation { return NewTanh() },
	"hardtanh":       func() Activation { return NewHardTanh() },
	"hardsigmoid":    func() Activation { return NewHardSigmoidAct() },
	"hardswish":      func() Activation { return NewHardSwish() },
	"softplus":       func() Activation { return NewSoftplus() },
	"mish":           func() Activation { return NewMish() },
	"gelu":           func() Activation { return NewGELUAct() },
	"silu":           func() Activation { return NewSiLUAct() },
	"swish":          func() Activation { return NewSiLUAct() },
	"softmax":        func() Activation { return NewSoftmax() },
	"binary_step":    func() Activation { return NewBinaryStep() },
	"quantized_relu": func() Activation { return NewQuantizedReLU(6.0, 8) },
}

// GetActivation retrieves an activation function by name.
func GetActivation(name string) Activation {
	if factory, ok := ActivationRegistry[name]; ok {
		return factory()
	}
	return NewReLU() // Default
}

// ApplyActivation applies activation to 2D tensor.
func ApplyActivation(activation Activation, x [][]float64) [][]float64 {
	result := make([][]float64, len(x))
	for i := range result {
		result[i] = activation.ForwardVec(x[i])
	}
	return result
}

// ApplyActivation3D applies activation to 3D tensor (feature maps).
func ApplyActivation3D(activation Activation, x [][][]float64) [][][]float64 {
	result := make([][][]float64, len(x))
	for c := range result {
		result[c] = make([][]float64, len(x[c]))
		for h := range result[c] {
			result[c][h] = activation.ForwardVec(x[c][h])
		}
	}
	return result
}

// HardwareActivationConfig specifies activation for hardware deployment.
type HardwareActivationConfig struct {
	Type       string  // Activation type
	Quantize   bool    // Quantize output
	Bits       int     // Output bits if quantized
	ClampMin   float64 // Output clamp minimum
	ClampMax   float64 // Output clamp maximum
	LUTBased   bool    // Use lookup table
	LUTEntries int     // Number of LUT entries
}

// DefaultHardwareActivation returns hardware-friendly activation config.
func DefaultHardwareActivation() *HardwareActivationConfig {
	return &HardwareActivationConfig{
		Type:       "hardsigmoid",
		Quantize:   true,
		Bits:       8,
		ClampMin:   0,
		ClampMax:   1,
		LUTBased:   false,
		LUTEntries: 256,
	}
}

// GenerateLUT generates lookup table for activation function.
func GenerateLUT(activation Activation, minInput, maxInput float64, entries int) []float64 {
	lut := make([]float64, entries)
	step := (maxInput - minInput) / float64(entries-1)

	for i := 0; i < entries; i++ {
		x := minInput + float64(i)*step
		lut[i] = activation.Forward(x)
	}

	return lut
}

// LUTActivation applies activation using lookup table.
func LUTActivation(lut []float64, x float64, minInput, maxInput float64) float64 {
	entries := len(lut)
	if entries == 0 {
		return x
	}

	// Clamp input
	if x <= minInput {
		return lut[0]
	}
	if x >= maxInput {
		return lut[entries-1]
	}

	// Linear interpolation
	normalized := (x - minInput) / (maxInput - minInput)
	index := normalized * float64(entries-1)
	lowerIdx := int(index)
	upperIdx := lowerIdx + 1
	if upperIdx >= entries {
		upperIdx = entries - 1
	}

	frac := index - float64(lowerIdx)
	return lut[lowerIdx]*(1-frac) + lut[upperIdx]*frac
}
