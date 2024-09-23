package main

import (
	"fmt"
	"literedis/pkg/client"
	"sync"
)

// MonitorCallback 是用于处理监控数据的回调函数类型
type MonitorCallback func(string)

type Client struct {
	*client.Client
	addr string
}

type ClientPool struct {
	pool chan *Client
	addr string
	size int
	mtx  sync.Mutex
}

func NewClientPool(host string, port int, size int) *ClientPool {
	addr := fmt.Sprintf("%s:%d", host, port)
	return &ClientPool{
		pool: make(chan *Client, size),
		addr: addr,
		size: size,
	}
}

func (p *ClientPool) Get() (*Client, error) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	select {
	case client := <-p.pool:
		return client, nil
	default:
		c, err := client.NewClient(p.addr)
		if err != nil {
			return nil, err
		}
		return &Client{Client: c, addr: p.addr}, nil
	}
}

func (p *ClientPool) Put(c *Client) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	select {
	case p.pool <- c:
	default:
		c.Close()
	}
}

func (c *Client) Addr() string {
	return c.addr
}

// Monitor 方法用于实时监控服务器接收的命令
func (c *Client) Monitor(callback MonitorCallback) error {
	// 发送 MONITOR 命令
	_, err := c.Do("MONITOR")
	if err != nil {
		return err
	}

	// 持续读取并处理监控数据
	for {
		reply, err := c.Do("") // 使用空命令来接收数据
		if err != nil {
			return err
		}

		// 将回复转换为字符串并调用回调函数
		if str, ok := reply.(string); ok {
			callback(str)
		}
	}
}

func (c *Client) Do(commandName string, args ...interface{}) (interface{}, error) {
	return c.Client.Do(commandName, args...)
}
