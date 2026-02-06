#!/bin/bash
# Screenshot automation at full HD resolution
set -e

DISPLAY_NUM=98
SCREENSHOT_DIR="screenshots/ui-review-hd"
APP_BIN="./fecim-lattice-tools"
WINDOW_W=1400
WINDOW_H=900

MODULES=("home" "hysteresis" "crossbar" "mnist" "circuits" "comparison" "eda" "docs")

cleanup() {
    pkill -f "fecim-lattice-tools" 2>/dev/null || true
    if [ -n "$XVFB_PID" ]; then
        kill "$XVFB_PID" 2>/dev/null || true
    fi
}
trap cleanup EXIT

rm -rf "$SCREENSHOT_DIR"
mkdir -p "$SCREENSHOT_DIR"

echo "Starting Xvfb on :${DISPLAY_NUM} at ${WINDOW_W}x${WINDOW_H}..."
Xvfb ":${DISPLAY_NUM}" -screen 0 "${WINDOW_W}x${WINDOW_H}x24" +extension RANDR &
XVFB_PID=$!
sleep 1

if ! kill -0 "$XVFB_PID" 2>/dev/null; then
    echo "ERROR: Xvfb failed to start"
    exit 1
fi

export DISPLAY=":${DISPLAY_NUM}"

# Clear saved window size preferences so app uses default 1400x900
# Fyne stores prefs in XDG_CONFIG_HOME or ~/.config/fyne
PREFS_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/fyne/com.fecim.visualizer"

for module in "${MODULES[@]}"; do
    echo ""
    echo "=== Capturing module: $module ==="

    "$APP_BIN" --module "$module" &
    APP_PID=$!

    FOUND=0
    for i in $(seq 1 30); do
        WID=$(xdotool search --name "FeCIM Lattice Tools" 2>/dev/null | head -1) || true
        if [ -n "$WID" ]; then
            FOUND=1
            # Resize window to fill screen
            xdotool windowsize "$WID" "$WINDOW_W" "$WINDOW_H" 2>/dev/null || true
            xdotool windowmove "$WID" 0 0 2>/dev/null || true
            echo "Window found: $WID (after ${i}x500ms)"
            break
        fi
        sleep 0.5
    done

    if [ "$FOUND" -eq 0 ]; then
        echo "WARNING: Window not found for module $module, skipping"
        kill "$APP_PID" 2>/dev/null || true
        wait "$APP_PID" 2>/dev/null || true
        continue
    fi

    # Extra render time
    sleep 4

    OUTFILE="${SCREENSHOT_DIR}/${module}.png"
    import -window "$WID" "$OUTFILE" 2>/dev/null || \
        xwd -id "$WID" | convert xwd:- "$OUTFILE" 2>/dev/null || \
        echo "WARNING: Failed to capture screenshot for $module"

    if [ -f "$OUTFILE" ]; then
        SIZE=$(identify -format "%wx%h" "$OUTFILE" 2>/dev/null || echo "unknown")
        echo "Screenshot saved: $OUTFILE ($SIZE)"
    fi

    kill "$APP_PID" 2>/dev/null || true
    wait "$APP_PID" 2>/dev/null || true
    sleep 1
done

echo ""
echo "=== All HD screenshots captured ==="
ls -la "$SCREENSHOT_DIR/"
