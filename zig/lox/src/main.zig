const std = @import("std");
const VM = @import("vm.zig").VM;
const dbg = @import("debug.zig");
const Allocator = std.mem.Allocator;

const stderr = std.io.getStdErr().writer();
const stdout = std.io.getStdOut().writer();
const stdin = std.io.getStdIn().reader();

const inputFileMaxSize: usize = 2*1024*1042*1024;

pub fn main() anyerror!void {
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    var alloc = gpa.allocator();
    defer _ = gpa.deinit();

    var vm = VM.init(alloc);
    vm.deinit();

    const args = try std.process.argsAlloc(alloc);
    defer std.process.argsFree(alloc, args);

    switch (args.len) {
        1 => startRepl(&vm),
        2 => runScript(alloc, &vm, args[1]),
        else => {
          try stderr.print("Usage: lox [path]\n", .{});
          std.process.exit(64);
        },
    }
}

fn startRepl(vm: *VM) void {
  var buf = std.io.bufferedReader(stdin);
  var reader = buf.reader();
  var line_buf: [1024]u8 = undefined;

  while (true) {
    stdout.writeAll("> ") catch std.debug.panic("Could not write to stdout", .{});
    var line = reader.readUntilDelimiterOrEof(line_buf[0..], '\n') catch {
      std.debug.panic("Could not read from stdin", .{});
    } orelse {
      stdout.writeAll("\n") catch std.debug.panic("Could not write to stdout", .{});
      break;
    };

    // TODO: Provide feedback on results.
    _ = vm.interpret(line) catch {
      std.debug.panic("Could not interpret line", .{});
    };
  }
}

fn runScript(alloc: Allocator, vm: *VM, filename: []u8) void {
  const source = readFile(alloc, filename);
  switch (vm.interpret(source) catch {
    std.debug.panic("Could not interpret script", .{});
  }) {
    .ok => {},
    .err_comptime => std.process.exit(65),
    .err_runtime => std.process.exit(70),
  }
}

fn readFile(alloc: Allocator, filename: []u8) []u8 {
  const fd  = std.fs.cwd().openFile(filename, .{
    .read = true,
  }) catch |err| {
      stderr.print("Could not open file \"{s}\", error: {any}.\n", .{ filename, err }) catch {};
      return std.process.exit(74);
  };
  defer fd.close();

  return fd.readToEndAlloc(alloc, inputFileMaxSize) catch |err| {
      stderr.print("Could not read file \"{s}\", error: {any}.\n", .{ filename, err }) catch {};
      return std.process.exit(74);
  };
}

