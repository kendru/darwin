const std = @import("std");

pub const Token = struct {
  tag: Tag,
  start: usize,

  pub const Tag = enum(u8) {
    // Single-character tokens
    left_paren, right_paren,
    left_brace, right_brace,
    comma, dot, minus, plus,
    semicolon, slash, star,
    // One- or two-character tokens
    bang, bang_equal,
    equal, equal_equal,
    greater, greater_equal,
    less, less_equal,
    // Literals
    identifier, string, number,
    // Keywords
    @"and", class, @"else", @"false",
    @"for", fun, @"if", nil, @"or",
    print, @"return", super, this,
    @"true", @"var", @"while",

    // Sentinel values
    invalid, eof,
  };

  pub fn lexeme(self: Token) ?[]const u8 {
    return switch (self.tag) {
      .invalid,
      .eof,
      .identifier,
      .string,
      .number,
      => null,

      .left_paren => "(",
      .right_paren => ")",
      .left_brace => "{",
      .right_brace => "}",
      .comma => ",",
      .dot => ".",
      .minus => "-",
      .plus => "+",
      .semicolon => ";",
      .slash => "/",
      .star => "*",
      .bang => "!",
      .bang_equal => "!=",
      .equal => "=",
      .equal_equal => "==",
      .greater => ">",
      .greater_equal => ">=",
      .less => "<",
      .less_equal => "<=",
      .@"and" => "and",
      .class => "class",
      .@"else" => "else",
      .@"false" => "false",
      .@"for" => "for",
      .fun => "fun",
      .@"if" => "if",
      .nil => "nil",
      .@"or" => "or",
      .print => "print",
      .@"return" => "return",
      .super => "super",
      .this => "this",
      .@"true" => "true",
      .@"var" => "var",
      .@"while" => "while",
    };
  }
};

pub const Tokenizer = struct {
  source: []const u8,
  iStart: usize,
  iCurrent: usize,

  pub fn init(source: []const u8) Tokenizer {
    // TODO: Do we need to worry about a UTF-8 BOM?

    return Tokenizer {
      .source = source,
      .iStart = 0,
      .iCurrent = 0,
    };
  }

  pub fn scanToken(self: *Tokenizer) Token {
    if (self.isEof()) return self.makeToken(.eof);

    self.skipWhitespace();
    self.iStart = self.iCurrent;

    const ch = self.advance();
    if (isDigit(ch)) return self.number();
    if (isAlpha(ch)) return self.identifier();
    return switch (ch) {
        '(' => self.makeToken(.left_paren),
        ')' => self.makeToken(.right_paren),
        '{' => self.makeToken(.left_brace),
        '}' => self.makeToken(.right_brace),
        ';' => self.makeToken(.semicolon),
        ',' => self.makeToken(.comma),
        '.' => self.makeToken(.dot),
        '-' => self.makeToken(.minus),
        '+' => self.makeToken(.plus),
        '/' => self.makeToken(.slash),
        '*' => self.makeToken(.star),
        '!' => self.makeToken(if (self.match('=')) .bang_equal else .bang),
        '=' => self.makeToken(if (self.match('=')) .equal_equal else .equal),
        '>' => self.makeToken(if (self.match('=')) .greater_equal else .greater),
        '<' => self.makeToken(if (self.match('=')) .less_equal else .less),
        '"' => self.string(),
        else => self.makeInvalid(),
    };
  }

  fn skipWhitespace(self: *Tokenizer) void {
    while (true) {
      switch (self.peek()) {
        ' ', '\t', '\r', '\n' => self.consumeNext(),
        '/' => {
          if (self.peekNext() == '/') {
            // Skip entire line of comment.
            while (self.peek() != '\n' and !self.isEof()) self.consumeNext();
          } else {
            return;
          }
        },
        else => {
          return;
        },
      }
    }
  }

  fn identifier(self: *Tokenizer) Token {
    while (isAlpha(self.peek()) or isDigit(self.peek())) self.consumeNext();
    return self.makeToken(self.identifierType());
  }

  fn identifierType(self: *Tokenizer) Token.Tag {
    switch (self.source[self.iStart]) {
        'a' => return self.checkKeyword(1, "nd", .@"and"),
        'c' => return self.checkKeyword(1, "lass", .class),
        'e' => return self.checkKeyword(1, "lse", .@"else"),
        'f' => {
          if (self.lexemeLength() > 1) {
            switch (self.source[self.iStart+1]) {
              'a' => return self.checkKeyword(2, "lse", .@"false"),
              'o' => return self.checkKeyword(2, "r", .@"for"),
              'u' => return self.checkKeyword(2, "n", .fun),
              else => {},
            }
          }
        },
        'i' => return self.checkKeyword(1, "f", .@"if"),
        'n' => return self.checkKeyword(1, "il", .nil),
        'o' => return self.checkKeyword(1, "r", .@"or"),
        'p' => return self.checkKeyword(1, "rint", .print),
        'r' => return self.checkKeyword(1, "eturn", .@"return"),
        's' => return self.checkKeyword(1, "uper", .super),
        't' => {
          if (self.lexemeLength() > 1) {
            switch (self.source[self.iStart+1]) {
              'h' => return self.checkKeyword(2, "is", .this),
              'r' => return self.checkKeyword(2, "ue", .@"true"),
              else => {},
            }
          }
        },
        'v' => return self.checkKeyword(1, "ar", .@"var"),
        'w' => return self.checkKeyword(1, "hile", .@"while"),
        else => {},
    }

    return .identifier;
  }

  fn checkKeyword(self: *Tokenizer, relativeStartIdx: usize, cmp: []const u8, tt: Token.Tag) Token.Tag {
    // If the current token is the wrong length for the keyword, see if
    const compareKwLen = relativeStartIdx + cmp.len;
    if (self.lexemeLength() != compareKwLen) return .identifier;

    const startIdx = self.iStart + relativeStartIdx;
    const endIdx = startIdx + cmp.len;
    return if (std.mem.eql(u8, self.source[startIdx..endIdx], cmp)) tt else .identifier;
  }

  fn lexemeLength(self: *Tokenizer) usize {
    return self.iCurrent - self.iStart;
  }

  fn string(self: *Tokenizer) Token {
    var is_escaped: bool = false;
    while (!self.isEof()) : (self.consumeNext()) {
      switch (self.peek()) {
        '\\' => {
          is_escaped = true;
          continue;
        },
        '"' => {
          if (!is_escaped) {
            break;
          }
        },
        else => {},
      }
      is_escaped = false;
    }
    if (self.isEof()) {
      return self.makeInvalid();
    }
    self.consumeNext(); // Consume closing quote.

    return self.makeToken(.string);
  }

  fn number(self: *Tokenizer) Token {
    while (isDigit(self.peek())) self.consumeNext();
    if (self.peek() == '.' and isDigit(self.peekNext())) {
      self.consumeNext(); // Consume dot.
      while (isDigit(self.peek())) self.consumeNext();
    }

    return self.makeToken(.number);
  }

  fn peek(self: *Tokenizer) u8 {
    if (self.isEof()) return 0;
    return self.source[self.iCurrent];
  }

  fn peekNext(self: *Tokenizer) u8 {
    if (self.isEof()) return 0;
    return self.source[self.iCurrent+1];
  }

  pub fn consumeNext(self: *Tokenizer) void {
    _ = self.advance();
  }

  fn advance(self: *Tokenizer) u8 {
    if (self.isEof()) return 0;
    defer self.iCurrent += 1;
    return self.source[self.iCurrent];
  }

  fn match(self: *Tokenizer, ch: u8) bool {
    if (self.isEof()) return false;
    if (self.source[self.iCurrent] != ch) return false;
    self.iCurrent += 1;
    return true;
  }

  fn makeInvalid(self: *Tokenizer) Token {
    return self.makeToken(.invalid);
  }

  fn makeToken(self: *Tokenizer, t: Token.Tag) Token {
    return Token {
      .tag = t,
      .start = self.iStart,
    };
  }

  pub fn isEof(self: *Tokenizer) bool {
    return self.iCurrent >= self.source.len;
  }
};

fn isAlpha(ch: u8) bool {
  return switch (ch) {
    'a'...'z', 'A'...'Z', '_' => true,
    else => false,
  };
}

fn isDigit(ch: u8) bool {
  return switch (ch) {
    '0'...'9' => true,
    else => false,
  };
}
