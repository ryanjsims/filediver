#include <std/ptr.pat>
#include <std/io.pat>
#include <std/core.pat>

#pragma pattern_limit 8388608
#pragma array_limit 524288

fn relative_to_parents_parent(u128 offset) {
    return addressof(parent.parent.parent);
};

fn relative_to_parent(u128 offset) {
    return addressof(parent.parent);
};

struct vec3 {
    float data[3] [[inline]];
};

struct vec4 {
    float data[4] [[inline]];
};

struct Mat4x4 {
    vec4 data[4] [[inline]];
};

struct maxMin {
    float max, min;
};

struct MurmurHash {
    u64 value;
};

struct ThinMurmurHash {
    u32 value;
};

struct LODGroupHeader {
    u32 unk00;
    ThinMurmurHash unkHash00;
    u32 unk01, unk02;
};

struct LODEntry {
    maxMin detail;
    u32 count;
    u32 indices[count];
};

struct LODEntryPtr {
    LODEntry *ptr : u32 [[pointer_base("relative_to_parents_parent"), inline]];
};

struct AABB {
    float x0, y0, z0;
    float x1, y1, z1;
};

struct LODGroupFooter {
    AABB boundingBox;
    float unkFloat0;
    u32 unk00;
    float unkFloat1;
    u32 unk01;
};

struct LODGroup {
    LODGroupHeader header;
    u32 entryCount;
    LODEntryPtr entryOffsets[entryCount];
    LODGroupFooter *footer : u32 [[pointer_base("relative_to_parent")]];
};

struct LODGroupPtr {
    LODGroup *ptr : u32 [[pointer_base("relative_to_parents_parent"), inline]];
};

struct LODGroupPtrList {
    u32 count;
    LODGroupPtr groups[count];
};

struct JointListHeader {
    u32 NumJoints;
    u8 unk00[12];
};

struct JointTransform {
    vec3 Rotation[3] [[color("00ff00")]];
    float Translation[3] [[color("0000ff")]];
    float Scale[3] [[color("ff0000")]];
    float Skew [[color("ffff00")]];
};

struct JointMapEntry {
    u16 increment;
    u16 parentBone;
};

struct JointList {
    JointListHeader header;
    JointTransform transforms[header.NumJoints];
    Mat4x4 transformMatrices[header.NumJoints];
    JointMapEntry jointMap[header.NumJoints];
    ThinMurmurHash nameHashes[header.NumJoints];
};

struct MeshGroup {
    u32 StartGroupId;
    u32 VertexOffset;
    u32 NumVertices;
    u32 IndexOffset;
    u32 NumIndices;
    u32 EndGroupId; // Group IDs should be identical
};

struct MeshGroupList {
    MeshGroup Groups[parent.NumGroups];
};

struct MeshHeader {
    u32 unk00[2];
    AABB bounding_box;
    float unkFloat;
    u32 unkInt0;
    ThinMurmurHash unkHash;
    u32 unkInts[4];
    s32 LayoutIdx;
    u8 unk01[40];
    u32 NumMaterials;
    u8 *MaterialOffset : u32 [[pointer_base("relative_to_parent")]];
    u8 unk02[8];
    u32 NumGroups;
    MeshGroupList *Groups : u32 [[pointer_base("relative_to_parent")]];
};

enum MeshLayoutItemType : u32 {
    Position,
    Color,
    UVCoords = 4,
    BoneIdx = 6,
    BoneWeight
};

enum MeshLayoutItemFormat : u32 {
    FormatF32,
    FormatVec2F,
    FormatVec3F,
    FormatVec4F,
    FormatVec4S8 = 24,
    FormatVec4U8 = 26,
    FormatF16 = 28,
    FormatVec2F16,
    FormatVec3F16,
    FormatVec4F16
};

fn formatSize(MeshLayoutItemFormat format) {
    match(format) {
        (MeshLayoutItemFormat::FormatF32): return 4;
        (MeshLayoutItemFormat::FormatVec2F): return 8;
        (MeshLayoutItemFormat::FormatVec3F): return 12;
        (MeshLayoutItemFormat::FormatVec4F): return 16;
        (MeshLayoutItemFormat::FormatVec4Norm8): return 4;
        (MeshLayoutItemFormat::FormatVec4U8): return 4;
        (MeshLayoutItemFormat::FormatF16): return 2;
        (MeshLayoutItemFormat::FormatVec2F16): return 4;
        (MeshLayoutItemFormat::FormatVec3F16): return 6;
        (MeshLayoutItemFormat::FormatVec4F16): return 8;
        (_): std::error("Unknown format {}", format);
    }
    return 0;
};

struct MeshLayoutItem {
    MeshLayoutItemType Type;
    MeshLayoutItemFormat Format;
    u32 Layer;
    u8 unk00[8];
};

struct MeshLayout {
    u8 magic0[4];
    u8 unk00[4];
    MeshLayoutItem Items[16];
    u32 NumItems;
    u8 unk01[4];
    u8 magic1[4];
    u8 unk02[12];
    u32 NumVertices, VertexStride;
    u8 unk03[16];
    u8 magic2[4];
    u8 unk04[12];
    u32 NumIndices;
    u8 unk05[20];
    u32 VertexOffset, PositionsSize, IndexOffset, IndicesSize;
    u8 unk06[16];
};

fn layoutStride(MeshLayout layout) {
    u32 toRet = 0;
    for(u32 i = 0, i < layout.NumItems, i = i + 1) {
        toRet += formatSize(layout.Items[i].Format);
    }
    return toRet;
};

struct MeshLayoutPtr {
    MeshLayout *data : u32 [[pointer_base("relative_to_parents_parent"), inline]];
};

struct MeshLayoutList {
    u32 count;
    MeshLayoutPtr layouts[count];
    ThinMurmurHash hashes[count];
};

struct MeshInfo {
    MeshHeader header;
    ThinMurmurHash materials[header.NumMaterials];
};

struct MeshInfoPtr {
    MeshInfo *ptr : u32 [[pointer_base("relative_to_parents_parent"), inline]];
};

struct MeshInfoList {
    u32 MeshInfoCount;
    MeshInfoPtr MeshInfos[MeshInfoCount];
    // Not sure if these are actually hashes or not
    ThinMurmurHash unkHashes[MeshInfoCount];
};

struct SkeletonMapJointIndices {
    u32 data[parent.count];
};

struct SkeletonMapRemapIndices {
    u32 unk00;
    u32 unk01;
    u32 count;
    u32 data[count];
};

struct SkeletonMapItem {
    u32 count;
    u32 unk00;
    SkeletonMapJointIndices *unitIndices : u32 [[pointer_base("relative_to_parent")]];
    SkeletonMapRemapIndices *remapIndices : u32 [[pointer_base("relative_to_parent")]];
    Mat4x4 matrices[count];
};

struct SkeletonMapItemPtr {
    SkeletonMapItem *data : u32 [[inline, pointer_base("relative_to_parents_parent")]];
};

struct SkeletonMapList {
    u32 count;
    SkeletonMapItemPtr offsets[count];
};

struct MaterialList {
    u32 count;
    ThinMurmurHash keys[count];
    MurmurHash values[count];
};

std::mem::Section meshWorkspace = std::mem::create_section("Mesh Workspace");
MeshInfo currentInfo @0x00 in meshWorkspace;
MeshLayout currentLayout @sizeof(currentInfo) in meshWorkspace;
MeshGroup Groups[32] @sizeof(currentInfo) + sizeof(currentLayout) in meshWorkspace;
MeshGroup currentGroup @sizeof(currentInfo) + sizeof(currentLayout) + sizeof(Groups) in meshWorkspace;

struct Vertex<auto Stride> {
    u8 data[Stride];
};

struct VertexGroup<auto Stride> {
    std::mem::copy_value_to_section(Groups[std::core::array_index()], meshWorkspace, addressof(currentGroup));
    $ = addressof(parent.parent) + currentLayout.VertexOffset + currentGroup.VertexOffset * currentLayout.VertexStride + 12;
    Vertex<Stride> vertices[currentGroup.NumVertices];
};

struct IndexGroup<auto Stride> {
    std::mem::copy_value_to_section(Groups[std::core::array_index()], meshWorkspace, addressof(currentGroup));
    std::print("Base: {:x} Layout offset: {:x} Group offset: {} Stride: {}", addressof(parent.parent), currentLayout.IndexOffset, currentGroup.IndexOffset, Stride);
    $ = addressof(parent.parent) + currentLayout.IndexOffset + currentGroup.IndexOffset * Stride + 12;
    match(Stride) {
        (1): u8 Indices[currentGroup.NumIndices];
        (2): u16 Indices[currentGroup.NumIndices];
        (4): u32 Indices[currentGroup.NumIndices];
    }
};

fn copyGroupsToSection() {
    for(u32 i = 0, i < currentInfo.header.NumGroups, i = i + 1) {
        std::mem::copy_value_to_section(parent.parent.parent.MeshInfoListOffset.MeshInfos[std::core::array_index()].ptr.header.Groups.Groups[i], meshWorkspace, addressof(Groups) + sizeof(currentGroup) * i);
    }
};

struct MeshData {
    std::mem::copy_value_to_section(parent.parent.MeshInfoListOffset.MeshInfos[std::core::array_index()].ptr, meshWorkspace, addressof(currentInfo));
    std::mem::copy_value_to_section(parent.parent.MeshLayoutListOffset.layouts[currentInfo.header.LayoutIdx].data, meshWorkspace, addressof(currentLayout));
    copyGroupsToSection();
    std::print("Stride: {}", currentLayout.VertexStride);
    std::print("Groups: {}", currentInfo.header.NumGroups);
    VertexGroup<currentLayout.VertexStride> vertexGroups[currentInfo.header.NumGroups];
    IndexGroup<currentLayout.IndicesSize / currentLayout.NumIndices> indexGroups[currentInfo.header.NumGroups];
};

struct MeshDataList {
    u32 count;
    MeshData meshes[count];
};

struct SoundCallbackData {
    u32 count;
    u8 *unkOffset00 : u32 [[pointer_base("relative_to_parent")]];
    u8 *unkOffset01 : u32 [[pointer_base("relative_to_parent")]];
    u8 *unkOffset02 : u32 [[pointer_base("relative_to_parent")]];
};

struct UnknownMatrixItem {
    ThinMurmurHash unkHash00;
    u32 unk00;
    Mat4x4 matrix;
    u8 unkData[40];
};

struct UnknownMatricesList {
    u32 count;
    u32 unk00[3];
    UnknownMatrixItem matrices[count];
};

struct Header {
    u8 unk00[8];
    MurmurHash bones;
    u8 unk01[8];
    MurmurHash unkHash00;
    MurmurHash stateMachine;
    u8 unk02[8];
    LODGroupPtrList *LODGroupListOffset : u32;
    JointList *JointListOffset : u32;
    UnknownMatricesList *UnkOffset01 : u32;
    u8 *UnkOffset02 : u32;
    SoundCallbackData *UnkOffset03 : u32;
    u8 unk03[8];
    u8 *UnkOffset04 : u32;
    u8 *UnkOffset05 : u32;
    u8 unk04[4];
    SkeletonMapList *skeletonMaps : u32;
    MeshLayoutList *MeshLayoutListOffset: u32;
    u64 meshDataListAddress = $;
    $ += 4;
    MeshInfoList *MeshInfoListOffset: u32;
    u8 unk05[8];
    MaterialList *MaterialListOffset : u32;
    $ = meshDataListAddress;
    MeshDataList *MeshDataOffset : u32;
};

Header hdr @0x0;