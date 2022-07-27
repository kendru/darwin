// This file contains tests that represent my attempt at understanding things
// about the language that were not clear to me from documentation or examples.

const std = @import("std");

const expect = std.testing.expect;

test "enums work the way i think they do" {
  const Hai = enum(u8) {
    ok,
    _, // Make enum non-exhaustive.
  };

  // We are parsing from string so that Zig does not statically determine that `47`
  // cannot match any tag of `Hai`.
  const tag_val = try std.fmt.parseUnsigned(u8, "47", 10);
  const result = switch (@intToEnum(Hai, tag_val)) { // Not an error.
      .ok => "Ok",
      _ => "Not Ok",
  };
  try expect(std.mem.eql(u8, "Not Ok", result));
}

test "identifiers may break naming rules with @\"name\" syntax" {
  const @"my value" = 123; // Contains spaces.
  const @"456" = 456;      // Starts with non-alpha.

  try expect(@"my value" + @"456" == 579);
}

// test "slices and pointer flavors" {
//   const some_val = "Hello, World";
//   // TODO: Use the corrext syntax here. These are complete guesses.
//   const val_ptr: *u8 = &some_val[0];
//   const many_val_ptr: [*]u8 = &some_val[0..];
// }

test "files define a struct" {
  const ImplicitStruct = @import("./ImplicitStruct.zig");
  const my_string = "Some data.";
  const s = ImplicitStruct.initialize(my_string[0..]);

  try expect(s.length() == 10);
}

test "pointers inherit constness" {
  var data = "abcdefg".*;
  const mutable = data[0..3]; // abc
  const immutable = @as([]const u8, data[3..]); // defg

  // These functions are given exactly the type that they expect.
  try expect(var_len(mutable) == 3);
  try expect(const_len(immutable) == 4);

  // A function that expects a []const T can also take a []T.
  try expect(const_len(mutable) == 3);

  // However, a function that expects a (mutable) []T cannot take
  // a []const T.
  // try expect(var_len(immutable) == 4);
}

fn var_len(a: []u8) usize {
  return a.len;
}

fn const_len(a: []const u8) usize {
  return a.len;
}
