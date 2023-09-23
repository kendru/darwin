const std = @import("std");

const Allocator = std.mem.Allocator;

pub const Scanner = struct {
    source: []const u8,

    const Self = @This();
    const TokenList = std.ArrayList(Token);

    pub fn scanTokens(self: Self, allocator: Allocator) !TokenList {
        var tokens = TokenList.init(allocator);
        for (self.source) |_| {
            try tokens.append(Token{});
        }
        return tokens;
    }
};

pub const Token = struct {
    pub fn format(
        self: Token,
        comptime fmt: []const u8,
        options: std.fmt.FormatOptions,
        writer: anytype,
    ) !void {
        _ = self;
        _ = fmt;
        _ = options;
        try writer.writeAll("TODO: print token");
    }
};
