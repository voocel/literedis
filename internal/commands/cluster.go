package commands

import (
	"fmt"
	"literedis/internal/cluster"
	"literedis/internal/consts"
	"literedis/internal/storage"
	"literedis/pkg/protocol"
	"strings"
)

func init() {
	RegisterCommand("CLUSTER", ClusterCommand)
}

func ClusterCommand(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) < 1 {
		return nil, consts.ErrInvalidArgument
	}

	var clusterInstance *cluster.Cluster
	if ms, ok := s.(*storage.MemoryStorage); ok {
		clusterInstance = ms.GetCluster()
	} else {
		return nil, consts.ErrClusterNotEnabled
	}

	if clusterInstance == nil {
		return nil, consts.ErrClusterNotEnabled
	}

	switch strings.ToUpper(args[0]) {
	case "JOIN":
		if len(args) != 3 {
			return nil, consts.ErrInvalidArgument
		}
		nodeID, address := args[1], args[2]
		err := clusterInstance.AddNode(&cluster.Node{ID: nodeID, Address: address})
		if err != nil {
			return nil, err
		}
		return &protocol.Message{Type: "SimpleString", Content: []byte("OK")}, nil

	case "LEAVE":
		if len(args) != 2 {
			return nil, consts.ErrInvalidArgument
		}
		nodeID := args[1]
		err := clusterInstance.RemoveNode(nodeID)
		if err != nil {
			return nil, err
		}
		return &protocol.Message{Type: "SimpleString", Content: []byte("OK")}, nil

	case "NODES":
		nodes := clusterInstance.GetNodes()
		nodeList := make([]string, 0, len(nodes))
		for _, node := range nodes {
			nodeList = append(nodeList, fmt.Sprintf("%s %s", node.ID, node.Address))
		}
		return &protocol.Message{Type: "Array", Content: nodeList}, nil

	default:
		return nil, consts.ErrUnknownCommand
	}
}
