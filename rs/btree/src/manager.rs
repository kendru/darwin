use std::borrow::BorrowMut;
use std::cell::Ref;
use std::sync::RwLock;

// XXX: The manager should be discarded. Instead, a parent node should own a Box to
// a child node.

use crate::node::{LeafNode, Node};
use crate::page::{Page, Pool};

const INITIAL_PAGE_CAP: usize = 64;

trait PushReturn<T> {
    fn push_return(&mut self, t: T) -> &mut T;
}

impl<T> PushReturn<T> for Vec<T> {
    fn push_return(&mut self, t: T) -> &mut T {
       self.push(t);
       self.last_mut().unwrap()
    }
}

pub struct PageManager<'a> {
    // Nodes are identified by their index in the `active` vector.
    active: RwLock<Vec<Node<'a>>>,
    page_pool: Pool,
}

impl<'a> PageManager<'a> {
    pub fn new() -> PageManager<'a> {
        PageManager {
            active: RwLock::new(Vec::with_capacity(INITIAL_PAGE_CAP)),
            page_pool: Pool::new(),
        }
    }

    pub(crate) fn new_leaf(&'a mut self, val_size: usize) -> &Node {
        let mut active = self.active.write().unwrap();
        let idx = active.len();
        let n = LeafNode::new(&self.page_pool, idx.into(), val_size);
        // SAFETY:
        // The PageManager will be guaranteed to live for at least as long as
        // any code that may borrow a reference (TODO), so we know that the
        // nodes will not be destructed while references exist.
        unsafe { &*(active.push_return(Node::LeafNode(n)) as *const Node) }
    }
}
