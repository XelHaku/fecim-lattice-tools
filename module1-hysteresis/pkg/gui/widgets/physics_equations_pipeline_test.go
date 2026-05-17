//go:build legacy_fyne

package widgets

import (
	"encoding/json"
	"encoding/xml"
	"strconv"
	"strings"
	"testing"

	eqassets "fecim-lattice-tools/shared/assets/equations"
)

type svgRoot struct {
	XMLName xml.Name `xml:"svg"`
	ViewBox string   `xml:"viewBox,attr"`
	Width   string   `xml:"width,attr"`
	Height  string   `xml:"height,attr"`
}

func TestEquationPipeline_EmbeddedHotspotsAreSingleSourceOfTruth(t *testing.T) {
	var cfg hotspotConfig
	if err := json.Unmarshal(eqassets.LkHotspotsJSON, &cfg); err != nil {
		t.Fatalf("embedded hotspots json parse failed: %v", err)
	}
	if len(cfg.Hotspots) == 0 {
		t.Fatal("embedded hotspots must not be empty")
	}

	loadedHotspots, loadedSize := loadLkHotspots()
	if len(loadedHotspots) != len(cfg.Hotspots) {
		t.Fatalf("hotspot source-of-truth mismatch: loaded=%d embedded=%d", len(loadedHotspots), len(cfg.Hotspots))
	}
	if loadedSize.Width != cfg.BaseWidth || loadedSize.Height != cfg.BaseHeight {
		t.Fatalf("base size mismatch loaded=(%g,%g) embedded=(%g,%g)", loadedSize.Width, loadedSize.Height, cfg.BaseWidth, cfg.BaseHeight)
	}

	details := termDetails()
	for _, spot := range loadedHotspots {
		if _, ok := details[spot.ID]; !ok {
			t.Fatalf("hotspot id %q has no term detail mapping", spot.ID)
		}
		if spot.X < 0 || spot.Y < 0 || spot.W <= 0 || spot.H <= 0 {
			t.Fatalf("invalid hotspot geometry %+v", spot)
		}
		if spot.X+spot.W > 1.0001 || spot.Y+spot.H > 1.0001 {
			t.Fatalf("hotspot out of normalized bounds %+v", spot)
		}
	}
}

func TestEquationPipeline_SVGViewBoxAlignsWithHotspotBaseSize(t *testing.T) {
	decoder := xml.NewDecoder(strings.NewReader(string(eqassets.LkEquationSVG)))
	var root svgRoot
	for {
		tok, err := decoder.Token()
		if err != nil {
			t.Fatalf("failed to decode svg: %v", err)
		}
		start, ok := tok.(xml.StartElement)
		if !ok || start.Name.Local != "svg" {
			continue
		}
		if err := decoder.DecodeElement(&root, &start); err != nil {
			t.Fatalf("decode root svg element: %v", err)
		}
		break
	}
	if root.ViewBox == "" {
		t.Fatal("svg missing viewBox")
	}

	parts := strings.Fields(root.ViewBox)
	if len(parts) != 4 {
		t.Fatalf("unexpected viewBox format: %q", root.ViewBox)
	}
	vbW, err := strconv.ParseFloat(parts[2], 32)
	if err != nil {
		t.Fatalf("parse viewbox width: %v", err)
	}
	vbH, err := strconv.ParseFloat(parts[3], 32)
	if err != nil {
		t.Fatalf("parse viewbox height: %v", err)
	}

	_, size := loadLkHotspots()
	if size.Width <= 0 || size.Height <= 0 {
		t.Fatalf("invalid hotspot base size (%g,%g)", size.Width, size.Height)
	}
	svgAspect := float32(vbW / vbH)
	hotspotAspect := size.Width / size.Height
	if absf(svgAspect-hotspotAspect) > 0.02 {
		t.Fatalf("svg/hotspot aspect mismatch: svg=%.5f hotspots=%.5f", svgAspect, hotspotAspect)
	}
}

func TestEquationPipeline_HotspotLayoutOrderMatchesEquationTerms(t *testing.T) {
	hotspots, _ := loadLkHotspots()
	byID := make(map[string]hotspotDef, len(hotspots))
	for _, spot := range hotspots {
		byID[spot.ID] = spot
	}

	orderedTopRow := []string{"rho_eff_main", "e_applied", "k_dep", "alpha", "beta", "gamma", "noise"}
	lastX := float32(-1)
	for _, id := range orderedTopRow {
		spot, ok := byID[id]
		if !ok {
			t.Fatalf("missing expected top-row hotspot id=%q", id)
		}
		if spot.Y > 0.12 {
			t.Fatalf("top-row hotspot %q unexpectedly low (y=%g)", id, spot.Y)
		}
		if spot.X <= lastX {
			t.Fatalf("top-row hotspot %q is out of left-to-right order (x=%g <= %g)", id, spot.X, lastX)
		}
		lastX = spot.X
	}

	if terms := byID["lk_terms"]; terms.Y < 0.20 || terms.Y > 0.36 {
		t.Fatalf("lk_terms hotspot expected around nonlinearity row, got y=%g", terms.Y)
	}
}

func absf(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
