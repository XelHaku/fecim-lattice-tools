# FeCIM Lattice Tools CLI Reference

> **Note:** This file was previously located at `docs/CLI.md`. It has moved to `docs/1-getting-started/cli-reference.md`.

This document describes the command-line interface for all FeCIM Lattice Tools modules.

## Top-level launcher flags (`cmd/fecim-lattice-tools`)

The primary binary exposes these flags:

| Flag | Description | Default |
|------|-------------|---------|
| `--logger` | Enable file logging (or pass level token) | `false` |
| `--verbosity` | Log level (`off\|info\|debug\|trace`) when logger is enabled | `info` |
| `--calibrate` | Run hysteresis calibration and exit | `false` |
| `--material` | Material to calibrate (`all` or explicit name) | `all` |
| `--force` | Force recalibration even if cache exists | `false` |
| `--verify` | Verify calibration quality after calibration | `false` |
| `--list-materials` | Print available materials and exit | `false` |
| `--mode` | Run headless mode and exit (for example `hysteresis`) | empty |
| `--engine` | Headless hysteresis engine: `preisach` or `lk` | `preisach` |
| `--screenshot-dir` | Output directory for GUI screenshots | `screenshots` |
| `--recording-dir` | Output directory for GUI recordings | `recordings` |
| `--module` | Initial module (`home`, `hysteresis`, `crossbar`, `mnist`, `circuits`, `comparison`, `eda`, `docs`) | `home` |

Use `fecim-lattice-tools --help` for the canonical, runtime-generated list.

## Common CLI Flags

All CLI commands support these common flags:

| Flag | Description |
|------|-------------|
| `--json` | Output results as JSON instead of human-readable text |
| `-q, --quiet` | Suppress informational output (only show results/errors) |
| `-c, --config FILE` | Load configuration from YAML or JSON file |
| `--batch FILE` | Process multiple items from a batch file |
| `-o, --output FILE` | Write output to file (default: stdout) |
| `-h, --help` | Show help message |

## Configuration Files

All CLI commands support loading configuration from YAML or JSON files via the `--config` flag.

### Example YAML Config (hysteresis.yaml)
```yaml
material: superlattice
frequency: 1000000
temperature: 300
```

### Example JSON Config (mnist.json)
```json
{
  "hidden_size": 128,
  "noise_level": 0.02,
  "epochs": 10,
  "levels": [8, 16, 24, 31]
}
```

## Batch Processing

Use the `--batch` flag to process multiple items from a file. Batch files can be:

1. **Line-based** (one item per line, `#` for comments):
```
superlattice
fecim
cryogenic
# This is a comment
hzo32
```

2. **JSON array**:
```json
["superlattice", "fecim", "cryogenic", "hzo32"]
```

---

## Module 1: Hysteresis

Visualize ferroelectric hysteresis in HfO2-ZrO2 superlattice materials.

### Basic Usage
```bash
fecim-lattice-tools hysteresis [options]
```

### Options
| Flag | Description | Default |
|------|-------------|---------|
| `--material NAME` | Material type | superlattice |
| `--freq HZ` | Waveform frequency | 1e6 |
| `--headless` | Static ASCII output | false |
| `--tui` | Terminal UI mode | false |
| `--vulkan` | Vulkan graphics mode | false |
| `--list-materials` | List available materials | - |

### Available Materials
- `default` - HZO (Si-doped)
- `fecim` - FeCIM HZO
- `superlattice` - Literature Superlattice (Cheema 2020)
- `cryogenic` - Cryogenic HZO (4K)
- `hzo32` - HZO Standard (32 states)
- `ftj140` - HZO FTJ (140 states)
- `alscn` - AlScN (8-16 states)

Mode precedence (if flags are combined): `--headless` → `--tui` → `--vulkan` → GUI default.

For authoritative run-mode and Preisach/L-K default behavior, see:
`docs/documentation/module1-hysteresis/RUN_MODES.md`.

### Examples

List materials as JSON:
```bash
fecim-lattice-tools hysteresis --list-materials --json
```

Headless mode with JSON output:
```bash
fecim-lattice-tools hysteresis --headless --json --material superlattice
```

Batch process multiple materials:
```bash
echo -e "superlattice\nfecim\ncryogenic" > materials.txt
fecim-lattice-tools hysteresis --headless --batch materials.txt --json
```

Using config file:
```bash
fecim-lattice-tools hysteresis --config hysteresis.yaml --headless
```

---

## Module 2: Crossbar

Crossbar array matrix-vector multiplication (MVM) visualization.

### Basic Usage
```bash
fecim-lattice-tools crossbar inference [options]
```

### Options
| Flag | Description | Default |
|------|-------------|---------|
| `--size N` | Array size (NxN) | 64 |
| `--layers N` | Neural network layers | 3 |
| `--batch N` | Inference batch size | 1 |
| `--noise LEVEL` | Device noise level (0-1) | 0.02 |
| `--adc BITS` | ADC resolution | 6 |
| `--seed N` | Random seed (0=time) | 1 |
| `--no-color` | Disable colored output | false |
| `--benchmark` | Run benchmark | false |
| `--show-array` | Show array state | false |
| `--show-mvm` | Show MVM operation | false |
| `--show-irdrop` | Show IR drop analysis | false |
| `--show-sneak` | Show sneak path analysis | false |
| `--show-nonidealities` | Show all effects | false |

### Examples

Run MVM with JSON output:
```bash
fecim-lattice-tools crossbar inference --show-mvm --json --size 32
```

Using config file:
```bash
fecim-lattice-tools crossbar inference --config crossbar.yaml
```

---

## Module 3: MNIST

MNIST digit recognition on ferroelectric crossbar arrays.

### Basic Usage
```bash
fecim-lattice-tools mnist cli [options]
```

### Options
| Flag | Description | Default |
|------|-------------|---------|
| `--train` | Train the network | false |
| `--evaluate` | Evaluate on test set | false |
| `--interactive` | Interactive mode | false |
| `--epochs N` | Training epochs | 5 |
| `--hidden N` | Hidden layer size | 128 |
| `--noise LEVEL` | Device noise (0-1) | 0.02 |
| `--load FILE` | Load weights from file | - |
| `--save FILE` | Save weights to file | - |
| `--core-eval` | Dual-mode evaluation (FP vs CIM) | false |
| `--core-samples N` | Samples for core-eval | 1000 |
| `--core-levels LIST` | Comma-separated levels | - |
| `--export-levels LIST` | Export quantized weights | - |

### Examples

Evaluate with JSON output:
```bash
fecim-lattice-tools mnist cli --evaluate --json --load weights.json
```

Core evaluation with level sweep:
```bash
fecim-lattice-tools mnist cli --core-eval --core-levels 8,16,24,31 --json
```

Using config file:
```bash
fecim-lattice-tools mnist cli --config mnist.yaml --evaluate
```

---

## Module 4: Circuits

Peripheral circuits visualization (DAC, ADC, TIA, Charge Pump).

### Basic Usage
```bash
fecim-lattice-tools circuits cli [options]
```

### Options
| Flag | Description | Default |
|------|-------------|---------|
| `--dac` | Show DAC details | false |
| `--adc` | Show ADC details | false |
| `--tia` | Show TIA details | false |
| `--pump` | Show Charge Pump details | false |
| `--all` | Show all circuits | false |
| `--linearity` | Show INL/DNL analysis | false |
| `--timing` | Show timing diagrams | false |
| `--power` | Show power breakdown | false |
| `--ispp` | Run ISPP demo | false |
| `--level N` | Demo level (0-29) | 15 |
| `--logger` | Enable file logging | false |
| `--verbosity N` | Log verbosity (0-3) | 2 |

### Examples

Get all circuit specs as JSON:
```bash
fecim-lattice-tools circuits cli --all --json
```

Show specific circuits with JSON output:
```bash
fecim-lattice-tools circuits cli --dac --adc --json
```

Using config file:
```bash
fecim-lattice-tools circuits cli --config circuits.yaml --all
```

---

## Module 5: Comparison

Architecture comparison (CPU vs GPU vs FeCIM).

### Basic Usage
```bash
fecim-lattice-tools comparison cli [options]
```

### Options
| Flag | Description | Default |
|------|-------------|---------|
| `--all` | Show all comparisons | true |
| `--specs` | Show architecture specs | false |
| `--inference` | Show inference comparison | false |
| `--datacenter` | Show datacenter comparison | false |
| `--advantages` | Show FeCIM advantages | false |
| `--workload NAME` | Workload type | mnist |
| `--throughput N` | Target throughput (inf/sec) | 10000 |
| `--no-color` | Disable colored output | false |

### Workloads
- `mnist` - MNIST digit recognition
- `resnet` - ResNet-50 image classification
- `bert` - BERT-base NLP
- `gpt2` - GPT-2 language model
- `llm` - Large language model

### Examples

Compare architectures with JSON output:
```bash
fecim-lattice-tools comparison cli --all --json --workload resnet
```

Specific comparisons:
```bash
fecim-lattice-tools comparison cli --specs --advantages --json
```

Using config file:
```bash
fecim-lattice-tools comparison cli --config comparison.yaml
```

---

## Module 6: EDA

EDA design generation and export.

### Basic Usage
```bash
fecim-lattice-tools eda cli [options]
```

### Options
| Flag | Description | Default |
|------|-------------|---------|
| `--mode MODE` | Operation mode (storage/memory/compute) | compute |
| `--input FILE` | Input weights JSON | - |
| `--output DIR` | Output directory | data |
| `--name NAME` | Design name | fecim_array |
| `--rows N` | Array rows | 128 |
| `--cols N` | Array columns | 128 |
| `--levels N` | Conductance levels (2-30) | 30 |
| `--tech NAME` | Technology (SKY130/GF180MCU/IHP_SG13G2) | SKY130 |
| `--arch TYPE` | Architecture (passive/1T1R) | passive |
| `--vdd V` | Supply voltage | 1.8 |
| `--gmin US` | Min conductance (μS) | 10.0 |
| `--gmax US` | Max conductance (μS) | 100.0 |
| `--export-json` | Export JSON mapping | true |
| `--csv` | Export CSV | true |
| `--spice` | Export SPICE netlist | true |
| `--verilog` | Export Verilog | true |
| `--def` | Export DEF placement | true |
| `--json-output` | Output results as JSON | false |
| `--quiet` | Suppress output | false |
| `--config FILE` | Config file | - |

### Examples

Generate design with JSON result:
```bash
fecim-lattice-tools eda cli --rows 64 --cols 64 --json-output
```

Compute mode with weights:
```bash
fecim-lattice-tools eda cli --mode compute --input weights.json --json-output
```

Using config file:
```bash
fecim-lattice-tools eda cli --config eda.yaml --json-output
```

---

## JSON Output Examples

### Hysteresis Material Info
```json
{
  "material": "Literature Superlattice (Cheema 2020)",
  "remanent_polarization_uC_cm2": 25.0,
  "saturation_polarization_uC_cm2": 30.0,
  "coercive_field_MV_cm": 1.5,
  "coercive_voltage_V": 0.75,
  "thickness_nm": 5.0,
  "permittivity": 30.0,
  "switching_time_ns": 100.0,
  "endurance_cycles": 1e12,
  "discrete_levels": 30,
  "bits_per_cell": 4.91
}
```

### Circuits Result
```json
{
  "dac": {
    "bits": 5,
    "levels": 30,
    "vref_low_v": -1.5,
    "vref_high_v": 1.5,
    "resolution_v": 0.1,
    "energy_fj": 10.5
  },
  "adc": {
    "bits": 5,
    "levels": 30,
    "conversion_time_ns": 100,
    "enob": 4.5,
    "energy_fj": 15.2
  }
}
```

### Comparison Result
```json
{
  "workload": "mnist",
  "target_throughput": 10000,
  "architectures": [
    {
      "name": "CPU+DRAM",
      "tdp_watts": 65,
      "tops_per_watt": 0.1,
      "latency_ms": 5.0,
      "energy_mj": 10.0,
      "throughput_infs": 200
    }
  ],
  "advantages": {
    "vs_cpu_energy_reduction": 100,
    "vs_cpu_cost_reduction": 10,
    "vs_gpu_power_reduction": 50,
    "vs_gpu_area_reduction": 20
  }
}
```

### EDA Result
```json
{
  "design_name": "fecim_array",
  "mode": "compute",
  "rows": 128,
  "cols": 128,
  "total_cells": 16384,
  "active_cells": 16384,
  "area_mm2": 0.25,
  "power_mw": 10.5,
  "throughput_gops": 100.0,
  "technology": "SKY130",
  "output_files": [
    "data/fecim_array_design.json",
    "data/fecim_array_cells.csv",
    "data/fecim_array.sp",
    "data/fecim_array.v",
    "data/fecim_array.def"
  ]
}
```

---

## Scripting Examples

### Batch Processing Script
```bash
#!/bin/bash
# Process all materials and collect results

materials=(superlattice fecim cryogenic hzo32 ftj140 alscn)

for mat in "${materials[@]}"; do
    echo "Processing $mat..."
    fecim-lattice-tools hysteresis --headless --json --material "$mat" \
        -o "results/${mat}.json"
done
```

### JSON Processing with jq
```bash
# Get all material properties
fecim-lattice-tools hysteresis --list-materials --json | jq '.[].name'

# Extract specific field from circuits
fecim-lattice-tools circuits cli --all --json | jq '.dac.energy_fj'

# Filter comparison results
fecim-lattice-tools comparison cli --json | \
    jq '.architectures[] | select(.name | contains("FeCIM"))'
```

### Pipeline Example
```bash
# Generate design, validate, and report
fecim-lattice-tools eda cli --rows 64 --cols 64 --json-output | \
    jq '{cells: .total_cells, area: .area_mm2, power: .power_mw}'
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `FECIM_LOG_LEVEL` | Logging level (debug/info/warn/error) |
| `FECIM_CONFIG_DIR` | Default config directory |

---

## See Also

- [OpenLane CLI Reference](eda/references/cli-reference.md) - External EDA tools
- [API Documentation](../3-develop/api-reference.md) - Programmatic access
- [Examples](../examples/) - Sample code and configs
