#!/bin/bash
# Launch the unified FeCIM Visualizer
# Usage: ./launch.sh [--verbosity LEVEL]
#   LEVEL: 0|off, 1|info, 2|debug, 3|trace
cd "$(dirname "$0")"
go build ./cmd/fecim-visualizer && ./fecim-visualizer "$@"
