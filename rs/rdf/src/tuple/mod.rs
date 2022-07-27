use std::mem::size_of;
use crate::schema::{Schema, Type};

type TypeTag = u8;

pub struct Tuple <'a> {
    schema: &'a Schema,
    buf: Vec<u8>,
    // Index in `buf` for next write.
    offset: usize,
    // Index of field in schema for next write.
    field_ofset: u16,
}

impl<'a> Tuple<'a> {
    fn new(schema: &Schema) -> Tuple {
        Tuple {
            buf: Vec::with_capacity(schema.size_hint() + size_of::<TypeTag>() * schema.len()),
            schema,
        }
    }

    // TODO: Introduce error types.
    fn write_unicode<S: AsRef<str>>(&mut self, val: S) -> Result<(), String> {
        if let Some(field) = self.schema.get(field_num) {
            if field.field_type != Type::Unicode {
                return Err(format!("Expected {:?} for field {} but got {:?}", Type::Unicode, field_num, field.field_type));
            }
        } else {
            return Err(format!("Field out of bounds: {}", field_num));
        }

        Ok(())
    }
}

fn type_tag(typ: Type) -> TypeTag {
    match typ {
        Type::Bool => 1,
        Type::Int8 => 2,
        Type::Int16 => 3,
        Type::Int32 => 4,
        Type::Int64 => 5,
        Type::UInt8 => 6,
        Type::UInt16 => 7,
        Type::UInt32 => 8,
        Type::UInt64 => 9,
        Type::Unicode => 10,
        Type::Binary => 11,
        Type::Array { .. } => 12,
        Type::Map { .. } => 13,
        Type::Tuple { .. } => 14,
        Type::Unknown => 255,
    }
}
