package protocol

import (
	"fmt"
	"net"
	"testing"
)

// RedisClient represents a client to a Redis server
type RedisClient struct {
	conn     net.Conn
	protocol Protocol
}

// NewRedisClient creates a new RedisClient
func NewRedisClient(address string) (*RedisClient, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	return &RedisClient{
		conn:     conn,
		protocol: &RESPProtocol{},
	}, nil
}

// Close closes the connection to the Redis server
func (c *RedisClient) Close() error {
	return c.conn.Close()
}

// SendCommand sends a command to the Redis server and returns the response
func (c *RedisClient) SendCommand(cmd *Message) (*Message, error) {
	packedCmd, err := c.protocol.Pack(cmd)
	if err != nil {
		return nil, err
	}

	_, err = c.conn.Write(packedCmd)
	if err != nil {
		return nil, err
	}

	return c.protocol.Unpack(c.conn)
}

// Create a new Message for a Redis command
func NewCommand(command string, args ...string) *Message {
	elements := make([]*Message, len(args)+1)
	elements[0] = &Message{Type: "BulkString", Content: []byte(command)}
	for i, arg := range args {
		elements[i+1] = &Message{Type: "BulkString", Content: []byte(arg)}
	}
	return &Message{Type: "Array", Content: elements}
}

func TestResp(t *testing.T) {
	client, err := NewRedisClient("localhost:6379")
	if err != nil {
		fmt.Println("Error connecting to Redis server:", err)
		return
	}
	defer client.Close()

	// Example SET command
	setCmd := NewCommand("SET", "mykey", "myvalue")
	setResp, err := client.SendCommand(setCmd)
	if err != nil {
		fmt.Println("Error sending SET command:", err)
		return
	}
	fmt.Printf("SET response: %+v\n", setResp)

	// Example GET command
	getCmd := NewCommand("GET", "mykey")
	getResp, err := client.SendCommand(getCmd)
	if err != nil {
		fmt.Println("Error sending GET command:", err)
		return
	}
	fmt.Printf("GET response: %+v\n", getResp)

	// Example HSET command
	hsetCmd := NewCommand("HSET", "myhash", "field1", "value1")
	hsetResp, err := client.SendCommand(hsetCmd)
	if err != nil {
		fmt.Println("Error sending HSET command:", err)
		return
	}
	fmt.Printf("HSET response: %+v\n", hsetResp)

	// Example HGET command
	hgetCmd := NewCommand("HGET", "myhash", "field1")
	hgetResp, err := client.SendCommand(hgetCmd)
	if err != nil {
		fmt.Println("Error sending HGET command:", err)
		return
	}
	fmt.Printf("HGET response: %+v\n", hgetResp)
}
