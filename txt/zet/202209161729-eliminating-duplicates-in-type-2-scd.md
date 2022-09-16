---
tags: ["sql", "window-functions"]
created: Fri Sep 16 17:29:41 MDT 2022
---

# Eliminating Duplicates In Type 2 SCD table

Imagine that we have a type 2 slowly changing dimension table with as_of and
until columns bounding each row's period of validity.

```sql
create table t2scd (
    id int,
    "name" text,
    age int,
    as_of int,
    until int
);
```

Now imagine that there was a bug in the process that populated this table such
that there were duplicates.

```sql
insert into t2scd (id, "name", age, as_of, until)
values
   (1, 'andrew', 32, 1, 2),    -- ok
   (1, 'andrew', 33, 2, 3),    -- ok
   (1, 'andrew', 33, 3, 4),    -- dup!
   (1, 'andrew', 33, 4, 5),    -- dup!
   (1, 'andrew', 33, 5, 6),    -- dup!
   (1, 'andrew', 34, 6, null); -- ok
```

Note that there were three consecutive rows where the table data did not change
from version to version. What we want is to collapse those rows into a single
row that looks like this: `(1, 'andrew', 33, 2, 6)`.

We can accomplish this by using a window function that partitions over all table
data (that is, all columns except `as_of` and `until`) and orders by `as_of`. We
then emit only the first row for each partition, taking the maximum `until`
value. Since the partition is ordered by `as_of`, the resulting row's validity
period will span from the earliest row's `as_of` until the latest row's `until`.

```sql
with runs as (
    select
        id, "name", age, as_of,
        max(until) over w as until,
        row_number() over w as row_num
    from t2scd
    window w as (
        partition by id, "name", age
        order by as_of asc
        rows between current row and unbounded following
    )
)
select
    id, "name", age, as_of, until
from runs
where
    row_num = 1;
```

This results in exactly the data we want:

```
 id |  name  | age | as_of | until
----+--------+-----+-------+-------
  1 | andrew |  32 |     1 |     2
  1 | andrew |  33 |     2 |     6
  1 | andrew |  34 |     6 |
```

