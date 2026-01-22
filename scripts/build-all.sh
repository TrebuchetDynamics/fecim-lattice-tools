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
echo "[1/8] Building Demo 1: Hysteresis..."
go build -o module1-hysteresis/hysteresis ./module1-hysteresis/cmd/hysteresis
echo "  -> module1-hysteresis/hysteresis"

# Build Demo 2: Crossbar
echo "[2/8] Building Demo 2: Crossbar MVM..."
go build -o module2-crossbar/crossbar-gui ./module2-crossbar/cmd/crossbar-gui
echo "  -> module2-crossbar/crossbar-gui"

# Build Demo 3: MNIST
echo "[3/8] Building Demo 3: MNIST..."
go build -o module3-mnist/mnist-gui ./module3-mnist/cmd/mnist-gui
echo "  -> module3-mnist/mnist-gui"

# Build Demo 4: Circuits
echo "[4/8] Building Demo 4: Circuits..."
go build -o module4-circuits/circuits-gui ./module4-circuits/cmd/circuits-gui
echo "  -> module4-circuits/circuits-gui"

# Build Demo 6: Multilayer 3D Stack
echo "[5/8] Building Demo 6: 3D Stack..."
go build -o demo6-multilayer/multilayer-gui ./demo6-multilayer/cmd/multilayer-gui
echo "  -> demo6-multilayer/multilayer-gui"

# Build Demo 7: Non-Idealities
echo "[6/8] Building Demo 7: Non-Idealities..."
go build -o demo7-nonidealities/nonidealities-gui ./demo7-nonidealities/cmd/nonidealities-gui
echo "  -> demo7-nonidealities/nonidealities-gui"

# Build Demo 8: Comparison
echo "[7/8] Building Demo 8: Comparison..."
go build -o module5-comparison/comparison-gui ./module5-comparison/cmd/comparison-gui
echo "  -> module5-comparison/comparison-gui"

# Build Launcher
echo "[8/8] Building Launcher..."
go build -o launcher ./cmd/launcher
echo "  -> launcher"

echo ""
echo "Build complete! Run ./launcher to start the demo suite."
echo "7/8 demos ready (Demo 5: Thermal coming soon)"
