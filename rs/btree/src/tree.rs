use std::{alloc::Layout, convert::TryInto, mem};

use crate::{
    entry::ValuesIterator,
    node::{InnerEntry, InnerNode, LeafNode, Node},
    page::Pool,
};

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
            Node::LeafNode(n) => {
                match n.insert(self.val_layout, key, val) {
                    Some(_) => {}
                    None => {
                        let old_root: LeafNode =
                            mem::replace(&mut self.root, Node::new_inner(self.pool))
                                .try_into()
                                .unwrap();
                        let (left, right) = old_root.split(self.val_layout);
                        // TODO: Is this the correct behaviour if the left node only has the capacity for a single entry?
                        let pivot = right
                            .scan()
                            .next()
                            .map(|(_, ptr)| {
                                let entry = unsafe { &*ptr };
                                entry.key().to_vec()
                            })
                            .unwrap_or_else(|| key.to_vec());

                        self.root
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
                unimplemented!("TODO: Implement insert over Node::InnerNode");
            }
        }
    }
}
