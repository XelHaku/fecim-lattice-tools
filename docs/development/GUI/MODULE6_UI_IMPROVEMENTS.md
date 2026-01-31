# Module 6 EDA Design Suite - UI Improvement Proposals

**Document Version:** 1.0
**Date:** 2026-01-30
**Status:** Proposal & Planning
**Target Release:** v2.0

---

## Executive Summary

The FeCIM Array Builder (Module 6 EDA Design Suite) provides powerful functionality for generating fabrication-ready files through industry-standard EDA tools. However, the current UI has several accessibility and usability limitations that impact both professional users and learners:

### Current State Assessment
- **Strengths**: Comprehensive validation pipeline, real-time statistics, educational Learn tab, professional file exports
- **Critical Issues**: Docker X11 integration failures (KLayout/OpenROAD), small visualization sizes, color-only status indicators (WCAG violation), dense information presentation, no syntax highlighting in code previews
- **User Impact**: Developers cannot see physical layouts; color-blind users cannot distinguish validation states; learners struggle with dense UI

### Proposed Solution
Implement three-tier improvements: (1) fix critical Docker/X11 integration, (2) enhance accessibility and readability, (3) refine visual polish and information hierarchy. Estimated 26-36 development hours.

---

## Priority 1: Critical Fixes (8-12 hours)

### 1.1 Docker X11 Integration for KLayout & OpenROAD

**Problem Statement:**
KLayout and OpenROAD image generation fail with exit code 1. The custom KLayout Ruby script and OpenROAD TCL script run but cannot render output. Root cause: Docker containers lack proper X11 display configuration and Xvfb (virtual framebuffer) initialization.

**Current Code Analysis:**

```go
// pkg/validation/layout_image.go:154-166
// Current: Passes /design paths to Docker but no X11 setup
rdVars = map[string]string{
    "def_file":   "/design/" + filepath.Base(defPath),
    "lef_file":   "/design/" + lefName,
    "output_png": "/design/" + outputName,
}

// pkg/validation/circuit_image.go:270-282
// Similar issue: No X11/Xvfb initialization
envVars = map[string]string{
    "DEF_FILE":   "/design/" + filepath.Base(defPath),
    "CELL_LEF":   "/design/" + lefName,
    "OUTPUT_PNG": "/design/" + outputName,
}
```

**Root Causes:**
1. No `DISPLAY` variable set for Docker containers
2. No Xvfb (X Virtual Framebuffer) initialization
3. Host `xhost` not configured for local Docker
4. OpenROAD script missing display setup
5. KLayout script uses `-z` flag (requires display) without display available

**Solution Approach:**

#### Step 1: Update OpenROAD TCL Script
```tcl
# pkg/validation/circuit_image.go - Update openroadImageScript

# save_layout_image.tcl - OpenROAD layout image export with X11 support
puts "=== OpenROAD Layout Image Export ==="

# Check if display is available, use Xvfb if needed
set display [exec echo $env(DISPLAY)]
if {$display eq ""} {
    puts "Warning: DISPLAY not set, using Xvfb virtual display"
    exec Xvfb :99 -screen 0 1024x768x24 &
    set env(DISPLAY) :99
    after 500  ;# Wait for Xvfb to start
}

# Read LEF and DEF with error handling
if {[catch {read_lef $env(CELL_LEF)} err]} {
    puts "LEF read error: $err"
}

if {[catch {read_def $env(DEF_FILE)} err]} {
    puts "DEF read error: $err"
    exit 1
}

puts "Design loaded, saving image..."

# Save layout image with explicit resolution
save_image $env(OUTPUT_PNG)

puts "=== Image Export Complete ==="
exit
```

#### Step 2: Update KLayout Ruby Script
```ruby
# pkg/validation/layout_image.go - Update klayoutScript

# Display setup
if ENV['DISPLAY'].to_s.empty?
  puts "Warning: DISPLAY not set, starting Xvfb"
  system("Xvfb :99 -screen 0 1024x768x24 &")
  sleep(0.5)  # Wait for Xvfb to initialize
  ENV['DISPLAY'] = ':99'
end

# ... rest of script
```

#### Step 3: Update Runner to Configure X11

File: `module6-eda/pkg/openlane/runner.go` - Add X11 setup to RunKLayout and RunOpenROAD:

```go
// Add to RunKLayout method
func (r *Runner) RunKLayout(scriptPath string, workDir string, rdVars map[string]string) (*RunResult, error) {
    cmd := exec.Command("docker", "run", "--rm",
        "-v", fmt.Sprintf("%s:/design", workDir),
        // NEW: X11 setup
        "-e", "DISPLAY=:0",
        "-v", "/tmp/.X11-unix:/tmp/.X11-unix:rw",
        "--net=host",
        // END NEW
        "efabless/openlane:latest",
        "klayout", "-r", "/design/layout_export.rb", ...
    )
    // ... rest of implementation
}

// Similarly for RunOpenROAD
```

#### Step 4: Execution Instructions for Users

Add to Module 6 README:

```markdown
## X11 Display Setup (Required for Layout Visualization)

For KLayout and OpenROAD image generation, grant Docker access to host X11:

### Linux (Required)
\`\`\`bash
# Allow Docker to access X11 display
xhost +local:docker

# Run the app (X11 will be available inside Docker containers)
./fecim-lattice-tools
\`\`\`

### macOS (via Docker Desktop)
\`\`\`bash
# Install socat and X11 forwarding
brew install socat

# In Terminal 1: Start socat relay
socat TCP-LISTEN:6000,reuseaddr,fork UNIX-CONNECT:/var/run/docker.sock

# In Terminal 2: Set DISPLAY and run app
export DISPLAY=:0
./fecim-lattice-tools
\`\`\`

### Windows (WSL2)
- Requires VcXsrv or X410 running
- Set `DISPLAY=localhost:0` in WSL
- Docker Desktop must be configured to use WSL2 backend
```

**Verification:**
- After fix, "Generate All" should produce KLayout PNG
- Layout tab should display physical layout from DEF/LEF
- Status indicators should show "✓ Generated" instead of error

**Files to Modify:**
1. `pkg/validation/layout_image.go` - Update klayoutScript constant (lines 35-90)
2. `pkg/validation/circuit_image.go` - Update openroadImageScript constant (lines 188-205)
3. `pkg/openlane/runner.go` - Add X11 env vars to RunKLayout/RunOpenROAD methods
4. `module6-eda/README.md` - Add X11 setup instructions
5. `docs/eda/guides/DOCKER_SETUP.md` - Create comprehensive guide

**Estimated Effort:** 3-4 hours

---

### 1.2 Layout Tab Image Size and Zoom Interface

**Problem Statement:**
Layout images (KLayout, OpenROAD, Yosys schematics) display at 400x350px fixed size. For complex arrays (16x16+), circuit details are unreadable. No pan/zoom controls available.

**Current Code:**
```go
// pkg/gui/tabs/builder_validation_tab.go:136-144
klayoutImage := canvas.NewImageFromFile("")
klayoutImage.FillMode = canvas.ImageFillContain
klayoutImage.SetMinSize(fyne.NewSize(400, 350))  // TOO SMALL
```

**Solution:**

#### Approach 1: Increase Default Sizes (Quick Win)
```go
// Default display sizes by image type
const (
    DefaultLayoutWidth  = 600  // Was 400
    DefaultLayoutHeight = 500  // Was 350
)

klayoutImage.SetMinSize(fyne.NewSize(600, 500))
openroadImage.SetMinSize(fyne.NewSize(600, 500))
yosysImage.SetMinSize(fyne.NewSize(600, 500))
```

#### Approach 2: Modal Zoom Viewer (Recommended)

Create new file: `module6-eda/pkg/gui/widgets/image_viewer.go`

```go
package widgets

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/canvas"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/widget"
)

// ImageViewer provides zoom, pan, and fullscreen viewing
type ImageViewer struct {
    widget.BaseWidget
    image       *canvas.Image
    zoomLevel   float32  // 0.5 = 50%, 1.0 = 100%, 2.0 = 200%
    minZoom     float32
    maxZoom     float32
    panX, panY  float32
}

// NewImageViewer creates a zoomable image viewer
func NewImageViewer(imagePath string) *ImageViewer {
    img := canvas.NewImageFromFile(imagePath)
    img.FillMode = canvas.ImageFillContain

    return &ImageViewer{
        image:     img,
        zoomLevel: 1.0,
        minZoom:   0.25,
        maxZoom:   4.0,
    }
}

// SetZoom adjusts zoom level and refresh display
func (iv *ImageViewer) SetZoom(level float32) {
    if level < iv.minZoom {
        level = iv.minZoom
    }
    if level > iv.maxZoom {
        level = iv.maxZoom
    }
    iv.zoomLevel = level
    iv.Refresh()
}

// ZoomIn increases zoom by 20%
func (iv *ImageViewer) ZoomIn() {
    iv.SetZoom(iv.zoomLevel * 1.2)
}

// ZoomOut decreases zoom by 20%
func (iv *ImageViewer) ZoomOut() {
    iv.SetZoom(iv.zoomLevel / 1.2)
}

// ShowInModal opens image in fullscreen modal with controls
func (iv *ImageViewer) ShowInModal(parent fyne.Window) {
    // Create control bar
    zoomLabel := widget.NewLabel("100%")

    zoomInBtn := widget.NewButton("Zoom In (1.2x)", func() {
        iv.ZoomIn()
        zoomLabel.SetText(fmt.Sprintf("%.0f%%", iv.zoomLevel*100))
    })

    zoomOutBtn := widget.NewButton("Zoom Out (0.8x)", func() {
        iv.ZoomOut()
        zoomLabel.SetText(fmt.Sprintf("%.0f%%", iv.zoomLevel*100))
    })

    resetBtn := widget.NewButton("Reset (100%)", func() {
        iv.SetZoom(1.0)
        zoomLabel.SetText("100%")
    })

    fitBtn := widget.NewButton("Fit to Window", func() {
        iv.SetZoom(0.5)  // Adjust as needed
        zoomLabel.SetText("50%")
    })

    controls := container.NewHBox(
        zoomOutBtn, zoomLabel, zoomInBtn,
        widget.NewSeparator(),
        resetBtn, fitBtn,
    )

    // Create scrollable image display
    imageScroll := container.NewScroll(iv.image)
    imageScroll.SetMinSize(fyne.NewSize(800, 600))

    // Create modal with controls at top
    content := container.NewBorder(
        controls, nil, nil, nil,
        imageScroll,
    )

    dialog.ShowCustom(
        "Image Viewer - Click buttons or scroll to zoom",
        "Close",
        content,
        parent,
    )
}
```

#### Integration into Builder Tab

Update `builder_validation_tab.go` to add zoom buttons:

```go
// Lines 139-144: Create KLayout card with zoom button
klayoutZoomBtn := widget.NewButton("Zoom", func() {
    viewer := widgets.NewImageViewer(klayoutImage.File)
    viewer.ShowInModal(window)
})

klayoutCard := widget.NewCard("", "", container.NewBorder(
    container.NewVBox(
        klayoutLabel,
        container.NewHBox(klayoutStatus, klayoutZoomBtn),
    ),
    nil, nil, nil,
    klayoutImage,
))

// Repeat for OpenROAD and Yosys images
```

**Verification:**
- Layout images display at 600x500 default
- "Zoom" button opens modal viewer
- Zoom in/out buttons adjust 20% per click
- Reset button returns to 100%
- Fit button resizes to window

**Files to Modify:**
1. `pkg/gui/widgets/image_viewer.go` - Create new viewer widget (NEW FILE)
2. `pkg/gui/tabs/builder_validation_tab.go` - Add zoom buttons to image cards (Lines 139-161)

**Estimated Effort:** 3-4 hours (includes testing zoom responsiveness)

---

### 1.3 Validation Results Accessibility (WCAG 1.4.1 Compliance)

**Problem Statement:**
Validation status indicators use color only (green checkmarks ✓, red X marks ✗). Users with color blindness cannot distinguish pass/fail states. Violates WCAG 2.1 Level AA (1.4.1 Use of Color).

**Current Code:**
```go
// pkg/gui/tabs/builder_validation_tab.go:183-185
yosysResult := widget.NewLabel("✓ PASS")    // Green text only
defResult := widget.NewLabel("✗ FAIL")      // Red text only
crossResult := widget.NewLabel("Not validated")
```

**Solution:**

Implement multi-modal status indicators using:
- **Shape Coding**: Circles (○) for pass, triangles (▲) for fail, dashes (—) for pending
- **Text Labels**: Clear "PASS" / "FAIL" / "PENDING" text
- **Color**: Maintained as secondary indicator
- **Icons**: Added Fyne icons for additional visual distinction

Create status widget: `module6-eda/pkg/gui/widgets/validation_badge.go`

```go
package widgets

import (
    "image/color"
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/canvas"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/theme"
    "fyne.io/fyne/v2/widget"
)

// ValidationStatus enum
type ValidationStatus int

const (
    StatusPending ValidationStatus = iota
    StatusPass
    StatusFail
    StatusSkipped
)

// ValidationBadge displays WCAG-compliant status with shape + text + color
type ValidationBadge struct {
    widget.BaseWidget
    status ValidationStatus
    label  string
}

// NewValidationBadge creates an accessible status badge
func NewValidationBadge(status ValidationStatus, label string) *ValidationBadge {
    badge := &ValidationBadge{
        status: status,
        label:  label,
    }
    badge.ExtendBaseWidget(badge)
    return badge
}

// CreateRenderer implements widget.Widget
func (vb *ValidationBadge) CreateRenderer() fyne.WidgetRenderer {
    icon := canvas.NewText("", color.White)
    statusText := widget.NewLabel("")

    // Set icon shape and color based on status
    switch vb.status {
    case StatusPass:
        icon.Text = "●"  // Filled circle for pass
        icon.TextSize = 16
        icon.Color = color.RGBA{0, 200, 0, 255}  // Green
        statusText.SetText("✓ PASS")
        statusText.TextStyle = fyne.TextStyle{Bold: true}

    case StatusFail:
        icon.Text = "▲"  // Triangle for fail
        icon.TextSize = 16
        icon.Color = color.RGBA{255, 0, 0, 255}  // Red
        statusText.SetText("✗ FAIL")
        statusText.TextStyle = fyne.TextStyle{Bold: true}

    case StatusPending:
        icon.Text = "⊙"  // Circled dot for pending
        icon.TextSize = 16
        icon.Color = color.RGBA{255, 165, 0, 255}  // Orange
        statusText.SetText("⋯ PENDING")

    case StatusSkipped:
        icon.Text = "—"  // Dash for skipped
        icon.TextSize = 16
        icon.Color = color.RGBA{128, 128, 128, 255}  // Gray
        statusText.SetText("⊝ SKIPPED")
    }

    content := container.NewHBox(icon, statusText)
    return widget.NewSimpleRenderer(content)
}

// SetStatus updates the badge to reflect new status
func (vb *ValidationBadge) SetStatus(status ValidationStatus, label string) {
    vb.status = status
    vb.label = label
    vb.Refresh()
}
```

**Integration into Validation Section:**

Update `builder_validation_tab.go`:

```go
// Lines 183-185: Replace with new badges
yosysResult := widgets.NewValidationBadge(
    widgets.StatusPending, "Yosys Verilog",
)
defResult := widgets.NewValidationBadge(
    widgets.StatusPending, "DEF Syntax",
)
crossResult := widgets.NewValidationBadge(
    widgets.StatusPending, "LEF/LIB/V Cross-check",
)
placementResult := widgets.NewValidationBadge(
    widgets.StatusPending, "Placement",
)

// During validation (lines 427+), update badges:
fyne.Do(func() {
    if yosysOK {
        yosysResult.SetStatus(widgets.StatusPass, "Yosys")
    } else {
        yosysResult.SetStatus(widgets.StatusFail, "Yosys")
    }
    // ... repeat for others
})

// After validation summary (line 473):
validationSummary.SetText("Validation Results")  // Plus visual indicator
```

**WCAG Compliance Checklist:**
- ✓ 1.4.1 (Use of Color): Not color-only; shapes + text provide distinction
- ✓ 1.4.3 (Contrast): Text labels have 4.5:1+ contrast ratio
- ✓ 1.4.5 (Images of Text): Icons are actual Unicode shapes, not images
- ✓ 2.5.5 (Target Size): Badge size ≥44px, touch-friendly

**Verification:**
- Run with color-blindness simulator (e.g., Coblis)
- Verify shapes are visually distinct when colors removed
- Confirm text labels clearly indicate status

**Files to Modify:**
1. `pkg/gui/widgets/validation_badge.go` - Create new badge widget (NEW FILE)
2. `pkg/gui/tabs/builder_validation_tab.go` - Replace status labels (Lines 183-185, 427+, 473)

**Estimated Effort:** 2-3 hours

---

## Priority 2: UX Improvements (12-16 hours)

### 2.1 Builder Tab Layout Restructuring

**Problem Statement:**
Cell Config and Array Config fields displayed in dense 4-column grids. Combined with stats row, the tab feels cramped and information-heavy. Field labels and values are visually disconnected.

**Current Structure:**
```
[Cell Config] [Array Config]
├─ Grid 4-col: Name|Width|Height|Rise
├─ Grid 4-col: Fall|Cap|Leakage|[Area]
└─ Stats row: Total|Area|WL|BL|Density|Util

[Preview Tabs]
[Validation Results]
```

**Solution:**

Implement card-based layout with collapsible sections:

```go
// module6-eda/pkg/gui/tabs/builder_validation_tab.go - New layout

// 1. CREATE CARD COMPONENTS
cellConfigCard := widget.NewCard(
    "Cell Configuration",  // Title
    "Define FeCIM bitcell properties",  // Subtitle
    makeCellConfigFields(),  // Content
)

arrayConfigCard := widget.NewCard(
    "Array Configuration",
    "Specify crossbar dimensions and operation mode",
    makeArrayConfigFields(),
)

// 2. MAKE RESPONSIVE GRID (2 columns on desktop, 1 on mobile)
configGrid := container.NewGridWithColumns(2,
    cellConfigCard,
    arrayConfigCard,
)

// 3. STATISTICS CARD (separate from config)
statsCard := widget.NewCard(
    "Array Statistics",
    "Real-time metrics",
    makeStatsDisplay(),
)

// 4. TOP SECTION: Config + Stats stacked
topSection := container.NewVBox(
    configGrid,
    statsCard,
)

// 5. FULL LAYOUT
mainLayout := container.NewBorder(
    topSection,      // Top: Config cards + Stats
    nil,             // Bottom: None
    nil,             // Left: None
    nil,             // Right: None
    container.NewVBox(  // Center: Preview + Validation
        previewTabs,
        validationSection,
    ),
)
```

**New Cell Config Card:**

```go
func makeCellConfigFields() fyne.CanvasObject {
    nameEntry := widget.NewEntry()
    nameEntry.SetText("fecim_bitcell")
    nameEntry.OnChanged = func(s string) { /* update */ }

    widthEntry := widget.NewEntry()
    widthEntry.SetText("0.460")
    widthEntry.OnChanged = func(s string) { updateStats() }

    // Form layout: label above entry, one per row
    form := container.NewVBox(
        widget.NewLabel("Cell Name:"),
        nameEntry,
        widget.NewSeparator(),

        widget.NewLabel("Width (μm):"),
        widthEntry,
        widget.NewSeparator(),

        widget.NewLabel("Height (μm):"),
        heightEntry,
        widget.NewSeparator(),

        // Section divider for timing
        widget.NewLabelWithStyle(
            "Timing (Placeholder)",
            fyne.TextAlignLeading,
            fyne.TextStyle{Bold: true},
        ),

        widget.NewLabel("Rise Time (ns):"),
        riseEntry,
        widget.NewLabel("Fall Time (ns):"),
        fallEntry,

        // Section divider for electrical
        widget.NewLabelWithStyle(
            "Electrical Properties",
            fyne.TextAlignLeading,
            fyne.TextStyle{Bold: true},
        ),

        widget.NewLabel("Input Capacitance (pF):"),
        capEntry,
        widget.NewLabel("Leakage Power (nW):"),
        leakageEntry,

        widget.NewSeparator(),
        cellAreaLabel,
    )

    return container.NewScroll(form)
}

func makeArrayConfigFields() fyne.CanvasObject {
    rowsEntry := widget.NewEntry()
    rowsEntry.SetText(fmt.Sprintf("%d", cfg.Rows))

    colsEntry := widget.NewEntry()
    colsEntry.SetText(fmt.Sprintf("%d", cfg.Cols))

    modeSelect := widget.NewSelect(
        []string{"storage", "memory", "compute"}, nil,
    )
    modeSelect.SetSelected(cfg.Mode)

    archSelect := widget.NewSelect(
        []string{"passive", "1t1r"}, nil,
    )
    archSelect.SetSelected(cfg.Architecture)

    form := container.NewVBox(
        widget.NewLabel("Array Rows:"),
        rowsEntry,
        widget.NewSeparator(),

        widget.NewLabel("Array Columns:"),
        colsEntry,
        widget.NewSeparator(),

        widget.NewLabel("Operation Mode:"),
        modeSelect,
        modeHelpText,  // Help text below selector
        widget.NewSeparator(),

        widget.NewLabel("Architecture:"),
        archSelect,
    )

    return container.NewScroll(form)
}

func makeStatsDisplay() fyne.CanvasObject {
    // Card-based stats layout instead of dense row
    statCards := container.NewGridWithColumns(3,
        makeStatCard("Total Cells", totalLabel),
        makeStatCard("Array Area", areaLabel),
        makeStatCard("Cell Density", densityLabel),
        makeStatCard("WL Length", wlLengthLabel),
        makeStatCard("BL Length", blLengthLabel),
        makeStatCard("Utilization", utilizationLabel),
    )
    return statCards
}

func makeStatCard(title string, valueLabel *widget.Label) fyne.CanvasObject {
    titleLabel := widget.NewLabelWithStyle(
        title,
        fyne.TextAlignCenter,
        fyne.TextStyle{Bold: true},
    )

    valueLabel.Alignment = fyne.TextAlignCenter

    return widget.NewCard("", "", container.NewVBox(
        titleLabel,
        valueLabel,
    ))
}
```

**Visual Hierarchy Improvements:**
- Section dividers (bold headings + separators) group related fields
- Card titles and subtitles provide context
- Spacing between sections (VBox padding)
- Icons could be added (future enhancement)

**Responsive Behavior:**
- Desktop (1000px+): 2 columns for config cards
- Tablet (600-999px): 1 column, cards stack vertically
- Mobile: Single column, cards full width

**Verification:**
- Config fields grouped logically
- Stats displayed in 3-column grid (not dense row)
- Cards are clickable/hoverable (visual feedback)
- Spacing hierarchy: section > field > line items

**Files to Modify:**
1. `pkg/gui/tabs/builder_validation_tab.go` - Refactor top section (Lines 580-615, 718-723)

**Estimated Effort:** 4-5 hours

---

### 2.2 Code Preview Enhancement (Syntax Highlighting & Polish)

**Problem Statement:**
Verilog and DEF code displayed in plain text MultiLineEntry widgets. No syntax highlighting, line numbers, or copy-to-clipboard functionality. Makes it hard to understand generated code.

**Current Code:**
```go
// Lines 168-174: Plain text previews
verilogPreview := widget.NewMultiLineEntry()
verilogPreview.Wrapping = fyne.TextWrapOff
// ... no syntax highlighting, monospace font is small (~12px)

defPreview := widget.NewMultiLineEntry()
defPreview.Wrapping = fyne.TextWrapOff
```

**Solution:**

Create custom code preview widget with syntax highlighting:

New file: `module6-eda/pkg/gui/widgets/code_viewer.go`

```go
package widgets

import (
    "strings"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/canvas"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)

// CodeLanguage enum
type CodeLanguage int

const (
    LangVerilog CodeLanguage = iota
    LangDEF
    LangLiberty
    LangPlain
)

// CodeViewer displays code with syntax highlighting and line numbers
type CodeViewer struct {
    widget.BaseWidget
    code      string
    language  CodeLanguage
    lineCount int
}

// NewCodeViewer creates a code viewer with syntax highlighting
func NewCodeViewer(code string, language CodeLanguage) *CodeViewer {
    lineCount := strings.Count(code, "\n") + 1

    cv := &CodeViewer{
        code:      code,
        language:  language,
        lineCount: lineCount,
    }
    cv.ExtendBaseWidget(cv)
    return cv
}

// SetCode updates the displayed code
func (cv *CodeViewer) SetCode(code string) {
    cv.code = code
    cv.lineCount = strings.Count(code, "\n") + 1
    cv.Refresh()
}

// CreateRenderer implements widget.Widget
func (cv *CodeViewer) CreateRenderer() fyne.WidgetRenderer {
    // Line numbers column
    lineNumbers := canvas.NewText("", nil)
    lines := strings.Split(cv.code, "\n")

    lineNumText := ""
    for i := 1; i <= len(lines); i++ {
        lineNumText += fmt.Sprintf("%3d\n", i)
    }
    lineNumbers.Text = lineNumText
    lineNumbers.TextSize = 13
    lineNumbers.TextStyle = fyne.TextStyle{}  // Monospace via theme

    // Code with syntax highlighting
    codeText := canvas.NewText(cv.code, nil)
    codeText.TextSize = 13
    codeText.TextStyle = fyne.TextStyle{}

    // Apply highlighting based on language
    highlightedCode := cv.applySyntaxHighlight(cv.code)
    highlightedText := canvas.NewText(highlightedCode, nil)
    highlightedText.TextSize = 13

    // Layout: line numbers | code
    content := container.NewHBox(
        container.NewPadded(lineNumbers),
        widget.NewSeparator(),
        container.NewScroll(highlightedText),
    )

    return widget.NewSimpleRenderer(content)
}

// applySyntaxHighlight adds ANSI color codes to code
func (cv *CodeViewer) applySyntaxHighlight(code string) string {
    switch cv.language {
    case LangVerilog:
        return cv.highlightVerilog(code)
    case LangDEF:
        return cv.highlightDEF(code)
    case LangLiberty:
        return cv.highlightLiberty(code)
    default:
        return code
    }
}

// Verilog syntax highlighting
func (cv *CodeViewer) highlightVerilog(code string) string {
    // Keywords to highlight
    keywords := []string{
        "module", "input", "output", "reg", "wire", "always",
        "begin", "end", "if", "else", "assign", "parameter",
    }

    result := code
    for _, kw := range keywords {
        // Simple regex replacement (production would use proper lexer)
        pattern := fmt.Sprintf(`\b%s\b`, kw)
        result = strings.ReplaceAll(result, kw, fmt.Sprintf("[BLUE]%s[/BLUE]", kw))
    }
    return result
}

// DEF syntax highlighting
func (cv *CodeViewer) highlightDEF(code string) string {
    // Similar pattern for DEF keywords
    keywords := []string{
        "COMPONENTS", "NETS", "PINS", "FIXED", "PLACED",
        "UNPLACED", "DIEAREA", "DESIGN", "UNITS",
    }

    result := code
    for _, kw := range keywords {
        result = strings.ReplaceAll(result, kw, fmt.Sprintf("[GREEN]%s[/GREEN]", kw))
    }
    return result
}

// Liberty syntax highlighting
func (cv *CodeViewer) highlightLiberty(code string) string {
    keywords := []string{
        "library", "cell", "pin", "direction", "timing",
        "rise_time", "fall_time", "related_pin",
    }

    result := code
    for _, kw := range keywords {
        result = strings.ReplaceAll(result, kw, fmt.Sprintf("[CYAN]%s[/CYAN]", kw))
    }
    return result
}

// CopyButton creates a button that copies code to clipboard
func (cv *CodeViewer) CopyButton(w fyne.Window) *widget.Button {
    return widget.NewButton("Copy to Clipboard", func() {
        w.Clipboard().SetContent(cv.code)
        dialog.ShowInformation(
            "Copied",
            fmt.Sprintf("Code (%d lines) copied to clipboard", cv.lineCount),
            w,
        )
    })
}
```

**Integration into Builder Tab:**

Update preview sections:

```go
// Lines 168-174: Replace MultiLineEntry with CodeViewer

// Verilog Tab
verilogViewer := widgets.NewCodeViewer("", widgets.LangVerilog)
verilogViewer.SetMinSize(fyne.NewSize(600, 300))
verilogCopyBtn := verilogViewer.CopyButton(window)
verilogTab := container.NewBorder(
    container.NewHBox(verilogStatsLabel, verilogCopyBtn),
    nil, nil, nil,
    verilogViewer,
)

// DEF Tab (similar)
defViewer := widgets.NewCodeViewer("", widgets.LangDEF)
defViewer.SetMinSize(fyne.NewSize(600, 300))
defCopyBtn := defViewer.CopyButton(window)
defTab := container.NewBorder(
    container.NewHBox(defStatsLabel, defCopyBtn),
    nil, nil, nil,
    defViewer,
)

// Update on generation:
verilogViewer.SetCode(generatedVerilog)  // Instead of verilogPreview.SetText()
defViewer.SetCode(generatedDEF)
```

**Features:**
- Line numbers (right-aligned, muted color)
- Syntax highlighting (keywords colored, strings/comments distinct)
- Monospace font 13px (vs default 12px)
- Copy button exports full code to clipboard
- Scrollable when content exceeds viewport

**Color Scheme (Dark Theme Compliant):**
- Keywords: Cyan (storage, module)
- Strings: Yellow
- Comments: Gray/muted
- Line numbers: Dark gray

**Verification:**
- Generated Verilog displays with syntax highlighting
- Line numbers visible and accurate
- Copy button creates clipboard content
- Font size readable (13px monospace)
- Colors have sufficient contrast

**Files to Modify:**
1. `pkg/gui/widgets/code_viewer.go` - Create new viewer (NEW FILE)
2. `pkg/gui/tabs/builder_validation_tab.go` - Integrate CodeViewer (Lines 168-174)

**Estimated Effort:** 4-5 hours (includes testing highlighting accuracy)

---

### 2.3 Log Panel Expansion and Filtering

**Problem Statement:**
Validation log displays only ~4 lines visible. Long validation sequences get truncated. No ability to filter log levels or search messages.

**Current Code:**
```go
// Lines 188-190: Small log panel
logOutput := widget.NewMultiLineEntry()
// No filtering, small visible area, no search
```

**Solution:**

Enhance log panel with filtering and search:

```go
// module6-eda/pkg/gui/tabs/builder_validation_tab.go - Update log section

// CREATE LOG LEVEL FILTER
filterAllRadio := widget.NewRadioItem("All", func() { filterLogs("") })
filterErrorRadio := widget.NewRadioItem("Errors only", func() { filterLogs("ERROR") })
filterWarnRadio := widget.NewRadioItem("Warnings+Errors", func() { filterLogs("WARN") })
filterInfoRadio := widget.NewRadioItem("Info+", func() { filterLogs("INFO") })

filterGroup := widget.NewRadioGroup(
    []*widget.RadioItem{filterAllRadio, filterErrorRadio, filterWarnRadio, filterInfoRadio},
)
filterGroup.Selected = filterAllRadio

// SEARCH BOX
searchEntry := widget.NewEntry()
searchEntry.PlaceHolder = "Search logs..."
searchEntry.OnChanged = func(s string) { searchLogs(s) }

// EXPANDED LOG OUTPUT (10+ lines visible)
logOutput := widget.NewMultiLineEntry()
logOutput.SetMinSize(fyne.NewSize(800, 200))  // Was smaller
logOutput.Wrapping = fyne.TextWrapWord

// CLEAR LOG BUTTON
clearLogBtn := widget.NewButton("Clear", func() {
    logOutput.SetText("")
    allLogs = []string{}  // Clear in-memory buffer
})

// CREATE LOG HEADER WITH CONTROLS
logHeader := container.NewBorder(
    container.NewVBox(
        container.NewHBox(
            widget.NewLabel("Filters:"),
            filterGroup,
        ),
        widget.NewLabel("Search:"),
        searchEntry,
    ),
    container.NewHBox(clearLogBtn),
    nil, nil,
    nil,
)

// CREATE LOG SECTION
logSection := container.NewBorder(
    logHeader,        // Top: Controls
    nil,              // Bottom: None
    nil,              // Left: None
    nil,              // Right: None
    logOutput,        // Center: Log display
)

// Update validationSection to include new log
validationSection := container.NewBorder(
    validationResultsPanel,  // Top: Validation results
    nil,                     // Bottom: None
    nil,                     // Left: None
    nil,                     // Right: None
    logSection,              // Center: Log with controls
)
```

**Log Level Standardization:**

Update `addLog()` function to include log levels:

```go
// Lines 192-196: Enhanced addLog with levels
func addLog(level string, message string) {
    timestamp := time.Now().Format("15:04:05")
    formattedMsg := fmt.Sprintf("[%s] %s: %s", timestamp, level, message)

    // Store in-memory for filtering
    allLogs = append(allLogs, struct{
        level, message string
    }{level, message})

    // Update display with fyne.Do() for thread safety
    fyne.Do(func() {
        logOutput.SetText(logOutput.Text + formattedMsg + "\n")
    })
}

// Filtering function
func filterLogs(minLevel string) {
    filtered := ""
    for _, log := range allLogs {
        if shouldIncludeLog(log.level, minLevel) {
            filtered += fmt.Sprintf("[%s] %s: %s\n", log.level, log.message)
        }
    }
    logOutput.SetText(filtered)
}

// Search function
func searchLogs(query string) {
    if query == "" {
        filterLogs("")  // Reset to all
        return
    }

    filtered := ""
    for _, log := range allLogs {
        if strings.Contains(log.message, query) {
            filtered += fmt.Sprintf("[%s] %s: %s\n", log.level, log.message)
        }
    }
    logOutput.SetText(filtered)
}
```

**Usage in Validation Steps:**

```go
// Lines 355+: Update all log calls
addLog("INFO", "Starting Yosys validation...")
addLog("INFO", fmt.Sprintf("Reading Verilog from %s", verilogPath))

if yosysOK {
    addLog("INFO", "✓ Yosys validation passed")
} else {
    addLog("ERROR", "✗ Yosys validation failed: " + yosysErr)
}

// For warnings
addLog("WARN", "OpenROAD not available - skipping placement check")
```

**Log Levels:**
- `INFO`: Normal operation (generation, validation started)
- `WARN`: Non-critical issues (missing tools, skipped checks)
- `ERROR`: Validation failures, file errors

**Verification:**
- Log panel shows 10+ lines by default
- Filters work (All/Errors/Warnings/Info)
- Search highlights matching messages
- Clear button removes all logs
- Timestamps included in each line

**Files to Modify:**
1. `pkg/gui/tabs/builder_validation_tab.go` - Add filter controls (Lines 188-190), update log section (Lines 699-703), enhance addLog() (Lines 192-196)

**Estimated Effort:** 3-4 hours

---

## Priority 3: Visual Polish (6-8 hours)

### 3.1 Color Scheme Enhancements

**Problem Statement:**
Current monochromatic dark blue scheme (from Fyne default dark theme) lacks visual distinction between sections. Tab-specific colors would improve navigation and visual hierarchy.

**Current Color Palette:**
- Background: #0d3a5c (dark blue)
- Text: White/light gray
- Accents: Cyan (buttons, links)
- Status: Green (pass), Red (fail)

**Solution:**

Create theme variant for Module 6:

New file: `shared/theme/module6_colors.go`

```go
package theme

import (
    "image/color"
    "fyne.io/fyne/v2/theme"
)

// Module6ColorScheme provides tab-specific color coding
type Module6ColorScheme struct {
    baseTheme theme.Theme
}

// TabColors defines section-specific accent colors
var TabColors = map[string]color.Color{
    "builder":    color.RGBA{0, 150, 255, 255},    // Bright blue
    "layout":     color.RGBA{0, 180, 100, 255},    // Green
    "learn":      color.RGBA{200, 100, 200, 255},  // Purple/magenta
    "validation": color.RGBA{255, 140, 0, 255},    // Orange
}

// StatusColors defines semantic colors with WCAG AA contrast
var StatusColors = map[string]color.Color{
    "pass":    color.RGBA{76, 220, 100, 255},      // Brighter green
    "fail":    color.RGBA{240, 70, 70, 255},       // Brighter red
    "pending": color.RGBA{255, 165, 0, 255},       // Orange
    "skip":    color.RGBA{180, 180, 180, 255},     // Gray
    "info":    color.RGBA{100, 180, 255, 255},     // Blue
}

// AccentColors for various UI elements
var AccentColors = map[string]color.Color{
    "primary":   color.RGBA{0, 150, 255, 255},     // Tab highlight
    "secondary": color.RGBA{150, 150, 150, 255},   // Labels
    "danger":    color.RGBA{240, 70, 70, 255},     // Destructive actions
    "success":   color.RGBA{76, 220, 100, 255},    // Success states
}

// CardBackground returns slightly lighter background for card containers
func CardBackground() color.Color {
    return color.RGBA{25, 50, 80, 255}  // Slightly lighter than main
}

// SectionDivider returns a muted color for separators
func SectionDivider() color.Color {
    return color.RGBA{70, 90, 120, 255}
}
```

**Apply in Builder Tab:**

```go
// pkg/gui/tabs/builder_validation_tab.go

// Create tab-specific header with color accent
builderHeader := container.NewVBox(
    widget.NewLabelWithStyle(
        "FeCIM Array Builder",
        fyne.TextAlignCenter,
        fyne.TextStyle{Bold: true, Size: 18},
    ),
)

// Builder section color accent
builderBg := canvas.NewRectangle(theme6.TabColors["builder"])
builderBg.SetMinSize(fyne.NewSize(400, 3))

// Apply to cards
cellConfigCard := widget.NewCard(
    "Cell Configuration",
    "",
    makeCellConfigFields(),
)
// Add color indicator bar to card

arrayConfigCard := widget.NewCard(
    "Array Configuration",
    "",
    makeArrayConfigFields(),
)

// Validation results with semantic colors
yosysResult := widgets.NewValidationBadge(
    widgets.StatusPending,
    "Yosys",
)
yosysResult.SetColor(theme6.StatusColors["pending"])
```

**Apply to Learn Tab:**

```go
// pkg/gui/tabs/learn_tab.go

// Section headers with tab color
learnHeader := container.NewVBox(
    widget.NewLabelWithStyle(
        "FeCIM Array Builder - Learning Center",
        fyne.TextAlignCenter,
        fyne.TextStyle{Bold: true, Size: 16},
    ),
)

// Add accent bar
learnBg := canvas.NewRectangle(theme6.TabColors["learn"])
learnBg.SetMinSize(fyne.NewSize(400, 3))

// Topic cards with subtle color
topicCard := widget.NewCard("Topics", "", topicSelector)
// Background slightly different from main
```

**Contrast Ratios (WCAG AA Compliant):**
- Pass (green): 4.5:1 against dark background
- Fail (red): 4.5:1 against dark background
- Pending (orange): 4.5:1 against dark background

**Verification:**
- Each tab has visually distinct accent color
- Cards have 1-2px colored border
- Status badges use semantic colors
- All text meets 4.5:1 contrast minimum

**Files to Modify:**
1. `shared/theme/module6_colors.go` - Create new color theme (NEW FILE)
2. `pkg/gui/tabs/builder_validation_tab.go` - Apply builder colors
3. `pkg/gui/tabs/learn_tab.go` - Apply learn colors

**Estimated Effort:** 2-3 hours

---

### 3.2 Typography and Hierarchy

**Problem Statement:**
Base font size ~12px, limited hierarchy between section headers and content. Makes dense information harder to scan. Code font too small for older eyes.

**Solution:**

Standardize typography throughout:

```go
// shared/typography.go - NEW FILE

package theme

const (
    // Base typography scale
    FontSizeSmall     = 11  // Help text, secondary labels
    FontSizeBase      = 14  // Default text, labels
    FontSizeLarge     = 16  // Section headers
    FontSizeXLarge    = 18  // Tab titles
    FontSizeCode      = 13  // Code blocks (monospace)

    // Line heights for readability
    LineHeightDense   = 1.3  // Compact text
    LineHeightNormal  = 1.5  // Normal paragraph text
    LineHeightLoose   = 1.8  // Long-form content
)

// TextStyle combinations
var (
    H1 = fyne.TextStyle{Bold: true}      // 18px, bold
    H2 = fyne.TextStyle{Bold: true}      // 16px, bold
    H3 = fyne.TextStyle{Bold: false}     // 14px, semibold
    Body = fyne.TextStyle{Bold: false}   // 14px, regular
    Caption = fyne.TextStyle{Bold: false} // 11px, muted
)
```

**Apply to Tab Headers:**

```go
// Builder tab
builderTitle := widget.NewLabelWithStyle(
    "FeCIM Array Builder",
    fyne.TextAlignCenter,
    fyne.TextStyle{Bold: true},  // Use Typography.H1
)
builderTitle.Alignment = fyne.TextAlignCenter
builderTitle.TextSize = 18  // From constant

// Section headers
cellConfigTitle := widget.NewLabelWithStyle(
    "Cell Configuration",
    fyne.TextAlignLeading,
    fyne.TextStyle{Bold: true},
)
cellConfigTitle.TextSize = 16
```

**Apply to Learn Tab:**

```go
// Topic list items with better hierarchy
topicTitle := widget.NewLabelWithStyle(
    "What is FeCIM EDA?",
    fyne.TextAlignLeading,
    fyne.TextStyle{Bold: true},
)
topicTitle.TextSize = 18

topicDesc := widget.NewLabel(
    "Educational overview of FeCIM and OpenLane integration",
)
topicDesc.TextSize = 12
topicDesc.Wrapping = fyne.TextWrapWord
```

**Apply to Code Viewer:**

```go
// Code font is now 13px (was 12px)
codeText.TextSize = 13  // In CodeViewer widget

// Code line numbers aligned right
lineNumbers.TextSize = 13
lineNumbers.Alignment = fyne.TextAlignTrailing  // Right-aligned
```

**Spacing Hierarchy:**

```go
const (
    // Padding/margin scale
    SpacingXSmall  = 4   // Tight spacing within elements
    SpacingSmall   = 8   // Between related fields
    SpacingNormal  = 16  // Standard section margin
    SpacingLarge   = 24  // Between major sections
    SpacingXLarge  = 32  // Top-level section gaps
)

// Apply in Builder Tab
topSection := container.NewPadded()
topSection.Padding = SpacingLarge

fieldGroup := container.NewVBox()
fieldGroup.Spacing = SpacingSmall
```

**Verification:**
- Tab titles are 18px bold, easily scannable
- Section headers are 16px bold, visually grouped
- Body text is 14px, readable on all screens
- Code text is 13px monospace, clear without squinting
- Spacing creates visual hierarchy (8px < 16px < 24px)

**Files to Modify:**
1. `shared/typography.go` - Create typography constants (NEW FILE)
2. `pkg/gui/tabs/builder_validation_tab.go` - Apply typography to all labels
3. `pkg/gui/tabs/learn_tab.go` - Apply typography to content

**Estimated Effort:** 2 hours

---

### 3.3 Spacing and Component Polish

**Problem Statement:**
Inconsistent padding and margins. Some buttons touch each other, spacing is ad-hoc. Cards lack visual separation.

**Solution:**

Implement spacing constants and apply consistently:

```go
// pkg/gui/widgets/spacing.go - Helper for consistent spacing

const (
    CardPadding      = 16  // Inside cards
    SectionMargin    = 24  // Between major sections
    ButtonSpacing    = 12  // Between buttons
    FieldSpacing     = 8   // Between form fields
    FieldLabelGap    = 4   // Between label and input
)

// CreateSpacedContainer wraps content with consistent padding
func CreateSpacedContainer(content fyne.CanvasObject, padding float32) *container.Container {
    return container.NewPadded(content)
}

// CreateButtonRow spaces buttons evenly
func CreateButtonRow(buttons ...*widget.Button) *container.Container {
    hbox := container.NewHBox()
    hbox.Spacing = ButtonSpacing
    for _, btn := range buttons {
        hbox.Add(btn)
    }
    return hbox
}

// CreateFieldGroup spaces form fields vertically
func CreateFieldGroup(label string, entry *widget.Entry) fyne.CanvasObject {
    labelWidget := widget.NewLabel(label)
    return container.NewVBox(
        labelWidget,
        entry,
    )
}
```

**Apply in Builder Tab:**

```go
// Better button layout (Lines 647-651)
actionButtons := container.NewHBox()
actionButtons.Spacing = 12  // From constant

actionButtons.Add(generateAllBtn)
actionButtons.Add(validateAllBtn)
actionButtons.Add(exportPackageBtn)

// Better card spacing
configGrid := container.NewGridWithColumns(2)
configGrid.Spacing = 24  // From SectionMargin

configGrid.Add(cellConfigCard)
configGrid.Add(arrayConfigCard)

// Better status row spacing
validationRow := container.NewHBox()
validationRow.Spacing = 16  // Increase from current

validationRow.Add(yosysResult)
validationRow.Add(widget.NewSeparator())
validationRow.Add(defResult)
validationRow.Add(widget.NewSeparator())
validationRow.Add(crossResult)
```

**Card Polish:**

```go
// Create standard card style
func CreateStyledCard(title, subtitle string, content fyne.CanvasObject) fyne.CanvasObject {
    card := widget.NewCard(title, subtitle, content)

    // Add subtle border/shadow effect (if theme supports)
    return container.NewCenter(card)
}

// Apply to all data cards
cellCard := CreateStyledCard(
    "Cell Configuration",
    "Define bitcell dimensions and timing",
    makeCellConfigFields(),
)

arrayCard := CreateStyledCard(
    "Array Configuration",
    "Specify array size and operation mode",
    makeArrayConfigFields(),
)

statsCard := CreateStyledCard(
    "Array Statistics",
    "Real-time metrics",
    makeStatsDisplay(),
)
```

**Button Polish:**

```go
// Larger, more touchable buttons (44px minimum)
generateAllBtn := widget.NewButton("Generate All Files", func() { /* ... */ })
generateAllBtn.SetMinSize(fyne.NewSize(120, 44))

validateAllBtn := widget.NewButton("Validate All", func() { /* ... */ })
validateAllBtn.SetMinSize(fyne.NewSize(120, 44))

exportPackageBtn := widget.NewButton("Export Package", func() { /* ... */ })
exportPackageBtn.SetMinSize(fyne.NewSize(120, 44))
```

**Separators and Dividers:**

```go
// Section dividers with muted color
sectionDivider := widget.NewSeparator()
// (Color customization in theme)

// Use consistently between major sections
mainLayout := container.NewVBox(
    builderHeader,
    sectionDivider,
    configSection,
    sectionDivider,
    previewSection,
    sectionDivider,
    validationSection,
)
```

**Verification:**
- All buttons are ≥44px tall (touch target)
- Card padding is consistent (16px)
- Section margins are 24px
- Button spacing is 12px
- Field spacing is 8px
- Separators visible but muted

**Files to Modify:**
1. `pkg/gui/widgets/spacing.go` - Create spacing constants (NEW FILE)
2. `pkg/gui/tabs/builder_validation_tab.go` - Apply spacing constants throughout

**Estimated Effort:** 2 hours

---

## Implementation Roadmap

### Phase 1: Critical Fixes (Week 1, ~12 hours)
1. **Docker X11 Integration** (3-4h)
   - Update KLayout Ruby script with Xvfb setup
   - Update OpenROAD TCL script with display configuration
   - Update Docker runner environment variables
   - Add user documentation for X11 setup

2. **Image Viewer Zoom** (3-4h)
   - Create ImageViewer widget
   - Add zoom in/out/reset buttons
   - Integrate into Layout tab
   - Test zoom responsiveness

3. **Validation Badges** (2-3h)
   - Create ValidationBadge widget with shape + text
   - Replace inline labels in validation section
   - Verify WCAG 1.4.1 compliance
   - Test with color-blindness simulator

### Phase 2: UX Improvements (Week 2, ~14 hours)
4. **Builder Tab Restructure** (4-5h)
   - Refactor Cell Config to vertical form layout
   - Refactor Array Config to vertical form layout
   - Create Statistics card grid
   - Test responsive layout

5. **Code Viewer with Syntax Highlighting** (4-5h)
   - Create CodeViewer widget with line numbers
   - Implement Verilog syntax highlighting
   - Implement DEF syntax highlighting
   - Add copy-to-clipboard button
   - Test highlighting accuracy

6. **Log Panel Enhancements** (3-4h)
   - Add filter radio buttons
   - Add search entry
   - Increase visible log lines to 10+
   - Implement log level tagging in all output
   - Test filtering and search

### Phase 3: Visual Polish (Week 3, ~8 hours)
7. **Color Theme Extensions** (2-3h)
   - Create module6_colors.go with tab-specific colors
   - Apply accent colors to Builder/Learn/Validation sections
   - Update badge colors for semantic meaning
   - Test contrast ratios

8. **Typography Standardization** (2h)
   - Create typography.go with size constants
   - Apply H1/H2/H3/Body styles to all labels
   - Update code font to 13px
   - Test readability

9. **Spacing and Polish** (2-3h)
   - Create spacing.go with margin/padding constants
   - Update all containers to use constants
   - Increase button sizes to 44px minimum
   - Add consistent separators

### Testing & Verification (1-2 hours per phase)
- Accessibility checks (WCAG 2.1 AA)
- Cross-platform testing (Linux/macOS/Windows)
- Screenshot comparisons (before/after)
- Performance impact (no slowdowns)

---

## Estimated Total Effort

| Phase | Task | Hours | Dependencies |
|-------|------|-------|--------------|
| 1 | Docker X11 Integration | 3-4 | Module 6 Docker setup |
| 1 | Image Viewer & Zoom | 3-4 | canvas.Image widget knowledge |
| 1 | Validation Badges | 2-3 | Custom widget patterns |
| 2 | Builder Tab Restructure | 4-5 | Form layout patterns |
| 2 | Code Viewer & Highlighting | 4-5 | Syntax highlighting libraries |
| 2 | Log Panel Enhancements | 3-4 | String filtering/search |
| 3 | Color Theme | 2-3 | Current theme system |
| 3 | Typography | 2 | Font size constants |
| 3 | Spacing & Polish | 2-3 | Layout proportions |
| **Total** | | **26-36 hours** | ~3-4 weeks |

---

## Files to Create

### New Files
1. `module6-eda/pkg/gui/widgets/image_viewer.go` - Zoomable image viewer
2. `module6-eda/pkg/gui/widgets/validation_badge.go` - WCAG-compliant status badges
3. `module6-eda/pkg/gui/widgets/code_viewer.go` - Syntax-highlighted code display
4. `shared/theme/module6_colors.go` - Tab-specific color palette
5. `shared/typography.go` - Typography scale and constants
6. `module6-eda/pkg/gui/widgets/spacing.go` - Spacing and layout helpers
7. `docs/eda/guides/DOCKER_SETUP.md` - Docker X11 configuration guide

### Modified Files
1. `module6-eda/pkg/validation/layout_image.go` - Update klayoutScript (X11 setup)
2. `module6-eda/pkg/validation/circuit_image.go` - Update openroadImageScript (X11 setup)
3. `module6-eda/pkg/openlane/runner.go` - Add DISPLAY env vars to Docker commands
4. `module6-eda/pkg/gui/tabs/builder_validation_tab.go` - Integrate all improvements
5. `module6-eda/pkg/gui/tabs/learn_tab.go` - Apply color theme
6. `module6-eda/README.md` - Add X11 setup instructions

---

## Success Criteria

### Critical Fixes
- ✓ KLayout generates PNG images (no exit 1 errors)
- ✓ OpenROAD generates PNG images (no exit 1 errors)
- ✓ Layout tab displays 3 images (KLayout, OpenROAD, Yosys)
- ✓ Validation badges show pass/fail with shapes + text
- ✓ All badges pass WCAG 1.4.1 color-only test

### UX Improvements
- ✓ Builder tab cards readable without scrolling
- ✓ Code preview shows syntax highlighting
- ✓ Copy button exports complete code to clipboard
- ✓ Log panel shows 10+ lines by default
- ✓ Log filtering shows/hides based on level
- ✓ Log search highlights matching messages

### Visual Polish
- ✓ Tab sections have distinct color accents
- ✓ Typography hierarchy clear (H1 > H2 > Body)
- ✓ All text meets 4.5:1 contrast ratio (WCAG AA)
- ✓ Card padding consistent (16px)
- ✓ Button sizes meet 44px minimum (touch targets)
- ✓ Spacing proportional (8px:16px:24px:32px)

### Non-Regression
- ✓ All validation checks still work (Yosys/DEF/Cross/Placement)
- ✓ All export functions produce correct output
- ✓ Learn tab content unchanged (visual refresh only)
- ✓ Performance: No slowdowns vs baseline

---

## References & Resources

### WCAG 2.1 Level AA Compliance
- 1.4.1 Use of Color: https://www.w3.org/WAI/WCAG21/Understanding/use-of-color.html
- 1.4.3 Contrast (Minimum): https://www.w3.org/WAI/WCAG21/Understanding/contrast-minimum.html
- 2.5.5 Target Size: https://www.w3.org/WAI/WCAG21/Understanding/target-size.html

### Accessibility Testing Tools
- Contrast Checker: https://webaim.org/resources/contrastchecker/
- Color Blindness Simulator: https://www.color-blindness.com/coblis-color-blindness-simulator/
- WAVE (Web Accessibility Tool): https://wave.webaim.org/

### Design References
- Material Design 3 (spacing): https://m3.material.io/foundations/layout/understanding-layout
- Apple Human Interface Guidelines: https://developer.apple.com/design/human-interface-guidelines/
- Fyne Theme Documentation: https://docs.fyne.io/v2.1/custom/theme.html

### Syntax Highlighting Libraries
- Chroma (Go syntax highlighting): https://github.com/alecthomas/chroma
- Highlight.js (JavaScript alternative): https://highlightjs.org/

---

## Appendix: ASCII Layout Mockups

### Builder Tab - Before vs After

**Before (Dense):**
```
┌─ Cell Config (Grid) ─┬─ Array Config (Grid) ─┐
│ Name|Width|Height|Rise│ Rows|Cols|Mode|Arch    │
│ Fall|Cap|Leakage|Area │ ModeHelp... | Stats   │
└─────────────────────┬─────────────────────┘
│ Stats: Total | Area | WL | BL | Density | Util
└──────────────────────────────────────────────┘
```

**After (Hierarchical):**
```
┌─ Cell Configuration ──────────┬─ Array Configuration ──────┐
│ Cell Name:        [fecim_bct] │ Array Rows:        [ 4  ]   │
│ Width (μm):         [  0.460 ] │ Array Columns:     [ 4  ]   │
│ Height (μm):        [  2.720 ] │ Operation Mode:    [storage] │
│                                │   Storage mode explains...  │
│ ─── Timing ─────────────────  │ Architecture:      [passive] │
│ Rise Time (ns):     [   0.1  ] │                              │
│ Fall Time (ns):     [   0.1  ] │                              │
│                                │                              │
│ ─── Electrical ──────────────  │                              │
│ Input Cap (pF):     [  0.002 ] │                              │
│ Leakage (nW):       [  0.001 ] │                              │
│ Cell Area: 1.25 µm²            │                              │
└────────────────────┴──────────────────────────┘

┌─ Array Statistics ────────────────────────────────────┐
│  Total Cells  │  Array Area   │  Cell Density        │
│    16 cells   │  20.0 µm²     │  0.80 cells/µm²      │
├──────────────┼───────────────┼──────────────────────┤
│  Word Line    │  Bit Line     │  Utilization         │
│  3.68 µm      │  21.76 µm     │  100.0%              │
└───────────────────────────────────────────────────────┘
```

### Layout Tab - Image Sizes

**Before (400x350):**
```
┌─ KLayout ─┬─ OpenROAD ─┬─ Yosys ──┐
│ 400x350   │  400x350   │ 400x350  │
│ [small]   │  [small]   │ [small]  │
│ Hard to   │  Hard to   │ Hard to  │
│ see       │  read      │ parse    │
└───────────┴────────────┴──────────┘
```

**After (600x500 + Zoom):**
```
┌─ KLayout ────────────┬─ OpenROAD ────────────┬─ Yosys ──────────────┐
│ [Zoom] [Generated]   │ [Zoom] [Generated]    │ [Zoom] [Generated]   │
│                      │                       │                      │
│ 600x500 image        │ 600x500 image         │ 600x500 image        │
│ Clear layout         │ Clear placement       │ Clear schematic      │
│ details visible      │ details visible       │ details visible      │
└──────────────────────┴───────────────────────┴──────────────────────┘

Click [Zoom] → Full-screen modal with pan/zoom controls
```

### Code Preview - Before vs After

**Before (Plain):**
```
[Preview] [DEF]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
DESIGN fecim_crossbar_4x4
UNITS DISTANCE MICRONS 1000
DIEAREA (0 0) (1.84 10.88)
COMPONENTS 16 ;
- fecim_0 fecim_bitcell + PLACED (0.46 0) N ;
- fecim_1 fecim_bitcell + PLACED (0.46 2.72) N ;
...more lines...
```

**After (Highlighted + Line Numbers + Copy):**
```
[DEF Preview] [Verilog] | [Copy to Clipboard]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  1 | DESIGN fecim_crossbar_4x4
  2 | UNITS DISTANCE MICRONS 1000
  3 | DIEAREA (0 0) (1.84 10.88)
  4 | COMPONENTS 16 ;
  5 | - fecim_0 [KEYWORD]fecim_bitcell[/] + PLACED (0.46 0) N ;
  6 | - fecim_1 [KEYWORD]fecim_bitcell[/] + PLACED (0.46 2.72) N ;
  7 | ...
  (Color: keywords=cyan, numbers=yellow, comments=gray)
```

### Validation Badges - Before vs After

**Before (Color Only):**
```
✓ PASS  ✗ FAIL  ✓ PASS  ⊝ SKIP
 (GRN)  (RED)  (GRN)  (GRAY)
   ↑ Color-blind users can't distinguish
```

**After (Shape + Text + Color):**
```
● ✓ PASS    ▲ ✗ FAIL    ● ✓ PASS    — ⊝ SKIP
 ↑ Shape     ↑ Shape     ↑ Shape     ↑ Shape
(Green)    (Red)      (Green)    (Gray)
 Text clearly labels each state
```

---

## Questions for Architect Review

Before implementation, clarify:

1. **X11 Setup**: Should we detect X11 availability and gracefully degrade, or require it?
2. **Image Viewer**: Zoom via modal only, or inline zoom controls too?
3. **Syntax Highlighting**: Full regex highlighting or simplified keyword matching?
4. **Log Levels**: Should validation steps automatically emit INFO/WARN/ERROR?
5. **Theme**: Should colors be user-configurable, or fixed per module?
6. **Breaking Changes**: Acceptable to refactor tab layouts (no backward compat needed)?

---

**Document End**

Version History:
- v1.0 (2026-01-30): Initial proposal with 3 priority tiers, 26-36 hour estimate
