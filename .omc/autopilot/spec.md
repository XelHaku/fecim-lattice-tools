# Autopilot Spec: Module Refactoring to Shared Utilities

## Overview
Refactor Module 2 (crossbar) and Module 4 (circuits) to use the new shared utilities in `shared/physics` and `shared/peripherals`.

## Functional Requirements

### Module 2 Changes
Replace local definitions with shared imports:
- `DefaultQuantizationLevels` → `physics.DefaultLevels`
- `GMin`, `GMax` → `physics.GMin`, `physics.GMax`
- `ConductanceModel` enum → `physics.ConductanceModel`
- `QuantizeToLevels()` → `physics.QuantizeTo30Levels()`
- `GetLevel()` → `physics.GetLevelFor30()`

### Module 4 Changes
- Verify `ADCType` enum aligns with `peripherals.ADCType`
- Can optionally use `peripherals.DefaultDACConfig()` / `DefaultADCConfig()` for reference

### Backward Compatibility
- Keep type aliases for exported types
- Keep wrapper functions that call shared versions
- All existing tests must pass

## Files to Modify

### Module 2 Core (PRIMARY)
1. `module2-crossbar/pkg/crossbar/array.go` - Main constants and functions
2. `module2-crossbar/pkg/crossbar/drift.go` - Uses GMin, GMax
3. `module2-crossbar/pkg/crossbar/temperature.go` - Uses GMin, GMax
4. `module2-crossbar/pkg/crossbar/enhanced.go` - Uses ConductanceModel

### Module 2 Tests
5. `module2-crossbar/pkg/crossbar/array_test.go`
6. `module2-crossbar/pkg/crossbar/improvements_test.go`

### Module 2 GUI (reference only, minimal changes)
7. `module2-crossbar/pkg/gui/tooltips.go`
8. `module2-crossbar/pkg/gui/callbacks.go`
9. `module2-crossbar/pkg/gui/app_tabs.go`
10. `module2-crossbar/pkg/gui/app_analysis.go`
11. `module2-crossbar/pkg/gui/vectors.go`

### Module 4 (verify alignment only)
12. `module4-circuits/pkg/peripherals/adc.go` - ADCType enum

## Implementation Strategy

### Phase 1: Add Backward-Compatible Aliases in Module 2
In `array.go`, keep existing constants as aliases:
```go
import "fecim-lattice-tools/shared/physics"

// Backward-compatible aliases
const DefaultQuantizationLevels = physics.DefaultLevels
const GMin = physics.GMin
const GMax = physics.GMax

type ConductanceModel = physics.ConductanceModel
const (
    ConductanceLinear = physics.ConductanceLinear
    ConductanceExponential = physics.ConductanceExponential
    ConductanceLookup = physics.ConductanceLookup
)

// Wrapper functions for backward compatibility
func QuantizeToLevels(value float64) float64 {
    return physics.QuantizeTo30Levels(value)
}

func GetLevel(conductance float64) int {
    return physics.GetLevelFor30(conductance)
}
```

### Phase 2: Update Internal Files
Update `drift.go`, `temperature.go`, `enhanced.go` to import from shared/physics directly.

### Phase 3: Verify Module 4 Alignment
Confirm `ADCType` in Module 4 matches `peripherals.ADCType`.

### Phase 4: Run Tests
Verify all 117+ tests pass.

## Acceptance Criteria
1. `go build ./...` succeeds
2. `go test ./...` passes all tests
3. No breaking changes to public API
4. Shared utilities are properly imported

---
**EXPANSION_COMPLETE**
**PLANNING_COMPLETE**
