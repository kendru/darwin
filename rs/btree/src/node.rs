use std::alloc::Layout;
use std::intrinsics::copy_nonoverlapping;
use std::slice::from_raw_parts_mut;

use crate::page::{Allocation, HEADER_SIZE, Page, Pool};
use crate::entry::{ENTRY_REF_SIZE, EntryRef, PAGE_ENTRY_HEADER_SIZE, PageEntry};

#[derive(Debug, PartialEq, Eq, PartialOrd, Ord, Clone, Copy)]
struct NodeId(usize);

#[derive(Debug, PartialEq)]
struct LeafNode {
    page: Page,
    val_size: usize,
}

impl LeafNode {
    fn new(pool: &mut Pool, val_size: usize) -> Self {
        LeafNode {
            page: pool.get(),
            val_size,
        }
    }

    // TODO: add versions of search functions that take a key comparator
    // for more complex keys that do not have byte-wise comparison semantics.
    // We should leave the raw versions that perform byte comparisons so that
    // the client code can use this (faster) version whenever possible.

    // TODO: do not require &mut self, and use a version of find_entry that returns shared
    // references.
    pub fn find(&self, key: &[u8]) -> Option<&PageEntry> {
        self.find_entry(key).map(|(_, entry)| &*entry)
    }

    fn find_entry(&self, key: &[u8]) -> Option<(&mut EntryRef, &mut PageEntry)> {
        // Perform a linear scan over keys.
        // TODO: Allow this to be replaced by binary search.
        let start = HEADER_SIZE;
        let end = start + self.page.free_len() as usize;
        for offset in (start..end).step_by(ENTRY_REF_SIZE) {
            // XXX: Can we prove that it is safe to "upgrade" the pointer to a unique reference?
            let entry_ref = unsafe {
                let ptr = self.page.offset_ptr_unchecked(offset, ENTRY_REF_SIZE) as *mut EntryRef;
                &mut *ptr
            };

            let entry = unsafe {
                let start = entry_ref.offset as usize;
                let length = entry_ref.length as usize;
                let ptr = self.page.offset_ptr_unchecked(start, length) as *mut PageEntry;
                &mut *ptr
            };

            if entry.key_len as usize == key.len() && &entry.data[0..entry.key_len as usize] == key
            {
                return Some((entry_ref, entry));
            }
        }
        None
    }

    pub fn insert(&mut self, key: &[u8], val: &[u8]) {
        debug_assert_eq!(val.len(), self.val_size);

        // TODO: Factor allocations of PageEntry and EntryPtr objects out.
        if let Some((entry_ref, old_entry)) = self.find_entry(key) {
            // Copy old entry into new empty slot in data. The old
            // memory becomes dead and will eventually get garbage collected.
            let old_size = PAGE_ENTRY_HEADER_SIZE + old_entry.data.len();
            let new_size = old_size + self.val_size;
            // Align to 8 bytes.
            let entry_start = ((self.page.free_end() as usize - new_size) / 8) * 8;

            if entry_start <= self.page.free_start() as usize + ENTRY_REF_SIZE {
                todo!("Deal with full pages");
            }

            // 1. Copy bytes from old slot into new slot.
            unsafe {
                copy_nonoverlapping(
                    entry_ref as *mut EntryRef as *const u8,
                    self.page.offset_ptr_unchecked(entry_start, old_size) as *mut u8,
                    old_size,
                );
            }
            // 2. Reinterpret pointer to new slot as &mut PageEntry.
            let entry = unsafe {
                let s = from_raw_parts_mut(
                    self.page.offset_ptr_unchecked(entry_start, new_size) as *mut u8,
                    new_size,
                );
                &mut *(s as *mut [u8] as *mut PageEntry)
            };
            // 3. Update new slot: increment val_count and append to data.
            entry.val_count += 1;
            entry.data[old_entry.data.len()..].copy_from_slice(val);

            // Update entry pointer.
            entry_ref.offset = entry_start as u16;
            entry_ref.length += self.val_size as u16;
        } else {
            // Initial entry holds the sized struct fields of PageEntry plus a data buffer large
            // enough to fit a single value.
            let key_len = key.len();
            let initial_data_size = key_len + self.val_size;
            let size = PAGE_ENTRY_HEADER_SIZE + initial_data_size;
            // TODO: Handle data alignment.
            let Allocation {
                ptr: entry_ptr,
                offset: entry_start,
            } = self.page.alloc_end(Layout::from_size_align(size, 8).unwrap())
                .expect("TODO: handle allocation failure for page");
            let entry = unsafe {
                &mut *(entry_ptr as *mut PageEntry)
            };
            // XXX: What if a key cannot fit in a u16?
            entry.key_len = key.len() as u16;
            entry.val_count = 1;
            entry.data[0..key_len].copy_from_slice(key);
            let val_start = key_len + (key_len % 8);
            entry.data[val_start..].copy_from_slice(val);

            let entry_ref = &EntryRef{
                offset: entry_start as u16,
                length: size as u16,
            };
            let entry_ref_ptr = self.page.alloc_start(Layout::for_value(entry_ref))
                .expect("TODO: handle allocation failure for page");

            unsafe {
                copy_nonoverlapping(
                    entry_ref as *const EntryRef as *const u8,
                    entry_ref_ptr.ptr as *mut u8,
                    ENTRY_REF_SIZE,
                );
            }
        }
    }
}


#[cfg(test)]
mod tests {
    use std::mem::size_of;

    use super::*;

    #[test]
    fn test_insert_into_leaf() {
        let mut pool = Pool::new();
        let val_size = size_of::<u64>();
        let mut leaf_node = LeafNode::new(&mut pool, val_size);

        leaf_node.insert(&[0, 1, 45, 23], &2345u64.to_le_bytes());
        // TODO: The return type should expose a value iterator that is aware of the value size.
        let res = leaf_node.find(&[0, 1, 45, 23]).unwrap();

        assert_eq!(res.key(), &[0, 1, 45, 23]);
        assert_eq!(res.values(val_size), vec![&2345u64.to_le_bytes()]);
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
