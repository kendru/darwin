---
tags: ["databases","recursive-queries"]
created: Tue Apr  6 10:47:34 MDT 2021
---

# Recursive Query Plans

When considering SQL, a recursive CTE builds up a result set by using some _anchor query_[^1] to establish an initial result set, which we will call the _initial set_, that is combined (using `UNION` or `UNION ALL`) with the results of another query that can join against the CTE itself. The query is considered "satisfied" when no additional rows are returned by the recursive query.

Execution of a recursive query is a fixpoint process in which the recursive query + join is run repeatedly until an iteration is reached where the set no longer grows.

A key consideration in planning a recursive query is what execution strategies and optimizations are legal. The operations that may be run inside the recursive procedure are limited to those that distribute over union all. For example, simple projection and inner joins are allowed, but distinct and top are not allowed.

#### Example: projection distributes over union all

```sql
-- Project(Union All(tbl_1, tbl_2), p)
SELECT a, b, c FROM (
  SELECT * FROM  tbl_1
  UNION ALL
  SELECT * FROM tbl_2
)

-- is equivalent to
-- Union All(Project(tbl_1, p), Project(tbl_2, p))
SELECT a, b, c FROM tbl_1
UNION ALL
SELECT a, b, c FROM tbl_2
```

#### Example: selection distributes over union all

```sql
-- Select(Union All(tbl_1, tbl_2), pred)
SELECT * FROM (
  SELECT * FROM  tbl_1
  UNION ALL
  SELECT * FROM tbl_2
) WHERE pred

-- is equivalent to

-- Union All(Select(tbl_1, pred), Select(tbl_2, pred))
SELECT * FROM tbl_1 WHERE pred
UNION ALL
SELECT * FROM tbl_2 WHERE pred
```

#### Example: distinct does not distribute over union all

```sql
-- Distinct(Union All(tbl_1, tbl_2), expr)
SELECT DISTINCT expr FROM (
  SELECT * FROM  tbl_1
  UNION ALL
  SELECT * FROM tbl_2
) WHERE p

-- is NOT equivalent to

-- Union All(Distinct(tbl_1, expr), Distinct(tbl_2, expr))
SELECT DISTINCT expr FROM tbl_1 WHERE p
UNION ALL
SELECT DISTINCT expr FROM tbl_2 WHERE p
```

If the recursive query contains operators like distinct, they cannot be "pushed down" to the recursive step and must be performed after a fixpoint is reached (or disallowed).

## Postgres Implementation

The Postgres documentation outlines the basic algorithm used to evaluate a recursive query[^2]. In pseudocode, it works as follows:

```
 1: intermediate_tbl <- {}
 2: working_tbl, result_tbl <- eval(initial)
 3:
 4: while not empty?(working_tbl):
 5:   intermediate_table <- eval(recursive(working_tbl))
 6:   result_tbl <- union_all(result_tbl, intermediate_table)
 7:   working_tbl <- intermediate_table
 8:   intermediate_table <- {}
 9:
10: yield result_tbl
```

The algorithm above works for a recursive CTE that is combined with `UNION ALL`. If combining results with `UNION` instead, the following modifications must be made:

1. The evaluation of `initial` on line 2 must be deduplicated.
2. The evaluation of `recursive` on line 5 must be deduplicated, and lines matching any row in result_tbl must also be discarded.


## Infinite Recursion

Recursive queries can be written in such a way that they never terminate. For example, the following query (adapted from the Postgres docs) computes the sequence of positive integers:

```sql
WITH RECURSIVE pos_ints(n) AS (
  SELECT 1
  UNION ALL
  SELECT n+1 FROM pos_ints
)
```

The database can attempt to detect non-terminating cycles through allowing a maximum depth of recursion or similar, but it could also use conditions or limits in the query that selects from the recursive query to determine termination criteria. For example, it could take a query like the following and use the LIMIT clause to determine the maximum depth of recursion:

```sql
WITH RECURSIVE pos_ints(n) AS (
  SELECT 1
  UNION ALL
  SELECT n+1 FROM pos_ints
)
SELECT n FROM pos_ints LIMIT 100
```

## Equivalence in Datalog

A recursive rule in datalog is equivalent to the recursive CTE in SQL, since they can both be expressed in terms of relational algebra. Just as a recursive CTE is defined in terms of an anchor query and a recursive query, a recursive rule in datalog is defined in terms of a non-recursive head and a recursive one. For example, the following rule expresses the transitive closure of a graph:

```prolog
reachable(X, Y) :- connected(X, Y). % non-recursive rule head
reachable(X, Y) :- connected(X, Z), reachable(Z, Y). % recursive
```

The semantics of this program can be similarly expressed in a recursive CTE, so the same evaluation strategy can be used.

### Terminology

- *anchor query*: query that returns the initial resultset in a recursive query
- *initial set*: result set returned by the _anchor query_
- *recursive query*: query that returns a resultset that is joined against the entire CTE
- *next set*: result set returned by the _recursive query_

### See Also

See [Fossil SCM Design](202012091138-fossil-scm-design) for an example of recursive queries using both SQLite and Datalog syntax.

[^1]: Anchor is a term found in the SQL Server docs. Check to see if this is a common term in the literature or if it is specific to SQL server.
[^2]: https://www.postgresql.org/docs/9.1/queries-with.html
