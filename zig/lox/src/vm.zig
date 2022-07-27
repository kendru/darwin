const std = @import("std");
const Allocator = std.mem.Allocator;
const print = std.debug.print;

const mod_chunk = @import("chunk.zig");
const Chunk = mod_chunk.Chunk;
const OpCode = mod_chunk.OpCode;
const mod_val = @import("value.zig");
const Value = mod_val.Value;
const mod_compiler = @import("compiler.zig");
const compile = mod_compiler.compile;
const printValue = mod_val.print;
const parse = @import("parser.zig").parse;
const dbg = @import("debug.zig");

const trace_exec = true;

const stack_max = 256;

pub const VM = struct {
  const Self = @This();

  alloc: Allocator,
  chunk: *Chunk = undefined,
  // Test performance of tracking the index within the bytecode
  // versus keeping a raw pointer that we increment. Can we
  // somehow guarantee that we do not index past the boundary
  // of the bytecode?
  ip: usize,
  stack: [stack_max]Value = undefined,
  stack_top: usize, // Test perf of offset vs ptr here too.

  pub fn init(alloc: Allocator) Self {
    return VM {
      .alloc = alloc,
      .ip = 0,
      .stack_top = 0,
    };
  }

  pub fn deinit(self: Self) void {
    _ = self;
  }

  pub fn interpret(self: *Self, source: []const u8) !InterpretResult {
    var chunk = Chunk.init(self.alloc);
    defer chunk.deinit();

    const ast = parse(self.alloc, source) catch {
      return .err_comptime;
    };

    if (!try compile(&ast, &chunk)) {
      return .err_comptime;
    }

    self.chunk = &chunk;
    self.ip = 0;

    return self.run();
  }

  pub fn run(self: *Self) InterpretResult {
    if (self.chunk.code.items.len == 0) return .ok;

    while (true) {
      if (comptime trace_exec) {
        print("          ", .{});
        var i: usize = 0;
        while (i < self.stack_top) : (i += 1) {
          print("[ ", .{});
          printValue(self.stack[i]);
          print(" ]", .{});
        }
        print("\n", .{});

        _ = dbg.disassembleInstruction(self.chunk, self.ip);
      }

      switch (self.readOpcode()) {
          .op_constant => {
            const constant = self.readConstant();
            self.push(constant);
          },
          .op_negate => {
            const x = self.pop();
            self.push(-x);
          },
          .op_add => self.binaryOp(.add),
          .op_subtract => self.binaryOp(.subtract),
          .op_multiply => self.binaryOp(.multiply),
          .op_divide => self.binaryOp(.divide),
          .op_return => {
            printValue(self.pop());
            print("\n", .{});
            return .ok;
          },
          _ => return .err_runtime
      }
    }
  }

  fn binaryOp(self: *Self, op: BinOp) void {
    const a = self.pop();
    const b = self.pop();
    self.push(switch (op) {
      .add => a + b,
      .subtract => a - b,
      .multiply => a * b,
      .divide => a / b,
    });
  }

  fn readOpcode(self: *Self) OpCode {
    return @intToEnum(OpCode, self.readByte());
  }

  inline fn readConstant(self: *Self) Value {
    return self.chunk.constants.items[self.readByte()];
  }

  inline fn readByte(self: *Self) u8 {
    const res = self.chunk.code.items[self.ip];
    self.ip += 1;
    return res;
  }

  fn push(self: *Self, value: Value) void {
    self.stack[self.stack_top] = value;
    self.stack_top += 1;
  }

  fn pop(self: *Self) Value {
    self.stack_top -= 1;
    return self.stack[self.stack_top];
  }

  fn resetStack(self: *Self) void {
    self.stack_top = 0;
  }

  const BinOp = enum {
      add,
      subtract,
      multiply,
      divide,
  };
};

const InterpretResult = enum(u8) {
  ok,
  err_comptime,
  err_runtime,
};
