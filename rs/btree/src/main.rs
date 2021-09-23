use std::mem::size_of;

// use std::mem::{self, MaybeUninit};
// use std::rc::Rc;
// use std::sync::RwLock;

// const BRANCHING: usize = 32;
// const PAGE_SIZE: usize = 1024*1024;

// #[derive(Debug, PartialEq, Eq, PartialOrd, Ord, Clone, Copy)]
// struct NodeId(usize);

// trait KeyType: PartialEq + PartialOrd {}

// #[derive(Debug, PartialEq, Clone)]
// struct InnerNode<K: KeyType> {
//     keys: [K; BRANCHING-1],
//     children: [NodeId; BRANCHING],
// }

// #[derive(Debug, PartialEq)]
// struct LeafNode {
//     page: [u8; PAGE_SIZE], // Should PAGE_SIZE be a const generic?
//     next_node: Option<NodeId>,
// }

// impl LeafNode {
//     fn new() -> Self {
//         LeafNode {
//             page: [0; PAGE_SIZE],
//             next_node: None, // Does this need to be protected by a mutex or be an atomic pointer?
//         }
//     }
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

const PAGE_SIZE: usize = 1024 * 8;

#[derive(Debug)]
#[repr(C)]
struct EntryPtr {
    offset: u16,
    length: u16,
}
const ENTRY_PTR_SIZE: usize = size_of::<EntryPtr>();

impl EntryPtr {
    #[inline]
    fn start(&self) -> usize {
        self.offset as usize
    }
}

// Possible optimization: allocate some fixed number of value slots for each entry and
// keep track of how many are free. This way, we minimize the frequency of allocating
// new slots.
#[derive(Debug)]
#[repr(packed, C)]
struct PageEntry {
    key_len: u16,
    val_count: u16, // TODO: Do we need to know the length of each element? The total length?
    data: [u8],
}

const PAGE_ENTRY_FIXED_SIZE: usize = ::std::mem::size_of::<u16>() * 2;

// Page implements an index or table page that contains keys with multiple associated values.
// The page itself stores key pointers from the beginning of the data array and entries from
// the end of the array. The key pointers are fixed-size logical pointers that are offsets
// to where the corresponding entry starts
#[derive(Debug)]
struct Page {
    // Require that the values stored in the page are fixed-size. For indexes,
    // this will typically be a RowID that can be looked up by a heap manager.
    val_size: usize,
    // Offset in `data` where the next EntryPtr should start.
    next_entry_ptr: usize,
    // Offset in `data` where the next entry should end.
    next_entry: usize,
    // Raw page data. Key pointers grow from the beginning of the page, and value pointers grow from the end.
    data: [u8; PAGE_SIZE],
}

impl Page {
    fn new(val_size: usize) -> Self {
        Page {
            val_size,
            next_entry_ptr: 0,
            next_entry: PAGE_SIZE,
            data: [0; PAGE_SIZE],
        }
    }

    // Prepares the node to be returned to a pool of unused pages.
    fn reset(&mut self) {
        self.next_entry_ptr = 0;
        self.next_entry = PAGE_SIZE - 1;
        // Data does not need to be reset because when we reset the offsets,
        // it becomes logically uninitialized memory.
    }

    // TODO: add versions of search functions that take a key comparator
    // for more complex keys that do not have byte-wise comparison semantics.
    // We should leave the raw versions that perform byte comparisons so that
    // the client code can use this (faster) version whenever possible.

    fn find_entry(&self, key: &[u8]) -> Option<(*mut EntryPtr, *const PageEntry)> {
        // Perform a linear scan over keys. TODO: Allow this to be replaced by
        // a binary search.
        for offset in (0..self.next_entry_ptr).step_by(ENTRY_PTR_SIZE) {
            let end = offset + ENTRY_PTR_SIZE;
            let raw_entry_ptr = &self.data[offset..end] as *const [u8] as *mut EntryPtr;
            let entry_ptr = unsafe { &*raw_entry_ptr };

            let entry = unsafe {
                let s = ::std::slice::from_raw_parts(
                    &self.data[entry_ptr.start()] as *const u8,
                    entry_ptr.length as usize,
                );
                &*(s as *const [u8] as *const PageEntry)
            };

            if entry.key_len as usize == key.len() && &entry.data[0..entry.key_len as usize] == key
            {
                return Some((raw_entry_ptr, entry));
            }
        }
        None
    }

    fn insert(&mut self, key: &[u8], val: &[u8]) {
        debug_assert_eq!(val.len(), self.val_size);

        // TODO: Factor allocations of PageEntry and EntryPtr objects out.
        if let Some((entry_ptr_ptr, entry_ptr)) = self.find_entry(key) {
            // Copy old entry into new empty slot in data. The old
            // memory becomes dead and will eventually get garbage collected.
            let old_entry = unsafe { &*entry_ptr };
            let old_size = PAGE_ENTRY_FIXED_SIZE + old_entry.data.len();
            let new_size = old_size + self.val_size;
            // Align to 8 bytes.
            let entry_start = ((self.next_entry - new_size) / 8) * 8;

            if entry_start <= self.next_entry_ptr + ENTRY_PTR_SIZE {
                todo!("Deal with full pages");
            }

            // 1. Copy bytes from old slot into new slot.
            unsafe {
                std::ptr::copy_nonoverlapping(
                    entry_ptr as *const u8,
                    &mut self.data[entry_start] as *mut u8,
                    old_size,
                );
            }
            // 2. Reinterpret pointer to new slot as &mut PageEntry.
            let entry = unsafe {
                let s = ::std::slice::from_raw_parts_mut(
                    &mut self.data[entry_start] as *mut u8,
                    old_entry.data.len() + self.val_size,
                );
                &mut *(s as *mut [u8] as *mut PageEntry)
            };
            // 3. Update new slot: increment val_count and append to data.
            entry.val_count += 1;
            entry.data[old_entry.data.len()..].copy_from_slice(val);

            // Update entry pointer.
            unsafe {
                (&mut *entry_ptr_ptr).offset = entry_start as u16;
                (&mut *entry_ptr_ptr).length += self.val_size as u16;
            }

            self.next_entry = entry_start;
        } else {
            // Initial entry holds the sized struct fields of PageEntry plus a data buffer large
            // enough to fit a single value.
            let initial_data_size = key.len() + self.val_size;
            // Align to 8 bytes.
            let entry_start =
                ((self.next_entry - PAGE_ENTRY_FIXED_SIZE - initial_data_size) / 8) * 8;

            if entry_start <= self.next_entry_ptr + ENTRY_PTR_SIZE {
                todo!("Deal with full pages");
            }

            let mut next_field_offset = entry_start;

            // Set key_len.
            self.data[next_field_offset..next_field_offset + 2]
                .copy_from_slice(&(key.len() as u16).to_le_bytes());
            next_field_offset += 2;

            // Set val_count.
            self.data[next_field_offset..next_field_offset + 2]
                .copy_from_slice(&1u16.to_le_bytes());
            next_field_offset += 2;

            // Set data slice.
            self.data[next_field_offset..next_field_offset + key.len()].copy_from_slice(key);
            next_field_offset += key.len();
            self.data[next_field_offset..next_field_offset + self.val_size].copy_from_slice(val);

            let entry_ptr_offset = self.next_entry_ptr;
            let entry_ptr_end = entry_ptr_offset + ENTRY_PTR_SIZE;
            let raw_entry_ptr =
                &self.data[entry_ptr_offset..entry_ptr_end] as *const [u8] as *mut EntryPtr;
            let mut entry_ptr = unsafe { &mut *raw_entry_ptr };
            entry_ptr.offset = entry_start as u16;
            // Mimic fat pointer layout to a DST. It appears that the length field of a fat
            // pointer describes the length of the dynamic component, not the entire field.
            entry_ptr.length = initial_data_size as u16; // This field is us

            self.next_entry_ptr += size_of::<EntryPtr>();
            self.next_entry = entry_start;
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_page_insert_lookup() {
        let mut p = Page::new(8);
        p.insert(&[0, 0, 12, 3], &[0, 0, 0, 0, 0, 0, 0, 42]);

        {
            let (_, found) = p.find_entry(&[0, 0, 12, 3]).unwrap();
            let entry = unsafe { &*found };

            // TODO: extract these accesses to functions that unsafely assert the proper alignment.
            assert_eq!(entry.key_len, 4);
            assert_eq!(entry.val_count, 1);
            assert_eq!(&entry.data[..], &[0, 0, 12, 3, 0, 0, 0, 0, 0, 0, 0, 42]);
        }

        // Insert second value for same key.
        p.insert(&[0, 0, 12, 3], &[0, 0, 0, 0, 0, 0, 0, 43]);
        {
            let (_, found) = p.find_entry(&[0, 0, 12, 3]).unwrap();
            let entry = unsafe { &*found };

            assert_eq!(entry.val_count, 2);
            assert_eq!(
                &entry.data[..],
                &[0, 0, 12, 3, 0, 0, 0, 0, 0, 0, 0, 42, 0, 0, 0, 0, 0, 0, 0, 43]
            );
        }

        // Insert value for another key.
        p.insert(&[99], &[0, 0, 0, 0, 0, 0, 0, 44]);
        {
            let (_, found) = p.find_entry(&[99]).unwrap();
            let entry = unsafe { &*found };

            assert_eq!(entry.key_len, 1);
            assert_eq!(entry.val_count, 1);
            assert_eq!(&entry.data[..], &[99, 0, 0, 0, 0, 0, 0, 0, 44]);
        }
        for n in 0i64..289 {
            let key = n.to_le_bytes();
            p.insert(&key, &key);
        }

        println!("{:?}", p);

        assert_eq!(None, p.find_entry(&[11, 22, 33, 44]));
    }

    // #[test]
    // fn test_arena_insert() {
    //     let arena = NodeArena::new();
    //     let leaf = LeafNode::new();
    //     leaf.keys[0] = Some
    //     let my_node = Node::Leaf();
    //     let id = arena.insert(my_node.clone());
    //     assert_eq!(NodeId(0), id);

    //     let retrieved = arena.get(id).unwrap();
    //     assert_eq!(&my_node, retrieved.as_ref());
    // }
}

fn main() {
    let mut p = Page::new(std::mem::size_of::<i64>());
    p.insert(&[0, 0, 12, 3], &[0, 0, 0, 0, 0, 0, 0, 42]);
    // println!("{:?}", p);
    let (_, found) = p.find_entry(&[0, 0, 12, 3]).unwrap();
    let entry = unsafe { &*found };

    println!("{:?}", entry)
}
