mod schema;
mod tuple;

#[derive(Debug, PartialEq)]
struct Fact {
    subject: u64,
    predicate: String,
    object: Vec<u8>,
}

impl Fact {
    fn new(subject: u64, predicate: String, object: Vec<u8>) -> Fact {
        Fact {
            subject,
            predicate,
            object
        }
    }
}

fn main() {
    println!("{:?}", Fact::new(1, "person:name".into(), "Andrew".as_bytes().to_vec()));
}
