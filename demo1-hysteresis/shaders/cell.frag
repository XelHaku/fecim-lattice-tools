#version 450

// Fragment shader for ferroelectric cell and P-E plot visualization

// Input from vertex shader
layout(location = 0) in vec4 fragColor;

// Output color
layout(location = 0) out vec4 outColor;

// Uniform buffer for visualization options
layout(std140, binding = 0) uniform VisualizationParams {
    float gamma;         // Gamma correction
    float brightness;    // Brightness adjustment
    float contrast;      // Contrast adjustment
    float saturation;    // Color saturation
} params;

// Apply gamma correction
vec3 gammaCorrect(vec3 color, float gamma) {
    return pow(color, vec3(1.0 / gamma));
}

// Adjust saturation
vec3 adjustSaturation(vec3 color, float saturation) {
    float luminance = dot(color, vec3(0.299, 0.587, 0.114));
    return mix(vec3(luminance), color, saturation);
}

void main() {
    vec3 color = fragColor.rgb;

    // Apply brightness and contrast
    color = (color - 0.5) * params.contrast + 0.5 + params.brightness;

    // Apply saturation
    color = adjustSaturation(color, params.saturation);

    // Apply gamma correction
    color = gammaCorrect(color, params.gamma);

    // Clamp and output
    outColor = vec4(clamp(color, 0.0, 1.0), fragColor.a);
}
