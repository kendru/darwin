const std = @import("std");

const stdout = std.io.getStdOut().writer();

pub fn main() anyerror!void {
    try stdout.print("Hello, World!\n", .{});
}

// This is required so that running test against this file tests all code
// in dependent files.
// See https://ziglang.org/documentation/master/#Nested-Container-Tests
test {
    std.testing.refAllDecls(@This());
    _ = @import("learn-by-test.zig");
}


