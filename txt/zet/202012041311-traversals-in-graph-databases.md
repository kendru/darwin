---
tags: ["graphs", "rfd"]
created: Fri Dec  4 13:11:55 MST 2020
---

# Traversals in Graph Databases

The Gremlin Graph Traversal Machine[][bib:rodriguez2015] uses a high-level model where a _traversal_ **Ψ** contains a set of _traversers_ **T**. The traversers navigate through a _graph_ **G**. The result of Ψ is the set of locations where all T have stopped. This is a similar execution model to the [specter](https://github.com/redplanetlabs/specter) library's _navigator_. For specter, a navigator describes a traversal into a tree-like data structure, but the navigator _selects_ all elements where the traversal terminates.
