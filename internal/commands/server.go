package commands

import (
	"errors"
	"literedis/internal/storage"
	"literedis/pkg/protocol"
	"strconv"
)

func registerServerCommands() {
	RegisterCommand("FLUSHALL", handleFlushAll)
	RegisterCommand("FLUSHDB", handleFlushDB)
	RegisterCommand("SELECT", handleSelect)
}

func handleFlushAll(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 0 {
		return nil, errors.New("FLUSHALL command takes no arguments")
	}

	err := s.(storage.ServerStorage).Flush()
	if err != nil {
		return nil, err
	}

	return &protocol.Message{Type: "SimpleString", Content: "OK"}, nil
}

func handleFlushDB(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 0 {
		return nil, errors.New("FLUSHDB command does not require parameters")
	}

	err := s.(storage.ServerStorage).FlushDB()
	if err != nil {
		return nil, err
	}

	return &protocol.Message{Type: "SimpleString", Content: "OK"}, nil
}

func handleSelect(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 1 {
		return nil, errors.New("SELECT command requires a parameter")
	}

	index, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("invalid database index")
	}

	err = s.(storage.ServerStorage).Select(index)
	if err != nil {
		return nil, err
	}

	return &protocol.Message{Type: "SimpleString", Content: "OK"}, nil
}
