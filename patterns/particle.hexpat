import std.core;

// import stingray; // Not sure where to put custom library files yet =/
// stingray.hexpat
import std.io;

using ThinMurmurHash;
using MurmurHash;

fn formatThinHash(ThinMurmurHash hash) {
    return std::format("{:08X}", hash.value);
};

fn formatHash(MurmurHash hash) {
    return std::format("{:016X}", hash.value);
};

struct ThinMurmurHash {
    u32 value;
} [[format("formatThinHash")]];

struct MurmurHash {
    u64 value;
} [[format("formatHash")]];
// end stingray.hexpat

#pragma pattern_limit 67108864
#pragma array_limit 524288
#pragma loop_limit 65536

fn relative_to_parents_parent(u128 offset) {
    return addressof(parent.parent.parent);
};

fn relative_to_parent(u128 offset) {
    return addressof(parent.parent);
};

u32 magic = 0;

struct Pair<T1, T2> {
    T1 val1;
    T2 val2;
};

struct Vector3 {
    float x, y, z;
};

struct Vector4 {
    float x, y, z, w;
};

struct Matrix4 {
    Vector4 rows[4];
};

enum GraphType : u32 {
    XY = 4,
    Color = 5,
};

struct XYGraph {
    float xValues[10];
    float yValues[10];
};

struct GradientGraph {
    float xValues[10];
    Vector3 yValues[10];
};

enum ControllerFormat : u32 {
    Unk0,
    MinMax,
    VelocityCone,
    Unk3,
    Unk4,
    Unk5,
    Unk6,
    Unk7,
    VelocityBox,
    Unk9,
    UnkA,
    UnkB,
    VelocityCylinder,
};

enum ControllerType : u32 {
    Luminance = 40,
    Life = 36,
};

struct Component {
    ControllerFormat format;
    ControllerType type;
    match(format) {
        (ControllerFormat::MinMax): {
            float min, max;
        }
        (ControllerFormat::VelocityCone): {
            float minVelocity, maxVelocity;
            float minTheta, maxTheta;
            Vector3 axis;
        }
        (ControllerFormat::Unk3): {
            u32 unkInts[2];
        }
        (ControllerFormat::Unk4): {
            Vector3 unkVectors[2];
        }
        (ControllerFormat::Unk6): {
            u32 unk1;
        }
        (ControllerFormat::Unk7): {
            u32 unkInts[6];
        }
        (ControllerFormat::VelocityBox): {
            Vector3 min;
            Vector3 max;
        }
        (ControllerFormat::VelocityCylinder): {
            u32 unkInt2;
            float minRadial, maxRadial;
            float minZ, maxZ;
            float minAngular, maxAngular;
        }
        (_): std::print("Unknown controller format {} at index {}\n", format, std::core::array_index());
    }
};

enum EmitterType : u32 {
    Unk0,
    Rate = 11,
    Burst,
};

struct ZeroEmitter {
    u32 unk1, unk2;
    if(magic == 0x73) {
        u32 unk3, unk4, unk5;
    }
};

struct BurstEmitter {
    Pair<float, u32> timeAmounts[10];
};

struct RateEmitter {
    float min, max;
    XYGraph changeOverTime;
    if(magic == 0x73) {
        XYGraph unkGraph;
        u32 unknown[4];
        padding[4];
    }
};

struct Emitter {
    EmitterType type;
    match(type) {
        (EmitterType::Unk0): ZeroEmitter emitter;
        (EmitterType::Rate): RateEmitter emitter;
        (EmitterType::Burst): BurstEmitter emitter;
        (_): std::print("Unknown emitter type {}", type);
    }
};

enum VisualizerType : u32 {
    Billboard,
    Light,
    Mesh,
    Unk3,
    Trail,
};

struct BillboardVisualizer {
    u32 unk0, unk1;
    MurmurHash material;
};

struct MeshVisualizer {
    MurmurHash unit;
    ThinMurmurHash mesh;
    padding[4];
    MurmurHash material;
};

struct TrailVisualizer {
    MurmurHash material;
};

struct Visualizer {
    VisualizerType type;
    match(type) {
        (VisualizerType::Billboard): BillboardVisualizer visualizer;
        (VisualizerType::Mesh): MeshVisualizer visualizer;
        (VisualizerType::Trail): TrailVisualizer visualizer;
        (_): std::print("Unimplemented visualizer type {}", type);
    }
    u32 unkValues[51];
    u32 unkint1;
    u32 unkCount;
    u32 unkOffset;
    u32 unkOffset2;
    u32 unkCount3;
    u32 unkOffset3;
    u32 totalSize;
    padding[totalSize - sizeof(this)];
};

struct ParticleSystem {
    u64 base = $;
    u32 maxParticleCount;
    u32 numComponents;
    u32 unkInt;
    if(magic == 0x73) {
        u32 componentBitFlags[16];
    } else {
        u32 componentBitFlags[8];
    }
    s32 unkInt2;
    s32 unkInt3;

    s32 unkInt4; // 4/5 might just be padding
    s32 unkInt5;

    u32 unkInt6;
    s32 unkInt7;
    u32 unkInt8;
    s32 unkInt9;
    ThinMurmurHash type1;
    ThinMurmurHash type2;
    padding[4];
    Matrix4 transform;
    if (magic == 0x73) {
        float unknownFloats[11];
    } else {
        float unknownFloats[3];
    }

    u32 componentListCount;
    u32 componentListOffset;
    if (componentListCount > 0) {
        Component components[componentListCount] @ base + componentListOffset;
    }

    u32 emitterCount;
    u32 emitterOffset;
    if (emitterCount > 0) {
        Emitter emitters[emitterCount] @ base + emitterOffset;
    }

    u32 unknown;
    u32 visualizerCount;
    u32 visualizerOffset;
    if (visualizerCount > 0) {
        Visualizer visualizers[visualizerCount] @ base + visualizerOffset;
    }

    u32 totalSize;
    if(magic == 0x73) {
        u32 unk2;
    }
    padding[totalSize - sizeof(this)];
};

struct Particle {
    u32 _magic;
    magic = _magic;
    float minLifetime;
    float maxLifetime;
    float unkFloat;
    u32 unkInt1;
    u32 variableCount;
    u32 particleSystemCount;
    if(magic == 0x31) {
        ThinMurmurHash variableNames[variableCount];
        Vector3 variableValues[variableCount];
    }
    if(magic == 0x73) {
        u32 unkItemCount;
        padding[48];
        ThinMurmurHash variableNames[variableCount];
        Vector3 variableValues[variableCount];
    }
    ParticleSystem systems[particleSystemCount];
};

Particle particle @ 0x0;
