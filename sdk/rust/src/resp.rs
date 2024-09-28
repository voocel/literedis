use crate::error::RedisError;
use std::io::{BufRead, Write};

pub enum Value {
    SimpleString(String),
    Error(String),
    Integer(i64),
    BulkString(Option<Vec<u8>>),
    Array(Vec<Value>),
}

pub trait RespCodec: Sized {
    fn encode<W: Write>(&self, writer: &mut W) -> Result<(), RedisError>;
    fn decode<R: BufRead>(reader: &mut R) -> Result<Self, RedisError>;
}

impl RespCodec for Value {
    fn encode<W: Write>(&self, writer: &mut W) -> Result<(), RedisError> {
        match self {
            Value::SimpleString(s) => {
                writer.write_all(b"+")?;
                writer.write_all(s.as_bytes())?;
                writer.write_all(b"\r\n")?;
            }
            Value::Error(s) => {
                writer.write_all(b"-")?;
                writer.write_all(s.as_bytes())?;
                writer.write_all(b"\r\n")?;
            }
            Value::Integer(i) => {
                writer.write_all(b":")?;
                writer.write_all(i.to_string().as_bytes())?;
                writer.write_all(b"\r\n")?;
            }
            Value::BulkString(Some(s)) => {
                writer.write_all(b"$")?;
                writer.write_all(s.len().to_string().as_bytes())?;
                writer.write_all(b"\r\n")?;
                writer.write_all(s)?;
                writer.write_all(b"\r\n")?;
            }
            Value::BulkString(None) => {
                writer.write_all(b"$-1\r\n")?;
            }
            Value::Array(arr) => {
                writer.write_all(b"*")?;
                writer.write_all(arr.len().to_string().as_bytes())?;
                writer.write_all(b"\r\n")?;
                for item in arr {
                    item.encode(writer)?;
                }
            }
        }
        Ok(())
    }

    fn decode<R: BufRead>(reader: &mut R) -> Result<Self, RedisError> {
        let mut line = String::new();
        reader.read_line(&mut line)?;
        match line.chars().next() {
            Some('+') => Ok(Value::SimpleString(line[1..].trim().to_string())),
            Some('-') => Ok(Value::Error(line[1..].trim().to_string())),
            Some(':') => {
                let num: i64 = line[1..].trim().parse()?;
                Ok(Value::Integer(num))
            }
            Some('$') => {
                let len: i64 = line[1..].trim().parse()?;
                if len == -1 {
                    Ok(Value::BulkString(None))
                } else {
                    let mut buf = vec![0; len as usize];
                    reader.read_exact(&mut buf)?;
                    reader.read_line(&mut String::new())?; // Discard CRLF
                    Ok(Value::BulkString(Some(buf)))
                }
            }
            Some('*') => {
                let len: usize = line[1..].trim().parse()?;
                let mut arr = Vec::with_capacity(len);
                for _ in 0..len {
                    arr.push(Value::decode(reader)?);
                }
                Ok(Value::Array(arr))
            }
            _ => Err(RedisError::ProtocolError("Unknown reply type".to_string())),
        }
    }
}
