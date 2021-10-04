use std::alloc::{alloc_zeroed, Layout};
use std::cell::UnsafeCell;
use std::collections::LinkedList;
use std::convert::TryFrom;
use std::mem::{align_of, size_of};
use std::num::NonZeroU64;
use std::slice::{from_raw_parts, from_raw_parts_mut};
use std::sync::{Arc, Mutex};

// Since we are not mapping to disk pages, we can use fairly large page sizes.
// We should to some perf tests to determine a good page size in the general case.
pub(crate) const PAGE_SIZE: usize = 1024 * 16;
pub(crate) const HEADER_SIZE: usize = size_of::<Header>();
pub(crate) const ALIGNMENT: usize = align_of::<Header>();

fn uninitialized_page() -> Page {
    let layout = Layout::from_size_align(PAGE_SIZE, ALIGNMENT).unwrap();
    unsafe {
        let ptr = alloc_zeroed(layout);
        let cell_ptr = fatten(ptr, PAGE_SIZE);
        Page { ptr: cell_ptr }
    }
}

/// <https://users.rust-lang.org/t/construct-fat-pointer-to-struct/29198/9>
/// Borrowed from [sled](https://github.com/spacejam/sled/blob/main/src/node.rs#L1148).
#[allow(trivial_casts)]
fn fatten(data: *mut u8, len: usize) -> *mut UnsafeCell<[u8]> {
    // Requirements of slice::from_raw_parts.
    assert!(!data.is_null());
    assert!(isize::try_from(len).is_ok());

    let slice = unsafe { core::slice::from_raw_parts(data as *const (), len) };
    slice as *const [()] as *mut _
}

#[derive(Debug, Clone, Copy)]
#[repr(C)]
pub struct Header {
    pub next: Option<NonZeroU64>,
    pub(super) free_start: u16,
    pub(super) free_len: u16,
}

impl Header {
    pub(crate) fn reset(&mut self) {
        let start = HEADER_SIZE;
        self.next = None;
        self.free_start = start as u16;
        self.free_len = (PAGE_SIZE - start) as u16;
    }
}

pub(crate) struct Allocation {
    pub(crate) ptr: *mut [u8],
    pub(crate) offset: u16,
}

#[derive(Debug, PartialEq)]
pub struct Page {
    ptr: *mut UnsafeCell<[u8]>,
}

impl Page {
    pub fn new(next: Option<::std::num::NonZeroU64>) -> Page {
        let mut inner = uninitialized_page();
        // Get "zeroed" header from fresh page, then set fields manually.
        let mut header = inner.header_mut();
        header.reset();
        header.next = next;

        inner
    }

    // Resets a page to be returned to a buffer of pages that can be reclaimed for
    // future use.
    pub fn reset(&mut self) {
        // We do not have to worry about zeroing the data portion of the page because it
        // is logically uninitialized when the header's is reset.
        self.header_mut().reset();
    }



    pub fn free_start(&self) -> u16 {
        self.header().free_start
    }

    pub fn free_len(&self) -> u16 {
        self.header().free_len
    }

    pub fn free_end(&self) -> u16 {
        let header = self.header();
        header.free_start + header.free_len
    }

    pub(crate) fn alloc_start(&mut self, layout: Layout) -> Option<Allocation> {
        let mut header = self.header_mut();
        let alloc_len = ((layout.size() / 8) * 8) as u16;
        if header.free_len < alloc_len {
            return None;
        }

        let start = header.free_start;
        header.free_start += alloc_len;

        Some(Allocation {
            ptr: self.offset_ptr_unchecked_mut(start as usize, layout.size()),
            offset: start,
        })
    }

    pub(crate) fn alloc_end(&mut self, layout: Layout) -> Option<Allocation> {
        let mut header = self.header_mut();
        let alloc_len = ((layout.size() / 8) * 8) as u16;
        if header.free_len < alloc_len {
            return None;
        }

        let start = header.free_len - alloc_len;
        header.free_len -= alloc_len;

        Some(Allocation {
            ptr: self.offset_ptr_unchecked_mut(start as usize, layout.size()),
            offset: start,
        })
    }

    pub(crate) fn header(&self) -> &Header {
        // SAFETY:
        // - self.ptr is a pointer to an UnsafeCell
        // - UnsafeCell has repr(transparent)
        // - Each page starts with a header
        unsafe { & *(self.ptr as *const Header) }
    }

    // #[inline]
    // pub(crate) fn data(&self) -> &[u8] {
    //     let len = self.free_len() as usize;
    //     unsafe {
    //         let ptr = self.offset_ptr_unchecked(HEADER_SIZE, len);
    //         from_raw_parts(ptr as *const u8, len)
    //     }
    // }

    // #[inline]
    // pub(crate) fn data_mut(&mut self) -> &mut [u8] {
    //     let len = self.free_len() as usize;
    //     unsafe {
    //         from_raw_parts_mut(self.data_ptr_mut() as *mut u8, len)
    //     }
    // }

    // #[inline]
    // pub(crate) fn data_ptr_mut(&mut self) -> *mut [u8] {
    //     let len = self.free_len() as usize;
    //     unsafe { self.offset_ptr_unchecked_mut(HEADER_SIZE, len) }
    // }

    #[inline]
    pub(crate) fn buf(&self) -> &[u8] {
        unsafe { &*(*self.ptr).get() }
    }

    #[inline]
    pub(crate) fn buf_mut(&mut self) -> &mut [u8] {
        unsafe { &mut *(*self.ptr).get() }
    }

    #[inline]
    pub(crate) fn buf_ptr(&mut self) -> *mut [u8] {
        unsafe { (*self.ptr).get() }
    }

    #[inline]
    pub(crate) fn offset_ptr_unchecked(&self, start: usize, len: usize) -> *const [u8] {
        unsafe {
            let ptr = self.buf().as_ptr().offset(start as isize);
            from_raw_parts(ptr as *const u8, len) as *const [u8]
        }
    }

    #[inline]
    pub(crate) fn offset_ptr_unchecked_mut(&mut self, start: usize, len: usize) -> *mut [u8] {
        unsafe {
            let ptr = self.buf_mut().as_mut_ptr().offset(start as isize);
            from_raw_parts_mut(ptr as *mut u8, len) as *mut [u8]
        }
    }

    fn header_mut(&mut self) -> &mut Header {
        unsafe { &mut *(self.ptr as *mut Header) }
    }
}

#[derive(Clone)]
pub(crate) struct Pool {
    pages: Arc<Mutex<LinkedList<Page>>>,
}

impl Pool {
    pub(crate) fn new() -> Pool {
        Pool {
            pages: Arc::new(Mutex::new(LinkedList::new())),
        }
    }

    pub(crate) fn get(&mut self) -> Page {
        let mut pages = self.pages.lock().unwrap();
        pages.pop_front().unwrap_or_else(|| Page::new(None))
    }

    pub(crate) fn check_in(&mut self, mut page: Page) {
        page.reset();
        let mut pages = self.pages.lock().unwrap();
        pages.push_back(page);
    }
}

// #[cfg(test)]
// mod tests {
//     use super::*;

    // TODO: Recreate equivalent tests in the node module.
    // #[test]
    // fn test_page_insert_lookup() {
    //     let mut p = Page::new(8);
    //     p.insert(&[0, 0, 12, 3], &[0, 0, 0, 0, 0, 0, 0, 42]);

    //     {
    //         let found = p.find(&[0, 0, 12, 3]).unwrap();
    //         let entry = unsafe { &*found };

    //         // TODO: extract these accesses to functions that unsafely assert the proper alignment.
    //         assert_eq!(entry.key_len, 4);
    //         assert_eq!(entry.val_count, 1);
    //         assert_eq!(&entry.data[..], &[0, 0, 12, 3, 0, 0, 0, 0, 0, 0, 0, 42]);
    //     }

    //     // Insert second value for same key.
    //     p.insert(&[0, 0, 12, 3], &[0, 0, 0, 0, 0, 0, 0, 43]);
    //     {
    //         let found = p.find(&[0, 0, 12, 3]).unwrap();
    //         let entry = unsafe { &*found };

    //         assert_eq!(entry.val_count, 2);
    //         assert_eq!(
    //             &entry.data[..],
    //             &[0, 0, 12, 3, 0, 0, 0, 0, 0, 0, 0, 42, 0, 0, 0, 0, 0, 0, 0, 43]
    //         );
    //     }

    //     // Insert value for another key.
    //     p.insert(&[99], &[0, 0, 0, 0, 0, 0, 0, 44]);
    //     {
    //         let found = p.find(&[99]).unwrap();
    //         let entry = unsafe { &*found };

    //         assert_eq!(entry.key_len, 1);
    //         assert_eq!(entry.val_count, 1);
    //         assert_eq!(&entry.data[..], &[99, 0, 0, 0, 0, 0, 0, 0, 44]);
    //     }

    //     assert_eq!(None, p.find(&[11, 22, 33, 44]));
    // }

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
// }