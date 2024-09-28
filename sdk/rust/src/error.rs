use std::error::Error;
use std::fmt;
use std::io;

#[derive(Debug)]
pub enum RedisError {
    IoError(io::Error),
    ParseError(String),
    ProtocolError(String),
    ConnectionError(String),
}

impl fmt::Display for RedisError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            RedisError::IoError(err) => write!(f, "IO error: {}", err),
            RedisError::ParseError(err) => write!(f, "Parse error: {}", err),
            RedisError::ProtocolError(err) => write!(f, "Protocol error: {}", err),
            RedisError::ConnectionError(err) => write!(f, "Connection error: {}", err),
        }
    }
}

impl Error for RedisError {
    fn source(&self) -> Option<&(dyn Error + 'static)> {
        match self {
            RedisError::IoError(err) => Some(err),
            _ => None,
        }
    }
}

impl From<io::Error> for RedisError {
    fn from(err: io::Error) -> Self {
        RedisError::IoError(err)
    }
}
