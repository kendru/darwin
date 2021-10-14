use std::{alloc::Layout, mem::{align_of, size_of}};

use crate::{page::Page, util::{pad_for, round_to}};

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
// Size of the statically sized portion of PageEntry.
pub(super) const PAGE_ENTRY_HEADER_SIZE: usize = size_of::<u16>() * 2;
pub(super) const PAGE_ENTRY_HEADER_ALIGN: usize = align_of::<u16>();


#[repr(C)]
pub(crate) struct EntryRef {
    pub(crate) offset: u16,
    pub(crate) length: u16,
}

impl EntryRef {
    pub(crate) fn new(offset: u16, length: u16) -> EntryRef {
        EntryRef { offset, length }
    }


    pub(crate) fn reset(&mut self) {
        self.offset = 0;
        self.length = 0;
    }
}


// TODO: Embed the header in an entry like we do with `page::header::Header` in `page::Page`.
// Possible optimization: allocate some fixed number of value slots for each entry and
// keep track of how many are free. This way, we minimize the frequency of allocating
// new slots.
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

    pub(crate) fn values_iter(&self, val_layout: Layout) -> ValuesIterator {
        let start_offset = self.key_len as usize + pad_for(PAGE_ENTRY_HEADER_SIZE + self.key_len as usize, val_layout.align());
        ValuesIterator {
            layout: val_layout,
            data: &self.data,
            offset: start_offset,
        }
    }
}

pub struct ValuesIterator<'a> {
    layout: Layout,
    data: &'a [u8],
    offset: usize,
}

impl<'a> Iterator for ValuesIterator<'a> {
    type Item = &'a [u8];

    fn next(&mut self) -> Option<Self::Item> {
        let size = self.layout.size();
        let offset = self.offset;
        let end = offset+size;
        if end > self.data.len() {
            return None;
        }

        self.offset += size;
        Some(&self.data[offset..end])
    }

    fn size_hint(&self) -> (usize, Option<usize>) {
        let size = self.data.len() / self.layout.size();
        (size, Some(size))
    }
}
