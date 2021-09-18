---
tags: ["db-internals","virtual-machines"]
created: Wed Sep 15 15:44:24 MDT 2021
---

# SQLite Virtual Machine

The SQLite VM defines the data format for records[^1].

A record has the following format:

```
|-------------|-----------|
| Header Size | Type Tags |
|-------------|-----------|
```

followed by the data for the record. Since each record has type tags, the schema is not used for decoding a record. This means that there is some space overhead of a fixed length (assuming that the schema is uniform) for each record.

[^1]: https://www.youtube.com/watch?v=gpxnbly9bz4&t=2850
