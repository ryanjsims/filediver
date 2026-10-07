// note - this file is meant to be included in other shader source, not to be used standalone
// hence the missing #version directive

uniform samplerBuffer asset_grading_lut;

mat4 gradingMatrix(int groupId) {
    int idx = groupId * 4 - 4;
    if(idx < 0) {
        return mat4(1.0);
    }
    vec4 row0 = texelFetch(asset_grading_lut, idx);
    vec4 row1 = texelFetch(asset_grading_lut, idx + 1);
    vec4 row2 = texelFetch(asset_grading_lut, idx + 2);
    vec4 row3 = texelFetch(asset_grading_lut, idx + 3);
    return mat4(row0, row1, row2, row3);
}

vec3 gradeColor(vec3 color, mat4 matrix) {
    return (color.y * matrix[1].xyz) + (color.x * matrix[0].xyz) + (color.z * matrix[2].xyz) + matrix[3].xyz;
}

