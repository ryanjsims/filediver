#include <std/ptr.pat>

fn relative_to_parents_parent(u128 offset) {
    return addressof(parent.parent.parent);
};

fn relative_to_parent(u128 offset) {
    return addressof(parent.parent);
};

struct ThinMurmurHash {
    u32 value;
};

struct Name {
    char data[];
};

struct Bones {
    u32 boneCount;
    u32 parentCount;
    float floats[parentCount];
    ThinMurmurHash nameHashes[boneCount];
    u32 roots[parentCount];
    Name names[boneCount];
};

Bones b @0x0;