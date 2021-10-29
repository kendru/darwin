const std = @import("std");
const page = @import("page.zig");

const native_endian = @import("builtin").target.cpu.arch.endian();
const expect = std.testing.expect;

const Gpa = std.heap.GeneralPurposeAllocator(.{});

pub fn main() !void {
    // align(8) is required to avoid unaligned reads of the fields of Header. 
    var data: [12]u8 align(8) = [12]u8{
        0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // next
        0x00, 0x00,                                     // freeStart
        0xbe, 0xef,                                     // freeLen
    };
    
    // Create a pointer to data as a *page.Header.
    var h = @ptrCast(*page.Header, &data);
    // The fields should be initialized as expected from the binary `data`.
    try expect(h.next == null);
    try expect(h.freeStart == 0);
    switch (native_endian) {
        .Big => {
            try expect(h.freeLen == 0xbeef);
        },
        .Little => {
            try expect(h.freeLen == 0xefbe);
        },
    }

    // Set free start.
    const startOffset = @bitOffsetOf(page.Header, "freeStart") / 8;
    var freeStart = @intToPtr(*u16, @ptrToInt(&data) + startOffset);
    freeStart.* = 12;
    try expect(h.freeStart == 12);   
}
