// Package gpu provides GPU-accelerated neural network operations using Vulkan compute shaders.
package gpu

import (
	"encoding/binary"
	"math"
)

// ActivationType defines the activation function to apply after a layer.
type ActivationType int32

const (
	ActivationNone    ActivationType = 0
	ActivationReLU    ActivationType = 1
	ActivationSoftmax ActivationType = 2
)

// LayerParams matches the shader's uniform buffer layout (std140).
// Must match dense_mvm.comp binding 0 EXACTLY.
//
// std140 layout: all int32/float32 are 4-byte aligned, 4-byte size.
type LayerParams struct {
	Rows       int32 // Output features (offset 0)
	Cols       int32 // Input features (offset 4)
	Activation int32 // 0=none, 1=relu (offset 8)
	Padding    int32 // std140 alignment (offset 12)
}

// LayerParamsSize is the byte size of LayerParams (16 bytes).
const LayerParamsSize = 16

// SoftmaxParams matches the shader's uniform buffer layout (std140).
// Must match softmax.comp binding 0 EXACTLY.
type SoftmaxParams struct {
	Size     int32 // Vector length (offset 0)
	Padding1 int32 // std140 alignment (offset 4)
	Padding2 int32 // std140 alignment (offset 8)
	Padding3 int32 // std140 alignment (offset 12)
}

// SoftmaxParamsSize is the byte size of SoftmaxParams (16 bytes).
const SoftmaxParamsSize = 16

// EncodeLayerParams encodes LayerParams into a byte slice for GPU upload.
func EncodeLayerParams(p LayerParams) []byte {
	buf := make([]byte, LayerParamsSize)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(p.Rows))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(p.Cols))
	binary.LittleEndian.PutUint32(buf[8:12], uint32(p.Activation))
	binary.LittleEndian.PutUint32(buf[12:16], uint32(p.Padding))
	return buf
}

// EncodeSoftmaxParams encodes SoftmaxParams into a byte slice for GPU upload.
func EncodeSoftmaxParams(p SoftmaxParams) []byte {
	buf := make([]byte, SoftmaxParamsSize)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(p.Size))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(p.Padding1))
	binary.LittleEndian.PutUint32(buf[8:12], uint32(p.Padding2))
	binary.LittleEndian.PutUint32(buf[12:16], uint32(p.Padding3))
	return buf
}

// LayerWeights holds weight data for a single dense layer.
type LayerWeights struct {
	Weights    []float32      // Row-major [Rows x Cols]
	Bias       []float32      // [Rows]
	Rows       int            // Output features
	Cols       int            // Input features
	Activation ActivationType // Activation function
}

// Float64ToFloat32 converts a float64 slice to float32.
func Float64ToFloat32(input []float64) []float32 {
	output := make([]float32, len(input))
	for i, v := range input {
		output[i] = float32(v)
	}
	return output
}

// Float32ToFloat64 converts a float32 slice to float64.
func Float32ToFloat64(input []float32) []float64 {
	output := make([]float64, len(input))
	for i, v := range input {
		output[i] = float64(v)
	}
	return output
}

// Weights2DToFlat converts a 2D weight matrix [rows][cols] to flat row-major []float32.
func Weights2DToFlat(weights [][]float64) []float32 {
	if len(weights) == 0 {
		return nil
	}
	rows := len(weights)
	cols := len(weights[0])
	flat := make([]float32, rows*cols)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			flat[i*cols+j] = float32(weights[i][j])
		}
	}
	return flat
}

// Compile-time size assertions
var _ = [1]struct{}{}[LayerParamsSize-16]   // LayerParams must be 16 bytes
var _ = [1]struct{}{}[SoftmaxParamsSize-16] // SoftmaxParams must be 16 bytes

// Unused import guard
var _ = math.Float32bits
