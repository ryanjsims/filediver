#version 430 core

layout(location = 0) out vec4 outColor;

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

void main() {
    outColor = color;
}