#!/bin/bash
# Build all FeCIM demo binaries

set -e

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Building FeCIM Demo Suite..."
echo "Project root: $PROJECT_ROOT"
echo ""

cd "$PROJECT_ROOT"

# Build Demo 1: Hysteresis
echo "[1/6] Building Demo 1: Hysteresis..."
go build -o demo1-hysteresis/hysteresis ./demo1-hysteresis/cmd/hysteresis
echo "  -> demo1-hysteresis/hysteresis"

# Build Demo 2: Crossbar
echo "[2/6] Building Demo 2: Crossbar MVM..."
go build -o demo2-crossbar/crossbar-gui ./demo2-crossbar/cmd/crossbar-gui
echo "  -> demo2-crossbar/crossbar-gui"

# Build Demo 3: MNIST
echo "[3/6] Building Demo 3: MNIST..."
go build -o demo3-mnist/mnist-gui ./demo3-mnist/cmd/mnist-gui
echo "  -> demo3-mnist/mnist-gui"

# Build Demo 4: Circuits (if exists)
if [ -d "demo4-circuits/cmd/circuits-gui" ]; then
    echo "[4/6] Building Demo 4: Circuits..."
    go build -o demo4-circuits/circuits-gui ./demo4-circuits/cmd/circuits-gui
    echo "  -> demo4-circuits/circuits-gui"
else
    echo "[4/6] Skipping Demo 4: Circuits (not ready)"
fi

# Build Demo 8: Comparison (if exists)
if [ -d "demo8-comparison/cmd/comparison-gui" ]; then
    echo "[5/6] Building Demo 8: Comparison..."
    go build -o demo8-comparison/comparison-gui ./demo8-comparison/cmd/comparison-gui
    echo "  -> demo8-comparison/comparison-gui"
else
    echo "[5/6] Skipping Demo 8: Comparison (not ready)"
fi

# Build Launcher
echo "[6/6] Building Launcher..."
go build -o launcher ./cmd/launcher
echo "  -> launcher"

echo ""
echo "Build complete! Run ./launcher to start the demo suite."
