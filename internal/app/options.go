package app

import (
	"compress/gzip"
	"literedis/pkg/network"
	"strings"
	"time"
)

const defaultName = "literedis"

type RDBConfig struct {
	Filename         string
	SaveInterval     time.Duration
	CompressionLevel int
	AutoSaveChanges  int
}

type options struct {
	id           string
	name         string
	nodeID       string
	server       network.Server
	clusterNodes []string
	clusterMode  bool
	rdbConfig    RDBConfig
}

type OptionFunc func(o *options)

func defaultOptions() *options {
	return &options{
		name: defaultName,
		rdbConfig: RDBConfig{
			Filename:         "dump.rdb",
			SaveInterval:     5 * time.Minute,
			CompressionLevel: gzip.DefaultCompression,
			AutoSaveChanges:  1000,
		},
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

func WithRDBConfig(config RDBConfig) OptionFunc {
	return func(o *options) { o.rdbConfig = config }
}
