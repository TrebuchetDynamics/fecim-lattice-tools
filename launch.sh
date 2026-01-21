#!/bin/bash
# Launch the FeCIM Demo Suite

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Build all demos
echo "Building all demos..."
./scripts/build-all.sh

# Run launcher
./launcher
