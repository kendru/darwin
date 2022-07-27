//! In Zig, each source file defines a struct, so it can have top-level struct
//! members, use @This() to get a comptime reference to the struct type, etc.

data: []const u8,

const ImplicitStruct = @This();

pub fn initialize (data: []const u8) ImplicitStruct {
  return ImplicitStruct {
    .data = data,
  };
}

pub fn length(self: ImplicitStruct) usize {
  return self.data.len;
}
