import std.mem;

s64 offset;
struct Animation {
    $ = offset;
    u32 size;
    offset = offset - size;
    $ = offset;
    u8 data[size];
    offset = offset - 4;
};

offset = std::mem::size() - 4;
Animation animation[while(offset > 0)] @0x00;
