---
tags: ["scm", "sqlite"]
created: Wed Dec  9 11:38:28 MST 2020
---

# Fossil SCM Design

There are several interesting features about the [Fossil](https://www.fossil-scm.org/) version control system. It is built on top of SQLite and is used to track the development of [SQLite](https://sqlite.org/src/doc/trunk/README.md) as well as [Fossil itself](https://www.fossil-scm.org/home/doc/trunk/www/webui.wiki). It makes use of the graph querying features of SQLite to provide common SCM functionality[^1]. The graph functionality is provided in the form of recursive CTEs. The general pattern for graph traversal is as follows:

```sql
CREATE TABLE node(
  id INTEGER PRIMARY KEY,
  title TEXT
);
CREATE TABLE edge(
  efrom INTEGER NOT NULL REFERENCES node(id),
  eto INTEGER NOT NULL REFERENCES node(id),
  PRIMARY KEY(efrom, eto)
);
INSERT INTO node(id, title) VALUES
  (1, 'andrew'),
  (2, 'stephen'),
  (3, 'abel'),
  (4, 'jonah'),
  (5, 'ken');
INSERT INTO edge(efrom, eto) VALUES
  (5, 2),
  (2, 1),
  (1, 3),
  (1, 4);

WITH RECURSIVE link(id, title, depth) AS (
  -- base case: start @ node 2
  SELECT id, title, 0 FROM node WHERE id = 2
  UNION
  -- recursive case: given a link representing the FROM side,
  -- find the next links representing the TO side
  SELECT edge.eto, node.title, link.depth + 1
  FROM link
  JOIN edge ON link.id = edge.efrom
  JOIN node ON edge.eto = node.id
)
SELECT * FROM link;
```

This is equivalent to a Datalog program:

```prolog
node(1, "andrew").
node(2, "stephen").
node(3, "abel").
node(4, "jonah").
node(5, "ken").
edge(5, 2).
edge(2, 1).
edge(1, 3).
edge(1, 4).

link(X, X).               % include the starting node
link(X, Y) :- edge(X, Y). % base case: link exists
link(X, Y) :- edge(X, Z), link(Z, Y), Z \= Y. % transitive closure

%% Query the graph
link(2, Id),
node(Id, Title).
```

[^1]: https://sqlite.org/lang_with.html#rcex3 references a query used by Fossil to get the `n` most recent ancestors of a checkin (the Fossil term for a commit).
