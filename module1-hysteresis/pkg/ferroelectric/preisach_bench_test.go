package ferroelectric

import (
	"math"
	"testing"
)

func BenchmarkPreisachModelUpdate_SineExcitation(b *testing.B) {
	mat := DefaultHZO()
	m := NewPreisachModel(mat)

	// Precompute a periodic excitation to exercise the wipeout stack logic.
	const n = 1024 // power-of-two for cheap wrap
	fields := make([]float64, n)
	Emax := 2.0 * mat.Ec
	for i := 0; i < n; i++ {
		fields[i] = Emax * math.Sin(2*math.Pi*float64(i)/float64(n))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Update(fields[i&(n-1)])
	}
}

func BenchmarkPreisachModelPolarization_NoHistoryMutation(b *testing.B) {
	mat := DefaultHZO()
	m := NewPreisachModel(mat)
	_ = m.Update(-2.0 * mat.Ec)
	_ = m.Update(2.0 * mat.Ec)
	_ = m.Update(0.3 * mat.Ec)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Polarization()
	}
}
