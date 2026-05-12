#version 430 core

layout(location = 0) in vec3 inPosition;
layout(location = 2) in vec2 inUV;

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

bool isShown() {
    int udim = int(inUV.x) | int(1-inUV.y)<<5;
    return udim < 64 && udimShown[udim];
}

void main() {
    if (!isShown()) return;
    gl_Position = mvp * vec4(inPosition, 1.0);
}