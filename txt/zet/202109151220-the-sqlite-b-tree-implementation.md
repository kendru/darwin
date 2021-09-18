---
tags: ["btree","sqlite","data-structures","db-internals"]
created: Wed Sep 15 12:20:10 MDT 2021
---

# The SQLite B(+) Tree Implementation

## B+ Trees for Tables

SQlite B tree interface defines a db image made up of pages. The image may be persisted to disk or stored in memory only.

All interactions happen through a _page cache_.

It appears that the B tree itself does not enforce transactionality in its read/write. This needs to be handled at another layer within the B tree module that is dedicated to locking. This must not use latch crabbing.

Leaf nodes have header, row offsets, free space, and row data:

```
|--------|---...>--|-----...-----|---<...----|
| HEADER | OFFSETS | FREE SPACE  | ROW DATA  |
|--------|---...>--|-----...-----|---<...----|
```

B+ trees for tables and B trees for indexes manage their pages in a large array that the page cache maps to disk pages within the database file.

### Overflow

In order to support very large rows, part of the data is stored on a normal (leaf node) page, and the rest is stored on one or more overflow pages. Overflow pages are essentially linked lists with the page content and a pointer to the next overflow page.

## Indexes

Indexes use a B tree (not a B+ tree) so that keys are not duplicated. A B+ Tree for a table only uses 64-bit integers for keys, so duplication is fine. Secondary indexes can have arbitrarily large keys, so reducing duplication is desirable.

Supports querying from a covering index[^1].

### References:

- https://sqlite.org/btreemodule.html
- https://www.youtube.com/watch?v=gpxnbly9bz4

[^1]: https://youtu.be/gpxnbly9bz4?t=3190
