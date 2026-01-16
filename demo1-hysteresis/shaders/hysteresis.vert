#version 450

// Vertex shader for hysteresis curve line rendering
// Supports line strip with varying alpha for trail effect

layout(location = 0) in vec2 inPosition;
layout(location = 1) in vec4 inColor;

layout(location = 0) out vec4 fragColor;
layout(location = 1) out vec2 lineCoord;

layout(push_constant) uniform PushConstants {
    float lineWidth;
    float aspectRatio;
    float time;
    float padding;
} pc;

void main() {
    gl_Position = vec4(inPosition, 0.0, 1.0);
    fragColor = inColor;
    lineCoord = inPosition;
}
