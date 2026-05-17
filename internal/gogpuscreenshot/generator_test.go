//go:build !cgo

package gogpuscreenshot

import (
	"encoding/binary"
	"hash/fnv"
	"image"
	"image/png"
	"os"
	"testing"

	"fecim-lattice-tools/internal/gogpuapp"
	"fecim-lattice-tools/shared/viewmodel"
)

func TestGenerateCapturesRealGogpuAppFrame(t *testing.T) {
	opts := DefaultOptions()
	opts.OutputDir = t.TempDir()
	opts.Only = "docs"
	opts.Width = 420
	opts.Height = 260

	if err := Generate(opts); err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	got := readPNG(t, opts.OutputPath("docs-overview.png"))
	want, err := gogpuapp.CaptureFrameImage(viewmodel.ModuleDocs, opts.Width, opts.Height)
	if err != nil {
		t.Fatalf("CaptureFrameImage error: %v", err)
	}

	if !imagesEqual(got, want) {
		t.Fatal("generated screenshot did not match the real gogpu app frame")
	}
}

func TestGenerateAllModulesWritesDistinctNonBlankPNGs(t *testing.T) {
	opts := DefaultOptions()
	opts.OutputDir = t.TempDir()
	opts.Width = 900
	opts.Height = 600

	if err := Generate(opts); err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	assertAllModuleScreenshotsValid(t, opts)
}

func assertAllModuleScreenshotsValid(t *testing.T, opts Options) {
	t.Helper()

	entries, err := os.ReadDir(opts.OutputDir)
	if err != nil {
		t.Fatalf("read output dir %s: %v", opts.OutputDir, err)
	}
	if len(entries) != len(appFrameScreenshots) {
		t.Fatalf("generated file count = %d, want %d", len(entries), len(appFrameScreenshots))
	}

	signatures := map[uint64]string{}
	for _, screenshot := range appFrameScreenshots {
		img := readPNG(t, opts.OutputPath(screenshot.filename))
		bounds := img.Bounds()
		if bounds.Dx() != opts.Width || bounds.Dy() != opts.Height {
			t.Fatalf("%s dimensions = %dx%d, want %dx%d", screenshot.filename, bounds.Dx(), bounds.Dy(), opts.Width, opts.Height)
		}

		signature, colorCount := imageSignatureAndColorCount(img)
		if colorCount < 2 {
			t.Fatalf("%s is blank: only %d unique colors", screenshot.filename, colorCount)
		}
		if prior, exists := signatures[signature]; exists {
			t.Fatalf("%s pixel signature matched %s", screenshot.filename, prior)
		}
		signatures[signature] = screenshot.filename
	}
}

func readPNG(t *testing.T, path string) image.Image {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return img
}

func imageSignatureAndColorCount(img image.Image) (uint64, int) {
	bounds := img.Bounds()
	hash := fnv.New64a()
	colors := map[[4]uint32]struct{}{}
	var buf [16]byte
	binary.LittleEndian.PutUint32(buf[0:4], uint32(bounds.Dx()))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(bounds.Dy()))
	_, _ = hash.Write(buf[0:8])

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			key := [4]uint32{r, g, b, a}
			colors[key] = struct{}{}
			binary.LittleEndian.PutUint32(buf[0:4], r)
			binary.LittleEndian.PutUint32(buf[4:8], g)
			binary.LittleEndian.PutUint32(buf[8:12], b)
			binary.LittleEndian.PutUint32(buf[12:16], a)
			_, _ = hash.Write(buf[:])
		}
	}
	return hash.Sum64(), len(colors)
}

func imagesEqual(a, b image.Image) bool {
	if !a.Bounds().Eq(b.Bounds()) {
		return false
	}
	bounds := a.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if a.At(x, y) != b.At(x, y) {
				return false
			}
		}
	}
	return true
}
