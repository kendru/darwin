---
tags: ["graph","indexing"]
created: Mon Mar 15 11:46:37 MDT 2021
---

# Multi Valued Attributes

There is an equivalence between allowing an entity to have multi-valued attributes and supporting non-unique indexes for entities. A non-unique index means that the EAV table may legally have multiple entries for some common `(E, A)` prefix. Similarly, allowing multi-valued attributes means that the AVE table may have multiple entries for a common `(AV)` prefix. It is possible to encode multiple values in a single array/list property, but the advantage of using a multi-valued property comes when storing foreign keys, in which case we can perform a join without doing any additional decoding of values. Also, we can use the same limiting/pagination mechanisms to fetch a subset of references when the full set of references is potentially very large (consider the followers set for a celebrity Twitter account).

Additionally, by encoding foreign keys as a multi-valued attribute, it is trivial to traverse the "reverse" relationship, making explicit join tables unnecessary[^1].

[^1]: This is not true if there is metadata about the relationship that needs to be maintained. In the Twitter example, tracking the follow timestamp of a "follows" relationship would require either properties on edges (which cannot be represented as a simple foreign key) or an explicit join table. Question: could this be solved by storing quadruplets and keeping additional metadata in the named graph/transaction?
