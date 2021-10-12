use std::alloc::Layout;
use std::intrinsics::copy_nonoverlapping;
use std::ops::{Deref, DerefMut};
use std::slice;

use crate::entry::{EntryRef, PageEntry, ENTRY_REF_SIZE, PAGE_ENTRY_HEADER_SIZE};
use crate::page::{Allocation, Page, Pool, HEADER_SIZE};
use crate::util::pad_for;

// TODO:
// Instead of letting the caller align the data (or letting LeafNode hold a Layout for entry data),
// we are hard-coding alignment to 8. We should change this ASAP to avoid over-padding data that does
// not need such large alignment.
const TODO_DATA_ALIGN: usize = 8;

#[derive(Debug, PartialEq, Eq, PartialOrd, Ord, Clone, Copy)]
#[repr(transparent)]
pub struct NodeId(u64);

pub(crate) enum Node<'a> {
    LeafNode(LeafNode<'a>),
}

// TODO: Consider a locking scheme for nodes. Where should the lock live?
pub(crate) struct LeafNode<'a> {
    val_layout: Layout,
    page: Option<Page>, // Always Some<Page> until dropped.
    pool: &'a Pool,
}

impl LeafNode<'_> {
    // XXX: This should take a `val_align: Layout` instead of a `val_size: usize`.
    pub(crate) fn new<'a>(pool: &'a Pool, val_layout: Layout) -> LeafNode<'a> {
        LeafNode {
            val_layout,
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

    fn find_entry(&self, key: &[u8]) -> Option<(*mut EntryRef, *mut PageEntry)> {
        // Perform a linear scan over keys.
        // TODO: Allow this to be replaced by binary search.
        let start = HEADER_SIZE;
        let end = start + self.free_start() as usize;
        for offset in (start..end).step_by(ENTRY_REF_SIZE) {
            let entry_ref_ptr: *mut EntryRef;
            let entry_ref = unsafe {
                entry_ref_ptr = self.offset_ptr_unchecked(offset, ENTRY_REF_SIZE) as *mut EntryRef;
                &mut *entry_ref_ptr
            };

            let entry_ptr: *mut PageEntry;
            let entry = unsafe {
                let start = entry_ref.offset as usize;
                let length = entry_ref.length as usize;
                entry_ptr = self.offset_ptr_unchecked(start, length) as *mut PageEntry;
                &mut *entry_ptr
            };

            if entry.key_len as usize == key.len() && &entry.data[0..entry.key_len as usize] == key
            {
                return Some((entry_ref_ptr, entry_ptr));
            }
        }
        None
    }

    pub(crate) fn insert(&mut self, key: &[u8], val: &[u8]) {
        debug_assert_eq!(val.len(), self.val_layout.size());

        if let Some((entry_ref_ptr, old_entry_ptr)) = self.find_entry(key) {
            self.insert_extend(entry_ref_ptr, old_entry_ptr, val);
        } else {
            self.insert_initial(key, val);
        }
    }

    fn insert_initial(&mut self, key: &[u8], val: &[u8]) {
        // Initial entry holds the sized struct fields of PageEntry plus a data buffer large
        // enough to fit a single value.
        // TODO: Does the key need to be aligned?
        let key_len = key.len();

        // The data that each entry starts with is the key followed by enough bytes of padding
        // to align the values appropriately, followed by a single value.
        let initial_data_size =
            key_len + pad_for(PAGE_ENTRY_HEADER_SIZE + key_len, self.val_layout.align()) + self.val_layout.size();

        let size = PAGE_ENTRY_HEADER_SIZE + initial_data_size;
        // TODO: Handle data alignment.
        let Allocation {
            ptr: entry_ptr,
            offset: entry_start,
        } = self
            .alloc_end(Layout::from_size_align(size, TODO_DATA_ALIGN).unwrap())
            .expect("TODO: handle allocation failure for page");
        let entry = unsafe {
            // Reinterpret entry_ptr as a fat pointer to `initial_data_size` bytes, since this
            // is the dynamic portion of the PageEntry DST.
            let dst_ptr =
                slice::from_raw_parts_mut(entry_ptr as *mut u8, initial_data_size) as *mut [u8];
            &mut *(dst_ptr as *mut PageEntry)
        };
        // XXX: What if a key cannot fit in a u16?
        entry.key_len = key_len as u16;
        entry.val_count = 1;
        entry.data[0..key_len].copy_from_slice(key);
        let val_start = key_len + pad_for(PAGE_ENTRY_HEADER_SIZE + key_len, TODO_DATA_ALIGN);
        entry.data[val_start..].copy_from_slice(val);

        let entry_ref = &EntryRef {
            offset: entry_start as u16,
            length: initial_data_size as u16,
        };
        let entry_ref_ptr = self
            .alloc_start(Layout::for_value(entry_ref))
            .expect("TODO: handle allocation failure for page");

        unsafe {
            copy_nonoverlapping(
                entry_ref as *const EntryRef as *const u8,
                entry_ref_ptr.ptr as *mut u8,
                ENTRY_REF_SIZE,
            );
        }
    }

    fn insert_extend(
        &mut self,
        entry_ref_ptr: *mut EntryRef,
        old_entry_ptr: *mut PageEntry,
        val: &[u8],
    ) {
        let (entry_ref, old_entry) = unsafe { (&mut *entry_ref_ptr, &*old_entry_ptr) };
        // Copy old entry into new empty slot in data. The old
        // memory becomes dead and will eventually get garbage collected.
        let old_size = PAGE_ENTRY_HEADER_SIZE + old_entry.data.len();
        let new_size = old_size + self.val_layout.size();

        // 1. Allocate a new slot.
        let Allocation {
            ptr: entry_ptr,
            offset: entry_start,
        } = self
            .alloc_end(Layout::from_size_align(new_size, TODO_DATA_ALIGN).unwrap())
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
                old_entry.data.len() + self.val_layout.size(),
            );
            &mut *(sized_slice as *mut [u8] as *mut PageEntry)
        };

        // 4. Update new slot: increment val_count and append to data.
        // NOTE [DATA ALIGNMENT]:
        // We pack values contiguously, potentially inserting padding between the key and
        // the start of the value array. This assumes that, like structs, the value size
        // is a multiple of the value alignment. As long as self.value_layout was constructed
        // safely, this is a sound assumption.
        let new_val_offset = old_entry.data.len();
        entry.val_count += 1;
        entry.data[new_val_offset..].copy_from_slice(val);
        println!("Data: {:?}", &entry.data);

        // Update entry pointer.
        println!("Updating offset from {} to {}", entry_ref.offset, entry_start);
        entry_ref.offset = entry_start as u16;
        entry_ref.length += self.val_layout.size() as u16;
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
        // Page should not get dropped, so return it to the pool, and replace it
        // with a None. When Rust gets better support for telling drocpck to not
        // drop some field, that method should be used instead. See <https://doc.rust-lang.org/nightly/nomicon/destructors.html>.
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
        let mut leaf_node = LeafNode::new(&pool, val_layout);

        leaf_node.insert(&[0, 1, 45, 23], &2345u64.to_le_bytes());
        {
            // TODO: The return type should expose a value iterator that is aware of the value size.
            let res = leaf_node.find(&[0, 1, 45, 23]).unwrap();
            assert_eq!(res.key(), &[0, 1, 45, 23]);
            assert_eq!(res.values(val_layout), vec![&2345u64.to_le_bytes()]);
        }

        leaf_node.insert(&[0, 1, 45, 23], &4985355u64.to_le_bytes());
        {
            let res = leaf_node.find(&[0, 1, 45, 23]).unwrap();
            assert_eq!(
                res.values(val_layout),
                vec![&2345u64.to_le_bytes(), &4985355u64.to_le_bytes()]
            );
        }
    }
}


// #[derive(Debug, PartialEq)]
// struct InnerNode {
//     // An inner node's page holds tuples with the key as a prefix and the page id for
//     // the next level page containing all values <= that key. We store a separate entry
//     // for the high key in the page, which will either be a special value for infinity
//     // or a value equal to the low key of the next page.
//     page: Page,
//     high_pointer: Option<NodeId>,
// }

// #[derive(Debug, PartialEq)]
// enum Node<K: KeyType, V> {
//     Inner(InnerNode<K>),
//     Leaf(LeafNode),
// }

// struct NodeArena<K: KeyType, V> {
//     storage: RwLock<Vec<Rc<Node<K, V>>>>,
// }

// impl<K: KeyType, V> NodeArena<K, V> {
//     fn new() -> Self {
//         NodeArena {
//             storage: RwLock::new(Vec::new()),
//         }
//     }

//     fn insert(&self, node: Node<K, V>) -> NodeId {
//         let mut storage = self.storage.write().expect("cannot take lock");
//         let idx = storage.len();
//         storage.push(Rc::new(node));
//         NodeId(idx)
//     }

//     fn get(&self, id: NodeId) -> Option<Rc<Node<K, V>>> {
//         let storage = self.storage.read().expect("cannot take lock");
//         storage.get(id.0).map(|n| n.clone())
//     }
// }

// struct BTree<K: KeyType, V> {
//     root: RwLock<NodeId>,
//     arena: NodeArena<K, V>,
// }

// impl<K: KeyType, V> BTree<K, V> {
//     fn new() -> Self {
//         let arena = NodeArena::new();
//         let rootId = arena.insert(Node::Leaf(LeafNode::new()));
//         let root = RwLock::new(rootId);

//         BTree { root, arena }
//     }

//     // TODO: Should this return a result/option?
//     fn insert(&self, key: K, val: V) {}
// }
