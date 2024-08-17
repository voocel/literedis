package app

import (
	"bufio"
	"bytes"
	"fmt"
	"literedis/internal/storage"
	"literedis/pkg/log"
	"literedis/pkg/network"
	"literedis/pkg/network/tcp"
	"literedis/pkg/protocol"
	"strconv"
	"strings"
)

type App struct {
	srv      network.Server
	opts     *options
	storage  storage.Storage
	protocol protocol.Protocol
}

func NewApp(opts ...OptionFunc) *App {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}
	return &App{
		opts:     o,
		storage:  storage.NewMapStorage(),
		protocol: protocol.NewRESPProtocol(),
	}
}

func (a *App) Start() {
	srv := tcp.NewServer(":8080")
	srv.OnConnect(a.handleConnect)
	srv.OnDisconnect(a.handleDisconnect)
	srv.OnReceive(a.handleReceive)

	srv.Start()
	a.srv = srv
}

func (a *App) handleConnect(conn network.Conn) {
	log.Debugf("[Gateway] user connect successful: %v", conn.RemoteAddr())
}

func (a *App) handleDisconnect(conn network.Conn, err error) {
	log.Debugf("[Gateway] user connection disconnected: %v, err: %v", conn.RemoteAddr(), err)
}

func (a *App) handleReceive(conn network.Conn, data []byte) {
	msg, err := a.protocol.Unpack(bytes.NewReader(data))
	if err != nil {
		log.Errorf("unpack data to struct failed: %v", err)
		return
	}
	log.Debugf("receive message type:%v, value: %v", msg.Type, msg.Content)
	response, err := a.processCommand(msg)
	if err != nil {
		log.Infof("Error processing command:%v", err)
		return
	}
	fmt.Println(response)
}

func (a *App) processCommand(msg *protocol.Message) (*protocol.Message, error) {
	switch msg.Type {
	case "SimpleString":
		fmt.Println("SimpleString:", msg.Content)
	case "Error":
		fmt.Println("Error:", msg.Content)
	case "Integer":
		fmt.Println("Integer:", msg.Content)
	case "BulkString":
		fmt.Println("BulkString:", msg.Content)
	case "Array":
		fmt.Println("Array:", msg.Content)
		//command := string(msg.Value.([]byte))
		cmd := msg.Content.([]*protocol.Message)
		switch cmd[0].Content.(string) {
		case "set":
			key, err := readBulkArgs(cmd[1].Content.(*bufio.Reader))
			if err != nil {
				return &protocol.Message{Type: "Error", Content: "Error reading key"}, err
			}
			value, err := readBulkArgs(cmd[2].Content.(*bufio.Reader))
			if err != nil {
				return &protocol.Message{Type: "Error", Content: "Error reading value"}, err
			}
			a.storage.Set(key, value)
			return &protocol.Message{Type: "SimpleString", Content: "OK"}, err
		case "get":
			key, err := readBulkArgs(cmd[1].Content.(*bufio.Reader))
			if err != nil {
				return &protocol.Message{Type: "Error", Content: "Error reading key"}, err
			}
			value, ok := a.storage.Get(key)
			if !ok {
				return &protocol.Message{Type: "BulkString", Content: nil}, err
			}
			return &protocol.Message{Type: "BulkString", Content: value}, err
		case "del":
			key, err := readBulkArgs(cmd[1].Content.(*bufio.Reader))
			if err != nil {
				return &protocol.Message{Type: "Error", Content: "Error reading key"}, err
			}
			a.storage.Del(key)
			return &protocol.Message{Type: "Integer", Content: 0}, err
		default:
			return &protocol.Message{Type: "Error", Content: "Unknown command"}, nil
		}
	default:
		return &protocol.Message{Type: "error", Content: "Invalid message type"}, nil
	}
	return &protocol.Message{Type: "error", Content: "Invalid message type"}, nil
}

// readBulkArgs reads bulk string arguments from the reader.
func readBulkArgs(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	length, err := strconv.Atoi(strings.TrimSuffix(line, protocol.CRLF))
	if err != nil {
		return "", err
	}
	if length == -1 {
		return "", nil // Nil indicates a nil bulk string
	}
	data := make([]byte, length+2)
	_, err = reader.Read(data)
	if err != nil {
		return "", err
	}
	return string(data[:length]), nil
}

func (a *App) Stop() {
	a.srv.Stop()
}
