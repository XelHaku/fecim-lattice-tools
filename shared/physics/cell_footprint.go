package physics

import "strings"

// TechNode is the process technology node in nm.
type TechNode float64

// CellFootprint captures architecture-level area and density estimates.
type CellFootprint struct {
	Architecture string
	F            float64 // feature size in meters
	FeFETArea    float64 // m^2
	SelectorArea float64 // m^2
	RoutingOH    float64 // m^2
	TotalArea    float64 // m^2
	Density      float64 // cells/mm^2
}

// CalculateFootprint estimates per-cell footprint from architecture and node.
func CalculateFootprint(arch string, techNodeNm float64) CellFootprint {
	a := strings.ToUpper(strings.TrimSpace(arch))
	if a == "" {
		a = "0T1R"
	}
	f := techNodeNm * 1e-9
	f2 := f * f

	var fefetF2, selectorF2, routingF2 float64
	switch a {
	case "0T1R":
		fefetF2 = 4
		selectorF2 = 0
		routingF2 = 0
	case "1T1R":
		// Typical midpoint in 6-12 F^2 range.
		fefetF2 = 4
		selectorF2 = 2
		routingF2 = 3 // total 9 F^2
	case "2T1R":
		// Typical midpoint in 12-20 F^2 range.
		fefetF2 = 4
		selectorF2 = 4
		routingF2 = 8 // total 16 F^2
	case "SRAM":
		// Typical midpoint in 120-150 F^2 range.
		fefetF2 = 0
		selectorF2 = 120
		routingF2 = 15 // total 135 F^2
	default:
		a = "0T1R"
		fefetF2 = 4
	}

	feArea := fefetF2 * f2
	selArea := selectorF2 * f2
	routingArea := routingF2 * f2
	total := feArea + selArea + routingArea

	density := 0.0
	if total > 0 {
		density = 1e-6 / total
	}

	return CellFootprint{
		Architecture: a,
		F:            f,
		FeFETArea:    feArea,
		SelectorArea: selArea,
		RoutingOH:    routingArea,
		TotalArea:    total,
		Density:      density,
	}
}

// NewCellFootprint is a compatibility wrapper.
func NewCellFootprint(arch string, tech TechNode, _ CellGeometry, _ MOSFETSelector) CellFootprint {
	return CalculateFootprint(arch, float64(tech))
}

// DensityCellsPerMM2 returns areal density in cells/mm^2.
func (c CellFootprint) DensityCellsPerMM2() float64 {
	if c.Density > 0 {
		return c.Density
	}
	if c.TotalArea <= 0 {
		return 0
	}
	return 1e-6 / c.TotalArea
}
