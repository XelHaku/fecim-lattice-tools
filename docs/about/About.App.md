# About FeCIM Lattice Tools

## Overview

FeCIM Lattice Tools is an educational visualization suite for Ferroelectric Compute-in-Memory (FeCIM) concepts and models.

## Purpose

This application demonstrates model-based behavior of ferroelectric memory for learning and exploration. It serves as:

- **Educational tool** for understanding ferroelectric materials and compute-in-memory
- **Research platform** for exploring FeCIM model behavior
- **Visualization suite** for demonstrating tool capabilities

## The 7-Module Story

| Module | Name | Purpose |
|--------|------|---------|
| 1 | Hysteresis | How the memory cell works (P-E curves, Preisach model) |
| 2 | Crossbar | How we compute (matrix-vector multiplication, non-idealities) |
| 3 | MNIST | What we can build (neural network digit recognition) |
| 4 | Circuits | How it fits in a chip (DAC/ADC/TIA peripherals) |
| 5 | Comparison | Model-based comparison (simulation-only) |
| 6 | EDA | Educational EDA artifacts (OpenLane-style formats) |
| 7 | Docs | Built-in documentation browser |

## Technology Stack

- **Language:** Go
- **GUI Framework:** Fyne v2
- **Platform:** Cross-platform (Linux, macOS, Windows)

## Model Defaults

- Analog levels default to 30 (configurable).
- Physical parameters are model defaults defined in `config/physics.yaml`.

## Research Foundation

This project is inspired by published literature and conference talks on ferroelectric compute-in-memory. It is not affiliated with any specific lab.

## License

Open source software. See LICENSE file for details.

## Links

- Project repository (see README)
