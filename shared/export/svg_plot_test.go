package export

import (
	"encoding/xml"
	"math"
	"os"
	"strings"
	"testing"
)

// TestGeneratePELoopSVG_ValidXML verifies the output is well-formed XML.
func TestGeneratePELoopSVG_ValidXML(t *testing.T) {
	e, p := sampleLoop(50)
	svg, err := GeneratePELoopSVG(e, p, DefaultSVGPlotConfig())
	if err != nil {
		t.Fatalf("GeneratePELoopSVG returned error: %v", err)
	}
	if err := xml.Unmarshal([]byte(svg), new(interface{})); err != nil {
		t.Fatalf("SVG is not valid XML: %v", err)
	}
}

// TestGeneratePELoopSVG_ContainsExpectedElements checks for axes, polyline, title, labels.
func TestGeneratePELoopSVG_ContainsExpectedElements(t *testing.T) {
	cfg := DefaultSVGPlotConfig()
	cfg.Title = "HZO P-E Loop"
	cfg.Citation = "HZO — 2026-03-29 — abc1234"
	e, p := sampleLoop(100)

	svg, err := GeneratePELoopSVG(e, p, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []string{
		"<svg",
		"</svg>",
		"<polyline",
		"HZO P-E Loop",             // title
		"E [MV/cm]",                // x-axis label
		"\u00b5C/cm\u00b2",         // y-axis label (µC/cm²)
		"HZO",                      // citation
		"abc1234",                   // commit in citation
		"class=\"loop\"",           // polyline class
		"class=\"axis\"",           // axis lines
		"class=\"tick\"",           // tick marks
		"P-E Loop",                 // legend text
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("SVG missing expected content: %q", want)
		}
	}
}

// TestGeneratePELoopSVG_CustomConfig verifies custom dimensions are respected.
func TestGeneratePELoopSVG_CustomConfig(t *testing.T) {
	cfg := SVGPlotConfig{
		Width:   800,
		Height:  600,
		MarginL: 100,
		MarginR: 40,
		MarginT: 60,
		MarginB: 80,
		Title:   "Custom",
		XLabel:  "X",
		YLabel:  "Y",
	}
	e, p := sampleLoop(20)
	svg, err := GeneratePELoopSVG(e, p, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(svg, `width="800"`) {
		t.Error("custom width not applied")
	}
	if !strings.Contains(svg, `height="600"`) {
		t.Error("custom height not applied")
	}
}

// TestGeneratePELoopSVG_ErrorCases ensures proper validation.
func TestGeneratePELoopSVG_ErrorCases(t *testing.T) {
	cfg := DefaultSVGPlotConfig()

	// Mismatched lengths.
	_, err := GeneratePELoopSVG([]float64{1, 2}, []float64{1}, cfg)
	if err == nil {
		t.Error("expected error for mismatched lengths")
	}

	// Too few points.
	_, err = GeneratePELoopSVG([]float64{1}, []float64{2}, cfg)
	if err == nil {
		t.Error("expected error for < 2 points")
	}

	// Empty slices.
	_, err = GeneratePELoopSVG(nil, nil, cfg)
	if err == nil {
		t.Error("expected error for nil slices")
	}
}

// TestGeneratePELoopSVG_XMLEscaping verifies special characters are escaped.
func TestGeneratePELoopSVG_XMLEscaping(t *testing.T) {
	cfg := DefaultSVGPlotConfig()
	cfg.Title = "A < B & C > D"
	cfg.Citation = `"quoted" & <tagged>`

	e, p := sampleLoop(10)
	svg, err := GeneratePELoopSVG(e, p, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Must not contain unescaped special chars in text content.
	if strings.Contains(svg, "A < B") {
		t.Error("title contains unescaped '<'")
	}
	if err := xml.Unmarshal([]byte(svg), new(interface{})); err != nil {
		t.Fatalf("SVG with special chars is not valid XML: %v", err)
	}
}

// TestExporter_ExportSVG verifies the file-writing integration.
func TestExporter_ExportSVG(t *testing.T) {
	tmpDir := t.TempDir()
	exporter := NewExporter(tmpDir, "pe_loop")

	e, p := sampleLoop(30)
	svg, err := GeneratePELoopSVG(e, p, DefaultSVGPlotConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := exporter.ExportSVG(svg)
	if result.Error != nil {
		t.Fatalf("ExportSVG failed: %v", result.Error)
	}
	if result.Format != FormatSVG {
		t.Errorf("expected format SVG, got %s", result.Format)
	}
	if !strings.HasSuffix(result.FilePath, ".svg") {
		t.Errorf("expected .svg extension, got %s", result.FilePath)
	}

	content, err := os.ReadFile(result.FilePath)
	if err != nil {
		t.Fatalf("cannot read exported file: %v", err)
	}
	if !strings.Contains(string(content), "<svg") {
		t.Error("exported file does not contain SVG content")
	}
}

// TestNiceStep verifies tick spacing selection.
func TestNiceStep(t *testing.T) {
	tests := []struct {
		dataRange float64
		want      float64
	}{
		{10, 2},
		{100, 20},
		{0.5, 0.1},
		{0, 1},   // edge case: zero range
		{-5, 1},  // edge case: negative range
	}
	for _, tc := range tests {
		got := niceStep(tc.dataRange)
		if got != tc.want {
			t.Errorf("niceStep(%g) = %g, want %g", tc.dataRange, got, tc.want)
		}
	}
}

// sampleLoop generates a synthetic P-E hysteresis loop with n points.
func sampleLoop(n int) (eField, polarization []float64) {
	eField = make([]float64, 2*n)
	polarization = make([]float64, 2*n)
	for i := 0; i < n; i++ {
		th := float64(i) / float64(n) * math.Pi
		eField[i] = 2.0 * math.Cos(th)
		polarization[i] = 20.0*math.Tanh(3*math.Cos(th)) + 2.0*math.Sin(th)
	}
	for i := 0; i < n; i++ {
		th := math.Pi + float64(i)/float64(n)*math.Pi
		eField[n+i] = 2.0 * math.Cos(th)
		polarization[n+i] = 20.0*math.Tanh(3*math.Cos(th)) + 2.0*math.Sin(th)
	}
	return
}
