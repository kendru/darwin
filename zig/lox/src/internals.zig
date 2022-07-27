const std = @import("std");
const Allocator = std.mem.Allocator;
const assert = std.debug.assert;

pub fn DynamicArray(comptime T: type) type {
  return struct {
    const Self = @This(); // Cool!

    alloc: Allocator,
    capacity: usize,
    items: []T,

    pub fn init(alloc: Allocator) Self {
      return Self {
        .alloc = alloc,
        .capacity = 0,
        .items = &[_]T{},
      };
    }

    pub fn deinit(self: Self) void {
      self.alloc.free(self.items);
    }

    pub fn append(self: *Self, item: T) Allocator.Error!void {
      var empty = try self.addOne();
      // This is what a dereference looks like.
      empty.* = item;
    }

    pub fn get(self: *Self, idx: usize) !T {
      assert(idx < self.items.len);

      return self.items[idx];
    }

    fn addOne(self: *Self) Allocator.Error!*T {
      if (self.capacity < self.items.len + 1) {
        try self.grow();
      }
      self.items.len += 1;
      return &self.items[self.items.len - 1];
    }

    fn grow(self: *Self) Allocator.Error!void {
      const requested_cap = if (self.capacity < 8) 8 else self.capacity * 2;
      const new_alloc = try self.alloc.reallocAtLeast(self.fullAllocation(), requested_cap);
      self.items.ptr = new_alloc.ptr;
      self.capacity = new_alloc.len; // Set to actual capacity in case allocator returned more.
    }

    // Returns the full allocation of items rather than just the used ones.
    fn fullAllocation(self: *Self) []T {
      return self.items.ptr[0..self.capacity];
    }
  };
}

test "append" {
  const Bytes = DynamicArray(u8);
  var b = Bytes.init(std.heap.page_allocator);
  try b.append(8);
  try b.append(0);
  try b.append(8);
  try b.append(3);
  try b.append(1);

  try std.testing.expectEqual(@as(u8, 8), b.items[0]);
  try std.testing.expectEqual(@as(u8, 1), b.items[4]);
}
