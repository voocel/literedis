package literedis

import (
	"errors"
	"sync"
	"time"
)

type Pool struct {
	address     string
	maxSize     int
	maxIdleSize int
	idleTimeout time.Duration
	clients     chan *Client
	mu          sync.Mutex
	activeConn  int
	idleClients []*idleClient
}

type idleClient struct {
	client   *Client
	lastUsed time.Time
}

func NewPool(address string, maxSize int, maxIdleSize int, idleTimeout time.Duration) *Pool {
	p := &Pool{
		address:     address,
		maxSize:     maxSize,
		maxIdleSize: maxIdleSize,
		idleTimeout: idleTimeout,
		clients:     make(chan *Client, maxSize),
	}
	go p.cleanIdleConnections()
	return p
}

func (p *Pool) Get() (*Client, error) {
	select {
	case client := <-p.clients:
		if err := client.Ping(); err != nil {
			client.Close()
			p.activeConn--
			return p.Get() // 递归调用以获取新的连接
		}
		return client, nil
	default:
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.activeConn < p.maxSize {
			client, err := NewClient(p.address)
			if err != nil {
				return nil, err
			}
			p.activeConn++
			return client, nil
		}
		return nil, errors.New("connection pool exhausted")
	}
}

func (p *Pool) Put(client *Client) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.idleClients) < p.maxIdleSize {
		p.idleClients = append(p.idleClients, &idleClient{client: client, lastUsed: time.Now()})
	} else {
		client.Close()
		p.activeConn--
	}
}

func (p *Pool) Close() {
	close(p.clients)
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, ic := range p.idleClients {
		ic.client.Close()
	}
	p.idleClients = nil
	p.activeConn = 0
}

func (p *Pool) cleanIdleConnections() {
	ticker := time.NewTicker(p.idleTimeout / 2)
	defer ticker.Stop()

	for range ticker.C {
		p.mu.Lock()
		now := time.Now()
		remainingClients := make([]*idleClient, 0, len(p.idleClients))
		for _, ic := range p.idleClients {
			if now.Sub(ic.lastUsed) > p.idleTimeout {
				ic.client.Close()
				p.activeConn--
			} else {
				remainingClients = append(remainingClients, ic)
			}
		}
		p.idleClients = remainingClients
		p.mu.Unlock()
	}
}

func (p *Pool) HealthCheck() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i := 0; i < len(p.idleClients); i++ {
		ic := p.idleClients[i]
		err := ic.client.Ping()
		if err != nil {
			ic.client.Close()
			p.activeConn--
			p.idleClients = append(p.idleClients[:i], p.idleClients[i+1:]...)
			i--
		}
	}
}
