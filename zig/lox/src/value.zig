const std = @import("std");
const DynamicArray = @import("internals.zig").DynamicArray;

pub const Value = f64;
pub const ValueArray = DynamicArray(Value);

pub fn print(v: Value) void {
  std.debug.print("{d:.4}", .{v});
}
