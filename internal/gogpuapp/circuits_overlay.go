//go:build !cgo

package gogpuapp

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"fecim-lattice-tools/shared/viewmodel"

	"github.com/gogpu/gg"
)

type circuitsOverlayState struct {
	rows               int
	cols               int
	mode               string
	architecture       string
	selectedRow        int
	selectedCol        int
	writeTarget        int
	coupling           string
	isppEngine         string
	lastOperation      string
	halfSelectState    string
	halfSelectCells    int
	disturbVoltage     string
	stressBudget       string
	stressPerPulse     string
	pvtTempSweep       string
	pvtProcessYield    string
	pvtCornerENOB      string
	pvtNoiseCeiling    string
	specPowerLatency   string
	specThroughput     string
	specCompliance     string
	timingActive       string
	timingActivePhases string
}

func drawCircuitsOverlay(cc *gg.Context, snapshot viewmodel.ModuleSnapshot, w, h int) {
	if w < 520 || h < 360 {
		return
	}

	state := circuitsOverlayStateFromSnapshot(snapshot)
	panelX := 260.0
	panelY := 96.0
	panelW := minFloat(760, float64(w)-panelX-42)
	panelH := minFloat(560, float64(h)-panelY-54)
	if panelW < 320 || panelH < 260 {
		return
	}

	gridSize := minFloat(panelH-132, panelW*0.56)
	gridSize = clampFloat(gridSize, 160, 430)
	gridX := 26.0
	gridY := 76.0
	detailX := gridX + gridSize + 28
	detailW := panelW - detailX - 24
	if detailW < 150 {
		detailW = 150
	}

	cc.Push()
	cc.Translate(panelX, panelY)
	drawCircuitsPanelBackground(cc, panelW, panelH, state)
	drawCircuitsGrid(cc, state, gridX, gridY, gridSize)
	drawHalfSelectStressOverlay(cc, state, gridX, gridY, gridSize)
	drawCircuitPathState(cc, state, gridX, gridY, gridSize, detailX, panelH)
	drawCircuitsDetails(cc, state, detailX, 76, detailW, panelH-106)
	cc.Pop()
}

func circuitsOverlayStateFromSnapshot(snapshot viewmodel.ModuleSnapshot) circuitsOverlayState {
	metrics := map[string]string{}
	for _, metric := range snapshot.Metrics {
		metrics[metric.ID] = metric.Value
	}
	rows, cols := parseCircuitArray(metrics["array"])
	selectedRow, selectedCol := parseCircuitCell(metrics["selected_cell"])
	target := parseCircuitTarget(metrics["write_target"])
	lastOperation := metrics["last_operation"]
	if lastOperation == "" {
		lastOperation = "Ready"
	}
	mode := strings.ToUpper(metrics["mode"])
	if mode == "" {
		mode = "READ"
	}
	return circuitsOverlayState{
		rows:               rows,
		cols:               cols,
		mode:               mode,
		architecture:       valueOr(metrics["architecture"], "0T1R (Passive)"),
		selectedRow:        clampInt(selectedRow, 0, rows-1),
		selectedCol:        clampInt(selectedCol, 0, cols-1),
		writeTarget:        target,
		coupling:           valueOr(metrics["coupling"], "Tier-A"),
		isppEngine:         valueOr(metrics["ispp_engine"], "Preisach (Level-based)"),
		lastOperation:      lastOperation,
		halfSelectState:    valueOr(metrics["half_select_state"], "inactive"),
		halfSelectCells:    parseLeadingInt(metrics["half_select_cells"]),
		disturbVoltage:     valueOr(metrics["disturb_voltage"], "0.00 V"),
		stressBudget:       valueOr(metrics["stress_budget"], "inactive"),
		stressPerPulse:     valueOr(metrics["stress_per_pulse"], "0.000000 level/pulse"),
		pvtTempSweep:       valueOr(metrics["pvt_temperature_sweep"], "not evaluated"),
		pvtProcessYield:    valueOr(metrics["pvt_process_yield"], "not evaluated"),
		pvtCornerENOB:      valueOr(metrics["pvt_corner_enob"], "not evaluated"),
		pvtNoiseCeiling:    valueOr(metrics["pvt_noise_ceiling"], "not evaluated"),
		specPowerLatency:   valueOr(metrics["spec_power_latency"], "not evaluated"),
		specThroughput:     valueOr(metrics["spec_throughput"], "not evaluated"),
		specCompliance:     valueOr(metrics["spec_compliance"], "not evaluated"),
		timingActive:       valueOr(metrics["timing_active"], "not evaluated"),
		timingActivePhases: valueOr(metrics["timing_active_phases"], "not evaluated"),
	}
}

func drawCircuitsPanelBackground(cc *gg.Context, width, height float64, state circuitsOverlayState) {
	cc.SetRGBA(0.04, 0.07, 0.06, 0.93)
	cc.DrawRoundedRectangle(0, 0, width, height, 14)
	cc.Fill()

	cc.SetRGBA(0.82, 0.92, 0.88, 0.9)
	cc.DrawStringAnchored("Module 4 Unified Circuit Canvas", 24, 28, 0, 0.5)
	cc.SetRGBA(0.58, 0.68, 0.63, 1)
	cc.DrawStringAnchored(fmt.Sprintf("%s | %s | %dx%d array", state.mode, state.architecture, state.rows, state.cols), 24, 50, 0, 0.5)

	r, g, b := modeColor(state.mode)
	cc.SetRGBA(r, g, b, 0.95)
	cc.DrawRoundedRectangle(width-150, 18, 126, 30, 15)
	cc.Fill()
	cc.SetRGBA(0.04, 0.07, 0.06, 1)
	cc.DrawStringAnchored(state.mode, width-87, 37, 0.5, 0.5)
}

func drawCircuitsGrid(cc *gg.Context, state circuitsOverlayState, x, y, size float64) {
	visibleRows := minInt(state.rows, 32)
	visibleCols := minInt(state.cols, 32)
	cell := size / float64(maxInt(visibleRows, visibleCols))
	if cell < 4 {
		cell = 4
	}
	gridW := float64(visibleCols) * cell
	gridH := float64(visibleRows) * cell

	cc.SetRGBA(0.02, 0.03, 0.03, 1)
	cc.DrawRoundedRectangle(x-8, y-8, gridW+16, gridH+16, 10)
	cc.Fill()

	for r := range visibleRows {
		for c := range visibleCols {
			t := float64((r*visibleCols+c)%30) / 29.0
			cellR := 0.09 + t*0.18
			cellG := 0.22 + t*0.32
			cellB := 0.20 + t*0.22
			cc.SetRGBA(cellR, cellG, cellB, 1)
			cc.DrawRectangle(x+float64(c)*cell+1, y+float64(r)*cell+1, cell-2, cell-2)
			cc.Fill()
		}
	}

	displayRow := scaledIndex(state.selectedRow, state.rows, visibleRows)
	displayCol := scaledIndex(state.selectedCol, state.cols, visibleCols)
	selX := x + float64(displayCol)*cell
	selY := y + float64(displayRow)*cell
	r, g, b := modeColor(state.mode)
	cc.SetRGBA(r, g, b, 0.96)
	cc.SetLineWidth(3)
	cc.DrawRectangle(selX+1, selY+1, cell-2, cell-2)
	cc.Stroke()
	cc.SetRGBA(r, g, b, 0.32)
	cc.DrawRectangle(selX+2, selY+2, cell-4, cell-4)
	cc.Fill()

	cc.SetRGBA(0.74, 0.82, 0.78, 1)
	cc.DrawStringAnchored(fmt.Sprintf("selected [%d,%d]", state.selectedRow, state.selectedCol), x, y+gridH+22, 0, 0.5)
	if state.rows > visibleRows || state.cols > visibleCols {
		cc.SetRGBA(0.55, 0.64, 0.60, 1)
		cc.DrawStringAnchored(fmt.Sprintf("compressed %dx%d view", visibleRows, visibleCols), x+gridW, y+gridH+22, 1, 0.5)
	}
}

func drawHalfSelectStressOverlay(cc *gg.Context, state circuitsOverlayState, x, y, size float64) {
	if state.halfSelectCells <= 0 || state.halfSelectState == "inactive" || state.halfSelectState == "isolated" {
		return
	}
	visibleRows := minInt(state.rows, 32)
	visibleCols := minInt(state.cols, 32)
	cell := size / float64(maxInt(visibleRows, visibleCols))
	if cell < 4 {
		cell = 4
	}
	displayCol := scaledIndex(state.selectedCol, state.cols, visibleCols)
	displayRow := scaledIndex(state.selectedRow, state.rows, visibleRows)
	colX := x + float64(displayCol)*cell

	cc.SetRGBA(1.0, 0.66, 0.20, 0.16)
	cc.DrawRectangle(colX+1, y+1, cell-2, float64(visibleRows)*cell-2)
	cc.Fill()
	cc.SetRGBA(1.0, 0.66, 0.20, 0.75)
	cc.SetLineWidth(2)
	cc.DrawRectangle(colX+1, y+1, cell-2, float64(visibleRows)*cell-2)
	cc.Stroke()

	cc.SetRGBA(1.0, 0.78, 0.32, 0.58)
	for r := range visibleRows {
		if r == displayRow {
			continue
		}
		cc.DrawRectangle(colX+3, y+float64(r)*cell+3, cell-6, cell-6)
		cc.Fill()
	}
}

func drawCircuitPathState(cc *gg.Context, state circuitsOverlayState, gridX, gridY, gridSize, detailX, panelH float64) {
	visibleRows := minInt(state.rows, 32)
	visibleCols := minInt(state.cols, 32)
	cell := gridSize / float64(maxInt(visibleRows, visibleCols))
	if cell < 4 {
		cell = 4
	}
	gridW := float64(visibleCols) * cell
	gridH := float64(visibleRows) * cell
	selectedX := gridX + float64(scaledIndex(state.selectedCol, state.cols, visibleCols))*cell + cell/2
	selectedY := gridY + float64(scaledIndex(state.selectedRow, state.rows, visibleRows))*cell + cell/2
	r, g, b := modeColor(state.mode)
	cc.SetRGBA(r, g, b, 0.9)
	cc.SetLineWidth(4)

	switch state.mode {
	case "WRITE":
		sourceX := detailX + 22
		sourceY := gridY + 34
		drawCircuitBlock(cc, sourceX, sourceY-22, 100, 44, "DAC/CP", r, g, b)
		drawPolyline(cc, sourceX+100, sourceY, selectedX, selectedY)
		cc.DrawStringAnchored(fmt.Sprintf("target L%d", state.writeTarget), selectedX+10, selectedY-14, 0, 1)
	case "COMPUTE":
		inputY := gridY + gridH + 44
		drawCircuitBlock(cc, gridX, inputY-20, gridW, 40, "Input vector", r, g, b)
		drawPolyline(cc, gridX+gridW/2, inputY-20, selectedX, selectedY)
		drawPolyline(cc, selectedX, selectedY, detailX+66, panelH-88)
		drawCircuitBlock(cc, detailX+14, panelH-110, 118, 44, "Column sum", r, g, b)
	default:
		senseX := detailX + 16
		senseY := gridY + gridH/2
		drawPolyline(cc, selectedX, selectedY, senseX, senseY)
		drawCircuitBlock(cc, senseX, senseY-22, 92, 44, "TIA", r, g, b)
		drawCircuitBlock(cc, senseX+112, senseY-22, 92, 44, "ADC", r, g, b)
		drawPolyline(cc, senseX+92, senseY, senseX+112, senseY)
	}
}

func drawCircuitsDetails(cc *gg.Context, state circuitsOverlayState, x, y, width, height float64) {
	cc.SetRGBA(0.09, 0.12, 0.11, 0.96)
	cc.DrawRoundedRectangle(x, y, width, height, 10)
	cc.Fill()

	lines := []string{
		"Path: " + state.mode,
		"Cell: " + fmt.Sprintf("[%d,%d]", state.selectedRow, state.selectedCol),
		"Target: L" + strconv.Itoa(state.writeTarget),
		"Coupling: " + state.coupling,
		"ISPP: " + compactISPPEngine(state.isppEngine),
		"Stress: " + state.halfSelectState,
		"Cells: " + strconv.Itoa(state.halfSelectCells),
		"Budget: " + state.stressBudget,
		"PVT: " + state.pvtProcessYield,
		"Temp: " + state.pvtTempSweep,
		"ENOB: " + compactPVTENOB(state.pvtCornerENOB),
		"Ceil: " + state.pvtNoiseCeiling,
		"Spec: " + state.specPowerLatency,
		"Perf: " + state.specThroughput,
		"Rule: " + compactSpecCompliance(state.specCompliance),
		"Time: " + state.timingActive,
		"Phase: " + compactTimingPhases(state.timingActivePhases),
	}
	cc.SetRGBA(0.84, 0.91, 0.87, 1)
	cc.DrawStringAnchored("State", x+14, y+22, 0, 0.5)
	cc.SetRGBA(0.67, 0.76, 0.72, 1)
	lineStep := 18.0
	statusY := y + height - 76
	maxLines := int((statusY - (y + 44)) / lineStep)
	if maxLines < 1 {
		maxLines = 1
	}
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	for i, line := range lines {
		cc.DrawStringAnchored(elideCircuitStatus(line, 46), x+14, y+52+float64(i)*lineStep, 0, 0.5)
	}

	cc.SetRGBA(0.03, 0.05, 0.05, 1)
	cc.DrawRoundedRectangle(x+12, statusY, width-24, 52, 8)
	cc.Fill()
	cc.SetRGBA(0.82, 0.92, 0.88, 1)
	cc.DrawStringAnchored("Last operation", x+24, statusY+17, 0, 0.5)
	cc.SetRGBA(0.60, 0.70, 0.66, 1)
	cc.DrawStringAnchored(elideCircuitStatus(state.lastOperation, 43), x+24, statusY+38, 0, 0.5)
}

func drawCircuitBlock(cc *gg.Context, x, y, w, h float64, label string, r, g, b float64) {
	cc.SetRGBA(r, g, b, 0.18)
	cc.DrawRoundedRectangle(x, y, w, h, 8)
	cc.Fill()
	cc.SetRGBA(r, g, b, 0.88)
	cc.SetLineWidth(2)
	cc.DrawRoundedRectangle(x, y, w, h, 8)
	cc.Stroke()
	cc.SetRGBA(0.88, 0.94, 0.90, 1)
	cc.DrawStringAnchored(label, x+w/2, y+h/2, 0.5, 0.5)
}

func drawPolyline(cc *gg.Context, x1, y1, x2, y2 float64) {
	midX := x1 + (x2-x1)*0.55
	cc.MoveTo(x1, y1)
	cc.LineTo(midX, y1)
	cc.LineTo(midX, y2)
	cc.LineTo(x2, y2)
	cc.Stroke()
	cc.DrawCircle(x2, y2, 4)
	cc.Fill()
}

func modeColor(mode string) (float64, float64, float64) {
	switch strings.ToUpper(mode) {
	case "WRITE":
		return 0.94, 0.50, 0.22
	case "COMPUTE":
		return 0.38, 0.62, 0.95
	default:
		return 0.38, 0.76, 0.62
	}
}

func parseCircuitArray(value string) (int, int) {
	parts := strings.Split(strings.ToLower(value), "x")
	if len(parts) != 2 {
		return 8, 8
	}
	rows, errRows := strconv.Atoi(strings.TrimSpace(parts[0]))
	cols, errCols := strconv.Atoi(strings.TrimSpace(parts[1]))
	if errRows != nil || errCols != nil || rows <= 0 || cols <= 0 {
		return 8, 8
	}
	return rows, cols
}

func parseCircuitCell(value string) (int, int) {
	var row, col int
	if _, err := fmt.Sscanf(value, "[%d,%d]", &row, &col); err != nil {
		return 0, 0
	}
	return row, col
}

func parseCircuitTarget(value string) int {
	beforeSlash := value
	if idx := strings.Index(value, "/"); idx >= 0 {
		beforeSlash = value[:idx]
	}
	target, err := strconv.Atoi(strings.TrimSpace(beforeSlash))
	if err != nil {
		return 15
	}
	return target
}

func parseLeadingInt(value string) int {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return 0
	}
	parsed, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0
	}
	return parsed
}

func scaledIndex(index, total, visible int) int {
	if visible <= 1 || total <= 1 {
		return 0
	}
	scaled := int(math.Round(float64(index) * float64(visible-1) / float64(total-1)))
	return clampInt(scaled, 0, visible-1)
}

func compactISPPEngine(engine string) string {
	if strings.Contains(engine, "Landau") {
		return "L-K ODE"
	}
	return "Preisach"
}

func compactPVTENOB(value string) string {
	value = strings.TrimSuffix(value, " bits")
	value = strings.ReplaceAll(value, " / ", " ")
	return value
}

func compactSpecCompliance(value string) string {
	return strings.TrimPrefix(value, "OK: ")
}

func compactTimingPhases(value string) string {
	value = strings.TrimSuffix(value, " ns")
	return strings.ReplaceAll(value, " / ", " ")
}

func elideCircuitStatus(status string, limit int) string {
	if len(status) <= limit {
		return status
	}
	if limit <= 3 {
		return status[:limit]
	}
	return status[:limit-3] + "..."
}

func valueOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func clampFloat(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clampInt(value, min, max int) int {
	if max < min {
		return min
	}
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
