package main

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

func TestSavePNG_WritesNonEmptyFile(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "shot.png")

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	img.Set(1, 1, color.RGBA{B: 255, A: 255})

	if err := savePNG(out, img); err != nil {
		t.Fatalf("savePNG failed: %v", err)
	}

	st, err := os.Stat(out)
	if err != nil {
		t.Fatalf("stat output failed: %v", err)
	}
	if st.Size() == 0 {
		t.Fatal("expected non-empty png output")
	}
}
