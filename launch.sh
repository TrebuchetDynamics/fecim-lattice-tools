#!/bin/bash
# Launch the FeCIM Demo Suite

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Build launcher if it doesn't exist
if [ ! -f "launcher" ]; then
    echo "Building launcher..."
    go build -o launcher ./cmd/launcher
fi

# Run launcher
./launcher
