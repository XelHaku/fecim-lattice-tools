#!/bin/bash
# Screenshot automation for FeCIM Lattice Tools
# Takes headless screenshots of each module using Xvfb

set -e

DISPLAY_NUM="${DISPLAY_NUM:-99}"

# Find a free display number if requested/default is occupied
if [ -e "/tmp/.X${DISPLAY_NUM}-lock" ]; then
    for candidate in 100 101 102 103 104 105; do
        if [ ! -e "/tmp/.X${candidate}-lock" ]; then
            DISPLAY_NUM="$candidate"
            break
        fi
    done
fi
SCREENSHOT_DIR="screenshots/ui-review"
APP_BIN="./bin/fecim-lattice-tools"
WINDOW_W=1400
WINDOW_H=900

# Modules to screenshot (matching --module flag values)
MODULES=("home" "hysteresis" "crossbar" "mnist" "circuits" "comparison" "eda" "docs")

# Cleanup function
cleanup() {
    # Kill any leftover app processes
    pkill -f "fecim-lattice-tools" 2>/dev/null || true
    # Kill Xvfb
    if [ -n "$XVFB_PID" ]; then
        kill "$XVFB_PID" 2>/dev/null || true
    fi
}
trap cleanup EXIT

# Create output directory
rm -rf "$SCREENSHOT_DIR"
mkdir -p "$SCREENSHOT_DIR"
mkdir -p "$(dirname "$APP_BIN")"

echo "Building gogpu/ui GUI binary ($APP_BIN)..."
CGO_ENABLED=0 go build -o "$APP_BIN" ./cmd/fecim-lattice-tools

CAPTURED=0

# Start Xvfb
echo "Starting Xvfb on :${DISPLAY_NUM}..."
Xvfb ":${DISPLAY_NUM}" -screen 0 "${WINDOW_W}x${WINDOW_H}x24" +extension RANDR &
XVFB_PID=$!
sleep 1

# Verify Xvfb is running
if ! kill -0 "$XVFB_PID" 2>/dev/null; then
    echo "ERROR: Xvfb failed to start"
    exit 1
fi
echo "Xvfb running (PID: $XVFB_PID)"

export DISPLAY=":${DISPLAY_NUM}"
# Force UTF-8 locale for stable headless runs.
export LANG="C.UTF-8"
export LC_ALL="C.UTF-8"
export LANGUAGE="C.UTF-8"

for module in "${MODULES[@]}"; do
    echo ""
    echo "=== Capturing module: $module ==="

    # Launch app with specific module
    "$APP_BIN" --module "$module" &
    APP_PID=$!

    # Wait for app window to appear (up to 15 seconds)
    echo "Waiting for window to appear..."
    FOUND=0
    for i in $(seq 1 30); do
        WID=$(xdotool search --name "FeCIM Lattice Tools" 2>/dev/null | head -1) || true
        if [ -n "$WID" ]; then
            FOUND=1
            echo "Window found: $WID (after ${i}x500ms)"
            break
        fi
        sleep 0.5
    done

    # In headless/Xvfb environments xdotool may fail to resolve a top-level window
    # even when Fyne has rendered to the root display. Fall back to root capture.
    if [ "$FOUND" -eq 0 ]; then
        echo "WARNING: Window not found for module $module, falling back to root window capture"
        WID=$(xwininfo -root 2>/dev/null | awk '/Window id:/ {print $4}') || true
    fi

    # Give it extra time to render content (modules have async Start())
    sleep 3

    # Take screenshot of the window (or root fallback)
    OUTFILE="${SCREENSHOT_DIR}/${module}.png"
    if [ -n "$WID" ]; then
        import -window "$WID" "$OUTFILE" 2>/dev/null || \
            xwd -id "$WID" | convert xwd:- "$OUTFILE" 2>/dev/null || \
            scrot "$OUTFILE" 2>/dev/null || \
            echo "WARNING: Failed to capture screenshot for $module"
    else
        scrot "$OUTFILE" 2>/dev/null || echo "WARNING: Failed to capture screenshot for $module (no window id available)"
    fi

    if [ -f "$OUTFILE" ]; then
        SIZE=$(identify -format "%wx%h" "$OUTFILE" 2>/dev/null || echo "unknown")
        echo "Screenshot saved: $OUTFILE ($SIZE)"
        CAPTURED=$((CAPTURED + 1))
    fi

    # Kill the app
    kill "$APP_PID" 2>/dev/null || true
    wait "$APP_PID" 2>/dev/null || true

    # Brief pause between launches
    sleep 1
done

echo ""
echo "=== All screenshots captured ==="
ls -la "$SCREENSHOT_DIR/"
echo "Captured screenshots: $CAPTURED"

if [ "$CAPTURED" -eq 0 ]; then
    echo "ERROR: No screenshots were captured."
    exit 1
fi
