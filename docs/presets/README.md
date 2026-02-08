# FeCIM Preset System

The preset system allows saving, loading, and sharing named configurations across all FeCIM Lattice Tools modules.

## Overview

Presets enable you to:
- **Save** your current configuration for later use
- **Load** pre-configured settings for specific use cases
- **Share** configurations with colleagues
- **Switch** quickly between educational, research, and demo modes

## Built-in Presets

FeCIM Lattice Tools comes with several built-in presets organized by category:

### Categories

| Category | Icon | Description |
|----------|------|-------------|
| Educational | 📚 | Slower animations, detailed explanations for learning |
| Research | 🔬 | High precision, data logging for analysis |
| Demo | 🎬 | Visual, impressive settings for presentations |
| Custom | 📁 | User-created presets |
| Benchmark | 📊 | Standardized settings for reproducible results |

### Module-Specific Presets

#### Hysteresis Module
- **Basic P-E Loop** - Classic hysteresis demonstration
- **Ferroelectric Switching** - Polarization switching dynamics
- **L-K Dynamics** - Time-domain switching analysis
- **Material Comparison** - Compare HZO vs PZT
- **Temperature Effects** - Temperature-dependent behavior
- **Investor Demo** - Impressive visual demo

#### Crossbar Module
- **Basic MVM Operation** - Simple matrix-vector multiplication
- **IR Drop Analysis** - Voltage drop visualization
- **Sneak Path Effects** - Sneak current demonstration
- **Large Array Analysis** - 128×128 scalability research
- **High Precision Mode** - Maximum ADC resolution
- **Visual Demo** - Animated presentation mode

#### MNIST Module
- **FP32 Baseline** - Standard floating-point inference
- **CIM Introduction** - Compute-in-memory basics
- **Noise Impact Analysis** - Noise effects on accuracy
- **Quantization Study** - Precision comparison
- **Live Demo** - Interactive digit recognition

#### Circuits Module
- **Basic 1T1R Cell** - Memory cell structure
- **Read Operation** - Read cycle visualization
- **Write Operation** - Write cycle visualization
- **Timing Analysis** - Detailed timing diagrams
- **Power Analysis** - Energy consumption

#### Comparison Module
- **FeCIM vs SRAM** - Technology comparison
- **Energy Advantage** - Efficiency focus
- **Technical Briefing** - Business presentation

#### EDA Module
- **EDA Introduction** - Design flow overview
- **Layout Basics** - Physical layout concepts
- **SKY130 PDK Flow** - SkyWater 130nm flow
- **Timing Analysis** - Static timing setup

## Using Presets

### Quick Select

Each module has a preset dropdown in the control panel:

1. Click the preset dropdown (shows "-- Select Preset --")
2. Choose a preset from the list (★ indicates built-in presets)
3. The configuration is applied immediately

### Preset Browser

For full preset management, click the folder icon to open the Preset Browser:

1. **Filter** by category, module, or search terms
2. **Select** a preset to view its details
3. **Apply** to load the configuration
4. **Save Current** to create a new preset from current settings
5. **Import/Export** to share presets as JSON files

### Saving Presets

1. Configure the module to your desired settings
2. Click the save icon (💾) or use "Save Current" in the browser
3. Enter a name and description
4. Choose a category
5. Add tags for easy searching (optional)

### Importing/Exporting

Presets can be shared as JSON files:

**Export:**
1. Select a preset in the browser
2. Click "Export"
3. Choose a save location

**Import:**
1. Click "Import" in the browser
2. Select a preset JSON file
3. The preset is added to your collection

## Preset File Format

Presets are stored as JSON files in the `presets/` directory:

```json
{
  "metadata": {
    "id": "hysteresis-basic-loop",
    "name": "Basic P-E Loop",
    "description": "Classic hysteresis loop demonstration",
    "category": "educational",
    "module": "hysteresis",
    "author": "FeCIM Team",
    "version": "1.0.0",
    "created_at": "2025-02-07T00:00:00Z",
    "updated_at": "2025-02-07T00:00:00Z",
    "tags": ["learning", "hysteresis", "P-E loop"],
    "built_in": true
  },
  "config": {
    "waveform": "Sine Wave",
    "frequency": 0.5,
    "amplitude": 1.5,
    "material": "HZO (optimized)",
    "physics_engine": "preisach",
    "num_levels": 30
  }
}
```

## Configuration Keys

Each module supports specific configuration keys:

### Hysteresis Module
| Key | Type | Description |
|-----|------|-------------|
| waveform | string | Waveform type (Manual, Sine Wave, Triangle Wave, ISPP, Time-Resolved) |
| frequency | float | Waveform frequency in Hz |
| amplitude | float | E-field amplitude as multiple of Ec |
| material | string | Material name |
| physics_engine | string | Physics engine (preisach, lk) |
| num_levels | int | Number of discrete levels |
| time_scale | float | Animation time scale |
| temperature | float | Temperature in Kelvin |

### Crossbar Module
| Key | Type | Description |
|-----|------|-------------|
| array_size | int | Array dimensions (N×N) |
| noise_level | float | Noise as fraction (0-1) |
| adc_bits | int | ADC resolution |
| temperature | float | Temperature in Kelvin |
| architecture | string | Array architecture (passive, 1T1R, 2T1R) |
| colormap | string | Colormap name |

### MNIST Module
| Key | Type | Description |
|-----|------|-------------|
| mode | string | Inference mode (fp, cim) |
| quantization | int | Bit precision |
| noise_enabled | bool | Enable noise simulation |
| noise_level | float | Noise level (0-1) |
| show_weights | bool | Display weight matrices |

### Circuits Module
| Key | Type | Description |
|-----|------|-------------|
| cell_type | string | Cell architecture |
| operation | string | Current operation (read, write) |
| animation_mode | string | Animation type (step, continuous) |
| animation_speed | float | Speed multiplier |

### Comparison Module
| Key | Type | Description |
|-----|------|-------------|
| comparison_mode | string | Comparison type |
| show_energy | bool | Show energy metrics |
| show_market | bool | Show market data |

### EDA Module
| Key | Type | Description |
|-----|------|-------------|
| pdk | string | Target PDK |
| target_freq | float | Target frequency MHz |
| show_flow | bool | Show design flow |

## API Usage

For programmatic access:

```go
import "fecim-lattice-tools/shared/presets"

// Get global preset manager
manager := presets.Global()

// List presets for a module
hysteresisPresets := manager.List(
    presets.FilterByModule(presets.ModuleHysteresis),
    presets.FilterByCategory(presets.CategoryEducational),
)

// Load and apply a preset
preset, err := manager.Load("hysteresis-basic-loop")
if err != nil {
    log.Fatal(err)
}
err = manager.Apply(preset.Metadata.ID)

// Create a new preset from current config
newPreset, err := manager.CreateFromCurrent(
    "My Custom Preset",
    "Description of my preset",
    presets.ModuleHysteresis,
    presets.CategoryCustom,
)

// Export/Import
err = manager.Export("preset-id", "exported_preset.json")
imported, err := manager.Import("preset_file.json")
```

## Best Practices

1. **Use descriptive names** - Make presets easy to identify
2. **Add meaningful descriptions** - Explain what the preset is for
3. **Tag appropriately** - Use tags for searchability
4. **Version your presets** - Update version when making changes
5. **Test before sharing** - Verify presets work as expected
6. **Document dependencies** - Note any special requirements

## Troubleshooting

**Preset not loading:**
- Check that the module matches
- Verify all config keys are valid
- Look for JSON syntax errors in exported files

**Configuration not applied:**
- Ensure the module has registered its provider
- Check for conflicting settings
- Verify the preset is for the current module version
