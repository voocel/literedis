package commands

import (
	"literedis/internal/storage"
	"literedis/pkg/protocol"
)

type CommandHandler func(storage storage.Storage, args []string) (*protocol.Message, error)

type Command struct {
	Name    string
	Handler CommandHandler
}

var CommandList []Command

func RegisterCommand(name string, handler CommandHandler) {
	CommandList = append(CommandList, Command{Name: name, Handler: handler})
}

func init() {
	registerStringCommands()
	registerHashCommands()
	registerListCommands()
	registerSetCommands()
	registerZSetCommands()
	registerGenericCommands()
}
