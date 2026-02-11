package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// TestNewModel verifies default model initialization
func TestNewModel(t *testing.T) {
	m := NewModel()

	// Check default values
	if m.material == nil {
		t.Error("Expected material to be initialized, got nil")
	}
	if m.preisach == nil {
		t.Error("Expected preisach to be initialized, got nil")
	}
	if len(m.materials) == 0 {
		t.Error("Expected materials to be populated, got empty slice")
	}
	if m.matIndex != 0 {
		t.Errorf("Expected matIndex=0, got %d", m.matIndex)
	}
	if m.maxHistory != 200 {
		t.Errorf("Expected maxHistory=200, got %d", m.maxHistory)
	}
	if cap(m.eHistory) != 200 {
		t.Errorf("Expected eHistory capacity=200, got %d", cap(m.eHistory))
	}
	if cap(m.pHistory) != 200 {
		t.Errorf("Expected pHistory capacity=200, got %d", cap(m.pHistory))
	}
	if m.waveform != WaveformSine {
		t.Errorf("Expected waveform=WaveformSine, got %v", m.waveform)
	}
	if !m.autoMode {
		t.Error("Expected autoMode=true, got false")
	}
	if m.paused {
		t.Error("Expected paused=false, got true")
	}
	if m.frequency != 2.0 {
		t.Errorf("Expected frequency=2.0, got %f", m.frequency)
	}
	if m.plotWidth != 60 {
		t.Errorf("Expected plotWidth=60, got %d", m.plotWidth)
	}
	if m.plotHeight != 20 {
		t.Errorf("Expected plotHeight=20, got %d", m.plotHeight)
	}
}

// TestNewModelWithMaterial verifies material selection
func TestNewModelWithMaterial(t *testing.T) {
	// Test with default material (first one)
	m1 := NewModelWithMaterial("")
	if m1.material == nil {
		t.Error("Expected material to be initialized")
	}
	if m1.matIndex != 0 {
		t.Errorf("Expected matIndex=0 for unknown material, got %d", m1.matIndex)
	}

	// Test with known material name
	if len(m1.materials) > 1 {
		testMaterialName := m1.materials[1].Name
		m2 := NewModelWithMaterial(testMaterialName)
		if m2.material.Name != testMaterialName {
			t.Errorf("Expected material.Name=%s, got %s", testMaterialName, m2.material.Name)
		}
		if m2.matIndex != 1 {
			t.Errorf("Expected matIndex=1, got %d", m2.matIndex)
		}
	}
}

// TestDefaultKeyMap verifies keyboard binding initialization
func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	// Check that all key bindings are initialized
	testCases := []struct {
		name    string
		binding key.Binding
		wantKey string
	}{
		{"Up", km.Up, "up"},
		{"Down", km.Down, "down"},
		{"Left", km.Left, "left"},
		{"Right", km.Right, "right"},
		{"Space", km.Space, " "},
		{"Tab", km.Tab, "tab"},
		{"Enter", km.Enter, "enter"},
		{"Reset", km.Reset, "r"},
		{"Help", km.Help, "?"},
		{"Quit", km.Quit, "q"},
		{"ToggleAuto", km.ToggleAuto, "a"},
		{"Material", km.Material, "m"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keys := tc.binding.Keys()
			if len(keys) == 0 {
				t.Errorf("Expected binding for %s to have keys, got none", tc.name)
				return
			}
			found := false
			for _, k := range keys {
				if k == tc.wantKey {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected binding for %s to include key %q, got %v", tc.name, tc.wantKey, keys)
			}
		})
	}
}

// TestKeyMapShortHelp verifies short help returns expected bindings
func TestKeyMapShortHelp(t *testing.T) {
	km := DefaultKeyMap()
	shortHelp := km.ShortHelp()

	if len(shortHelp) == 0 {
		t.Error("Expected ShortHelp to return bindings, got empty slice")
	}

	// Should include at least Up, Down, Space, Tab, Reset, Quit, Help
	expectedMin := 7
	if len(shortHelp) < expectedMin {
		t.Errorf("Expected ShortHelp to have at least %d bindings, got %d", expectedMin, len(shortHelp))
	}
}

// TestKeyMapFullHelp verifies full help returns expected binding groups
func TestKeyMapFullHelp(t *testing.T) {
	km := DefaultKeyMap()
	fullHelp := km.FullHelp()

	if len(fullHelp) != 3 {
		t.Errorf("Expected FullHelp to have 3 groups, got %d", len(fullHelp))
	}

	// Check first group has directional keys
	if len(fullHelp) > 0 && len(fullHelp[0]) != 4 {
		t.Errorf("Expected first group to have 4 bindings (directional), got %d", len(fullHelp[0]))
	}

	// Check second group has toggle keys
	if len(fullHelp) > 1 && len(fullHelp[1]) != 4 {
		t.Errorf("Expected second group to have 4 bindings (toggles), got %d", len(fullHelp[1]))
	}

	// Check third group has action keys
	if len(fullHelp) > 2 && len(fullHelp[2]) != 4 {
		t.Errorf("Expected third group to have 4 bindings (actions), got %d", len(fullHelp[2]))
	}
}

// TestWaveformTypeString verifies waveform string conversion
func TestWaveformTypeString(t *testing.T) {
	testCases := []struct {
		waveform WaveformType
		want     string
	}{
		{WaveformManual, "Manual"},
		{WaveformSine, "Sine"},
		{WaveformTriangle, "Triangle"},
		{WaveformSquare, "Square"},
		{WaveformType(99), "Unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.want, func(t *testing.T) {
			got := tc.waveform.String()
			if got != tc.want {
				t.Errorf("WaveformType(%d).String() = %q, want %q", tc.waveform, got, tc.want)
			}
		})
	}
}

// TestWaveformTypeConstants verifies waveform type constant values
func TestWaveformTypeConstants(t *testing.T) {
	if WaveformManual != 0 {
		t.Errorf("Expected WaveformManual=0, got %d", WaveformManual)
	}
	if WaveformSine != 1 {
		t.Errorf("Expected WaveformSine=1, got %d", WaveformSine)
	}
	if WaveformTriangle != 2 {
		t.Errorf("Expected WaveformTriangle=2, got %d", WaveformTriangle)
	}
	if WaveformSquare != 3 {
		t.Errorf("Expected WaveformSquare=3, got %d", WaveformSquare)
	}
}

// TestModelInit verifies Init returns tick command
func TestModelInit(t *testing.T) {
	m := NewModel()
	cmd := m.Init()

	if cmd == nil {
		t.Error("Expected Init to return a command, got nil")
	}
}

// TestModelUpdateQuit verifies quit message handling
func TestModelUpdateQuit(t *testing.T) {
	m := NewModel()

	// Send 'q' key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Error("Expected quit command, got nil")
	}
}

// TestModelUpdateHelp verifies help toggle
func TestModelUpdateHelp(t *testing.T) {
	m := NewModel()
	initialShowHelp := m.showHelp

	// Press '?' to toggle help
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	if m.showHelp == initialShowHelp {
		t.Errorf("Expected showHelp to toggle from %v, but it didn't", initialShowHelp)
	}

	// Press '?' again to toggle back
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	if m.showHelp != initialShowHelp {
		t.Errorf("Expected showHelp to toggle back to %v, got %v", initialShowHelp, m.showHelp)
	}
}

// TestModelUpdatePause verifies pause toggle
func TestModelUpdatePause(t *testing.T) {
	m := NewModel()
	initialPaused := m.paused

	// Press space to toggle pause
	msg := tea.KeyMsg{Type: tea.KeySpace}
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	if m.paused == initialPaused {
		t.Errorf("Expected paused to toggle from %v, but it didn't", initialPaused)
	}

	// Press space again to toggle back
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	if m.paused != initialPaused {
		t.Errorf("Expected paused to toggle back to %v, got %v", initialPaused, m.paused)
	}
}

// TestModelUpdateDirectional verifies directional key handling
func TestModelUpdateDirectional(t *testing.T) {
	m := NewModel()
	initialE := m.electricField

	// Press up arrow
	msg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	if m.electricField <= initialE {
		t.Errorf("Expected electricField to increase after up key, got %f (was %f)", m.electricField, initialE)
	}
	if m.waveform != WaveformManual {
		t.Errorf("Expected waveform to switch to Manual after directional key, got %v", m.waveform)
	}

	// Press down arrow
	currentE := m.electricField
	msg = tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	if m.electricField >= currentE {
		t.Errorf("Expected electricField to decrease after down key, got %f (was %f)", m.electricField, currentE)
	}
}

// TestModelUpdateTab verifies waveform cycling
func TestModelUpdateTab(t *testing.T) {
	m := NewModel()
	m.waveform = WaveformManual

	// Press tab to cycle waveform
	msg := tea.KeyMsg{Type: tea.KeyTab}
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	if m.waveform != WaveformSine {
		t.Errorf("Expected waveform to cycle to Sine, got %v", m.waveform)
	}

	// Continue cycling
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	if m.waveform != WaveformTriangle {
		t.Errorf("Expected waveform to cycle to Triangle, got %v", m.waveform)
	}

	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	if m.waveform != WaveformSquare {
		t.Errorf("Expected waveform to cycle to Square, got %v", m.waveform)
	}

	// Should wrap around to Manual
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	if m.waveform != WaveformManual {
		t.Errorf("Expected waveform to cycle back to Manual, got %v", m.waveform)
	}
}

// TestModelUpdateReset verifies reset functionality
func TestModelUpdateReset(t *testing.T) {
	m := NewModel()

	// Modify state
	m.electricField = 1e8
	m.polarization = 0.5
	m.normalizedP = 0.8
	m.discreteLevel = 25
	m.eHistory = append(m.eHistory, 1.0, 2.0, 3.0)
	m.pHistory = append(m.pHistory, 0.1, 0.2, 0.3)
	m.simTime = 10.0

	// Press 'r' to reset
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	if m.electricField != 0 {
		t.Errorf("Expected electricField=0 after reset, got %f", m.electricField)
	}
	if m.polarization != 0 {
		t.Errorf("Expected polarization=0 after reset, got %f", m.polarization)
	}
	if m.normalizedP != 0 {
		t.Errorf("Expected normalizedP=0 after reset, got %f", m.normalizedP)
	}
	if m.discreteLevel != 15 {
		t.Errorf("Expected discreteLevel=15 after reset, got %d", m.discreteLevel)
	}
	if len(m.eHistory) != 0 {
		t.Errorf("Expected eHistory to be empty after reset, got length %d", len(m.eHistory))
	}
	if len(m.pHistory) != 0 {
		t.Errorf("Expected pHistory to be empty after reset, got length %d", len(m.pHistory))
	}
	if m.simTime != 0 {
		t.Errorf("Expected simTime=0 after reset, got %f", m.simTime)
	}
}

// TestModelUpdateMaterial verifies material cycling
func TestModelUpdateMaterial(t *testing.T) {
	m := NewModel()
	if len(m.materials) < 2 {
		t.Skip("Need at least 2 materials to test cycling")
	}

	initialMat := m.material.Name
	initialIndex := m.matIndex

	// Press 'm' to cycle material
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	if m.matIndex == initialIndex {
		t.Errorf("Expected matIndex to change from %d, but it didn't", initialIndex)
	}
	if m.material.Name == initialMat {
		t.Errorf("Expected material to change from %s, but it didn't", initialMat)
	}
	if len(m.eHistory) != 0 || len(m.pHistory) != 0 {
		t.Error("Expected history to be cleared after material change")
	}
}

// TestModelUpdateWindowSize verifies window resize handling
func TestModelUpdateWindowSize(t *testing.T) {
	m := NewModel()

	// Send window size message
	msg := tea.WindowSizeMsg{Width: 100, Height: 40}
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	// plotWidth should be min(100-30, 80) = 70
	expectedWidth := 70
	if m.plotWidth != expectedWidth {
		t.Errorf("Expected plotWidth=%d after resize, got %d", expectedWidth, m.plotWidth)
	}

	// plotHeight should be min(40-15, 25) = 25
	expectedHeight := 25
	if m.plotHeight != expectedHeight {
		t.Errorf("Expected plotHeight=%d after resize, got %d", expectedHeight, m.plotHeight)
	}
}

// TestModelUpdateTick verifies tick message handling
func TestModelUpdateTick(t *testing.T) {
	m := NewModel()
	m.paused = false
	m.lastTick = time.Now().Add(-100 * time.Millisecond)
	initialSimTime := m.simTime

	// Send tick message
	msg := tickMsg(time.Now())
	updatedModel, cmd := m.Update(msg)
	m = updatedModel.(Model)

	if m.simTime <= initialSimTime {
		t.Errorf("Expected simTime to increase after tick, got %f (was %f)", m.simTime, initialSimTime)
	}
	if cmd == nil {
		t.Error("Expected tick to return next tick command, got nil")
	}
}

// TestModelUpdateTickPaused verifies tick doesn't update when paused
func TestModelUpdateTickPaused(t *testing.T) {
	m := NewModel()
	m.paused = true
	m.lastTick = time.Now().Add(-100 * time.Millisecond)
	initialSimTime := m.simTime
	initialE := m.electricField

	// Send tick message
	msg := tickMsg(time.Now())
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	// When paused, simulation should not advance
	if m.simTime != initialSimTime {
		t.Errorf("Expected simTime to remain %f when paused, got %f", initialSimTime, m.simTime)
	}
	if m.electricField != initialE {
		t.Errorf("Expected electricField to remain %f when paused, got %f", initialE, m.electricField)
	}
}

// TestModelView verifies View returns non-empty string
func TestModelView(t *testing.T) {
	m := NewModel()
	view := m.View()

	if view == "" {
		t.Error("Expected View to return non-empty string, got empty")
	}
	if !strings.Contains(view, "FeCIM") {
		t.Error("Expected View to contain 'FeCIM' in title")
	}
}

// TestRenderPEPlot verifies PE plot rendering
func TestRenderPEPlot(t *testing.T) {
	m := NewModel()
	plot := m.renderPEPlot()

	if plot == "" {
		t.Error("Expected renderPEPlot to return non-empty string, got empty")
	}
	if !strings.Contains(plot, "P") {
		t.Error("Expected plot to contain polarization label 'P'")
	}
	if !strings.Contains(plot, "E") {
		t.Error("Expected plot to contain electric field label 'E'")
	}
}

// TestRenderPEPlotEdgeCases verifies plot handles edge cases
func TestRenderPEPlotEdgeCases(t *testing.T) {
	m := NewModel()

	// Test with very small but valid dimensions
	m.plotWidth = 5
	m.plotHeight = 3
	plot := m.renderPEPlot()
	if plot == "" {
		t.Error("Expected renderPEPlot to handle small dimensions")
	}

	// Test with large history
	m.plotWidth = 60
	m.plotHeight = 20
	for i := 0; i < 300; i++ {
		m.eHistory = append(m.eHistory, float64(i))
		m.pHistory = append(m.pHistory, float64(i)*0.1)
	}
	plot = m.renderPEPlot()
	if plot == "" {
		t.Error("Expected renderPEPlot to handle large history")
	}

	// Test with extreme polarization values
	m.eHistory = []float64{0, 1e8, -1e8}
	m.pHistory = []float64{0, m.material.Ps, -m.material.Ps}
	plot = m.renderPEPlot()
	if plot == "" {
		t.Error("Expected renderPEPlot to handle extreme values")
	}
}

// TestRenderInfoPanel verifies info panel rendering
func TestRenderInfoPanel(t *testing.T) {
	m := NewModel()
	info := m.renderInfoPanel()

	if info == "" {
		t.Error("Expected renderInfoPanel to return non-empty string, got empty")
	}
	if !strings.Contains(info, "Material") {
		t.Error("Expected info panel to contain 'Material'")
	}
	if !strings.Contains(info, "Electric Field") {
		t.Error("Expected info panel to contain 'Electric Field'")
	}
	if !strings.Contains(info, "Polarization") {
		t.Error("Expected info panel to contain 'Polarization'")
	}
	if !strings.Contains(info, "Discrete Level") {
		t.Error("Expected info panel to contain 'Discrete Level'")
	}
}

// TestRenderLevelBar verifies level bar rendering
func TestRenderLevelBar(t *testing.T) {
	m := NewModel()
	levelBar := m.renderLevelBar()

	if levelBar == "" {
		t.Error("Expected renderLevelBar to return non-empty string, got empty")
	}
	if !strings.Contains(levelBar, "30 Levels") {
		t.Error("Expected level bar to mention '30 Levels'")
	}

	// Test with different discrete levels
	testLevels := []int{0, 15, 29}
	for _, level := range testLevels {
		m.discreteLevel = level
		levelBar = m.renderLevelBar()
		if levelBar == "" {
			t.Errorf("Expected renderLevelBar to handle level %d, got empty", level)
		}
	}
}

// TestRenderStatusBar verifies status bar rendering
func TestRenderStatusBar(t *testing.T) {
	m := NewModel()

	// Test running status
	m.paused = false
	statusBar := m.renderStatusBar()
	if statusBar == "" {
		t.Error("Expected renderStatusBar to return non-empty string, got empty")
	}
	if !strings.Contains(statusBar, "RUNNING") {
		t.Error("Expected status bar to show 'RUNNING' when not paused")
	}

	// Test paused status
	m.paused = true
	statusBar = m.renderStatusBar()
	if !strings.Contains(statusBar, "PAUSED") {
		t.Error("Expected status bar to show 'PAUSED' when paused")
	}

	// Should contain time and switched fraction
	if !strings.Contains(statusBar, "t =") {
		t.Error("Expected status bar to contain simulation time")
	}
	if !strings.Contains(statusBar, "Switched") {
		t.Error("Expected status bar to contain switched percentage")
	}
}

// TestUpdateSimulationAutoMode verifies auto mode simulation updates
func TestUpdateSimulationAutoMode(t *testing.T) {
	m := NewModel()
	m.autoMode = true
	m.waveform = WaveformSine
	m.lastTick = time.Now()

	initialHistLen := len(m.eHistory)

	// Advance simulation
	m.updateSimulation()

	// Check that history was updated
	if len(m.eHistory) != initialHistLen+1 {
		t.Errorf("Expected eHistory length to increase by 1, got %d (was %d)", len(m.eHistory), initialHistLen)
	}
	if len(m.pHistory) != len(m.eHistory) {
		t.Errorf("Expected pHistory length to match eHistory (%d), got %d", len(m.eHistory), len(m.pHistory))
	}
}

// TestUpdateSimulationWaveforms verifies different waveform types
func TestUpdateSimulationWaveforms(t *testing.T) {
	testCases := []struct {
		name     string
		waveform WaveformType
	}{
		{"Sine", WaveformSine},
		{"Triangle", WaveformTriangle},
		{"Square", WaveformSquare},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := NewModel()
			m.autoMode = true
			m.waveform = tc.waveform
			m.lastTick = time.Now()

			// Run multiple simulation steps
			for i := 0; i < 10; i++ {
				m.updateSimulation()
			}

			if len(m.eHistory) != 10 {
				t.Errorf("Expected 10 history entries, got %d", len(m.eHistory))
			}
		})
	}
}

// TestUpdateSimulationManualMode verifies manual mode doesn't auto-update E-field
func TestUpdateSimulationManualMode(t *testing.T) {
	m := NewModel()
	m.autoMode = false
	m.waveform = WaveformManual
	m.electricField = 1e8
	m.lastTick = time.Now()

	initialE := m.electricField

	// Advance simulation
	m.updateSimulation()

	// In manual mode, E-field should not change automatically
	if m.electricField != initialE {
		t.Errorf("Expected electricField to remain %f in manual mode, got %f", initialE, m.electricField)
	}
}

// TestUpdateSimulationHistoryCap verifies history is capped at maxHistory
func TestUpdateSimulationHistoryCap(t *testing.T) {
	m := NewModel()
	m.autoMode = true
	m.waveform = WaveformSine
	m.maxHistory = 10
	m.lastTick = time.Now()

	// Add more than maxHistory entries
	for i := 0; i < 20; i++ {
		m.updateSimulation()
	}

	if len(m.eHistory) > m.maxHistory {
		t.Errorf("Expected eHistory to be capped at %d, got %d", m.maxHistory, len(m.eHistory))
	}
	if len(m.pHistory) > m.maxHistory {
		t.Errorf("Expected pHistory to be capped at %d, got %d", m.maxHistory, len(m.pHistory))
	}
}

// TestUpdateSimulationDiscreteLevelBounds verifies discrete level clamping
func TestUpdateSimulationDiscreteLevelBounds(t *testing.T) {
	m := NewModel()
	m.lastTick = time.Now()

	// Test extreme normalized polarization values
	testCases := []struct {
		normalizedP   float64
		expectedLevel int
	}{
		{-1.0, 0},   // Min
		{0.0, 15},   // Middle
		{1.0, 29},   // Max
		{-2.0, 0},   // Below min (should clamp)
		{2.0, 29},   // Above max (should clamp)
	}

	for _, tc := range testCases {
		m.normalizedP = tc.normalizedP
		m.updateSimulation()

		if m.discreteLevel < 0 {
			t.Errorf("discreteLevel should not be negative, got %d", m.discreteLevel)
		}
		if m.discreteLevel > 29 {
			t.Errorf("discreteLevel should not exceed 29, got %d", m.discreteLevel)
		}
	}
}

// TestMinHelper verifies the min helper function
func TestMinHelper(t *testing.T) {
	testCases := []struct {
		a, b, want int
	}{
		{1, 2, 1},
		{5, 3, 3},
		{0, 0, 0},
		{-1, -5, -5},
		{100, 100, 100},
	}

	for _, tc := range testCases {
		got := min(tc.a, tc.b)
		if got != tc.want {
			t.Errorf("min(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

// TestTickCommand verifies tick command generation
func TestTickCommand(t *testing.T) {
	cmd := tick()
	if cmd == nil {
		t.Error("Expected tick() to return a command, got nil")
	}
}
