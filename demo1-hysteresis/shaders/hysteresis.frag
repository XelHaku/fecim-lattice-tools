#version 450

// Fragment shader for hysteresis curve rendering
// Implements smooth line rendering with anti-aliasing

layout(location = 0) in vec4 fragColor;
layout(location = 1) in vec2 lineCoord;

layout(location = 0) out vec4 outColor;

layout(push_constant) uniform PushConstants {
    float lineWidth;
    float aspectRatio;
    float time;
    float padding;
} pc;

// Simple anti-aliasing for lines
float lineAA(float dist, float width) {
    float edge = width * 0.5;
    return 1.0 - smoothstep(edge - 1.0, edge + 1.0, abs(dist));
}

void main() {
    // Apply color with alpha from vertex
    outColor = fragColor;

    // Add subtle glow effect for active curve
    if (fragColor.a > 0.9) {
        // Full alpha points get a subtle highlight
        outColor.rgb *= 1.1;
    }
}
