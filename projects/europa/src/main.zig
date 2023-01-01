const std = @import("std");
const testing = std.testing;
const allocator = std.heap.page_allocator;

extern "env" fn log(ptr: u32, len: usize) void;
pub fn _log(msg: []u8) void {
    log(@truncate(u32, stringToPtr(msg)), msg.len);
}

pub fn stringToPtr(s: []const u8) u64 {
    const p: u64 = @ptrToInt(s.ptr);
    return p << 32;
}

export fn add(a: i32, b: i32) i32 {
  // var buffer: [1028]u8 = undefined;
  // var fba = std.heap.FixedBufferAllocator.init(&buffer);
  // const allocator1 = fba.allocator();
  // const string = std.fmt.allocPrint(
  //     allocator,
  //     "{d} + {d} = {d}",
  //     .{ a, b, a+b },
  // ) catch unreachable;
  // defer allocator.free(string);
  // const bs = allocator1.alloc(u8, 8) catch label: {
  //   var tmp: [10]u8 = undefined;
  //   break :label &tmp;
  // };
  var buf: [100]u8 = undefined;
  var buf_slice: [:0]u8 = std.fmt.bufPrintZ(&buf, "{d} + {d} = {d}", .{a, b, a+b}) catch unreachable;
  _log(buf_slice);
  return a + b;
}

test "basic add functionality" {
    try testing.expect(add(3, 7) == 10);
}
