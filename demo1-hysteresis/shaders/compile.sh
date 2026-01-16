#!/bin/bash
# Compile GLSL shaders to SPIR-V for Vulkan

# Check for glslc
if ! command -v glslc &> /dev/null; then
    echo "Error: glslc not found. Install Vulkan SDK."
    echo "  Ubuntu: sudo apt install glslc"
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "Compiling shaders..."

# Compile compute shader
if [ -f "preisach.comp" ]; then
    glslc preisach.comp -o preisach.comp.spv
    echo "  preisach.comp -> preisach.comp.spv"
fi

# Compile vertex shader (when available)
if [ -f "cell.vert" ]; then
    glslc cell.vert -o cell.vert.spv
    echo "  cell.vert -> cell.vert.spv"
fi

# Compile fragment shader (when available)
if [ -f "cell.frag" ]; then
    glslc cell.frag -o cell.frag.spv
    echo "  cell.frag -> cell.frag.spv"
fi

# Compile hysteresis curve shaders
if [ -f "hysteresis.vert" ]; then
    glslc hysteresis.vert -o hysteresis.vert.spv
    echo "  hysteresis.vert -> hysteresis.vert.spv"
fi

if [ -f "hysteresis.frag" ]; then
    glslc hysteresis.frag -o hysteresis.frag.spv
    echo "  hysteresis.frag -> hysteresis.frag.spv"
fi

echo "Done."
echo ""
echo "Shader files ready for Vulkan:"
ls -la *.spv 2>/dev/null || echo "  No .spv files found (run this script after creating shaders)"
