# Getting Started with FeCIM Lattice Tools

**Quick start guide for new users - from installation to running your first demo.**

---

## 📋 Prerequisites

### System Requirements

- **OS:** Linux, macOS, or Windows
- **Memory:** 4GB RAM minimum, 8GB recommended
- **Disk:** 500MB free space
- **Display:** 1280×720 minimum resolution

### Required Software

- **Go:** Version 1.24 or later ([download](https://go.dev/dl/))
- **GCC:** C compiler (for Fyne GUI)
- **Git:** For cloning the repository

See [installation.md](installation.md) for platform-specific instructions.

---

## 🚀 Quick Start (5 Minutes)

### 1. Install Dependencies

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get install -y gcc libgl1-mesa-dev xorg-dev
```

**macOS:**
```bash
xcode-select --install
```

**Windows:**
- Install TDM-GCC or MinGW-w64
- See [installation.md](installation.md#windows) for details

### 2. Clone Repository

```bash
git clone https://github.com/[your-repo]/fecim-lattice-tools.git
cd fecim-lattice-tools
```

### 3. Build

```bash
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
```

### 4. Run

```bash
./fecim-lattice-tools
```

The GUI will launch with 7 interactive modules!

---

## 📚 What's Next?

After installation, explore the documentation:

### For Learners
1. Read [ELI5 Overview](../2-learn/eli5-overview.md)
2. Watch [demo videos](demo-videos/)
3. Try [Module 1: Hysteresis](../2-learn/module1-hysteresis/)
4. Progress through modules 2-7

### For Developers
1. Read [API Reference](../3-develop/api-reference.md)
2. Study [Architecture](../3-develop/architecture/)
3. Review [Testing Guide](../3-develop/testing/)
4. Check [Code Quality](../3-develop/code-quality.md)

### For Researchers
1. Review [Honesty Audit](../4-research/honesty-audit.md)
2. Read [Physics Validation](../4-research/physics-validation.md)
3. Browse [Research Papers](../4-research/papers/)
4. Check [Literature Reviews](../4-research/literature-review/)

---

## 📖 Documentation Map

| Document | Purpose | Time |
|----------|---------|------|
| [installation.md](installation.md) | Detailed install instructions | 10 min |
| [runbook.md](runbook.md) | Operations & troubleshooting | 15 min |
| [cli-reference.md](cli-reference.md) | Command-line options | 5 min |
| [demo-videos/](demo-videos/) | Visual walkthroughs | 15 min |

---

## 🎮 Try the Demos

### Demo 1: Hysteresis Loop (2 minutes)

Visualize how ferroelectric materials remember:

1. Launch app: `./fecim-lattice-tools`
2. Select **Module 1: Hysteresis**
3. Click "AC Mode" to see the P-E loop
4. Try different materials from the dropdown

**What you'll see:** The butterfly-shaped hysteresis curve showing path-dependent memory.

### Demo 2: Crossbar Computation (3 minutes)

Watch matrix multiplication happen in hardware:

1. Select **Module 2: Crossbar**
2. Click "Load MNIST Weights"
3. Apply input vector
4. See output currents compute instantly

**What you'll see:** Real-time MVM with IR drop visualization.

### Demo 3: MNIST Recognition (2 minutes)

Draw a digit and watch the network recognize it:

1. Select **Module 3: MNIST**
2. Draw a digit (0-9) in the canvas
3. Click "Classify"
4. See confidence scores for each digit

**What you'll see:** Neural network inference in action.

---

## 🛠️ Command-Line Usage

### Basic Commands

```bash
# Run GUI (default)
./fecim-lattice-tools

# Enable logging
./fecim-lattice-tools --logger --verbosity debug

# List available materials
./fecim-lattice-tools --list-materials

# Run calibration (headless)
./fecim-lattice-tools --calibrate --material fecim_hzo
```

Full reference: [cli-reference.md](cli-reference.md)

---

## 🎥 Video Tutorials

Visual guides for each module (see [demo-videos/](demo-videos/)):

| Video | Duration | Topic |
|-------|----------|-------|
| `01-hysteresis-loop.mp4` | 2 min | P-E curves and materials |
| `02-crossbar-mvm.mp4` | 3 min | Matrix multiplication |
| `03-mnist-inference.mp4` | 2 min | Neural network demo |
| `04-circuits-dac-adc.mp4` | 3 min | Peripheral circuits |
| `05-full-workflow.mp4` | 5 min | End-to-end walkthrough |

---

## 🐛 Troubleshooting

### Build Errors

**Error:** `gcc: command not found`
→ **Fix:** Install GCC (see [installation.md](installation.md))

**Error:** `cannot find package "fyne.io/fyne/v2"`
→ **Fix:** Run `go mod download`

**Error:** `undefined: GL_VERSION`
→ **Fix:** Install OpenGL dev packages (see [installation.md](installation.md))

### Runtime Issues

**Problem:** Black screen on launch
→ **Fix:** Try `FYNE_NO_GL=1 ./fecim-lattice-tools`

**Problem:** Window resize loop (Wayland/Sway)
→ **Fix:** Use `GDK_BACKEND=x11 ./fecim-lattice-tools`

**Problem:** Font rendering issues
→ **Fix:** Set `export FYNE_SCALE=1.0`

Full troubleshooting: [runbook.md#common-issues](runbook.md#common-issues)

---

## 💡 Learning Paths

### Path A: Complete Beginner (2 hours)

```
1. Install (this guide) ..................... 10 min
2. Watch demo videos ....................... 15 min
3. Read ELI5 overview ...................... 20 min
4. Try Module 1 ............................ 15 min
5. Try Module 2 ............................ 20 min
6. Try Module 3 ............................ 15 min
7. Explore remaining modules ............... 25 min
```

### Path B: Quick Developer Start (30 minutes)

```
1. Install (this guide) ..................... 10 min
2. Skim runbook ............................. 5 min
3. Read API reference ....................... 10 min
4. Run test suite ........................... 5 min
```

### Path C: Researcher Validation (1 hour)

```
1. Install (this guide) ..................... 10 min
2. Read honesty audit ....................... 10 min
3. Review physics validation ................ 20 min
4. Browse research papers ................... 20 min
```

---

## 📦 What's Included

### Interactive Modules

- **Module 1:** Hysteresis & Materials (8 materials, P-E loops)
- **Module 2:** Crossbar Arrays (MVM, IR drop, sneak paths)
- **Module 3:** MNIST Neural Network (handwriting recognition)
- **Module 4:** Peripheral Circuits (DAC, ADC, TIA, charge pump)
- **Module 5:** Technology Comparison (CPU vs GPU vs FeCIM)
- **Module 6:** EDA Tools (chip design workflow)
- **Module 7:** Documentation Viewer (glossary, search)

### Simulation Features

✅ Physics-based hysteresis (Preisach + Landau-Khalatnikov)
✅ Non-ideal crossbar effects (IR drop, sneak paths, drift)
✅ Dual-mode MNIST (floating-point vs CIM comparison)
✅ Peripheral circuit models (4-bit DAC/ADC baseline)
✅ Energy and timing analysis
✅ Material property explorer
✅ Export to JSON/CSV/SPICE

---

## 🔗 Next Steps

### Explore the Modules

- **[Module 1: Hysteresis](../2-learn/module1-hysteresis/)** - Start here!
- **[Module 2: Crossbar](../2-learn/module2-crossbar/)** - See computation
- **[Module 3: MNIST](../2-learn/module3-mnist/)** - Try recognition
- **[Module 4: Circuits](../2-learn/module4-circuits/)** - Learn peripherals
- **[Module 5: Comparison](../2-learn/module5-comparison/)** - Compare tech
- **[Module 6: EDA](../2-learn/module6-eda/)** - Design chips

### Dig Deeper

- **[Learn Section](../2-learn/)** - Educational content
- **[Develop Section](../3-develop/)** - API and architecture
- **[Research Section](../4-research/)** - Papers and validation

---

## 🤝 Getting Help

### Documentation

- **[GLOSSARY](../GLOSSARY.md)** - Technical terms explained
- **[FAQ](runbook.md#common-issues)** - Common questions
- **[API Docs](../3-develop/api-reference.md)** - Package reference

### Community

- **Issues:** Report bugs or request features
- **Discussions:** Ask questions, share ideas
- **Contributing:** See [CONTRIBUTING.md](../../CONTRIBUTING.md)

---

## ✅ Installation Checklist

Use this to verify your setup:

- [ ] Go 1.24+ installed (`go version`)
- [ ] GCC installed (`gcc --version`)
- [ ] Repository cloned
- [ ] Dependencies downloaded (`go mod download`)
- [ ] Binary built successfully
- [ ] GUI launches without errors
- [ ] Module 1 demo runs
- [ ] Module 2 demo runs
- [ ] Module 3 demo runs

If all checked, you're ready to explore! 🎉

---

## 📝 Build Options

### Standard Build
```bash
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
```

### Optimized Release Build
```bash
go build -ldflags="-s -w" -o fecim-lattice-tools ./cmd/fecim-lattice-tools
```

### Cross-Compilation
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o fecim-linux ./cmd/fecim-lattice-tools

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o fecim-mac ./cmd/fecim-lattice-tools

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o fecim-mac-arm ./cmd/fecim-lattice-tools

# Windows
GOOS=windows GOARCH=amd64 go build -o fecim.exe ./cmd/fecim-lattice-tools
```

---

## 🎯 Quick Links

**Essential:**
- [Installation Guide](installation.md)
- [Operations Runbook](runbook.md)
- [CLI Reference](cli-reference.md)

**Learning:**
- [ELI5 Overview](../2-learn/eli5-overview.md)
- [Module Documentation](../2-learn/)

**Development:**
- [API Reference](../3-develop/api-reference.md)
- [Architecture](../3-develop/architecture/)

**Research:**
- [Honesty Audit](../4-research/honesty-audit.md)
- [Physics Validation](../4-research/physics-validation.md)

---

**Last Updated:** 2026-02-16
**Status:** All modules functional and tested
**Support:** See [runbook.md](runbook.md) for troubleshooting
