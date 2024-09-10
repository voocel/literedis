package commands

import (
	"errors"
	"literedis/internal/storage"
	"literedis/pkg/protocol"
)

func registerHashCommands() {
	RegisterCommand("HSET", handleHSet)
	RegisterCommand("HGET", handleHGet)
	RegisterCommand("HDEL", handleHDel)
	RegisterCommand("HLEN", handleHLen)
}

func handleHSet(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) < 3 || len(args)%2 == 0 {
		return nil, errors.New("HSET command requires at least three arguments and an odd number of arguments")
	}

	key := args[0]
	fieldsCount := (len(args) - 1) / 2
	fields := make(map[string][]byte)

	for i := 0; i < fieldsCount; i++ {
		fieldName := args[2*i+1]
		fieldValue := args[2*i+2]
		fields[fieldName] = []byte(fieldValue)
	}

	count, err := s.HSet(key, fields)
	if err != nil {
		return nil, err
	}

	return &protocol.Message{Type: "Integer", Content: count}, nil
}

func handleHGet(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 2 {
		return nil, errors.New("HGET command requires two parameters")
	}

	key, field := args[0], args[1]
	value, err := s.HGet(key, field)
	if err != nil {
		if err == storage.ErrKeyNotFound {
			return &protocol.Message{Type: "BulkString", Content: nil}, nil
		}
		return nil, err
	}

	return &protocol.Message{Type: "BulkString", Content: value}, nil
}

func handleHDel(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) < 2 {
		return nil, errors.New("HDEL command requires at least two parameters")
	}

	key := args[0]
	fields := args[1:]

	count, err := s.HDel(key, fields...)
	if err != nil {
		return nil, err
	}

	return &protocol.Message{Type: "Integer", Content: count}, nil
}

func handleHLen(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 1 {
		return nil, errors.New("HLEN command requires a parameter")
	}

	key := args[0]
	length, err := s.HLen(key)
	if err != nil {
		return nil, err
	}

	return &protocol.Message{Type: "Integer", Content: length}, nil
}
