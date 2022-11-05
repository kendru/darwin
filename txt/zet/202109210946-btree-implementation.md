---
tags: ["btree","data-structures","db-internals"]
created: Tue Sep 21 09:46:48 MDT 2021
---

# BTree Implementation

Rows are identified by RID (Row ID)/TID (Tuple ID), which is a logical pointer to some location in a heap page. In this design, some storage manager needs to own the heap pages and dereference the RID.

Layout:

- Metadata
- Key map
- Entries `(key, [value])` - variable-length

### InnoDB

InnoDB adds a next record pointer to each record inside a leaf node so that records can be inserted in any order, but once the starting record has been found, the remaining records within the page can be traversed in order by following these pointers.
