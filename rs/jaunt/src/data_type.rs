/// This package contains the specification of all native data types
/// along with the implementations of encoding and decoding logic.
/// 
/// # Examples
/// 
/// ``FIXME
/// use std::convert::TryFrom;
/// use std::io::Cursor;
/// use std::str::FromStr;
/// 
/// use jaunt::data_type::{DataType, Value};
/// 
/// let data_type = DataType::from_str("string").unwrap();
/// let value = Value::String("hello".to_string());
/// 
/// let mut cursor = Cursor::new(Vec::new());
/// data_type.encode(&value, &mut cursor).unwrap();
/// 
/// let mut cursor = Cursor::new(cursor.into_inner());
/// let decoded_value = data_type.decode(&mut cursor).unwrap();
/// 
/// assert_eq!(value, decoded_value);
/// ```

use std::io::{Read, Write};
use crate::error;

trait DataType<T> {
    fn encode<W: Write>(&self, value: &T, writer: W) -> error::Result<()>;
    fn decode<R: Read>(&self, reader: R) -> error::Result<T>;
}

struct Uint8;

impl DataType<u8> for Uint8 {
    fn encode<W: Write>(&self, value: &u8, mut writer: W) -> error::Result<()> {
        writer.write_all(&[*value])?;
        Ok(())
    }

    fn decode<R: Read>(&self, mut reader: R) -> error::Result<u8> {
        let mut buffer = [0u8; 1];
        reader.read_exact(&mut buffer)?;
        Ok(buffer[0])
    }
}

struct Uint16;

impl DataType<u16> for Uint16 {
    fn encode<W: Write>(&self, value: &u16, mut writer: W) -> error::Result<()> {
        writer.write_all(&value.to_be_bytes())?;
        Ok(())
    }

    fn decode<R: Read>(&self, mut reader: R) -> error::Result<u16> {
        let mut buffer = [0u8; 2];
        reader.read_exact(&mut buffer)?;
        Ok(u16::from_be_bytes(buffer))
    }
}