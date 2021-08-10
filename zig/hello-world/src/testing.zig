const expect = @import("std").testing.expect;

test "if statement" {
  const a = true;
  var x: u16 = 0;
  if (a) {
    x += 1;
  } else {
    x += 2;
  }

  try expect(x == 1);
}

test "if statement expression" {
  const a = true;
  var x: u16 = 0;
  x += if (a) 1 else 2;

  try expect(x == 1);
}

test "while" {
  var i: u8 = 2;
  while (i < 100) {
    i *= 2;
  }

  try expect(i == 128);
}

test "while with continue expression" {
  var sum: u8 = 0;
  var i: u8 = 1;
  while (i <= 10) : (i += 1) {
    sum += i;
  }

  try expect(sum == 55);
}

// test "for" {
//   const string = [_]u8{'a', 'b', 'c'};

//   for (string) |character, index| {}
//   for (string) |character| {}
//   for (string) |_, index| {}
//   for (string) |_| {}
// }

fn addFive(x: u32) u32 {
  return x + 5;
}

test "function" {
  const y = addFive(0);
  try expect(@TypeOf(y) == u32);
  try expect(y == 5);
}

fn fib(n: u16) u16 {
  if (n == 0 or n == 1) return n;
  return fib(n-1) + fib(n-2);
}

test "function recursion" {
  const x = fib(10);
  try expect(x == 55);
}

test "defer" {
  var x: i16 = 5;
  {
    defer x += 2;
    try expect(x == 5);
  }
    try expect(x == 7);
}

fn failingFunction() error{Oops}!void {
  return error.Oops;
}

test "returning an error" {
  failingFunction() catch |err| {
    try expect(err == error.Oops);
    return;
  };
}

fn increment(num: *u8) void {
  num.* += 1;
}

test "pointers" {
  var x: u8 = 1;
  increment(&x);
  try expect(x == 2);
}

fn total(values: []const u8) usize {
  var count: usize = 0;
  for (values) |value| count += value;
  return count;
}

test "slices" {
  const array = [_]u8{1,2,3,4,5};
  const slice = array[0..3];
  try expect(total(slice) == 6);
  try expect(@TypeOf(slice) == *const [3]u8);
}

const Vec3 = struct {
  x: f32, y: f32, z: f32
};

test "struct usage" {
  const my_vector = Vec3{
    .x = 0,
    .y = 100,
    .x = 50,
  };
}
