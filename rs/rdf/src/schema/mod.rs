#[derive(Debug, PartialEq, Eq, Clone)]
pub enum Type {
    Bool,
    Int8,
    Int16,
    Int32,
    Int64,
    UInt8,
    UInt16,
    UInt32,
    UInt64,

    Unicode,
    Binary,

    Array {
        Element: Box<Type>,
    },
    Map {
        Key: Box<Type>,
        Value: Box<Type>,
    },
    Tuple {
        Members: Vec<Type>,
    },

    Unknown,
}

impl Type {
    fn size_hint(&self) -> usize {
        match self {
            Type::Bool => 1,
            Type::Int8 | Type::UInt8 => 1,
            Type::Int16 | Type::UInt16 => 2,
            Type::Int32 | Type::UInt32 => 4,
            Type::Int64 | Type::UInt64 => 8,
            // The following types are dynamically sized, and there is not a logical minimum size
            Type::Unicode | Type::Binary | Type::Array{ .. } | Type::Map{ .. } | Type::Tuple{ .. } => 0,
            Type::Unknown => 0,
        }
    }
}

#[derive(Debug, PartialEq)]
pub struct Schema {
    fields: Vec<Field>,
}

impl Schema {
    pub fn size_hint(&self) -> usize {
        self.fields.iter().map(|field| field.size_hint()).sum()
    }

    pub fn get(&self, field_num: usize) -> Option<&Field> {
        self.fields.get(field_num)
    }

    pub fn len(&self) -> usize {
        self.fields.len()
    }
}

#[derive(Debug, PartialEq)]
pub struct Field {
    pub field_type: Type,
    pub field_name: Option<String>,
}

impl Field {
    pub fn new<S: Into<String>>(field_name: S, field_type: Type) -> Self {
        Field { field_type: field_type, field_name: Some(field_name.into()) }
    }

    pub fn new_unnamed(field_type: Type) -> Self {
        Field { field_type: field_type, field_name: None }
    }

    fn size_hint(&self) -> usize {
        self.field_type.size_hint()
    }
}

struct SchemaBuilder {
    fields: Vec<Field>
}

impl SchemaBuilder {
    fn new() -> Self {
        SchemaBuilder { fields: Vec::new() }
    }

    fn with_field(mut self, field: Field) -> Self {
        self.fields.push(field);
        self
    }

    fn build(self) -> Schema {
        Schema { fields: self.fields }
    }
}


#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_build() {
        let mut sb = SchemaBuilder::new();
        let s = sb
            .with_field(Field::new("id", Type::Int64))
            .with_field(Field::new("name", Type::Unicode))
            .build();

        assert_eq!(s, Schema {
            fields: vec![
                Field {
                    field_name: Some("id".into()),
                    field_type: Type::Int64,
                },
                Field {
                    field_name: Some("name".into()),
                    field_type: Type::Unicode,
                },
            ]
        })
    }
}
