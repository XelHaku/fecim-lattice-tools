//go:build !cgo

package gpu

import "fmt"

type GPUNetwork struct{}

func NewGPUNetwork() *GPUNetwork { return &GPUNetwork{} }

func (g *GPUNetwork) IsAvailable() bool { return false }

func (g *GPUNetwork) Forward(input []float32, layers []LayerWeights) ([]float32, error) {
	return nil, fmt.Errorf("GPU not available in zero-CGO builds")
}

func (g *GPUNetwork) ForwardBatch(inputs [][]float32, layers []LayerWeights) ([][]float32, error) {
	return nil, fmt.Errorf("GPU not available in zero-CGO builds")
}

func (g *GPUNetwork) Destroy() {}

type DenseLayerGPU struct{}

func NewDenseLayerGPU(ctx any, maxRows, maxCols int) (*DenseLayerGPU, error) {
	return nil, fmt.Errorf("GPU not available in zero-CGO builds")
}

func (d *DenseLayerGPU) IsAvailable() bool { return false }

func (d *DenseLayerGPU) Forward(weights, bias, input []float32, rows, cols int, activation ActivationType) ([]float32, error) {
	return nil, fmt.Errorf("GPU not available in zero-CGO builds")
}

func (d *DenseLayerGPU) ForwardBatch(weights, bias, batchInput []float32, batchSize, rows, cols int, activation ActivationType) ([]float32, error) {
	return nil, fmt.Errorf("GPU not available in zero-CGO builds")
}

func (d *DenseLayerGPU) Destroy() {}
