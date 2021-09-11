use std::mem::{self, MaybeUninit};
use std::rc::Rc;
use std::sync::RwLock;

#[derive(Debug, PartialEq, Eq, PartialOrd, Ord, Clone, Copy)]
struct NodeId(usize);

#[derive(Debug, PartialEq, Clone)]
struct InnerNode<K: PartialEq + PartialOrd, const N: usize> {
    // Note that we over-allocate by 1 key because the maximum number
    // of keys actually stored is N-1. We should consider using constants
    // for the array sizes instead of using const generics. We need to
    // consider whether there is a need for building trees with different
    // branching factors in the same binary.
    keys: [K; N],
    children: [NodeId; N],
}

#[derive(Debug, PartialEq, Clone)]
struct LeafNode<K: PartialEq + PartialOrd, V, const N: usize> {
    keys: [Option<K>; N],
    vals: [Option<V>; N],
}

impl<K: PartialEq + PartialOrd, V, const N: usize> LeafNode<K, V, N> {
    fn new() -> Self {
        LeafNode {
            keys: array_with_none::<K, N>(),
            vals: array_with_none::<V, N>(),
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
enum Node<K: PartialEq + PartialOrd, V, const N: usize> {
    Inner(InnerNode<K, N>),
    Leaf(LeafNode<K, V, N>),
}

struct NodeArena<K: PartialEq + PartialOrd, V, const N: usize> {
    storage: RwLock<Vec<Rc<Node<K, V, N>>>>,
}

impl<K: PartialEq + PartialOrd, V, const N: usize> NodeArena<K, V, N> {
    fn new() -> Self {
        NodeArena {
            storage: RwLock::new(Vec::new()),
        }
    }

    fn insert(&self, node: Node<K, V, N>) -> NodeId {
        let mut storage = self.storage.write().expect("cannot take lock");
        let idx = storage.len();
        storage.push(Rc::new(node));
        NodeId(idx)
    }

    fn get(&self, id: NodeId) -> Option<Rc<Node<K, V, N>>> {
        let storage = self.storage.read().expect("cannot take lock");
        storage.get(id.0).map(|n| n.clone())
    }
}

struct BTree<K: PartialEq + PartialOrd, V, const N: usize> {
    root: RwLock<NodeId>,
    arena: NodeArena<K, V, N>,
}

impl<K: PartialEq + PartialOrd, V, const N: usize> BTree<K, V, N> where K: Default {
    fn new() -> Self {
        let arena = NodeArena::new();
        let rootId = arena.insert(Node::Leaf(LeafNode::new()));
        let root = RwLock::new(rootId);

        BTree { root, arena }
    }

    // TODO: Should this return a result/option?
    fn insert(&self, key: K, val: V) {}
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_arena_insert() {
        let arena = NodeArena::new();
        let my_node = Node::Leaf(LeafNode {
            keys: [1],
            vals: [123],
        });
        let id = arena.insert(my_node.clone());
        assert_eq!(NodeId(0), id);

        let retrieved = arena.get(id).unwrap();
        assert_eq!(&my_node, retrieved.as_ref());
    }
}

fn array_with_none<K>() -> [Option<K>; 32] {
    let mut data: [MaybeUninit<Option<K>>; 32] = unsafe {
        MaybeUninit::uninit().assume_init()
    };

    for elem in &mut data[..] {
        elem.write(None);
    }

    unsafe { mem::transmute::<_, [Option<K>; 32]>(data) }
}

fn main() {
    println!("Hello, world!");
}
