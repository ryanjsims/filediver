#version 430 core
#extension GL_ARB_shading_language_include : enable

#include "lighting.frag"

out vec4 fragColor;

in vec3 fragPosition;
in vec3 fragPositionObjectSpace;
in vec2 fragUV0;
in vec2 fragUV1;
in vec2 fragUV2;
in mat3 dbg_fragTBN;
in mat3 dbg_fragITBN;

uniform sampler2D decal_sheet;
uniform sampler2DArray composite_array;
uniform sampler2DArray customization_camo_tiler_array;
uniform sampler2DArray customization_material_detail_tiler_array;
uniform sampler2D pattern_lut;
uniform sampler2D base_data;
uniform sampler2D material_lut;
uniform sampler2DArray pattern_masks_array;
uniform sampler2DArray id_masks_array;

uniform sampler2D weathering_dirt;
uniform sampler2D weathering_special;
uniform sampler2D blood_splatter_tiler;
uniform sampler2D bug_splatter_tiler;

layout(shared, binding = 0) uniform DecalSettingsBlock {
    uint seed;
    uint debug;
    bool use_decals;
    float damage_mask_selector;
    float detail_tile_factor_mult;
    float decal_id;
    float decal_id_offset_x;
    float decal_id_offset_y;
    float decal_size_x;
    float decal_size_y;
    float decal_lut_id;
    float decal_lut_size_x;
    float decal_lut_size_y;
    float decal_normal_intentity;
    float decal_normal_offset;
    vec2 decal_scalarfield_end;
    float decal_alpha_offset;
    float decal_alpha_sharpness;
    float decal_channel_selection;
    float use_decal_rgb_channels;
    float underlying_normal_behind_decal_opacity;
};

layout(shared, binding = 1) uniform WeatheringSettingsBlock {
    float blood_gunk_normal_intensity;
    float weathering_tile_factor;
    vec3 blood_color;
    vec3 special_gunk_color;
    float blood_normal_fade;
    float blood_scale;
    float bug_gunk_minimum;
    float gunk_normal_fade;
    float gunk_scale;

    uint weathering_special_case;
    float frost_weight;

    // positive/negative blood?? weights
    vec3 blood_weights_positive;
    vec3 blood_weights_negative;

    // positive/negative bug_gunk?? weights
    vec3 bug_gunk_weights_positive;
    vec3 bug_gunk_weights_negative;

    // positive/negative weathering_dirt?? weights
    vec3 weathering_dirt_weights_positive;
    vec3 weathering_dirt_weights_negative;

    // positive/negative weathering_special?? weights
    vec3 weathering_special_weights_positive;
    vec3 weathering_special_weights_negative;

    vec2 setting_18;

    // Height wetness and wash
    vec4 height_wetness_and_wash;
};

layout(shared, binding = 1) uniform SnowBlock {
    float material_wetness;
    vec3 snow_base_color;
    vec3 snow_fuzz_color;
    vec3 snow_subsurface_color;
    float snow_subsurface_intensity;
    float snow_subsurface_wrap;
    float snow_subsurface_diffusion;
    float snow_specular;
};

vec3 rowColors[8] = {
    {0, 0, 0},
    {1, 0, 0},
    {0, 1, 0},
    {0, 0, 1},
    {1, 1, 0},
    {1, 0, 1},
    {0, 1, 1},
    {1, 1, 1},
};

uint getMaterialLutRow() {
    mat4 identity = mat4(1.0);
    uint id_masks_len = textureSize(id_masks_array, 0).z;
    uint id_mask_layer = 0;
    uint material_lut_row = 0;
    while(id_mask_layer < id_masks_len) {
        vec4 id_mask_sample = texture(id_masks_array, vec3(fragUV0, float(id_mask_layer)));
        uint row_base = id_mask_layer * 4;
        uint channel = 0;
        while(channel < 4) {
            if (dot(id_mask_sample, identity[channel]) >= 0.496) {
                material_lut_row = row_base + channel;
                break;
            }
            channel = channel + 1;
        }

        id_mask_layer = id_mask_layer + 1;
    }
    return material_lut_row;
}

vec3 gamma(vec3 color, float value) {
    return exp2(log2(color) * vec3(value));
}

// base and detail should be unpacked already, and this will return an unpacked normal
vec3 reorientNormalMaps(vec3 base, vec3 detail) {
    vec3 t = base + vec3(0, 0, 1);
    vec3 u = detail * vec3(-1, -1, 1);
    return normalize(t*dot(t, u)/t.z - u);
}

vec4 getBloodGunkDirtSpecialWeights() {
    vec3 positive_normal = clamp(dbg_fragTBN[2], 0.0, 1.0);
    vec3 negative_normal = clamp(-dbg_fragTBN[2], 0.0, 1.0);
    vec4 tempPos = vec4(blood_weights_positive.x, bug_gunk_weights_positive.x, weathering_dirt_weights_positive.x, weathering_special_weights_positive.x);
    vec4 tempNeg = vec4(blood_weights_negative.x, bug_gunk_weights_negative.x, weathering_dirt_weights_negative.x, weathering_special_weights_negative.x);
    tempNeg = negative_normal.x * tempNeg;
    vec4 combined = positive_normal.x * tempPos + tempNeg;

    tempPos = vec4(blood_weights_positive.y, bug_gunk_weights_positive.y, weathering_dirt_weights_positive.y, weathering_special_weights_positive.y);
    combined = positive_normal.y * tempPos + combined;

    tempNeg = vec4(blood_weights_negative.y, bug_gunk_weights_negative.y, weathering_dirt_weights_negative.y, weathering_special_weights_negative.y);
    combined = negative_normal.y * tempNeg + combined;

    tempPos = vec4(blood_weights_positive.z, bug_gunk_weights_positive.z, weathering_dirt_weights_positive.z, weathering_special_weights_positive.z);
    combined = positive_normal.z * tempPos + combined;
    
    tempNeg = vec4(blood_weights_negative.z, bug_gunk_weights_negative.z, weathering_dirt_weights_negative.z, weathering_special_weights_negative.z);
    return clamp(negative_normal.z * tempNeg + combined, 0.0, 1.0);
}

void main() {
    float detail_tile_factor_mult = 1.0;

    vec4 base_data_sample = texture(base_data, fragUV0);
    vec3 normal = vec3(base_data_sample.xy * 2.0 - 1.0, 0);
    normal.z = reconstructNormalZ(normal.xy);
    normal.x = -normal.x;
    vec3 normal_modified = normal;
    float ao = 1 - (base_data_sample.z * 0.4 + 0.6);
    float curvature = base_data_sample.w;

    uint material_lut_row = getMaterialLutRow();

    vec4 base_color = texelFetch(material_lut, ivec2(0, material_lut_row), 0);
    vec4 detail_layer = texelFetch(material_lut, ivec2(1, material_lut_row), 0);
    vec4 lut_2 = texelFetch(material_lut, ivec2(2, material_lut_row), 0);
    vec4 lut_3 = texelFetch(material_lut, ivec2(3, material_lut_row), 0);
    vec4 lut_4 = texelFetch(material_lut, ivec2(4, material_lut_row), 0);
    vec4 lut_5 = texelFetch(material_lut, ivec2(5, material_lut_row), 0);
    vec4 lut_6 = texelFetch(material_lut, ivec2(6, material_lut_row), 0);
    vec4 metallic_detail_controls = texelFetch(material_lut, ivec2(7, material_lut_row), 0);
    vec4 specular_detail_controls = texelFetch(material_lut, ivec2(8, material_lut_row), 0);
    vec4 lut_9 = texelFetch(material_lut, ivec2(9, material_lut_row), 0);
    vec4 lut_10 = texelFetch(material_lut, ivec2(10, material_lut_row), 0);
    vec4 lut_11 = texelFetch(material_lut, ivec2(11, material_lut_row), 0);
    vec4 lut_12 = texelFetch(material_lut, ivec2(12, material_lut_row), 0);
    vec4 lut_13 = texelFetch(material_lut, ivec2(13, material_lut_row), 0);
    vec4 lut_14 = texelFetch(material_lut, ivec2(14, material_lut_row), 0);
    vec4 lut_15 = texelFetch(material_lut, ivec2(15, material_lut_row), 0);
    float roughness = lut_10.x;
    vec4 camo_controls = texelFetch(material_lut, ivec2(21, material_lut_row), 0);
    vec4 detail_tiling = texelFetch(material_lut, ivec2(22, material_lut_row), 0);

    // overwritten by composite_array, if used
    vec4 material_detailing_xy = vec4(0);

    vec4 material_detailing_zw = vec4(0);
    vec4 material_detail_sample = vec4(0);

    vec4 r15 = vec4(clamp(lut_13.zw, 0.0, 1.0), 0.0, 0.0);

    // r12.w
    float detail_normal_2_alignment;
    vec4 material_detail_sample_2 = vec4(0);
    uvec2 random_detail_offset = uvec2(seed, seed >> 9);
    uvec2 bitmask = (((uvec2(1) << uvec2(23)) - uvec2(1)) << uvec2(0)) & 0xffffffff;
    random_detail_offset = (random_detail_offset & bitmask.xy) | (uvec2(0x3f800000) & ~bitmask.xy);
    vec2 detail_offset = uintBitsToFloat(random_detail_offset) - 1;

    uint material_detail_len = textureSize(customization_material_detail_tiler_array, 0).z;
    if (base_color.w <= 3.0) {
        float detail_scaling = detail_tiling.x * detail_tile_factor_mult;
        vec3 detail_uvs = vec3(fragUV1 * detail_scaling + detail_offset, float(uint(detail_layer.x) % material_detail_len));
        material_detail_sample = texture(customization_material_detail_tiler_array, detail_uvs) - 0.5;
        material_detailing_xy.xy = material_detail_sample.zw;
        material_detailing_zw.xy = material_detail_sample.xy * vec2(-1, 1);
    }

    vec3 material_detailing_normal;
    if (base_color.w >= 2.0 && lut_10.z > 0.0) {
        vec3 detail_uvs_2 = vec3(detail_tiling.x * detail_tile_factor_mult * fragUV1 + detail_offset, float(uint(lut_10.y) % material_detail_len));
        material_detail_sample_2 = texture(customization_material_detail_tiler_array, detail_uvs_2) - 0.5;
        // r16.zw
        material_detailing_zw.zw = material_detail_sample_2.xy * vec2(-2, 2);
        vec2 temp_sq = material_detailing_zw.zw * material_detailing_zw.zw;
        float material_detailing_normal_z = max(max(temp_sq.x, temp_sq.y) * 0.000061, reconstructNormalZ(material_detailing_zw.zw));
        material_detailing_zw.zw = -inversesqrt(material_detailing_normal_z) * material_detailing_zw.zw * lut_10.w;
        normal_modified = normalize(vec3(material_detailing_zw.zw + normal_modified.xy, normal_modified.z * material_detailing_normal_z));
        material_detail_sample_2.xyz = vec3(material_detail_sample_2.zw, clamp((curvature - 0.5), 0.0, 1.0));
        detail_normal_2_alignment = dot(lut_11.xyz, material_detail_sample_2.xyz) + lut_11.w;
    } else {
        lut_10.z = 0.0;
        detail_normal_2_alignment = 0.0;
    }
    material_detailing_normal = normal_modified;

    if (base_color.w == 4.0) {
        vec3 composite_uvs = vec3(detail_tiling.x * detail_tile_factor_mult * 4.0 * fragUV1, float(uint(detail_layer.x) % material_detail_len /*not composite length, yes*/));
        vec3 composite_sample = texture(composite_array, composite_uvs).xyz;
        material_detailing_xy.xyz = composite_sample.zxy - 0.5;
        material_detailing_xy.w = -(composite_sample.x - 0.5);
        material_detailing_zw.zw = vec2(-(composite_sample.x - 0.5), composite_sample.y - 0.5);
        material_detailing_xy.y = 0.0;
    } else {
        material_detailing_zw.zw = material_detailing_zw.xy;
    }

    vec4 scaled_lut_4 = vec4(lut_4.xy * max((curvature - 0.5) * lut_4.z, 0.0) * 4, 0.0, 0.0);
    lut_4 = lut_4 + scaled_lut_4;
    material_detailing_xy.zw = base_data_sample.wz * vec2(1, -0.4) + vec2(-0.5, 0.4);
    // r15.z
    float detail_roughness = clamp(dot(lut_4, material_detailing_xy) + detail_layer.w, 0.0, 1.0);
    
    vec4 scaled_lut_3 = vec4(lut_3.xy * max((curvature - 0.5) * lut_3.z, 0.0) * 4, 0.0, 0.0);
    lut_3 = lut_3 + scaled_lut_3;
    vec4 modified_material_detail_tiling = material_detailing_xy;
    float temp_r15w = clamp(material_detailing_xy.y + 0.5, 0.0, 1.0);
    modified_material_detail_tiling.w = ao * (ao * (1 - temp_r15w) + temp_r15w);
    // r15.w, r14.x in voidwalker
    float detail_intensity = clamp(dot(lut_3, modified_material_detail_tiling) + detail_layer.z, 0.0, 1.0);

    vec4 scaled_metallic_detail_controls = vec4(metallic_detail_controls.xy * max((curvature - 0.5) * metallic_detail_controls.z, 0.0) * 4, 0.0, 0.0);
    metallic_detail_controls = metallic_detail_controls + scaled_metallic_detail_controls;
    // r15.x
    float detail_metallic = clamp(lut_6.w + dot(metallic_detail_controls, material_detailing_xy), 0.0, 1.0);

    vec3 scaled_specular_detail_controls = vec3(specular_detail_controls.xy * max((curvature - 0.5) * specular_detail_controls.z, 0.0) * 4, 0.0);
    scaled_specular_detail_controls = scaled_specular_detail_controls + specular_detail_controls.xyz;
    // r15.y
    float detail_specular = clamp(specular_detail_controls.w + dot(scaled_specular_detail_controls, material_detailing_xy.xyz), 0.0, 1.0);

    vec3 scaled_lut_14 = vec3(lut_14.xy * max((curvature - 0.5) * lut_14.z, 0.0) * 4.0, 0.0);
    vec3 temp_lut_14 = scaled_lut_14 + lut_14.xyz;
    float r16y = clamp(lut_14.w + dot(temp_lut_14, modified_material_detail_tiling.xyz), 0.0, 1.0);

    vec3 scaled_lut_15 = vec3(lut_15.xy * max((curvature - 0.5) * lut_15.z, 0.0) * 4.0, 0.0);
    vec3 temp_lut_15 = scaled_lut_15 + lut_15.xyz;
    r15.w = clamp(lut_15.w + dot(temp_lut_15, modified_material_detail_tiling.xyz), 0.0, 1.0);

    vec2 temp_detailing;
    if (base_color.w <= 3.0) {
        lut_2.w = detail_intensity * (lut_2.w - 1.0) + 1.0;
        lut_2.w = detail_roughness * (lut_5.w - lut_2.w) + lut_2.w;

        temp_detailing = material_detailing_zw.xy + material_detailing_zw.xy;
        vec2 temp_detailing_sq = temp_detailing * temp_detailing;
        float detailing_z = 1 - temp_detailing.x * temp_detailing.x - temp_detailing.y * temp_detailing.y;
        detailing_z = inversesqrt(max(max(temp_detailing_sq.x, temp_detailing_sq.y) / 16384.0, detailing_z));
        temp_detailing = temp_detailing * -detailing_z * detail_layer.y * lut_2.w;

        normal_modified = normalize(vec3(temp_detailing + normal_modified.xy, normal_modified.z * detailing_z));
    }

    if (base_color.w == 4.0) {
        temp_detailing = material_detailing_zw.zw * 2;
        vec2 temp_detailing_sq = temp_detailing * temp_detailing;
        float detailing_z = 1 - temp_detailing.x * temp_detailing.x - temp_detailing.y * temp_detailing.y;
        detailing_z = inversesqrt(max(max(temp_detailing_sq.x, temp_detailing_sq.y) / 16384.0, detailing_z));
        temp_detailing = temp_detailing * -detailing_z * detail_layer.y * lut_2.w;

        normal_modified = normalize(vec3(temp_detailing + normal_modified.xy, normal_modified.z * detailing_z));
    }

    vec4 scaled_lut_9 = vec4(lut_9.xy * max((curvature - 0.5) * lut_9.z, 0.0) * 4.0, 0.0, 0.0);
    lut_9 = lut_9 + scaled_lut_9;
    vec4 temp_modified_material_detail_tiling = vec4(modified_material_detail_tiling.xy, clamp(curvature - 0.5, 0.0, 1.0), clamp(-(curvature - 0.5), 0.0, 1.0));
    // r16.w in voidwalker shader
    float detail_control_lut_9 = clamp(roughness + dot(lut_9, temp_modified_material_detail_tiling), 0.0, 1.0);
    

    vec3 camo_sample = vec3(0);
    vec4 camo_result = vec4(0);
    vec3 camo_combined_1 = vec3(0);
    float camo_mix = 0;

    if (camo_controls.w >= 0.0) {
        vec4 camo_color_1 = texelFetch(material_lut, ivec2(16, material_lut_row), 0);
        vec4 camo_color_2 = texelFetch(material_lut, ivec2(17, material_lut_row), 0);
        vec4 camo_color_3 = texelFetch(material_lut, ivec2(18, material_lut_row), 0);
        vec4 camo_color_4 = texelFetch(material_lut, ivec2(19, material_lut_row), 0);
        float camo_control_extra = texelFetch(material_lut, ivec2(20, material_lut_row), 0).w;

        vec3 camo_uvs = vec3(detail_tile_factor_mult * camo_controls.z * fragUV1, camo_controls.w);
        camo_sample = clamp((texture(customization_camo_tiler_array, camo_uvs).xyz - 0.5) * camo_controls.xxx + camo_controls.yyy, 0.0, 1.0);
        camo_result = mix(camo_color_1, camo_color_2, camo_sample.x);
        camo_result = mix(camo_result, camo_color_3, camo_sample.y);
        camo_result = mix(camo_result, camo_color_4, camo_sample.z);

        if (base_color.w != 1.0) {
            float camo_edge_wear = max((curvature - 0.5) * camo_color_4.w, 0.0) * 4;
            camo_combined_1 = vec3(camo_edge_wear * camo_color_2.w, camo_edge_wear * camo_color_3.w, 0.0);
            vec3 camo_combined_2 = vec3(camo_color_2.w, camo_color_3.w, camo_color_4.w);
            camo_combined_1 = camo_combined_1 + camo_combined_2;
            camo_mix = clamp(dot(camo_combined_1, material_detailing_xy.xyz) - camo_control_extra, 0.0, 1.0);
            base_color.xyz = mix(camo_result.xyz, base_color.xyz, camo_mix);
            detail_metallic = max(detail_metallic - (1.0 - camo_mix), 0.0);
            float camo_tiling = detail_tiling.w * (1.0 - camo_mix);
            roughness = camo_tiling * (camo_color_1.w - roughness) + roughness;
            r16y = r16y * camo_mix;
        } else {
            base_color.xyz = camo_result.xyz;
            camo_mix = 1.0;
        }
    } else {
        camo_mix = 1.0;
    }

    // pattern mask
    if (base_color.w != 3.0) {
        vec4 pattern_color = texelFetch(pattern_lut, ivec2(0, 0), 0);
        if (pattern_color.w != -1.0) {
            int depth = textureSize(pattern_masks_array, 0).z;
            vec4 pattern_controls_1 = texelFetch(pattern_lut, ivec2(1, 0), 0);
            float pattern_edge_wear = max((curvature - 0.5) * pattern_controls_1.z, 0.0) * 4.0;
            vec3 pattern_normal = vec3(pattern_controls_1.xy * pattern_edge_wear, 0.0) + pattern_controls_1.xyz;
            float pattern_mix = clamp(dot(pattern_normal, material_detailing_xy.xyz) + pattern_controls_1.w, 0.0, 1.0);
            vec3 pattern_mask_uv = vec3(fragUV0, float(int(pattern_color.w) % depth));
            float mask_sample = clamp(100 * (texture(pattern_masks_array, pattern_mask_uv).x - 0.5), 0.0, 1.0);
            float pattern_mix_masked = pattern_mix * mask_sample;
            if (pattern_mix_masked > 0.0) {
                base_color.xyz = mix(base_color.xyz, pattern_color.xyz, pattern_mix_masked);
                if (base_color.w != 1.0) {
                    float pattern_roughness = min(detail_roughness - (pattern_mix_masked * pattern_mix_masked) + 1.0, 1.0);
                    detail_metallic = detail_metallic * pattern_roughness;
                    float pattern_control_2 = texelFetch(pattern_lut, ivec2(2, 0), 0).w;
                    float r18w = max(pattern_mix * mask_sample - detail_roughness, 0.0);
                    float temp_pattern_mix = r18w * (1.0 / (2.0 - roughness));
                    roughness = temp_pattern_mix * (pattern_roughness - roughness) + roughness;
                    r16y = r16y * (1.0 - pattern_mix_masked);
                    detail_specular = pattern_mix_masked * (0.5 - detail_specular) + detail_specular;
                }
                //float pattern_roughness = 1.0;
            }
        }
    }

    bvec2 decal_range = lessThan(abs(fragUV2 - 0.5), vec2(0.5));
    if(use_decals && decal_range.x && decal_range.y) {
        vec4 decal_sheet_size = vec4(textureSize(decal_sheet, 0).xy, 0, 0);
        decal_sheet_size.zw = max(decal_sheet_size.xy, vec2(1.0));
        
        // note: spelled as "decal_normal_intentity" in files
        float decal_normal_intensity = decal_normal_intentity;
        float decal_id_val = max(floor(decal_id + 0.5), 0.0);

        // r27
        vec4 decal_size_xyzw = vec4(max(decal_size_x, 1.0), max(decal_size_y, 1.0), decal_sheet_size.zw * fragUV2);
        vec2 decal_offset = fragUV2 * decal_sheet_size.zw + (decal_id_val * vec2(decal_id_offset_x, -decal_id_offset_y));
        decal_offset = decal_offset / decal_sheet_size.zw;

        vec2 decal_lut_size = vec2(decal_lut_size_x, decal_lut_size_y);
        if (decal_lut_size_x > 1.0 || decal_lut_size_y > 1.0) {
            decal_lut_size = decal_lut_size  / decal_sheet_size.zw;
        }
        vec2 temp = min(max(vec2(1.0) / decal_sheet_size.zw, decal_lut_size), vec2(1.0));
        decal_lut_size = vec2(1) - temp;
        temp = temp * vec2(1.0/6.0, 1.0/3.0);

        float decal_lut_id_val = clamp(floor(decal_lut_id + 0.5), 0.0, 5.0);

        vec4 decal_sample = texture(decal_sheet, fragUV2);

        bvec2 decal_size_scaled_down = lessThan(decal_size_xyzw.zw, decal_size_xyzw.xy);
        if (decal_size_scaled_down.x && decal_size_scaled_down.y) {
            decal_sample = texture(decal_sheet, decal_offset);
            if (use_decal_rgb_channels > 0.5) {
                float channel = (decal_channel_selection == 2.0) ? decal_sample.z : decal_sample.x;
                channel = (decal_channel_selection == 1.0) ? decal_sample.y : channel;

                float decal_lut_id_plus_half = decal_lut_id_val + 0.5;
                vec4 decal_lut_offset = vec4(decal_lut_id_plus_half, 0.5, decal_lut_id_plus_half, 1.5);
                decal_lut_offset = decal_lut_offset * temp.xyxy + decal_lut_size.xyxy;

                vec2 decal_lut_offset_2 = vec2(decal_lut_id_plus_half, 2.5);
                decal_lut_offset_2 = decal_lut_offset_2 * temp.xy + decal_lut_size.xy;

                vec3 decal_red_sample = texture(decal_sheet, decal_lut_offset.xy).xyz;
                vec3 decal_green_sample = texture(decal_sheet, decal_lut_offset.zw).xyz;
                vec3 decal_blue_sample = texture(decal_sheet, decal_lut_offset_2.xy).xyz;

                float channel_mix_1 = (channel < 1.0/3.0) ? 1.0 : 0.0;
                float channel_mix_2 = (channel < 2.0/3.0 && channel >= 1.0/3.0) ? 1.0 : 0.0;
                float channel_mix_3 = (channel >= 2.0/3.0) ? 1.0 : 0.0;

                decal_sample.xyz = channel_mix_2 * decal_red_sample + (channel_mix_1 * decal_blue_sample + (channel_mix_3 * decal_green_sample));
            }
        }

        bvec2 decal_uv_in_scalarfield = lessThan(vec2(fragUV2.x, 1 - fragUV2.y), decal_scalarfield_end);
        if (decal_uv_in_scalarfield.x && decal_uv_in_scalarfield.y) {
            float inv_normal_offset = max(1 - decal_normal_offset, 0.001);

            vec2 offset = vec2(0.5) / decal_sheet_size.xy;

            float decal_sheet_alpha_neg_x = texture(decal_sheet, vec2(fragUV2.x - offset.x, fragUV2.y)).w;
            decal_sheet_alpha_neg_x = decal_sheet_alpha_neg_x - decal_normal_offset;
            decal_sheet_alpha_neg_x = clamp(decal_sheet_alpha_neg_x / inv_normal_offset, 0.0, 1.0);

            float decal_sheet_alpha_pos_x = texture(decal_sheet, vec2(fragUV2.x + offset.x, fragUV2.y)).w;
            decal_sheet_alpha_pos_x = decal_sheet_alpha_pos_x - decal_normal_offset;
            decal_sheet_alpha_pos_x = clamp(decal_sheet_alpha_pos_x / inv_normal_offset, 0.0, 1.0);

            float decal_sheet_alpha_neg_y = texture(decal_sheet, vec2(fragUV2.x, fragUV2.y - offset.y)).w;
            decal_sheet_alpha_neg_y = decal_sheet_alpha_neg_y - decal_normal_offset;
            decal_sheet_alpha_neg_y = clamp(decal_sheet_alpha_neg_y / inv_normal_offset, 0.0, 1.0);

            float decal_sheet_alpha_pos_y = texture(decal_sheet, vec2(fragUV2.x, fragUV2.y + offset.y)).w;
            decal_sheet_alpha_pos_y = decal_sheet_alpha_pos_y - decal_normal_offset;
            decal_sheet_alpha_pos_y = clamp(decal_sheet_alpha_pos_y / inv_normal_offset, 0.0, 1.0);

            float temp_alpha_x = decal_sheet_alpha_pos_x - decal_sheet_alpha_neg_x;
            float temp_alpha_y = decal_sheet_alpha_pos_y - decal_sheet_alpha_neg_y;
            vec2 temp_intensity = vec2(temp_alpha_x, temp_alpha_y) * decal_normal_intensity;

            float offset_alpha = decal_sample.w - decal_alpha_offset;
            decal_sample.w = clamp((offset_alpha / inv_normal_offset) * decal_alpha_sharpness, 0.0, 1.0);

            float normal_opacity = (underlying_normal_behind_decal_opacity - 1.0) * decal_sample.w + 1.0;
            vec3 temp_normal_modified = normal_modified * normal_opacity;

            float inv_normal_opacity = (1.0 - underlying_normal_behind_decal_opacity) * decal_sample.w;

            vec3 result_normal = -normal_modified * normal_opacity + normal;
            normal_modified = inv_normal_opacity * result_normal + temp_normal_modified;
        }

        float decal_mix = clamp(decal_sample.w - detail_roughness, 0.0, 1.0);
        vec3 temp_color = decal_sample.xyz - base_color.xyz;
        base_color.xyz = temp_color * decal_mix + base_color.xyz;
        r16y = decal_mix * -r16y + r16y;
        detail_metallic = detail_metallic * (1.0 - decal_mix);
        decal_mix = max(decal_mix - detail_intensity, 0.0);
        roughness = decal_mix * (0.4 - roughness) + roughness;
    }

    // r14.w
    float r17z, inv_detail_intensity;
    if (base_color.w == 1.0) {
        lut_13.xy = clamp(lut_13.xy, 0.0, 1.0);
        inv_detail_intensity = (1.0 - detail_intensity) * lut_13.x;
        vec3 sqrt_color = sqrt(base_color.xyz) - lut_12.xyz;
        lut_12.xyz = lut_13.y * sqrt_color + lut_12.xyz;
        r17z = (1 - (roughness * inv_detail_intensity));
        r17z = clamp(dot(base_color.xyz, vec3(0.3, 0.59, 0.11)), 0.0, 1.0) * (1.0 - r17z) + r17z;
    } else {
        lut_12.xyzw = vec4(0.0);
        inv_detail_intensity = 0.0;
        r17z = 1.0;
    }

    lut_2.xyz = detail_intensity * (lut_2.xyz - base_color.xyz) + base_color.xyz;
    detail_intensity = detail_tiling.z * detail_intensity * (detail_tiling.y - roughness) + roughness;
    float camo_roughness = detail_roughness * camo_mix;
    lut_2.xyz = camo_roughness * (lut_6.xyz - lut_2.xyz) + lut_2.xyz;
    camo_roughness = camo_roughness * detail_roughness;
    lut_2.xyz = camo_roughness * (lut_5.xyz - lut_2.xyz) + lut_2.xyz;
    lut_2.xyz = max(lut_2.xyz, vec3(0.000061));
    bvec3 notVerySmall = lessThan(vec3(0.04045), lut_2.xyz);
    vec3 srgb_temp = lut_2.xyz * vec3(0.947867) + vec3(0.052133);
    srgb_temp = gamma(srgb_temp, 2.4);
    lut_2.xyz = lut_2.xyz * vec3(0.0774);
    vec3 temp;
    temp.x = notVerySmall.x ? srgb_temp.x : lut_2.x;
    temp.y = notVerySmall.y ? srgb_temp.y : lut_2.y;
    temp.z = notVerySmall.z ? srgb_temp.z : lut_2.z;
    base_color.xyz = temp;

    vec2 height_wash = clamp(10 * (fragPositionObjectSpace.zz - height_wetness_and_wash.xz), 0.0, 1.0);
    height_wash = clamp((height_wash - height_wetness_and_wash.yw) + 1.0, 0.0, 1.0);
    vec4 bgds_weights = getBloodGunkDirtSpecialWeights() * height_wash.y;

    float special_weights_z = min(weathering_special_weights_negative.z * clamp(weathering_special_weights_negative.z * 1.25 - fragPositionObjectSpace.z, 0.0, 1.0), 0.85);
    special_weights_z = special_weights_z * 0.95 + bgds_weights.w;

    vec2 weathering_dirt_uvs = weathering_tile_factor * 0.66 * fragUV1;
    vec2 weathering_special_uvs = weathering_tile_factor * 0.33 * fragUV1;

    vec4 weathering_dirt_sample = texture(weathering_dirt, weathering_dirt_uvs);
    vec4 weathering_special_sample = texture(weathering_special, weathering_special_uvs);

    // Maybe metallic, maybe roughness
    float r14z = 4.0 * (curvature - 0.5);
    float weathering_dirt_weight = clamp(bgds_weights.z * -r14z + clamp(2 * (bgds_weights.z + weathering_dirt_sample.w - 1.0), 0.0, 1.0), 0.0, 1.0);
    float weathering_special_metallic = min((1.0 - detail_metallic) * clamp(curvature - 0.5, 0.0, 1.0) * 4.0 + (base_color.w == 1.0 ? 1.0 : 0.0), 1.0);

    float weathering_weight = (1.0 - normal.z) * weathering_special_metallic;
    weathering_weight = weathering_weight * weathering_weight + abs(weathering_dirt_sample.w - 0.5) + abs(weathering_special_sample.w - 0.5);
    weathering_weight = min(clamp(curvature - 0.5, 0.0, 1.0) * 2.0 + weathering_weight, 1.0);
    float temp_weathering_weight = weathering_weight * weathering_weight;
    weathering_weight = 3.0 - (2.0 * weathering_weight);
    float result_weathering_weight = temp_weathering_weight * weathering_weight;

    float weathering_mix = 0.0;
    if (weathering_special_case != 0) {
        if (weathering_special_case == 2) {
            vec2 temp_special_weights = vec2(clamp(special_weights_z + 0.65, 0.0, 1.0), special_weights_z - 0.1);
            float temp_weathering_mix = clamp(temp_special_weights.y * 10.0 / 9.0, 0.0, 1.0); 
            float temp_spec_weights_y2 = 3.0 - 2 * temp_weathering_mix;
            temp_weathering_mix = clamp((temp_weathering_mix * temp_weathering_mix) * temp_spec_weights_y2 - 0.5, 0.0, 1.0);
            temp_spec_weights_y2 = (1 - temp_weathering_mix) * special_weights_z;
            temp_weathering_mix = temp_special_weights.x * 0.2 + (temp_spec_weights_y2 * 0.7 + temp_weathering_mix);
            temp_spec_weights_y2 = clamp(weathering_special_sample.y + temp_weathering_mix - 1.0, 0.0, 1.0);
            temp_weathering_mix = clamp(temp_weathering_mix * -r14z + temp_spec_weights_y2, 0.0, 1.0);
            weathering_mix = clamp(1.0 - height_wetness_and_wash.y, 0.0, 1.0) * temp_weathering_mix;
        } else {
            weathering_mix = clamp(special_weights_z * -r14z + clamp((special_weights_z + weathering_special_sample.w - 1.0)* 8.0, 0.0, 1.0), 0.0, 1.0);
        }
        float special_case = weathering_mix * min(abs(float(weathering_special_case) - 2.0), 1.0) + weathering_dirt_weight;
        weathering_dirt_weight = min(special_case, 1.0);
    }

    float inv_dirt_weight = 1.0 - weathering_dirt_weight;
    base_color.xyz = inv_dirt_weight * base_color.xyz;
    inv_detail_intensity = inv_dirt_weight * inv_detail_intensity;
    r15.z = inv_dirt_weight * r16y;
    // r16.y
    float one_minus_dirt_weight_cubed = 1.0 - pow(weathering_dirt_weight, 3.0);
    float dirt_detail_metallic = one_minus_dirt_weight_cubed * detail_metallic;
    inv_dirt_weight = inv_dirt_weight * lut_10.x;
    detail_intensity = weathering_dirt_weight * (1.0 - detail_intensity) + detail_intensity;

    vec3 dirt_normal = vec3(weathering_dirt_sample.xy * 2 - 1, 0.0);
    dirt_normal.x = -dirt_normal.x;
    dirt_normal.z = reconstructNormalZ(dirt_normal.xy);
    vec3 temp_dirt_normal = vec3(dirt_normal.xy * weathering_dirt_weight, 0.0);
    temp_dirt_normal.z = reconstructNormalZ(temp_dirt_normal.xy);
    normal_modified = reorientNormalMaps(normal_modified, temp_dirt_normal);

    // r11.z = clamp((curvature - 0.5) /*r12.x*/, 0.0, 1.0);
    // r11.w = clamp(-(curvature - 0.5) /*r12.x*/, 0.0, 1.0);

    vec3 special_normal = vec3(weathering_special_sample.xy * 2 - 1, 0.0);
    special_normal.x = -special_normal.x;
    if (setting_18.x > 0.0) {
        vec3 temp_base_color = base_color.xyz * min(2 * result_weathering_weight, 1.0);
        temp_base_color = temp_base_color * vec3(-0.769, -0.847, -0.882) + base_color.xyz;
        temp_base_color = (result_weathering_weight * temp_base_color.xyz) * vec3(-0.902, -0.890, -0.871) + temp_base_color.xyz;
        float r14y = (result_weathering_weight * 0.5);
        r14y = r14y * r14y;
        base_color.xyz = r14y * (vec3(0.98, 0.11, 0.129) - temp_base_color) + temp_base_color;
        float r12z = min(temp_weathering_weight * weathering_weight + clamp(curvature - 0.5, 0.0, 1.0), 1.0);
        r15 = r12z * (vec4(0.1, 0.35, 1.0, 1.0) - r15) + r15;

        vec3 combined_dirt_normal = 4.0 * result_weathering_weight * (dirt_normal + special_normal) * (1 - detail_metallic * one_minus_dirt_weight_cubed);
        float strength = length(combined_dirt_normal.xy);
        if (strength > 0.99) {
             combined_dirt_normal.xy = combined_dirt_normal.xy / strength;
        }
       
        combined_dirt_normal.z = reconstructNormalZ(combined_dirt_normal.xy);
        normal_modified = reorientNormalMaps(normal_modified, combined_dirt_normal);

        detail_intensity = result_weathering_weight * (1 - 0.5 * dirt_detail_metallic) + detail_intensity;
    }

    vec2 blood_uvs = fragUV1 * blood_scale * weathering_tile_factor;
    vec4 blood_splatter_sample = texture(blood_splatter_tiler, blood_uvs);

    vec2 inv_splatter_weights = vec2(8.0, 4.0) - inv_detail_intensity * vec2(6.0, 3.0);
    float blood_weight = clamp((bgds_weights.x * blood_splatter_sample.w) * inv_splatter_weights.x - inv_splatter_weights.y, 0.0, 1.0);
    vec4 splatter_weights = vec4(vec2(0), blood_splatter_sample.zw * blood_weight);

    vec2 gunk_uvs = fragUV1 * gunk_scale * weathering_tile_factor + 0.5;
    vec4 bug_splatter_sample = texture(bug_splatter_tiler, gunk_uvs);
    float gunk_weight = clamp(min(max(bgds_weights.y, bug_gunk_minimum), 1.0) * bug_gunk_weights_negative.z * inv_splatter_weights.x - inv_splatter_weights.y, 0.0, 1.0);

    splatter_weights.xy = bug_splatter_sample.zw * gunk_weight;
    vec3 secondary_splatter_weights = vec3(blood_splatter_sample.wz * blood_weight + splatter_weights.yx, 0.0);
    secondary_splatter_weights.yz = clamp(secondary_splatter_weights.xy, 0.0, 1.0);

    float r16w = 1.0 - secondary_splatter_weights.y;
    float r17x = r16w * inv_detail_intensity;
    float gunk_detail_weight = clamp(bug_splatter_sample.z * gunk_weight + (blood_splatter_sample.z * blood_weight + r17x), 0.0, 1.0);

    vec4 lut_12_splatter = lut_12 * r16w;
    float r14y = r15.z * r16w;

    detail_intensity = secondary_splatter_weights.y * (0.75 - detail_intensity) + detail_intensity;
    detail_specular = secondary_splatter_weights.y * (0.5 - detail_specular) + detail_specular;

    r17z = 1.0 - 0.96 * dirt_detail_metallic;
    float r16x = detail_metallic * one_minus_dirt_weight_cubed + 1.0;
    vec3 blood_gunk_modified_color = r17z * (base_color.xyz * vec3(0.6, 0.016, 0.016)) - base_color.xyz;
    base_color.xyz = clamp(r16x * splatter_weights.w, 0.0, 1.0) * blood_gunk_modified_color + base_color.xyz;

    vec3 blood_color_linear = gamma(abs(blood_color), 2.2);
    splatter_weights.xz = clamp(splatter_weights.xz, 0.0, 1.0);
    base_color.xyz = splatter_weights.z * (blood_color_linear - base_color.xyz) + base_color.xyz;

    r16x = clamp(splatter_weights.y * r16x, 0.0, 1.0);
    blood_gunk_modified_color = (r17z * base_color.xyz) * vec3(0.816, 0.835, 0.267) - base_color.xyz;
    base_color.xyz = r16x * blood_gunk_modified_color + base_color.xyz;

    vec3 special_gunk_color_linear = gamma(abs(special_gunk_color), 2.2);
    base_color.xyz = splatter_weights.x * (special_gunk_color_linear - base_color.xyz) + base_color.xyz;

    r16x = 1.0 - clamp(secondary_splatter_weights.x * 8.0 - 3.0, 0.0, 1.0);
    r16y = dirt_detail_metallic * r16x;

    blood_gunk_modified_color = clamp(0.75 * ((splatter_weights.y * blood_color_linear) + (splatter_weights.w * special_gunk_color_linear)), 0.0, 1.0);
    float subsurface_intensity = secondary_splatter_weights.z * 0.025;
    float subsurface_wrap;
    float subsurface_diffusion;

    if (frost_weight.x >= 0.001) {
        vec4 modified_dirt_color = weathering_dirt_sample.z * vec4(0.384, 0.25, 0.18, 0.25) + vec4(0.616, 0.75, 0.82, 0.25);
        r17z = clamp(frost_weight.x, 0.0, 0.41);
        float dirt_curvature = clamp(weathering_special_sample.z * 0.5 + (weathering_dirt_sample.w - 2.0 * (curvature - 0.5)), 0.0, 1.0) - (1.0 - r17z);
        dirt_curvature = (max(dirt_curvature, 0.0) + 0.1) * r17z;
        r17z = dirt_curvature * 2;
        vec3 combined_dirt_normal = dirt_curvature * 4.0 * (dirt_normal + special_normal) * (1.0 - dirt_detail_metallic * r16x);
        float strength = length(combined_dirt_normal.xy);
        if (strength > 0.99) {
             combined_dirt_normal.xy = combined_dirt_normal.xy / strength;
        }
        combined_dirt_normal.z = reconstructNormalZ(combined_dirt_normal.xy);
        normal_modified = reorientNormalMaps(normal_modified, combined_dirt_normal);

        base_color.xyz = (modified_dirt_color.xyz - base_color.xyz) * dirt_curvature + base_color.xyz;

        detail_specular = dirt_curvature * (1.0 - detail_specular) + detail_specular;
        detail_intensity = dirt_curvature * (0.1 - detail_intensity) + detail_intensity;
        blood_gunk_modified_color = dirt_curvature * (modified_dirt_color.xyz - blood_gunk_modified_color) + blood_gunk_modified_color;

        subsurface_intensity = dirt_curvature * (modified_dirt_color.y - 0.025 * secondary_splatter_weights.z) + subsurface_intensity;
        subsurface_wrap = dirt_curvature * modified_dirt_color.w + 0.5;
        subsurface_diffusion = dirt_curvature * (weathering_dirt_sample.z * 0.25 + 0.15) + 0.6;
        secondary_splatter_weights.y = dirt_curvature * secondary_splatter_weights.y;
    } else {
        subsurface_wrap = 0.5;
        subsurface_diffusion = 0.6;
        r17z = 1.0;
    }

    float snow_mix;
    if (weathering_special_case == 2) {
        float inv_weathering_mix = 1.0 - weathering_mix;
        special_weights_z = special_weights_z * (1.0 - abs(setting_18.x * 2.0 - 1.0));
        snow_mix = secondary_splatter_weights.y * inv_weathering_mix + special_weights_z;

        float inv_scorch = 1.0 - setting_18.x;
        float r18w = weathering_mix * inv_scorch * inv_scorch;
        inv_scorch = 1.0 - (weathering_mix * weathering_mix) * (inv_scorch * inv_scorch);

        r16y = inv_scorch * r16y;
        r17z = 1.0 - 2.0 * r18w;
        weathering_mix = r16w * r18w;
        base_color.xyz = weathering_mix * (snow_base_color - base_color.xyz) + base_color.xyz;
        lut_12_splatter = weathering_mix * (vec4(snow_fuzz_color.xyz - (r16w * lut_12.xyz), 1.0 - r16w * lut_12.w)) + lut_12_splatter;
        blood_gunk_modified_color = weathering_mix * (snow_subsurface_color.xyz - blood_gunk_modified_color.xyz) + blood_gunk_modified_color;
        subsurface_intensity = weathering_mix * (snow_subsurface_intensity - subsurface_intensity) + subsurface_intensity;
        subsurface_wrap = weathering_mix * (snow_subsurface_wrap - subsurface_wrap) + subsurface_wrap;
        subsurface_diffusion = weathering_mix * (snow_subsurface_diffusion - subsurface_diffusion) + subsurface_diffusion;
        detail_specular = weathering_mix * (snow_specular - detail_specular) + detail_specular;
        detail_intensity = weathering_mix * (0.75 - detail_intensity) + detail_intensity;
    } else {
        snow_mix = weathering_mix * clamp((1.0 - setting_18.x) - (result_weathering_weight * setting_18.x), 0.0, 1.0) + secondary_splatter_weights.y;
    }

    snow_mix = clamp(snow_mix + setting_18.y + (1.0 - height_wash.x), 0.0, 1.0);
    normal = weathering_mix * (normal + special_normal - normal_modified) + normal_modified;

    vec3 blood_gunk_normal = vec3(r17z * vec2(-1.0, 1.0) * (blood_splatter_sample.xy + bug_splatter_sample.xy - 1.0) * blood_gunk_normal_intensity, 0.0);
    vec2 bgn_sq = blood_gunk_normal.xy * blood_gunk_normal.xy;
    float bgn_inv_z = inversesqrt(max(1.0 - length(blood_gunk_normal.xy), max(bgn_sq.x, bgn_sq.y) / 16384));
    blood_gunk_normal.xy = blood_gunk_normal.xy * -bgn_inv_z;
    //blood_gunk_normal = dbg_TBN[0] * blood_gunk_normal.x + dbg_TBN[1] * blood_gunk_normal.y;
    
    float blood_gunk_normal_factor = clamp(splatter_weights.w * blood_normal_fade + splatter_weights.y * gunk_normal_fade, 0.0, 1.0);
    blood_gunk_normal_factor = exp2(log2(1.0 - (r16w * inv_detail_intensity)) * 20.0) * blood_gunk_normal_factor;

    blood_gunk_normal.xy = blood_gunk_normal.xy * -blood_gunk_normal_factor;
    blood_gunk_normal.z = reconstructNormalZ(blood_gunk_normal.xy);
    normal = normalize(reorientNormalMaps(normal, blood_gunk_normal));

    float r13w = max(r16y, 0.0);
    float r14w = r17x;
    if (r13w > 0.0) {
        r14w = 0.0;
    }

    if (r14w > 0.0) {
        inv_dirt_weight = 0.0;
    }

    detail_specular = clamp(detail_specular, 0.0, 1.0);
    detail_intensity = clamp(detail_intensity, 0.0, 1.0) * 0.955 + 0.045;

    float r2z = clamp(r15.w * r14y, 0.0, 1.0);
    float r1w = clamp(r14y, 0.0, 1.0);
    vec2 r3zw = clamp(r15.xy, 0.0, 1.0);
    r1w = (1.0 - r3zw.x) * r3zw.y * r1w + r3zw.x;
    r1w = r1w * r1w;
    vec3 viewDirection = normalize(fragTangentViewPosition - fragTangentFragmentPosition);

    float r2w = max(dot(normal, viewDirection), 0.0001);
    lut_12_splatter = clamp(lut_12_splatter, 0.0, 1.0);
    float r12y = max(lut_12_splatter.w - (r2w * lut_12_splatter.w), 0.0);
    base_color.xyz = r12y * (lut_12_splatter.xyz - base_color.xyz) + base_color.xyz;

    subsurface_wrap = clamp(subsurface_wrap, 0.0, 1.0);
    subsurface_intensity = clamp(subsurface_intensity, 0.0, 1.0);

    r12y = 1.0 - 2.0 * r2w;
    r2w = subsurface_wrap * r12y + r2w;
    r12y = subsurface_wrap - 0.5;
    float r12z = abs(r12y) * 2.0 + 1.0;
    r2w = (exp2(log2(abs(r2w)) * r12z) - abs(r12y) + 0.5) * subsurface_intensity;
    if (r13w > 0.0) {
        r2w = 0.0;
    }

    snow_mix = clamp(snow_mix + material_wetness, 0.0, 1.0);
    if (r2w <= 0.01) {
        r2w = 0.0;
    }

    r12y = inv_dirt_weight;
    if (inv_dirt_weight <= 0.01) {
        r12y = 0.0;
    }

    // if (r2w > 0.0) {
        // subsurface_color = dither * (clamp(blood_gunk_modified_color, 0.0, 1.0) - base_color.xyz) + base_color.xyz;
        blood_gunk_modified_color = sRGBFromIntensity(r2w * (clamp(blood_gunk_modified_color, 0.0, 1.0) - base_color.xyz) + base_color.xyz);
    // }

    if (r12y > 0.0) {
        detail_normal_2_alignment = clamp(detail_normal_2_alignment, 0.0, 1.0) * 0.955 + 0.045;
        //subsurface_detail_intensity = dither * (detail_normal_2_alignment - detail_intensity) + detail_intensity;
        //material_detailing_normal = normalize(normalize(dbg_fragTBN[2]) + material_detailing_normal);
        //subsurface_detail_normal = dither * (normalize(material_detailing_normal) - normal) + normal;
        material_detailing_normal.xy = material_detailing_normal.xy * r12y;
        material_detailing_normal.z = reconstructNormalZ(material_detailing_normal.xy);
        normal = reorientNormalMaps(normal, material_detailing_normal);
    }

    float r2x = max((1.0 - r2w + r13w) - 1.0, 0.0);

    base_color.xyz = sRGBFromIntensity(clamp(base_color.xyz, 0.0, 1.0));

    normal = normalize(normal);

    // perform basic lighting calcs
    // fragColor = calculateLighting(base_color, roughness, detail_metallic, base_data_sample.z, normal_modified);
    if (debug == 0) {
        fragColor = calculateLighting(base_color, blood_gunk_modified_color, detail_intensity, /*detail_metallic*/ r2x, base_data_sample.z, normal);
    } else if (debug == 1) {
        fragColor = vec4(normal * 0.5 + 0.5, 1.0);
    } else if (debug == 2) {
        fragColor = vec4(vec3(detail_intensity), 1.0);
    } else if (debug == 3) {
        fragColor = vec4(vec3(detail_specular + r2w * (snow_mix - detail_specular)), 1.0);
    } else if (debug == 4) {
        fragColor = vec4(vec3(detail_metallic), 1.0);
    } else if (debug == 5) {
        fragColor = vec4(blood_gunk_modified_color, 1.0);
    } else if (debug == 6) {
        fragColor = vec4(vec3(weathering_dirt_weight), 1.0);
    } else if (debug == 7) {
        fragColor = vec4(vec3(r2x), 1.0);
    }
    // fragColor = vec4(vec3(detail_intensity), 1.0);
    //fragColor = vec4(normal_modified, 1.0);
}