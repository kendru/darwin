use std::fmt;

#[derive(PartialEq, Eq, PartialOrd, Ord, Clone)]
pub struct Uri(String);

impl From<String> for Uri {
    fn from(value: String) -> Self {
        Uri(value)
    }
}

impl Into<String> for Uri {
    fn into(self) -> String {
        self.0
    }
}

impl From<&str> for Uri {
    fn from(value: &str) -> Self {
        Uri(value.to_string())
    }
}

impl AsRef<str> for Uri {
    fn as_ref(&self) -> &str {
        &self.0
    }
}

impl AsRef<[u8]> for Uri {
    fn as_ref(&self) -> &[u8] {
        &self.0.as_bytes()
    }
}

impl fmt::Debug for Uri {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "<{}>", self.0)
    }
}
