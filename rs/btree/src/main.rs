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
    let mut leaf = LeafNode::new(&pool, Layout::from_size_align(8, 8).unwrap());
    leaf.insert("Andrew".as_bytes(), &124u64.to_be_bytes());
    leaf.insert("Andrew".as_bytes(), &248u64.to_be_bytes());
    let found = leaf.find("Andrew".as_bytes()).unwrap();
    println!("Key: {:?}", found.key());
    println!("Vals: {:?}", found.values(Layout::from_size_align(8, 8).unwrap()));

    leaf.insert("Diana".as_bytes(), &34u64.to_be_bytes());
    leaf.insert("Audrey".as_bytes(), &9u64.to_be_bytes());
    leaf.insert("Jonah".as_bytes(), &7u64.to_be_bytes());
    leaf.insert("Arwen".as_bytes(), &7u64.to_be_bytes());

    let mut keys = Vec::<String>::with_capacity(5);
    for (_ref, entry) in leaf.scan() {
        keys.push(unsafe { String::from_utf8_unchecked((&*entry).key().to_vec()) })
    }

    println!("Keys: {:?}", keys);
}
