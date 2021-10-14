use std::alloc::Layout;

use node::Node;

use crate::{page::Pool, tree::BTree};

pub mod entry;
pub mod page;
pub mod node;
mod util;
pub mod tree;

fn main() {
    let pool = Pool::new();
    let mut index = BTree::new(Layout::from_size_align(8, 8).unwrap(), &pool);
    index.insert("Andrew".as_bytes(), &124u64.to_be_bytes());
    index.insert("Andrew".as_bytes(), &248u64.to_be_bytes());
    println!("Vals: {:?}", index.get("Andrew".as_bytes()).unwrap().collect::<Vec<_>>());

    index.insert("Diana".as_bytes(), &34u64.to_be_bytes());
    index.insert("Audrey".as_bytes(), &9u64.to_be_bytes());
    index.insert("Jonah".as_bytes(), &7u64.to_be_bytes());
    index.insert("Arwen".as_bytes(), &9u64.to_be_bytes());

    println!("Vals: {:?}", index.get("Jonah".as_bytes()).unwrap().collect::<Vec<_>>());

    // let mut keys = Vec::<String>::with_capacity(5);
    // for (_ref, entry) in leaf.scan() {
    //     keys.push(unsafe { String::from_utf8_unchecked((&*entry).key().to_vec()) })
    // }
    // println!("Keys: {:?}", keys);
}
