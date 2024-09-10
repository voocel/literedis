package tcp

import (
	"literedis/pkg/network"
	"literedis/pkg/protocol"
	"net"
)

type client struct {
	addr              string
	opts              *clientOptions
	protocol          protocol.Protocol
	connectHandler    network.ConnectHandler
	disconnectHandler network.DisconnectHandler
	receiveHandler    network.ReceiveHandler
}

func NewClient(addr string, options ...ClientOptionFunc) network.Client {
	o := defaultClientOptions()
	o.addr = addr
	for _, option := range options {
		option(o)
	}
	return &client{
		opts:     o,
		protocol: protocol.NewRESPProtocol(),
	}
}

func (c *client) Dial() (network.Conn, error) {
	addr, err := net.ResolveTCPAddr("tcp", c.opts.addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		return nil, err
	}

	return newClientConn(c, conn), nil
}

func (c *client) OnConnect(handler network.ConnectHandler) {
	c.connectHandler = handler
}

func (c *client) OnReceive(handler network.ReceiveHandler) {
	c.receiveHandler = handler
}

func (c *client) OnDisconnect(handler network.DisconnectHandler) {
	c.disconnectHandler = handler
}

func (c *client) Codec() protocol.Protocol {
	return c.protocol
}
