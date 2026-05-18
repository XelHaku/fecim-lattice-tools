//go:build !cgo

package gogpuapp

import (
	"encoding/binary"
	"hash/fnv"
	"image"
	"os"
	"os/exec"
	"testing"

	"github.com/gogpu/gg"

	"fecim-lattice-tools/shared/viewmodel"
)

const noopAcceleratorChildEnv = "FECIM_TEST_NOOP_ACCELERATOR_CHILD"

func TestCaptureFrameImageUsesSoftwareOffscreenRasterizerWithNoopAccelerator(t *testing.T) {
	if os.Getenv(noopAcceleratorChildEnv) == "1" {
		if err := gg.RegisterAccelerator(noopAccelerator{}); err != nil {
			t.Fatalf("register noop accelerator: %v", err)
		}
		img, err := CaptureFrameImage(viewmodel.ModuleCircuits, 640, 420)
		if err != nil {
			t.Fatalf("CaptureFrameImage error: %v", err)
		}
		_, colors := offscreenImageSignatureAndColorCount(img)
		if colors < 2 {
			t.Fatalf("captured offscreen frame is blank with noop accelerator: %d unique colors", colors)
		}
		return
	}

	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("test executable: %v", err)
	}
	cmd := exec.Command(exe, "-test.run=^TestCaptureFrameImageUsesSoftwareOffscreenRasterizerWithNoopAccelerator$", "-test.v")
	cmd.Env = append(os.Environ(), noopAcceleratorChildEnv+"=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("child offscreen capture test failed:\n%s", out)
	}
}

type noopAccelerator struct{}

func (noopAccelerator) Name() string { return "noop-test" }
func (noopAccelerator) Init() error  { return nil }
func (noopAccelerator) Close()       {}
func (noopAccelerator) CanAccelerate(gg.AcceleratedOp) bool {
	return true
}
func (noopAccelerator) FillPath(gg.GPURenderTarget, *gg.Path, *gg.Paint) error {
	return nil
}
func (noopAccelerator) StrokePath(gg.GPURenderTarget, *gg.Path, *gg.Paint) error {
	return nil
}
func (noopAccelerator) FillShape(gg.GPURenderTarget, gg.DetectedShape, *gg.Paint) error {
	return nil
}
func (noopAccelerator) StrokeShape(gg.GPURenderTarget, gg.DetectedShape, *gg.Paint) error {
	return nil
}
func (noopAccelerator) Flush(gg.GPURenderTarget) error {
	return nil
}
func (noopAccelerator) DrawText(gg.GPURenderTarget, any, string, float64, float64, gg.RGBA, gg.Matrix, float64) error {
	return nil
}
func (noopAccelerator) DrawGlyphMaskText(gg.GPURenderTarget, any, string, float64, float64, gg.RGBA, gg.Matrix, float64) error {
	return nil
}

func offscreenImageSignatureAndColorCount(img image.Image) (uint64, int) {
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
			colors[[4]uint32{r, g, b, a}] = struct{}{}
			binary.LittleEndian.PutUint32(buf[0:4], r)
			binary.LittleEndian.PutUint32(buf[4:8], g)
			binary.LittleEndian.PutUint32(buf[8:12], b)
			binary.LittleEndian.PutUint32(buf[12:16], a)
			_, _ = hash.Write(buf[:])
		}
	}
	return hash.Sum64(), len(colors)
}
