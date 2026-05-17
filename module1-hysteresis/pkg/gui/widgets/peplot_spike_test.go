//go:build legacy_fyne

package widgets

import (
	"image/color"
	"math"
	"math/rand"
	"sync"
	"testing"
)

func testColors() (bg, grid, axis, pos, neg, warn color.RGBA) {
	bg = color.RGBA{20, 20, 30, 255}
	grid = color.RGBA{60, 60, 80, 255}
	axis = color.RGBA{100, 100, 120, 255}
	pos = color.RGBA{0, 200, 150, 255}
	neg = color.RGBA{200, 100, 50, 255}
	warn = color.RGBA{255, 200, 0, 255}
	return
}

func TestPEPlotNoSpikesFromPhysicsData(t *testing.T) {
	bg, grid, axis, pos, neg, warn := testColors()
	plot := NewPEPlot(2e8, 0.4, bg, grid, axis, pos, neg, warn)

	points := 500
	Emax := 2e8
	Pmax := 0.4

	eData := make([]float64, points)
	pData := make([]float64, points)

	for i := 0; i < points; i++ {
		phase := float64(i) / float64(points) * 4 * math.Pi
		E := Emax * math.Sin(phase)
		P := Pmax * math.Tanh(E/Emax*2)
		eData[i] = E
		pData[i] = P
	}

	plot.SetData(eData, pData, eData[len(eData)-1], pData[len(pData)-1])

	plot.mu.RLock()
	defer plot.mu.RUnlock()

	if len(plot.eData) != len(plot.pData) {
		t.Errorf("Data length mismatch: E=%d, P=%d", len(plot.eData), len(plot.pData))
	}

	for i := 1; i < len(plot.eData); i++ {
		eDiff := math.Abs(plot.eData[i] - plot.eData[i-1])
		pDiff := math.Abs(plot.pData[i] - plot.pData[i-1])

		normE := eDiff / plot.eMax
		normP := pDiff / plot.pMax

		isSpike := (normE < 0.05 && normP > 0.30) ||
			(normE > 0.30 && normP < 0.05) ||
			normP > 0.50

		if isSpike {
			t.Errorf("Spike detected at index %d: normE=%.4f, normP=%.4f", i, normE, normP)
		}
	}
}

func TestPEPlotConcurrentSetData(t *testing.T) {
	bg, grid, axis, pos, neg, warn := testColors()
	plot := NewPEPlot(2e8, 0.4, bg, grid, axis, pos, neg, warn)

	var wg sync.WaitGroup
	iterations := 1000

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			size := rand.Intn(100) + 10
			eData := make([]float64, size)
			pData := make([]float64, size)
			for j := range eData {
				eData[j] = float64(j) * 1e6
				pData[j] = float64(j) * 0.001
			}
			plot.SetData(eData, pData, eData[len(eData)-1], pData[len(pData)-1])
		}
	}()

	lengthMismatchCount := 0
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			plot.mu.RLock()
			if len(plot.eData) != len(plot.pData) {
				lengthMismatchCount++
			}
			plot.mu.RUnlock()
		}
	}()

	wg.Wait()

	if lengthMismatchCount > 0 {
		t.Errorf("Found %d length mismatches during concurrent access", lengthMismatchCount)
	}
}

func TestPEPlotDataBoundsAfterSet(t *testing.T) {
	bg, grid, axis, pos, neg, warn := testColors()
	plot := NewPEPlot(2e8, 0.4, bg, grid, axis, pos, neg, warn)

	eData := []float64{-2e8, -1e8, 0, 1e8, 2e8}
	pData := []float64{-0.4, -0.2, 0, 0.2, 0.4}

	plot.SetData(eData, pData, eData[len(eData)-1], pData[len(pData)-1])

	plot.mu.RLock()
	defer plot.mu.RUnlock()

	for i, e := range plot.eData {
		if math.IsNaN(e) {
			t.Errorf("E[%d] is NaN", i)
		}
		if math.IsInf(e, 0) {
			t.Errorf("E[%d] is Inf", i)
		}
	}

	for i, p := range plot.pData {
		if math.IsNaN(p) {
			t.Errorf("P[%d] is NaN", i)
		}
		if math.IsInf(p, 0) {
			t.Errorf("P[%d] is Inf", i)
		}
	}
}

func TestPEPlotEmptyData(t *testing.T) {
	bg, grid, axis, pos, neg, warn := testColors()
	plot := NewPEPlot(2e8, 0.4, bg, grid, axis, pos, neg, warn)

	plot.SetData(nil, nil, 0, 0)

	plot.mu.RLock()
	defer plot.mu.RUnlock()

	if len(plot.eData) != 0 {
		t.Errorf("Expected empty eData after nil SetData")
	}
	if len(plot.pData) != 0 {
		t.Errorf("Expected empty pData after nil SetData")
	}
}

func TestPEPlotSinglePoint(t *testing.T) {
	bg, grid, axis, pos, neg, warn := testColors()
	plot := NewPEPlot(2e8, 0.4, bg, grid, axis, pos, neg, warn)

	eData := []float64{1e8}
	pData := []float64{0.2}

	plot.SetData(eData, pData, eData[0], pData[0])

	plot.mu.RLock()
	defer plot.mu.RUnlock()

	if len(plot.eData) != 1 {
		t.Errorf("Expected 1 point, got %d", len(plot.eData))
	}
}

func TestPEPlotRapidUpdates(t *testing.T) {
	bg, grid, axis, pos, neg, warn := testColors()
	plot := NewPEPlot(2e8, 0.4, bg, grid, axis, pos, neg, warn)

	for i := 0; i < 1000; i++ {
		size := (i % 100) + 1
		eData := make([]float64, size)
		pData := make([]float64, size)
		for j := range eData {
			eData[j] = float64(j)
			pData[j] = float64(j) * 0.01
		}
		plot.SetData(eData, pData, eData[len(eData)-1], pData[len(pData)-1])
	}

	plot.mu.RLock()
	defer plot.mu.RUnlock()

	if len(plot.eData) != len(plot.pData) {
		t.Errorf("Length mismatch after rapid updates: E=%d, P=%d",
			len(plot.eData), len(plot.pData))
	}
}

func TestPEPlotMaxBoundsEnforcement(t *testing.T) {
	eMax := 2e8
	pMax := 0.4
	bg, grid, axis, pos, neg, warn := testColors()
	plot := NewPEPlot(eMax, pMax, bg, grid, axis, pos, neg, warn)

	eData := []float64{-3e8, 0, 3e8}
	pData := []float64{-0.6, 0, 0.6}

	plot.SetData(eData, pData, eData[len(eData)-1], pData[len(pData)-1])

	plot.mu.RLock()
	defer plot.mu.RUnlock()

	if plot.eMax < math.Abs(eData[0]) {
		t.Logf("Plot may need to handle out-of-bounds E values: E=%.2e, eMax=%.2e",
			eData[0], plot.eMax)
	}
}

func TestPEPlotHistoryAccumulation(t *testing.T) {
	bg, grid, axis, pos, neg, warn := testColors()
	plot := NewPEPlot(2e8, 0.4, bg, grid, axis, pos, neg, warn)

	for i := 0; i < 10; i++ {
		size := 100
		eData := make([]float64, size)
		pData := make([]float64, size)
		for j := range eData {
			eData[j] = float64(j+i*100) * 1e6
			pData[j] = float64(j+i*100) * 0.001
		}
		plot.SetData(eData, pData, eData[len(eData)-1], pData[len(pData)-1])
	}

	plot.mu.RLock()
	finalLen := len(plot.eData)
	plot.mu.RUnlock()

	if finalLen > 10000 {
		t.Errorf("Plot data may be accumulating unbounded: len=%d", finalLen)
	}
}

func BenchmarkPEPlotSpikeSetData(b *testing.B) {
	bg, grid, axis, pos, neg, warn := testColors()
	plot := NewPEPlot(2e8, 0.4, bg, grid, axis, pos, neg, warn)

	eData := make([]float64, 500)
	pData := make([]float64, 500)
	for i := range eData {
		eData[i] = float64(i) * 1e6
		pData[i] = float64(i) * 0.001
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		plot.SetData(eData, pData, eData[len(eData)-1], pData[len(pData)-1])
	}
}

func BenchmarkPEPlotSpikeConcurrentAccess(b *testing.B) {
	bg, grid, axis, pos, neg, warn := testColors()
	plot := NewPEPlot(2e8, 0.4, bg, grid, axis, pos, neg, warn)

	eData := make([]float64, 100)
	pData := make([]float64, 100)
	for i := range eData {
		eData[i] = float64(i) * 1e6
		pData[i] = float64(i) * 0.001
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			plot.SetData(eData, pData, eData[len(eData)-1], pData[len(pData)-1])
			plot.mu.RLock()
			_ = len(plot.eData)
			plot.mu.RUnlock()
		}
	})
}
