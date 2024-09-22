package app

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"literedis/config"
	"literedis/internal/cluster"
	"literedis/internal/commands"
	"literedis/internal/consts"
	"literedis/internal/storage"
	"literedis/pkg/log"
	"literedis/pkg/network"
	"literedis/pkg/network/tcp"
	"literedis/pkg/protocol"
	"strconv"
	"strings"
	"time"
)

type App struct {
	srv           network.Server
	opts          *options
	storage       storage.Storage
	protocol      protocol.Protocol
	cluster       *cluster.Cluster
	handlers      map[string]commands.CommandHandler
	rdbSaveTicker *time.Ticker
	rdbConfig     storage.RDBConfig
}

func NewApp(opts ...OptionFunc) *App {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	app := &App{
		// ... 其他字段的初始化 ...
		opts: options,
	}

	// 加载配置
	config.LoadConfig()

	rdbConfig := config.GetRDBConfig()
	app.storage = storage.NewMemoryStorage()
	app.storage.SetRDBConfig(rdbConfig)

	if err := app.storage.LoadRDB(); err != nil {
		log.Errorf("Failed to load RDB: %v", err)
	}

	app.startRDBSaver()

	return app
}

func (a *App) startRDBSaver() {
	ticker := time.NewTicker(a.opts.rdbConfig.SaveInterval)
	go func() {
		for range ticker.C {
			if err := a.storage.SaveRDB(); err != nil {
				log.Errorf("Failed to start background RDB save: %v", err)
			}
		}
	}()
}

// Implement handleClusterOps method
func (a *App) handleClusterOps(conn network.Conn, msg *protocol.Message) {
	cmdArray, ok := msg.Content.([]*protocol.Message)
	if !ok || len(cmdArray) < 2 {
		a.sendError(conn, errors.New("invalid cluster command"))
		return
	}

	args := make([]string, len(cmdArray)-1)
	for i, arg := range cmdArray[1:] {
		args[i] = string(arg.Content.([]byte))
	}

	response, err := commands.ClusterCommand(a.storage, args)
	if err != nil {
		a.sendError(conn, err)
		return
	}

	respData, err := a.protocol.Pack(response)
	if err != nil {
		a.sendError(conn, err)
		return
	}

	conn.Send(respData)
}

// Helper method to send errors
func (a *App) sendError(conn network.Conn, err error) {
	errResp := &protocol.Message{Type: "Error", Content: []byte(err.Error())}
	respData, _ := a.protocol.Pack(errResp)
	conn.Send(respData)
}

func (a *App) forwardRequest(msg *protocol.Message) (*protocol.Message, error) {
	cmdArray, ok := msg.Content.([]*protocol.Message)
	if !ok || len(cmdArray) < 2 {
		return nil, errors.New("invalid command")
	}

	key := string(cmdArray[1].Content.([]byte))
	node := a.cluster.GetNodeForKey(key)
	if node == nil {
		return nil, errors.New("no node found for key")
	}

	return a.cluster.ForwardRequest(node.ID, msg)
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
		a.sendErrorResponse(conn, "ERR unpack")
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

		// Check if the command should be executed on this node
		if len(args) > 0 {
			node := a.cluster.GetNodeForKey(args[0])
			if !a.cluster.IsLocalNode(node.ID) {
				return nil, consts.ErrWrongNode
			}
		}

		cmd, ok := a.handlers[cmdName]
		if !ok {
			return nil, fmt.Errorf("unknown command: %s", cmdName)
		}
		return cmd(a.storage, args)
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
	a.rdbSaveTicker.Stop()
	if err := a.storage.SaveRDB(); err != nil {
		log.Errorf("Failed to save final RDB: %v", err)
	}
}

func (a *App) sendErrorResponse(conn network.Conn, errMsg string) {
	errResp := &protocol.Message{Type: "Error", Content: []byte(errMsg)}
	respData, _ := a.protocol.Pack(errResp)
	conn.Send(respData)
}

// 添加一个方法来获取RDB统计信息
func (a *App) GetRDBStats() storage.RDBStats {
	return a.storage.GetRDBStats()
}
