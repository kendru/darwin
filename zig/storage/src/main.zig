const std = @import("std");
const lex = @import("lex.zig");
const Parser = @import("parse.zig").Parser;
const query = @import("query.zig");
const QueryPrinter = @import("print.zig").QueryPrinter;

const Lexer = lex.Lexer;
const Token = lex.Token;

pub fn main() !void {
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    defer _ = gpa.deinit();
    const alloc = gpa.allocator();

    try startRepl(alloc);
}

fn startRepl(alloc: std.mem.Allocator) !void {
    // Start a repl.
    const stdin = std.io.getStdIn();
    while (true) {
        std.debug.print("db> ", .{});
        // Read up to 4k bytes from stdin until a newline is encountered,
        // allocating a new buffer for the input
        const input = try stdin.reader().readUntilDelimiterAlloc(alloc, '\n', 1024*4);
        defer alloc.free(input);

        if (input.len == 0) {
            // If the user entered an empty line, continue to the next iteration.
            continue;
        }

        if (input[0] == '.') {
            var si = std.mem.split(u8, input[1..], " ");
            if (si.next()) |cmd| {
                const should_quit = try processCommand(alloc, cmd, si.rest());
                if (should_quit) {
                    break;
                }
            }
        }


        var arena = std.heap.ArenaAllocator.init(std.heap.page_allocator);
        defer arena.deinit();

        var lexer = Lexer.init(input, alloc);
        var parser = Parser.init(arena.allocator(), lexer);
        defer parser.deinit();

        var ast = try parser.parseQuery();

        // Print the AST to stdout.
        const stderr_writer = std.io.getStdErr().writer();
        var printer = QueryPrinter(@TypeOf(stderr_writer)).init(stderr_writer);
        printer.print(ast);

        // const out = try query.parse(alloc, input);
        // defer out.deinit();

        // Otherwise, print the input.
        // std.debug.print("Unrecognized input: \"{s}\".\nPlease type \".help\" for help.\n", .{input});
    }
}

fn processCommand(alloc: std.mem.Allocator, cmd: []const u8, args: []const u8) !bool {
    _ = args;
    _ = alloc;
    if (stringEq(cmd, "quit")) {
        return true;
    }

    return false;
}

fn stringEq(a: []const u8, b: []const u8) bool {
    if (a.len != b.len) return false;
    return std.mem.eql(u8, a, b);
}

test "simple test" {
    var list = std.ArrayList(i32).init(std.testing.allocator);
    defer list.deinit(); // try commenting this out and see if zig detects the memory leak!
    try list.append(42);
    try std.testing.expectEqual(@as(i32, 42), list.pop());
}
