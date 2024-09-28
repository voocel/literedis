use crate::client::{AsyncClient, Client};
use crate::error::RedisError;
use std::sync::Arc;
use tokio::sync::{Mutex, Semaphore};

pub struct Pool {
    clients: Vec<Arc<Mutex<Client>>>,
    semaphore: Arc<Semaphore>,
}

impl Pool {
    pub async fn new(address: &str, size: usize) -> Result<Self, RedisError> {
        let mut clients = Vec::with_capacity(size);
        for _ in 0..size {
            clients.push(Arc::new(Mutex::new(Client::connect(address).await?)));
        }
        Ok(Self {
            clients,
            semaphore: Arc::new(Semaphore::new(size)),
        })
    }

    pub async fn get(&self) -> Result<PooledClient, RedisError> {
        let permit = self.semaphore.clone().acquire_owned().await.map_err(|e| RedisError::ConnectionError(e.to_string()))?;
        let client = self.clients[permit.0].clone();
        Ok(PooledClient {
            client,
            _permit: permit,
        })
    }
}

pub struct PooledClient {
    client: Arc<Mutex<Client>>,
    _permit: tokio::sync::OwnedSemaphorePermit,
}

#[async_trait::async_trait]
impl AsyncClient for PooledClient {
    async fn connect(_address: &str) -> Result<Self, RedisError> {
        unimplemented!("Use Pool::new instead")
    }

    async fn execute(&mut self, command: &[u8]) -> Result<Value, RedisError> {
        let mut client = self.client.lock().await;
        client.execute(command).await
    }
}

impl Commands for PooledClient {}
