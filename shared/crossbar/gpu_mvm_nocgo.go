//go:build !cgo && !js

package crossbar

import "fmt"

type GPUAccelerator struct{}

type CrossbarParams struct {
	Rows           int32
	Cols           int32
	NoiseLevel     float32
	ADCBits        int32
	DACBits        int32
	Time           float32
	WireResistance float32
	DriftCoeff     float32
}

func NewGPUAccelerator(maxRows, maxCols int) (*GPUAccelerator, error) {
	if maxRows <= 0 || maxCols <= 0 {
		return nil, fmt.Errorf("invalid dimensions: rows=%d, cols=%d", maxRows, maxCols)
	}
	return nil, nil
}

func (g *GPUAccelerator) IsAvailable() bool { return false }

func (g *GPUAccelerator) MVM(_ []float32, _ []float32, _ CrossbarParams) ([]float32, error) {
	return nil, fmt.Errorf("GPU compute is unavailable in zero-CGO builds")
}

func (g *GPUAccelerator) SetDeviceVariation(_ []float32) error {
	return fmt.Errorf("GPU compute is unavailable in zero-CGO builds")
}

func (g *GPUAccelerator) Destroy() {}
