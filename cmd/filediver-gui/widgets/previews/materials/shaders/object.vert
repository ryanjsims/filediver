#version 430 core

layout(location = 0) in vec3 inPosition;
layout(location = 1) in vec3 inNormal;
layout(location = 2) in vec4 inTangent;
layout(location = 3) in vec3 inBitangent;
layout(location = 4) in vec2 inUV;
layout(location = 5) in vec2 inUV1;
layout(location = 6) in vec2 inUV2;
layout(location = 7) in mat4 instModel;
layout(location = 11) in mat3 normalMat;
layout(location = 14) in ivec4 indices;
layout(location = 15) in vec4 weights;

#define MAX_BONES 256
layout(shared, binding = 11) uniform BoneMatricesBlock {
    uint animate;
    int debugBone;
    mat4 bones[MAX_BONES];
};

out VertexOutput
{
    vec3 fragPosition;
    vec3 fragInstPosition;
    vec3 objPosition;
    vec2 fragUV0;
    vec2 fragUV1;
    vec2 fragUV2;
    vec3 fragTangentLightPosition; // tangent meaning in tangent space
    vec3 fragTangentViewPosition;
    vec3 fragTangentFragmentPosition;
    mat3 dbg_fragTBN;
    mat3 dbg_fragITBN;
    mat4 instanceModel;
    mat4 boneOut;
    vec4 weightsOut;
    vec4 indicesOut;
};

uniform mat4 projection;
uniform mat4 view;
uniform mat4 model;
uniform vec3 viewPosition;

const mat4 stingrayToGL = mat4(
    1, 0, 0, 0,
    0, 0, 1, 0,
    0, -1, 0, 0,
    0, 0, 0, 1
);

void main() {
    instanceModel = stingrayToGL * instModel;
    if (debugBone >= 0) {
        boneOut = bones[debugBone];
    } else {
        boneOut = mat4(0.0);
    }
    weightsOut = weights;
    indicesOut = vec4(indices) / 256.0;
    vec4 position;
    if (animate != 0) {
        position = bones[indices.x] * vec4(inPosition, 1.0) * weights.x
            + bones[indices.y] * vec4(inPosition, 1.0) * weights.y
            + bones[indices.z] * vec4(inPosition, 1.0) * weights.z
            + bones[indices.w] * vec4(inPosition, 1.0) * weights.w;
    } else {
        position = vec4(inPosition, 1.0);
    }
    gl_Position = projection * view * model * instModel * position;
    fragPosition = vec3(model * vec4(inPosition, 1.0));
    fragInstPosition = vec3(instanceModel * vec4(inPosition, 1.0));
    objPosition = inPosition;
    fragUV0 = inUV;
    fragUV1 = inUV1;
    fragUV2 = inUV2;

    {
        vec3 t = normalize(normalMat * inTangent.xyz);
        vec3 n = normalize(normalMat * inNormal);
        vec3 b = normalize(normalMat * inBitangent);
        //t = normalize(t - dot(t, n) * n);
        //vec3 b = cross(n, t);

        mat3 tbn = mat3(t, b, n);
        mat3 itbn = inverse(tbn);

        dbg_fragTBN = tbn;
        dbg_fragITBN = itbn;

        vec3 lightPosition = normalize(viewPosition) * 10000.0; // light behind camera

        fragTangentLightPosition    = itbn * lightPosition;
        fragTangentViewPosition     = itbn * viewPosition;
        fragTangentFragmentPosition = itbn * fragPosition;
    }
}
