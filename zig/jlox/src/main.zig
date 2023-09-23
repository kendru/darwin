const std = @import("std");
const builtin = @import("builtin");
const Scanner = @import("scanner.zig").Scanner;

const os = std.os;
const fs = std.fs;
const Allocator = std.mem.Allocator;

pub fn main() !void {
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    defer _ = gpa.deinit();
    const allocator = gpa.allocator();

    // Get command-line arguments.
    const args = try std.process.argsAlloc(allocator);
    defer std.process.argsFree(allocator, args);

    // Prints to stderr (it's a shortcut based on `std.io.getStdErr()`)
    std.debug.print("Arguments: {s}\n", .{args});

    switch (args.len) {
        1 => {
            try runPrompt(allocator);
        },
        2 => {
            try runFile(allocator, args[1]);
        },
        else => {
            std.debug.print("Usage: jlox [script]\n", .{});
            os.exit(64);
        },
    }

    // stdout is for the actual output of your application, for example if you
    // are implementing gzip, then only the compressed bytes should be sent to
    // stdout, not any debugging messages.
    //const stdout_file = std.io.getStdOut().writer();
    //var bw = std.io.bufferedWriter(stdout_file);
    //const stdout = bw.writer();

    //try stdout.print("Run `zig build test` to run the tests.\n", .{});

    //try bw.flush(); // don't forget to flush!
}

fn runPrompt(allocator: Allocator) !void {
    const stdout = std.io.getStdOut();
    const stdin = std.io.getStdIn();

    var buf: [1024]u8 = undefined;
    while (true) {
        try stdout.writeAll("> ");
        if (try readLine(stdin.reader(), &buf)) |input| {
            if (std.mem.eql(u8, input, "quit")) {
                break;
            }
            try run(allocator, input);
        } else {
            try stdout.writer().print("\n", .{});
            break;
        }
    }

    try stdout.writer().print("Goodbye.\n", .{});
}

fn runFile(allocator: Allocator, path: []const u8) !void {
    const file = try fs.cwd().openFile(
        path,
        .{ .mode = .read_only },
    );
    defer file.close();

    const data = try file.readToEndAlloc(allocator, 1024 * 1024 * 1024 * 2);
    defer allocator.free(data);

    try run(allocator, data);
}

fn run(allocator: Allocator, data: []const u8) !void {
    var scanner = Scanner{
        .source = data,
    };

    const tokens = try scanner.scanTokens(allocator);
    defer tokens.deinit();

    for (tokens.items, 0..) |token, i| {
        if (i > 0) {
            std.debug.print(" ", .{});
        }
        std.debug.print("{s}", .{token});
    }
    std.debug.print("\n", .{});
}

fn readLine(reader: fs.File.Reader, buffer: []u8) !?[]const u8 {
    var line = (try reader.readUntilDelimiterOrEof(
        buffer,
        '\n',
    )) orelse return null;

    // Trim carriage return on Windows.
    if (builtin.os.tag == .windows) {
        return std.mem.trimRight(u8, line, '\r');
    } else {
        return line;
    }
}

test "simple test" {
    var list = std.ArrayList(i32).init(std.testing.allocator);
    defer list.deinit(); // try commenting this out and see if zig detects the memory leak!
    try list.append(42);
    try std.testing.expectEqual(@as(i32, 42), list.pop());
}
