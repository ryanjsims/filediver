#version 430 core

layout (points) in;
layout (line_strip, max_vertices = 6) out;

in vec4 normalEndPosition[];
in vec4 tangentEndPosition[];
in vec4 bitangentEndPosition[];

out vec4 lineColor;

layout(shared) uniform FilediverBlock {
    mat4 mvp; // projection*view*model
    mat4 model;
    mat3 normalMat; // normal matrix = transpose(inverse(model))
    vec3 viewPosition;
    vec4 color;
    float opacityThreshold;
    float len;
    bool shouldReconstructNormalZ;
    bool showTangentBitangent;
    bool udimShown[64];
};

void drawLine(vec4 endPosition) {
    gl_Position = gl_in[0].gl_Position;
    EmitVertex();
    gl_Position = endPosition;
    EmitVertex();
    EndPrimitive();
}

void main() {
    lineColor = vec4(0, 0, 1, 1);
    drawLine(normalEndPosition[0]);
    if (showTangentBitangent) {
        lineColor = vec4(1, 0, 0, 1);
        drawLine(tangentEndPosition[0]);
        lineColor = vec4(0, 1, 0, 1);
        drawLine(bitangentEndPosition[0]);
    }
}