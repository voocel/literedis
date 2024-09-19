package literedis

import (
	"fmt"
	"time"
)

func (c *Client) Set(key string, value interface{}, expiration time.Duration) error {
	var args []interface{}
	args = append(args, key, value)
	if expiration > 0 {
		args = append(args, "PX", int64(expiration/time.Millisecond))
	}
	_, err := c.Do("SET", args...)
	if err != nil {
		return fmt.Errorf("SET command failed: %w", err)
	}
	return nil
}

func (c *Client) Get(key string) (string, error) {
	reply, err := c.Do("GET", key)
	if err != nil {
		return "", fmt.Errorf("GET command failed: %w", err)
	}
	if reply == nil {
		return "", nil
	}
	return reply.(string), nil
}

func (c *Client) Del(keys ...string) (int64, error) {
	args := make([]interface{}, len(keys))
	for i, key := range keys {
		args[i] = key
	}
	reply, err := c.Do("DEL", args...)
	if err != nil {
		return 0, fmt.Errorf("DEL command failed: %w", err)
	}
	return reply.(int64), nil
}

func (c *Client) Incr(key string) (int64, error) {
	reply, err := c.Do("INCR", key)
	if err != nil {
		return 0, fmt.Errorf("INCR command failed: %w", err)
	}
	return reply.(int64), nil
}

func (c *Client) Expire(key string, expiration time.Duration) (bool, error) {
	reply, err := c.Do("EXPIRE", key, int64(expiration/time.Second))
	if err != nil {
		return false, fmt.Errorf("EXPIRE command failed: %w", err)
	}
	return reply.(int64) == 1, nil
}

func (c *Client) TTL(key string) (time.Duration, error) {
	reply, err := c.Do("TTL", key)
	if err != nil {
		return 0, fmt.Errorf("TTL command failed: %w", err)
	}
	ttl := reply.(int64)
	if ttl < 0 {
		return -1, nil
	}
	return time.Duration(ttl) * time.Second, nil
}

func (c *Client) MSet(pairs map[string]interface{}) error {
	args := make([]interface{}, 0, len(pairs)*2)
	for k, v := range pairs {
		args = append(args, k, v)
	}
	_, err := c.Do("MSET", args...)
	if err != nil {
		return fmt.Errorf("MSET command failed: %w", err)
	}
	return nil
}

func (c *Client) MGet(keys ...string) ([]interface{}, error) {
	args := make([]interface{}, len(keys))
	for i, key := range keys {
		args[i] = key
	}
	reply, err := c.Do("MGET", args...)
	if err != nil {
		return nil, fmt.Errorf("MGET command failed: %w", err)
	}
	return reply.([]interface{}), nil
}
