---
tags: ["db-internals","query-planning"]
created: Wed Sep 15 16:00:36 MDT 2021
---

# Relational Query Planning

- For a query with an `OR` of predicates that can be fulfilled via an index scan/lookup, it is possible to perform the index operations then union the results before probing the table[^1].

- If a table is indexed on `(a, b)` where `a` has low cardinality, then a select of the form `SELECT "x" FROM "t" WHERE "b" = ?` could be rewritten as `SELECT "x" FROM "t" WHERE "a" in (SELECT "a" FROM "t") AND "b" = ?`, which for a small enough cardinality of `a` will out-perform a full table scan[^2].

- Query plan stability is very important for commercial database products. Even if the average plan improves, it is usually unacceptable to have a noticeable slowdown for some percentage of queries.

[^1]: https://youtu.be/gpxnbly9bz4?t=3205
[^2]: https://youtu.be/gpxnbly9bz4?t=3245
