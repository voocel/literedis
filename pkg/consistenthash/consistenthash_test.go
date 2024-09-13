package consistenthash

import (
	"fmt"
	"testing"
)

func TestConsistentHash(t *testing.T) {
	hash := New(3, nil)

	// Add some nodes
	hash.Add("10.0.0.1:8080", "10.0.0.2:8080", "10.0.0.3:8080")

	// Get the node for some keys
	testKeys := []string{"key1", "key2", "key3", "key4", "key5"}
	for _, key := range testKeys {
		server := hash.Get(key)
		fmt.Printf("%s is mapped to %s\n", key, server)
	}

	fmt.Println("\nAdding a new node 10.0.0.4:8080")
	hash.Add("10.0.0.4:8080")

	// Check the distribution again
	for _, key := range testKeys {
		server := hash.Get(key)
		fmt.Printf("%s is mapped to %s\n", key, server)
	}

	fmt.Println("\nRemoving node 10.0.0.2:8080")
	hash.Remove("10.0.0.2:8080")

	// Check the distribution once more
	for _, key := range testKeys {
		server := hash.Get(key)
		fmt.Printf("%s is mapped to %s\n", key, server)
	}
}
