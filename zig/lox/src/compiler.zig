const std = @import("std");
const Ast = @import("ast.zig");
const Node = Ast.Node;
const Value = @import("value.zig").Value;
const chunk_mod = @import("chunk.zig");
const Chunk = chunk_mod.Chunk;
const OpCode = chunk_mod.OpCode;

const Allocator = std.mem.Allocator;

pub fn compile(ast: *const Ast, chunk: *Chunk) Allocator.Error!bool {
  var compiler = Compiler {
    .ast = ast,
    .base_chunk = chunk,
  };

  try compiler.render(0);
  try compiler.end();

  return true;
}

pub const Compiler = struct {
  ast: *const Ast,
  base_chunk: *Chunk,

  fn render(self: *Compiler, node_idx: u32) Allocator.Error!void {
    const tags = self.ast.nodes.items(.tag);
    const tok_idxs = self.ast.nodes.items(.token);
    const tag = tags[@as(usize, node_idx)];
    const tokIdx = tok_idxs[@as(usize, node_idx)];
    std.debug.print("Tag: {s}\n", .{@tagName(tag)});
    switch (tag) {
        .lit_integer => {
          // std.debug.print("Got number: {s}", .{self.ast.lexeme(tokIdx)});
          const n = std.fmt.parseFloat(f64, self.ast.lexeme(tokIdx)) catch {
            @panic("Should have only lexed numbers that can be parsed");
          };
          const c = try self.makeConstant(n);
          try self.emitSimpleOp(.op_constant, @as(u8, c));
        },
        .negate => {
          std.debug.print("Here 1", .{});
          const data = self.ast.data(Node.UnaryOpData, node_idx);
          std.debug.print("Here 2", .{});
          try self.render(data.operand);
          std.debug.print("Here 3", .{});
          try self.emitOpcode(.op_negate);
          std.debug.print("Here 4", .{});
        },
        else => unreachable,
    }
  }

  fn makeConstant(self: *Compiler, value: Value) Allocator.Error!u8 {
    const idx = try self.currentChunk().addConstant(value);
    if (idx > std.math.maxInt(u8)) {
      err("Too many constants in one chunk");
      return 0;
    }
    return @intCast(u8, idx);
  }

  fn emitReturn(self: *Compiler) Allocator.Error!void {
    try self.emitOpcode(.op_return);
  }

  fn end(self: *Compiler) Allocator.Error!void {
    // XXX: Temporarily add a return so that the VM will return the value
    // of whatever expression we generate. We do not yet have statements.
    try self.emitReturn();
  }

  fn emitSimpleOp(self: *Compiler, op: OpCode, operand: u8) Allocator.Error!void {
    try self.emitOpcode(op);
    try self.emitByte(operand);
  }

  fn emitOpcode(self: *Compiler, op: OpCode) Allocator.Error!void {
    return self.currentChunk().writeOpcode(op, 0); // TODO: Get the current node's line.
  }

  fn emitByte(self: *Compiler, byte: u8) Allocator.Error!void {
    return self.currentChunk().writeByte(byte, 0); // TODO: Get the current node's line.
  }

  fn currentChunk(self: *Compiler) *Chunk {
    return self.base_chunk;
  }

  fn err(msg: []const u8) void {
    _ = msg;
    // TODO: include more context; determine whether we should crash.
    // std.debug.print("ERROR: {s}", .{msg}) catch {
    //   @panic("Cannot print to stderr.");
    // };
  }
};

