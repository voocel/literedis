package app

import (
	"literedis/pkg/network"
	"strings"
)

const defaultName = "literedis"

type options struct {
	id           string
	name         string
	nodeID       string
	server       network.Server
	clusterNodes []string
	clusterMode  bool
}

type OptionFunc func(o *options)

func defaultOptions() *options {
	return &options{
		name: defaultName,
	}
}

func WithID(id string) OptionFunc {
	return func(o *options) { o.id = id }
}

func WithName(name string) OptionFunc {
	return func(o *options) { o.name = name }
}

func WithServer(s network.Server) OptionFunc {
	return func(o *options) { o.server = s }
}

func WithNodeID(id string) OptionFunc {
	return func(o *options) {
		o.nodeID = id
		o.clusterMode = true
	}
}

func WithClusterNodes(nodes string) OptionFunc {
	return func(o *options) {
		o.clusterNodes = strings.Split(nodes, ",")
		o.clusterMode = true
	}
}

func WithClusterMode(enable bool) OptionFunc {
	return func(o *options) { o.clusterMode = enable }
}
