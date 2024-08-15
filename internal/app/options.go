package app

import "literedis/pkg/network"

const defaultName = "literedis"

type options struct {
	id     string
	name   string
	server network.Server
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
