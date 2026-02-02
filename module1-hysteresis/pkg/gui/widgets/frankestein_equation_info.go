// Package widgets provides custom GUI widgets for the hysteresis visualization.
package widgets

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/physics"
)

type termDetail struct {
	ID           string
	Title        string
	Equation     string
	Meaning      string
	Units        string
	DefaultValue string
	CodeRef      string
	References   []string
	Notes        string
}

type termDetailPanel struct {
	titleLabel   *widget.Label
	equation     *widget.Label
	meaning      *widget.Label
	units        *widget.Label
	defaultValue *widget.Label
	codeRef      *widget.Label
	references   *widget.Label
	notes        *widget.Label
}

func newTermDetailPanel() (*termDetailPanel, fyne.CanvasObject) {
	panel := &termDetailPanel{
		titleLabel:   widget.NewLabel("Tap a term to see its details."),
		equation:     widget.NewLabel(""),
		meaning:      widget.NewLabel(""),
		units:        widget.NewLabel(""),
		defaultValue: widget.NewLabel(""),
		codeRef:      widget.NewLabel(""),
		references:   widget.NewLabel(""),
		notes:        widget.NewLabel(""),
	}

	panel.titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	panel.equation.TextStyle = fyne.TextStyle{Monospace: true}

	panel.equation.Wrapping = fyne.TextWrapWord
	panel.meaning.Wrapping = fyne.TextWrapWord
	panel.units.Wrapping = fyne.TextWrapWord
	panel.defaultValue.Wrapping = fyne.TextWrapWord
	panel.codeRef.Wrapping = fyne.TextWrapWord
	panel.references.Wrapping = fyne.TextWrapWord
	panel.notes.Wrapping = fyne.TextWrapWord

	content := container.NewVBox(
		panel.titleLabel,
		fieldBlock("Equation", panel.equation),
		fieldBlock("Meaning", panel.meaning),
		fieldBlock("Units", panel.units),
		fieldBlock("Default / Typical", panel.defaultValue),
		fieldBlock("Code Mapping", panel.codeRef),
		fieldBlock("References", panel.references),
		fieldBlock("Notes", panel.notes),
	)

	card := widget.NewCard("Selected Term", "", content)
	return panel, card
}

func (p *termDetailPanel) SetDetail(termID, fallback string) {
	detail, ok := termDetails()[termID]
	if !ok {
		p.titleLabel.SetText("Selected Term")
		p.equation.SetText("")
		p.meaning.SetText(fallback)
		p.units.SetText("")
		p.defaultValue.SetText("")
		p.codeRef.SetText("")
		p.references.SetText("")
		p.notes.SetText("")
		return
	}

	p.titleLabel.SetText(detail.Title)
	p.equation.SetText(detail.Equation)
	p.meaning.SetText(detail.Meaning)
	p.units.SetText(detail.Units)
	p.defaultValue.SetText(detail.DefaultValue)
	p.codeRef.SetText(detail.CodeRef)
	p.references.SetText(bullets(detail.References))
	p.notes.SetText(detail.Notes)
}

func fieldBlock(title string, value *widget.Label) fyne.CanvasObject {
	label := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	return container.NewVBox(label, value)
}

func termDetails() map[string]termDetail {
	mat := physics.FeCIMMaterial()
	return map[string]termDetail{
		"rho_eff_main": {
			ID:           "rho_eff_main",
			Title:        "ρ_eff (effective viscosity)",
			Equation:     "ρ_eff dP/dt",
			Meaning:      "Effective damping that folds intrinsic viscosity and RC delay into a single term.",
			Units:        "Ω·m (effective viscosity term in the L-K equation).",
			DefaultValue: fmt.Sprintf("ρ_eff = ρ + (R_series·A/d) with ρ=%.2f Ω·m, R_series=%.0f Ω, A=%.0f nm², d=%.0f nm.", mat.RhoViscosity, mat.SeriesResistanceOhm, mat.Area*1e18, mat.Thickness*1e9),
			CodeRef:      "shared/physics/landau.go (UseEffectiveViscosity, effectiveRho)",
			References: []string{
				"Khalatnikov viscous dynamics (L-K form)",
				"RC delay aggregation (effective viscosity approximation)",
			},
		},
		"e_applied": {
			ID:           "e_applied",
			Title:        "E_applied (external field)",
			Equation:     "E_applied = V/d",
			Meaning:      "External drive field applied across the ferroelectric film thickness.",
			Units:        "V/m",
			DefaultValue: fmt.Sprintf("Ec ≈ %.2f MV/cm for FeCIM HZO defaults.", mat.Ec/1e8),
			CodeRef:      "shared/physics/landau.go (Step, dPdT)",
			References: []string{
				"Landau-Khalatnikov dynamics (drive term)",
			},
		},
		"k_dep": {
			ID:           "k_dep",
			Title:        "k_dep (depolarization factor)",
			Equation:     "E_dep = k_dep · P",
			Meaning:      "Depolarization field from interfacial layers; adds slant for analog multi-level states.",
			Units:        "V·m/C",
			DefaultValue: fmt.Sprintf("k_dep = %.2e V·m/C (Golden Set I).", mat.K_dep),
			CodeRef:      "shared/physics/landau.go (dPdT, E_depolarization)",
			References: []string{
				"Thin-film depolarization field models (dead-layer)",
			},
		},
		"alpha": {
			ID:           "alpha",
			Title:        "α (dynamic stiffness)",
			Equation:     "α(T,σ) = (T−T_C)/(2ε0C) − 2 Q12 σ",
			Meaning:      "Temperature + stress dependent curvature of the Landau free energy wells.",
			Units:        "Effective field contribution (V/m).",
			DefaultValue: fmt.Sprintf("Derived from T_C=%.0f K, C=%.1e K, Q12=%.3f m⁴/C², σ=%.1f GPa.", mat.CurieTemp, mat.CurieConst, mat.Q12, mat.StressGPa),
			CodeRef:      "shared/physics/landau.go (UpdateParams)",
			References: []string{
				"Curie-Weiss law (α temperature dependence)",
				"Electrostriction coupling (Q12 stress term)",
			},
		},
		"beta": {
			ID:           "beta",
			Title:        "β (first-order nonlinearity)",
			Equation:     "4βP³",
			Meaning:      "Negative β creates the first-order switching barrier in HZO.",
			Units:        "β in J·m⁵/C⁴ (contributes an effective field term).",
			DefaultValue: fmt.Sprintf("β = %.3e J·m⁵/C⁴ (Golden Set I).", mat.BetaLandau),
			CodeRef:      "shared/physics/landau.go (dPdT, Beta)",
			References: []string{
				"Landau-Devonshire free energy expansion",
			},
		},
		"gamma": {
			ID:           "gamma",
			Title:        "γ (sixth-order stabilizer)",
			Equation:     "6γP⁵",
			Meaning:      "Positive γ keeps the energy bounded at high polarization.",
			Units:        "γ in J·m⁹/C⁶ (contributes an effective field term).",
			DefaultValue: fmt.Sprintf("γ = %.3e J·m⁹/C⁶ (Golden Set I).", mat.GammaLandau),
			CodeRef:      "shared/physics/landau.go (dPdT, Gamma)",
			References: []string{
				"Landau-Devonshire free energy expansion",
			},
		},
		"lk_terms": {
			ID:           "lk_terms",
			Title:        "Landau-Khalatnikov nonlinearity",
			Equation:     "2αP + 4βP³ + 6γP⁵",
			Meaning:      "Nonlinear energy gradient that shapes the ferroelectric double-well potential.",
			Units:        "Effective field contribution (V/m).",
			DefaultValue: "Computed each step from α(T,σ), β, and γ.",
			CodeRef:      "shared/physics/landau.go (dPdT, UpdateParams)",
			References: []string{
				"Landau-Devonshire free energy expansion",
				"Khalatnikov viscous dynamics (L-K form)",
			},
		},
		"noise": {
			ID:           "noise",
			Title:        "ξ(t) (stochastic noise)",
			Equation:     "ξ(t)",
			Meaning:      "Optional Langevin noise for thermal variability and cycle-to-cycle spread.",
			Units:        "Effective field contribution (V/m).",
			DefaultValue: "Disabled by default (EnableNoise=false).",
			CodeRef:      "shared/physics/landau.go (noiseTerm)",
			References: []string{
				"Langevin dynamics for ferroelectric variability",
			},
		},
		"rho_eff_def": {
			ID:           "rho_eff_def",
			Title:        "ρ_eff definition",
			Equation:     "ρ_eff = ρ + (R_series·A/d)",
			Meaning:      "Aggregates series resistance delay into effective viscosity.",
			Units:        "Ω·m",
			DefaultValue: fmt.Sprintf("ρ=%.2f Ω·m, R_series=%.0f Ω, A=%.0f nm², d=%.0f nm.", mat.RhoViscosity, mat.SeriesResistanceOhm, mat.Area*1e18, mat.Thickness*1e9),
			CodeRef:      "shared/physics/landau.go (effectiveRho)",
			References: []string{
				"Khalatnikov viscous dynamics (L-K form)",
			},
		},
		"rho": {
			ID:           "rho",
			Title:        "ρ (viscosity / damping)",
			Equation:     "ρ",
			Meaning:      "Intrinsic damping in polarization dynamics.",
			Units:        "Ω·m",
			DefaultValue: fmt.Sprintf("ρ = %.2f Ω·m (Golden Set I).", mat.RhoViscosity),
			CodeRef:      "shared/physics/landau.go (Rho)",
			References: []string{
				"Khalatnikov viscous dynamics (L-K form)",
			},
		},
		"r_series": {
			ID:           "r_series",
			Title:        "R_series (series resistance)",
			Equation:     "R_series·A/d",
			Meaning:      "Series resistance term folded into effective viscosity.",
			Units:        "Ω",
			DefaultValue: fmt.Sprintf("R_series = %.0f Ω.", mat.SeriesResistanceOhm),
			CodeRef:      "shared/physics/landau.go (SeriesResistance)",
			References: []string{
				"RC delay aggregation (effective viscosity approximation)",
			},
		},
	}
}

func buildEquationInfoTabs() fyne.CanvasObject {
	tabs := container.NewAppTabs(
		container.NewTabItem("Overview", buildOverviewSection()),
		container.NewTabItem("α(T,σ)", buildAlphaSection()),
		container.NewTabItem("Parameters", buildGoldenSetSection()),
		container.NewTabItem("Materials", buildMaterialDefaultsSection()),
		container.NewTabItem("Dynamics", buildDynamicsSection()),
		container.NewTabItem("Assumptions", buildAssumptionsSection()),
		container.NewTabItem("References", buildReferencesSection()),
	)
	return tabs
}

func buildOverviewSection() fyne.CanvasObject {
	return container.NewVBox(
		sectionTitle("Master Equation (First-Order L-K / TDGL)"),
		bodyLabel("We solve polarization dynamics using the Landau–Khalatnikov form:"),
		equationBlock("ρ dP/dt = −∂G/∂P"),
		bodyLabel("With effective field and Landau energy expansion:"),
		equationBlock("E_eff = E_applied − k_dep P"),
		equationBlock("∂G/∂P = 2αP + 4βP³ + 6γP⁵"),
		bodyLabel("This widget adds depolarization and series-resistance aggregation used in the headless L-K path."),
	)
}

func buildAlphaSection() fyne.CanvasObject {
	return container.NewVBox(
		sectionTitle("Unified α(T,σ) Coefficient"),
		equationBlock("α(T,σ) = (T − T_C) / (2 ε0 C) − 2 Q12 σ"),
		bodyLabel("As temperature approaches T_C, α → 0 and the wells flatten (more volatile)."),
		bodyLabel("Stress shifts α via electrostriction; tensile vs compressive sign depends on Q12."),
	)
}

func buildGoldenSetSection() fyne.CanvasObject {
	rows := [][]string{
		{"Parameter", "Value", "Units", "Role"},
		{"β", "−2.160e8", "J·m⁵/C⁴", "First‑order barrier (negative)"},
		{"γ", "1.653e10", "J·m⁹/C⁶", "Stability (positive)"},
		{"ρ", "0.05", "Ω·m", "Viscosity / damping"},
		{"Q12", "−0.026", "m⁴/C²", "Electrostriction"},
		{"T_C", "723", "K", "Curie temperature"},
		{"k_dep", "2.5e8", "V·m/C", "Depolarization"},
	}
	return container.NewVBox(
		sectionTitle("Golden Parameter Set (10 nm HZO, Set I)"),
		tableFromRows(rows, []float32{120, 120, 140, 280}),
	)
}

func buildMaterialDefaultsSection() fyne.CanvasObject {
	mat := physics.FeCIMMaterial()
	rows := [][]string{
		{"Parameter", "Value"},
		{"Pr", fmt.Sprintf("%.2f C/m² (%.0f µC/cm²)", mat.Pr, mat.Pr*100)},
		{"Ps", fmt.Sprintf("%.2f C/m² (%.0f µC/cm²)", mat.Ps, mat.Ps*100)},
		{"Ec", fmt.Sprintf("%.2f MV/cm", mat.Ec/1e8)},
		{"Thickness", fmt.Sprintf("%.0f nm", mat.Thickness*1e9)},
		{"Area", fmt.Sprintf("%.0f nm²", mat.Area*1e18)},
		{"Tau (pulse width)", fmt.Sprintf("%.0f ns", mat.Tau*1e9)},
		{"BetaLandau", fmt.Sprintf("%.3e J·m⁵/C⁴", mat.BetaLandau)},
		{"GammaLandau", fmt.Sprintf("%.3e J·m⁹/C⁶", mat.GammaLandau)},
		{"RhoViscosity", fmt.Sprintf("%.2f Ω·m", mat.RhoViscosity)},
		{"SeriesResistance", fmt.Sprintf("%.0f Ω", mat.SeriesResistanceOhm)},
		{"K_dep", fmt.Sprintf("%.2e V·m/C", mat.K_dep)},
		{"Q12", fmt.Sprintf("%.3f m⁴/C²", mat.Q12)},
		{"Stress", fmt.Sprintf("%.1f GPa", mat.StressGPa)},
		{"CurieTemp", fmt.Sprintf("%.0f K", mat.CurieTemp)},
		{"CurieConst", fmt.Sprintf("%.1e K", mat.CurieConst)},
		{"Gmin/Gmax", fmt.Sprintf("%.0f / %.0f µS", mat.Gmin*1e6, mat.Gmax*1e6)},
	}

	return container.NewVBox(
		sectionTitle("FeCIM HZO Material Defaults"),
		bodyLabel("Defaults are pulled from shared/physics/material.go (FeCIMMaterial)."),
		tableFromRows(rows, []float32{180, 360}),
	)
}

func buildDynamicsSection() fyne.CanvasObject {
	return container.NewVBox(
		sectionTitle("Numerical Solver & Write Loop"),
		bodyLabel("RK4 integration is used to stabilize the stiff L-K dynamics at sub‑ns steps."),
		bodyLabel("Effective viscosity aggregates series resistance: ρ_eff = ρ + (R_series·A/d)."),
		bodyLabel("Headless mode runs adaptive binary ISPP with overshoot handling."),
		bodyLabel("Optional NLS (Merz’s law) modulates switching time at low fields."),
	)
}

func buildAssumptionsSection() fyne.CanvasObject {
	return container.NewVBox(
		sectionTitle("Model Assumptions"),
		bodyLabel(bullets([]string{
			"Single‑domain effective medium (no explicit multi‑domain walls).",
			"Depolarization modeled by k_dep term (interfacial layer approximation).",
			"Series resistance folded into ρ_eff for RC delay.",
			"Noise term optional; default disabled for deterministic checks.",
			"GUI hysteresis still uses Preisach; L‑K is exercised in headless mode.",
		})),
	)
}

func buildReferencesSection() fyne.CanvasObject {
	refs := []string{
		"Landau–Devonshire free energy expansion (phenomenological ferroelectric theory).",
		"Khalatnikov viscous dynamics (Landau–Khalatnikov equation).",
		"Curie–Weiss law for α(T) temperature dependence.",
		"Electrostriction coupling via Q12 under stress.",
		"Merz’s law for nucleation‑limited switching (NLS).",
		"Park et al., Adv. Mater. 2015 (HZO ferroelectricity).",
		"Müller et al., Nano Lett. 2012 (HfO₂ ferroelectric properties).",
		"Pesic et al., Adv. Funct. Mater. 2016 (wake‑up, fatigue).",
		"Cheema et al., Nature 2020 (superlattice enhancement).",
	}
	return container.NewVBox(
		sectionTitle("Literature & Canonical References"),
		bodyLabel(bullets(refs)),
	)
}

func sectionTitle(text string) *widget.Label {
	label := widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	label.Wrapping = fyne.TextWrapWord
	return label
}

func bodyLabel(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	return label
}

func equationBlock(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.TextStyle = fyne.TextStyle{Monospace: true}
	label.Wrapping = fyne.TextWrapWord
	return label
}

func tableFromRows(rows [][]string, colWidths []float32) fyne.CanvasObject {
	rowCount := len(rows)
	colCount := 0
	if rowCount > 0 {
		colCount = len(rows[0])
	}
	table := widget.NewTable(
		func() (int, int) {
			return rowCount, colCount
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.Wrapping = fyne.TextWrapWord
			return label
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)
			if id.Row >= rowCount || id.Col >= colCount {
				label.SetText("")
				return
			}
			label.SetText(rows[id.Row][id.Col])
			label.TextStyle = fyne.TextStyle{Bold: id.Row == 0}
		},
	)
	for col, width := range colWidths {
		if col < colCount {
			table.SetColumnWidth(col, width)
		}
	}
	return table
}

func bullets(items []string) string {
	if len(items) == 0 {
		return ""
	}
	var b strings.Builder
	for i, item := range items {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("• ")
		b.WriteString(item)
	}
	return b.String()
}
