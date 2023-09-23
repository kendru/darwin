use std::io::Write;
use byteorder::{WriteBytesExt, BigEndian, ByteOrder}; 
use super::error;
use super::uri;

const TAG_URI: u8 = 0;
const TAG_STRING: u8 = 1;
const TAG_LANG_STRING: u8 = 2;
const TAG_U64: u8 = 3;
const TAG_I64: u8 = 4;
const TAG_F64: u8 = 5;
const TAG_U32: u8 = 6;
const TAG_I32: u8 = 7;
const TAG_F32: u8 = 8;
const TAG_U16: u8 = 9;
const TAG_I16: u8 = 10;
const TAG_U8: u8 = 11;
const TAG_I8: u8 = 12;
const TAG_BOOL: u8 = 13;
const TAG_DECIMAL: u8 = 14;
const TAG_TYPED_LITERAL: u8 = 15;
const TAG_ILL_FORMED: u8 = 255;

pub trait Identity {
    fn is_identical(&self, other: &Self) -> bool;
}

#[derive(Debug)]
pub enum Value {
    Uri(uri::Uri),
    String(String),
    LangString { str: String, language_tag: String },
    U64(u64),
    I64(i64),
    F64(f64),
    U32(u32),
    I32(i32),
    F32(f32),
    U16(u16),
    I16(i16),
    U8(u8),
    I8(i8),
    Bool(bool),

    Decimal(String),

    TypedLiteral { source: String, datatype: uri::Uri },

    IllFormed(String),
}

impl Value {
    pub fn type_name(&self) -> &'static str {
        match self {
            Value::Uri(_) => "uri",
            Value::String(_) => "string",
            Value::LangString { .. } => "langString",
            Value::U64(_) => "u64",
            Value::I64(_) => "i64",
            Value::F64(_) => "f64",
            Value::U32(_) => "u32",
            Value::I32(_) => "i32",
            Value::F32(_) => "f32",
            Value::U16(_) => "u16",
            Value::I16(_) => "i16",
            Value::U8(_) => "u8",
            Value::I8(_) => "i8",
            Value::Bool(_) => "bool",

            Value::Decimal { .. } => "decimal",

            Value::TypedLiteral { .. } => "typedLiteral",

            Value::IllFormed(_) => "illFormedValue",
        }
    }

    pub fn type_code(&self) -> u8 {
        match self {
            Value::Uri(_) => TAG_URI,
            Value::String(_) => TAG_STRING,
            Value::LangString { .. } => TAG_LANG_STRING,
            Value::U64(_) => TAG_U64,
            Value::I64(_) => TAG_I64,
            Value::F64(_) => TAG_F64,
            Value::U32(_) => TAG_U32,
            Value::I32(_) => TAG_I32,
            Value::F32(_) => TAG_F32,
            Value::U16(_) => TAG_U16,
            Value::I16(_) => TAG_I16,
            Value::U8(_) => TAG_U8,
            Value::I8(_) =>     TAG_I8,
            Value::Bool(_) => TAG_BOOL,

            Value::Decimal { .. } => TAG_DECIMAL,

            Value::TypedLiteral { .. } => TAG_TYPED_LITERAL,

            Value::IllFormed(_) => TAG_ILL_FORMED,
        }
    }

    pub fn serialize<W: Write>(&self, w: &mut W) -> error::Result<usize> {
        let mut bytes_written = 0;
        w.write_u8(self.type_code())?;
        bytes_written += 1;
        match self {
            Value::Uri(u) => {
                bytes_written += write_escaped_bytes(w, u.as_ref())?;
            },
            Value::String(s) => {
                w.write_all(s.as_bytes())?;
                bytes_written += s.as_bytes().len();
            },
            Value::LangString { str, language_tag } => {
                w.write_all(language_tag.as_bytes())?;
                bytes_written += language_tag.as_bytes().len();
                w.write_all(str.as_bytes())?;
                bytes_written += str.as_bytes().len();
            },
            Value::U64(u) => {
                w.write_u64::<BigEndian>(*u)?;
                bytes_written += 8;
            },
            Value::I64(i) => {
                w.write_i64::<BigEndian>(*i)?;
                bytes_written += 8;
            },
            Value::F64(f) => {
                w.write_f64::<BigEndian>(*f)?;
                bytes_written += 8;
            },
            Value::U32(u) => {
                w.write_u32::<BigEndian>(*u)?;
                bytes_written += 4;
            },
            Value::I32(i) => {
                w.write_i32::<BigEndian>(*i)?;
                bytes_written += 4;
            },
            Value::F32(f) => {
                w.write_f32::<BigEndian>(*f)?;
                bytes_written += 4;
            },
            Value::U16(u) => {
                w.write_u16::<BigEndian>(*u)?;
                bytes_written += 2;
            },
            Value::I16(i) => {
                w.write_i16::<BigEndian>(*i)?;
                bytes_written += 2;
            },
            Value::U8(u) => {
                w.write_u8(*u)?;
                bytes_written += 1;
            },
            Value::I8(i) => {
                w.write_i8(*i)?;
                bytes_written += 1;
            },
            Value::Bool(b) => {
                w.write_u8(*b as u8)?;
                bytes_written += 1;
            },
            Value::Decimal(n) =>
                return Err(error::Error::Todo(
                    format!("Decimal not implemented: {}", n))),
            Value::TypedLiteral { source, datatype } =>
                return Err(error::Error::Todo(
                    format!("TypedLiteral not implemented: {} {:?}", source, datatype))),
            Value::IllFormed(s) =>
                return Err(error::Error::Todo(
                    format!("IllFormed not implemented: {}", s))),
        }

        Ok(bytes_written)
    }

    pub fn parse(buf: &[u8]) -> error::Result<Value> {
        if buf.is_empty() {
            return Err(error::Error::Todo("no data to parse".to_string()));
        }
        let head = buf[0];
        let tail = &buf[1..];
        match head {
            TAG_URI => {
                let str = String::from_utf8(tail.to_vec())?;
                let u = uri::Uri::from(str);
                Ok(Value::Uri(u))
            },
            TAG_STRING => {
                let str = String::from_utf8(tail.to_vec())?;
                Ok(Value::String(str))
            },
            TAG_LANG_STRING => {
                if tail.len() < 3 {
                    return Ok(Value::IllFormed("Invalid langString".to_string()))
                }
                let language_tag = String::from_utf8(tail[0..2].to_vec())?;
                let str = String::from_utf8(tail[2..].to_vec())?;
                
                Ok(Value::LangString { str, language_tag })
            },
            TAG_U64 => {
                let u = BigEndian::read_u64(tail);
                Ok(Value::U64(u))
            },
            TAG_I64 => {
                let i = BigEndian::read_i64(tail);
                Ok(Value::I64(i))
            },
            // TODO: Implement the rest of the types.
            _ => Ok(Value::IllFormed("unknown type code".to_string())),
        }
    }
}

impl Identity for Value {
    fn is_identical(&self, other: &Self) -> bool {
        match (self, other) {
            (Value::Uri(a), Value::Uri(b)) => *a == *b,
            (Value::String(a), Value::String(b)) => *a == *b,
            (
                Value::LangString {
                    str: str_a,
                    language_tag: language_tag_a,
                },
                Value::LangString {
                    str: str_b,
                    language_tag: language_tag_b,
                },
            ) => *str_a == *str_b && *language_tag_a == *language_tag_b,
            // How do we want a numeric tower to behave?
            (Value::I64(a), Value::I64(b)) => *a == *b,
            (Value::I32(a), Value::I32(b)) => *a == *b,
            (Value::I16(a), Value::I16(b)) => *a == *b,
            (Value::I8(a), Value::I8(b)) => *a == *b,
            (Value::U64(a), Value::U64(b)) => *a == *b,
            (Value::U32(a), Value::U32(b)) => *a == *b,
            (Value::U16(a), Value::U16(b)) => *a == *b,
            (Value::U8(a), Value::U8(b)) => *a == *b,
            (Value::F32(a), Value::F32(b)) => !a.is_nan() && !b.is_nan() && unsafe {
                let a_bytes = std::slice::from_raw_parts(a as *const f32 as *const u8, 4);
                let b_bytes = std::slice::from_raw_parts(b as *const f32 as *const u8, 4);
                a_bytes == b_bytes
            },
            (Value::F64(a), Value::F64(b)) => !a.is_nan() && !b.is_nan() && unsafe {
                let a_bytes = std::slice::from_raw_parts(a as *const f64 as *const u8, 8);
                let b_bytes = std::slice::from_raw_parts(b as *const f64 as *const u8, 8);
                println!("{:?} == {:?}", a_bytes, b_bytes);
                a_bytes == b_bytes
            },
            // TODO
            _ => false,
        }
    }
}

impl PartialEq for Value {
    fn eq(&self, other: &Self) -> bool {
        match (self, other) {
            (Value::F32(a), Value::F32(b)) => a == b,
            (Value::F64(a), Value::F64(b)) => a == b,
            _ => self.is_identical(other)
        }
    }
}

impl TryInto<String> for Value {
    type Error = error::Error;

    fn try_into(self) -> Result<String, Self::Error> {
        match self {
            Value::String(s) => Ok(s),
            Value::LangString { str, .. } => Ok(str),
            Value::Uri(uri) => Ok(uri.into()),
            Value::IllFormed(source) => Ok(format!("\"{}\"", source)), // TODO: escape source.
            other => Err(error::Error::CannotCoerceValue {
                expected: "string",
                encountered: other.type_name(),
            }),
        }
    }
}

impl From<String> for Value {
    fn from(value: String) -> Self {
        Value::String(value)
    }
}

impl From<&str> for Value {
    fn from(value: &str) -> Self {
        Value::String(value.to_string())
    }
}

impl From<u64> for Value {
    fn from(value: u64) -> Self {
        Value::U64(value)
    }
}

impl From<i64> for Value {
    fn from(value: i64) -> Self {
        Value::I64(value)
    }
}

impl From<f64> for Value {
    fn from(value: f64) -> Self {
        Value::F64(value)
    }
}

impl From<f32> for Value {
    fn from(value: f32) -> Self {
        Value::F32(value)
    }
}

// TODO: Additional from impls.

fn write_escaped_bytes<W: Write>(w: &mut W, bytes: &[u8]) -> error::Result<usize> {
    let mut bytes_written = 0;
    // TODO: This is a bit inefficient. Instead of calling write for each byte, we should
    // write as many bytes as possible at once. We only need to call write when we encounter
    // a null byte or the end of the slice.
    for byte in bytes {
        match byte {
            0x00 => {
                // Note that since each string of bytes is written with an unescaped null byte,
                // the escaping value (0xff) must never be used as a type tag.
                bytes_written += w.write(&[0x00, 0xff])?;
            },
            &b => {
                bytes_written += w.write(&[b])?;
            }
        }
    }
    bytes_written += w.write(&[0x00])?;
    
    Ok(bytes_written)
}
