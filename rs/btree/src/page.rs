use std::alloc::{alloc_zeroed, Layout};
use std::cell::UnsafeCell;
use std::collections::LinkedList;
use std::convert::TryFrom;
use std::mem::{align_of, size_of};
use std::num::NonZeroU64;
use std::slice;
use std::sync::{Arc, Mutex};

use crate::util::{round_down, round_to};

// Since we are not mapping to disk pages, we can use fairly large page sizes.
// We should to some perf tests to determine a good page size in the general case.
pub(crate) const PAGE_SIZE: usize = 1024 * 16;
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

    let slice = unsafe { slice::from_raw_parts(data as *const (), len) };
    slice as *const [()] as *mut _
}

#[derive(Debug, Clone, Copy)]
#[repr(C)]
pub struct Header {
    // TODO: Should the links be kept in the pages or in the nodes?
    pub next: Option<NonZeroU64>,
    pub(super) free_start: u16,
    // Marks the spot after the last free index.
    pub(super) free_end: u16,
}

impl Header {
    pub(crate) fn reset(&mut self) {
        let start = size_of::<Header>();
        self.next = None;
        self.free_start = start as u16;
        self.free_end = PAGE_SIZE as u16;
    }

    #[inline]
    pub(super) fn free_len(&self) -> u16 {
        self.free_end - self.free_start
    }

    #[inline]
    pub(super) fn can_fit(&self, layout: Layout) -> bool {
        let alloc_start = round_to(self.free_start as usize, layout.align());
        alloc_start + layout.size() <= self.free_end as usize
    }
}

pub(crate) struct Allocation {
    pub(crate) ptr: *mut [u8],
    pub(crate) offset: u16,
}

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

    #[inline]
    pub fn free_start(&self) -> u16 {
        self.header().free_start
    }

    #[inline]
    pub fn free_len(&self) -> u16 {
        self.header().free_len()
    }

    #[inline]
    pub fn free_end(&self) -> u16 {
        self.header().free_end
    }

    #[inline]
    pub fn can_fit(&self, layout: Layout) -> bool {
        self.header().can_fit(layout)
    }

    pub(crate) fn alloc_start(&mut self, layout: Layout) -> Option<Allocation> {
        if self.can_fit(layout) {
            unsafe { Some(self.alloc_start_unchecked(layout)) }
        } else {
            None
        }

    }

    pub(crate) unsafe fn alloc_start_unchecked(&mut self, layout: Layout) -> Allocation {
        let mut header = self.header_mut();
        let start = round_to(header.free_start as usize, layout.align()) as u16;
        header.free_start = start + layout.size() as u16;

        Allocation {
            ptr: self.offset_ptr_unchecked_mut(start as usize, layout.size()),
            offset: start,
        }
    }

    /// Shifts bytes from `offset` to the right by `len` bytes.
    /// Returns None when there is not enough free space left to perform the shift.
    /// This function is unsafe because it is up to the caller to ensure that the shifted data
    /// is still properly aligned. Additionally, any references to shifted data will be
    /// invalid.
    pub(crate) unsafe fn shift_start(&mut self, offset: usize, len: usize) -> Option<()> {
        let self_ptr = self.ptr as *const u8;
        let mut header = self.header_mut();
        if (header.free_len() as usize) < len {
            return None;
        }

        let src = self_ptr.offset(offset as isize);
        let dst = src.offset(len as isize) as *mut u8;
        let count = (header.free_start as usize) - offset;
        std::ptr::copy(src, dst, count);

        // Shift free start marker to the right.
        header.free_start += len as u16;

        Some(())
    }

    pub(crate) fn alloc_end(&mut self, layout: Layout) -> Option<Allocation> {
        if self.can_fit(layout) {
            unsafe { Some(self.alloc_end_unchecked(layout)) }
        } else {
            None
        }
    }

    pub(crate) unsafe fn alloc_end_unchecked(&mut self, layout: Layout) -> Allocation {
        let mut header = self.header_mut();
        let start = round_down(header.free_end as usize - layout.size(), layout.align()) as u16;
        header.free_end = start;

        Allocation {
            ptr: self.offset_ptr_unchecked_mut(start as usize, layout.size()),
            offset: start,
        }
    }

    #[inline]
    pub(crate) fn header(&self) -> &Header {
        // SAFETY:
        // - self.ptr is a pointer to an UnsafeCell
        // - UnsafeCell has repr(transparent)
        // - Each page starts with a header
        unsafe { &*(self.ptr as *const Header) }
    }

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
    pub(crate) unsafe fn offset_ptr_unchecked(&self, start: usize, len: usize) -> *const [u8] {
        let ptr = self.buf().as_ptr().offset(start as isize);
        slice::from_raw_parts(ptr as *const u8, len) as *const [u8]
    }

    #[inline]
    pub(crate) unsafe fn offset_ptr_unchecked_mut(&mut self, start: usize, len: usize) -> *mut [u8] {
        let ptr = self.buf_mut().as_mut_ptr().offset(start as isize);
        slice::from_raw_parts_mut(ptr as *mut u8, len) as *mut [u8]
    }

    pub(crate) fn header_mut(&mut self) -> &mut Header {
        unsafe { &mut *(self.ptr as *mut Header) }
    }
}

#[derive(Clone)]
pub struct Pool {
    pages: Arc<Mutex<LinkedList<Page>>>,
}

impl Pool {
    pub fn new() -> Pool {
        Pool {
            pages: Arc::new(Mutex::new(LinkedList::new())),
        }
    }

    pub(crate) fn get(&self) -> Page {
        let mut pages = self.pages.lock().unwrap();
        pages.pop_front().unwrap_or_else(|| Page::new(None))
    }

    pub(crate) fn check_in(&self, mut page: Page) {
        page.reset();
        let mut pages = self.pages.lock().unwrap();
        pages.push_back(page);
    }
}
