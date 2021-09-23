---
tags: ["btree","postgres","data-structures","db-internals"]
created: Thu Sep 16 17:02:38 MDT 2021
---

# The Postgres B Tree Implementation

Both the interior and lef nodes contain standard database pages[^1]. The storage format is decoupled from the index structure itself.

Postgres uses a technique in indexes to store the key only once when many consecutive records in a page have the same key value. They call this a "posting list tuple"[^1]. It appears that the overhead is small enough that they do not disable this feature automatically for unique indexes, even though it would always be a no-op.

Postgres avoids concurrent splits with 2 page-level additions (from Lehman & Yao Algorithm): a right-link ("next page") pointer, and an upper bound on all keys in the page. When searching down the tree, the search key is compared to the high key. If the search key is greater, then the page was split, and the next step should be to follow the right-link pointer. If the page had not been split, then the search would not have ended up on a page that contained only keys less than the search key.[^2]

[^1]: https://www.postgresql.org/docs/13/btree-implementation.html
[^2]: https://github.com/postgres/postgres/blob/REL_13_4/src/backend/access/nbtree/README
