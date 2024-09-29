# LiteRedis Rust SDK

This is the Rust SDK for LiteRedis, a lightweight Redis-like in-memory storage system.

## Installation

Add this to your `Cargo.toml`:

```toml
[dependencies]
literedis = { git = "https://github.com/voocel/literedis.git" }
tokio = { version = "1.0", features = ["full"] }
```


## Usage

Here's a basic example of how to use the LiteRedis Rust SDK:

```rust
use literedis::{Pool, RedisError, Commands};
use tokio;
#[tokio::main]
async fn main() -> Result<(), RedisError> {
// Create a connection pool
let pool = Pool::new("127.0.0.1:6379", 10).await?;
// Get a connection from the pool
let mut conn = pool.get().await?;
// Set a key
conn.set("my_key", "my_value").await?;
// Get a key
let value = conn.get("my_key").await?;
println!("Value: {:?}", value);
// Delete a key
let deleted = conn.del(&["my_key"]).await?;
println!("Deleted {} key(s)", deleted);
// Increment a counter
let new_value = conn.incr("counter").await?;
println!("Incremented counter: {}", new_value);
// Set a key with expiration
use std::time::Duration;
conn.set_ex("expiring_key", "expiring_value", Duration::from_secs(10)).await?;
// Get the TTL of a key
let ttl = conn.ttl("expiring_key").await?;
println!("TTL of expiring_key: {} seconds", ttl);
Ok(())
}
```


For more examples, check the `examples` directory in the source code.

## Features

- Asynchronous API using Tokio
- Connection pooling
- Basic Redis commands (SET, GET, DEL, INCR, EXPIRE, TTL, SETEX)
- Error handling
