# FFI Example

This project demonstrates calling C code from Rust. In order to make this happen,
the `Makefile` contains a `build` target that causes the following to take place:

1. Compile `point.c` into `libpoint.so`. This shared library contains the `Point`
struct as well as a `point_distance` function that calculates the distance between
two points.
2. Run the cargo build script in `build.sh`, which invokes `bindgen` to generate Rust
bindings from the `point.h` header file and instruct the linker that `libpoint.so`
should be loaded.
3.

## Getting started

```
make build
./target/release/ffi-example
```

### Dependencies

- Rust
- LLVM/Clang

#### Example setup (Ubuntu)

```
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
apt install llvm-dev libclang-dev clang
```
