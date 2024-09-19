package main

import (
	"fmt"
	"log"
	"time"

	"github.com/voocel/literedis"
)

func main() {
	// 创建连接池
	pool := literedis.NewPool("localhost:6379", 10, 5, 30*time.Second)
	defer pool.Close()

	// 从连接池获取客户端
	client, err := pool.Get()
	if err != nil {
		log.Fatalf("Failed to get client: %v", err)
	}
	defer pool.Put(client)

	// 字符串操作
	if err := client.Set("key1", "value1", 0); err != nil {
		log.Fatalf("Failed to set key1: %v", err)
	}

	value, err := client.Get("key1")
	if err != nil {
		log.Fatalf("Failed to get key1: %v", err)
	}
	fmt.Printf("key1: %s\n", value)

	// 设置带过期时间的键
	if err := client.Set("key2", "value2", 5*time.Second); err != nil {
		log.Fatalf("Failed to set key2: %v", err)
	}

	ttl, err := client.TTL("key2")
	if err != nil {
		log.Fatalf("Failed to get TTL of key2: %v", err)
	}
	fmt.Printf("TTL of key2: %v\n", ttl)

	// 自增操作
	if err := client.Set("counter", "10", 0); err != nil {
		log.Fatalf("Failed to set counter: %v", err)
	}

	newValue, err := client.Incr("counter")
	if err != nil {
		log.Fatalf("Failed to increment counter: %v", err)
	}
	fmt.Printf("Incremented counter: %d\n", newValue)

	// 删除操作
	deleted, err := client.Del("key1", "key2")
	if err != nil {
		log.Fatalf("Failed to delete keys: %v", err)
	}
	fmt.Printf("Deleted %d key(s)\n", deleted)

	// 批量操作示例
	if err := client.MSet(map[string]interface{}{
		"batch_key1": "batch_value1",
		"batch_key2": "batch_value2",
	}); err != nil {
		log.Fatalf("Failed to set batch keys: %v", err)
	}

	values, err := client.MGet("batch_key1", "batch_key2", "nonexistent_key")
	if err != nil {
		log.Fatalf("Failed to get batch keys: %v", err)
	}
	fmt.Printf("Batch get results: %v\n", values)
}
