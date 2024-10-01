package commands

import (
	"errors"
	"literedis/internal/consts"
	"literedis/internal/storage"
	"literedis/pkg/protocol"
	"strconv"
)

func registerListCommands() {
	RegisterCommand("LPUSH", handleLPush)
	RegisterCommand("RPUSH", handleRPush)
	RegisterCommand("LPOP", handleLPop)
	RegisterCommand("RPOP", handleRPop)
	RegisterCommand("LLEN", handleLLen)
	RegisterCommand("LRANGE", handleLRange)
}

func handleLPush(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) < 2 {
		return nil, consts.ErrInvalidArgument
	}

	key := args[0]
	values := make([][]byte, len(args)-1)
	for i, v := range args[1:] {
		values[i] = []byte(v)
	}

	length, err := s.LPush(key, values...)
	if err != nil {
		return nil, err
	}

	return &protocol.Message{Type: "Integer", Content: length}, nil
}

func handleRPush(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) < 2 {
		return nil, errors.New("RPUSH command requires at least two arguments")
	}

	key := args[0]
	values := make([][]byte, len(args)-1)
	for i, v := range args[1:] {
		values[i] = []byte(v)
	}

	length, err := s.RPush(key, values...)
	if err != nil {
		return nil, err
	}

	return &protocol.Message{Type: "Integer", Content: length}, nil
}

func handleLPop(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 1 {
		return nil, errors.New("LPOP command requires one argument")
	}

	value, err := s.LPop(args[0])
	if err != nil {
		if err == storage.ErrKeyNotFound {
			return &protocol.Message{Type: "BulkString", Content: nil}, nil
		}
		return nil, err
	}

	return &protocol.Message{Type: "BulkString", Content: value}, nil
}

func handleRPop(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 1 {
		return nil, errors.New("RPOP command requires one argument")
	}

	value, err := s.RPop(args[0])
	if err != nil {
		if err == storage.ErrKeyNotFound {
			return &protocol.Message{Type: "BulkString", Content: nil}, nil
		}
		return nil, err
	}

	return &protocol.Message{Type: "BulkString", Content: value}, nil
}

func handleLLen(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 1 {
		return nil, errors.New("LLEN command requires one argument")
	}

	length, err := s.LLen(args[0])
	if err != nil {
		return nil, err
	}

	return &protocol.Message{Type: "Integer", Content: length}, nil
}

func handleLRange(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 3 {
		return nil, consts.ErrInvalidArgument
	}

	key := args[0]
	start, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, consts.ErrNotInteger
	}
	stop, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, consts.ErrNotInteger
	}

	values, err := s.LRange(key, start, stop)
	if err != nil {
		if err == consts.ErrKeyNotFound {
			return &protocol.Message{Type: "Array", Content: [][]byte{}}, nil
		}
		return nil, err
	}

	return &protocol.Message{Type: "Array", Content: values}, nil
}
