// This code was inspired by Zig's own AST.
const std = @import("std");
const tokenizer_mod = @import("tokenizer.zig");
const Token = tokenizer_mod.Token;
const Tokenizer = tokenizer_mod.Tokenizer;
const Allocator = std.mem.Allocator;

pub const TokenIndex = u32;

const Ast = @This();

source: []const u8,
tokens: []const Token,
nodes: NodeList.Slice,
errors: []const Error,
// data is a slice of Node.Index, but in practice, we interpret this as an array
// of structs whose members are all Node.Index. Each node may have a `data`
// member that is an index into the main `node_data` list that the AST maintains.
node_data: []const Node.Index,

// TODO: Should this be an ArrayList instead? What is the access pattern?
pub const NodeList = std.MultiArrayList(Node);
// The ErrorList is unmanaged because most of the data in the AST is built by
// the parser then passed to the AST and deallocated afterwards, so the lifecycle
// must start at or before parsing beings and must not end until after we are
// done with the AST.
pub const ErrorList = std.ArrayListUnmanaged(Error);

pub fn deinit(self: *Ast, gpa: Allocator) void {
  self.nodes.deinit(gpa);
  gpa.free(self.tokens);
  gpa.free(self.errors);
  gpa.free(self.node_data);
}

pub fn lexeme(self: *const Ast, idx: TokenIndex) []const u8 {
  const tok = self.token(idx);
  if (tok.lexeme()) |l| {
    return l;
  }

  // Initialize a new tokenizer that is pointing to the start of the token whose
  // lexeme we want.
  var tokenizer = Tokenizer {
    .source = self.source,
    .iStart = tok.start,
    .iCurrent = tok.start,
  };

  // Consume this token so that the tokenizer is now positioned at the start of
  // the next token.
  _ = tokenizer.scanToken();

  return self.source[tok.start..tokenizer.iCurrent];
}

pub fn data(self: *const Ast, comptime T: type, idx: Node.Index) *T {
  const datas = self.nodes.items(.data);
  const start = &datas[idx];

  return @ptrCast(*T, start);
}

pub fn token(self: *const Ast, idx: TokenIndex) Token {
  const tokIdxs = self.nodes.items(.token);
  const tokIdx = tokIdxs[idx];

  return self.tokens[tokIdx];
}

pub const Error = struct {
  tag: Tag,
  tokenIdx: usize,

  pub const Tag = enum(u8) {
    expected_eof,
    unexpected_token,
    unknown,
  };
};

pub const Node = struct {
  tag: Tag,
  token: TokenIndex,

  // Data is an index into the `node_data` slice. The exact type of data stored
  // can vary by node, but all (data-containing) nodes have a struct with one or
  // more Index fields, and `data` points to the offset in `node_data` where
  // this struct starts.
  // It is important that the data struct contain only `Index` fields so that
  // when we interpret a portion of `node_data`, it is correctly aligned.
  data: DataIndex,

  pub const DataList = std.ArrayListUnmanaged(Index);

  pub const Index = u32; // TODO: Change to u32 and update parser.

  // DataIndex points to the first index in the data array where the struct
  // containing pointers to child nodes exists.
  pub const DataIndex = u32;


  // The struct pointed to by the Data member is below
  pub const Tag = enum {
    lit_integer,
    lit_string,
    add,
    sub,
    mul,
    div,
    negate,
  };

  pub const UnaryOpData = struct {
    operand: Index,
  };

  pub const BinaryOpData = struct {
    operand: Index,
  };
};
