---
tags: ["rust","learnings"]
created: Thu Sep 30 10:14:55 MDT 2021
---

# Learning Rust Notes

## Data Types

### Marker Types

These types do not have any behaviour associated with them. They are used only as directives to the compiler to indicate that objects obey certain properties.

#### std::marker::PhantomData

A zero-sized generic type that can be used as a struct member to force the owning struct to use a type parameter that it would not otherwise use.

*Use when:*

- You need to artificially tie a struct to a lifetime parameter when it does not actually own a reference with that lifetime.
- Make a type contravariant or invariant over some generic parameter, e.g.:

```rust
struct ContravariantGerericUsize<T> {
  val: usize,
  _marker: PhantomData<fn(T)>
}
```

- Hint to the drop checker that a type that does not own a `T` may still drop a `T`, e.g.:

```rust
struct RemoteAloc<T> {
  data: *const T, // Not an owned T.
  _marker: PhantomData<T>,
}
```

#### UnsafeCell

The lowest-level primitive type that allows for interior mutability types to be built. `UnsafeCell` itself does not enforce any invariants that make access safe (hence `Unsafe`). It is used within `Cell` and `RefCell` to provide checked access to shared references.
