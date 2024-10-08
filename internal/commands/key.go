package commands

import (
	"errors"
	"literedis/internal/storage"
	"literedis/pkg/protocol"
	"strconv"
	"time"
)

func registerKeyCommands() {
	RegisterCommand("KEYS", handleKeys)
	RegisterCommand("DEL", handleDel)
	RegisterCommand("EXISTS", handleExists)
	RegisterCommand("EXPIRE", handleExpire)
	RegisterCommand("TTL", handleTTL)
	RegisterCommand("TYPE", handleType)
}

func handleKeys(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 1 {
		return nil, errors.New("ERR wrong number of arguments for 'keys' command")
	}
	pattern := args[0]
	keys := s.(storage.KeyStorage).Keys(pattern)
	return &protocol.Message{Type: "Array", Content: keys}, nil
}

func handleDel(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) < 1 {
		return nil, errors.New("DEL command requires at least one argument")
	}

	count := 0
	for _, key := range args {
		deleted, err := s.(storage.KeyStorage).Del(key)
		if err != nil {
			return nil, err
		}
		if deleted {
			count++
		}
	}

	return &protocol.Message{Type: "Integer", Content: count}, nil
}

func handleExists(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) < 1 {
		return nil, errors.New("EXISTS command requires at least one argument")
	}

	count := 0
	for _, key := range args {
		exists := s.(storage.KeyStorage).Exists(key)
		if exists {
			count++
		}
	}

	return &protocol.Message{Type: "Integer", Content: count}, nil
}

func handleExpire(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 2 {
		return nil, errors.New("EXPIRE command requires two arguments")
	}

	key := args[0]
	seconds, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, errors.New("invalid expiration time")
	}

	ok, err := s.(storage.KeyStorage).Expire(key, time.Duration(seconds)*time.Second)
	if err != nil {
		return nil, err
	}

	result := 0
	if ok {
		result = 1
	}

	return &protocol.Message{Type: "Integer", Content: result}, nil
}

func handleTTL(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 1 {
		return nil, errors.New("TTL command requires one argument")
	}

	ttl, err := s.(storage.KeyStorage).TTL(args[0])
	if err != nil {
		return nil, err
	}

	return &protocol.Message{Type: "Integer", Content: int64(ttl.Seconds())}, nil
}

func handleType(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 1 {
		return nil, errors.New("TYPE command requires one argument")
	}

	keyType, err := s.(storage.KeyStorage).Type(args[0])
	if err != nil {
		return nil, err
	}

	return &protocol.Message{Type: "SimpleString", Content: keyType}, nil
}
