---
tags:
  - database
  - tuple
  - encoding
created: Wed Apr 27 14:21:42 MDT 2022
---

# Tuple Representation

There are two cases that I have run into for representing tuples in the data
systems that I have worked on:

1. Serializing keys using an order-preserving encoding
2. Representing a row of data in memory or on disk

In the first case, it is advantageous to use an encoding that is trivially
binary-comparable. This is exactly the approach that FoundationDB uses with its
[Tuple
Layer](https://apple.github.io/foundationdb/data-modeling.html#data-modeling-tuples).
The advantages of this approach is that comparisons are fast, compound keys are
supported, and the data type is self-describing (i.e. no external schema need be
supplied). The primary disadvantage is that it takes up more space because each
tuple encodes the type of each field, and strings may require certain bytes to
be escaped with additional characters. The ability to iterate in order allows
this method to be used with key-value stores that support prefix scans.
Otherwise, a lower-level interface must be used.

For the second case, it may be preferrable to supply the schema separately from
the data because the additional storage required for redundant type tags may be
too significant, and unlike the tuple layer approach, a single field may be read
from the tuple before decoding the entire thing because the offset of each field
is known at the outset.

In general, the tuple layer approach is much easier from an engineering perspective, but the packed data+schema approach is more effecient both in terms of storage and processing.
