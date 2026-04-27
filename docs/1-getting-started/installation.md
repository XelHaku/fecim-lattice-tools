# Installation

> **Note:** This file was previously located at `docs/INSTALLATION.md`. It has moved to `docs/1-getting-started/installation.md`.

This document lists required prerequisites and optional dependencies for specific features.

## Prerequisites

- **Go 1.25+** — https://go.dev/dl/
- **C compiler** (gcc/clang) for CGO
- **OpenGL libraries**

## Optional Dependencies

- **Docker** — For Module 6 EDA tools (OpenLane/OpenROAD/KLayout)
- **Graphviz** — For Yosys circuit schematic visualization
- **LaTeX + dvisvgm** — For regenerating equation SVG assets (Frankestein equation)

### Linux (Ubuntu/Debian)

```bash
sudo apt-get update
sudo apt-get install -y gcc libgl1-mesa-dev xorg-dev
# Optional: for Module 6 Yosys schematic visualization
sudo apt-get install -y graphviz
# Optional: run GUI/layout tests on a headless server
sudo apt-get install -y xvfb
```

### Linux (Fedora/RHEL)

```bash
sudo dnf install -y gcc mesa-libGL-devel libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel libXxf86vm-devel
```

### macOS

```bash
xcode-select --install  # Install command line tools
```

### Windows

1. Install MSYS2 (https://www.msys2.org/) or TDM-GCC (https://jmeubank.github.io/tdm-gcc/)
2. Ensure `gcc` is in your PATH
3. Run: `go build -o fecim-lattice-tools.exe ./cmd/fecim-lattice-tools`

## Equation SVG (Optional, Ubuntu/Debian)

```bash
sudo apt-get update
sudo apt-get install -y texlive-latex-base texlive-latex-recommended texlive-latex-extra texlive-fonts-recommended dvisvgm ghostscript
```

Regenerate the equation SVG after edits:

```bash
go run ./cmd/latex-svg -in shared/assets/equations/frankestein.tex -out shared/assets/equations/frankestein.svg
```
