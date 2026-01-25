#!/bin/bash
# Launch the unified FeCIM Visualizer
# Usage: ./launch.sh [--verbosity LEVEL]
#   LEVEL: 0|off, 1|info, 2|debug, 3|trace
cd "$(dirname "$0")"
echo "Building fecim-visualizer..."
if go build -v ./cmd/fecim-visualizer 2>&1; then
    echo "Build successful, launching..."
    ./fecim-visualizer "$@"
else
    echo "Build failed!"
    exit 1
fi
