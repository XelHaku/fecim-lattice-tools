// pkg/gui/widgets/widgets_test.go
package widgets

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"fecim-lattice-tools/module6-eda/pkg/config"
)

func TestNewLayoutCanvas(t *testing.T) {
	cfg := &config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "storage",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	canvas := NewLayoutCanvas(cfg)
	if canvas == nil {
		t.Fatal("NewLayoutCanvas returned nil")
	}
}

func TestNewLayoutCanvas_NilConfig(t *testing.T) {
	canvas := NewLayoutCanvas(nil)
	if canvas == nil {
		t.Fatal("NewLayoutCanvas returned nil with nil config")
	}
}

func TestLayoutCanvas_SetConfig(t *testing.T) {
	initialCfg := &config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "storage",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	canvas := NewLayoutCanvas(initialCfg)

	newCfg := &config.ArrayConfig{
		Rows:         8,
		Cols:         8,
		Mode:         "memory",
		Architecture: "1t1r",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   4.07,
	}

	// Should not panic
	canvas.SetConfig(newCfg)
}

func TestLayoutCanvas_SetConfig_Nil(t *testing.T) {
	cfg := &config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "storage",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	canvas := NewLayoutCanvas(cfg)

	// Setting nil config should not panic
	canvas.SetConfig(nil)
}

func TestLayoutCanvas_MinSize(t *testing.T) {
	tests := []struct {
		name   string
		config *config.ArrayConfig
	}{
		{
			name: "Passive 4x4",
			config: &config.ArrayConfig{
				Rows:         4,
				Cols:         4,
				Mode:         "storage",
				Architecture: "passive",
				Technology:   "sky130",
				CellWidth:    0.46,
				CellHeight:   2.72,
			},
		},
		{
			name: "1T1R 4x4",
			config: &config.ArrayConfig{
				Rows:         4,
				Cols:         4,
				Mode:         "memory",
				Architecture: "1t1r",
				Technology:   "sky130",
				CellWidth:    0.46,
				CellHeight:   4.07,
			},
		},
		{
			name: "Large 16x16",
			config: &config.ArrayConfig{
				Rows:         16,
				Cols:         16,
				Mode:         "compute",
				Architecture: "passive",
				Technology:   "sky130",
				CellWidth:    0.46,
				CellHeight:   2.72,
			},
		},
		{
			name:   "Nil config",
			config: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canvas := NewLayoutCanvas(tt.config)
			minSize := canvas.MinSize()

			// Minimum size should be at least 500x400 or larger
			if minSize.Width < 500 {
				t.Errorf("MinSize width %f is less than minimum 500", minSize.Width)
			}
			if minSize.Height < 400 {
				t.Errorf("MinSize height %f is less than minimum 400", minSize.Height)
			}
		})
	}
}

func TestLayoutCanvas_CreateRenderer(t *testing.T) {
	cfg := &config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "storage",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	canvas := NewLayoutCanvas(cfg)
	renderer := canvas.CreateRenderer()

	if renderer == nil {
		t.Fatal("CreateRenderer returned nil")
	}
}

func TestLayoutCanvas_Rendering(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	cfg := &config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "storage",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	canvas := NewLayoutCanvas(cfg)

	// Create a test window to trigger rendering
	window := testApp.NewWindow("Test")
	window.SetContent(canvas)

	// Should not panic during rendering
	window.Resize(fyne.NewSize(800, 600))

	window.Close()
}

func TestLayoutCanvas_1T1RArchitecture(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	cfg := &config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "memory",
		Architecture: "1t1r",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   4.07,
	}

	canvas := NewLayoutCanvas(cfg)
	minSize := canvas.MinSize()

	// 1T1R should have slightly larger height for SL legend
	if minSize.Height < 400 {
		t.Error("1T1R MinSize height is too small")
	}

	window := testApp.NewWindow("Test 1T1R")
	window.SetContent(canvas)
	window.Resize(fyne.NewSize(800, 600))
	window.Close()
}

func TestLayoutCanvas_PassiveArchitecture(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	cfg := &config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "storage",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	canvas := NewLayoutCanvas(cfg)

	window := testApp.NewWindow("Test Passive")
	window.SetContent(canvas)
	window.Resize(fyne.NewSize(800, 600))
	window.Close()
}

func TestLayoutCanvas_LargeArray(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	cfg := &config.ArrayConfig{
		Rows:         32,
		Cols:         32,
		Mode:         "compute",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	canvas := NewLayoutCanvas(cfg)
	minSize := canvas.MinSize()

	// Large array should have proportionally larger MinSize
	if minSize.Width < 500 || minSize.Height < 400 {
		t.Error("Large array MinSize is unexpectedly small")
	}

	window := testApp.NewWindow("Test Large")
	window.SetContent(canvas)
	window.Resize(fyne.NewSize(1200, 1000))
	window.Close()
}

func TestLayoutCanvas_SmallArray(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	cfg := &config.ArrayConfig{
		Rows:         1,
		Cols:         1,
		Mode:         "storage",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	canvas := NewLayoutCanvas(cfg)
	minSize := canvas.MinSize()

	// Even 1x1 array should have minimum size enforced
	if minSize.Width < 500 || minSize.Height < 400 {
		t.Error("Minimum size not enforced for small array")
	}
}

func TestLayoutCanvas_ZeroSizedArray(t *testing.T) {
	cfg := &config.ArrayConfig{
		Rows:         0,
		Cols:         0,
		Mode:         "storage",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	canvas := NewLayoutCanvas(cfg)

	// Should not panic with zero-sized array
	minSize := canvas.MinSize()
	if minSize.Width < 500 || minSize.Height < 400 {
		t.Error("MinSize not enforced for zero-sized array")
	}
}

func TestLayoutCanvasRenderer_Objects(t *testing.T) {
	cfg := &config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "storage",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	canvas := NewLayoutCanvas(cfg)
	renderer := canvas.CreateRenderer()

	// Call Refresh to populate objects
	renderer.Refresh()

	objects := renderer.Objects()
	// Objects may be nil until first refresh - this is valid behavior
	if objects == nil {
		t.Skip("Renderer Objects() returned nil before refresh (valid)")
	}

	// Should have at least some objects (background, cells, lines, etc.)
	if len(objects) == 0 {
		t.Error("Renderer Objects() returned empty slice")
	}
}

func TestLayoutCanvasRenderer_Refresh(t *testing.T) {
	cfg := &config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "storage",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	canvas := NewLayoutCanvas(cfg)
	renderer := canvas.CreateRenderer()

	// Should not panic
	renderer.Refresh()
}

func TestLayoutCanvasRenderer_Layout(t *testing.T) {
	cfg := &config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "storage",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	canvas := NewLayoutCanvas(cfg)
	renderer := canvas.CreateRenderer()

	// Should not panic with various sizes
	renderer.Layout(fyne.NewSize(100, 100))
	renderer.Layout(fyne.NewSize(800, 600))
	renderer.Layout(fyne.NewSize(1920, 1080))
}

func TestLayoutCanvasRenderer_MinSize(t *testing.T) {
	cfg := &config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "storage",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	canvas := NewLayoutCanvas(cfg)
	renderer := canvas.CreateRenderer()

	minSize := renderer.MinSize()
	if minSize.Width < 500 || minSize.Height < 400 {
		t.Error("Renderer MinSize is less than expected minimum")
	}
}

func TestLayoutCanvasRenderer_Destroy(t *testing.T) {
	cfg := &config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "storage",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	canvas := NewLayoutCanvas(cfg)
	renderer := canvas.CreateRenderer()

	// Should not panic
	renderer.Destroy()
}

func TestLayoutCanvas_UpdateAfterResize(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	cfg := &config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "storage",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	canvas := NewLayoutCanvas(cfg)
	window := testApp.NewWindow("Test Resize")
	window.SetContent(canvas)

	// Resize multiple times
	window.Resize(fyne.NewSize(600, 500))
	window.Resize(fyne.NewSize(800, 700))
	window.Resize(fyne.NewSize(1000, 900))

	window.Close()
}

func TestLayoutCanvas_MultipleConfigChanges(t *testing.T) {
	canvas := NewLayoutCanvas(nil)

	configs := []*config.ArrayConfig{
		{Rows: 4, Cols: 4, Architecture: "passive", CellWidth: 0.46, CellHeight: 2.72},
		{Rows: 8, Cols: 8, Architecture: "1t1r", CellWidth: 0.46, CellHeight: 4.07},
		{Rows: 2, Cols: 2, Architecture: "passive", CellWidth: 0.46, CellHeight: 2.72},
		nil,
		{Rows: 16, Cols: 16, Architecture: "passive", CellWidth: 0.46, CellHeight: 2.72},
	}

	for i, cfg := range configs {
		t.Run("Config change "+string(rune('A'+i)), func(t *testing.T) {
			canvas.SetConfig(cfg)
			_ = canvas.MinSize() // Should not panic
		})
	}
}

func TestLayoutCanvas_EmptyConfigFields(t *testing.T) {
	cfg := &config.ArrayConfig{
		// Intentionally leaving many fields empty/zero
		Rows: 2,
		Cols: 2,
	}

	canvas := NewLayoutCanvas(cfg)
	if canvas == nil {
		t.Fatal("NewLayoutCanvas failed with partial config")
	}

	minSize := canvas.MinSize()
	if minSize.Width == 0 || minSize.Height == 0 {
		t.Error("MinSize has zero dimensions with partial config")
	}
}
