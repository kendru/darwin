---
tags: ["db-internals"]
created: Wed Sep 15 15:27:40 MDT 2021
---

# In Memory Database Distinctives

## Storage Format

- For an in-memory db, do we need to have page-oriented structures, or is it preferrable to maintain a heap within a single byte array or other variable-sized data structure?
- It is not necessary to map pages to data that can be serialized to a fixed-size disk page.

### Pointers to Records

In CMU system, the entry stored in the tree is a 64-bit pointer that contains both the start of the block (44-bit) and the offset within the block (20-bit)[^1]. They use fixed-size records so that pointer arithmetic is simple. This makes it appear that the offset is a row offset rather than a byte offset. Is there any advantage to this scheme, or would it work to have variable-length records and have the offset be a byte offset? Variable length data is _all_ stored in a variable length data "pool", and fields in the fixed-length records store a 64-bit pointer to the start of the variable-length data. This sounds like TOAST from Postgres.

In CMU system, the schema is external to each record, so the schema does not need to be included in a row header, and each row can be a fixed length[^2]. Additionally, since the for header is a fixed size, a given field of a given tuple can be given as `blockStart+rowOffset+headerSize+sizeBeforeField`.

When storing variable-length data, it can be helpful to include a fixed-length prefix of the data along with the pointer to the full variable length data so that prefix (and short equality) matches on that column can ofen be done without the indirection of following the pointer. Possible to store hash instead of prefix so that any equality match works, but prefix scans still require the indirect.

## Type Representation

Can often use native types of programming language for integers. Endianness may need to be considered if we are comparing bytes directly.

See HyPer for an example of storage and manipulation of fixed-point decimals - better than Postgres.

It is important to word-align values to avoid multiple reads (or even errors in older systems). There are profilers that warn on unaligned accesses.

## Access Patterns

There is almost always "hot" data that is still receiving updates and "cold" data that is essentially static. This indicates that it may be advantageous to keep data in a row store until is is "cold" enough to be moved to a column store. You still need to deal with updates to cold data, even though it should be infrequent and probably does not need to be optimal. Pavlo believes that this is the wrong approach and that column stores can be used for transactional workloads as well[^3].

For a hybrid approach, do you maintain 2 storage managers and make the execution engine aware of how to query both, or do you maintain a single storage manager that can locate and access tuples in either rows or columns?

Separate engines:
- Fractured Mirrors (Oracle, IBM)
    - All R/W go to row store
    - Asynchronously, copy new rows to DSM
    - Fulfil OLAP queries against DSM, optionally unioning with new data in row store.
    - Oracle's column store is _ephemeral_.
- Delta Store (previously, SAP HANA)
    - Less work if using MVCC

## Indexes

BW Tree came out of Microsoft's Hekaton project. Motivation: B+ Tree-like structure, but latch-free.

[^1]: https://youtu.be/y6qFHu0YKMM?t=235
[^2]: https://youtu.be/y6qFHu0YKMM?t=390
[^3]: https://youtu.be/y6qFHu0YKMM?t=3600
