#!/bin/bash
# Compile module4-circuits compute shaders to SPIR-V

if ! command -v glslc &> /dev/null; then
    echo "Error: glslc not found. Install Vulkan SDK."
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "Compiling module4-circuits compute shaders..."

for shader in *.comp; do
    if [ -f "$shader" ]; then
        glslc "$shader" -o "${shader}.spv"
        echo "  $shader -> ${shader}.spv"
    fi
done

echo "Done."
