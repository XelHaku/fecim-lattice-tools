# FeCIM Config Validator

A comprehensive validation system for FeCIM configuration JSON files. Validates structure, types, ranges, and physical constraints.

## Supported Config Types

| Type | Description | Key Indicators |
|------|-------------|----------------|
| `calibration` | Ferroelectric calibration data | `version`, `num_levels`, `calibrations` |
| `preisach_state` | Preisach hysteron states | `hysteron_states` |
| `array_design` | Crossbar array designs | `config.array_rows`, `config.array_cols` |
| `weight_matrix` | Neural network weight matrices | `rows`, `cols`, `weights` |
| `openlane` | OpenLane ASIC flow configs | `DESIGN_NAME`, `VERILOG_FILES` |

## Installation

```bash
# Build the CLI tool
go build -o bin/fecim-config-validate ./validation/configvalidator/cmd/validate

# Or install to GOPATH
go install ./validation/configvalidator/cmd/validate
```

## CLI Usage

```bash
# Validate a single file
fecim-config-validate data/calibrations/fecim_hzo.json

# Validate a directory
fecim-config-validate data/calibrations/

# Recursively validate all JSON files
fecim-config-validate -r .

# Show warnings (not just errors)
fecim-config-validate -w data/

# Summary only (for CI pipelines)
fecim-config-validate -r -s .

# Quiet mode (exit code only)
fecim-config-validate -q config.json && echo "Valid"
```

### Exit Codes

- `0` - All files valid
- `1` - One or more files invalid

## Library Usage

```go
package main

import (
    "fmt"
    "fecim-lattice-tools/validation/configvalidator"
)

func main() {
    // Validate a file
    result, err := configvalidator.ValidateFile("config.json")
    if err != nil {
        panic(err)
    }
    
    if result.Valid {
        fmt.Println("Config is valid!")
    } else {
        for _, e := range result.Errors {
            fmt.Printf("Error: %s\n", e.Error())
        }
    }
    
    // Validate JSON data directly
    jsonData := []byte(`{"version": 2, "num_levels": 30, ...}`)
    result = configvalidator.ValidateJSON(jsonData)
    
    // Validate entire directory
    results, err := configvalidator.ValidateDirectory("data/calibrations")
    for _, r := range results {
        fmt.Printf("%s: valid=%v\n", r.FilePath, r.Valid)
    }
}
```

## Validation Rules

### Calibration Configs

| Field | Type | Constraints |
|-------|------|-------------|
| `version` | int | 1-4 |
| `material_name` | string | non-empty |
| `num_levels` | int | 2-256 |
| `calibrations.<temp>.temperature_k` | int | 1-1000 K |
| `calibrations.<temp>.relax_comp_*` | float[] | 0.0-1.0 |
| All calibration arrays | float[] | length = num_levels |

### Preisach State Configs

| Field | Type | Constraints |
|-------|------|-------------|
| `version` | int | 1-2 |
| `material` | string | non-empty |
| `temperature_k` | int | 1-1000 K |
| `grid_size` | int | 1-1000 |
| `distribution_type` | string | gaussian, lorentzian, bimodal, uniform |
| `hysteron_states` | int[] | values must be -1 or 1 |
| `alpha_sigma`, `beta_sigma` | float | > 0 |
| `correlation` | float | -1.0 to 1.0 |
| `current_wakeup` | float | 0.0-1.0 |

### Array Design Configs

| Field | Type | Constraints |
|-------|------|-------------|
| `config.array_rows` | int | 1-4096 |
| `config.array_cols` | int | 1-4096 |
| `config.levels` | int | 2-256 |
| `config.g_min` | float | ≥ 0, < g_max |
| `config.g_max` | float | > g_min |
| `config.v_prog_min` | float | ≥ 0, < v_prog_max |
| `config.v_prog_max` | float | > v_prog_min, ≤ 10V |
| `cells` | array | length = rows × cols |

### Weight Matrix Configs

| Field | Type | Constraints |
|-------|------|-------------|
| `name` | string | non-empty |
| `rows` | int | 1-65536 |
| `cols` | int | 1-65536 |
| `weights` | float[][] | dimensions match rows × cols |

### OpenLane Configs

| Field | Type | Constraints |
|-------|------|-------------|
| `DESIGN_NAME` | string | valid Verilog identifier |
| `VERILOG_FILES` | string | non-empty |
| `CLOCK_PERIOD` | float | ≥ 0.1 ns |
| `PDK` | string | sky130A, sky130B, gf180mcuC, gf180mcuD, asap7 |
| `FP_SIZING` | string | absolute, relative |
| `DIE_AREA` | string | format "x0 y0 x1 y1" |

## Running Tests

```bash
# Run all tests
go test ./validation/configvalidator/...

# With verbose output
go test -v ./validation/configvalidator/...

# Run benchmarks
go test -bench=. ./validation/configvalidator/...
```

## Integration with CI

Add to your CI pipeline:

```yaml
# GitHub Actions
- name: Validate configs
  run: |
    go build -o validate ./validation/configvalidator/cmd/validate
    ./validate -r -s data/
```

```bash
# Shell script
#!/bin/bash
if ./bin/fecim-config-validate -r -q .; then
    echo "✓ All configs valid"
else
    echo "✗ Config validation failed"
    exit 1
fi
```

## Adding New Config Types

1. Add type constant in `validator.go`
2. Update `detectConfigType()` to recognize the new type
3. Create `<type>.go` with validation function
4. Add tests in `validator_test.go`
