#!/bin/bash
# Compile GLSL shaders to SPIR-V for Vulkan
#
# Supports glslc (Vulkan SDK) or glslangValidator as a fallback.
# Pre-compiled .spv files are checked in so this script only needs to be
# re-run when shader sources change.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Detect compiler.
COMPILER=""
if command -v glslc &> /dev/null; then
    COMPILER="glslc"
elif command -v glslangValidator &> /dev/null; then
    COMPILER="glslangValidator"
else
    echo "Error: no GLSL compiler found. Install Vulkan SDK (glslc) or glslangValidator."
    echo "  Ubuntu: sudo apt install glslc          # Vulkan SDK"
    echo "          sudo apt install glslang-tools   # glslangValidator"
    exit 1
fi

echo "Using compiler: $COMPILER"

compile_shader() {
    local src="$1"
    local dst="${src}.spv"
    if [ ! -f "$src" ]; then
        return
    fi
    if [ "$COMPILER" = "glslc" ]; then
        glslc "$src" -o "$dst"
    else
        glslangValidator -V "$src" -o "$dst"
    fi
    echo "  $src -> $dst"
}

echo "Compiling shaders..."

# Compute shaders
compile_shader preisach.comp
compile_shader heatmap.comp

# Cell shaders (module 1 lattice visualisation)
compile_shader cell.vert
compile_shader cell.frag

# Simple passthrough shaders
compile_shader simple.vert
compile_shader simple.frag

# Hysteresis curve shaders
compile_shader hysteresis.vert
compile_shader hysteresis.frag

# Heatmap shaders (L09 GPU crossbar rendering)
compile_shader heatmap.vert
compile_shader heatmap.frag

echo "Done."
echo ""
echo "Shader files ready for Vulkan:"
ls -la *.spv 2>/dev/null || echo "  No .spv files found (run this script after creating shaders)"
