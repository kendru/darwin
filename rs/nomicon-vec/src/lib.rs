use std::alloc::{self, Layout};
use std::marker::PhantomData;
use std::mem::size_of;
use std::ops::{Deref, DerefMut};
use std::ptr::{self, NonNull};
use std::slice;

pub struct NVec<T> {
    buf: RawNVec<T>,
    len: usize,
}

unsafe impl<T: Send> Send for NVec<T> {}
unsafe impl<T: Sync> Sync for NVec<T> {}

impl<T> NVec<T> {
    pub fn new() -> NVec<T> {
        NVec {
            buf: RawNVec::new(),
            len: 0,
        }
    }

    pub fn capacity(&self) -> usize {
        self.cap()
    }

    #[inline]
    fn cap(&self) -> usize {
        self.buf.cap
    }

    #[inline]
    fn ptr(&self) -> *mut T {
        self.buf.ptr.as_ptr()
    }

    pub fn push(&mut self, elem: T) {
        if self.len == self.cap() {
            self.buf.grow();
        }

        unsafe {
            ptr::write(self.ptr().add(self.len), elem);
        }

        self.len += 1;
    }

    pub fn pop(&mut self) -> Option<T> {
        if self.len == 0 {
            None
        } else {
            self.len -= 1;
            unsafe { Some(ptr::read(self.ptr().add(self.len))) }
        }
    }

    pub fn insert(&mut self, index: usize, elem: T) {
        assert!(index <= self.len, "Index out of bounds");
        if self.cap() == self.len {
            self.buf.grow();
        }

        unsafe {
            let p = self.ptr();
            ptr::copy(p.add(index), p.add(index + 1), self.len - index);
            ptr::write(p.add(index), elem);
            self.len += 1;
        }
    }

    pub fn remove(&mut self, index: usize) -> T {
        assert!(index < self.len, "Index out of bounds");

        unsafe {
            self.len -= 1;
            let p = self.ptr();
            let res = ptr::read(p.add(index));
            ptr::copy(p.add(index + 1), p.add(index), self.len - index);
            res
        }
    }

    pub fn into_iter(self) -> IntoIter<T> {
        unsafe {
            let iter = RawValIter::new(&self);
            // Unsafely move buf out of self.
            let buf = ptr::read(&self.buf);
            // Do not drop.
            ::std::mem::forget(self);

            IntoIter { iter, _buf: buf }
        }
    }

    pub fn drain(&mut self) -> Drain<T> {
        unsafe {
            let iter = RawValIter::new(&self);
            self.len = 0;

            Drain {
                vec: PhantomData,
                iter: iter,
            }
        }
    }
}

impl<T> Drop for NVec<T> {
    fn drop(&mut self) {
        // Drop elements.
        while let Some(_) = self.pop() {}
    }
}

impl<T> Deref for NVec<T> {
    type Target = [T];

    fn deref(&self) -> &Self::Target {
        unsafe { slice::from_raw_parts(self.ptr(), self.len) }
    }
}

impl<T> DerefMut for NVec<T> {
    fn deref_mut(&mut self) -> &mut Self::Target {
        unsafe { slice::from_raw_parts_mut(self.ptr(), self.len) }
    }
}

pub struct IntoIter<T> {
    _buf: RawNVec<T>,
    iter: RawValIter<T>,
}

impl<T> Iterator for IntoIter<T> {
    type Item = T;
    fn next(&mut self) -> Option<Self::Item> {
        self.iter.next()
    }

    fn size_hint(&self) -> (usize, Option<usize>) {
        self.iter.size_hint()
    }
}

impl<T> DoubleEndedIterator for IntoIter<T> {
    fn next_back(&mut self) -> Option<Self::Item> {
        self.iter.next_back()
    }
}

impl<T> Drop for IntoIter<T> {
    fn drop(&mut self) {
        // Drop any remaining elements.
        for _ in &mut *self {}
    }
}

struct RawNVec<T> {
    ptr: NonNull<T>,
    cap: usize,
    _marker: PhantomData<T>,
}

unsafe impl<T: Send> Send for RawNVec<T> {}
unsafe impl<T: Sync> Sync for RawNVec<T> {}

impl<T> RawNVec<T> {
    fn new() -> RawNVec<T> {
        // Set max capacity for ZST.
        let cap = if size_of::<T>() == 0 { !0 } else { 0 };
        RawNVec {
            ptr: NonNull::dangling(),
            cap,
            _marker: PhantomData,
        }
    }

    fn grow(&mut self) {
        assert!(size_of::<T>() != 0, "capacity overflow");

        let (new_cap, new_layout) = if self.cap == 0 {
            (1, Layout::array::<T>(1).unwrap())
        } else {
            let new_cap = 2 * self.cap;
            let new_layout = Layout::array::<T>(new_cap).unwrap();
            (new_cap, new_layout)
        };
        assert!(
            new_layout.size() <= isize::MAX as usize,
            "Allocation too large"
        );

        let new_ptr = if self.cap == 0 {
            unsafe { alloc::alloc(new_layout) }
        } else {
            let old_layout = Layout::array::<T>(self.cap).unwrap();
            let old_ptr = self.ptr.as_ptr() as *mut u8;
            unsafe { alloc::realloc(old_ptr, old_layout, new_layout.size()) }
        };

        self.ptr = match NonNull::new(new_ptr as *mut T) {
            Some(p) => p,
            None => alloc::handle_alloc_error(new_layout),
        };
        self.cap = new_cap;
    }
}

impl<T> Drop for RawNVec<T> {
    fn drop(&mut self) {
        let elem_size = size_of::<T>();
        if self.cap != 0 && elem_size != 0 {
            let layout = Layout::array::<T>(self.cap).unwrap();
            unsafe {
                alloc::dealloc(self.ptr.as_ptr() as *mut u8, layout);
            }
        }
    }
}

pub struct Drain<'a, T: 'a> {
    vec: PhantomData<&'a mut NVec<T>>,
    iter: RawValIter<T>,
}

impl<'a, T: 'a> Iterator for Drain<'a, T> {
    type Item = T;

    fn next(&mut self) -> Option<Self::Item> {
        self.iter.next()
    }

    fn size_hint(&self) -> (usize, Option<usize>) {
        self.iter.size_hint()
    }
}

impl<'a, T: 'a> DoubleEndedIterator for Drain<'a, T> {
    fn next_back(&mut self) -> Option<Self::Item> {
        self.iter.next_back()
    }
}

impl<'a, T: 'a> Drop for Drain<'a, T> {
    fn drop(&mut self) {
        for _ in &mut *self {}
    }
}

struct RawValIter<T> {
    start: *const T,
    end: *const T,
}

impl<T> RawValIter<T> {
    unsafe fn new(slice: &[T]) -> RawValIter<T> {
        RawValIter {
            start: slice.as_ptr(),
            end: if size_of::<T>() == 0 {
                // NOTE [ZST POINTER]:
                // For zero-sized types, we use a dangling pointer that is never
                // dereferenced (ptr::read<T> is a no-op when T is a ZST). Instead
                // of an extra count field for use with ZSTs, we pretend that each
                // one occupies a single byte and rely on pointer arithmetic to
                // keep track of where we are in the iterator.
                ((slice.as_ptr() as usize) + slice.len()) as *const _
            } else if slice.len() == 0 {
                slice.as_ptr()
            } else {
                slice.as_ptr().add(slice.len())
            },
        }
    }
}

impl<T> Iterator for RawValIter<T> {
    type Item = T;

    fn next(&mut self) -> Option<Self::Item> {
        if self.start == self.end {
            None
        } else {
            unsafe {
                let res = ptr::read(self.start);
                self.start = if size_of::<T>() == 0 {
                    // See NOTE: [ZST POINTER].
                    (self.start as usize + 1) as *const _
                } else {
                    self.start.offset(1)
                };
                Some(res)
            }
        }
    }

    fn size_hint(&self) -> (usize, Option<usize>) {
        let byte_range = self.end as usize - self.start as usize;
        let mut elem_size = size_of::<T>();
        if elem_size == 0 {
            // See NOTE: [ZST POINTER].
            elem_size = 1;
        }
        let len = byte_range / elem_size;
        (len, Some(len))
    }
}

impl<T> DoubleEndedIterator for RawValIter<T> {
    fn next_back(&mut self) -> Option<Self::Item> {
        if self.start == self.end {
            None
        } else {
            unsafe {
                self.end = if size_of::<T>() == 0 {
                    // See NOTE: [ZST POINTER].
                    ((self.end as usize) - 1) as *const _
                } else {
                    self.end.offset(-1)
                };
                Some(ptr::read(self.end))
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::{rc::Rc, sync::atomic::AtomicUsize};

    #[test]
    fn test_push_pop() {
        let mut v = NVec::new();
        v.push(1);
        v.push(2);
        v.push(3);

        assert_eq!(Some(3), v.pop());
        assert_eq!(Some(2), v.pop());
        assert_eq!(Some(1), v.pop());
        assert_eq!(None, v.pop());
    }

    #[test]
    fn test_insert() {
        let mut v = NVec::new();
        v.push("Hello");
        v.push("World");
        v.insert(1, "there");

        assert_eq!("World", v.pop().unwrap());
        assert_eq!("there", v.pop().unwrap());
        assert_eq!("Hello", v.pop().unwrap());
    }

    #[test]
    fn test_remove() {
        let mut v = NVec::new();
        v.push("Hello");
        v.push("there");
        v.push("World");
        assert_eq!("there", v.remove(1));

        assert_eq!("World", v.pop().unwrap());
        assert_eq!("Hello", v.pop().unwrap());
    }

    #[test]
    fn test_slice_deref() {
        let mut v = NVec::new();
        v.push(1);
        v.push(2);
        v.push(3);

        assert_eq!(3, v.len());

        let sum = v.iter().map(|x| *x).reduce(|acc, elem| acc + elem).unwrap();
        assert_eq!(sum, 6);
    }

    struct Elem {
        s: &'static str,
        drop_ctr: Rc<AtomicUsize>,
    }

    impl Drop for Elem {
        fn drop(&mut self) {
            let dropped = self
                .drop_ctr
                .fetch_add(1, std::sync::atomic::Ordering::SeqCst);
            println!("Dropping {} (dropped {})", &self.s, dropped + 1);
        }
    }

    struct ElemTracker {
        drop_ctr: Rc<AtomicUsize>,
    }

    impl ElemTracker {
        fn new() -> ElemTracker {
            ElemTracker {
                drop_ctr: Rc::new(AtomicUsize::new(0)),
            }
        }

        fn get(&self, s: &'static str) -> Elem {
            Elem {
                s: s,
                drop_ctr: self.drop_ctr.clone(),
            }
        }

        fn drop_count(&self) -> usize {
            self.drop_ctr.load(std::sync::atomic::Ordering::SeqCst)
        }
    }

    #[test]
    fn test_calls_destructors() {
        let elem_tracker = ElemTracker::new();
        let mut v = NVec::new();
        v.push(elem_tracker.get("Hello"));
        v.push(elem_tracker.get("World"));
        assert_eq!(0, elem_tracker.drop_count());

        drop(v);
        assert_eq!(2, elem_tracker.drop_count());
    }

    #[test]
    fn test_into_iter() {
        let mut v = NVec::new();
        v.push("taco");
        v.push("nacho");
        v.push("burrito");

        assert_eq!(vec!["taco", "nacho", "burrito"], v.into_iter().collect::<Vec<&str>>());
    }

    #[test]
    fn test_double_ended_iteration() {
        let mut v = NVec::new();
        v.push("taco");
        v.push("nacho");
        v.push("burrito");

        assert_eq!(vec!["burrito", "nacho", "taco"], v.into_iter().rev().collect::<Vec<&str>>());
    }

    #[test]
    fn test_drain() {
        let mut v = NVec::new();
        let elem_tracker = ElemTracker::new();
        v.push(elem_tracker.get("first"));
        v.push(elem_tracker.get("second"));
        v.push(elem_tracker.get("third"));

        let cap = v.capacity();
        for _ in v.drain() {}

        // Drained elements were dropped.
        assert_eq!(3, elem_tracker.drop_count());
        // Still has allocated capacity.
        assert_eq!(cap, v.capacity());
        // Length has been reset.
        assert_eq!(0, v.len());

        // Push/pop behaviour still normal after drain.
        v.push(elem_tracker.get("another"));
        assert_eq!("another", v.pop().unwrap().s);
        assert!(v.pop().is_none());
    }

    #[derive(Debug, PartialEq)]
    struct Nothing;

    #[test]
    fn test_zst() {
        let mut v = NVec::new();
        v.push(Nothing);
        v.push(Nothing);
        assert_eq!(v.len(), 2);
        assert_eq!(Some(Nothing), v.pop());
        assert_eq!(Some(Nothing), v.pop());
        assert_eq!(None, v.pop());
    }
}
