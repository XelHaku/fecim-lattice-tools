#version 450

// Vertex shader for ferroelectric cell and P-E plot visualization

// Vertex input
layout(location = 0) in vec2 inPosition;
layout(location = 1) in vec4 inColor;

// Output to fragment shader
layout(location = 0) out vec4 fragColor;

// Push constants for transform
layout(push_constant) uniform PushConstants {
    mat4 transform;  // Model-view-projection matrix
    float time;      // Animation time
} pc;

void main() {
    // Apply transformation (for future camera/zoom support)
    gl_Position = pc.transform * vec4(inPosition, 0.0, 1.0);

    // Pass color to fragment shader
    fragColor = inColor;
}
