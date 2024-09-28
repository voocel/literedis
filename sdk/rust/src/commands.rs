use crate::client::AsyncClient;
use crate::error::RedisError;
use crate::resp::Value;

#[async_trait::async_trait]
pub trait Commands: AsyncClient {
    async fn set(&mut self, key: &str, value: &str) -> Result<(), RedisError> {
        let command = format!("*3\r\n$3\r\nSET\r\n${}\r\n{}\r\n${}\r\n{}\r\n", 
                              key.len(), key, value.len(), value);
        match self.execute(command.as_bytes()).await? {
            Value::SimpleString(s) if s == "OK" => Ok(()),
            _ => Err(RedisError::ProtocolError("Unexpected response".to_string())),
        }
    }

    async fn get(&mut self, key: &str) -> Result<Option<String>, RedisError> {
        let command = format!("*2\r\n$3\r\nGET\r\n${}\r\n{}\r\n", key.len(), key);
        match self.execute(command.as_bytes()).await? {
            Value::BulkString(Some(bytes)) => Ok(Some(String::from_utf8(bytes)?)),
            Value::BulkString(None) => Ok(None),
            _ => Err(RedisError::ProtocolError("Unexpected response".to_string())),
        }
    }

  async fn del(&mut self, keys: &[&str]) -> Result<i64, RedisError> {
        let mut command = format!("*{}\r\n$3\r\nDEL\r\n", keys.len() + 1);
        for key in keys {
            command.push_str(&format!("${}\r\n{}\r\n", key.len(), key));
        }
        match self.execute(command.as_bytes()).await? {
            Value::Integer(n) => Ok(n),
            _ => Err(RedisError::ProtocolError("Unexpected response".to_string())),
        }
    }

    async fn incr(&mut self, key: &str) -> Result<i64, RedisError> {
        let command = format!("*2\r\n$4\r\nINCR\r\n${}\r\n{}\r\n", key.len(), key);
        match self.execute(command.as_bytes()).await? {
            Value::Integer(n) => Ok(n),
            _ => Err(RedisError::ProtocolError("Unexpected response".to_string())),
        }
    }

    async fn expire(&mut self, key: &str, seconds: u64) -> Result<bool, RedisError> {
        let command = format!("*3\r\n$6\r\nEXPIRE\r\n${}\r\n{}\r\n${}\r\n{}\r\n",
                              key.len(), key, seconds.to_string().len(), seconds);
        match self.execute(command.as_bytes()).await? {
            Value::Integer(1) => Ok(true),
            Value::Integer(0) => Ok(false),
            _ => Err(RedisError::ProtocolError("Unexpected response".to_string())),
        }
    }

    async fn ttl(&mut self, key: &str) -> Result<i64, RedisError> {
        let command = format!("*2\r\n$3\r\nTTL\r\n${}\r\n{}\r\n", key.len(), key);
        match self.execute(command.as_bytes()).await? {
            Value::Integer(n) => Ok(n),
            _ => Err(RedisError::ProtocolError("Unexpected response".to_string())),
        }
    }

    async fn set_ex(&mut self, key: &str, value: &str, expiration: Duration) -> Result<(), RedisError> {
        let seconds = expiration.as_secs();
        let command = format!("*4\r\n$5\r\nSETEX\r\n${}\r\n{}\r\n${}\r\n{}\r\n${}\r\n{}\r\n",
                              key.len(), key, seconds.to_string().len(), seconds, value.len(), value);
        match self.execute(command.as_bytes()).await? {
            Value::SimpleString(s) if s == "OK" => Ok(()),
            _ => Err(RedisError::ProtocolError("Unexpected response".to_string())),
        }
    }
}

impl Commands for crate::client::Client {}
