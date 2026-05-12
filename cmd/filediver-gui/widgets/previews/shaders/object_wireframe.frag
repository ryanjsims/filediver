#version 430 core

out vec4 fragColor;

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
    fragColor = color;
}