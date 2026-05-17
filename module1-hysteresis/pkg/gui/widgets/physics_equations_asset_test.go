//go:build legacy_fyne

package widgets

import (
	"bytes"
	"sync"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	eqassets "fecim-lattice-tools/shared/assets/equations"
)

func resetEquationCachesForTest() {
	equationSVGCacheMu.Lock()
	equationSVGCache = map[string]*equationSVGCacheEntry{}
	equationSVGCacheMu.Unlock()

	lkHotspotsOnce = sync.Once{}
	cachedLkSpots = nil
	cachedLkSize = fyne.Size{}
}

func TestEquationSVGResource_CacheReturnsStableResource(t *testing.T) {
	resetEquationCachesForTest()

	res1, ok1 := loadEquationSVGResource(lkEquationID)
	res2, ok2 := loadEquationSVGResource(lkEquationID)
	if !ok1 || !ok2 {
		t.Fatal("expected embedded LK SVG to load")
	}
	if res1 == nil || res2 == nil {
		t.Fatal("expected non-nil resources")
	}
	if res1 != res2 {
		t.Fatal("expected cached resource pointer to be reused for same theme")
	}
}

func TestEquationSVGAssets_VectorOnly_NoBitmapPayloads(t *testing.T) {
	assets := map[string][]byte{
		"lk":       eqassets.LkEquationSVG,
		"preisach": eqassets.PreisachEquationSVG,
	}
	for name, data := range assets {
		if len(data) == 0 {
			t.Fatalf("%s svg must not be empty", name)
		}
		lower := bytes.ToLower(data)
		if bytes.Contains(lower, []byte("<image")) {
			t.Fatalf("%s svg contains <image> tag (bitmap fallback)", name)
		}
		if bytes.Contains(lower, []byte("data:image")) {
			t.Fatalf("%s svg embeds raster data URI", name)
		}
		if bytes.Contains(lower, []byte("<foreignobject")) {
			t.Fatalf("%s svg contains <foreignObject>, expected pure vector primitives", name)
		}
	}
}

func TestEquationWidget_LKFallsBackToTextWhenSVGMissing(t *testing.T) {
	test.NewApp()
	win := test.NewWindow(widget.NewLabel("host"))

	resetEquationCachesForTest()
	orig := eqassets.LkEquationSVG
	eqassets.LkEquationSVG = nil
	t.Cleanup(func() { eqassets.LkEquationSVG = orig })

	panel := buildLkEquationPanel(win, func(string, string) {})
	if panel == nil {
		t.Fatal("expected non-nil panel")
	}
	if countImages(panel) != 0 {
		t.Fatal("expected LK fallback to text-only layout when SVG is missing")
	}
}

func TestEquationWidget_PreisachFallsBackToTextWhenSVGMissing(t *testing.T) {
	test.NewApp()
	win := test.NewWindow(widget.NewLabel("host"))

	resetEquationCachesForTest()
	orig := eqassets.PreisachEquationSVG
	eqassets.PreisachEquationSVG = nil
	t.Cleanup(func() { eqassets.PreisachEquationSVG = orig })

	panel := buildPreisachEquationPanel(win, func(string, string) {})
	if panel == nil {
		t.Fatal("expected non-nil panel")
	}
	if countImages(panel) != 0 {
		t.Fatal("expected Preisach fallback to text-only layout when SVG is missing")
	}
}

func countImages(obj fyne.CanvasObject) int {
	if obj == nil {
		return 0
	}
	count := 0
	if _, ok := obj.(*canvas.Image); ok {
		count++
	}
	if c, ok := obj.(*fyne.Container); ok {
		for _, child := range c.Objects {
			count += countImages(child)
		}
	}
	return count
}
