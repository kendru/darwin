/// Header describes a btree page in memory.
/// It is assumed that the data described by the page resides immediately after
/// the page in memory. It is a low level struct that nodes are built on.
pub const Header = packed struct {
    next: ?*Header = null,
    freeStart: u16,
    freeLen: u16,
};

/// ItemRef is the type that is appended to the beginning of the page and
/// references data that is stored elsewhere in the page.
/// offset points to the start of the data item as the number of bytes from
/// the start of the page data (not including the header), and len indicates
/// how many bytes the item occupies.
pub const ItemRef = packed struct {
  offset: u16,
  len: u16,
};

// Cr
