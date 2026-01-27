// Package tui provides an enhanced terminal user interface for the hysteresis demo.
// Uses charmbracelet/bubbletea for interactive TUI and lipgloss for styling.
package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
)

// Styles using lipgloss
var (
	// Color palette - FeCIM theme
	primaryColor   = lipgloss.Color("#00D4FF") // Cyan
	secondaryColor = lipgloss.Color("#FF6B6B") // Coral red
	accentColor    = lipgloss.Color("#4ECDC4") // Teal
	warningColor   = lipgloss.Color("#FFE66D") // Yellow
	bgColor        = lipgloss.Color("#1A1A2E") // Dark blue
	fgColor        = lipgloss.Color("#EAEAEA") // Light gray

	// Title style
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Background(lipgloss.Color("#16213E")).
			Padding(0, 2).
			MarginBottom(1)

	// Box style for panels
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	// Info panel style
	infoPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accentColor).
			Padding(0, 1)

	// Label style
	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Width(20)

	// Value style
	valueStyle = lipgloss.NewStyle().
			Foreground(fgColor).
			Bold(true)

	// Highlight style
	highlightStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)

	// Help style
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	// Plot axis style
	axisStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))

	// Positive polarization style
	posPolarStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	// Negative polarization style
	negPolarStyle = lipgloss.NewStyle().
			Foreground(primaryColor)
)

// KeyMap defines keyboard bindings
type KeyMap struct {
	Up         key.Binding
	Down       key.Binding
	Left       key.Binding
	Right      key.Binding
	Space      key.Binding
	Tab        key.Binding
	Enter      key.Binding
	Reset      key.Binding
	Help       key.Binding
	Quit       key.Binding
	ToggleAuto key.Binding
	Material   key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "increase E-field"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "decrease E-field"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "fine decrease"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "fine increase"),
		),
		Space: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "pause/resume"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "cycle waveform"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "apply field"),
		),
		Reset: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reset"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c", "esc"),
			key.WithHelp("q", "quit"),
		),
		ToggleAuto: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "auto mode"),
		),
		Material: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "cycle material"),
		),
	}
}

// ShortHelp returns short help text
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Space, k.Tab, k.Reset, k.Quit, k.Help}
}

// FullHelp returns full help text
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Space, k.Tab, k.ToggleAuto, k.Material},
		{k.Reset, k.Enter, k.Help, k.Quit},
	}
}

// WaveformType represents the input waveform type
type WaveformType int

const (
	WaveformManual WaveformType = iota
	WaveformSine
	WaveformTriangle
	WaveformSquare
)

func (w WaveformType) String() string {
	switch w {
	case WaveformManual:
		return "Manual"
	case WaveformSine:
		return "Sine"
	case WaveformTriangle:
		return "Triangle"
	case WaveformSquare:
		return "Square"
	default:
		return "Unknown"
	}
}

// Model represents the TUI application state
type Model struct {
	// Physics
	material  *ferroelectric.HZOMaterial
	preisach  *ferroelectric.MayergoyzPreisach
	materials []*ferroelectric.HZOMaterial
	matIndex  int

	// Simulation state
	electricField float64 // Current E-field (V/m)
	polarization  float64 // Current P (C/m²)
	normalizedP   float64 // P/Ps
	discreteLevel int     // 0-29

	// History for plotting
	eHistory   []float64
	pHistory   []float64
	maxHistory int

	// UI state
	waveform  WaveformType
	autoMode  bool
	paused    bool
	simTime   float64
	frequency float64
	showHelp  bool

	// Plot dimensions
	plotWidth  int
	plotHeight int

	// UI components
	keys KeyMap
	help help.Model

	// Tick for animation
	lastTick time.Time
}

// NewModel creates a new TUI model
func NewModel() Model {
	materials := ferroelectric.AllMaterials()

	mat := materials[0]
	preisach := ferroelectric.NewMayergoyzPreisach(mat, 30)

	return Model{
		material:   mat,
		preisach:   preisach,
		materials:  materials,
		matIndex:   0,
		maxHistory: 200,
		eHistory:   make([]float64, 0, 200),
		pHistory:   make([]float64, 0, 200),
		waveform:   WaveformSine,
		autoMode:   true,
		paused:     false,
		frequency:  2.0, // 2 Hz for visible animation
		plotWidth:  60,
		plotHeight: 20,
		keys:       DefaultKeyMap(),
		help:       help.New(),
		lastTick:   time.Now(),
	}
}

// tickMsg is sent on each animation frame
type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tick()
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp

		case key.Matches(msg, m.keys.Space):
			m.paused = !m.paused

		case key.Matches(msg, m.keys.Up):
			m.waveform = WaveformManual
			m.electricField += m.material.Ec * 0.1

		case key.Matches(msg, m.keys.Down):
			m.waveform = WaveformManual
			m.electricField -= m.material.Ec * 0.1

		case key.Matches(msg, m.keys.Right):
			m.waveform = WaveformManual
			m.electricField += m.material.Ec * 0.02

		case key.Matches(msg, m.keys.Left):
			m.waveform = WaveformManual
			m.electricField -= m.material.Ec * 0.02

		case key.Matches(msg, m.keys.Tab):
			m.waveform = WaveformType((int(m.waveform) + 1) % 4)
			if m.waveform != WaveformManual {
				m.autoMode = true
			}

		case key.Matches(msg, m.keys.ToggleAuto):
			m.autoMode = !m.autoMode
			if m.autoMode && m.waveform == WaveformManual {
				m.waveform = WaveformSine
			}

		case key.Matches(msg, m.keys.Reset):
			m.preisach.Reset()
			m.electricField = 0
			m.polarization = 0
			m.normalizedP = 0
			m.discreteLevel = 15
			m.eHistory = m.eHistory[:0]
			m.pHistory = m.pHistory[:0]
			m.simTime = 0

		case key.Matches(msg, m.keys.Material):
			m.matIndex = (m.matIndex + 1) % len(m.materials)
			m.material = m.materials[m.matIndex]
			m.preisach = ferroelectric.NewMayergoyzPreisach(m.material, 30)
			m.eHistory = m.eHistory[:0]
			m.pHistory = m.pHistory[:0]
		}

	case tickMsg:
		if !m.paused {
			m.updateSimulation()
		}
		return m, tick()

	case tea.WindowSizeMsg:
		m.plotWidth = min(msg.Width-30, 80)
		m.plotHeight = min(msg.Height-15, 25)
	}

	return m, nil
}

// updateSimulation advances the physics simulation
func (m *Model) updateSimulation() {
	dt := time.Since(m.lastTick).Seconds()
	m.lastTick = time.Now()
	m.simTime += dt

	// Generate E-field based on waveform
	if m.autoMode && m.waveform != WaveformManual {
		Emax := m.material.Ec * 2
		phase := 2 * math.Pi * m.frequency * m.simTime

		switch m.waveform {
		case WaveformSine:
			m.electricField = Emax * math.Sin(phase)
		case WaveformTriangle:
			p := math.Mod(phase, 2*math.Pi) / (2 * math.Pi)
			if p < 0.25 {
				m.electricField = Emax * (4 * p)
			} else if p < 0.75 {
				m.electricField = Emax * (2 - 4*p)
			} else {
				m.electricField = Emax * (4*p - 4)
			}
		case WaveformSquare:
			if math.Sin(phase) >= 0 {
				m.electricField = Emax
			} else {
				m.electricField = -Emax
			}
		}
	}

	// Update physics
	m.polarization = m.preisach.Update(m.electricField)
	m.normalizedP = m.preisach.NormalizedPolarization()
	m.discreteLevel = int(math.Round((m.normalizedP + 1) / 2 * 29))
	if m.discreteLevel < 0 {
		m.discreteLevel = 0
	}
	if m.discreteLevel > 29 {
		m.discreteLevel = 29
	}

	// Record history
	m.eHistory = append(m.eHistory, m.electricField)
	m.pHistory = append(m.pHistory, m.polarization)
	if len(m.eHistory) > m.maxHistory {
		m.eHistory = m.eHistory[1:]
		m.pHistory = m.pHistory[1:]
	}
}

// View renders the TUI
func (m Model) View() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("  FeCIM Hysteresis Visualizer - Demo 1  ")
	b.WriteString(title + "\n\n")

	// Main content: P-E plot on left, info panel on right
	plot := m.renderPEPlot()
	info := m.renderInfoPanel()
	levelBar := m.renderLevelBar()

	// Layout: plot | level bar | info
	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		plot,
		"  ",
		levelBar,
		"  ",
		info,
	)
	b.WriteString(mainContent + "\n\n")

	// Status bar
	statusBar := m.renderStatusBar()
	b.WriteString(statusBar + "\n")

	// Help
	if m.showHelp {
		helpView := m.help.View(m.keys)
		b.WriteString("\n" + helpStyle.Render(helpView))
	} else {
		b.WriteString(helpStyle.Render("Press ? for help"))
	}

	return b.String()
}

// renderPEPlot renders the P-E hysteresis curve
func (m Model) renderPEPlot() string {
	width := m.plotWidth
	height := m.plotHeight

	// Create plot grid
	grid := make([][]rune, height)
	colors := make([][]lipgloss.Style, height)
	for i := range grid {
		grid[i] = make([]rune, width)
		colors[i] = make([]lipgloss.Style, width)
		for j := range grid[i] {
			grid[i][j] = ' '
			colors[i][j] = axisStyle
		}
	}

	// Draw axes
	midX := width / 2
	midY := height / 2
	for x := 0; x < width; x++ {
		grid[midY][x] = '─'
	}
	for y := 0; y < height; y++ {
		grid[y][midX] = '│'
	}
	grid[midY][midX] = '┼'

	// Draw coercive field markers
	Emax := m.material.Ec * 2.5
	Pmax := m.material.Ps * 1.2
	EcX := midX + int(float64(width/2)*m.material.Ec/Emax)
	if EcX >= 0 && EcX < width {
		grid[midY][EcX] = '┃'
		colors[midY][EcX] = highlightStyle
	}
	EcX = midX - int(float64(width/2)*m.material.Ec/Emax)
	if EcX >= 0 && EcX < width {
		grid[midY][EcX] = '┃'
		colors[midY][EcX] = highlightStyle
	}

	// Draw remanent polarization markers
	PrY := midY - int(float64(height/2)*m.material.Pr/Pmax)
	if PrY >= 0 && PrY < height {
		grid[PrY][midX] = '━'
		colors[PrY][midX] = highlightStyle
	}
	PrY = midY + int(float64(height/2)*m.material.Pr/Pmax)
	if PrY >= 0 && PrY < height {
		grid[PrY][midX] = '━'
		colors[PrY][midX] = highlightStyle
	}

	// Plot hysteresis trail
	for i := range m.eHistory {
		e := m.eHistory[i]
		p := m.pHistory[i]

		x := midX + int(float64(width/2)*e/Emax)
		y := midY - int(float64(height/2)*p/Pmax)

		if x >= 0 && x < width && y >= 0 && y < height {
			// Fade older points
			age := float64(len(m.eHistory)-i) / float64(len(m.eHistory))
			if age < 0.3 {
				grid[y][x] = '●'
				if p >= 0 {
					colors[y][x] = posPolarStyle
				} else {
					colors[y][x] = negPolarStyle
				}
			} else if age < 0.6 {
				grid[y][x] = '○'
				colors[y][x] = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
			} else {
				grid[y][x] = '·'
				colors[y][x] = lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))
			}
		}
	}

	// Draw current position
	if len(m.eHistory) > 0 {
		x := midX + int(float64(width/2)*m.electricField/Emax)
		y := midY - int(float64(height/2)*m.polarization/Pmax)
		if x >= 0 && x < width && y >= 0 && y < height {
			grid[y][x] = '◆'
			colors[y][x] = highlightStyle
		}
	}

	// Build the plot string with colors
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("    P (µC/cm²) [Ps=%.1f]\n", m.material.Ps*100))
	sb.WriteString(fmt.Sprintf("    ↑ +%.1f\n", Pmax*100))

	for y := 0; y < height; y++ {
		sb.WriteString("    ")
		for x := 0; x < width; x++ {
			sb.WriteString(colors[y][x].Render(string(grid[y][x])))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("    ↓ -%.1f\n", Pmax*100))
	sb.WriteString(fmt.Sprintf("    %.1f ←─ E (MV/cm) ─→ +%.1f\n", -Emax/1e8, Emax/1e8))

	return boxStyle.Render(sb.String())
}

// renderInfoPanel renders the information panel
func (m Model) renderInfoPanel() string {
	var sb strings.Builder

	// Material info
	sb.WriteString(highlightStyle.Render("Material: ") + valueStyle.Render(m.material.Name) + "\n\n")

	// Current state
	sb.WriteString(labelStyle.Render("Electric Field:"))
	eVal := fmt.Sprintf("%.3f MV/cm", m.electricField/1e8)
	sb.WriteString(valueStyle.Render(eVal) + "\n")

	sb.WriteString(labelStyle.Render("Polarization:"))
	pVal := fmt.Sprintf("%.2f µC/cm²", m.polarization*100)
	if m.polarization >= 0 {
		sb.WriteString(posPolarStyle.Render(pVal) + "\n")
	} else {
		sb.WriteString(negPolarStyle.Render(pVal) + "\n")
	}

	sb.WriteString(labelStyle.Render("Normalized P/Ps:"))
	sb.WriteString(valueStyle.Render(fmt.Sprintf("%.3f", m.normalizedP)) + "\n")

	sb.WriteString(labelStyle.Render("Discrete Level:"))
	sb.WriteString(highlightStyle.Render(fmt.Sprintf("%d/30", m.discreteLevel+1)) + "\n\n")

	// Material parameters
	sb.WriteString(axisStyle.Render("─── Parameters ───") + "\n")
	sb.WriteString(labelStyle.Render("Pr (remanent):"))
	sb.WriteString(valueStyle.Render(fmt.Sprintf("%.1f µC/cm²", m.material.Pr*100)) + "\n")
	sb.WriteString(labelStyle.Render("Ec (coercive):"))
	sb.WriteString(valueStyle.Render(fmt.Sprintf("%.2f MV/cm", m.material.Ec/1e8)) + "\n")
	sb.WriteString(labelStyle.Render("τ (switching):"))
	sb.WriteString(valueStyle.Render(fmt.Sprintf("%.1f ns", m.material.Tau*1e9)) + "\n")
	sb.WriteString(labelStyle.Render("Endurance:"))
	sb.WriteString(valueStyle.Render(fmt.Sprintf("%.0e cycles", m.material.EnduranceCycles)) + "\n\n")

	// Waveform info
	sb.WriteString(axisStyle.Render("─── Waveform ───") + "\n")
	sb.WriteString(labelStyle.Render("Type:"))
	sb.WriteString(highlightStyle.Render(m.waveform.String()) + "\n")
	sb.WriteString(labelStyle.Render("Frequency:"))
	sb.WriteString(valueStyle.Render(fmt.Sprintf("%.1f Hz", m.frequency)) + "\n")
	sb.WriteString(labelStyle.Render("Auto Mode:"))
	if m.autoMode {
		sb.WriteString(highlightStyle.Render("ON") + "\n")
	} else {
		sb.WriteString(axisStyle.Render("OFF") + "\n")
	}

	return infoPanelStyle.Render(sb.String())
}

// renderLevelBar renders the 30-level indicator bar
func (m Model) renderLevelBar() string {
	var sb strings.Builder
	sb.WriteString("  30 Levels\n")
	sb.WriteString("  ┌─┐\n")

	for i := 29; i >= 0; i-- {
		var char string
		var style lipgloss.Style

		if i == m.discreteLevel {
			char = "█"
			style = highlightStyle
		} else if i < m.discreteLevel {
			// Below current - gradient blue
			intensity := float64(i) / 29.0
			color := lipgloss.Color(fmt.Sprintf("#%02x%02x%02xff",
				int(50+intensity*150), int(50+intensity*150), 255))
			style = lipgloss.NewStyle().Foreground(color)
			char = "▓"
		} else {
			// Above current - gradient red
			intensity := float64(i) / 29.0
			color := lipgloss.Color(fmt.Sprintf("#ff%02x%02x%02x",
				255, int(255-intensity*200), int(255-intensity*200)))
			style = lipgloss.NewStyle().Foreground(color)
			char = "░"
		}

		sb.WriteString("  │" + style.Render(char) + "│")
		if i == 29 || i == 15 || i == 0 {
			sb.WriteString(fmt.Sprintf(" %2d", i+1))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("  └─┘\n")
	sb.WriteString(fmt.Sprintf("  %.1f bits", math.Log2(30)))

	return sb.String()
}

// renderStatusBar renders the bottom status bar
func (m Model) renderStatusBar() string {
	status := "●"
	if m.paused {
		status = highlightStyle.Render("⏸ PAUSED")
	} else {
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render("● RUNNING")
	}

	time := fmt.Sprintf("t = %.2fs", m.simTime)
	switchedFrac := fmt.Sprintf("Switched: %.1f%%", m.preisach.GetSwitchedFraction()*100)

	return fmt.Sprintf("  %s  │  %s  │  %s  │  Press [q] to quit",
		status, time, switchedFrac)
}

// Run starts the TUI application
func Run() error {
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
