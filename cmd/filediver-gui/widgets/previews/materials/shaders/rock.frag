#version 430 core

#include "lighting.frag"
#include "triplanar.frag"
#include "grading.frag"

out vec4 fragColor;

in vec3 objPosition;
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

layout(shared, binding = 1) uniform DebugBlock {
    uint debug;
};

void main() {
    vec3 normal = normalize(dbg_fragTBN[2].xyz);
    vec3 mixed_vertex_normal = heightmap_normals * (vec3(0.0, 0.0, 1.0) - normal) + normal;
    float sqrlen = dot(mixed_vertex_normal, mixed_vertex_normal);
    mixed_vertex_normal = mixed_vertex_normal * inversesqrt(sqrlen);
    if (sqrlen == 0.0) {
        mixed_vertex_normal = vec3(0.0, 0.0, 1.0);
    }

    vec3 face_tangent = dFdx(objPosition);
    vec3 face_bitangent = dFdy(objPosition);
    vec3 face_mixed_tangent = face_tangent - dot(face_tangent, mixed_vertex_normal) * mixed_vertex_normal;
    vec3 face_mixed_bitangent = face_bitangent - dot(face_bitangent, mixed_vertex_normal) * mixed_vertex_normal;
    vec3 face_crossed_bitangent = cross(face_tangent, mixed_vertex_normal);
    float face_handedness = dot(face_bitangent, face_crossed_bitangent) < 0.0 ? -1.0 : 1.0;

    vec2 uv_dx = dFdx(fragUV0);
    vec2 uv_dy = dFdy(fragUV0);

    float uv_align = dot(uv_dx, vec2(uv_dy.y, -uv_dy.x));
    float uv_handedness = uv_align < 0.0 ? -1.0 : 1.0;
    vec3 uv_face_tangent = (uv_handedness * uv_dy.y) * face_mixed_tangent - ((uv_handedness * uv_dx.y) * face_mixed_bitangent);

    uv_face_tangent = abs(uv_align) > 0.0 ? normalize(uv_face_tangent) : uv_face_tangent;
    float handedness = face_handedness * uv_handedness;
    vec3 uv_face_bitangent = handedness * cross(uv_face_tangent, mixed_vertex_normal);

    vec4 nad = texture(base_normal_ao_dirt, fragUV0);
    vec2 details = texture(detail_mask_, fragUV0).xy;

    vec3 nad_normal = vec3((nad.xy * 2.0 - 1.0) * vec2(-1.0, 1.0), 0.0);

    vec2 nxy_sq = nad_normal.xy * nad_normal.xy;

    nad_normal.z = max(max(nxy_sq.x, nxy_sq.y) / 16384.0, 1.0 - nad_normal.x * nad_normal.x - nad_normal.y * nad_normal.y);
    nad_normal.z = inversesqrt(nad_normal.z);
    nad_normal.xy = nad_normal.xy * -nad_normal.z;
    nad_normal.xyz = base_normal_intensity * (nad_normal.x * uv_face_tangent.xyz + nad_normal.y * uv_face_bitangent.xyz);

    vec3 combined_normal = mixed_vertex_normal + nad_normal;
    combined_normal = triplanar_vertex_normal_lerp * (normal - combined_normal) + combined_normal;
    sqrlen = dot(combined_normal, combined_normal);
    combined_normal = normalize(combined_normal);
    if (sqrlen <= 0.0) {
        combined_normal = vec3(0.0, 0.0, 1.0);
    }

    vec2 uvs = getCustomTriplanarUVs(combined_normal);

    vec4 detail_albedo_x = texture(triplanar_detail_albedo, vec3(uvs * triplanar_detail_tiler.x, 0.0));
    vec4 detail_albedo_y = texture(triplanar_detail_albedo, vec3(uvs * triplanar_detail_tiler.y, 1.0));
    vec4 detail_albedo_z = texture(triplanar_detail_albedo, vec3(uvs * triplanar_detail_tiler.z, 2.0));

    float detail_mix_y = dot(details.xx, details.xx) - 3.0 * details.x + 2.0;
    detail_mix_y = dot(vec2(detail_mix_y), details.xx);
    detail_mix_y = clamp(((2.0 * detail_mix_y) + detail_albedo_y.w - detail_albedo_x.w - 2.0) * 1.5 + 0.5, 0.0, 1.0);

    vec3 combined_detail_albedo = detail_mix_y * (detail_albedo_y.xyz - detail_albedo_x.xyz) + detail_albedo_x.xyz;
    float combined_albedo_blend = details.x * (detail_albedo_y.w - detail_albedo_x.w) + detail_albedo_x.w;

    float detail_mix_z = dot(details.yy, details.yy) - 3.0 * details.y + 2.0;
    detail_mix_z = dot(vec2(detail_mix_z), details.yy);
    combined_albedo_blend = (combined_albedo_blend + 1.0) - detail_mix_z;
    detail_mix_z = detail_mix_z + detail_albedo_z.w - 1.0;
    detail_mix_z = clamp((detail_mix_z - combined_albedo_blend) * 1.5 + 0.5, 0.0, 1.0);
    combined_detail_albedo = detail_mix_z * (detail_albedo_z.xyz - combined_detail_albedo) + combined_detail_albedo;

    vec4 detail_data_x = texture(triplanar_detail_data, vec3(uvs * triplanar_detail_tiler.x, 0.0));
    vec4 detail_data_y = texture(triplanar_detail_data, vec3(uvs * triplanar_detail_tiler.y, 1.0));
    vec4 detail_data_z = texture(triplanar_detail_data, vec3(uvs * triplanar_detail_tiler.z, 2.0));

    vec4 combined_detail_data = detail_mix_y * (detail_data_y - detail_data_x) + detail_data_x;
    combined_detail_data = detail_mix_z * (detail_data_z - combined_detail_data) + combined_detail_data;

    float inv_ao = 1.0 - combined_detail_data.z;
    float half_ao = combined_detail_data.z * 0.5;

    vec3 detail_normal = vec3((combined_detail_data.xy * vec2(-1.0, 1.0) + vec2(1.0, 0.0)) * 2.0 - 1.0, 0.0);
    vec2 dn_sq = detail_normal.xy * detail_normal.xy;
    float dn_z = inversesqrt(max(max(dn_sq.x, dn_sq.y) / 16384.0, 1.0 - detail_normal.x * detail_normal.x - detail_normal.y * detail_normal.y));

    detail_normal.xz = detail_normal.xy * -dn_z;
    detail_normal.y = 0.0;

    detail_normal = getTriplanarCustomNormal(combined_normal, detail_normal);
    combined_normal = triplanar_normal_intensity * (detail_normal - dot(detail_normal, combined_normal) * combined_normal) + nad_normal;

    vec3 base_albedo = combined_detail_albedo * albedo_intensity;

    float dirt_height = clamp(((objPosition.y + fd_world_offset.y) - gradient_dirt_maxheight) / (gradient_dirt_minheight - gradient_dirt_maxheight), 0.0, 1.0);
    dirt_height = (3.0 - dirt_height * 2.0) * dirt_height * dirt_height;

    float dirt_value = dirt_height * (bottom_dirt_value - top_dirt_value) + top_dirt_value;

    float inv_dirt_detail_sharpness = 1.0 - dirt_detail_sharpness;
    half_ao = min(pow(abs(half_ao), 4.0), 1.0);

    float inv_dirt_ao_sharpness = 1.0 - dirt_ao_sharpness;
    float inv_dirt_ao = 1.0 - dirt_ao;
    float inv_dirt_detail_ao = 1.0 - dirt_detail_ao;

    float min_value = min(dirt_value, 0.0);
    float dirt_combined_inv_ao = clamp(clamp(clamp(inv_ao + min_value, 0.0, 1.0) - inv_dirt_detail_ao, 0.0, 1.0) / inv_dirt_detail_sharpness, 0.0, 1.0);
    dirt_combined_inv_ao = (3.0 - dirt_combined_inv_ao * 2.0) * dirt_combined_inv_ao * dirt_combined_inv_ao;

    float dirt_combined_ao = max(clamp(1.0 + (min_value - nad.z) - half_ao, 0.0, 1.0) - clamp(inv_dirt_ao, 0.0, 1.0), 0.0);
    dirt_combined_ao = clamp(dirt_combined_ao / inv_dirt_ao_sharpness, 0.0, 1.0);
    dirt_combined_ao = (3.0 - dirt_combined_ao * 2.0) * dirt_combined_ao * dirt_combined_ao;
    dirt_combined_ao = clamp(inv_dirt_ao, 1.0, 4.0) * dirt_combined_ao;

    float dirt_mix = max(dirt_combined_inv_ao, dirt_combined_ao);
    dirt_mix = max(dirt_mix, clamp(dirt_value, 0.0, 1.0));
    dirt_mix = max(dirt_mix, nad.w * dirt_detail_mask);

    float roughness = (dirt_mix * dirt_roughness_blend) * (dirt_roughness - combined_detail_data.w) + combined_detail_data.w;
    vec3 temp_dirt_color = gamma(abs(dirt_color), 2.2);

    vec3 base_color = max(dirt_mix, 0.001) * (temp_dirt_color - base_albedo) + base_albedo;

    float up_fade = max(up_fade_distance, 0.001);
    float up_fade_crv = max(up_fade_curve, 0.001);

    // what's up_val?
    float up_val = curve(clamp(objPosition.y / up_fade, 0.0, 1.0), up_fade_crv);

    vec3 tint_color = up_val * (top_color_tint - bottom_color_tint) + bottom_color_tint;
    base_color = tint_color * base_color;

    float ao = tiler_ao_intensity * (combined_detail_data.z - 1.0) + 1.0;
    ao = (base_ao_intensity * (nad.z - 1.0) + 1.0) * ao;

    vec3 primary_graded_color = base_color;
    if (grading_group_id > 0.0) {
        primary_graded_color = gradeColor(base_color, gradingMatrix(gradingPixel(grading_group_id)));
    }
    primary_graded_color = clamp(primary_graded_color, 0.0, 1.0);

    vec3 secondary_graded_color = base_color;
    if (grading_secondary_group_id > 0.0) {
        secondary_graded_color = gradeColor(base_color, gradingMatrix(gradingPixel(grading_secondary_group_id)));
    }
    secondary_graded_color = clamp(secondary_graded_color, 0.0, 1.0);
    base_color = nad.w * (secondary_graded_color - primary_graded_color) + primary_graded_color;

    base_color = dirt_mix * dirt_metallic * (dirt_color - base_color) + base_color;

    float weathering_height_crv = max(weathering_height_curve, 0.001);


    switch (debug) {
    case 1:
        fragColor = vec4(combined_detail_albedo, 1.0);
        break;
    case 2:
        fragColor = vec4(detail_normal, 1.0);
        break;
    case 3:
        fragColor = vec4(combined_normal, 1.0);
        break;
    case 4:
        fragColor = vec4(mixed_vertex_normal, 1.0);
        break;
    case 5:
        fragColor = vec4(nad_normal, 1.0);
        break;
    default:
        fragColor = vec4(base_color, 1.0);
    }
}
