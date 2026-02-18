package hysteresiscli

import (
	"testing"
)

func TestGetMaterialKnownKeys(t *testing.T) {
	keys := []string{"default", "fecim", "superlattice", "cryogenic", "hzo32", "ftj140", "alscn"}
	for _, k := range keys {
		m := getMaterial(k)
		if m == nil {
			t.Fatalf("getMaterial(%q) returned nil", k)
		}
		if m.Name == "" {
			t.Fatalf("getMaterial(%q) returned empty-name material", k)
		}
	}
}

func TestBuildMaterialResult_UnitConversions(t *testing.T) {
	m := getMaterial("superlattice")
	res := buildMaterialResult(m)
	if res.Material == "" {
		t.Fatal("material name should not be empty")
	}
	if res.RemanentPol <= 0 || res.SaturationPol <= 0 {
		t.Fatalf("polarization conversions invalid: Pr=%v Ps=%v", res.RemanentPol, res.SaturationPol)
	}
	if res.CoerciveField <= 0 {
		t.Fatalf("coercive field conversion invalid: %v", res.CoerciveField)
	}
	if res.Thickness <= 0 {
		t.Fatalf("thickness conversion invalid: %v", res.Thickness)
	}
}
