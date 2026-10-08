#version 430

#include "lighting.frag"
#include "grading.frag"

out vec4 fragColor;

in vec3 fragPosition;

uniform sampler2D albedo_blend_tex;
uniform sampler2D displacement_tex;
uniform sampler2D nar_tex;
//uniform sampler2D imp_map; // result of the material generator running

layout(shared, binding = 0) uniform TerrainBlock {
    float grading_group_id;
    float grading_group_id_masked_details;
    float grading_group_id_secondary_color;
    float material_index;
    float material_wetness;
    float subsurface_diff;
    float subsurface_int;
    float subsurface_wrap;
};

void main() {
    vec2 uv = fragPosition.xz * 0.1875;
    vec3 color = sRGBToIntensity(texture(albedo_blend_tex, uv).xyz);
    vec2 gradingParams = texture(displacement_tex, uv).zw;
    vec4 nar = texture(nar_tex, uv);
    vec3 normal = normalize(vec3(nar.xy * 2 - 1, reconstructNormalZ(nar.xy * 2 - 1)));
    //normal.x = -normal.x;
    float metallic = clamp((1.0 - material_wetness) * clamp(0.5 * -2 + 1, 0.0, 1.0) + material_wetness, 0.0, 1.0);
    metallic = metallic * 0.955 + 0.045;

    vec3 graded_color = color;
    if(grading_group_id > 0) {
        graded_color = gradeColor(color, gradingMatrix(int(floor(grading_group_id + 0.5))));
    }
    graded_color = clamp(graded_color, 0.0, 1.0);

    vec3 graded_secondary = color;
    if(grading_group_id_secondary_color > 0) {
        graded_secondary = gradeColor(color, gradingMatrix(int(floor(grading_group_id_secondary_color + 0.5))));
    }
    graded_secondary = clamp(graded_secondary, 0.0, 1.0);

    vec3 graded_masked_details = color;
    if(grading_group_id_masked_details > 0) {
        graded_masked_details = gradeColor(color, gradingMatrix(int(floor(grading_group_id_masked_details + 0.5))));
    }
    graded_masked_details = clamp(graded_masked_details, 0.0, 1.0);

    graded_color = gradingParams.x * (graded_secondary - graded_color) + graded_color;
    color = sRGBFromIntensity(gradingParams.y * (graded_masked_details - graded_color) + graded_color);

    fragColor = calculateLighting(vec4(color, 1.0), nar.w, metallic, nar.z, normal);
}