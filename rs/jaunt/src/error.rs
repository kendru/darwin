use std::fmt;
use std::string::FromUtf8Error;
use std::io::{Error as IoError};

#[derive(Debug)]
pub enum Error {
    Todo(String),
    ParseError(String),
    IOError(IoError),
    CannotCoerceValue {
        expected: &'static str,
        encountered: &'static str,
    },
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Error::CannotCoerceValue {
                expected,
                encountered,
            } => write!(f, "expected type {} but got {}", expected, encountered),
            Error::IOError(err) => write!(f, "IO Error: {}", err),
            Error::ParseError(msg) => write!(f, "Parse Error: {}", msg),
            Error::Todo(msg) => write!(f, "TODO Error: {}", msg),
        }
    }
}

impl From<FromUtf8Error> for Error {
    fn from(_: FromUtf8Error) -> Self {
        Error::ParseError("invalid utf-8 string".to_string())
    }
}

impl From<IoError> for Error {
    fn from(inner: IoError) -> Self {
        Error::IOError(inner)
    }
}

pub type Result<T> = std::result::Result<T, Error>;
