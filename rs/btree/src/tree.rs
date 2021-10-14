use std::alloc::Layout;

use crate::{entry::ValuesIterator, node::Node, page::Pool};

pub struct BTree<'a> {
    root: Node<'a>,
    val_layout: Layout,
    pool: &'a Pool,
}

impl<'a> BTree<'a> {
    pub fn new(val_layout: Layout, pool: &'a Pool) -> BTree<'a> {
        BTree {
            root: Node::new_leaf(pool),
            val_layout,
            pool,
        }
    }

    pub fn get(&self, key: &[u8]) -> Option<ValuesIterator> {
        match &self.root {
            Node::LeafNode(n) => n.find(key).map(|entry| entry.values_iter(self.val_layout)),
            Node::InnerNode(n) => {
                unimplemented!("TODO: Implement get over Node::InnerNode");
            }
        }
    }

    pub fn insert(&mut self, key: &[u8], val: &[u8]) {
        match &mut self.root {
            Node::LeafNode(n) => n.insert(self.val_layout, key, val),
            Node::InnerNode(n) => {
                unimplemented!("TODO: Implement insert over Node::InnerNode");
            }
        }
    }
}
