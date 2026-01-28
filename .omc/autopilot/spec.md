# VOLTAGE_RULES.md Specification

## Requirements (from Analyst)

### Functional Requirements
1. Document must cover read/write/compute operations for passive (0T1R), 1T1R, and 2T1R modes
2. 9 core sections minimum (3 operations × 3 architectures)
3. Each voltage specification needs Source column (peer-reviewed, codebase, or "ESTIMATED")
4. Must include code snippet mappings to codebase
5. Cross-reference at least 5 existing documents

### Non-Functional Requirements
- Markdown format, GitHub-compatible
- ASCII diagrams (version control friendly)
- Copy-paste ready quick reference card
- Consistent table formatting

### Implicit Requirements
- Temperature: Room temperature (300K) nominal values only
- Array sizes: Reference sizes (8x8, 32x32, 128x128)
- Timing: Exclude (separate document scope)

### Out of Scope
- Temperature-dependent voltage scaling (reference only)
- Pulse width/timing specifications
- PVT corner analysis
- Circuit implementation details (defer to PHYSICS.md)

---

## Voltage Data Collected

### Core Constants (from codebase)

| Parameter | Value | Source |
|-----------|-------|--------|
| DACVrefHigh | +1.5V | shared/peripherals/defaults.go:19 |
| DACVrefLow | -1.5V | shared/peripherals/defaults.go:22 |
| ADCVrefHigh | +1.0V | shared/peripherals/defaults.go:31 |
| ADCVrefLow | 0.0V | shared/peripherals/defaults.go:34 |
| Charge Pump Input | 1.0V | chargepump.go:22 |
| Charge Pump Output | 1.5V | chargepump.go:23 |
| TIA Max Output | 1.0V | tia.go:26 |

### Material Parameters (from physics.yaml)

| Parameter | Value | Source |
|-----------|-------|--------|
| Coercive Field (Ec) | 0.6-1.5 MV/cm | Peer-reviewed: Nature Commun. 2025 |
| Coercive Voltage (Vc) | 0.6-1.5V (at 10nm) | ESTIMATED: Vc = Ec × thickness |
| Half-Select (V/2) | 0.75V | ESTIMATED: Vwrite/2 |

### Operation Voltages

| Operation | Voltage Range | Notes |
|-----------|---------------|-------|
| Write (positive) | +1.2 to +1.5V | Above Vc with margin |
| Write (negative) | -1.2 to -1.5V | Erase operation |
| Read | +0.1 to +0.5V | Below Vc (non-destructive) |
| Compute (input) | 0 to +1.0V | MVM input range |

---

## Technical Specification

### Document Location
```
<local-path>
```

### Section Outline

```markdown
# FeCIM Voltage Rules: Architecture-Specific Operating Voltages

## Table of Contents
1. Overview
2. Voltage Constants Summary
3. Passive (0T1R) Mode
   - 3.1 Read Operation
   - 3.2 Write Operation
   - 3.3 Compute Operation
4. 1T1R Mode
   - 4.1 Read Operation
   - 4.2 Write Operation
   - 4.3 Compute Operation
5. 2T1R Mode
   - 5.1 Read Operation
   - 5.2 Write Operation
   - 5.3 Compute Operation
6. Voltage Biasing Schemes
7. Code Mappings
8. ASCII Diagrams
9. References
10. Quick Reference Card
```

### Required Cross-References

| Reference | Path |
|-----------|------|
| ARCHITECTURES.md | docs/crossbar/ARCHITECTURES.md |
| PHYSICS.md (crossbar) | docs/crossbar/PHYSICS.md |
| PHYSICS.md (peripheral) | docs/peripheral-circuits/PHYSICS.md |
| physics.yaml | config/physics.yaml |
| defaults.go | shared/peripherals/defaults.go |
| dac.go | module4-circuits/pkg/peripherals/dac.go |
| adc.go | module4-circuits/pkg/peripherals/adc.go |
| chargepump.go | module4-circuits/pkg/peripherals/chargepump.go |
| array.go | module2-crossbar/pkg/crossbar/array.go |
| sneakpath.go | module2-crossbar/pkg/crossbar/sneakpath.go |

### Required ASCII Diagrams

1. **Voltage Levels Overview** - Rail diagram showing all voltage levels
2. **0T1R Half-Select Biasing** - V/2 scheme for passive arrays
3. **1T1R Isolation** - Transistor gating mechanism

---

## Verification Checklist

- [ ] All 9 core sections present (3 operations x 3 architectures)
- [ ] Every voltage value has Source column
- [ ] Cross-references to 10+ files
- [ ] ASCII diagrams render correctly
- [ ] Code snippets have file paths
- [ ] Quick reference card included
- [ ] No claims without evidence

---

**EXPANSION_COMPLETE**
