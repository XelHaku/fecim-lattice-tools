//go:build !cgo

package gogpuapp

import (
	"fecim-lattice-tools/internal/gogpuapp/design"
	"fecim-lattice-tools/shared/viewmodel"
	hysteresisvm "fecim-lattice-tools/shared/viewmodel/hysteresis"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func buildHysteresisView(snapshot viewmodel.ModuleSnapshot, theme *material3.Theme) widget.Widget {
	return buildHysteresisViewWithActions(snapshot, theme, nil)
}

func buildHysteresisViewWithActions(snapshot viewmodel.ModuleSnapshot, theme *material3.Theme, onAction func(viewmodel.Action)) widget.Widget {
	children := []widget.Widget{
		primitives.Text(snapshot.Descriptor.Title).FontSize(22).Bold(),
		primitives.Text(snapshot.Descriptor.Description).FontSize(13).Color(theme.Colors.OnSurfaceVariant),
	}

	if snapshot.Descriptor.BoundaryNotice != "" {
		children = append(children, boundaryNotice(snapshot.Descriptor.BoundaryNotice))
	}

	metricBoxes := []widget.Widget{}
	for _, m := range snapshot.Metrics {
		unitStr := ""
		if m.Unit != "" {
			unitStr = " " + m.Unit
		}
		metricBoxes = append(metricBoxes, primitives.Box(
			primitives.Text(m.Label).FontSize(11).Color(theme.Colors.OnSurfaceVariant),
			primitives.Text(m.Value+unitStr).FontSize(16).Bold(),
		).
			Padding(12).
			Gap(4).
			Background(theme.Colors.SurfaceContainer).
			Rounded(6),
		)
	}
	children = append(children, primitives.Box(metricBoxes...).Gap(8))
	children = append(children, buildHysteresisControls(snapshot, theme, onAction))
	if panel := buildHysteresisDiagnosticPanels(snapshot, theme); panel != nil {
		children = append(children, panel)
	}

	for _, plot := range snapshot.Plots {
		plotData := design.NewPlotData(plot.Title, plot.XLabel, plot.YLabel)
		for _, s := range plot.Series {
			pts := make([]design.PlotPoint, len(s.Points))
			for i, p := range s.Points {
				pts[i] = design.PlotPoint{X: p.X, Y: p.Y}
			}
			plotData.AddSeries(s.Name, pts)
		}
		children = append(children, plotCard(plotData, theme))
	}

	eSections, rSections, dSections := partitionSections(snapshot.Sections)
	children = appendSectionGroup(children, "Education", eSections, widget.Hex(0xE8EEF0), theme)
	children = appendSectionGroup(children, "Research", rSections, widget.Hex(0xEBF5F0), theme)
	children = appendSectionGroup(children, "Design", dSections, widget.Hex(0xF5EEE8), theme)

	return primitives.Box(children...).
		Padding(24).
		Gap(14).
		Background(theme.Colors.Surface)
}

func buildHysteresisControls(snapshot viewmodel.ModuleSnapshot, theme *material3.Theme, onAction func(viewmodel.Action)) widget.Widget {
	actions := indexActions(snapshot.Actions)
	waveform := actionPayload(actions, hysteresisvm.EventSetWaveform, "waveform", hysteresisvm.WaveformSine)

	commandButtons := []widget.Widget{
		circuitButton("Run/Pause", false, actionOrDefault(actions, hysteresisvm.EventToggleSimulation, viewmodel.ActionToggle), theme, onAction),
		circuitButton("Export CSV", false, actionOrDefault(actions, hysteresisvm.EventExportCSV, viewmodel.ActionCommand), theme, onAction),
	}
	waveformButtons := []widget.Widget{
		circuitButton("Sine", waveform == hysteresisvm.WaveformSine, selectAction(hysteresisvm.EventSetWaveform, "waveform", hysteresisvm.WaveformSine), theme, onAction),
		circuitButton("Triangle", waveform == hysteresisvm.WaveformTriangle, selectAction(hysteresisvm.EventSetWaveform, "waveform", hysteresisvm.WaveformTriangle), theme, onAction),
		circuitButton("Square", waveform == hysteresisvm.WaveformSquare, selectAction(hysteresisvm.EventSetWaveform, "waveform", hysteresisvm.WaveformSquare), theme, onAction),
	}
	fieldButtons := []widget.Widget{
		circuitButton("+/-1500", false, fieldRangeAction("-1500", "1500"), theme, onAction),
		circuitButton("+/-3000", false, fieldRangeAction("-3000", "3000"), theme, onAction),
		circuitButton("+/-6000", false, fieldRangeAction("-6000", "6000"), theme, onAction),
	}
	diagnosticButtons := []widget.Widget{
		circuitButton("Run PUND", false, actionOrDefault(actions, hysteresisvm.EventRunPUND, viewmodel.ActionCommand), theme, onAction),
		circuitButton("Run FORC", false, actionOrDefault(actions, hysteresisvm.EventRunFORC, viewmodel.ActionCommand), theme, onAction),
	}

	return primitives.Box(
		primitives.Text("Hysteresis Controls").FontSize(14).Bold(),
		primitives.Box(commandButtons...).Gap(8),
		controlRow("Waveform", waveformButtons, theme),
		controlRow("Field", fieldButtons, theme),
		controlRow("Diagnostics", diagnosticButtons, theme),
	).Padding(12).Gap(8).Background(theme.Colors.SurfaceContainer).Rounded(6)
}

type hysteresisDiagnosticPanelState struct {
	pundAvailable bool
	pundSummary   string
	forcAvailable bool
	forcSummary   string
}

func hysteresisDiagnosticPanelStateFromSnapshot(snapshot viewmodel.ModuleSnapshot) hysteresisDiagnosticPanelState {
	var state hysteresisDiagnosticPanelState
	for _, section := range snapshot.Sections {
		switch section.ID {
		case "diagnostic_pund":
			state.pundAvailable = true
			state.pundSummary = section.Body
		case "diagnostic_forc":
			state.forcAvailable = true
			state.forcSummary = section.Body
		}
	}
	return state
}

func buildHysteresisDiagnosticPanels(snapshot viewmodel.ModuleSnapshot, theme *material3.Theme) widget.Widget {
	state := hysteresisDiagnosticPanelStateFromSnapshot(snapshot)
	if !state.pundAvailable && !state.forcAvailable {
		return nil
	}
	rows := []widget.Widget{
		primitives.Text("Diagnostic Summaries").FontSize(14).Bold(),
	}
	if state.pundAvailable {
		rows = append(rows, diagnosticSummaryCard("PUND Measurement", state.pundSummary, theme))
	}
	if state.forcAvailable {
		rows = append(rows, diagnosticSummaryCard("FORC Density", state.forcSummary, theme))
	}
	return primitives.Box(rows...).
		Padding(12).
		Gap(8).
		Background(widget.Hex(0xF6FAF7)).
		Rounded(8).
		BorderStyle(1, widget.Hex(0xD4DED8))
}

func diagnosticSummaryCard(title, body string, theme *material3.Theme) widget.Widget {
	return primitives.Box(
		primitives.Text(title).FontSize(13).Bold().Color(widget.Hex(0x183D34)),
		primitives.Text(body).FontSize(11).Color(theme.Colors.OnSurfaceVariant),
	).Padding(10).Gap(6).Background(theme.Colors.SurfaceContainer).Rounded(6)
}

func fieldRangeAction(min, max string) viewmodel.Action {
	return viewmodel.Action{
		ID:   hysteresisvm.EventSetFieldRange,
		Kind: viewmodel.ActionCommand,
		Payload: map[string]string{
			"min": min,
			"max": max,
		},
	}
}

func plotCard(data *design.PlotData, theme *material3.Theme) widget.Widget {
	summary := data.Title
	if len(data.Series) > 0 {
		summary += " | " + data.Series[0].Name
	}
	return primitives.Box(
		primitives.Text(data.Title).FontSize(16).Bold(),
		primitives.Text(data.XLabel+" vs "+data.YLabel).FontSize(12).Color(theme.Colors.OnSurfaceVariant),
	).
		Padding(20).
		Gap(8).
		Background(widget.Hex(0xF6FAF7)).
		Rounded(8).
		BorderStyle(1, widget.Hex(0xD4DED8))
}

func partitionSections(sections []viewmodel.Section) (edu, res, des []viewmodel.Section) {
	for _, s := range sections {
		switch s.Category {
		case "education":
			edu = append(edu, s)
		case "research":
			res = append(res, s)
		case "design":
			des = append(des, s)
		default:
			res = append(res, s)
		}
	}
	return
}

func appendSectionGroup(children []widget.Widget, label string, sections []viewmodel.Section, bg widget.Color, theme *material3.Theme) []widget.Widget {
	if len(sections) == 0 {
		return children
	}
	children = append(children, primitives.Text(label).FontSize(15).Bold().Color(widget.Hex(0x24483E)))
	for _, s := range sections {
		children = append(children, primitives.Box(
			primitives.Text(s.Title).FontSize(14).Bold().Color(widget.Hex(0x183D34)),
			primitives.Text(s.Body).FontSize(12).Color(widget.Hex(0x44504B)),
		).
			Padding(12).
			Gap(6).
			Background(bg).
			Rounded(8).
			BorderStyle(1, widget.Hex(0xD4DED8)),
		)
	}
	return children
}

func boundaryNotice(text string) widget.Widget {
	return primitives.Box(
		primitives.Text(text).FontSize(12).Color(widget.Hex(0x5C3B00)),
	).Padding(12).Background(widget.Hex(0xFFF4D8)).Rounded(8).BorderStyle(1, widget.Hex(0xE7C66A))
}
