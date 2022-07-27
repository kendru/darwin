const std = @import("std");
const chunk = @import("chunk.zig");
const printValue = @import("value.zig").print;

const print = std.debug.print;
const Chunk = chunk.Chunk;
const OpCode = chunk.OpCode;

pub fn disassembleChunk(c: *Chunk, name: []const u8) void {
  print("== {s} ==\n", .{name});
  var offset: usize = 0;
  while (offset < c.len()) {
    offset = disassembleInstruction(c, offset);
  }
}

pub fn disassembleInstruction(c: *Chunk, offset: usize) usize {
  print("{d:0>4} ", .{offset});
  if (offset == 0) {
    print("   | EOF", .{});
    return offset;
  }

  if (offset > 0 and c.line_numbers.items[offset] == c.line_numbers.items[offset-1]) {
    print("   | ", .{});
  } else {
    print("{d:>4} ", .{c.line_numbers.items[offset]});
  }

  const op = @intToEnum(OpCode, try c.code.get(offset));
  return switch (op) {
      .op_constant => constantInstruction("OP_CONSTANT", c, offset),
      .op_negate => simpleInstruction("OP_NEGAGTE", offset),
      .op_add => simpleInstruction("OP_ADD", offset),
      .op_subtract => simpleInstruction("OP_SUBTRACT", offset),
      .op_multiply => simpleInstruction("OP_MULTIPLY", offset),
      .op_divide => simpleInstruction("OP_DIVIDE", offset),
      .op_return => simpleInstruction("OP_RETURN", offset),
      _ => ret: {
        print("Unknown opcode {d}", .{op});
        break :ret offset + 1;
      },
  };
}

fn simpleInstruction(name: []const u8, offset: usize) usize {
  print("{s}\n", .{name});

  return offset + 1;
}

fn constantInstruction(name: []const u8, c: *Chunk, offset: usize) usize {
  const const_idx = c.code.items[offset+1];
  print("{s:<16} {d:>4} '", .{name, const_idx});
  printValue(c.constants.items[const_idx]);
  print("'\n", .{});

  return offset + 2;
}
