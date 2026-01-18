#version 450

// Simple fragment shader for ferroelectric visualization
// Minimal version for initial rendering

layout(location = 0) in vec4 fragColor;

layout(location = 0) out vec4 outColor;

void main() {
    outColor = fragColor;
}
