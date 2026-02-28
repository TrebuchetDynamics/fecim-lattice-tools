#version 450

// Instanced crossbar heatmap vertex shader.
// Each instance is one crossbar cell. The vertex buffer contains a unit quad
// (4 vertices, drawn as triangle strip). Per-instance data comes from a
// storage buffer with one float per cell (row-major, normalised 0-1).
//
// Push constants carry grid dimensions and the pixel rectangle so the shader
// can position each cell without a full MVP matrix.

layout(location = 0) out float vNormValue;

// Push constants: grid size and viewport mapping.
layout(push_constant) uniform PushConstants {
    uint  rows;        // Number of rows in the crossbar array
    uint  cols;        // Number of columns in the crossbar array
    float originX;     // NDC X of top-left corner
    float originY;     // NDC Y of top-left corner
    float cellWidth;   // NDC width of one cell
    float cellHeight;  // NDC height of one cell
} pc;

// Per-cell conductance values (row-major, length = rows*cols).
layout(std430, binding = 0) readonly buffer CellData {
    float values[];
} cellData;

void main() {
    // gl_VertexIndex: 0..3 for the unit quad (triangle strip).
    //   0 = (0,0)  top-left
    //   1 = (1,0)  top-right
    //   2 = (0,1)  bottom-left
    //   3 = (1,1)  bottom-right
    float qx = float(gl_VertexIndex & 1);
    float qy = float((gl_VertexIndex >> 1) & 1);

    // gl_InstanceIndex encodes row*cols + col.
    uint row = gl_InstanceIndex / pc.cols;
    uint col = gl_InstanceIndex % pc.cols;

    // Cell corner in NDC.
    float x = pc.originX + float(col) * pc.cellWidth  + qx * pc.cellWidth;
    float y = pc.originY + float(row) * pc.cellHeight + qy * pc.cellHeight;

    gl_Position = vec4(x, y, 0.0, 1.0);

    // Pass normalised value to fragment shader for colormap lookup.
    uint idx = row * pc.cols + col;
    vNormValue = (idx < cellData.values.length()) ? cellData.values[idx] : 0.0;
}
