#version 430 core

#include "lighting.frag"
#include "triplaner.frag"
#include "grading.frag"

out vec4 fragColor;

in vec2 fragUV0;

uniform sampler2D base_normal_ao_dirt;
uniform sampler2D detail_mask_;
uniform sampler2D weathering_albedo;
uniform sampler2D weathering_nar;
uniform sampler2DArray triplanar_detail_albedo;
uniform sampler2DArray triplanar_detail_data;

layout(shared, binding = 0) uniform RockBlock {
    float albedo_intensity;
    float base_ao_intensity;
    float base_normal_intensity;
    vec3  bottom_color_tint;
    float bottom_dirt_value;
    float dirt_ao;
    float dirt_ao_sharpness;
    vec3  dirt_color;
    float dirt_detail_ao;
    float dirt_detail_sharpness;
    float dirt_detail_mask;
    float dirt_metallic;
    float dirt_roughness;
    float dirt_roughness_blend;
    float gradient_dirt_maxheight;
    float gradient_dirt_minheight;
    float grading_group_id;
    float grading_group_id_weathering;
    float grading_secondary_group_id;
    float heightmap_normals;
    float material_wetness;
    float tiler_ao_intensity;
    vec3  top_color_tint;
    float top_dirt_value;
    vec3  triplanar_detail_tiler;
    float triplanar_normal_intensity;
    float triplanar_vertex_normal_lerp;
    float up_fade_curve;
    float up_fade_distance;
    float weathering_ao_amount;
    float weathering_ao_cutoff;
    float weathering_coverage;
    float weathering_coverage_multiplier;
    float weathering_height;
    float weathering_height_curve;
    float weathering_normal_mult;
    float weathering_on_dirt_amount;
    float weathering_thickness;
    float weathering_tiling;
};

void main() {
    vec3 normal = normalize(dbg_fragTBN[2].xyz);
    normal = heightmap_normals * (vec3(0.0, 0.0, 1.0) - normal) + normal;
    float sqrlen = dot(normal, normal);
    normal = normal * inversesqrt(sqrlen);
    if (sqrlen == 0.0) {
        normal = vec3(0.0, 0.0, 1.0);
    }

    vec4 nad = texture(base_normal_ao_dirt, fragUV0);
    vec2 details = texture(detail_mask_, fragUV0).xy;

    vec3 nad_normal = vec3((nad.xy * 2.0 - 1.0) * vec2(-1.0, 1.0), 0.0);
    nad_normal.z = reconstructNormalZ(nad_normal.xy);

    vec2 nxy_sq = nad_normal.xy * nad_normal.xy;
    nad_normal.z = max(max(nxy_sq.x, nxy_sq.y) / 16384.0, nad_normal.z);

    
}
