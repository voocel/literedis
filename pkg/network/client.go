package network

import "literedis/pkg/protocol"

type Client interface {
	Dial() (Conn, error)

	Codec() (p protocol.Protocol)

	OnConnect(handler ConnectHandler)

	OnReceive(handler ReceiveHandler)

	OnDisconnect(handler DisconnectHandler)
}
