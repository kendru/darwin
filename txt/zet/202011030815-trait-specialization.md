---
tags: rust
created: Tue Nov 3 08:15:22 MST 2020
---

# Trait Specialization

Trait Specialization is a feature proposed in Rust [RFC 1210](https://github.com/rust-lang/rfcs/blob/master/text/1210-impl-specialization.md). It would allow for multiple `impl` blocks to be provided for a single trait on a given struct, and the most specific implementation would be chosen. The primary motivation is performance, since default impls could be provided that are provably correct, but additional implementations that that can be more optimized can be used in certain cases.
