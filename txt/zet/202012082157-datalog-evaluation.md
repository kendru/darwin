---
tags: ["databases", "datalog"]
created: Tue Dec  8 21:57:20 MST 2020
---

# Datalog Evaluation

## Description of Datalog

A datalog _rule_ has the following form:

```
p(a1, ..., ak) :- p1(a11, ..., a1k), ..., ph(ah1, ..., ahk).
```

where `p`, `p1`, etc. through `ph` are fixed-arity _predicates_ and each `a` is is either a constant or variable argument. The combination of a predicate and its arguments is an _atom_. A _fact_ is a rule with only a _head_ (the portion before the `:-`), which is only allowed when if the arguments are all constants. For example, the following is a fact:

```
parent(stephen, andrew).
```

On the other hand, a query is an atom<!-- is the query an entire rule, or are individual atoms rules? --> prefixed by `?-` that may contain constants as well as variables.
