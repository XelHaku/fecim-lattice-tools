//go:build legacy_fyne

// Package gui provides Fyne-based GUI components for architecture comparison.
package gui

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"

	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// ComparisonMode represents the current demo mode.
type ComparisonMode int

const (
	ComparisonModeIdle ComparisonMode = iota
	ComparisonModeCalculating
	ComparisonModeComparing
)

func (m ComparisonMode) String() string {
	switch m {
	case ComparisonModeIdle:
		return "IDLE"
	case ComparisonModeCalculating:
		return "CALCULATING"
	case ComparisonModeComparing:
		return "COMPARING"
	default:
		return "UNKNOWN"
	}
}

// PresentationMode represents the presentation/demo mode.
type PresentationMode int

const (
	PresentationModeManual   PresentationMode = iota // User controls navigation
	PresentationModeAuto                             // Self-running 30s per section
	PresentationModeBriefing                         // Large numbers, minimal jargon
	PresentationModeEngineer                         // Technical deep-dive
)

func (p PresentationMode) String() string {
	switch p {
	case PresentationModeManual:
		return "Manual"
	case PresentationModeAuto:
		return "Auto Demo"
	case PresentationModeBriefing:
		return "Briefing"
	case PresentationModeEngineer:
		return "Engineer"
	default:
		return "Unknown"
	}
}

// PresentationModeFromString converts string to PresentationMode.
func PresentationModeFromString(s string) PresentationMode {
	switch s {
	case "Manual":
		return PresentationModeManual
	case "Auto Demo":
		return PresentationModeAuto
	case "Briefing":
		return PresentationModeBriefing
	case "Engineer":
		return PresentationModeEngineer
	default:
		return PresentationModeManual
	}
}

// AutoDemoPhase represents phases in the auto demo sequence.
type AutoDemoPhase int

const (
	AutoDemoPhaseEnergyRace AutoDemoPhase = iota
	AutoDemoPhaseMarket
	AutoDemoPhaseCompetitive
	AutoDemoPhaseStrategy
	AutoDemoPhaseCalculator
	AutoDemoPhaseCount // Total number of phases
)

func (p AutoDemoPhase) String() string {
	switch p {
	case AutoDemoPhaseEnergyRace:
		return "Energy Comparison"
	case AutoDemoPhaseMarket:
		return "Market Opportunity"
	case AutoDemoPhaseCompetitive:
		return "Competitive Matrix"
	case AutoDemoPhaseStrategy:
		return "Phased Strategy"
	case AutoDemoPhaseCalculator:
		return "Calculator Demo"
	default:
		return "Unknown"
	}
}

// PhaseDuration returns the duration for each auto-demo phase.
func (p AutoDemoPhase) PhaseDuration() time.Duration {
	switch p {
	case AutoDemoPhaseEnergyRace:
		return 10 * time.Second
	case AutoDemoPhaseMarket:
		return 10 * time.Second
	case AutoDemoPhaseCompetitive:
		return 10 * time.Second
	case AutoDemoPhaseStrategy:
		return 10 * time.Second
	case AutoDemoPhaseCalculator:
		return 15 * time.Second
	default:
		return 10 * time.Second
	}
}

// ComparisonEducationalPanel shows explanations.
type ComparisonEducationalPanel struct {
	*sharedwidgets.EducationalPanel

	mu               sync.RWMutex
	presentationMode PresentationMode
	currentPhase     AutoDemoPhase
}

// NewComparisonEducationalPanel creates a new educational panel.
func NewComparisonEducationalPanel() *ComparisonEducationalPanel {
	e := &ComparisonEducationalPanel{
		EducationalPanel: sharedwidgets.NewEducationalPanel(sharedwidgets.EducationalPanelConfig{
			Title:   "Why CIM Wins",
			Content: "Compute-in-memory eliminates\nthe memory bottleneck.",
			MinSize: fyne.NewSize(200, 200),
		}),
		presentationMode: PresentationModeManual,
		currentPhase:     AutoDemoPhaseEnergyRace,
	}
	return e
}

// SetPresentationMode sets the current presentation mode.
func (e *ComparisonEducationalPanel) SetPresentationMode(mode PresentationMode) {
	e.mu.Lock()
	e.presentationMode = mode
	e.mu.Unlock()
	e.updateForMode()
}

// SetPhase sets the current auto-demo phase.
func (e *ComparisonEducationalPanel) SetPhase(phase AutoDemoPhase) {
	e.mu.Lock()
	e.currentPhase = phase
	e.mu.Unlock()
	e.updateForPhase()
}

// updateForMode updates content based on presentation mode.
func (e *ComparisonEducationalPanel) updateForMode() {
	e.mu.RLock()
	mode := e.presentationMode
	e.mu.RUnlock()

	switch mode {
	case PresentationModeBriefing:
		e.SetContent("Scenario Summary",
			"THE OPPORTUNITY\n\n"+
				"$721B addressable market by 2030 (model input)\n"+
				"1000× energy reduction (model input)\n"+
				"CMOS-compatible fabrication (assumption)\n"+
				"Literature context (not endorsement)\n\n"+
				"IMPLEMENTATION SCENARIO\n"+
				"Phase 1: NAND replacement\n"+
				"  → Drop-in compatible design\n"+
				"  → Minimal integration risk\n\n"+
				"TRL 4 → TRL 9 roadmap (scenario)")

	case PresentationModeEngineer:
		e.SetContent("Technical Deep-Dive",
			"FERROELECTRIC PHYSICS (MODEL INPUTS)\n\n"+
				"HfO₂-ZrO₂ superlattice structure (context)\n"+
				"Remanent polarization: 15-34 µC/cm² (model input)\n"+
				"Coercive field: 1.0-1.5 MV/cm (model input)\n"+
				"30 discrete analog levels (model input; simulation baseline)\n\n"+
				"CROSSBAR ARCHITECTURE\n"+
				"Matrix-vector multiply: O(1) time\n"+
				"Physical parallelism via Kirchhoff's law\n"+
				"Current summation: I = Σ(Gᵢⱼ × Vⱼ)\n\n"+
				"ENGINEERING CHALLENGES\n"+
				"IR voltage drop mitigation\n"+
				"Sneak path current management\n"+
				"Long-term conductance stability")

	default:
		e.SetContent("Why Compute-in-Memory Wins",
			"THE MEMORY WALL PROBLEM\n\n"+
				"Von Neumann Architecture:\n"+
				"  • Data shuttles between\n"+
				"    separate memory and CPU\n"+
				"  • Energy dominated by data movement\n"+
				"  • Bandwidth bottleneck limits performance\n\n"+
				"Compute-in-Memory Solution:\n"+
				"  • Computation occurs at storage location\n"+
				"  • Eliminates data movement overhead\n"+
				"  • Massive parallel operations via physics")
	}
}

// updateForPhase updates content based on auto-demo phase.
func (e *ComparisonEducationalPanel) updateForPhase() {
	e.mu.RLock()
	phase := e.currentPhase
	mode := e.presentationMode
	e.mu.RUnlock()

	var title, content string

	switch phase {
	case AutoDemoPhaseEnergyRace:
		title = "Energy Comparison"
		if mode == PresentationModeBriefing {
			content = "THE HEADLINE\n\n" +
				"1000× less energy (model input)\n" +
				"than current CPUs\n" +
				"100× less than GPUs (model input)\n\n" +
				"= 90% cost reduction (model input)\n" +
				"= 10× more inference (model input)\n" +
				"= same power budget (model input)\n\n" +
				"* TRL 4 = Laboratory Validation\n" +
				"  (not production ready)"
		} else {
			content = "ENERGY PER MAC\n\n" +
				"CPU + DRAM: 1000 pJ (model input)\n" +
				"GPU + HBM: 100 pJ (model input)\n" +
				"FeCIM: ~1 pJ (model input)*\n\n" +
				"GPU NUANCE:\n" +
				"  • GPU advantage applies to large, batched GEMM/MVM workloads\n" +
				"  • Single-cell / per-device physics simulation is not GPU-favorable\n\n" +
				"* TRL 4 = Laboratory Validation\n" +
				"  (not production ready)\n" +
				"(1 pJ = 1000 fJ)\n\n" +
				"Model input references (not validated)"
		}

	case AutoDemoPhaseMarket:
		title = "Market Opportunity"
		content = "$721B BY 2030 (model input)\n\n" +
			"NAND Flash: $98B (model input)\n" +
			"DRAM: $220B (model input)\n" +
			"AI Semiconductor: $403B (model input)\n\n" +
			"FeCIM addresses all three segments (scenario)"

	case AutoDemoPhaseCompetitive:
		title = "Competitive Position"
		content = "COMPETITIVE LANDSCAPE\n\n" +
			"Google TPU v5: Von Neumann arch\n" +
			"Intel Loihi 2: Exotic fabrication\n" +
			"IBM Analog AI: Research phase\n\n" +
			"MODEL INPUT ADVANTAGES:\n" +
			"  ✓ True compute-in-memory (assumption)\n" +
			"  ✓ Standard CMOS process (assumption)\n" +
			"  ✓ Production scalability (scenario)"

	case AutoDemoPhaseStrategy:
		title = "Phased Strategy"
		content = "COMMERCIALIZATION ROADMAP\n\n" +
			"Phase 1: NAND Replacement\n" +
			"  → Drop-in compatible interface\n" +
			"  → Leverage existing infrastructure\n\n" +
			"Phase 2: DRAM Displacement\n" +
			"  → Non-volatile, zero refresh power\n" +
			"  → Higher density potential\n\n" +
			"Phase 3: Full Compute-in-Memory\n" +
			"  → 80-90% model input energy savings\n" +
			"  → Transform datacenter economics"

	case AutoDemoPhaseCalculator:
		title = "Real Impact"
		content = "DATA CENTER SAVINGS\n\n" +
			"At 10,000 inferences/sec:\n\n" +
			"GPU: N/A (workload-dependent)\n" +
			"FeCIM: N/A (workload-dependent)\n\n" +
			"Try the calculator\n" +
			"with your workload!"
	}

	e.SetContent(title, content)
}

// SetComparison sets comparison explanation with calculated ratios.
func (e *ComparisonEducationalPanel) SetComparison(cpuRatio, gpuRatio float64) {
	content := "THE MEMORY WALL PROBLEM\n\n" +
		"Von Neumann Architecture:\n" +
		"  • Data shuttles between\n" +
		"    separate memory and CPU\n" +
		"  • Energy dominated by movement\n" +
		"  • Bandwidth bottleneck\n\n" +
		"Compute-in-Memory Solution:\n" +
		"  • Computation at storage location\n" +
		"  • Eliminates data movement\n" +
		"  • Physics-based parallelism\n\n" +
		"MODEL INPUT ADVANTAGES:\n" +
		fmt.Sprintf("  • %.0f× less power vs CPU*\n", cpuRatio) +
		fmt.Sprintf("  • %.0f× less power vs GPU*\n", gpuRatio) +
		"\nGPU NUANCE:\n" +
		"  • GPU wins on large batched array ops (GEMM/MVM)\n" +
		"  • Not applicable to single-cell simulation throughput\n" +
		"\n* TRL 4 = Laboratory Validation\n" +
		"  (not production ready)\n" +
		"  Model inputs only; not validated\n"
	e.SetContent("Why Compute-in-Memory Wins", content)
}

// newComparisonOperationLog creates an operation log for calculation history.
func newComparisonOperationLog() *sharedwidgets.OperationLog {
	return sharedwidgets.NewOperationLog(sharedwidgets.OperationLogConfig{
		Title:        "Calculation Log",
		MaxEntries:   8,
		MinSize:      fyne.NewSize(200, 150),
		UseMonospace: true,
	})
}
