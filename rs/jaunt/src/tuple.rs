/// This module implements the tuple type.
/// A tuple is a fixed-length collection of values that can be of different types.
/// Our representation is inspired by
/// [FoundationDB's Tuple Layer](https://github.com/apple/foundationdb/blob/main/design/tuple.md).
/// 
/// A tuple acts as a descriptor that can be used to encode and decode values
/// into and from a byte representation.By their nature, tuples provide unsafe
/// interpretation of memory and must be used with care.
/// 
/// # Examples
/// 
/// ```
/// use std::convert::TryFrom;
/// use std::io::Cursor;
/// use std::str::FromStr;
/// use std::string::ToString;
/// 
/// use crate::tuple::{Tuple, Value};
/// 
/// let tuple = Tuple::from(vec![
///    Value::String("hello".to_string()),
///   Value::I64(42),
/// ]);
/// 
/// let mut cursor = Cursor::new(Vec::new());
/// tuple.encode(&mut cursor).unwrap();
/// 
/// let mut cursor = Cursor::new(cursor.into_inner());
/// let decoded_tuple = Tuple::decode(&mut cursor).unwrap();
/// 
/// assert_eq!(tuple, decoded_tuple);
/// ```
/// 
/// # Safety
/// 
/// The tuple type is unsafe because it provides unchecked access to memory.
/// This means that it is possible to create a tuple that is invalid.
/// 
/// For example, the following code will panic:
/// 
/// ```should_panic
/// use std::convert::TryFrom;
/// use std::io::Cursor;
/// use std::str::FromStr;
/// use std::string::ToString;
/// 
/// use crate::tuple::{Tuple, Value};
///     
/// let tuple = Tuple::from(vec![
///    Value::String("hello".to_string()),
///  Value::I64(42),
/// ]);
/// 
/// let mut cursor = Cursor::new(Vec::new());
/// tuple.encode(&mut cursor).unwrap();
/// 
/// let mut cursor = Cursor::new(cursor.into_inner());
/// let decoded_tuple = Tuple::decode(&mut cursor).unwrap();
/// 
/// assert_eq!(tuple, decoded_tuple);
/// ```

struct Element {
    data_type: DataType,
    nullable: bool,
}

struct Tuple {
    schema: Vec<Element>
}