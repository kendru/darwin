use std::marker::PhantomData;
use std::ops::Deref;
use std::ptr::NonNull;
use std::sync::atomic::{self, AtomicUsize, Ordering};

pub struct Arc<T> {
    ptr: NonNull<ArcInner<T>>,
    _marker: PhantomData<ArcInner<T>>,
}

impl<T> Arc<T> {
    pub fn new(data: T) -> Arc<T> {
        let boxed = Box::new(ArcInner {
            rc: AtomicUsize::new(1),
            data,
        });
        Arc {
            ptr: unsafe {
                // SAFETY:
                // Box::into_raw() is guaranteed to return a non-null pointer.
                NonNull::new_unchecked(Box::into_raw(boxed))
            },
            _marker: PhantomData,
        }
    }

    #[inline]
    fn inner_ref(&self) -> &ArcInner<T> {
        unsafe { self.ptr.as_ref() }
    }
}

impl<T> Deref for Arc<T> {
    type Target = T;

    fn deref(&self) -> &Self::Target {
        &self.inner_ref().data
    }
}

impl <T> Clone for Arc<T> {
    fn clone(&self) -> Self {
        let inner = self.inner_ref();
        let old_rc = inner.rc.fetch_add(1, Ordering::Relaxed);
        if old_rc >= isize::MAX as usize {
            // Something is seriously wrong with the program.
            ::std::process::abort();
        }
        Self {
            ptr: self.ptr,
            _marker: PhantomData,
        }
    }
}

impl<T> Drop for Arc<T> {
    fn drop(&mut self) {
        let inner = self.inner_ref();
        if inner.rc.fetch_sub(1, Ordering::Release) != 1 {
            return;
        }

        // Synchronize with the decrement of the reference count to ensure
        // that there is no race that would allow inner to be dropped twice.
        atomic::fence(Ordering::Acquire);

        // Cast to a box to drop contents automatically.
        unsafe { Box::from_raw(self.ptr.as_ptr()); }
    }
}

unsafe impl<T: Send + Sync> Send for Arc<T> {}
unsafe impl<T: Send + Sync> Sync for Arc<T> {}


struct ArcInner<T> {
    rc: AtomicUsize,
    data: T,
}

#[cfg(test)]
mod tests {
    use std::{sync::atomic::AtomicBool, thread};

    use super::*;

    #[test]
    fn test_create() {
        let _ = Arc::new("Test");
    }

    struct DropCheckedI32<'a> {
        val: i32,
        dropped: &'a AtomicBool,
    }

    impl<'a> Drop for DropCheckedI32<'a> {
        fn drop(&mut self) {
            self.dropped.store(true, Ordering::SeqCst);
        }
    }

    #[test]
    fn test_multi_threaded_access() {
        // Use Arc from the Stdlib to keep track of observed values.
        let observations = ::std::sync::Arc::new(
            ::std::sync::Mutex::new(
                Vec::<i32>::new(),
            ),
        );
        let o1 = observations.clone();
        let o2 = observations.clone();

        let dropped = &*Box::leak(Box::new(AtomicBool::new(false)));
        let a1 = Arc::new(DropCheckedI32{
            val: 10,
            dropped,
        });
        let a2 = a1.clone();
        let t1 = thread::spawn(move || {
            let mut o = o1.lock().unwrap();
            o.push(a1.val);
        });
        let t2 = thread::spawn(move || {
            let mut o = o2.lock().unwrap();
            o.push(a2.val);
        });

        t1.join().unwrap();
        t2.join().unwrap();
        assert!(dropped.load(Ordering::SeqCst));

        assert_eq!(vec![10i32, 10], observations.lock().unwrap().clone());
    }
}
