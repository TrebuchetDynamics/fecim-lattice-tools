#version 450

// Crossbar heatmap fragment shader with viridis-like colormap.
// Receives a normalised value [0,1] from the vertex shader and maps it to
// a perceptually uniform, colorblind-safe colour gradient.

layout(location = 0) in float vNormValue;

layout(location = 0) out vec4 outColor;

// Polynomial approximation of the matplotlib "viridis" colormap.
// Degree-4 fit to the 256-entry viridis table.
vec3 viridis(float t) {
    t = clamp(t, 0.0, 1.0);
    float r = 0.2777 - 0.0563*t + 2.4411*t*t - 5.9587*t*t*t + 4.3322*t*t*t*t;
    float g = 0.0046 + 1.5230*t - 1.4266*t*t + 1.7166*t*t*t - 0.8530*t*t*t*t;
    float b = 0.3292 + 1.1523*t - 3.5462*t*t + 5.4371*t*t*t - 3.3903*t*t*t*t;
    return clamp(vec3(r, g, b), 0.0, 1.0);
}

void main() {
    outColor = vec4(viridis(vNormValue), 1.0);
}
