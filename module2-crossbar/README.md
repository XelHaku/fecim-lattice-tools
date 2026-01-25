# Module 2: Crossbar Array Simulation

**Ferroelectric Compute-in-Memory - Matrix-Vector Multiplication Visualization**

---

## Quick Start

```bash
# From project root
cd <local-path>

# Run unified app and select "Crossbar" tab
./launch.sh

# OR run standalone
cd module2-crossbar
go build -o crossbar-gui ./cmd/crossbar-gui
./crossbar-gui
```

---

## Documentation

All documentation has been consolidated into `docs/crossbar/`:

### Start Here
- **[Demo Guide](../docs/crossbar/crossbar.demo.md)** - How to run the visualization
- **[ELI5 Explanation](../docs/crossbar/crossbar.ELI5.md)** - Simple water park analogy

### Deep Dive
- **[Physics Reference](../docs/crossbar/crossbar.physics.md)** - Complete technical details
- **[Research Papers](../docs/crossbar/crossbar.research.md)** - Academic citations
- **[Open Source Tools](../docs/crossbar/crossbar.opensource.md)** - Comparison with other simulators

---

## Code Structure

```
module2-crossbar/
├── cmd/crossbar-gui/          # Standalone application entry
├── pkg/
│   ├── crossbar/              # Core simulation
│   │   ├── array.go           # MVM implementation
│   │   ├── nonidealities.go   # IR drop, sneak paths
│   │   └── enhanced.go        # Integrated simulation
│   └── gui/                   # Fyne-based GUI
│       ├── app.go             # Main application
│       ├── embedded.go        # Embeddable interface
│       └── widgets.go         # Custom widgets
└── README.md                  # This file
```

---

## Features

- **30-level conductance quantization** - Matches Dr. Tour's FeCIM specification
- **Interactive heatmaps** - Click cells to see detailed physics data
- **Non-ideality modeling** - IR drop, sneak paths, device variation, ADC quantization
- **Real-time metrics** - Accuracy, energy efficiency, performance
- **Data export** - CSV and JSON formats for validation

---

## Tests

```bash
cd module2-crossbar
go test ./pkg/crossbar -v
```

---

## Related Modules

- **[Module 1: Hysteresis](../module1-hysteresis/)** - P-E curves and ferroelectric switching
- **[Module 3: MNIST](../module3-mnist/)** - Neural network digit recognition demo
- **[Module 6: EDA](../module6-eda/)** - Circuit design and layout tools

---

**Part of:** FeCIM Lattice Tools
**Source:** Dr. external research group's HfO₂-ZrO₂ superlattice research (COSM 2025)
