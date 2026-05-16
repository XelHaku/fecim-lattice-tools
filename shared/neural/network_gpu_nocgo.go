//go:build !cgo

package neural

import "fmt"

func InitGPU() {}

func IsGPUAvailable() bool { return false }

func DestroyGPU() {}

func (net *DualModeNetwork) SetUseGPU(use bool) {
	net.mu.Lock()
	defer net.mu.Unlock()
	net.useGPU = false
}

func (net *DualModeNetwork) UseGPU() bool {
	net.mu.RLock()
	defer net.mu.RUnlock()
	return false
}

func (net *DualModeNetwork) forwardFPGPU(input []float64, weights [][]float64, bias []float64) ([]float64, error) {
	return nil, fmt.Errorf("GPU not available in zero-CGO builds")
}
