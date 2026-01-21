#!/bin/bash
# Launch the unified FeCIM Visualizer
cd "$(dirname "$0")"
go build ./cmd/fecim-visualizer && ./fecim-visualizer
