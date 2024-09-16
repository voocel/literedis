package main

import (
	"fmt"
	"literedis/pkg/client"
	"log"
	"time"
)

func main() {
	// Connect to the server
	c, err := client.NewClient("localhost:8080")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer c.Close()

	// Set a key
	err = c.Set("mykey", "Hello, LiteRedis!", 10*time.Second)
	if err != nil {
		log.Fatalf("Failed to set key: %v", err)
	}

	// Get the key
	value, err := c.Get("mykey")
	if err != nil {
		log.Fatalf("Failed to get key: %v", err)
	}
	fmt.Printf("Got value: %s\n", value)

	// Delete the key
	deleted, err := c.Del("mykey")
	if err != nil {
		log.Fatalf("Failed to delete key: %v", err)
	}
	fmt.Printf("Deleted %d key(s)\n", deleted)

	// Try to get the deleted key
	value, err = c.Get("mykey")
	if err != nil {
		log.Fatalf("Failed to get key: %v", err)
	}
	if value == "" {
		fmt.Println("Key not found (as expected)")
	}

	// Use the generic Do method
	result, err := c.Do("PING")
	if err != nil {
		log.Fatalf("Failed to execute PING: %v", err)
	}
	fmt.Printf("PING response: %v\n", result)
}
