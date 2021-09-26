pub mod page;

// use std::mem::{self, MaybeUninit};
// use std::rc::Rc;
// use std::sync::RwLock;

// const BRANCHING: usize = 32;
// const PAGE_SIZE: usize = 1024*1024;

// #[derive(Debug, PartialEq, Eq, PartialOrd, Ord, Clone, Copy)]
// struct NodeId(usize);

// trait KeyType: PartialEq + PartialOrd {}

// #[derive(Debug, PartialEq, Clone)]
// struct InnerNode<K: KeyType> {
//     keys: [K; BRANCHING-1],
//     children: [NodeId; BRANCHING],
// }

// #[derive(Debug, PartialEq)]
// struct LeafNode {
//     page: [u8; PAGE_SIZE], // Should PAGE_SIZE be a const generic?
//     next_node: Option<NodeId>,
// }

// impl LeafNode {
//     fn new() -> Self {
//         LeafNode {
//             page: [0; PAGE_SIZE],
//             next_node: None, // Does this need to be protected by a mutex or be an atomic pointer?
//         }
//     }
// }

// #[derive(Debug, PartialEq)]
// enum Node<K: KeyType, V> {
//     Inner(InnerNode<K>),
//     Leaf(LeafNode),
// }

// struct NodeArena<K: KeyType, V> {
//     storage: RwLock<Vec<Rc<Node<K, V>>>>,
// }

// impl<K: KeyType, V> NodeArena<K, V> {
//     fn new() -> Self {
//         NodeArena {
//             storage: RwLock::new(Vec::new()),
//         }
//     }

//     fn insert(&self, node: Node<K, V>) -> NodeId {
//         let mut storage = self.storage.write().expect("cannot take lock");
//         let idx = storage.len();
//         storage.push(Rc::new(node));
//         NodeId(idx)
//     }

//     fn get(&self, id: NodeId) -> Option<Rc<Node<K, V>>> {
//         let storage = self.storage.read().expect("cannot take lock");
//         storage.get(id.0).map(|n| n.clone())
//     }
// }

// struct BTree<K: KeyType, V> {
//     root: RwLock<NodeId>,
//     arena: NodeArena<K, V>,
// }

// impl<K: KeyType, V> BTree<K, V> {
//     fn new() -> Self {
//         let arena = NodeArena::new();
//         let rootId = arena.insert(Node::Leaf(LeafNode::new()));
//         let root = RwLock::new(rootId);

//         BTree { root, arena }
//     }

//     // TODO: Should this return a result/option?
//     fn insert(&self, key: K, val: V) {}
// }



fn main() {
    let mut p = page::Page::new(std::mem::size_of::<i64>());
    p.insert(&[0, 0, 12, 3], &[0, 0, 0, 0, 0, 0, 0, 42]);
    // println!("{:?}", p);
    let found = p.find(&[0, 0, 12, 3]).unwrap();
    let entry = unsafe { &*found };

    println!("{:?}", entry)
}
