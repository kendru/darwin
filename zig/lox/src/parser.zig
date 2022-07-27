const std = @import("std");
const assert = std.debug.assert;
const Allocator = std.mem.Allocator;
const Ast = @import("ast.zig");
const Node = Ast.Node;
const ErrTag = Ast.Error.Tag;

const tokenizer_mod = @import("tokenizer.zig");
const Tokenizer = tokenizer_mod.Tokenizer;
const Token = tokenizer_mod.Token;

const TokenList = std.ArrayList(Token);

// The AST may have errors
pub fn parse(gpa: Allocator, source: []const u8) Allocator.Error!Ast {
  var tokens = TokenList.init(gpa);
  // Memory will be passed to AST before TokenList is deinitialized.
  defer tokens.deinit();

  var tokenizer = Tokenizer.init(source);

  // 1. Tokenize entire input. The tokens can be passed to the parser as well as the AST.
  //   - The AST will take ownership of the tokens list
  while (true) {
    const token = tokenizer.scanToken();
    try tokens.append(token);
    if (token.tag == .eof) break;
  }

  var parser = Parser {
    .gpa = gpa,
    .source = source,
    .tokens = tokens.toOwnedSlice(),
    .errors = .{},
    .nodes = .{},
    .node_data = .{},
    .idx = 0,
    .panic_mode = false,
  };

  // try parser.advance();
  _ = try parser.expression();
  _ = try parser.consume(.eof, .expected_eof);

  return parser.intoAst();
}

// The Parser builds the state that will eventually be passed to the AST.
// As such, it does ont own any allocations
const Parser = struct {
  // External state.
  gpa: Allocator,
  tokens: []const Token,
  source: []const u8,

  // Internal state: parser builds these but passes the allocation off
  // to an Ast.
  errors: Ast.ErrorList,
  nodes: Ast.NodeList,
  node_data: Node.DataList,
  idx: u32,
  panic_mode: bool,

  fn intoAst(self: *Parser) Ast {
    defer self.nodes.deinit(self.gpa);
    defer self.errors.deinit(self.gpa);
    // defer self.* = undefined;

    return Ast {
      .source = self.source,
      .tokens = self.tokens,
      .nodes = self.nodes.toOwnedSlice(),
      .errors = self.errors.toOwnedSlice(self.gpa),
      .node_data = self.node_data.toOwnedSlice(self.gpa),
    };
  }

  fn expression(self: *Parser) Allocator.Error!Node.Index {
    const current = self.tokens[self.idx];
    return switch (current.tag) {
        .number => self.number(),
        .minus => self.unary(),
        .left_paren => self.grouping(),
        else => @panic("Unhandled node type"),
    };
  }

  fn grouping(self: *Parser) Allocator.Error!Node.Index {
    _ = try self.consume(.left_paren, .unexpected_token);
    const idx = self.expression();
    _ = try self.consume(.right_paren, .unexpected_token);
    return idx;
  }

  fn unary(self: *Parser) Allocator.Error!Node.Index {
    const op_idx = self.idx;
    const op_tok = self.tokens[self.idx];
    std.debug.print("Got a unary expression: {s}\n", .{@tagName(op_tok.tag)});
    try self.advance();

    return switch (op_tok.tag) {
      .minus => self.addNode(.{
        .tag = .negate,
        .token = op_idx,
        .data = try self.addNodeData(Node.UnaryOpData{
          .operand = try self.expression(),
        }),
      }),
      else => unreachable,
    };
  }

  fn number(self: *Parser) Allocator.Error!Node.Index {
    const idx = try self.consume(.number, .unknown);
    return self.addNode(.{
      .tag = .lit_integer,
      .token = idx,
      .data = undefined,
    });
  }

  fn string(self: *Parser) Allocator.Error!void {
    _ = self;
    @panic("TODO: Parse string");
  }

  fn boolean(self: *Parser) Allocator.Error!void {
    _ = self;
    @panic("TODO: Parse boolean");
  }

  // Adds a node type-specific data struct. Since the only requirement of `data`
  // is that it contain only `Node.Index` fields, we accept any type but
  // validate that it is aligned to the size of `Node.Index`.
  fn addNodeData(self: *Parser, data: anytype) Allocator.Error!Node.DataIndex {
    const data_type = @TypeOf(data);
    comptime {
      assert(@alignOf(data_type) == @alignOf(Node.Index));
    }

    // The currrent length before appending new fields becomes the start index
    // for the new node_data struct.
    const start_idx = @intCast(Node.DataIndex, self.node_data.items.len);

    // Allocate enough space for all the fields, and append them individually.
    const fields = std.meta.fields(data_type);
    try self.node_data.ensureUnusedCapacity(self.gpa, fields.len);
    inline for (fields) |field| {
      comptime assert(field.field_type == Node.Index);
      self.node_data.appendAssumeCapacity(@field(data, field.name));
    }

    return start_idx;
  }

  fn addNode(self: *Parser, n: Ast.Node) Allocator.Error!Node.Index {
    const idx = @intCast(Node.Index, self.nodes.len);
    std.debug.print("Adding node {s} at {d}\n", .{ @tagName(n.tag), idx });
    try self.nodes.append(self.gpa, n);
    return idx;
  }

  fn advance(self: *Parser) Allocator.Error!void {
    // Skip over invalid tokens, accumulating errors. Stop at the next valid
    // token.`
    while (true) {
      const current = self.tokens[self.idx];
      self.idx += 1;
      if (current.tag != .invalid) break;

      try self.errorAtCurrent(.unknown);
    }
  }

  fn consume(self: *Parser, tag: Token.Tag, errTag: ErrTag) Allocator.Error!Ast.TokenIndex {
    const prev = self.idx;
    const current = self.tokens[prev];
    if (current.tag == tag) {
      try self.advance();
    } else {
      try self.errorAtCurrent(errTag);
    }

    return @intCast(Ast.TokenIndex, prev);
  }

  fn errorAtCurrent(self: *Parser, tag: ErrTag) Allocator.Error!void {
    try self.errorAt(self.idx, tag);
  }

  fn errorAtPrev(self: *Parser, tag: ErrTag) Allocator.Error!void {
    try self.errorAt(self.idx-1, tag);
  }

  fn errorAt(self: *Parser, idx: usize, tag: ErrTag) Allocator.Error!void {
    if (self.panic_mode) return;
    self.panic_mode = true;

    const tok = self.tokens[idx];
    const pos = self.posFor(tok);
    std.debug.print("[line {d}]", .{pos.line});
    switch (tok.tag) {
        .eof => std.debug.print(" at end", .{}),
        .invalid => {
          // Do nothing
        },
        else => std.debug.print(" at unknown", .{}),
    }
    std.debug.print(": {s}\n", .{@tagName(tag)});

    // TODO: include location information on error
    try self.errors.append(self.gpa, .{ .tag = tag, .tokenIdx = idx });
  }

  fn posFor(self: *Parser, tok: Token) Position {
    var loc = Position {
      .line = 1,
      .col = 1,
    };

    for (self.source) |ch,i| {
      if (i == tok.start) {
        break;
      }

      if (ch == '\n') {
        loc.line += 1;
        loc.col = 1;
      }
      loc.col += 1;
    }

    return loc;
  }
};


const Position = struct {
  line: u32,
  col: u32,
};
