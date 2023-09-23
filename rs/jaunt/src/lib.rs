pub mod error;
pub mod uri;
pub mod value;
pub mod data_type;

use std::{fmt, collections, io::Read};

use value::Value;

#[derive(Clone, PartialEq)]
pub struct RawTriple {
    subject: uri::Uri,
    predicate: uri::Uri,
    object: Vec<u8>,
}

impl RawTriple {
    pub fn new<S, P, O>(s: S, p: P, o: O) -> RawTriple
    where
        S: Into<uri::Uri>,
        P: Into<uri::Uri>,
        O: Into<Vec<u8>>,
    {
        RawTriple {
            subject: s.into(),
            predicate: p.into(),
            object: o.into(),
        }
    }
}

impl fmt::Debug for RawTriple {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(
            f,
            "{:?} {:?} ??? .",
            self.subject,
            self.predicate,
        )
    }
}

pub struct Triple {
    subject: uri::Uri,
    predicate: uri::Uri,
    object: value::Value,
}

impl Triple {
    pub fn new<S, P>(s: S, p: P, o: Value) -> Triple
    where
        S: Into<uri::Uri>,
        P: Into<uri::Uri>,
    {
        Triple {
            subject: s.into(),
            predicate: p.into(),
            object: o.into(),
        }
    }
}

impl fmt::Debug for Triple {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(
            f,
            "{:?} {:?} {:?} .",
            self.subject,
            self.predicate,
            self.object,
        )
    }
}

impl From<RawTriple> for Triple {
    fn from(value: RawTriple) -> Self {
        Triple {
            subject: value.subject,
            predicate: value.predicate,
            object: value::Value::parse(&value.object).unwrap_or(value::Value::IllFormed("TODO: handle error".to_string())),
        }
    }
}

impl Into<RawTriple> for Triple {
    fn into(self) -> RawTriple {
        let mut buf = Vec::new();
        self.object.serialize(&mut buf).unwrap();

        RawTriple {
            subject: self.subject,
            predicate: self.predicate,
            object: buf,
        }
    }
    
    // fn into(s) -> Self {
    //     Triple {
    //         subject: value.subject,
    //         predicate: value.predicate,
    //         object: value.object,
    //     }
    // }
}

pub struct Db {
    // Indexes
    spo: collections::BTreeMap<(uri::Uri, uri::Uri), Vec<u8>>,
    pso: collections::BTreeMap<(uri::Uri, uri::Uri), Vec<u8>>,
    pos: collections::BTreeMap<(uri::Uri, Vec<u8>), uri::Uri>,
}

impl Db {
    pub fn new() -> Db {
        Db {
            spo: collections::BTreeMap::new(),
            pso: collections::BTreeMap::new(),
            pos: collections::BTreeMap::new(),
        }
    }

    pub fn append(&mut self, triple: RawTriple) -> error::Result<()> {
        // TODO: Be more space-efficient by sharing the storage for the data that is shared between each index.
        self.spo.insert((triple.subject.clone(), triple.predicate.clone()), triple.object.clone());
        self.pso.insert((triple.predicate.clone(), triple.subject.clone()), triple.object.clone());
        self.pos.insert((triple.predicate, triple.object), triple.subject);
        Ok(())
    }

    pub fn iter(&self) -> TupleIter {
        TupleIter {
            inner: self.spo.iter(),
        }
    }

    pub fn scan_spo(&self, prefix: (uri::Uri, uri::Uri)) -> SPOScan {
        let (subject, predicate) = prefix.clone();
        SPOScan {
            inner: self.spo.range(prefix..),
            subject,
            predicate,
        }
    }

    pub fn scan_pso(&self, prefix: (uri::Uri, uri::Uri)) -> PSOScan {
        let (predicate, subject) = prefix.clone();
        PSOScan {
            inner: self.pso.range(prefix..),
            predicate,
            subject,
        }
    }

    pub fn scan_pos(&self, prefix: (uri::Uri, Vec<u8>)) -> POSScan {
        let (predicate, object) = prefix.clone();
        POSScan {
            inner: self.pos.range(prefix..),
            predicate,
            object,
        }
    }
}

pub struct SPOScan<'a> {
    inner: collections::btree_map::Range<'a, (uri::Uri, uri::Uri), Vec<u8>>,
    subject: uri::Uri,
    predicate: uri::Uri,
}

impl Iterator for SPOScan<'_> {
    type Item = RawTriple;

    fn next(&mut self) -> Option<Self::Item> {
        let sub_bytes = AsRef::<[u8]>::as_ref(&self.subject);
        let pred_bytes = AsRef::<[u8]>::as_ref(&self.predicate);

        self.inner.next().and_then(|((s, p), o)| {
            if
                AsRef::<[u8]>::as_ref(s).starts_with(sub_bytes)
                && AsRef::<[u8]>::as_ref(p).starts_with(pred_bytes)
            {
                Some(RawTriple {
                    subject: s.clone(),
                    predicate: p.clone(),
                    object: o.clone(),
                })
            } else {
                None
            }
        })
    }
}

pub struct PSOScan<'a> {
    inner: collections::btree_map::Range<'a, (uri::Uri, uri::Uri), Vec<u8>>,
    predicate: uri::Uri,
    subject: uri::Uri,
}

impl Iterator for PSOScan<'_> {
    type Item = RawTriple;

    fn next(&mut self) -> Option<Self::Item> {
        let pred_bytes = AsRef::<[u8]>::as_ref(&self.predicate);
        let sub_bytes = AsRef::<[u8]>::as_ref(&self.subject);

        self.inner.next().and_then(|((p, s), o)| {
            if
                AsRef::<[u8]>::as_ref(p).starts_with(pred_bytes)
                && AsRef::<[u8]>::as_ref(s).starts_with(sub_bytes)
            {
                Some(RawTriple {
                    subject: s.clone(),
                    predicate: p.clone(),
                    object: o.clone(),
                })
            } else {
                None
            }
        })
    }
}

pub struct POSScan<'a> {
    inner: collections::btree_map::Range<'a, (uri::Uri, Vec<u8>), uri::Uri>,
    predicate: uri::Uri,
    object: Vec<u8>,
}

impl Iterator for POSScan<'_> {
    type Item = RawTriple;

    fn next(&mut self) -> Option<Self::Item> {
        let pred_bytes = AsRef::<[u8]>::as_ref(&self.predicate);
        let obj_bytes = self.object.as_slice();

        self.inner.next().and_then(|((p, o), s)| {
            if
                AsRef::<[u8]>::as_ref(p).starts_with(pred_bytes)
                && AsRef::<[u8]>::as_ref(o).starts_with(obj_bytes)
            {
                Some(RawTriple {
                    subject: s.clone(),
                    predicate: p.clone(),
                    object: o.clone(),
                })
            } else {
                None
            }
        })
    }
}


pub struct TupleIter<'a> {
    inner: collections::btree_map::Iter<'a, (uri::Uri, uri::Uri), Vec<u8>>,
}

impl Iterator for TupleIter<'_> {
    type Item = RawTriple;

    fn next(&mut self) -> Option<Self::Item> {
        self.inner.next().map(|((s, p), o)| RawTriple {
            subject: s.clone(),
            predicate: p.clone(),
            object: o.clone(),
        })
    }
}


#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn scan_empty_db() {
        let db = Db::new();
        let mut iter = db.iter();
        assert_eq!(iter.next(), None);
    }

    #[test]
    fn insert_one_and_scan() {
        let mut db = Db::new();
        let triple = Triple::new(
            "http://example.com/graph/p1",
            "http://www.w3.org/1999/02/22-rdf-syntax-ns#type",
            Value::I64(1337)
        ).into();
        db.append(triple).unwrap();

        let mut iter = db.iter();
        assert_eq!(iter.next(), Some(Triple::new(
            "http://example.com/graph/p1",
            "http://www.w3.org/1999/02/22-rdf-syntax-ns#type",
            Value::I64(1337)
        ).into()));
        assert_eq!(iter.next(), None);
    }

    #[test]
    fn test_scan_ordered() {
        let mut db = Db::new();
        let triple1 = Triple::new(
            "http://example.com/graph/apple",
            "http://example.com/graph/fruit-type",
            Value::String("Apple".to_string())
        ).into();

        let triple2 = Triple::new(
            "http://example.com/graph/banana",
            "http://example.com/graph/fruit-color",
            Value::String("yellow".to_string())
        ).into();

        let triple3 = Triple::new(
            "http://example.com/graph/banana",
            "http://example.com/graph/fruit-type",
            Value::String("Banana".to_string())
        ).into();

        let triple4 = Triple::new(
            "http://example.com/graph/cantaloupe",
            "http://example.com/graph/fruit-type",
            Value::String("Cantaloupe".to_string())
        ).into();

        db.append(triple2).unwrap();
        db.append(triple4).unwrap();
        db.append(triple3).unwrap();
        db.append(triple1).unwrap();

        let mut iter = db.iter();
        assert_eq!(iter.next().unwrap().subject, "http://example.com/graph/apple".into());
        
        let out2 = iter.next().unwrap();
        assert_eq!(out2.subject, "http://example.com/graph/banana".into());
        assert_eq!(out2.predicate, "http://example.com/graph/fruit-color".into());

        let out3 = iter.next().unwrap();
        assert_eq!(out3.subject, "http://example.com/graph/banana".into());
        assert_eq!(out3.predicate, "http://example.com/graph/fruit-type".into());

        assert_eq!(iter.next().unwrap().subject, "http://example.com/graph/cantaloupe".into());
    }
    
    #[test]
    fn test_scan_prefix_spo() {
        let mut db = Db::new();
        let triple1 = Triple::new(
            "http://example.com/graph/apple",
            "http://example.com/graph/fruit-color",
            Value::String("Green".to_string())
        ).into();
        let triple2 = Triple::new(
            "http://example.com/graph/apple",
            "http://example.com/graph/fruit-type",
            Value::String("Apple".to_string())
        ).into();
        // Should not match prefix
        let triple3 = Triple::new(
            "http://example.com/graph/banana",
            "http://example.com/graph/fruit-type",
            Value::String("Banana".to_string())
        ).into();
        db.append(triple1).unwrap();
        db.append(triple2).unwrap();
        db.append(triple3).unwrap();

        let mut iter = db.scan_spo((
            "http://example.com/graph/apple".into(),
            "".into()
        ));
        assert_eq!(iter.next().unwrap(), Triple{
            subject: "http://example.com/graph/apple".into(),
            predicate: "http://example.com/graph/fruit-color".into(),
            object: Value::String("Green".to_string()),
        }.into());
        assert_eq!(iter.next().unwrap(), Triple{
            subject: "http://example.com/graph/apple".into(),
            predicate: "http://example.com/graph/fruit-type".into(),
            object: Value::String("Apple".to_string()),
        }.into());

        assert_eq!(None, iter.next());
    }

    #[test]
    fn test_scan_prefix_pso() {
        let mut db = Db::new();
        let triple1 = Triple::new(
            "http://example.com/graph/apple",
            "http://example.com/graph/fruit-color",
            Value::String("Green".to_string())
        ).into();
        let triple2 = Triple::new(
            "http://example.com/graph/apple",
            "http://example.com/graph/fruit-type",
            Value::String("Apple".to_string())
        ).into();
        // Should not match prefix
        let triple3 = Triple::new(
            "http://example.com/graph/banana",
            "http://example.com/graph/fruit-type",
            Value::String("Banana".to_string())
        ).into();
        db.append(triple1).unwrap();
        db.append(triple2).unwrap();
        db.append(triple3).unwrap();

        let mut iter = db.scan_pso((
            "http://example.com/graph/fruit-type".into(),
            "".into()
        ));
        assert_eq!(iter.next().unwrap(), Triple{
            subject: "http://example.com/graph/apple".into(),
            predicate: "http://example.com/graph/fruit-type".into(),
            object: Value::String("Apple".to_string()),
        }.into());
        assert_eq!(iter.next().unwrap(), Triple{
            subject: "http://example.com/graph/banana".into(),
            predicate: "http://example.com/graph/fruit-type".into(),
            object: Value::String("Banana".to_string()),
        }.into());

        assert_eq!(None, iter.next());
    }

    #[test]
    fn test_scan_prefix_pos() {
        let mut db = Db::new();
        let triple1 = Triple::new(
            "http://example.com/graph/apple",
            "http://example.com/graph/fruit-color",
            Value::String("Green".to_string())
        ).into();
        let triple2 = Triple::new(
            "http://example.com/graph/cucumber",
            "http://example.com/graph/fruit-color",
            Value::String("Green".to_string())
        ).into();
        // Should not match prefix
        let triple3 = Triple::new(
            "http://example.com/graph/apple",
            "http://example.com/graph/fruit-type",
            Value::String("Apple".to_string())
        ).into();
        db.append(triple1).unwrap();
        db.append(triple2).unwrap();
        db.append(triple3).unwrap();

        let mut buf = Vec::new();
        Value::String("Green".to_string()).serialize(&mut buf).unwrap();

        let mut iter = db.scan_pos((
            "http://example.com/graph/fruit-color".into(),
            buf,
        ));
        assert_eq!(iter.next().unwrap(), Triple{
            subject: "http://example.com/graph/cucumber".into(),
            predicate: "http://example.com/graph/fruit-color".into(),
            object: Value::String("Green".to_string()),
        }.into());
        assert_eq!(iter.next().unwrap(), Triple{
            subject: "http://example.com/graph/apple".into(),
            predicate: "http://example.com/graph/fruit-color".into(),
            object: Value::String("Green".to_string()),
        }.into());

        assert_eq!(None, iter.next());
    }
}
