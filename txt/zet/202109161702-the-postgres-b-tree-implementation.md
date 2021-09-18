---
tags: ["btree","postgres","data-structures","db-internals"]
created: Thu Sep 16 17:02:38 MDT 2021
---

# The Postgres B Tree Implementation

Both the interior and lef nodes contain standard database pages[^1]. The storage format is decoupled from the index structure itself.

Postgres uses a technique in indexes to store the key only once when many consecutive records in a page have the same key value. They call this a "posting list tuple"[^1]. It appears that the overhead is small enough that they do not disable this feature automatically for unique indexes, even though it would always be a no-op.

[^1]: https://www.postgresql.org/docs/13/btree-implementation.html
