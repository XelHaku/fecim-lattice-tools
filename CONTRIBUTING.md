# Contributing to FeCIM Macro Compiler

Thank you for your interest in contributing to the FeCIM Visualizer project.

## Quick Start for Contributors

1. **Fork** the repository
2. **Clone** your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/fecim-lattice-tools.git
   cd fecim-lattice-tools
   ```
3. **Install dependencies:**
   ```bash
   sudo apt-get install -y gcc libgl1-mesa-dev xorg-dev  # Linux
   go mod download
   ```
4. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature
   ```
5. **Make changes and test:**
   ```bash
   go test ./...
   go build ./...
   ```
6. **Commit and push:**
   ```bash
   git commit -m "feat: add feature"
   git push origin feature/your-feature
   ```
7. **Open a Pull Request** on GitHub

## Calibration update policy (drift-sensitive)

Calibration JSON files under:

- `cmd/fecim-lattice-tools/data/calibrations/*.json`

are treated as **drift-sensitive**.

### CI guard: no uncommitted calibration drift

Use:

```bash
scripts/calib-guard.sh
```

This script fails when any calibration JSON file is changed but not committed
(staged, unstaged, or untracked). It is intended for CI jobs that should leave
these files untouched.

### Intentional calibration updates (required evidence)

If calibration JSON changes are intentional, include evidence links in the commit
message/body (bench logs, lab notes, issue/PR discussion, etc.).

Recommended trailer format:

```text
Calibration-Evidence: https://...
```

### Optional pre-commit warning hook (template)

Install the provided warning-only hook:

```bash
git config core.hooksPath .githooks
chmod +x .githooks/pre-commit
```

The hook warns when staged calibration JSON files are detected, but does **not**
block commits.

## Code Standards

### Go Code
- Run `gofmt` before committing
- Add tests for new features (maintain >80% coverage)
- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use meaningful variable and function names

### Documentation
- Update documentation for user-facing changes
- Add godoc comments to exported functions
- Keep README files up to date

### Testing
```bash
# Run all tests
go test ./...

# Run with verbose output
go test ./... -v

# Run specific module tests
go test ./module2-crossbar/pkg/crossbar -v
```

## Commit Message Format

```
<type>: <subject>

<body>

Co-Authored-By: Your Name <email@example.com>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `test`: Adding tests
- `refactor`: Code refactoring
- `style`: Formatting, missing semicolons, etc.
- `chore`: Maintenance tasks

**Examples:**
```
feat: add IR drop visualization to crossbar module

Implements heatmap overlay showing voltage distribution
across the crossbar array during MVM operations.
```

```
fix: correct quantization formula for negative weights

The symmetric quantization was not properly handling
weights at the boundary conditions.
```

## Project Structure

```
fecim-lattice-tools/
├── cmd/fecim-lattice-tools/     # Unified GUI entry point
├── module1-hysteresis/       # P-E curve visualization
├── module2-crossbar/         # Crossbar MVM + non-idealities
├── module3-mnist/            # MNIST inference demo
├── module4-circuits/         # Peripheral circuits
├── module5-comparison/       # Technology comparison
├── module6-eda/              # EDA tooling (main development)
├── shared/                   # Shared packages
└── docs/                     # Documentation
```

## Module 6 Development (Priority)

Module 6 (FeCIM Design Suite) is a **universal chip design tool** supporting:
- **Storage chips** (NAND replacement) — no AI involved
- **Memory chips** (DRAM replacement) — no AI involved
- **Compute chips** (AI accelerator) — weights optional

```
module6-eda/
├── pkg/compiler/     # Array design generation (all three modes)
├── pkg/export/       # Verilog, DEF, SPICE, GDS export
├── pkg/gui/          # Fyne GUI tabs
└── examples/         # Working examples
```

**Key areas needing contribution:**
- [ ] Custom FeCIM cell design in Magic VLSI
- [ ] Liberty timing model generation
- [ ] ngspice simulation integration
- [ ] Design space explorer (Tab 4)
- [ ] Storage mode: Data retention optimization
- [ ] Memory mode: Speed/bandwidth optimization

## Getting Help

- **Questions:** Open an issue with `[Question]` prefix
- **Bugs:** Open an issue with `[Bug]` prefix and reproduction steps
- **Features:** Open an issue with `[Feature]` prefix for discussion first

## Code of Conduct

- Be respectful and constructive
- Focus on the technical merits of contributions
- Help newcomers learn

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
