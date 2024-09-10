package app

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"literedis/internal/commands"
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
	handlers map[string]commands.CommandHandler
}

func NewApp(opts ...OptionFunc) *App {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}
	app := &App{
		opts:     o,
		storage:  storage.NewMemoryStorage(),
		protocol: protocol.NewRESPProtocol(),
		handlers: make(map[string]commands.CommandHandler),
	}
	app.registerHandlers()
	return app
}

func (a *App) registerHandlers() {
	for _, cmd := range commands.CommandList {
		a.handlers[cmd.Name] = cmd.Handler
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
		errResp := &protocol.Message{Type: "Error", Content: []byte(err.Error())}
		respData, _ := a.protocol.Pack(errResp)
		conn.Send(respData)
		return
	}
	respData, err := a.protocol.Pack(response)
	if err != nil {
		log.Errorf("pack response failed: %v", err)
		return
	}
	conn.Send(respData)
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
		cmdArray, ok := msg.Content.([]*protocol.Message)
		if !ok || len(cmdArray) == 0 {
			return nil, errors.New("invalid command")
		}
		cmdName := strings.ToUpper(string(cmdArray[0].Content.([]byte)))
		args := make([]string, len(cmdArray)-1)
		for i, arg := range cmdArray[1:] {
			args[i] = string(arg.Content.([]byte))
		}
		cmd, ok := a.handlers[cmdName]
		if !ok {
			return nil, fmt.Errorf("未知命令: %s", cmdName)
		}
		return cmd.Handler(a.storage, args)
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
