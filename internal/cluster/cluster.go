package cluster

import (
	"errors"
	"literedis/internal/communication"
	"literedis/pkg/consistenthash"
	"literedis/pkg/protocol"
	"sync"
)

type Node struct {
	ID      string
	Address string
}

type Cluster struct {
	nodes       map[string]*Node
	hash        *consistenthash.Map
	mu          sync.RWMutex
	localNodeID string
	comm        *communication.NodeCommunicator
}

func NewCluster(localNodeID string) *Cluster {
	return &Cluster{
		nodes:       make(map[string]*Node),
		hash:        consistenthash.New(3, nil),
		localNodeID: localNodeID,
		comm:        communication.NewNodeCommunicator(),
	}
}

func (c *Cluster) AddNode(node *Node) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.nodes[node.ID]; exists {
		return errors.New("node already exists")
	}

	c.nodes[node.ID] = node
	c.hash.Add(node.ID)

	if node.ID != c.localNodeID {
		err := c.comm.Connect(node.ID, node.Address)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Cluster) RemoveNode(nodeID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.nodes[nodeID]; !exists {
		return errors.New("node not found")
	}

	delete(c.nodes, nodeID)
	c.hash.Remove(nodeID)

	if nodeID != c.localNodeID {
		err := c.comm.Close(nodeID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Cluster) GetNodeForKey(key string) *Node {
	c.mu.RLock()
	defer c.mu.RUnlock()
	nodeID := c.hash.Get(key)
	return c.nodes[nodeID]
}

func (c *Cluster) IsLocalNode(nodeID string) bool {
	return nodeID == c.localNodeID
}

func (c *Cluster) GetNodes() []*Node {
	c.mu.RLock()
	defer c.mu.RUnlock()

	nodes := make([]*Node, 0, len(c.nodes))
	for _, node := range c.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

func (c *Cluster) ForwardRequest(nodeID string, msg *protocol.Message) (*protocol.Message, error) {
	return c.comm.SendMessage(nodeID, msg)
}
