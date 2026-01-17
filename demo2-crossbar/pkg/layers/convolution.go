// Package layers provides neural network layer implementations for crossbar arrays.
package layers

import (
	"fmt"
	"math"
	"math/rand"
)

// Conv2DConfig holds configuration for 2D convolution layer.
type Conv2DConfig struct {
	InChannels  int // Number of input channels
	OutChannels int // Number of output channels (filters)
	KernelSize  int // Kernel size (assumed square)
	Stride      int // Stride
	Padding     int // Zero padding
	UseBias     bool
}

// DefaultConv2DConfig returns default convolution config.
func DefaultConv2DConfig(inCh, outCh, kernel int) *Conv2DConfig {
	return &Conv2DConfig{
		InChannels:  inCh,
		OutChannels: outCh,
		KernelSize:  kernel,
		Stride:      1,
		Padding:     0,
		UseBias:     true,
	}
}

// Conv2D implements 2D convolution layer optimized for crossbar mapping.
type Conv2D struct {
	config *Conv2DConfig

	// Weights: [outChannels][inChannels][kernelH][kernelW]
	Weights [][][][]float64

	// Biases: [outChannels]
	Biases []float64

	// Im2col buffer for efficient crossbar mapping
	im2colBuffer [][]float64
}

// NewConv2D creates a new 2D convolution layer.
func NewConv2D(config *Conv2DConfig) *Conv2D {
	if config == nil {
		config = &Conv2DConfig{
			InChannels:  1,
			OutChannels: 1,
			KernelSize:  3,
			Stride:      1,
			Padding:     0,
			UseBias:     true,
		}
	}

	conv := &Conv2D{
		config:  config,
		Weights: make([][][][]float64, config.OutChannels),
		Biases:  make([]float64, config.OutChannels),
	}

	// Initialize weights with Kaiming/He initialization
	fanIn := config.InChannels * config.KernelSize * config.KernelSize
	stddev := math.Sqrt(2.0 / float64(fanIn))

	for oc := 0; oc < config.OutChannels; oc++ {
		conv.Weights[oc] = make([][][]float64, config.InChannels)
		for ic := 0; ic < config.InChannels; ic++ {
			conv.Weights[oc][ic] = make([][]float64, config.KernelSize)
			for kh := 0; kh < config.KernelSize; kh++ {
				conv.Weights[oc][ic][kh] = make([]float64, config.KernelSize)
				for kw := 0; kw < config.KernelSize; kw++ {
					conv.Weights[oc][ic][kh][kw] = rand.NormFloat64() * stddev
				}
			}
		}
	}

	return conv
}

// Forward performs convolution on input tensor.
// Input shape: [channels][height][width]
// Output shape: [outChannels][outHeight][outWidth]
func (c *Conv2D) Forward(input [][][]float64) [][][]float64 {
	inH := len(input[0])
	inW := len(input[0][0])

	outH := (inH+2*c.config.Padding-c.config.KernelSize)/c.config.Stride + 1
	outW := (inW+2*c.config.Padding-c.config.KernelSize)/c.config.Stride + 1

	// Apply padding if needed
	padded := input
	if c.config.Padding > 0 {
		padded = c.padInput(input)
	}

	// Initialize output
	output := make([][][]float64, c.config.OutChannels)
	for oc := range output {
		output[oc] = make([][]float64, outH)
		for oh := range output[oc] {
			output[oc][oh] = make([]float64, outW)
		}
	}

	// Perform convolution
	for oc := 0; oc < c.config.OutChannels; oc++ {
		for oh := 0; oh < outH; oh++ {
			for ow := 0; ow < outW; ow++ {
				sum := 0.0

				// Convolve
				for ic := 0; ic < c.config.InChannels; ic++ {
					for kh := 0; kh < c.config.KernelSize; kh++ {
						for kw := 0; kw < c.config.KernelSize; kw++ {
							ih := oh*c.config.Stride + kh
							iw := ow*c.config.Stride + kw
							sum += padded[ic][ih][iw] * c.Weights[oc][ic][kh][kw]
						}
					}
				}

				// Add bias
				if c.config.UseBias {
					sum += c.Biases[oc]
				}

				output[oc][oh][ow] = sum
			}
		}
	}

	return output
}

// padInput adds zero padding to input tensor.
func (c *Conv2D) padInput(input [][][]float64) [][][]float64 {
	p := c.config.Padding
	inC := len(input)
	inH := len(input[0])
	inW := len(input[0][0])

	padded := make([][][]float64, inC)
	for ic := range padded {
		padded[ic] = make([][]float64, inH+2*p)
		for h := range padded[ic] {
			padded[ic][h] = make([]float64, inW+2*p)
		}
		// Copy original data
		for h := 0; h < inH; h++ {
			for w := 0; w < inW; w++ {
				padded[ic][h+p][w+p] = input[ic][h][w]
			}
		}
	}
	return padded
}

// Im2Col converts input patches to columns for efficient crossbar mapping.
// Each column represents one convolution window, suitable for matrix multiplication.
// Returns: [patchSize][numPatches] where patchSize = inChannels * kernelH * kernelW
func (c *Conv2D) Im2Col(input [][][]float64) [][]float64 {
	inH := len(input[0])
	inW := len(input[0][0])

	outH := (inH+2*c.config.Padding-c.config.KernelSize)/c.config.Stride + 1
	outW := (inW+2*c.config.Padding-c.config.KernelSize)/c.config.Stride + 1

	patchSize := c.config.InChannels * c.config.KernelSize * c.config.KernelSize
	numPatches := outH * outW

	// Apply padding
	padded := input
	if c.config.Padding > 0 {
		padded = c.padInput(input)
	}

	// Create im2col matrix
	cols := make([][]float64, patchSize)
	for i := range cols {
		cols[i] = make([]float64, numPatches)
	}

	patchIdx := 0
	for oh := 0; oh < outH; oh++ {
		for ow := 0; ow < outW; ow++ {
			elemIdx := 0
			for ic := 0; ic < c.config.InChannels; ic++ {
				for kh := 0; kh < c.config.KernelSize; kh++ {
					for kw := 0; kw < c.config.KernelSize; kw++ {
						ih := oh*c.config.Stride + kh
						iw := ow*c.config.Stride + kw
						cols[elemIdx][patchIdx] = padded[ic][ih][iw]
						elemIdx++
					}
				}
			}
			patchIdx++
		}
	}

	return cols
}

// GetKernelMatrix returns weights as 2D matrix for crossbar programming.
// Shape: [outChannels][inChannels * kernelH * kernelW]
func (c *Conv2D) GetKernelMatrix() [][]float64 {
	patchSize := c.config.InChannels * c.config.KernelSize * c.config.KernelSize
	matrix := make([][]float64, c.config.OutChannels)

	for oc := range matrix {
		matrix[oc] = make([]float64, patchSize)
		idx := 0
		for ic := 0; ic < c.config.InChannels; ic++ {
			for kh := 0; kh < c.config.KernelSize; kh++ {
				for kw := 0; kw < c.config.KernelSize; kw++ {
					matrix[oc][idx] = c.Weights[oc][ic][kh][kw]
					idx++
				}
			}
		}
	}

	return matrix
}

// ForwardCrossbar performs convolution using crossbar-style matrix multiplication.
// This simulates how the operation would execute on actual CIM hardware.
func (c *Conv2D) ForwardCrossbar(input [][][]float64) [][][]float64 {
	inH := len(input[0])
	inW := len(input[0][0])

	outH := (inH+2*c.config.Padding-c.config.KernelSize)/c.config.Stride + 1
	outW := (inW+2*c.config.Padding-c.config.KernelSize)/c.config.Stride + 1

	// Convert to im2col format
	cols := c.Im2Col(input)

	// Get kernel matrix
	kernels := c.GetKernelMatrix()

	// Matrix multiplication: output = kernels × cols
	// Shape: [outChannels][numPatches]
	numPatches := outH * outW
	flatOutput := make([][]float64, c.config.OutChannels)

	for oc := 0; oc < c.config.OutChannels; oc++ {
		flatOutput[oc] = make([]float64, numPatches)
		for p := 0; p < numPatches; p++ {
			sum := 0.0
			for k := 0; k < len(kernels[oc]); k++ {
				sum += kernels[oc][k] * cols[k][p]
			}
			if c.config.UseBias {
				sum += c.Biases[oc]
			}
			flatOutput[oc][p] = sum
		}
	}

	// Reshape to output tensor
	output := make([][][]float64, c.config.OutChannels)
	for oc := range output {
		output[oc] = make([][]float64, outH)
		for oh := range output[oc] {
			output[oc][oh] = make([]float64, outW)
			for ow := range output[oc][oh] {
				output[oc][oh][ow] = flatOutput[oc][oh*outW+ow]
			}
		}
	}

	return output
}

// LoadWeights loads pre-trained weights.
func (c *Conv2D) LoadWeights(weights [][][][]float64, biases []float64) error {
	if len(weights) != c.config.OutChannels {
		return fmt.Errorf("output channel mismatch")
	}
	c.Weights = weights
	if len(biases) == c.config.OutChannels {
		c.Biases = biases
	}
	return nil
}

// GetCrossbarDimensions returns dimensions needed for crossbar mapping.
func (c *Conv2D) GetCrossbarDimensions() (rows, cols int) {
	rows = c.config.OutChannels
	cols = c.config.InChannels * c.config.KernelSize * c.config.KernelSize
	return rows, cols
}

// MaxPool2D implements 2D max pooling layer.
type MaxPool2D struct {
	KernelSize int
	Stride     int
}

// NewMaxPool2D creates a new max pooling layer.
func NewMaxPool2D(kernelSize, stride int) *MaxPool2D {
	if stride == 0 {
		stride = kernelSize
	}
	return &MaxPool2D{
		KernelSize: kernelSize,
		Stride:     stride,
	}
}

// Forward performs max pooling on input tensor.
func (p *MaxPool2D) Forward(input [][][]float64) [][][]float64 {
	inC := len(input)
	inH := len(input[0])
	inW := len(input[0][0])

	outH := (inH - p.KernelSize) / p.Stride + 1
	outW := (inW - p.KernelSize) / p.Stride + 1

	output := make([][][]float64, inC)
	for c := range output {
		output[c] = make([][]float64, outH)
		for oh := range output[c] {
			output[c][oh] = make([]float64, outW)
			for ow := range output[c][oh] {
				maxVal := math.Inf(-1)
				for kh := 0; kh < p.KernelSize; kh++ {
					for kw := 0; kw < p.KernelSize; kw++ {
						ih := oh*p.Stride + kh
						iw := ow*p.Stride + kw
						if input[c][ih][iw] > maxVal {
							maxVal = input[c][ih][iw]
						}
					}
				}
				output[c][oh][ow] = maxVal
			}
		}
	}

	return output
}

// AvgPool2D implements 2D average pooling layer.
type AvgPool2D struct {
	KernelSize int
	Stride     int
}

// NewAvgPool2D creates a new average pooling layer.
func NewAvgPool2D(kernelSize, stride int) *AvgPool2D {
	if stride == 0 {
		stride = kernelSize
	}
	return &AvgPool2D{
		KernelSize: kernelSize,
		Stride:     stride,
	}
}

// Forward performs average pooling on input tensor.
func (p *AvgPool2D) Forward(input [][][]float64) [][][]float64 {
	inC := len(input)
	inH := len(input[0])
	inW := len(input[0][0])

	outH := (inH - p.KernelSize) / p.Stride + 1
	outW := (inW - p.KernelSize) / p.Stride + 1
	poolSize := float64(p.KernelSize * p.KernelSize)

	output := make([][][]float64, inC)
	for c := range output {
		output[c] = make([][]float64, outH)
		for oh := range output[c] {
			output[c][oh] = make([]float64, outW)
			for ow := range output[c][oh] {
				sum := 0.0
				for kh := 0; kh < p.KernelSize; kh++ {
					for kw := 0; kw < p.KernelSize; kw++ {
						ih := oh*p.Stride + kh
						iw := ow*p.Stride + kw
						sum += input[c][ih][iw]
					}
				}
				output[c][oh][ow] = sum / poolSize
			}
		}
	}

	return output
}

// GlobalAvgPool2D implements global average pooling (used before FC layers).
type GlobalAvgPool2D struct{}

// Forward performs global average pooling.
func (p *GlobalAvgPool2D) Forward(input [][][]float64) []float64 {
	inC := len(input)
	output := make([]float64, inC)

	for c := 0; c < inC; c++ {
		sum := 0.0
		count := 0
		for h := range input[c] {
			for w := range input[c][h] {
				sum += input[c][h][w]
				count++
			}
		}
		output[c] = sum / float64(count)
	}

	return output
}

// GlobalMaxPool2D implements global max pooling.
type GlobalMaxPool2D struct{}

// Forward performs global max pooling.
func (p *GlobalMaxPool2D) Forward(input [][][]float64) []float64 {
	inC := len(input)
	output := make([]float64, inC)

	for c := 0; c < inC; c++ {
		maxVal := math.Inf(-1)
		for h := range input[c] {
			for w := range input[c][h] {
				if input[c][h][w] > maxVal {
					maxVal = input[c][h][w]
				}
			}
		}
		output[c] = maxVal
	}

	return output
}

// AdaptiveAvgPool2D implements adaptive average pooling to target output size.
// Commonly used before fully connected layers in modern architectures.
type AdaptiveAvgPool2D struct {
	OutputH int
	OutputW int
}

// NewAdaptiveAvgPool2D creates adaptive average pooling layer.
func NewAdaptiveAvgPool2D(outputH, outputW int) *AdaptiveAvgPool2D {
	return &AdaptiveAvgPool2D{
		OutputH: outputH,
		OutputW: outputW,
	}
}

// Forward performs adaptive average pooling.
func (p *AdaptiveAvgPool2D) Forward(input [][][]float64) [][][]float64 {
	inC := len(input)
	inH := len(input[0])
	inW := len(input[0][0])

	output := make([][][]float64, inC)

	for c := 0; c < inC; c++ {
		output[c] = make([][]float64, p.OutputH)
		for oh := 0; oh < p.OutputH; oh++ {
			output[c][oh] = make([]float64, p.OutputW)
			for ow := 0; ow < p.OutputW; ow++ {
				// Calculate input region for this output position
				hStart := oh * inH / p.OutputH
				hEnd := (oh + 1) * inH / p.OutputH
				wStart := ow * inW / p.OutputW
				wEnd := (ow + 1) * inW / p.OutputW

				// Average over the region
				sum := 0.0
				count := 0
				for h := hStart; h < hEnd; h++ {
					for w := wStart; w < wEnd; w++ {
						sum += input[c][h][w]
						count++
					}
				}
				if count > 0 {
					output[c][oh][ow] = sum / float64(count)
				}
			}
		}
	}

	return output
}

// AdaptiveMaxPool2D implements adaptive max pooling to target output size.
type AdaptiveMaxPool2D struct {
	OutputH int
	OutputW int
}

// NewAdaptiveMaxPool2D creates adaptive max pooling layer.
func NewAdaptiveMaxPool2D(outputH, outputW int) *AdaptiveMaxPool2D {
	return &AdaptiveMaxPool2D{
		OutputH: outputH,
		OutputW: outputW,
	}
}

// Forward performs adaptive max pooling.
func (p *AdaptiveMaxPool2D) Forward(input [][][]float64) [][][]float64 {
	inC := len(input)
	inH := len(input[0])
	inW := len(input[0][0])

	output := make([][][]float64, inC)

	for c := 0; c < inC; c++ {
		output[c] = make([][]float64, p.OutputH)
		for oh := 0; oh < p.OutputH; oh++ {
			output[c][oh] = make([]float64, p.OutputW)
			for ow := 0; ow < p.OutputW; ow++ {
				// Calculate input region for this output position
				hStart := oh * inH / p.OutputH
				hEnd := (oh + 1) * inH / p.OutputH
				wStart := ow * inW / p.OutputW
				wEnd := (ow + 1) * inW / p.OutputW

				// Max over the region
				maxVal := math.Inf(-1)
				for h := hStart; h < hEnd; h++ {
					for w := wStart; w < wEnd; w++ {
						if input[c][h][w] > maxVal {
							maxVal = input[c][h][w]
						}
					}
				}
				output[c][oh][ow] = maxVal
			}
		}
	}

	return output
}

// LpPool2D implements Lp-norm pooling (generalization of avg/max pooling).
// p=1: L1 pooling (average of absolutes)
// p=2: L2 pooling (root mean square)
// p->inf: approaches max pooling
type LpPool2D struct {
	KernelSize int
	Stride     int
	P          float64 // Norm order
}

// NewLpPool2D creates Lp-norm pooling layer.
func NewLpPool2D(kernelSize, stride int, p float64) *LpPool2D {
	if stride == 0 {
		stride = kernelSize
	}
	return &LpPool2D{
		KernelSize: kernelSize,
		Stride:     stride,
		P:          p,
	}
}

// Forward performs Lp-norm pooling.
func (p *LpPool2D) Forward(input [][][]float64) [][][]float64 {
	inC := len(input)
	inH := len(input[0])
	inW := len(input[0][0])

	outH := (inH - p.KernelSize) / p.Stride + 1
	outW := (inW - p.KernelSize) / p.Stride + 1
	poolSize := float64(p.KernelSize * p.KernelSize)

	output := make([][][]float64, inC)
	for c := range output {
		output[c] = make([][]float64, outH)
		for oh := range output[c] {
			output[c][oh] = make([]float64, outW)
			for ow := range output[c][oh] {
				sum := 0.0
				for kh := 0; kh < p.KernelSize; kh++ {
					for kw := 0; kw < p.KernelSize; kw++ {
						ih := oh*p.Stride + kh
						iw := ow*p.Stride + kw
						sum += math.Pow(math.Abs(input[c][ih][iw]), p.P)
					}
				}
				output[c][oh][ow] = math.Pow(sum/poolSize, 1.0/p.P)
			}
		}
	}

	return output
}

// SpatialPyramidPooling implements SPP for variable input sizes.
// Outputs fixed-length representation regardless of input size.
type SpatialPyramidPooling struct {
	Levels []int // Pyramid levels (e.g., [1, 2, 4] for 1x1, 2x2, 4x4)
}

// NewSpatialPyramidPooling creates SPP layer.
func NewSpatialPyramidPooling(levels []int) *SpatialPyramidPooling {
	return &SpatialPyramidPooling{Levels: levels}
}

// Forward performs spatial pyramid pooling.
func (s *SpatialPyramidPooling) Forward(input [][][]float64) []float64 {
	inC := len(input)
	inH := len(input[0])
	inW := len(input[0][0])

	var output []float64

	for _, level := range s.Levels {
		// Adaptive pooling for this level
		binH := inH / level
		binW := inW / level

		for c := 0; c < inC; c++ {
			for bh := 0; bh < level; bh++ {
				for bw := 0; bw < level; bw++ {
					// Max pool within bin
					hStart := bh * binH
					hEnd := hStart + binH
					if bh == level-1 {
						hEnd = inH // Handle remainder
					}
					wStart := bw * binW
					wEnd := wStart + binW
					if bw == level-1 {
						wEnd = inW
					}

					maxVal := math.Inf(-1)
					for h := hStart; h < hEnd; h++ {
						for w := wStart; w < wEnd; w++ {
							if input[c][h][w] > maxVal {
								maxVal = input[c][h][w]
							}
						}
					}
					output = append(output, maxVal)
				}
			}
		}
	}

	return output
}

// GetSPPOutputSize returns the output size for SPP given number of channels.
func (s *SpatialPyramidPooling) GetSPPOutputSize(inChannels int) int {
	total := 0
	for _, level := range s.Levels {
		total += level * level
	}
	return total * inChannels
}

// Flatten converts 3D tensor to 1D vector.
func Flatten(input [][][]float64) []float64 {
	var output []float64
	for c := range input {
		for h := range input[c] {
			output = append(output, input[c][h]...)
		}
	}
	return output
}

// ReLU3D applies ReLU activation to 3D tensor.
func ReLU3D(input [][][]float64) [][][]float64 {
	output := make([][][]float64, len(input))
	for c := range input {
		output[c] = make([][]float64, len(input[c]))
		for h := range input[c] {
			output[c][h] = make([]float64, len(input[c][h]))
			for w := range input[c][h] {
				output[c][h][w] = math.Max(0, input[c][h][w])
			}
		}
	}
	return output
}

// DepthwiseSeparableConv implements efficient depthwise separable convolution.
// Used in MobileNet-style architectures for efficient CIM mapping.
type DepthwiseSeparableConv struct {
	Depthwise *Conv2D // Depthwise convolution (groups = inChannels)
	Pointwise *Conv2D // 1x1 pointwise convolution
}

// NewDepthwiseSeparableConv creates a new depthwise separable convolution.
func NewDepthwiseSeparableConv(inChannels, outChannels, kernelSize int) *DepthwiseSeparableConv {
	// Depthwise: each input channel convolved separately
	dw := &Conv2D{
		config: &Conv2DConfig{
			InChannels:  inChannels,
			OutChannels: inChannels, // Same as input
			KernelSize:  kernelSize,
			Stride:      1,
			Padding:     kernelSize / 2,
			UseBias:     false,
		},
		Weights: make([][][][]float64, inChannels),
		Biases:  make([]float64, inChannels),
	}

	// Initialize depthwise weights (single channel per filter)
	for i := 0; i < inChannels; i++ {
		dw.Weights[i] = make([][][]float64, 1)
		dw.Weights[i][0] = make([][]float64, kernelSize)
		for kh := 0; kh < kernelSize; kh++ {
			dw.Weights[i][0][kh] = make([]float64, kernelSize)
			for kw := 0; kw < kernelSize; kw++ {
				dw.Weights[i][0][kh][kw] = rand.NormFloat64() * math.Sqrt(2.0/float64(kernelSize*kernelSize))
			}
		}
	}

	// Pointwise: 1x1 convolution to mix channels
	pw := NewConv2D(&Conv2DConfig{
		InChannels:  inChannels,
		OutChannels: outChannels,
		KernelSize:  1,
		Stride:      1,
		Padding:     0,
		UseBias:     true,
	})

	return &DepthwiseSeparableConv{
		Depthwise: dw,
		Pointwise: pw,
	}
}

// Forward performs depthwise separable convolution.
func (dsc *DepthwiseSeparableConv) Forward(input [][][]float64) [][][]float64 {
	// Depthwise: apply each filter to corresponding input channel
	dw := dsc.forwardDepthwise(input)
	// Pointwise: 1x1 conv to mix channels
	return dsc.Pointwise.Forward(dw)
}

func (dsc *DepthwiseSeparableConv) forwardDepthwise(input [][][]float64) [][][]float64 {
	inC := len(input)
	inH := len(input[0])
	inW := len(input[0][0])
	k := dsc.Depthwise.config.KernelSize
	p := dsc.Depthwise.config.Padding

	outH := inH + 2*p - k + 1
	outW := inW + 2*p - k + 1

	output := make([][][]float64, inC)
	for c := 0; c < inC; c++ {
		output[c] = make([][]float64, outH)
		for oh := range output[c] {
			output[c][oh] = make([]float64, outW)
			for ow := range output[c][oh] {
				sum := 0.0
				for kh := 0; kh < k; kh++ {
					for kw := 0; kw < k; kw++ {
						ih := oh + kh - p
						iw := ow + kw - p
						if ih >= 0 && ih < inH && iw >= 0 && iw < inW {
							sum += input[c][ih][iw] * dsc.Depthwise.Weights[c][0][kh][kw]
						}
					}
				}
				output[c][oh][ow] = sum
			}
		}
	}

	return output
}

// GetCrossbarRequirements returns crossbar dimensions for depthwise separable conv.
// Depthwise separable is more efficient for CIM: smaller crossbars needed.
func (dsc *DepthwiseSeparableConv) GetCrossbarRequirements() (dwRows, dwCols, pwRows, pwCols int) {
	k := dsc.Depthwise.config.KernelSize
	dwRows = dsc.Depthwise.config.InChannels
	dwCols = k * k // Per-channel, no cross-channel

	pwRows = dsc.Pointwise.config.OutChannels
	pwCols = dsc.Pointwise.config.InChannels // 1x1 kernel

	return
}
