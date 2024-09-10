package protocol

import "io"

// Message represents a generic message structure
type Message struct {
	Type    string
	Content interface{}
}

type Protocol interface {
	// Pack packs Message into the packet to be written
	Pack(msg *Message) ([]byte, error)

	// Unpack unpacks the message packet from reader
	Unpack(reader io.Reader) (*Message, error)
}
