// parse.zig
//
// This file contains a recursive descent parser for our SQL
// dialect.
const std = @import("std");
const lex = @import("lex.zig");
const query = @import("query.zig");
const util = @import("util.zig");

const Allocator = std.mem.Allocator;
const String = util.String;
const Lexer = lex.Lexer;
const Token = lex.Token;

pub const Parser = struct {
    allocator: Allocator,
    lexer: Lexer,
    current: Token,
    next: ?Token,

    // Create a new parser that has ownership of `lexer`.
    // The parser will take care of deallocating the lexer.
    // We recommend using an arena allocator for the parser, since
    // it will perform many allocations that have the same lifetime
    // and benefit from adjacent allocation
    pub fn init(allocator: Allocator, lexer: Lexer) Parser {
        return Parser{
            .allocator = allocator,
            .lexer = lexer,
            .current = Token{
              .kind = .err,
              .start = 0,
              .end = 0,
              .err_idx = -1,
            },
            .next = null,
        };
    }

    pub fn deinit(self: Parser) void {
        self.lexer.deinit();
    }

    // TODO: Expose a function that parses zero or more many queries.

    // TODO: Create a helpful error set.
    pub fn parseQuery(self: *Parser) !query.Query {
        self.current = (try self.lexer.next()).?;
        self.next = try self.lexer.next();
        return self.parseStatement();
    }

    // Parse a top-level statement.
    fn parseStatement(self: *Parser) !query.Query {
        return switch (self.current.kind) {
            .kw_select => .{ .select = try self.parseSelect() },
            .err => .{ .err = "TODO: Pass top-level lex error" },
            else => .{ .err = "expected statement" },
        };
    }

    // Parse a select statement.
    fn parseSelect(self: *Parser) !query.SelectQuery {
        // Parse the select clause.
        const select = try self.parseSelectClause();

        // // Parse the from clause.
        // const from = try self.parseFromClause();

        // // Parse the where clause.
        // const where = try self.parseWhereClause();

        // // Parse the group by clause.
        // const group_by = try self.parseGroupByClause();

        // // Parse the having clause.
        // const having = try self.parseHavingClause();

        // // Parse the order by clause.
        // const order_by = try self.parseOrderByClause();

        // // Parse the limit clause.
        // const limit = try self.parseLimitClause();
        // // Parse the from clause.
        // const from = try self.parseFromClause();

        // // Parse the where clause.
        // const where = try self.parseWhereClause();

        // // Parse the group by clause.
        // const group_by = try self.parseGroupByClause();

        // // Parse the having clause.
        // const having = try self.parseHavingClause();

        // // Parse the order by clause.
        // const order_by = try self.parseOrderByClause();

        // // Parse the limit clause.
        // const limit = try self.parseLimitClause();

        return query.SelectQuery{
            .select = select,
            .from = null,
            .where = null,
            .group_by = null,
            .having = null,
            .order_by = null,
            .limit = null,
            // .from = from,
            // .where = where,
            // .group_by = group_by,
            // .having = having,
            // .order_by = order_by,
            // .limit = limit,
        };
    }

    // Parse a select clause.
    fn parseSelectClause(self: *Parser) !query.SelectClause {
        // Parse the select keyword.
        if (!try self.expect(.kw_select)) {
            return error { TodoError }.TodoError;
        }

        // Parse the select list.
        const list = try self.parseColumnList();

        return query.SelectClause{
            .column_list = list,
        };
    }

    // Parse a column list.
    fn parseColumnList(self: *Parser) ![]query.ColumnExpression {
        // XXX: This will likely way over-allocate since we are using an arena allocator.
        //      We need to be able to tokenize the list to get the length then parse the
        //      list in a second pass into a slice of known size.
        var list = std.ArrayList(query.ColumnExpression).init(self.allocator);

        // Parse the first column expression.
        var column = try self.parseSelectColumn();
        try list.append(column);
        // Parse the remaining column expressions.
        while (self.current.kind == .punc_comma) : ({ _ = try self.expect(.punc_comma); }) {
            _ = try self.accept(.punc_comma);
            column = try self.parseSelectColumn();
            try list.append(column);
        }

        return list.items;
    }

    // Parse a select column.
    fn parseSelectColumn(self: *Parser) !query.ColumnExpression {
        // Parse the column expression.
        const expr = try self.parseExpression();

        // Parse the optional column alias.
        var alias: ?String = null;
        if (try self.accept(.kw_as)) {
            const alias_expr = try self.parseIdentifier();
            alias = alias_expr.identifier;
        } else if (self.next) |next| {
            if (next.kind == .bare_ident or next.kind == .quoted_ident) {
                _ = try self.advance();
                const alias_expr = try self.parseIdentifier();
                alias = alias_expr.identifier;
            }
        }

        return query.ColumnExpression{
            .expr = expr,
            .alias = alias,
        };
    }

    // Parse an expression.
    fn parseExpression(self: *Parser) !query.Expression {
        return switch (self.current.kind) {
            .lit_int => try self.parseIntegerLiteral(),
            .lit_float => try self.parseFloatLiteral(),
            .bare_ident, .quoted_ident => try self.parseIdentifier(),
            else => unreachable,
        };
    }

    // Parse an integer literal.
    fn parseIntegerLiteral(self: *Parser) !query.Expression {
        const token = self.current;
        const num = try std.fmt.parseInt(i64, token.string(self.lexer.src), 10);
        return query.Expression{
            .integer = num,
        };
    }

    // Parse a floating point literal.
    fn parseFloatLiteral(self: *Parser) !query.Expression {
        const token = self.current;
        const num = try std.fmt.parseFloat(f64, self.tokenSource(token));
        return query.Expression{
            .floating_point = num,
        };
    }

    // Parse an identifier.
    fn parseIdentifier(self: *Parser) !query.Expression {
        return switch (self.current.kind) {
            .bare_ident => try self.parseBareIdentifier(),
            .quoted_ident => try self.parseQuotedIdentifier(),
            else => unreachable,
        };
    }

    // Parse a bare identifier.
    fn parseBareIdentifier(self: *Parser) !query.Expression {
        return try self.parseIdentifierFromStr(self.tokenSource(self.current));
    }

    // Parse a quoted identifier.
    fn parseQuotedIdentifier(self: *Parser) !query.Expression {
        const str = self.tokenSource(self.current);
        return try self.parseIdentifierFromStr(str[1..str.len - 1]);
    }

    // Parse an identifier from a string.
    fn parseIdentifierFromStr(self: *Parser, str: String) !query.Expression {
        // Allocate a string for the identifier. We allocate enough space for the
        // original identifier, but if there are escape sequences, we will use
        // less memory. Ideally, we could use copy-on-write and share the original
        // memory if there are no escape sequences.
        var out = try self.allocator.alloc(u8, str.len );

        // XXX: This parser was generated by Copilot, and we have not yet
        // audited the code for correctness.

        // Parse the identifier.
        var i: usize = 0;
        while (i < str.len) : (i += 1) {
            var c = str[i];
            if (c == '\\') {
                // Parse an escape sequence.
                if (i + 1 >= str.len) {
                    self.reportError("unexpected end of input", .{});
                }
                c = str[i + 1];
                switch (c) {
                    '0' => out[i] = 0,
                    'n' => out[i] = '\n',
                    'r' => out[i] = '\r',
                    't' => out[i] = '\t',
                    '\\' => out[i] = '\\',
                    '\'' => out[i] = '\'',
                    'x' => {
                        if (i + 3 >= str.len) {
                            self.reportError("unexpected end of input", .{});
                        }
                        const hex = str[i + 2..i + 4];
                        const num = try std.fmt.parseInt(u8, hex, 16);
                        out[i] = num;
                        i += 2;
                    },
                    'u' => {
                        if (i + 5 >= str.len) {
                            self.reportError("unexpected end of input", .{});
                        }
                        const hex = str[i + 2..i + 6];
                        const num = try std.fmt.parseInt(u16, hex, 16);
                        out[i] = @truncate(u8, num << 8);
                        out[i + 1] = @truncate(u8, num);
                        i += 4;
                    },
                    'U' => {
                        if (i + 9 >= str.len) {
                            self.reportError("unexpected end of input", .{});
                        }
                        const hex = str[i + 2..i + 10];
                        const num = try std.fmt.parseInt(u32, hex, 16);
                        out[i] = @truncate(u8, num << 24);
                        out[i + 1] = @truncate(u8, num << 16);
                        out[i + 2] = @truncate(u8, num << 8);
                        out[i + 3] = @truncate(u8, num);
                        i += 8;
                    },
                    else => {
                        self.reportError("invalid escape sequence", .{});
                    }
                }
                i += 1;
            } else {
                out[i] = c;
            }
        }

        return query.Expression{
            .identifier = out[0..i],
        };
    }

    fn expect(self: *Parser, kind: Token.Kind) !bool {
        const accepted = try accept(self, kind);
        if (!accepted) {
            self.reportError("expected token of kind {} but got {}", .{kind, self.current.kind});
        }
        return accepted;
    }

    // Do we need this, or is it only called by expect()?
    fn accept(self: *Parser, kind: Token.Kind) !bool {
        if (self.current.kind == kind) {
            return self.advance();
        }
        return false;
    }

    fn advance(self: *Parser) !bool {
        if (try self.lexer.next()) |next| {
            self.current = self.next.?;
            self.next = next;
            return true;
        }
        return false;
    }

    fn tokenSource(self: Parser, token: Token) String {
        return token.string(self.lexer.src);
    }

    // TODO: Improve error reporting.
    fn reportError(self: *Parser, comptime format: String, args: anytype) void {
        _ = self;
        std.debug.print(format, args);
    }
};
