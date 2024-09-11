package communication

import (
	"bytes"
	"errors"
	"literedis/pkg/network"
	"literedis/pkg/network/tcp"
	"literedis/pkg/protocol"
	"sync"
	"time"
)

type NodeCommunicator struct {
	connections map[string]network.Conn
	protocol    protocol.Protocol
	mu          sync.RWMutex
	responseCh  chan *protocol.Message
}

func NewNodeCommunicator() *NodeCommunicator {
	return &NodeCommunicator{
		connections: make(map[string]network.Conn),
		protocol:    protocol.NewRESPProtocol(),
		responseCh:  make(chan *protocol.Message),
	}
}

func (nc *NodeCommunicator) Connect(nodeID, address string) error {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	if _, exists := nc.connections[nodeID]; exists {
		return errors.New("connection already exists")
	}

	c := tcp.NewClient(address)
	conn, err := c.Dial()
	if err != nil {
		return err
	}

	nc.handleIncomingMessages(c)

	nc.connections[nodeID] = conn
	return nil
}

func (nc *NodeCommunicator) SendMessage(nodeID string, msg *protocol.Message) (*protocol.Message, error) {
	nc.mu.RLock()
	conn, exists := nc.connections[nodeID]
	nc.mu.RUnlock()

	if !exists {
		return nil, errors.New("connection not found")
	}

	data, err := nc.protocol.Pack(msg)
	if err != nil {
		return nil, err
	}

	err = conn.Send(data)
	if err != nil {
		return nil, err
	}

	select {
	case resp := <-nc.responseCh:
		return resp, nil
	case <-time.After(5 * time.Second):
		return nil, errors.New("timeout waiting for response")
	}
}

func (nc *NodeCommunicator) handleIncomingMessages(c network.Client) {
	c.OnReceive(func(conn network.Conn, msg []byte) {
		data, err := nc.protocol.Unpack(bytes.NewBuffer(msg))
		if err != nil {
			return
		}
		nc.responseCh <- data
	})
}

func (nc *NodeCommunicator) Close(nodeID string) error {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	conn, exists := nc.connections[nodeID]
	if !exists {
		return errors.New("connection not found")
	}

	err := conn.Close()
	if err != nil {
		return err
	}

	delete(nc.connections, nodeID)
	return nil
}
