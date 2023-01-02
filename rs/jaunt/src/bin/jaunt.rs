use jaunt::{Db, RawTriple};
use jaunt::error;

fn main() -> error::Result<()> {
   let mut db = Db::new();
   db.append(RawTriple::new("http://example.com/graph/p1", "http://www.w3.org/1999/02/22-rdf-syntax-ns#type", "http://example.com/graph/Person"))?;

   Ok(())
}
