package client

import (
	"bufio"
	"errors"
	"fmt"
	"literedis/pkg/protocol"
	"net"
	"time"
)

type Client struct {
	conn     net.Conn
	reader   *bufio.Reader
	protocol protocol.Protocol
}

func NewClient(address string) (*Client, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:     conn,
		reader:   bufio.NewReader(conn),
		protocol: protocol.NewRESPProtocol(),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Do(cmd string, args ...interface{}) (interface{}, error) {
	// Construct the command
	cmdArgs := make([]*protocol.Message, len(args)+1)
	cmdArgs[0] = &protocol.Message{Type: "BulkString", Content: []byte(cmd)}
	for i, arg := range args {
		cmdArgs[i+1] = &protocol.Message{Type: "BulkString", Content: []byte(fmt.Sprintf("%v", arg))}
	}
	message := &protocol.Message{Type: "Array", Content: cmdArgs}

	// Send the command
	data, err := c.protocol.Pack(message)
	if err != nil {
		return nil, err
	}
	_, err = c.conn.Write(data)
	if err != nil {
		return nil, err
	}

	// Read the response
	resp, err := c.protocol.Unpack(c.reader)
	if err != nil {
		return nil, err
	}

	return c.parseResponse(resp)
}

func (c *Client) parseResponse(resp *protocol.Message) (interface{}, error) {
	switch resp.Type {
	case "SimpleString", "BulkString":
		return string(resp.Content.([]byte)), nil
	case "Integer":
		return resp.Content.(int64), nil
	case "Array":
		array := resp.Content.([]*protocol.Message)
		result := make([]interface{}, len(array))
		for i, item := range array {
			parsed, err := c.parseResponse(item)
			if err != nil {
				return nil, err
			}
			result[i] = parsed
		}
		return result, nil
	case "Error":
		return nil, errors.New(string(resp.Content.([]byte)))
	default:
		return nil, fmt.Errorf("unknown response type: %s", resp.Type)
	}
}

func (c *Client) Set(key string, value interface{}, expiration time.Duration) error {
	var cmd string
	var args []interface{}

	if expiration > 0 {
		cmd = "SETEX"
		args = []interface{}{key, int(expiration.Seconds()), value}
	} else {
		cmd = "SET"
		args = []interface{}{key, value}
	}

	_, err := c.Do(cmd, args...)
	return err
}

func (c *Client) Get(key string) (string, error) {
	result, err := c.Do("GET", key)
	if err != nil {
		return "", err
	}
	if result == nil {
		return "", nil
	}
	return result.(string), nil
}

func (c *Client) Del(keys ...string) (int64, error) {
	args := make([]interface{}, len(keys))
	for i, key := range keys {
		args[i] = key
	}
	result, err := c.Do("DEL", args...)
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}
