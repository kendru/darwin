use std::alloc::Layout;
use std::intrinsics::copy_nonoverlapping;
use std::mem::{align_of, size_of};
use std::ops::{Deref, DerefMut};
use std::slice;

use crate::entry::{
    EntryRef, PageEntry, ENTRY_REF_SIZE, PAGE_ENTRY_HEADER_ALIGN, PAGE_ENTRY_HEADER_SIZE,
};
use crate::page::{Allocation, Page, Pool, HEADER_SIZE};
use crate::util::pad_for;

#[derive(Debug, PartialEq, Eq, PartialOrd, Ord, Clone, Copy)]
#[repr(transparent)]
pub struct NodeId(u64);

pub(crate) enum Node<'a> {
    LeafNode(LeafNode<'a>),
    InnerNode(InnerNode<'a>),
}

impl<'a> Node<'a> {
    pub(crate) fn new_leaf(pool: &'a Pool) -> Node<'a> {
        Node::LeafNode(LeafNode::new(pool))
    }

    pub(crate) fn new_inner(pool: &'a Pool) -> Node<'a> {
        Node::InnerNode(InnerNode::new(pool))
    }
}

// TODO: Consider a locking scheme for nodes. Where should the lock live?
pub(crate) struct LeafNode<'a> {
    page: Option<Page>, // Always Some<Page> until dropped.
    pool: &'a Pool,
}

impl LeafNode<'_> {
    pub(crate) fn new<'a>(pool: &'a Pool) -> LeafNode<'a> {
        LeafNode {
            page: Some(pool.get()),
            pool,
        }
    }

    // TODO: add versions of search functions that take a key comparator
    // for more complex keys that do not have byte-wise comparison semantics.
    // We should leave the raw versions that perform byte comparisons so that
    // the client code can use this (faster) version whenever possible.

    pub(crate) fn find(&self, key: &[u8]) -> Option<&PageEntry> {
        self.find_entry(key).map(|(_, entry)| unsafe { &*entry })
    }

    pub(crate) fn scan(&self) -> PageEntryIter {
        PageEntryIter {
            node: self,
            offset: HEADER_SIZE,
            end: self.free_start() as usize,
        }
    }

    fn find_entry(&self, key: &[u8]) -> Option<(*mut EntryRef, *mut PageEntry)> {
        self.scan()
            .find(|(_entry_ref_ptr, entry_ptr)| {
                let entry = unsafe { &mut *(*entry_ptr as *mut PageEntry) };
                entry.key_len as usize == key.len() && &entry.data[0..entry.key_len as usize] == key
            })
            .map(|(entry_ref_ptr, entry_ptr)| {
                (entry_ref_ptr as *mut EntryRef, entry_ptr as *mut PageEntry)
            })
    }

    pub(crate) fn insert(&mut self, val_layout: Layout, key: &[u8], val: &[u8]) {
        debug_assert_eq!(val.len(), val_layout.size());

        if let Some((entry_ref_ptr, old_entry_ptr)) = self.find_entry(key) {
            self.insert_extend(val_layout, entry_ref_ptr, old_entry_ptr, val);
        } else {
            self.insert_initial(val_layout, key, val);
        }
    }

    fn insert_initial(&mut self, val_layout: Layout, key: &[u8], val: &[u8]) {
        // Initial entry holds the sized struct fields of PageEntry plus a data buffer large
        // enough to fit a single value.
        // TODO: Does the key need to be aligned?
        let key_len = key.len();

        // The data that each entry starts with is the key followed by enough bytes of padding
        // to align the values appropriately, followed by a single value.
        let initial_data_size = key_len
            + pad_for(PAGE_ENTRY_HEADER_SIZE + key_len, val_layout.align())
            + val_layout.size();

        let size = PAGE_ENTRY_HEADER_SIZE + initial_data_size;
        let Allocation {
            ptr: entry_ptr,
            offset: entry_start,
        } = self
            .alloc_end(Layout::from_size_align(size, PAGE_ENTRY_HEADER_ALIGN).unwrap())
            .expect("TODO: handle allocation failure for page");
        let entry = unsafe {
            // Reinterpret entry_ptr as a fat pointer to `initial_data_size` bytes, since this
            // is the dynamic portion of the PageEntry DST.
            let dst_ptr =
                slice::from_raw_parts_mut(entry_ptr as *mut u8, initial_data_size) as *mut [u8];
            &mut *(dst_ptr as *mut PageEntry)
        };

        // XXX: What if a key length cannot fit in a u16? Should the key be split across multiple pages and have the
        // length be only what is on this page?
        entry.key_len = key_len as u16;
        entry.val_count = 1;
        entry.data[0..key_len].copy_from_slice(key);
        let val_start = key_len + pad_for(PAGE_ENTRY_HEADER_SIZE + key_len, val_layout.align());
        entry.data[val_start..].copy_from_slice(val);

        let entry_ref = self
            .new_entry_ref(key)
            .expect("TODO: handle allocation failure for page");
        entry_ref.offset = entry_start;
        entry_ref.length = initial_data_size as u16;
    }

    // TODO: This does not order the records within the leaf node. In order to make scans simpler and lookups
    // more performant, we should order the records on insert.
    fn insert_extend(
        &mut self,
        val_layout: Layout,
        entry_ref_ptr: *mut EntryRef,
        old_entry_ptr: *mut PageEntry,
        val: &[u8],
    ) {
        let (entry_ref, old_entry) = unsafe { (&mut *entry_ref_ptr, &*old_entry_ptr) };
        // Copy old entry into new empty slot in data. The old
        // memory becomes dead and will eventually get garbage collected.
        let old_size = PAGE_ENTRY_HEADER_SIZE + old_entry.data.len();
        let new_size = old_size + val_layout.size();

        // 1. Allocate a new slot.
        let Allocation {
            ptr: entry_ptr,
            offset: entry_start,
        } = self
            .alloc_end(Layout::from_size_align(new_size, PAGE_ENTRY_HEADER_ALIGN).unwrap())
            .expect("TODO: handle allocation failure for page");

        // 2. Copy bytes from old slot into new slot.
        unsafe {
            copy_nonoverlapping(
                old_entry as *const PageEntry as *const u8,
                entry_ptr as *mut u8,
                old_size,
            );
        }

        // 3. Reinterpret pointer to new slot as &mut PageEntry.
        // Note that we have to create a sized slice and cast that as a pointer instead
        // of doing a naive cast like `&mut *(entry_ptr as *mut PageEntry)` because Rust
        // would use the size of the entire allocation as the length of the dynamic portion
        // of the DST in the resulting fat pointer, which is incorrect. The allocation
        // contains the entire PageEntry struct, but the fat pointer's length field should
        // exclude the dynamic fields.
        // TODO: Consider mitigating this by representing the `PageEntry` as a fat raw pointer
        // with a prefix that can be interpreted as a `PageEntryHeader`.
        let entry = unsafe {
            let sized_slice = slice::from_raw_parts_mut(
                entry_ptr as *mut u8,
                old_entry.data.len() + val_layout.size(),
            );
            &mut *(sized_slice as *mut [u8] as *mut PageEntry)
        };

        // 4. Update new slot: increment val_count and append to data.
        // NOTE [DATA ALIGNMENT]:
        // We pack values contiguously, potentially inserting padding between the key and
        // the start of the value array. This assumes that, like structs, the value size
        // is a multiple of the value alignment. As long as value_layout was constructed
        // safely, this is a sound assumption.
        let new_val_offset = old_entry.data.len();
        entry.val_count += 1;
        entry.data[new_val_offset..].copy_from_slice(val);

        // Update entry pointer.
        entry_ref.offset = entry_start as u16;
        entry_ref.length += val_layout.size() as u16;
    }

    fn new_entry_ref(&mut self, key: &[u8]) -> Option<&mut EntryRef> {
        // Find the first key < `key`, and insert to the left of it, shifting all other EntryRefs to the right.
        // If no higher key is found, allocate a new EntryRef at the end of the free space.
        // If no space left on the page, return None, and let the call site deal with allocation failure.

        for offset in (HEADER_SIZE..(self.free_start() as usize)).step_by(ENTRY_REF_SIZE) {
            let entry_ref_ptr: *mut EntryRef;
            let entry_ref = unsafe {
                entry_ref_ptr = self.offset_ptr_unchecked(offset, ENTRY_REF_SIZE) as *mut EntryRef;
                &mut *entry_ref_ptr
            };
            let entry = unsafe {
                let start = entry_ref.offset as usize;
                let length = entry_ref.length as usize;
                let ptr = self.offset_ptr_unchecked(start, length) as *mut PageEntry;
                &*ptr
            };
            // NOTE [KEY COMPARISONS]:
            // We are naively comparing these keys, which will not always provide a valid comparison function
            // for the key type. We should eventually allow a comparator to be passed in to all tree operations
            // and use that for determining ordering.
            if entry.key() > key {
                unsafe {
                    self.shift_start(offset, ENTRY_REF_SIZE)
                        .expect("TODO: handle page allocation error");
                }
                // entry_ref now references the dead entry that is where the first shifted element was.
                entry_ref.reset();
                return Some(entry_ref);
            }
        }

        // All existing entries have keys <= supplied key.
        let layout =
            Layout::from_size_align(size_of::<EntryRef>(), align_of::<EntryRef>()).unwrap();
        self.alloc_start(layout)
            .map(|Allocation { ptr, .. }| unsafe { &mut *(ptr as *mut u8 as *mut EntryRef) })
    }
}

// XXX: The Deref/DerefMut impls are just convenience for getting to the page more easily
// from within LeafNode. It probably does not make much sense
impl<'a> Deref for LeafNode<'a> {
    type Target = Page;
    fn deref(&self) -> &Self::Target {
        self.page.as_ref().unwrap()
    }
}

impl<'a> DerefMut for LeafNode<'a> {
    fn deref_mut(&mut self) -> &mut Self::Target {
        self.page.as_mut().unwrap()
    }
}

impl<'a> Drop for LeafNode<'a> {
    fn drop(&mut self) {
        // NOTE []:
        // Page should not get dropped, so return it to the pool, and replace it
        // with a None. When Rust gets better support for telling drocpck to not
        // drop some field, that method should be used instead. See <https://doc.rust-lang.org/nightly/nomicon/destructors.html>.
        self.pool.check_in(self.page.take().unwrap());
    }
}

pub(crate) struct PageEntryIter<'a> {
    node: &'a LeafNode<'a>,
    offset: usize,
    end: usize,
}

impl<'a> Iterator for PageEntryIter<'a> {
    type Item = (*const EntryRef, *const PageEntry);

    // Perform a linear scan over keys.
    // TODO: Allow this to be replaced by binary search.
    // Ideally, this could iterate in order, and we would place the EntryRef at the correct position on insert.
    fn next(&mut self) -> Option<Self::Item> {
        if self.offset >= self.end {
            return None;
        }
        let entry_ref_ptr: *mut EntryRef;
        let entry_ref = unsafe {
            entry_ref_ptr =
                self.node.offset_ptr_unchecked(self.offset, ENTRY_REF_SIZE) as *mut EntryRef;
            &mut *entry_ref_ptr
        };

        let start = entry_ref.offset as usize;
        let length = entry_ref.length as usize;
        let entry_ptr = self.node.offset_ptr_unchecked(start, length) as *mut PageEntry;

        self.offset += ENTRY_REF_SIZE;

        Some((entry_ref_ptr, entry_ptr))
    }
}

/// InnerEntry is the type of an entry that is added to an inner node.
/// However, when an entry is stored, the pointers and key will be stored
/// separately.
pub(crate) struct InnerEntry {
    pub(crate) left: *const u8,
    pub(crate) right: *const u8,
    pub(crate) key: [u8],
}

pub(crate) struct InnerNode<'a> {
    page: Option<Page>, // Always Some<Page> until dropped.
    pool: &'a Pool,
}

impl InnerNode<'_> {
    pub(crate) fn new<'a>(pool: &'a Pool) -> InnerNode<'a> {
        InnerNode {
            page: Some(pool.get()),
            pool,
        }
    }

    pub(crate) fn insert_entry(&mut self, entry: &InnerEntry) -> Option<()> {
        let page = self.page.as_mut().unwrap();
        let key_len = entry.key.len();

        let free_space_needed = size_of::<usize>() * 2  // left and right pointers
            + size_of::<EntryRef>() // reference to key
            + key_len; // key
        if (page.free_len() as usize) < free_space_needed {
            return None;
        }

        let key_layout = Layout::from_size_align(key_len, 1).unwrap();
        let ptr_layout = Layout::from_size_align(size_of::<usize>(), align_of::<usize>()).unwrap();
        let entry_ref_layout =
            Layout::from_size_align(size_of::<EntryRef>(), align_of::<EntryRef>()).unwrap();

        unsafe {
            // Append key.
            // TODO: Handle when key length is larger than u16::MAX.
            // TODO: Handle key alignment so that clients can perform an aligned read of some typed key without a copy.

            // SAFETY:
            // We have already ensured that the page has enough free space for all allocations made in this
            // block, so none of the allocations will fail or return overlapping memory regions.
            let Allocation {
                ptr: key_ptr,
                offset: key_offset,
            } = page.alloc_end_unchecked(key_layout);
            (&mut *(key_ptr)).copy_from_slice(&entry.key);

            // Allocate left pointer.
            let Allocation { ptr: left_ptr, .. } = page.alloc_start_unchecked(ptr_layout);
            *(left_ptr as *mut usize) = entry.left as usize;

            // Allocate EntryRef.
            let Allocation {
                ptr: entry_ref_ptr, ..
            } = page.alloc_start_unchecked(entry_ref_layout);
            let entry_ref = &mut *(entry_ref_ptr as *mut u8 as *mut EntryRef);
            entry_ref.offset = key_offset;
            entry_ref.length = entry.key.len() as u16;

            // Allocate right pointer.
            let Allocation { ptr: right_ptr, .. } = page.alloc_start_unchecked(ptr_layout);
            *(right_ptr as *mut usize) = entry.right as usize;
        }

        Some(())
    }
}

impl<'a> Drop for InnerNode<'a> {
    fn drop(&mut self) {
        self.pool.check_in(self.page.take().unwrap());
    }
}

#[cfg(test)]
mod tests {
    use std::mem::{align_of, size_of};

    use super::*;

    #[test]
    fn test_insert_into_leaf() {
        let pool = Pool::new();
        let val_size = size_of::<u64>();
        let val_align = align_of::<u64>();
        let val_layout = Layout::from_size_align(val_size, val_align).unwrap();
        let mut leaf_node = LeafNode::new(&pool);

        leaf_node.insert(val_layout, &[0, 1, 45, 23], &2345u64.to_le_bytes());
        {
            // TODO: The return type should expose a value iterator that is aware of the value size.
            let res = leaf_node.find(&[0, 1, 45, 23]).unwrap();
            assert_eq!(res.key(), &[0, 1, 45, 23]);
            assert_eq!(res.values(val_layout), vec![&2345u64.to_le_bytes()]);
        }

        leaf_node.insert(val_layout, &[0, 1, 45, 23], &4985355u64.to_le_bytes());
        {
            let res = leaf_node.find(&[0, 1, 45, 23]).unwrap();
            assert_eq!(
                res.values(val_layout),
                vec![&2345u64.to_le_bytes(), &4985355u64.to_le_bytes()]
            );
        }
    }
}
