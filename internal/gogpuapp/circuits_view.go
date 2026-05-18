//go:build !cgo

package gogpuapp

import (
	"strconv"

	"fecim-lattice-tools/shared/viewmodel"
	circuitsvm "fecim-lattice-tools/shared/viewmodel/circuits"

	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func buildCircuitsView(snapshot viewmodel.ModuleSnapshot, theme *material3.Theme) widget.Widget {
	return buildCircuitsViewWithActions(snapshot, theme, nil)
}

func buildCircuitsViewWithActions(snapshot viewmodel.ModuleSnapshot, theme *material3.Theme, onAction func(viewmodel.Action)) widget.Widget {
	children := []widget.Widget{
		primitives.Text(snapshot.Descriptor.Title).FontSize(22).Bold(),
		primitives.Text(snapshot.Descriptor.Description).FontSize(13).Color(theme.Colors.OnSurfaceVariant),
	}
	if snapshot.Descriptor.BoundaryNotice != "" {
		children = append(children, boundaryNotice(snapshot.Descriptor.BoundaryNotice))
	}

	metricBoxes := []widget.Widget{}
	for _, m := range snapshot.Metrics {
		status := theme.Colors.Primary
		if m.ID == "ispp" && m.Value == "false" {
			status = theme.Colors.OnSurfaceVariant
		}
		metricBoxes = append(metricBoxes, primitives.Box(
			primitives.Text(m.Label).FontSize(11).Color(theme.Colors.OnSurfaceVariant),
			primitives.Text(m.Value).FontSize(14).Bold().Color(status),
		).Padding(12).Gap(4).Background(theme.Colors.SurfaceContainer).Rounded(6))
	}
	children = append(children, primitives.Box(metricBoxes...).Gap(8))
	children = append(children, buildCircuitControls(snapshot, theme, onAction))

	eSections, rSections, dSections := partitionSections(snapshot.Sections)
	children = appendSectionGroup(children, "Education", eSections, widget.Hex(0xE8EEF0), theme)
	children = appendSectionGroup(children, "Research", rSections, widget.Hex(0xEBF5F0), theme)
	children = appendSectionGroup(children, "Design", dSections, widget.Hex(0xF5EEE8), theme)

	return primitives.Box(children...).Padding(24).Gap(14).Background(theme.Colors.Surface)
}

func buildCircuitControls(snapshot viewmodel.ModuleSnapshot, theme *material3.Theme, onAction func(viewmodel.Action)) widget.Widget {
	actions := indexActions(snapshot.Actions)
	mode := actionPayload(actions, circuitsvm.ActionSetOperationMode, "mode", circuitsvm.OperationRead)
	timingOperation := actionPayload(actions, circuitsvm.ActionSetTimingOperation, "operation", "READ")
	architecture := actionPayload(actions, circuitsvm.ActionSetArchitecture, "architecture", circuitsvm.ArchitecturePassive)
	rows := actionPayload(actions, circuitsvm.ActionResizeArray, "rows", "8")
	cols := actionPayload(actions, circuitsvm.ActionResizeArray, "cols", rows)
	dacBits := actionPayload(actions, circuitsvm.ActionSetDACBits, "bits", "5")
	adcBits := actionPayload(actions, circuitsvm.ActionSetADCBits, "bits", "5")
	tiaGain := actionPayload(actions, circuitsvm.ActionSetTIAGain, "gain_ohm", "10000")
	coupling := actionPayload(actions, circuitsvm.ActionSetCouplingTier, "tier", circuitsvm.CouplingTierA)
	isppEngine := actionPayload(actions, circuitsvm.ActionSetISPPEngine, "engine", circuitsvm.ISPPEngineLevel)
	loggerVerbosity := actionPayload(actions, circuitsvm.ActionSetLoggerVerbosity, "verbosity", "off")
	isppEnabled := actionPayload(actions, circuitsvm.ActionToggleISPP, "enabled", "true") == "true"

	commandButtons := []widget.Widget{
		circuitButton("Simulate Read", false, actionOrDefault(actions, circuitsvm.ActionRunRead, viewmodel.ActionCommand), theme, onAction),
		circuitButton("Simulate Write", false, actionOrDefault(actions, circuitsvm.ActionRunWrite, viewmodel.ActionCommand), theme, onAction),
		circuitButton("Simulate Compute", false, actionOrDefault(actions, circuitsvm.ActionRunCompute, viewmodel.ActionCommand), theme, onAction),
		circuitButton("Export Log", false, actionOrDefault(actions, circuitsvm.ActionExportOperationLog, viewmodel.ActionCommand), theme, onAction),
		circuitButton("Export Specs", false, actionOrDefault(actions, circuitsvm.ActionExportReferenceSpecs, viewmodel.ActionCommand), theme, onAction),
		circuitButton("Export Timing", false, actionOrDefault(actions, circuitsvm.ActionExportReferenceTiming, viewmodel.ActionCommand), theme, onAction),
		circuitButton("Export Timing SVG", false, actionOrDefault(actions, circuitsvm.ActionExportReferenceTimingSVG, viewmodel.ActionCommand), theme, onAction),
		circuitButton("Animate Timing", false, actionOrDefault(actions, circuitsvm.ActionAnimateReferenceTiming, viewmodel.ActionCommand), theme, onAction),
	}

	modeButtons := []widget.Widget{
		circuitButton("READ", mode == circuitsvm.OperationRead, selectAction(circuitsvm.ActionSetOperationMode, "mode", circuitsvm.OperationRead), theme, onAction),
		circuitButton("WRITE", mode == circuitsvm.OperationWrite, selectAction(circuitsvm.ActionSetOperationMode, "mode", circuitsvm.OperationWrite), theme, onAction),
		circuitButton("COMPUTE", mode == circuitsvm.OperationCompute, selectAction(circuitsvm.ActionSetOperationMode, "mode", circuitsvm.OperationCompute), theme, onAction),
	}

	timingButtons := []widget.Widget{
		circuitButton("READ", timingOperation == "READ", selectAction(circuitsvm.ActionSetTimingOperation, "operation", "READ"), theme, onAction),
		circuitButton("WRITE", timingOperation == "WRITE", selectAction(circuitsvm.ActionSetTimingOperation, "operation", "WRITE"), theme, onAction),
		circuitButton("COMPUTE", timingOperation == "COMPUTE", selectAction(circuitsvm.ActionSetTimingOperation, "operation", "COMPUTE"), theme, onAction),
	}

	architectureButtons := []widget.Widget{
		circuitButton("0T1R", architecture == circuitsvm.ArchitecturePassive, selectAction(circuitsvm.ActionSetArchitecture, "architecture", circuitsvm.ArchitecturePassive), theme, onAction),
		circuitButton("1T1R", architecture == circuitsvm.Architecture1T1R, selectAction(circuitsvm.ActionSetArchitecture, "architecture", circuitsvm.Architecture1T1R), theme, onAction),
		circuitButton("2T1R", architecture == circuitsvm.Architecture2T1R, selectAction(circuitsvm.ActionSetArchitecture, "architecture", circuitsvm.Architecture2T1R), theme, onAction),
	}

	arrayButtons := make([]widget.Widget, 0, len(circuitsvm.ValidArraySizes))
	for _, size := range circuitsvm.ValidArraySizes {
		sizeText := strconv.Itoa(size)
		arrayButtons = append(arrayButtons, circuitButton(sizeText, rows == sizeText && cols == sizeText, viewmodel.Action{
			ID:   circuitsvm.ActionResizeArray,
			Kind: viewmodel.ActionSelect,
			Payload: map[string]string{
				"rows": sizeText,
				"cols": sizeText,
			},
		}, theme, onAction))
	}

	cellButtons := buildCellButtons(rows, cols, theme, onAction)
	targetButtons := []widget.Widget{
		circuitButton("L0", false, selectAction(circuitsvm.ActionSetWriteTarget, "level", "0"), theme, onAction),
		circuitButton("L15", false, selectAction(circuitsvm.ActionSetWriteTarget, "level", "15"), theme, onAction),
		circuitButton("L29", false, selectAction(circuitsvm.ActionSetWriteTarget, "level", "29"), theme, onAction),
	}
	dacButtons := bitButtons(circuitsvm.ActionSetDACBits, "bits", dacBits, []int{4, 5, 6, 7, 8}, theme, onAction)
	adcButtons := bitButtons(circuitsvm.ActionSetADCBits, "bits", adcBits, []int{5, 6, 7, 8}, theme, onAction)
	tiaButtons := []widget.Widget{
		circuitButton("10k", tiaGain == "10000", selectAction(circuitsvm.ActionSetTIAGain, "gain_ohm", "10000"), theme, onAction),
		circuitButton("50k", tiaGain == "50000", selectAction(circuitsvm.ActionSetTIAGain, "gain_ohm", "50000"), theme, onAction),
		circuitButton("100k", tiaGain == "100000", selectAction(circuitsvm.ActionSetTIAGain, "gain_ohm", "100000"), theme, onAction),
	}
	couplingButtons := []widget.Widget{
		circuitButton("Ideal", coupling == circuitsvm.CouplingIdeal, selectAction(circuitsvm.ActionSetCouplingTier, "tier", circuitsvm.CouplingIdeal), theme, onAction),
		circuitButton("Tier-A", coupling == circuitsvm.CouplingTierA, selectAction(circuitsvm.ActionSetCouplingTier, "tier", circuitsvm.CouplingTierA), theme, onAction),
		circuitButton("Tier-B", coupling == circuitsvm.CouplingTierB, selectAction(circuitsvm.ActionSetCouplingTier, "tier", circuitsvm.CouplingTierB), theme, onAction),
	}
	engineButtons := []widget.Widget{
		circuitButton("Preisach", isppEngine == circuitsvm.ISPPEngineLevel, selectAction(circuitsvm.ActionSetISPPEngine, "engine", circuitsvm.ISPPEngineLevel), theme, onAction),
		circuitButton("L-K ODE", isppEngine == circuitsvm.ISPPEngineLK, selectAction(circuitsvm.ActionSetISPPEngine, "engine", circuitsvm.ISPPEngineLK), theme, onAction),
	}
	loggerButtons := []widget.Widget{
		circuitButton("Off", loggerVerbosity == "off", selectAction(circuitsvm.ActionSetLoggerVerbosity, "verbosity", "off"), theme, onAction),
		circuitButton("Info", loggerVerbosity == "info", selectAction(circuitsvm.ActionSetLoggerVerbosity, "verbosity", "info"), theme, onAction),
		circuitButton("Debug", loggerVerbosity == "debug", selectAction(circuitsvm.ActionSetLoggerVerbosity, "verbosity", "debug"), theme, onAction),
		circuitButton("Trace", loggerVerbosity == "trace", selectAction(circuitsvm.ActionSetLoggerVerbosity, "verbosity", "trace"), theme, onAction),
	}
	isppToggle := checkbox.New(
		checkbox.Label("ISPP enabled"),
		checkbox.Checked(isppEnabled),
		checkbox.PainterOpt(material3.CheckboxPainter{Theme: theme}),
		checkbox.OnToggle(func(checked bool) {
			emitAction(onAction, viewmodel.Action{
				ID:      circuitsvm.ActionToggleISPP,
				Kind:    viewmodel.ActionToggle,
				Payload: map[string]string{"enabled": strconv.FormatBool(checked)},
			})
		}),
	)

	return primitives.Box(
		primitives.Text("Circuit Controls").FontSize(14).Bold(),
		primitives.Box(commandButtons...).Gap(8),
		controlRow("Mode", modeButtons, theme),
		controlRow("Timing", timingButtons, theme),
		controlRow("Architecture", architectureButtons, theme),
		controlRow("Array", arrayButtons, theme),
		controlRow("Cell", cellButtons, theme),
		controlRow("Write Target", targetButtons, theme),
		controlRow("DAC", dacButtons, theme),
		controlRow("ADC", adcButtons, theme),
		controlRow("TIA", tiaButtons, theme),
		controlRow("Coupling", couplingButtons, theme),
		controlRow("ISPP Engine", engineButtons, theme),
		controlRow("Logger", loggerButtons, theme),
		isppToggle,
	).Padding(12).Gap(8).Background(theme.Colors.SurfaceContainer).Rounded(6)
}

func circuitButton(label string, active bool, action viewmodel.Action, theme *material3.Theme, onAction func(viewmodel.Action)) widget.Widget {
	variant := button.Tonal
	if active {
		variant = button.Filled
	}
	return button.New(
		button.Text(label),
		button.VariantOpt(variant),
		button.SizeOpt(button.Small),
		button.PainterOpt(material3.ButtonPainter{Theme: theme}),
		button.OnClick(func() { emitAction(onAction, action) }),
	).MinWidth(64)
}

func controlRow(label string, controls []widget.Widget, theme *material3.Theme) widget.Widget {
	children := []widget.Widget{primitives.Text(label).FontSize(11).Color(theme.Colors.OnSurfaceVariant)}
	children = append(children, controls...)
	return primitives.Box(children...).Gap(8)
}

func buildCellButtons(rows, cols string, theme *material3.Theme, onAction func(viewmodel.Action)) []widget.Widget {
	rowCount := parsePositiveInt(rows, 8)
	colCount := parsePositiveInt(cols, rowCount)
	centerRow := (rowCount - 1) / 2
	centerCol := (colCount - 1) / 2
	lastRow := rowCount - 1
	lastCol := colCount - 1
	return []widget.Widget{
		cellButton("[0,0]", 0, 0, theme, onAction),
		cellButton("center", centerRow, centerCol, theme, onAction),
		cellButton("edge", lastRow, lastCol, theme, onAction),
	}
}

func cellButton(label string, row, col int, theme *material3.Theme, onAction func(viewmodel.Action)) widget.Widget {
	return circuitButton(label, false, viewmodel.Action{
		ID:   circuitsvm.ActionSelectCell,
		Kind: viewmodel.ActionSelect,
		Payload: map[string]string{
			"row": strconv.Itoa(row),
			"col": strconv.Itoa(col),
		},
	}, theme, onAction)
}

func bitButtons(actionID, payloadKey, active string, values []int, theme *material3.Theme, onAction func(viewmodel.Action)) []widget.Widget {
	buttons := make([]widget.Widget, 0, len(values))
	for _, value := range values {
		text := strconv.Itoa(value)
		buttons = append(buttons, circuitButton(text+"b", active == text, selectAction(actionID, payloadKey, text), theme, onAction))
	}
	return buttons
}

func indexActions(actions []viewmodel.Action) map[string]viewmodel.Action {
	indexed := make(map[string]viewmodel.Action, len(actions))
	for _, action := range actions {
		indexed[action.ID] = action
	}
	return indexed
}

func actionOrDefault(actions map[string]viewmodel.Action, id string, kind viewmodel.ActionKind) viewmodel.Action {
	if action, ok := actions[id]; ok {
		return action
	}
	return viewmodel.Action{ID: id, Kind: kind}
}

func actionPayload(actions map[string]viewmodel.Action, id, key, fallback string) string {
	action, ok := actions[id]
	if !ok || action.Payload == nil {
		return fallback
	}
	if value := action.Payload[key]; value != "" {
		return value
	}
	return fallback
}

func selectAction(id, key, value string) viewmodel.Action {
	return viewmodel.Action{
		ID:      id,
		Kind:    viewmodel.ActionSelect,
		Payload: map[string]string{key: value},
	}
}

func emitAction(onAction func(viewmodel.Action), action viewmodel.Action) {
	if onAction != nil {
		onAction(action)
	}
}

func parsePositiveInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
