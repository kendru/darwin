pub mod error;

pub struct Uri(String);

impl From<String> for Uri {
    fn from(value: String) -> Self {
        Uri(value)
    }
}

impl From<&str> for Uri {
    fn from(value: &str) -> Self {
        Uri(value.to_string())
    }
}

pub struct RawTriple {
    subject: Uri,
    predicate: Uri,
    object: Vec<u8>,
}

impl RawTriple {
    pub fn new<S, P, O>(s: S, p: P, o: O) -> RawTriple
    where
        S: Into<Uri>,
        P: Into<Uri>,
        O: Into<Vec<u8>>,
    {
        RawTriple {
            subject: s.into(),
            predicate: p.into(),
            object: o.into(),
        }
    }
}

pub struct Db {
    triples: Vec<RawTriple>,
}

impl Db {
    pub fn new() -> Db {
        Db {
            triples: Vec::new(),
        }
    }

    pub fn append(&mut self, triple: RawTriple) -> error::Result<()> {
        self.triples.push(triple);
        Ok(())
    }
}

impl IntoIter {
    
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn it_works() {
        let result = 2 + 2;
        assert_eq!(result, 4);
    }
}
