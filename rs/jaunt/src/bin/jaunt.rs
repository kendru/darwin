use jaunt::{Db, Triple};
use jaunt::value::{Value};
use jaunt::error;


fn main() -> error::Result<()> {
   let mut db = Db::new();
   let triple = Triple::new(
      "http://example.com/graph/p1",
      "http://www.w3.org/1999/02/22-rdf-syntax-ns#type",
      Value::I64(1337)
   );
   println!("Triple: {:?}", triple);
   db.append(triple.into())?;

   for triple in db.iter() {
      let triple: Triple = triple.clone().into();
      println!("{:?}", triple)
   }


   Ok(())
}
