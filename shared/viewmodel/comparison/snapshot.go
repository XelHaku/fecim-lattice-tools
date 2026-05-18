package comparison

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	pkg "fecim-lattice-tools/module5-comparison/pkg/comparison"
	"fecim-lattice-tools/shared/viewmodel"
)

func descriptor() viewmodel.ModuleDescriptor {
	return viewmodel.ModuleDescriptor{
		ID:             viewmodel.ModuleComparison,
		Title:          "FeCIM Comparison",
		Description:    "Evidence-first technology comparison and scenario analysis.",
		Status:         viewmodel.StatusFunctional,
		BoundaryNotice: "ESTIMATED DATA — Comparison benchmarks are educational estimates based on published architecture analyses. Not validated against production silicon measurements.",
	}
}

// buildSnapshot converts a slice of architectures into a UI-neutral
// ModuleSnapshot. Pure: same input → same output, no clock, no I/O.
func buildSnapshot(archs []*pkg.Architecture) viewmodel.ModuleSnapshot {
	sections := make([]viewmodel.Section, 0, len(archs))
	for _, a := range archs {
		if a == nil {
			continue
		}
		sections = append(sections, viewmodel.Section{
			ID:       sectionID(a.Name),
			Title:    a.Name,
			Body:     architectureBody(a),
			Category: "comparison",
		})
	}
	metrics := []viewmodel.Metric{
		{
			ID:         "count",
			Label:      "Architectures compared",
			Value:      strconv.Itoa(len(sections)),
			Confidence: "deterministic",
		},
	}
	return viewmodel.ModuleSnapshot{
		Descriptor: descriptor(),
		Metrics:    metrics,
		Sections:   sections,
		UpdatedAt:  time.Time{},
	}
}

func sectionID(name string) string {
	id := strings.ToLower(name)
	id = strings.ReplaceAll(id, " ", "-")
	id = strings.ReplaceAll(id, "+", "-")
	return id
}

func architectureBody(a *pkg.Architecture) string {
	estimated := ""
	if a.IsEstimated {
		estimated = " (estimated; not validated)"
	}
	memory := fmt.Sprintf("Memory: %.0f GB @ %.0f GB/s", a.MemorySize, a.MemoryBW)
	if a.MemoryBW == 0 {
		memory = fmt.Sprintf("Memory: %.0f GB (in-situ; compute-in-memory)", a.MemorySize)
	}
	return fmt.Sprintf(
		"Technology: %s%s\nProcess node: %.0f nm\nChip area: %.0f mm²\nTDP: %.1f W\nPeak TOPS: %.2f\nTOPS/W: %.3f\n%s",
		a.Technology, estimated,
		a.ProcessNode,
		a.ChipArea,
		a.TDP,
		a.PeakTOPS,
		a.TOPSPerWatt,
		memory,
	)
}
