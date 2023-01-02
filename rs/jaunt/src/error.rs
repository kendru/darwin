use std::fmt;

#[derive(Debug)]
pub enum Error {
    Todo(String),
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Error::Todo(msg) => write!(f, "TODO Error: {}", msg),
        }
    }
}

pub type Result<T> = std::result::Result<T, Error>;
