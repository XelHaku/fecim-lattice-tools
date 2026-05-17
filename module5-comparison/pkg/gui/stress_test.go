//go:build legacy_fyne

package gui

import (
	"runtime"
	"testing"
)

func TestModule5RapidControlsAndTabSwitch_MemoryBudget(t *testing.T) {
	switches := 1000
	if testing.Short() {
		switches = 150
	}

	ca := NewComparisonApp()
	ca.calculator = NewDataCenterCalculator()
	ca.dcTransformation = NewDataCenterTransformation()
	workloads := []string{"MNIST", "ResNet-50", "BERT-Base", "GPT-2", "LLM-70B"}

	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)

	for i := 0; i < switches; i++ {
		ca.currentWorkload = workloads[i%len(workloads)]
		ca.currentInferences = float64(1000 + (i*37)%50000)
		ca.updateCalculations()
		_ = competitors[i%len(competitors)] // tab-switch surrogate access pattern
	}

	runtime.GC()
	runtime.ReadMemStats(&after)
	growth := int64(after.Alloc) - int64(before.Alloc)
	const budget = int64(120 * 1024 * 1024)
	if growth > budget {
		t.Fatalf("memory growth exceeded budget: growth=%d bytes budget=%d bytes", growth, budget)
	}
}
