// src/client.rs
use crate::error::RedisError;
use crate::resp::{RespCodec, Value};
use async_trait::async_trait;
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use tokio::net::TcpStream;

pub struct Client {
    stream: TcpStream,
}

#[async_trait]
pub trait AsyncClient {
    async fn connect(address: &str) -> Result<Self, RedisError>
    where
        Self: Sized;
    async fn execute(&mut self, command: &[u8]) -> Result<Value, RedisError>;
}

#[async_trait]
impl AsyncClient for Client {
    async fn connect(address: &str) -> Result<Self, RedisError> {
        let stream = TcpStream::connect(address).await?;
        Ok(Self { stream })
    }

    async fn execute(&mut self, command: &[u8]) -> Result<Value, RedisError> {
        self.stream.write_all(command).await?;
        let mut reader = tokio::io::BufReader::new(&mut self.stream);
        Value::decode(&mut reader)
    }
}
