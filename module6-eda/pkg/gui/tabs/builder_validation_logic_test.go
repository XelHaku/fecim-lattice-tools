package tabs

import (
	"os"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module6-eda/pkg/config"
)

func TestMakeBuilderValidationTab_StateAndValidationLogicPaths(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	w := app.NewWindow("builder")
	defer w.Close()

	cfg := &config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "storage",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	root := MakeBuilderValidationTab(cfg, w)
	if root == nil {
		t.Fatal("MakeBuilderValidationTab returned nil")
	}

	rowsEntry := findEntryWithText(root, "4")
	colsEntry := findNthEntryWithText(root, "4", 2)
	widthEntry := findEntryWithText(root, "0.460")
	heightEntry := findEntryWithText(root, "2.720")
	if rowsEntry == nil || colsEntry == nil || widthEntry == nil || heightEntry == nil {
		t.Fatal("failed to discover expected entries in builder tab")
	}

	// Exercise updateStats parsing/validation branches via entry OnChanged handlers.
	rowsEntry.SetText("2048") // clamps to 1024
	if cfg.Rows != 1024 {
		t.Fatalf("expected rows clamped to 1024, got %d", cfg.Rows)
	}

	colsEntry.SetText("-3") // invalid -> keep previous cfg value
	if cfg.Cols <= 0 {
		t.Fatalf("expected cols to remain valid positive value, got %d", cfg.Cols)
	}

	widthEntry.SetText("bad") // parse error -> keep prior cfg value
	heightEntry.SetText("0")  // <=0 -> keep prior cfg value
	if cfg.CellWidth <= 0 || cfg.CellHeight <= 0 {
		t.Fatalf("expected cell dimensions to remain positive, got w=%f h=%f", cfg.CellWidth, cfg.CellHeight)
	}

	// Exercise architecture state transitions and early-return branches.
	passive := findButtonByText(root, "PASSIVE")
	oneT1R := findButtonByText(root, "1T1R")
	twoT1R := findButtonByText(root, "2T1R")
	if passive == nil || oneT1R == nil || twoT1R == nil {
		t.Fatal("failed to discover architecture buttons")
	}

	oneT1R.OnTapped()
	if cfg.Architecture != "1t1r" {
		t.Fatalf("expected architecture 1t1r, got %s", cfg.Architecture)
	}
	oneT1R.OnTapped() // already selected branch (early return)

	twoT1R.OnTapped()
	if cfg.Architecture != "2t1r" {
		t.Fatalf("expected architecture 2t1r, got %s", cfg.Architecture)
	}
	passive.OnTapped()
	if cfg.Architecture != "passive" {
		t.Fatalf("expected architecture passive, got %s", cfg.Architecture)
	}
}

func TestMakeBuilderValidationTab_RunPrimaryActions(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	cfg := &config.ArrayConfig{
		Rows:         2,
		Cols:         2,
		Mode:         "storage",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	defer func() { _ = os.Chdir(origWD) }()

	root := MakeBuilderValidationTab(cfg, nil)
	if root == nil {
		t.Fatal("MakeBuilderValidationTab returned nil")
	}

	gen := findButtonByText(root, "Generate All")
	val := findButtonByText(root, "Validate All")
	exp := findButtonByText(root, "Export Package")
	if gen == nil || val == nil || exp == nil {
		t.Fatal("failed to find primary action buttons")
	}

	// Trigger primary actions to ensure handlers are wired and do not panic in headless mode.
	gen.OnTapped()
	time.Sleep(100 * time.Millisecond)
	val.OnTapped()
	time.Sleep(100 * time.Millisecond)
	exp.OnTapped()
	time.Sleep(100 * time.Millisecond)
}

// TestArchitectureSwitch_UpdatesNameEntry verifies that the name entry auto-updates
// to the architecture-specific default when switching between passive/1T1R/2T1R,
// but leaves user-typed custom names untouched.
func TestArchitectureSwitch_UpdatesNameEntry(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	w := app.NewWindow("arch-name")
	defer w.Close()

	cfg := &config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "storage",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	root := MakeBuilderValidationTab(cfg, w)
	passive := findButtonByText(root, "PASSIVE")
	oneT1R := findButtonByText(root, "1T1R")
	twoT1R := findButtonByText(root, "2T1R")
	nameEntry := findEntryWithText(root, "fecim_bitcell")
	if passive == nil || oneT1R == nil || twoT1R == nil || nameEntry == nil {
		t.Fatal("failed to find required widgets")
	}

	// passive → 1T1R: name should update to fecim_1t1r_bitcell
	oneT1R.OnTapped()
	if nameEntry.Text != "fecim_1t1r_bitcell" {
		t.Errorf("passive→1T1R: nameEntry = %q, want %q", nameEntry.Text, "fecim_1t1r_bitcell")
	}

	// 1T1R → 2T1R: name should update to fecim_2t1r_bitcell
	twoT1R.OnTapped()
	if nameEntry.Text != "fecim_2t1r_bitcell" {
		t.Errorf("1T1R→2T1R: nameEntry = %q, want %q", nameEntry.Text, "fecim_2t1r_bitcell")
	}

	// 2T1R → passive: name should update back to fecim_bitcell
	passive.OnTapped()
	if nameEntry.Text != "fecim_bitcell" {
		t.Errorf("2T1R→passive: nameEntry = %q, want %q", nameEntry.Text, "fecim_bitcell")
	}

	// Custom name: switching arch must NOT overwrite it
	nameEntry.SetText("my_custom_cell")
	oneT1R.OnTapped()
	if nameEntry.Text != "my_custom_cell" {
		t.Errorf("custom name overwritten on arch switch: got %q, want %q", nameEntry.Text, "my_custom_cell")
	}
	twoT1R.OnTapped()
	if nameEntry.Text != "my_custom_cell" {
		t.Errorf("custom name overwritten on 2T1R switch: got %q, want %q", nameEntry.Text, "my_custom_cell")
	}
	passive.OnTapped()
	if nameEntry.Text != "my_custom_cell" {
		t.Errorf("custom name overwritten on passive switch: got %q, want %q", nameEntry.Text, "my_custom_cell")
	}
}

// TestArchitectureSwitch_UpdatesCellDimensions verifies that the width entry is set
// to the correct physical pitch when switching between passive/1T1R/2T1R.
// These defaults match the LEF generator's minimum thresholds, preventing silent
// dimension overrides in the generated output.
func TestArchitectureSwitch_UpdatesCellDimensions(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	w := app.NewWindow("arch-dims")
	defer w.Close()

	cfg := &config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "storage",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	root := MakeBuilderValidationTab(cfg, w)
	passive := findButtonByText(root, "PASSIVE")
	oneT1R := findButtonByText(root, "1T1R")
	twoT1R := findButtonByText(root, "2T1R")
	// Find widthEntry while it still shows the passive default.
	widthEntry := findEntryWithText(root, "0.460")
	if passive == nil || oneT1R == nil || twoT1R == nil || widthEntry == nil {
		t.Fatal("failed to find required widgets")
	}

	// passive → 1T1R: width must be 0.920 (2 SKY130 sites, matches LEF min)
	oneT1R.OnTapped()
	if widthEntry.Text != "0.920" {
		t.Errorf("passive→1T1R: widthEntry = %q, want %q", widthEntry.Text, "0.920")
	}

	// 1T1R → 2T1R: width must be 1.380 (3 SKY130 sites, matches LEF min)
	twoT1R.OnTapped()
	if widthEntry.Text != "1.380" {
		t.Errorf("1T1R→2T1R: widthEntry = %q, want %q", widthEntry.Text, "1.380")
	}

	// 2T1R → passive: width must return to 0.460
	passive.OnTapped()
	if widthEntry.Text != "0.460" {
		t.Errorf("2T1R→passive: widthEntry = %q, want %q", widthEntry.Text, "0.460")
	}
}

func waitUntil(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}

func findEntryWithText(root fyne.CanvasObject, text string) *widget.Entry {
	var found *widget.Entry
	walkObjects(root, func(obj fyne.CanvasObject) {
		if found != nil {
			return
		}
		if e, ok := obj.(*widget.Entry); ok && e.Text == text {
			found = e
		}
	})
	return found
}

func findNthEntryWithText(root fyne.CanvasObject, text string, n int) *widget.Entry {
	count := 0
	var found *widget.Entry
	walkObjects(root, func(obj fyne.CanvasObject) {
		if found != nil {
			return
		}
		if e, ok := obj.(*widget.Entry); ok && e.Text == text {
			count++
			if count == n {
				found = e
			}
		}
	})
	return found
}

func findButtonByText(root fyne.CanvasObject, label string) *widget.Button {
	var found *widget.Button
	walkObjects(root, func(obj fyne.CanvasObject) {
		if found != nil {
			return
		}
		if b, ok := obj.(*widget.Button); ok && b.Text == label {
			found = b
		}
	})
	return found
}

func walkObjects(obj fyne.CanvasObject, visit func(fyne.CanvasObject)) {
	if obj == nil {
		return
	}
	visit(obj)

	switch v := obj.(type) {
	case *fyne.Container:
		for _, child := range v.Objects {
			walkObjects(child, visit)
		}
	case *container.AppTabs:
		for _, item := range v.Items {
			walkObjects(item.Content, visit)
		}
	case *widget.Accordion:
		for _, item := range v.Items {
			walkObjects(item.Detail, visit)
		}
	}
}
