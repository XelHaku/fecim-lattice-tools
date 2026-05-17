//go:build legacy_fyne

package gui

import (
	"sort"
	"testing"
	"time"
)

func percentile(samples []float64, p float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	cp := append([]float64(nil), samples...)
	sort.Float64s(cp)
	idx := int(float64(len(cp)-1) * p)
	if idx < 0 {
		idx = 0
	}
	if idx >= len(cp) {
		idx = len(cp) - 1
	}
	return cp[idx]
}

func BenchmarkModule5_TabSwitch(b *testing.B) {
	comps := buildCompetitors()
	durations := make([]float64, 0, b.N)
	for i := 0; i < b.N; i++ {
		start := time.Now()
		_ = comps[(i+1)%len(comps)]
		durations = append(durations, float64(time.Since(start).Microseconds()))
	}
	b.ReportMetric(percentile(durations, 0.50), "tabswitch_p50_us")
	b.ReportMetric(percentile(durations, 0.95), "tabswitch_p95_us")
	b.ReportMetric(percentile(durations, 0.99), "tabswitch_p99_us")
}

func BenchmarkModule5_Resize(b *testing.B) {
	chart := NewMarketOpportunityChart()
	durations := make([]float64, 0, b.N)
	for i := 0; i < b.N; i++ {
		start := time.Now()
		chart.UpdateAnimation(0.016)
		durations = append(durations, float64(time.Since(start).Microseconds()))
	}
	b.ReportMetric(percentile(durations, 0.50), "resize_p50_us")
	b.ReportMetric(percentile(durations, 0.95), "resize_p95_us")
	b.ReportMetric(percentile(durations, 0.99), "resize_p99_us")
}

func BenchmarkModule5_CalculateToRender(b *testing.B) {
	ca := NewComparisonApp()
	ca.calculator = NewDataCenterCalculator()
	ca.dcTransformation = NewDataCenterTransformation()
	durations := make([]float64, 0, b.N)
	for i := 0; i < b.N; i++ {
		start := time.Now()
		ca.currentWorkload = "GPT-2"
		ca.currentInferences = float64(1000 + (i % 50000))
		ca.updateCalculations()
		durations = append(durations, float64(time.Since(start).Microseconds()))
	}
	b.ReportMetric(percentile(durations, 0.50), "calc_render_p50_us")
	b.ReportMetric(percentile(durations, 0.95), "calc_render_p95_us")
	b.ReportMetric(percentile(durations, 0.99), "calc_render_p99_us")
}
