package validation

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	sharederrors "fecim-lattice-tools/shared/errors"
)

var (
	errPhysicsBadMaterial = errors.New("physics: unknown material")
	errPeripheralBadADC   = errors.New("peripheral: invalid ADC config")
)

func physicsLayerLoadMaterial(name string) error {
	if name == "" || strings.Contains(strings.ToLower(name), "bad") {
		return sharederrors.InvalidData("material", fmt.Sprintf("unsupported material %q", name)).
			WithDetails("requested material is not in the physics material library").
			WithCause(errPhysicsBadMaterial)
	}
	return nil
}

func controllerLayerConfigureMaterial(name string) error {
	if err := physicsLayerLoadMaterial(name); err != nil {
		return sharederrors.WrapWithCategory(err, sharederrors.CategoryConfig,
			"controller: failed to configure material for write path").
			WithDetails("material=%q", name)
	}
	return nil
}

func peripheralLayerBuildADC(bits int, vrefLow, vrefHigh float64) error {
	if bits <= 0 {
		return sharederrors.InvalidParameter("adc.bits", bits, "bits > 0").WithCause(errPeripheralBadADC)
	}
	if vrefHigh <= vrefLow {
		return sharederrors.InvalidParameter("adc.vref", []float64{vrefLow, vrefHigh}, "vrefHigh > vrefLow").WithCause(errPeripheralBadADC)
	}
	return nil
}

func guiLayerInitializeADC(bits int, vrefLow, vrefHigh float64) error {
	if err := peripheralLayerBuildADC(bits, vrefLow, vrefHigh); err != nil {
		return sharederrors.Wrap(err, "gui: failed to initialize peripheral panel")
	}
	return nil
}

func TestErrorPropagation_PhysicsToController(t *testing.T) {
	err := controllerLayerConfigureMaterial("bad-material")
	if err == nil {
		t.Fatal("expected error for bad material")
	}

	if !strings.Contains(err.Error(), "controller: failed to configure material") {
		t.Fatalf("expected controller context in error, got: %v", err)
	}
	if !errors.Is(err, errPhysicsBadMaterial) {
		t.Fatalf("expected wrapped physics root cause via errors.Is")
	}

	var fecimErr *sharederrors.FeCIMError
	if !errors.As(err, &fecimErr) {
		t.Fatalf("expected FeCIMError in chain")
	}
	if fecimErr.Category != sharederrors.CategoryConfig {
		t.Fatalf("expected controller-level CategoryConfig, got %v", fecimErr.Category)
	}
	if !strings.Contains(fecimErr.Details, `material="bad-material"`) {
		t.Fatalf("expected material context in details, got %q", fecimErr.Details)
	}
}

func TestErrorPropagation_PeripheralToGUI(t *testing.T) {
	err := guiLayerInitializeADC(0, 1.0, 0.0)
	if err == nil {
		t.Fatal("expected error for invalid ADC config")
	}

	if !strings.Contains(err.Error(), "gui: failed to initialize peripheral panel") {
		t.Fatalf("expected GUI context in error, got: %v", err)
	}
	if !errors.Is(err, errPeripheralBadADC) {
		t.Fatalf("expected wrapped peripheral root cause via errors.Is")
	}

	var fecimErr *sharederrors.FeCIMError
	if !errors.As(err, &fecimErr) {
		t.Fatalf("expected FeCIMError in GUI chain")
	}
	if fecimErr.Category != sharederrors.CategoryUser {
		t.Fatalf("expected user-input category to survive wrap, got %v", fecimErr.Category)
	}
}

func TestErrorTypesImplementErrorInterface(t *testing.T) {
	var _ error = (*sharederrors.FeCIMError)(nil)

	types := []error{
		sharederrors.ConfigError("x"),
		sharederrors.ConfigNotFound("physics"),
		sharederrors.IOError("io", errors.New("disk")),
		sharederrors.FileNotFound("/tmp/file"),
		sharederrors.DataError("bad"),
		sharederrors.InvalidData("material", "missing field"),
		sharederrors.ResourceError("GPU", "OOM"),
		sharederrors.GPUError("failed", errors.New("driver")),
		sharederrors.InternalError("bug"),
		sharederrors.UserInputError("bad input"),
		sharederrors.InvalidParameter("adc.bits", 0, ">0"),
		sharederrors.ErrNotInitialized,
		sharederrors.ErrAlreadyInitialized,
		sharederrors.ErrClosed,
		sharederrors.ErrTimeout,
		sharederrors.ErrCanceled,
		sharederrors.ErrNotSupported,
	}

	for i, err := range types {
		if err == nil {
			t.Fatalf("error at index %d is nil", i)
		}
		if err.Error() == "" {
			t.Fatalf("error at index %d has empty Error() string", i)
		}
	}
}

func TestErrorWrappingPreservesOriginalCause(t *testing.T) {
	root := errors.New("root-cause")

	wrapped := sharederrors.WrapWithCategory(root, sharederrors.CategoryIO, "read calibration failed")
	wrapped = sharederrors.Wrap(wrapped, "controller: startup failed")
	finalErr := fmt.Errorf("gui: launch failed: %w", wrapped)

	if !errors.Is(finalErr, root) {
		t.Fatalf("errors.Is should find original root cause through all wrappers")
	}

	var extracted *sharederrors.FeCIMError
	if !errors.As(finalErr, &extracted) {
		t.Fatalf("errors.As should extract FeCIMError through all wrappers")
	}
	if extracted.Category != sharederrors.CategoryIO {
		t.Fatalf("expected preserved category CategoryIO, got %v", extracted.Category)
	}
}
