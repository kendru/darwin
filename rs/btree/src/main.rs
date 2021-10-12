use std::alloc::Layout;

use node::Node;

use crate::{node::LeafNode, page::Pool};

pub mod entry;
pub mod page;
// pub mod manager;
pub mod node;
mod util;

fn main() {
    let pool = Pool::new();
    let leaf = LeafNode::new(&pool, Layout::from_size_align(8, 8).unwrap());
    let mut mut_leaf_node = unsafe { &mut *(&leaf as *const LeafNode as *mut LeafNode) };
    mut_leaf_node.insert("Andrew".as_bytes(), &124u64.to_be_bytes());
    mut_leaf_node.insert("Andrew".as_bytes(), &248u64.to_be_bytes());
    let found = leaf.find("Andrew".as_bytes()).unwrap();
    println!("Key: {:?}", found.key());
    println!("Vals: {:?}", found.values(Layout::from_size_align(8, 8).unwrap()));
}
