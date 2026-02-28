package render3d

import (
	"image"
	"math"
	"testing"
)

func TestVec3_Basic(t *testing.T) {
	a := Vec3{1, 2, 3}
	b := Vec3{4, 5, 6}

	// Add
	sum := a.Add(b)
	if sum.X != 5 || sum.Y != 7 || sum.Z != 9 {
		t.Errorf("Add: got %v, want {5,7,9}", sum)
	}

	// Sub
	diff := b.Sub(a)
	if diff.X != 3 || diff.Y != 3 || diff.Z != 3 {
		t.Errorf("Sub: got %v, want {3,3,3}", diff)
	}

	// Scale
	scaled := a.Scale(2)
	if scaled.X != 2 || scaled.Y != 4 || scaled.Z != 6 {
		t.Errorf("Scale: got %v, want {2,4,6}", scaled)
	}

	// Dot
	dot := a.Dot(b)
	if dot != 32 { // 1*4 + 2*5 + 3*6 = 32
		t.Errorf("Dot: got %f, want 32", dot)
	}

	// Length
	unit := Vec3{1, 0, 0}
	if l := unit.Length(); math.Abs(l-1.0) > 1e-10 {
		t.Errorf("Length of unit X: got %f, want 1.0", l)
	}

	// Normalize
	v := Vec3{3, 0, 0}
	n := v.Normalize()
	if math.Abs(n.X-1.0) > 1e-10 || math.Abs(n.Y) > 1e-10 || math.Abs(n.Z) > 1e-10 {
		t.Errorf("Normalize: got %v, want {1,0,0}", n)
	}

	// Normalize zero vector
	zero := Vec3{0, 0, 0}
	nz := zero.Normalize()
	if nz.X != 0 || nz.Y != 0 || nz.Z != 0 {
		t.Errorf("Normalize zero: got %v, want {0,0,0}", nz)
	}
}

func TestMat4_Identity(t *testing.T) {
	m := Identity()
	v := Vec3{3, 7, 11}
	sx, sy := m.Project(v)
	// Identity should pass through X and Y unchanged (projection ignores Z in 2D output)
	if math.Abs(sx-3) > 1e-10 || math.Abs(sy-7) > 1e-10 {
		t.Errorf("Identity project: got (%f, %f), want (3, 7)", sx, sy)
	}
}

func TestMat4_Multiply(t *testing.T) {
	// Scale(2) * Translate(1,0,0) should put (0,0,0) at (2,0)
	s := ScaleMat(2, 2, 2)
	tr := Translate(1, 0, 0)
	m := s.Multiply(tr)
	sx, sy := m.Project(Vec3{0, 0, 0})
	if math.Abs(sx-2) > 1e-10 || math.Abs(sy) > 1e-10 {
		t.Errorf("Scale*Translate at origin: got (%f, %f), want (2, 0)", sx, sy)
	}
}

func TestMat4_Projection(t *testing.T) {
	// Standard isometric: azimuth=pi/6, elevation=pi/6
	proj := NewIsometricProjection(math.Pi/6, math.Pi/6, 100)

	// Origin should project to (0,0) before offset
	sx, sy := proj.Project(Vec3{0, 0, 0})
	if math.Abs(sx) > 1e-6 || math.Abs(sy) > 1e-6 {
		t.Errorf("Origin: got (%f, %f), want (0, 0)", sx, sy)
	}

	// Point at (1,0,0) should project to a non-zero X
	sx2, _ := proj.Project(Vec3{1, 0, 0})
	if sx2 == 0 {
		t.Error("Point (1,0,0) should not project to X=0")
	}

	// Point at (0,1,0) should project upward (negative Y in screen space due to flip)
	_, sy3 := proj.Project(Vec3{0, 1, 0})
	if sy3 >= 0 {
		t.Errorf("Point (0,1,0) should project to negative screen Y, got %f", sy3)
	}
}

func TestMat4_RotateY(t *testing.T) {
	// Rotate 90 degrees should move (1,0,0) to approximately (0,0,-1)
	m := RotateY(math.Pi / 2)
	v := m.TransformVec3(Vec3{1, 0, 0})
	if math.Abs(v.X) > 1e-10 || math.Abs(v.Z+1) > 1e-10 {
		t.Errorf("RotateY(pi/2) of (1,0,0): got %v, want ~(0,0,-1)", v)
	}
}

func TestStackRenderer_SingleLayer(t *testing.T) {
	s := NewStackRenderer()
	s.Width = 400
	s.Height = 300

	// Create a single 8x8 layer with gradient values
	data := make([]float64, 64)
	for i := range data {
		data[i] = float64(i) / 63.0
	}
	s.Layers = []LayerData{
		{Values: data, Rows: 8, Cols: 8, Label: "Layer 0"},
	}

	img := s.Render()

	// Verify it's a valid image with the expected dimensions
	bounds := img.Bounds()
	if bounds.Dx() != 400 || bounds.Dy() != 300 {
		t.Errorf("Image size: got %dx%d, want 400x300", bounds.Dx(), bounds.Dy())
	}

	// Verify it's not all one color (rendering actually happened)
	rgba := img.(*image.RGBA)
	nonBG := 0
	for i := 0; i < len(rgba.Pix); i += 4 {
		// Background is (25, 25, 35, 255)
		if rgba.Pix[i] != 25 || rgba.Pix[i+1] != 25 || rgba.Pix[i+2] != 35 {
			nonBG++
		}
	}
	if nonBG == 0 {
		t.Error("Image appears to be entirely background - no cells rendered")
	}
}

func TestStackRenderer_MultiLayer(t *testing.T) {
	s := NewStackRenderer()
	s.Width = 400
	s.Height = 300

	// Create 4 layers of 8x8
	for l := 0; l < 4; l++ {
		data := make([]float64, 64)
		for i := range data {
			data[i] = float64(l) / 3.0 // Each layer has uniform color
		}
		s.Layers = append(s.Layers, LayerData{
			Values: data, Rows: 8, Cols: 8, Label: "",
		})
	}

	img := s.Render()
	bounds := img.Bounds()
	if bounds.Dx() != 400 || bounds.Dy() != 300 {
		t.Errorf("Image size: got %dx%d, want 400x300", bounds.Dx(), bounds.Dy())
	}

	// Verify rendering happened
	rgba := img.(*image.RGBA)
	nonBG := 0
	for i := 0; i < len(rgba.Pix); i += 4 {
		if rgba.Pix[i] != 25 || rgba.Pix[i+1] != 25 || rgba.Pix[i+2] != 35 {
			nonBG++
		}
	}
	if nonBG == 0 {
		t.Error("Multi-layer image appears to be entirely background")
	}
}

func TestStackRenderer_512Layers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping 512-layer test in short mode")
	}

	s := NewStackRenderer()
	s.Width = 400
	s.Height = 300
	s.MaxVisibleLayers = 32 // Cap visible layers for performance

	// Create 512 layers of 4x4
	for l := 0; l < 512; l++ {
		data := make([]float64, 16)
		for i := range data {
			data[i] = float64(l) / 511.0
		}
		s.Layers = append(s.Layers, LayerData{
			Values: data, Rows: 4, Cols: 4, Label: "",
		})
	}

	// This should not OOM or take excessively long
	img := s.Render()
	bounds := img.Bounds()
	if bounds.Dx() != 400 || bounds.Dy() != 300 {
		t.Errorf("Image size: got %dx%d, want 400x300", bounds.Dx(), bounds.Dy())
	}

	// Verify rendering happened
	rgba := img.(*image.RGBA)
	nonBG := 0
	for i := 0; i < len(rgba.Pix); i += 4 {
		if rgba.Pix[i] != 25 || rgba.Pix[i+1] != 25 || rgba.Pix[i+2] != 35 {
			nonBG++
		}
	}
	if nonBG == 0 {
		t.Error("512-layer image appears to be entirely background")
	}
}

func TestStackRenderer_EmptyLayers(t *testing.T) {
	s := NewStackRenderer()
	s.Width = 200
	s.Height = 150

	// No layers - should not panic
	img := s.Render()
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 150 {
		t.Errorf("Empty image size: got %dx%d, want 200x150", bounds.Dx(), bounds.Dy())
	}
}

func TestStackRenderer_ZeroDimension(t *testing.T) {
	s := NewStackRenderer()
	s.Width = 0
	s.Height = 0

	// Should not panic
	img := s.Render()
	if img == nil {
		t.Error("Render with zero dimensions returned nil")
	}
}

func TestHitTest_Basic(t *testing.T) {
	s := NewStackRenderer()
	s.Width = 400
	s.Height = 300

	// Single layer centered in the viewport
	data := make([]float64, 64)
	for i := range data {
		data[i] = 0.5
	}
	s.Layers = []LayerData{
		{Values: data, Rows: 8, Cols: 8, Label: "Layer 0"},
	}

	// Hit test at the center of the image should hit layer 0
	layer := s.HitTest(float64(s.Width)/2, float64(s.Height)/2)
	if layer != 0 {
		t.Errorf("HitTest at center: got layer %d, want 0", layer)
	}

	// Hit test far outside should return -1
	layer = s.HitTest(0, 0)
	if layer != -1 {
		t.Errorf("HitTest at corner: got layer %d, want -1", layer)
	}
}

func TestHitTest_NoLayers(t *testing.T) {
	s := NewStackRenderer()
	s.Width = 400
	s.Height = 300

	// No layers
	layer := s.HitTest(200, 150)
	if layer != -1 {
		t.Errorf("HitTest with no layers: got %d, want -1", layer)
	}
}

func TestColormap_Range(t *testing.T) {
	cmaps := []struct {
		name string
		fn   ColormapFunc
	}{
		{"viridis", ViridisColor},
		{"plasma", PlasmaColor},
		{"coolwarm", CoolwarmColor},
	}

	for _, cm := range cmaps {
		// Test at boundaries and mid-point
		for _, val := range []float64{0, 0.5, 1.0} {
			c := cm.fn(val)
			if c.A != 255 {
				t.Errorf("%s(%f): alpha = %d, want 255", cm.name, val, c.A)
			}
		}

		// Test clamping
		c := cm.fn(-0.5)
		if c.A != 255 {
			t.Errorf("%s(-0.5): alpha = %d, want 255", cm.name, c.A)
		}
		c = cm.fn(1.5)
		if c.A != 255 {
			t.Errorf("%s(1.5): alpha = %d, want 255", cm.name, c.A)
		}
	}
}

func TestGetColormap(t *testing.T) {
	// Known colormaps
	for _, name := range []string{"viridis", "plasma", "coolwarm"} {
		fn := GetColormap(name)
		if fn == nil {
			t.Errorf("GetColormap(%q) returned nil", name)
		}
		// Verify it produces a color
		c := fn(0.5)
		if c.A != 255 {
			t.Errorf("GetColormap(%q)(0.5): alpha = %d, want 255", name, c.A)
		}
	}

	// Unknown colormap should fall back to viridis
	fn := GetColormap("unknown")
	if fn == nil {
		t.Error("GetColormap(\"unknown\") returned nil")
	}
}

func TestSelectVisibleLayers(t *testing.T) {
	s := NewStackRenderer()

	// When visible >= total, return all
	indices := s.selectVisibleLayers(5, 10)
	if len(indices) != 5 {
		t.Errorf("selectVisibleLayers(5, 10): got %d indices, want 5", len(indices))
	}

	// When subsampling, always include first and last
	indices = s.selectVisibleLayers(100, 5)
	if indices[0] != 0 {
		t.Error("Subsampled layers should start at 0")
	}
	if indices[len(indices)-1] != 99 {
		t.Errorf("Subsampled layers should end at 99, got %d", indices[len(indices)-1])
	}
	if len(indices) > 5 {
		t.Errorf("selectVisibleLayers(100, 5): got %d indices, want <= 5", len(indices))
	}
}

func TestPointInQuad(t *testing.T) {
	// Unit square
	corners := [4][2]float64{
		{0, 0}, {1, 0}, {1, 1}, {0, 1},
	}

	// Inside
	if !pointInQuad(0.5, 0.5, corners) {
		t.Error("Center of unit square should be inside")
	}

	// Outside
	if pointInQuad(2, 2, corners) {
		t.Error("Point (2,2) should be outside unit square")
	}

	// On edge (may be either true or false depending on convention, just verify no panic)
	_ = pointInQuad(0, 0.5, corners)
}

func TestFillQuad(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Fill a small quad
	corners := [4][2]float64{
		{20, 20}, {40, 20}, {40, 40}, {20, 40},
	}
	c := ViridisColor(0.5)
	fillQuad(img, corners, c)

	// Center of the quad should be filled
	px := img.RGBAAt(30, 30)
	if px.R == 0 && px.G == 0 && px.B == 0 && px.A == 0 {
		t.Error("Center of filled quad should not be transparent black")
	}
}

func TestDrawLine(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	c := ViridisColor(1.0)
	drawLine(img, 10, 10, 50, 50, c)

	// Midpoint of the diagonal should be drawn
	px := img.RGBAAt(30, 30)
	if px.A == 0 {
		t.Error("Midpoint of drawn line should not be transparent")
	}
}

func BenchmarkStackRenderer_16Layers_64x64(b *testing.B) {
	s := NewStackRenderer()
	s.Width = 800
	s.Height = 600

	// Create 16 layers of 64x64
	for l := 0; l < 16; l++ {
		data := make([]float64, 64*64)
		for i := range data {
			data[i] = float64(i) / float64(len(data)-1)
		}
		s.Layers = append(s.Layers, LayerData{
			Values: data, Rows: 64, Cols: 64,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Render()
	}
}

func BenchmarkStackRenderer_32Layers_8x8(b *testing.B) {
	s := NewStackRenderer()
	s.Width = 800
	s.Height = 600

	for l := 0; l < 32; l++ {
		data := make([]float64, 64)
		for i := range data {
			data[i] = float64(i) / 63.0
		}
		s.Layers = append(s.Layers, LayerData{
			Values: data, Rows: 8, Cols: 8,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Render()
	}
}

func BenchmarkStackRenderer_512Layers_4x4_Capped(b *testing.B) {
	s := NewStackRenderer()
	s.Width = 400
	s.Height = 300
	s.MaxVisibleLayers = 32

	for l := 0; l < 512; l++ {
		data := make([]float64, 16)
		for i := range data {
			data[i] = float64(l) / 511.0
		}
		s.Layers = append(s.Layers, LayerData{
			Values: data, Rows: 4, Cols: 4,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Render()
	}
}
