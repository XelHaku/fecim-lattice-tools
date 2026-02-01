#!/bin/bash
# Launch the unified FeCIM Visualizer
# Usage: ./launch.sh [--logger] [--verbosity LEVEL]
#   --logger: Enable file logging (logs to logs/ directory)
#   --verbosity LEVEL: Set logging verbosity (only used with --logger)
#     LEVEL: 0|off, 1|info, 2|debug, 3|trace
cd "$(dirname "$0")"
rm -f fecim-lattice-tools
echo "Building fecim-lattice-tools..."
if go build -v ./cmd/fecim-lattice-tools 2>&1; then
    echo "Build successful, launching..."
    ./fecim-lattice-tools "$@"
else
    echo "Build failed!"
    exit 1
fi
