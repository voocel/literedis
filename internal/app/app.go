package app

import (
	"bufio"
	"bytes"
	"fmt"
	"literedis/pkg/log"
	"literedis/pkg/network"
	"literedis/pkg/network/tcp"
	"literedis/pkg/protocol"
	"strings"
)

type App struct {
	srv      network.Server
	opts     *options
	protocol protocol.Protocol
}

func NewApp(opts ...OptionFunc) *App {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}
	return &App{
		opts:     o,
		protocol: protocol.NewRespProtocol(),
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
	log.Debugf("[Gateway] receive message route: %v(%v), cid: %v, uid: %v,data: %v",
		msg.GetCmd(), conn.Cid(), conn.Uid(), string(msg.GetData()))
	response, err := a.processCommand(msg)
	if err != nil {
		log.Println("Error processing command:", err)
		return
	}
}

func (a *App) processCommand(message *protocol.Message) (*protocol.Message, error) {
	switch message.Type {
	case "bulk":
		command := string(message.Value.([]byte))
		switch command {
		case "PING":
			return &protocol.Message{Type: "bulk", Value: []byte("PONG")}, nil
		case "SET":
			// SET 命令格式：SET <key> <value>
			args, err := readBulkArgs(reader)
			if err != nil || len(args) != 2 {
				return &protocol.Message{Type: "error", Value: "Invalid number of arguments for SET"}, nil
			}
			key, value := args[0], args[1]
			set(key, value)
			return &protocol.Message{Type: "bulk", Value: []byte("OK")}, nil
		case "GET":
			// GET 命令格式：GET <key>
			args, err := readBulkArgs(reader)
			if err != nil || len(args) != 1 {
				return &protocol.Message{Type: "error", Value: "Invalid number of arguments for GET"}, nil
			}
			key := args[0]
			value := get(key)
			return &protocol.Message{Type: "bulk", Value: value}, nil
		default:
			return &protocol.Message{Type: "error", Value: "Unsupported command"}, nil
		}
	default:
		return &protocol.Message{Type: "error", Value: "Invalid message type"}, nil
	}
}

func readBulkArgs(reader *bufio.Reader) ([][]byte, error) {
	var args [][]byte
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimRight(line, "\r\n")

		message, err := Unpack(strings.NewReader(line))
		if err != nil {
			return nil, err
		}

		if message.Type != "bulk" {
			return nil, fmt.Errorf("expected bulk data, got %s", message.Type)
		}

		args = append(args, message.Value.([]byte))

		// 假设每个命令只有一个参数
		break
	}

	return args, nil
}

func (a *App) Stop() {
	a.srv.Stop()
}
