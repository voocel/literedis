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

func handleHSet(storage storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) < 3 || len(args)%2 == 0 {
		return nil, errors.New("HSET 命令需要至少三个参数，且参数数量必须为奇数")
	}

	key := args[0]
	fieldsCount := (len(args) - 1) / 2
	fields := make(map[string][]byte)

	for i := 0; i < fieldsCount; i++ {
		fieldName := args[2*i+1]
		fieldValue := args[2*i+2]
		fields[fieldName] = []byte(fieldValue)
	}

	count, err := storage.HSet(key, fields)
	if err != nil {
		return nil, err
	}

	return &protocol.Message{Type: "Integer", Content: count}, nil
}

func handleHGet(storage storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 2 {
		return nil, errors.New("HGET 命令需要两个参数")
	}

	key, field := args[0], args[1]
	value, err := storage.HGet(key, field)
	if err != nil {
		if err == storage.ErrKeyNotFound {
			return &protocol.Message{Type: "BulkString", Content: nil}, nil
		}
		return nil, err
	}

	return &protocol.Message{Type: "BulkString", Content: value}, nil
}

func handleHDel(storage storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) < 2 {
		return nil, errors.New("HDEL 命令需要至少两个参数")
	}

	key := args[0]
	fields := args[1:]

	count, err := storage.HDel(key, fields...)
	if err != nil {
		return nil, err
	}

	return &protocol.Message{Type: "Integer", Content: count}, nil
}

func handleHLen(storage storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 1 {
		return nil, errors.New("HLEN 命令需要一个参数")
	}

	key := args[0]
	length, err := storage.HLen(key)
	if err != nil {
		return nil, err
	}

	return &protocol.Message{Type: "Integer", Content: length}, nil
}
