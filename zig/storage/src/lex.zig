// lex.zig
//
// This file contains the lexical analyzer for our minimal SQL dialect.
const std = @import("std");
const util = @import("util.zig");

const Allocator = std.mem.Allocator;
const String = util.String;

const ErrorList = std.ArrayList(String);

const TokenTableEntry = struct {
    char: u8,
    kind: Token.Kind,
};
const singleCharTokens = [_]TokenTableEntry{
    .{ .char = ',', .kind = .punc_comma },
    .{ .char = '(', .kind = .punc_paren_left },
    .{ .char = ')', .kind = .punc_paren_right },
    .{ .char = ';', .kind = .punc_semicolon },
    .{ .char = '*', .kind = .punc_star },
    .{ .char = '=', .kind = .op_eq },
    .{ .char = '+', .kind = .op_plus },
    .{ .char = '-', .kind = .op_minus },
    .{ .char = '/', .kind = .op_divide },
    .{ .char = '%', .kind = .op_modulo },
    .{ .char = '.', .kind = .op_dot },
};

fn tokenKindForChar(c: u8) ?Token.Kind {
    inline for (singleCharTokens) |entry| {
        if (entry.char == c) {
            return entry.kind;
        }
    }
    return null;
}

// Lexer is an iterator that produces tokens from a source string.
pub const Lexer = struct {
    alloc: Allocator,
    src: String,
    i: u32 = 0,
    errors: ErrorList,

    // TODO: Be able to save and restore the lexer state.
    // This would be useful to call from the parser when we
    // need to backtrack or when we want to walk a list to
    // determine the size, allocate, and parse.

    pub fn init(src: String, alloc: Allocator) Lexer {
        return Lexer{
            .alloc = alloc,
            .src = src,
            .i = 0,
            .errors = ErrorList.init(alloc),
        };
    }

    pub fn deinit(self: Lexer) void {
        self.errors.deinit();
    }

    pub fn next(self: *Lexer) !?Token {
        const src = self.src;

        self.skipWhitespace();
        if (self.i > self.src.len) {
            return null;
        }

        if (self.i == self.src.len) {
            defer self.i += 1;
            return Token.init(.eof, self.i, self.i);
        }

        // Test for single-character tokens.
        if (tokenKindForChar(self.currentUnchecked())) |kind| {
            defer self.advance();
            return Token.init(kind, self.i, self.i + 1);
        }

        // Test for multi-character tokens.
        switch (self.currentUnchecked()) {
            '<' => {
                if (self.peek()) |c| {
                    if (c == '>') {
                        defer self.advanceBy(2);
                        return Token.init(.op_ne, self.i, self.i + 2);
                    }
                    if (c == '=') {
                        defer self.advanceBy(2);
                        return Token.init(.op_lte, self.i, self.i + 2);
                    }
                }
                defer self.advance();
                return Token.init(.op_lt, self.i, self.i + 1);
            },
            '>' => {
                if (self.peek()) |c| {
                    if (c == '=') {
                        defer self.advanceBy(2);
                        return Token.init(.op_gte, self.i, self.i + 2);
                    }
                }
                defer self.advance();
                return Token.init(.op_gt, self.i, self.i + 1);
            },
            '|' => {
                if (self.peek()) |c| {
                    if (c == '|') {
                        defer self.advanceBy(2);
                        return Token.init(.op_or, self.i, self.i + 2);
                    } else {
                        return try self.errorToken("Unexpected character. Expected \"|\".");
                    }
                }
                return try self.errorToken("Unexpected EOF. Expected \"|\".");
            },
            '!' => {
                if (self.peek()) |c| {
                    if (c == '=') {
                        defer self.advanceBy(2);
                        return Token.init(.op_ne, self.i, self.i + 2);
                    }
                }
                return try  self.errorToken("invalid character. expected \"=\"");
            },
            'A'...'z' => {
                const start = self.i;
                var ch  = self.currentUnchecked();
                while (std.ascii.isAlphanumeric(ch) or ch == '_') {
                    self.advance();
                    if (self.isEof()) {
                        break;
                    }
                    ch = self.currentUnchecked();
                }
                var end = self.i;
                var kind = getKeywordKind(src[start..end]);
                return Token.init(kind, start, end);
            },
            '0'...'9' => {
                // TODO: Handle floating point numbers.
                var start = self.i;
                var ch  = self.currentUnchecked();
                while (std.ascii.isDigit(ch)) {
                    self.advance();
                    if (self.isEof()) {
                        break;
                    }
                    ch = self.currentUnchecked();
                }
                var end = self.i;
                return Token.init(.lit_int, start, end);
            },
            '\'' => {
                var start = self.i;
                self.advance(); // Skip the opening quote.
                if (self.isEof()) {
                    return try self.errorToken("unterminated string literal");
                }

                var invalidEscape = false;
                var ch  = self.currentUnchecked();
                while (ch != '\'') {
                    self.advance();
                    if (self.isEof()) {
                        return try self.errorToken("unterminated string literal");
                    }

                    ch = self.currentUnchecked();
                    if (ch == '\\') {
                        const next_ch = self.peek() orelse return try self.errorToken("unterminated string literal");
                        self.advance();
                        switch (next_ch) {
                            'n', 'r', 't', '\\', '\'' => {
                                if (self.peek()) |_| {
                                    self.advance(); // Skip over the escaped quote so that we do not exit the loop.
                                } else {
                                    return try self.errorToken("unterminated string literal");
                                }
                            },
                            else => {
                                invalidEscape = true;
                            }
                        }
                    }
                }
                self.advance(); // Skip the closing quote.
                if (invalidEscape) {
                    return Token.initError(start, self.i, try self.appendError("invalid escape sequence"));
                } else {
                    return Token.init(.lit_string, start, self.i);
                }
            },
            // TODO: Handle quoted identifiers.
            // TODO: Handle comments.
            else => {
                return try self.errorToken("invalid character");
            },
        }
    }

    fn errorToken(self: *Lexer, msg: String) !Token {
        const start = self.i;
        self.skipUntilWhitespace();

        return Token.initError(start, self.i, try self.appendError(msg));
    }

    fn appendError(self: *Lexer, msg: String) !i16 {
        const err_idx = @intCast(i16, self.errors.items.len);
        try self.errors.append(msg);

        return err_idx;
    }

    fn peek(self: Lexer) ?u8 {
        return self.peekN(1);
    }

    fn peekN(self: Lexer, i: usize) ?u8 {
        const next_i = self.i + i;
        if (next_i >= self.src.len) {
            return null;
        }
        return self.src[next_i];
    }

    fn skipWhitespace(self: *Lexer) void {
        if (self.i >= self.src.len) {
            return;
        }
        while (!self.isEof() and isWhitespace(self.currentUnchecked())) {
            _ = self.advance();
        }
    }

    fn skipUntilWhitespace(self: *Lexer) void {
        while (!self.isEof() and !isWhitespace(self.currentUnchecked())) {
            _ = self.advance();
        }
    }

    // TODO: Inline this.
    fn isEof(self: Lexer) bool {
        return self.i >= self.src.len;
    }

    // TODO: Inline this.
    fn currentUnchecked(self: Lexer) u8 {
        return self.src[self.i];
    }

    // TODO: Inline this.
    fn advanceBy(self: *Lexer, count: u32) void {
        self.i += count;
    }

    // TODO: Inline this.
    fn advance(self: *Lexer) void {
        self.advanceBy(1);
    }
};

// Token represents a lexical token.
// One goal of our lexer is to run with zero heap allocations. Thus,
// the Token contains only offsets into the input string. The parser
// is responsible for decoding tokens.
// Note that we use 32-bit offsets, so anyone wanting to lex a SQL
// string longer than 4GB is out of luck.
// We also do not support Unicode yet.
pub const Token = struct {
    // The token spans from start inclusive to end exclusive.
    start: u32,
    end: u32,
    err_idx: i16,
    kind: Kind,

    fn init(kind: Kind, start: u32, end: u32) Token {
        return Token{
            .start = start,
            .end = end,
            .kind = kind,
            .err_idx = -1,
        };
    }

    fn initError(start: u32, end: u32, err_idx: i16) Token {
        return Token{
            .start = start,
            .end = end,
            .kind = .err,
            .err_idx = err_idx,
        };
    }

    // Each token has a "Kind", which is the class of token it is.
    // Each keyword and punctuation has a dedicated Kind, which
    // allows us to avoid quite a few string comparisons in the parser.
    // The variable syntactic elements each belong to a general Kind,
    // which the parser can use to determine how to decode the token.
    pub const Kind = enum {
        // Reserved Keywords.
        kw_as,
        kw_by,
        kw_delete,
        kw_false,
        kw_from,
        kw_group,
        kw_inner,
        kw_insert,
        kw_into,
        kw_join,
        kw_offset,
        kw_on,
        kw_outer,
        kw_true,
        kw_select,
        kw_update,
        kw_where,

        // Punctuation.
        punc_comma,
        punc_paren_left,
        punc_paren_right,
        punc_semicolon,
        punc_star,

        // Operators.
        op_and,
        op_concat,
        op_divide,
        op_dot,
        op_eq,
        op_gt,
        op_gte,
        op_lt,
        op_lte,
        op_minus,
        op_modulo,
        op_ne,
        op_not,
        op_or,
        op_plus,

        // Literals.
        lit_float,
        lit_int,
        lit_null,
        lit_string,

        // Identifiers.
        // Note that in the lexer, we do not distinguish between non-reserved
        // keywords and bare identifiers. The parser is responsible for determining
        // whether a non-reserved keyword is being used as an identifier.
        bare_ident,
        quoted_ident,

        // End of file.
        eof,

        // Error.
        err,

        pub fn str(self: Kind) String {
            return switch (self) {
                .kw_as => "AS",
                .kw_by => "BY",
                .kw_delete => "DELETE",
                .kw_false => "FALSE",
                .kw_from => "FROM",
                .kw_group => "GROUP",
                .kw_inner => "INNER",
                .kw_insert => "INSERT",
                .kw_into => "INTO",
                .kw_join => "JOIN",
                .kw_offset => "OFFSET",
                .kw_on => "ON",
                .kw_outer => "OUTER",
                .kw_select => "SELECT",
                .kw_true => "TRUE",
                .kw_update => "UPDATE",
                .kw_where => "WHERE",

                .punc_comma => ",",
                .punc_paren_left => "(",
                .punc_paren_right => ")",
                .punc_semicolon => ";",
                .punc_star => "*",

                .op_and => "AND",
                .op_concat => "||",
                .op_divide => "/",
                .op_dot => ".",
                .op_eq => "=",
                .op_gt => ">",
                .op_gte => ">=",
                .op_lt => "<",
                .op_lte => "<=",
                .op_minus => "-",
                .op_modulo => "%",
                .op_ne => "<>",
                .op_not => "NOT",
                .op_or => "OR",
                .op_plus => "+",

                .lit_float => "float literal",
                .lit_int => "integer literal",
                .lit_null => "NULL",
                .lit_string => "string literal",

                .bare_ident => "identifier",
                .quoted_ident => "quoted identifier",

                .eof => "EOF",

                .err => "<ERROR>",
            };
        }
    };

    // string returns the slice of the source where the token is located.
    pub fn string(self: Token, src: String) String {
        return src[self.start..self.end];
    }

    // printDebug prints a deubg message with information about the token, the
    // line in which the token occurred in the source, and line/column numbers.
    //
    // Example:
    //  debug(src, tok, "expected \",\" or \"FROM\", got identifier")
    //  # Near identifier at 112:12: expected "," or "FROM", got identifier.
    //  # SELECT foo bar FROM baz
    //  #            ^
    pub fn printDebug(self: Token, src: String, msg: String) void {
        var line: u32 = 1;
        var col: u32 = 1;
        var line_cursor: usize = 0;

        var i: usize = 0;
        while (i < src.len) : (i += 1) {
            if (src[i] == '\n') {
                line_cursor = std.math.min(i + 1, src.len - 1);
                col = 1;
                line += 1;
            } else {
                col += 1;
            }

            if (i == self.start) {
                // Advance to the end of the line.
                while (i < src.len and src[i] != '\n') : (i += 1) {}
                break;
            }
        }

        // At this point, `line_start_idx` is the index of the start of the line
        // containing the `self` token, and `i` is the index of the end of the line.

        // Print the informational line.
        std.debug.print("Near {s} at {}:{}: {s}\n", .{ self.kind.str(), line, col, msg });

        // Print the line containing the token, substituting four spaces for a tab.
        var j: usize = line_cursor;
        while (j < i) : (j += 1) {
            if (src[j] == '\t') {
                std.debug.print("    ", .{});
            } else {
                std.debug.print("{s}", .{src[j .. j + 1]});
            }
        }
        std.debug.print("\n", .{});

        // Print the cursor line.
        while (line_cursor < self.start) : (line_cursor += 1) {
            if (src[line_cursor] == '\t') {
                std.debug.print("~~~~", .{});
            } else {
                std.debug.print("~", .{});
            }
        }
        std.debug.print("^", .{});
        line_cursor += 1;
        while (line_cursor < i) : (line_cursor += 1) {
            if (src[line_cursor] == '\t') {
                std.debug.print("~~~~", .{});
            } else {
                std.debug.print("~", .{});
            }
        }

        std.debug.print("\n", .{});
    }
};

inline fn isWhitespace(c: u8) bool {
    return c == ' ' or c == '\t' or c == '\n' or c == '\r';
}

const BuiltinKeyword = struct {
    name: String,
    kind: Token.Kind,
};
const builtins = [_]BuiltinKeyword{
    .{ .name = "AS", .kind = Token.Kind.kw_as },
    .{ .name = "BY", .kind = Token.Kind.kw_by },
    .{ .name = "DELETE", .kind = Token.Kind.kw_delete },
    .{ .name = "FROM", .kind = Token.Kind.kw_from },
    .{ .name = "GROUP", .kind = Token.Kind.kw_group },
    .{ .name = "INNER", .kind = Token.Kind.kw_inner },
    .{ .name = "INSERT", .kind = Token.Kind.kw_insert },
    .{ .name = "INTO", .kind = Token.Kind.kw_into },
    .{ .name = "JOIN", .kind = Token.Kind.kw_join },
    .{ .name = "OFFSET", .kind = Token.Kind.kw_offset },
    .{ .name = "ON", .kind = Token.Kind.kw_on },
    .{ .name = "OUTER", .kind = Token.Kind.kw_outer },
    .{ .name = "SELECT", .kind = Token.Kind.kw_select },
    .{ .name = "UPDATE", .kind = Token.Kind.kw_update },
    .{ .name = "WHERE", .kind = Token.Kind.kw_where },
};

fn getKeywordKind(str: String) Token.Kind {
    for (builtins) |builtin| {
        switch (std.ascii.orderIgnoreCase(builtin.name, str)) {
            .lt => {
                continue;
            },
            .eq => {
                return builtin.kind;
            },
            .gt => {
                continue;
            },
        }
    }

    // If no keyword matches, return ident.
    return Token.Kind.bare_ident;
}
