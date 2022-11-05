use std::alloc::Layout;
use std::convert::TryInto;
use std::intrinsics::copy_nonoverlapping;
use std::mem::{self, align_of, size_of};
use std::ops::{Deref, DerefMut};
use std::ptr::NonNull;
use std::{ptr, slice};

use crate::entry::{EntryRef, PageEntry, PAGE_ENTRY_HEADER_ALIGN, PAGE_ENTRY_HEADER_SIZE};
use crate::page::{Allocation, Header, Page, Pool};
use crate::util::{pad_for, round_to};

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

    // Does it make sense to extract find/insert to a trait?
    pub(crate) fn find(&self, key: &[u8]) -> Option<&PageEntry> {
        match self {
            Node::LeafNode(n) => n.find(key),
            Node::InnerNode(n) => n.find(key),
        }
    }

    pub(crate) fn insert(&mut self, pool: &Pool, val_layout: Layout, key: &[u8], val: &[u8]) {
        // Should this be refactored to use polymorphic methods?
        match self {
            Node::LeafNode(n) => {
                match n.insert(val_layout, key, val) {
                    Some(_) => {}
                    None => {
                        let old_node: LeafNode =
                            mem::replace(self, Node::new_inner(pool))
                                .try_into()
                                .unwrap();
                        let (left, right) = old_node.split(val_layout);
                        // TODO: Is this the correct behaviour if the left node only has the capacity for a single entry?
                        let pivot = right
                            .scan()
                            .next()
                            .map(|(_, ptr)| {
                                let entry = unsafe { &*ptr };
                                entry.key().to_vec()
                            })
                            .unwrap_or_else(|| key.to_vec());

                        self
                            .as_inner_mut()
                            .insert_entry(InnerEntry {
                                left: &left as *const LeafNode as *const u8,
                                right: &right as *const LeafNode as *const u8,
                                key: pivot,
                            })
                            .expect("Inner node must have capacity for entries after split");
                    }
                }
            }
            Node::InnerNode(n) => {
                match n.insert(val_layout, key, val) {
                    Some(_) => {},
                    None => {
                        let old_node: InnerNode =
                            mem::replace(self, Node::new_inner(pool))
                                .try_into()
                                .unwrap();
                        let (left, right) = old_node.split(val_layout);
                    }
                }
            }
        }




        match self {
            Node::LeafNode(n) => n.insert(val_layout, key, val),
            Node::InnerNode(n) => n.insert(val_layout, key, val),
        }
    }

    pub(crate) fn as_inner_mut(&mut self) -> &mut InnerNode<'a> {
        match self {
            Node::InnerNode(n) => n,
            _ => {
                panic!("Caller assumes that node is InnerNode")
            }
        }
    }

    pub(crate) fn as_leaf_mut(&mut self) -> &mut LeafNode<'a> {
        match self {
            Node::LeafNode(n) => n,
            _ => {
                panic!("Caller assumes that node is LeafNode")
            }
        }
    }
}

impl<'a> TryInto<InnerNode<'a>> for Node<'a> {
    type Error = ();

    fn try_into(self) -> Result<InnerNode<'a>, Self::Error> {
        match self {
            Node::InnerNode(n) => Ok(n),
            _ => Err(()),
        }
    }
}

impl<'a> TryInto<LeafNode<'a>> for Node<'a> {
    type Error = ();

    fn try_into(self) -> Result<LeafNode<'a>, Self::Error> {
        match self {
            Node::LeafNode(n) => Ok(n),
            _ => Err(()),
        }
    }
}

// TODO: Consider a locking scheme for nodes. Where should the lock live?
pub(crate) struct LeafNode<'a> {
    page: Option<Page>, // Always Some<Page> until dropped.
    pool: &'a Pool,
}

impl<'a> LeafNode<'a> {
    pub(crate) fn new(pool: &'a Pool) -> LeafNode {
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
            ref_iter: self.scan_entry_refs(),
        }
    }

    fn scan_entry_refs(&self) -> EntryRefIter {
        EntryRefIter {
            page: self.page.as_ref().unwrap(),
            offset: size_of::<Header>(),
            end: self.free_start() as usize,
        }
    }

    fn find_entry(&self, key: &[u8]) -> Option<(*mut EntryRef, *mut PageEntry)> {
        // TODO: Allow this linear scan to be replaced by binary search.
        self.scan()
            .find(|(_entry_ref_ptr, entry_ptr)| {
                let entry = unsafe { &mut *(*entry_ptr as *mut PageEntry) };
                entry.key_len as usize == key.len() && &entry.data[0..entry.key_len as usize] == key
            })
            .map(|(entry_ref_ptr, entry_ptr)| {
                (entry_ref_ptr as *mut EntryRef, entry_ptr as *mut PageEntry)
            })
    }

    pub(crate) fn insert(&mut self, val_layout: Layout, key: &[u8], val: &[u8]) -> Option<()> {
        debug_assert_eq!(val.len(), val_layout.size());

        if let Some((entry_ref_ptr, old_entry_ptr)) = self.find_entry(key) {
            self.insert_extend(val_layout, entry_ref_ptr, old_entry_ptr, val)
        } else {
            self.insert_initial(val_layout, key, val, 1)
        }
    }

    fn insert_initial(
        &mut self,
        val_layout: Layout,
        key: &[u8],
        vals: &[u8],
        val_count: u16,
    ) -> Option<()> {
        debug_assert_eq!(vals.len() / val_count as usize, val_layout.size());
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
        } = self.alloc_end(Layout::from_size_align(size, PAGE_ENTRY_HEADER_ALIGN).unwrap())?;
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
        entry.val_count = val_count;
        entry.data[0..key_len].copy_from_slice(key);
        let val_start = key_len + pad_for(PAGE_ENTRY_HEADER_SIZE + key_len, val_layout.align());
        entry.data[val_start..].copy_from_slice(vals);

        let entry_ref = self.new_entry_ref(key)?;
        entry_ref.offset = entry_start;
        entry_ref.length = initial_data_size as u16;

        Some(())
    }

    // TODO: This does not order the records within the leaf node. In order to make scans simpler and lookups
    // more performant, we should order the records on insert.
    fn insert_extend(
        &mut self,
        val_layout: Layout,
        entry_ref_ptr: *mut EntryRef,
        old_entry_ptr: *mut PageEntry,
        val: &[u8],
    ) -> Option<()> {
        let (entry_ref, old_entry) = unsafe { (&mut *entry_ref_ptr, &*old_entry_ptr) };
        // Copy old entry into new empty slot in data. The old
        // memory becomes dead and will eventually get garbage collected.
        let old_size = PAGE_ENTRY_HEADER_SIZE + old_entry.data.len();
        let new_size = old_size + val_layout.size();

        // 1. Allocate a new slot.
        let Allocation {
            ptr: entry_ptr,
            offset: entry_start,
        } = self.alloc_end(Layout::from_size_align(new_size, PAGE_ENTRY_HEADER_ALIGN).unwrap())?;

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

        Some(())
    }

    pub(crate) fn split(mut self, val_layout: Layout) -> (LeafNode<'a>, LeafNode<'a>) {
        // Compact the node first so that contiguous entry refs point to contiguous entries.
        // This allows us to simply copy two blocks of memory to the new node and update each
        // entry ref's offset by some constant that is the delta between this node's free_end
        // and the new node's free_end.
        self.compact();

        let mut right = LeafNode::new(self.pool);

        // Move half of the keys to `right`.
        // TODO: Try to divide in half by entry size rather than key count.
        let left_count = self.entry_count() / 2;

        // Iterate the entries to be copied, and insert them into the new page.
        // Note that we cannot simply copy the memory for the EntryRefs and PageEntries
        // to the new page then update the offsets of each EntryRef because each PageEntry
        // needs to be placed such that the values array is properly aligned. Each PageEntry
        // in the current table has some amount of padding between the key and the values
        // array to ensure proper alignment, and we cannot maintain that alignment with a
        // simple memory copy.
        for (_, entry_ptr) in self.scan().skip(left_count) {
            let entry = unsafe { &*entry_ptr };
            right
                .insert_initial(
                    val_layout,
                    entry.key(),
                    entry.values_buffer(val_layout),
                    entry.val_count,
                )
                .expect("New right node must have capacity to fit half of left node's entries");
        }

        // Reset the pointers for the left page to point to only the first half of the
        // prior contents.
        let &EntryRef {
            offset: last_entry_offset,
            length: last_entry_length,
        } = self.scan_entry_refs().last().unwrap();
        let header = self.header_mut();
        header.free_start = (size_of::<Header>() + left_count * size_of::<EntryRef>()) as u16;
        header.free_end = last_entry_offset + last_entry_length;

        (self, right)
    }

    pub(crate) fn entry_count(&self) -> usize {
        (self.free_start() as usize - size_of::<Header>()) / size_of::<EntryRef>()
    }

    /// Compacts the entries such that:
    /// 1. Old entries that do not have an EntryRef pointing to them are removed.
    /// 2. Entries are re-organized such that they grow from the end of the page
    ///    as their corresponding EntryRefs grow from the beginning of the page.
    /// In order to simplify the process, we allocate a new page, copy the data
    /// over, and replace the current page with the new one.
    fn compact(&mut self) {
        let mut new_page = self.pool.get();
        let entry_ref_layout = EntryRef::layout();
        unsafe {
            for (entry_ref, entry) in self.scan() {
                // Copy PageEntry.
                let entry_size = PAGE_ENTRY_HEADER_SIZE + (&*entry_ref).length as usize;
                let Allocation {
                    ptr: entry_ptr,
                    offset: entry_offset,
                } = new_page.alloc_end_unchecked(Layout::from_size_align_unchecked(
                    entry_size,
                    PAGE_ENTRY_HEADER_ALIGN,
                ));
                ptr::copy_nonoverlapping(entry as *const u8, entry_ptr as *mut u8, entry_size);

                // Copy EntryRef.
                let Allocation {
                    ptr: entry_ref_ptr, ..
                } = new_page.alloc_start_unchecked(entry_ref_layout);
                ptr::copy_nonoverlapping(entry_ref, entry_ref_ptr as *mut EntryRef, 1);

                // Update offset to point to new entry.
                (&mut *(entry_ref_ptr as *mut EntryRef)).offset = entry_offset;
            }
        }
        // Swap in new compacted page, and return the fragmented one to the pool.
        let old_page = mem::replace(self.page.as_mut().unwrap(), new_page);
        self.pool.check_in(old_page);
    }

    fn new_entry_ref(&mut self, key: &[u8]) -> Option<&mut EntryRef> {
        // Find the first key < `key`, and insert to the left of it, shifting all other EntryRefs to the right.
        // If no higher key is found, allocate a new EntryRef at the end of the free space.
        // If no space left on the page, return None, and let the call site deal with allocation failure.

        for offset in
            (size_of::<Header>()..(self.free_start() as usize)).step_by(size_of::<EntryRef>())
        {
            let entry_ref_ptr: *mut EntryRef;
            let entry_ref = unsafe {
                entry_ref_ptr =
                    self.offset_ptr_unchecked(offset, size_of::<EntryRef>()) as *mut EntryRef;
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
                    self.shift_start(offset, size_of::<EntryRef>())?;
                }
                // entry_ref now references the dead entry that is where the first shifted element was.
                entry_ref.reset();
                return Some(entry_ref);
            }
        }

        // All existing entries have keys <= supplied key.
        self.alloc_start(EntryRef::layout())
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
    ref_iter: EntryRefIter<'a>,
}

impl<'a> Iterator for PageEntryIter<'a> {
    type Item = (*const EntryRef, *const PageEntry);

    // Perform a linear scan over keys.
    // Ideally, this could iterate in order, and we would place the EntryRef at the correct position on insert.
    fn next(&mut self) -> Option<Self::Item> {
        self.ref_iter.next().map(|entry_ref| {
            let start = entry_ref.offset as usize;
            let length = entry_ref.length as usize;
            let entry_ptr = unsafe {
                self.ref_iter.page.offset_ptr_unchecked(start, length) as *const PageEntry
            };

            (entry_ref as *const EntryRef, entry_ptr)
        })
    }
}

pub(crate) struct EntryRefIter<'a> {
    page: &'a Page,
    offset: usize,
    end: usize,
}

impl<'a> Iterator for EntryRefIter<'a> {
    type Item = &'a EntryRef;

    fn next(&mut self) -> Option<Self::Item> {
        if self.offset >= self.end {
            return None;
        }
        let entry_ref = unsafe {
            let entry_ref_ptr = self
                .page
                .offset_ptr_unchecked(self.offset, size_of::<EntryRef>())
                as *mut EntryRef;
            &*entry_ref_ptr
        };

        self.offset += size_of::<EntryRef>();

        Some(entry_ref)
    }
}

/// InnerEntry is the type of an entry that is added to an inner node.
/// However, when an entry is stored, the pointers and key will be stored
/// separately.
pub(crate) struct InnerEntry {
    pub(crate) left: *const u8,
    pub(crate) right: *const u8,
    pub(crate) key: Vec<u8>,
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

    pub(crate) fn insert_entry(&mut self, entry: InnerEntry) -> Option<()> {
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
                ptr: entry_ref_ptr,
                ..
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

    pub(crate) fn split<'a>(mut self, val_layout: Layout) -> (InnerNode<'a>, InnerNode<'a>) {
        let mut right = InnerNode::new(self.pool);

        // Move half of the keys to `right`.
        let page = self.page.as_mut().unwrap();
        let entry_count = (page.header().free_start as usize - size_of::<Header>()) / size_of::<EntryRef>();
        let left_count = entry_count / 2;

        // Iterate the entries to be copied, and insert them into the new page.
        // Note that we cannot simply copy the memory for the EntryRefs and PageEntries
        // to the new page then update the offsets of each EntryRef because each PageEntry
        // needs to be placed such that the values array is properly aligned. Each PageEntry
        // in the current table has some amount of padding between the key and the values
        // array to ensure proper alignment, and we cannot maintain that alignment with a
        // simple memory copy.
        for (_, entry_ptr) in self.scan().skip(left_count) {
            let entry = unsafe { &*entry_ptr };
            right
                .insert_initial(
                    val_layout,
                    entry.key(),
                    entry.values_buffer(val_layout),
                    entry.val_count,
                )
                .expect("New right node must have capacity to fit half of left node's entries");
        }

        // Reset the pointers for the left page to point to only the first half of the
        // prior contents.
        let &EntryRef {
            offset: last_entry_offset,
            length: last_entry_length,
        } = self.scan_entry_refs().last().unwrap();
        let header = self.header_mut();
        header.free_start = (size_of::<Header>() + left_count * size_of::<EntryRef>()) as u16;
        header.free_end = last_entry_offset + last_entry_length;

        (self, right)
    }

    pub(crate) fn find(&self, key: &[u8]) -> Option<&PageEntry> {
        self.smallest_child_gte(key)
            .and_then(|child| child.find(key))
    }

    pub(crate) fn insert(&mut self, val_layout: Layout, key: &[u8], val: &[u8]) {
        if let Some(child) = self.smallest_child_gte_mut(key) {
            child.insert(self.pool, val_layout, key, val);
        }
    }

    fn smallest_child_gte(&self, key: &[u8]) -> Option<&Node> {
        self.smallest_child_gte_ptr(key).map(|ptr| unsafe { &*ptr })
    }

    fn smallest_child_gte_mut(&self, key: &[u8]) -> Option<&mut Node> {
        self.smallest_child_gte_ptr(key)
            .map(|ptr| unsafe { &mut *(ptr as *mut Node) })
    }

    fn smallest_child_gte_ptr(&self, key: &[u8]) -> Option<*const Node> {
        let page = self.page.as_ref().unwrap();

        // Calculate offsets of components of each entry
        let left_ptr_offset = 0usize;
        let entry_ref_offset = round_to(left_ptr_offset+size_of::<usize>(), align_of::<EntryRef>());
        let right_ptr_offset =  round_to(entry_ref_offset+size_of::<EntryRef>(), align_of::<usize>());
        let total_item_size = right_ptr_offset + size_of::<usize>();

        // Range from the start of allocatable memory to the end of the memory allocated at
        // the beginning of the page.
        let start = round_to(size_of::<Header>(), align_of::<usize>());
        let end = page.free_start() as usize - total_item_size;

        let mut last_right_child = None;
        unsafe {
            // The memory that we look at in each iteration overlaps the previous iteration
            // such that the previous right pointer becomes the current right pointer.
            for base_offset in (start..=end).step_by(right_ptr_offset) {
                let left_ptr = page.offset_ptr_unchecked(base_offset+left_ptr_offset, size_of::<usize>()) as *const usize;
                let entry_ref_ptr = page.offset_ptr_unchecked(base_offset+entry_ref_offset, size_of::<EntryRef>()) as *const EntryRef;
                let right_ptr = page.offset_ptr_unchecked(base_offset+right_ptr_offset, size_of::<usize>()) as *const usize;

                let entry_ref = &*entry_ref_ptr;
                let entry_key = page.offset_ptr_unchecked(entry_ref.offset as usize, entry_ref.length as usize) as *const [u8];
                if key < &*entry_key {
                    return Some(*left_ptr as *const Node);
                }

                last_right_child = Some(*right_ptr as *const Node);
            }
        }

        last_right_child
    }
}

impl<'a> Drop for InnerNode<'a> {
    fn drop(&mut self) {
        self.pool.check_in(self.page.take().unwrap());
    }
}

#[cfg(test)]
mod tests {
    use std::{
        convert::TryInto,
        mem::{align_of, size_of},
    };

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
            assert_eq!(
                res.values_iter(val_layout).collect::<Vec<_>>(),
                vec![&2345u64.to_le_bytes()]
            );
        }

        leaf_node.insert(val_layout, &[0, 1, 45, 23], &4985355u64.to_le_bytes());
        {
            let res = leaf_node.find(&[0, 1, 45, 23]).unwrap();
            assert_eq!(
                res.values_iter(val_layout).collect::<Vec<_>>(),
                vec![&2345u64.to_le_bytes(), &4985355u64.to_le_bytes()]
            );
        }
    }

    #[test]
    fn test_entry_count() {
        let pool = Pool::new();
        let val_layout = Layout::from_size_align(1, 1).unwrap();
        let mut leaf_node = LeafNode::new(&pool);

        leaf_node.insert(val_layout, &[1], &[1]);
        assert_eq!(1, leaf_node.entry_count());

        // Adding a new key should create a new entry.
        leaf_node.insert(val_layout, &[2], &[2]);
        assert_eq!(2, leaf_node.entry_count());

        // Inserting a new value for an existing key should not create a new entry (only garbage to clean up later).
        leaf_node.insert(val_layout, &[1], &[3]);
        assert_eq!(2, leaf_node.entry_count());
    }

    #[test]
    fn test_compact() {
        let pool = Pool::new();
        let val_layout = Layout::from_size_align(8, 4).unwrap();
        let mut leaf_node = LeafNode::new(&pool);

        leaf_node.insert(val_layout, "key 1".as_bytes(), &123u64.to_le_bytes());
        leaf_node.insert(val_layout, "key 1".as_bytes(), &456u64.to_le_bytes());
        leaf_node.insert(val_layout, "key 1".as_bytes(), &789u64.to_le_bytes());

        leaf_node.insert(val_layout, "other key".as_bytes(), &81235u64.to_le_bytes());

        let initial_free = leaf_node.free_len();
        let vals_for_key_1 = get_u64_values_for_key(&leaf_node, val_layout, "key 1".as_bytes());

        leaf_node.compact();

        assert!(leaf_node.free_len() > initial_free);
        assert_eq!(
            vals_for_key_1,
            get_u64_values_for_key(&leaf_node, val_layout, "key 1".as_bytes()),
        )
    }

    #[test]
    fn test_node_split() {
        let pool = Pool::new();
        let val_layout = Layout::from_size_align(8, 4).unwrap();
        let mut leaf_node = LeafNode::new(&pool);

        leaf_node.insert(val_layout, "key 1".as_bytes(), &123u64.to_le_bytes());
        leaf_node.insert(val_layout, "key 2".as_bytes(), &456u64.to_le_bytes());

        let vals_for_key_1 = get_u64_values_for_key(&leaf_node, val_layout, "key 1".as_bytes());
        let vals_for_key_2 = get_u64_values_for_key(&leaf_node, val_layout, "key 2".as_bytes());

        let (fst, snd) = leaf_node.split(val_layout);

        // First node has only first entry.
        assert_eq!(
            vals_for_key_1,
            get_u64_values_for_key(&fst, val_layout, "key 1".as_bytes()),
        );
        assert!(fst.find("key 2".as_bytes()).is_none());

        // Second node has only second entry.
        assert_eq!(
            vals_for_key_2,
            get_u64_values_for_key(&snd, val_layout, "key 2".as_bytes()),
        );
        assert!(snd.find("key 1".as_bytes()).is_none());
    }

    #[test]
    fn test_inner_node_basic() {
        let pool = Pool::new();
        let leaf_node_left = Node::new_leaf(&pool);
        let leaf_node_right = Node::new_leaf(&pool);
        let mut inner_node = InnerNode::new(&pool);

        inner_node.insert_entry(InnerEntry {
            key: "banana".as_bytes().to_vec(),
            left: &leaf_node_left as *const Node as *const u8,
            right: &leaf_node_right as *const Node as *const u8,
        });

        assert_eq!(
            &leaf_node_left as *const Node as usize,
            inner_node.smallest_child_gte("apple".as_bytes()).unwrap() as *const Node as usize
        );
        assert_eq!(
            &leaf_node_right as *const Node as usize,
            inner_node.smallest_child_gte("banana".as_bytes()).unwrap() as *const Node as usize
        );
        assert_eq!(
            &leaf_node_right as *const Node as usize,
            inner_node.smallest_child_gte("cherry".as_bytes()).unwrap() as *const Node as usize
        );
    }

    fn get_u64_values_for_key(n: &LeafNode, val_layout: Layout, key: &[u8]) -> Vec<u64> {
        n.find(key)
            .unwrap()
            .values_iter(val_layout)
            .map(|val| u64::from_le_bytes(val.try_into().unwrap()))
            .collect()
    }
}
