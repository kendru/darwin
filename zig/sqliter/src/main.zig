const std = @import("std");

const MAX_INPUT_LINE_SIZE = 2048;

fn nextLine(reader: anytype, buffer: []u8) !?[]const u8 {
  var line = (try reader.readUntilDelimiterOrEof(
    buffer,
    '\n',
  )) orelse return null;

  if (std.builtin.os.tag == .windows) {
    line = std.mem.trimRight(u8, line, '\r');
  }

  return line;
}

pub fn main() !void {
  const stdin = std.io.getStdIn();
  const stdout = std.io.getStdOut();

  var bufIn = std.io.bufferedReader(stdin);
  const st = bufIn.reader();
  var lineBuf: [MAX_INPUT_LINE_SIZE]u8 = undefined;

  while (true) {
    try stdout.print("sqliter> ");
    // MAX_INPUT_LINE_SIZE
    var inputLine: ?[]u8 = st.readUntilDelimiterOrEof(&lineBuf, '\n', MAX_INPUT_LINE_SIZE) catch |err| {
      try stdout.print("Error reading line longer than {} chars\n", .{ MAX_INPUT_LINE_SIZE });
      return;
    };

    try stdout.print("You said: {}\n", .{ inputLine });
  }
}
