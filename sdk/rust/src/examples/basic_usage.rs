use literedis::{Pool, RedisError, Commands};
use std::time::Duration;
use tokio;

#[tokio::main]
async fn main() -> Result<(), RedisError> {
    // 创建连接池
    let pool = Pool::new("127.0.0.1:6379", 10).await?;
    
    // 从连接池获取连接
    let mut conn = pool.get().await?;
    
    // 字符串操作
    conn.set("key1", "value1").await?;
    let value = conn.get("key1").await?;
    println!("key1: {:?}", value);

    // 设置带过期时间的键
    conn.set_ex("key2", "value2", Duration::from_secs(5)).await?;
    
    // 获取TTL
    let ttl = conn.ttl("key2").await?;
    println!("TTL of key2: {} seconds", ttl);

    // 自增操作
    conn.set("counter", "10").await?;
    let new_value = conn.incr("counter").await?;
    println!("Incremented counter: {}", new_value);

    // 删除操作
    let deleted = conn.del(&["key1", "key2"]).await?;
    println!("Deleted {} key(s)", deleted);

    // 批量操作
    conn.set("batch_key1", "batch_value1").await?;
    conn.set("batch_key2", "batch_value2").await?;
    let values = conn.mget(&["batch_key1", "batch_key2", "nonexistent_key"]).await?;
    println!("Batch get results: {:?}", values);

    // 列表操作
    conn.lpush("mylist", &["item1", "item2", "item3"]).await?;
    let list_length = conn.llen("mylist").await?;
    println!("List length: {}", list_length);
    let list_items = conn.lrange("mylist", 0, -1).await?;
    println!("List items: {:?}", list_items);

    // 哈希表操作
    conn.hset("myhash", "field1", "value1").await?;
    conn.hset("myhash", "field2", "value2").await?;
    let hash_value = conn.hget("myhash", "field1").await?;
    println!("Hash value for field1: {:?}", hash_value);
    let all_fields = conn.hgetall("myhash").await?;
    println!("All hash fields: {:?}", all_fields);

    Ok(())
}
