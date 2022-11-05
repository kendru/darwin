const std = @import("std");
const page = @import("page.zig");

const native_endian = @import("builtin").target.cpu.arch.endian();
const expect = std.testing.expect;

pub fn main() !void {
    var general_purpose_allocator = std.heap.GeneralPurposeAllocator(.{}){};
    defer std.debug.assert(!general_purpose_allocator.deinit());

    const gpa = general_purpose_allocator.allocator();

    // align(8) is required to avoid unaligned reads of the fields of Header.
    var data: [12]u8 align(8) = [12]u8{
        0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // next
        0x00, 0x00,                                     // freeStart
        0xbe, 0xef,                                     // freeLen
    };

    var data1 = try gpa.alloc(u8, 1024*8);
    defer gpa.free(data1);

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
