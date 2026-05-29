//go:build !cgo

package gogpuapp

import (
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
)

func TestFullAppE2ECapturesEveryReleasedModuleFrame(t *testing.T) {
	frames, err := CaptureFrameImagesForKnownModules(720, 480)
	if err != nil {
		t.Fatalf("CaptureFrameImagesForKnownModules: %v", err)
	}

	descriptors := viewmodel.KnownDescriptors()
	if len(frames) != len(descriptors) {
		t.Fatalf("captured frames = %d, want %d known modules", len(frames), len(descriptors))
	}
	for _, descriptor := range descriptors {
		img, ok := frames[descriptor.ID]
		if !ok {
			t.Fatalf("missing captured frame for %s", descriptor.ID)
		}
		signature, colors := offscreenImageSignatureAndColorCount(img)
		if signature == 0 || colors < 4 {
			t.Fatalf("captured frame for %s appears blank: signature=%d colors=%d", descriptor.ID, signature, colors)
		}
	}
}
