use std::mem::size_of;

// TODO: Find a home for this comment.
// Page implements an index or table page that contains keys with multiple associated values.
// The page itself stores key pointers from the beginning of the data array and entries from
// the end of the array. The key pointers are fixed-size logical pointers that are offsets
// to where the corresponding entry starts.
//
// Since the keys inside an index page are never referenced directly by a pointer or any other
// physical identifier, the entry pointers can be rearranged to keep them ordered by keys.
// TODO: we should benchmark the performance of keeping keys ordered versus keeping entries internal to
// a page unordered and always performing a linear scan (and potentially a sort).
// Another option would be keeping elements unsorted but including a "next element" pointer
// like InnoDB does. Again, these are things that need thorough benchmarking before we settle on an
// option.



pub(super) const ENTRY_REF_SIZE: usize = size_of::<EntryRef>();
// Size of the fixed port
pub(super) const PAGE_ENTRY_HEADER_SIZE: usize = size_of::<u16>() * 2;


#[derive(Debug)]
#[repr(C)]
pub(crate) struct EntryRef {
    pub(crate) offset: u16,
    pub(crate) length: u16,
}

// TODO: Embed the header in an entry like we do with `page::header::Header` in `page::Page`.
// Possible optimization: allocate some fixed number of value slots for each entry and
// keep track of how many are free. This way, we minimize the frequency of allocating
// new slots.
#[derive(Debug)]
#[repr(packed, C)]
pub struct PageEntry {
    pub(crate) key_len: u16,
    pub(crate) val_count: u16, // TODO: Do we need to know the length of each element? The total length?
    pub(crate) data: [u8],
}

impl PageEntry {
    pub(crate) fn key(&self) -> &[u8] {
        &self.data[0..(self.key_len as usize)]
    }

    // TODO: Create a ValueIterator type and make the primitive PageEntry::values_iter().
    pub(crate) fn values(&self, val_len: usize) -> Vec<&[u8]> {
        let start = (self.key_len + self.key_len % 8) as usize;
        let mut values = Vec::with_capacity(self.val_count as usize);
        for offset in (start..self.data.len()).step_by(val_len) {
            values.push(&self.data[offset..(offset+val_len)]);
        }

        values
    }
}
