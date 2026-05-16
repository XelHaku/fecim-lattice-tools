#!/bin/bash
# Launch the unified FeCIM Visualizer
# Usage: ./launch.sh [--logger [LEVEL]] [--verbosity LEVEL] [--clear]
#   --logger: Enable file logging (logs to logs/ directory)
#     Optional shorthand: --logger debug|info|trace|off
#   --verbosity LEVEL: Set logging verbosity (only used with --logger)
#     LEVEL: 0|off, 1|info, 2|debug, 3|trace
#   --clear: Delete logs/ and screenshots/ folders before running
cd "$(dirname "$0")"

# Check for --clear flag and remove it from args passed to the app
CLEAR_FLAG=false
ARGS=()
for arg in "$@"; do
    if [[ "$arg" == "--clear" ]]; then
        CLEAR_FLAG=true
    else
        ARGS+=("$arg")
    fi
done

# Clear logs and screenshots if requested
if $CLEAR_FLAG; then
    echo "Clearing logs/ and screenshots/ directories..."
    rm -rf logs/ screenshots/
fi
rm -f fecim-lattice-tools
echo "Building fecim-lattice-tools (gogpu/ui, zero-CGO)..."
if CGO_ENABLED=0 go build -v -o fecim-lattice-tools ./cmd/fecim-lattice-tools 2>&1; then
    echo "Build successful, launching..."
    ./fecim-lattice-tools "${ARGS[@]}"
else
    echo "Build failed!"
    exit 1
fi
