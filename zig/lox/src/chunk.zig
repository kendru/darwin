const std = @import("std");
const DynamicArray = @import("internals.zig").DynamicArray;
const value = @import("value.zig");

const Allocator = std.mem.Allocator;
const Value = value.Value;
const ValueArray = value.ValueArray;

pub const OpCode = enum(u8) {
  op_constant,
  op_negate,
  op_add,
  op_subtract,
  op_multiply,
  op_divide,
  op_return,
  _,
};


pub const Chunk = struct {
  const Bytes = DynamicArray(u8);
  const LineNums = DynamicArray(usize);

  code: Bytes,
  constants: ValueArray,
  line_numbers: LineNums,

  pub fn init(alloc: Allocator) Chunk {
    return Chunk {
      .code = Bytes.init(alloc),
      .constants = ValueArray.init(alloc),
      .line_numbers = LineNums.init(alloc),
    };
  }

  pub fn deinit(self: Chunk) void {
    self.code.deinit();
    self.constants.deinit();
    self.line_numbers.deinit();
  }

  pub fn writeOpcode(self: *Chunk, op: OpCode, line: usize) Allocator.Error!void {
    try self.writeByte(@enumToInt(op), line);
  }

  pub fn writeByte(self: *Chunk, byte: u8, line: usize) Allocator.Error!void {
    try self.code.append(byte);
    try self.line_numbers.append(line);
  }

  pub fn addConstant(self: *Chunk, val: Value) Allocator.Error!usize {
    try self.constants.append(val);
    return self.constants.items.len - 1;
  }

  pub fn len(self: *Chunk) usize {
    return self.code.items.len;
  }
};

