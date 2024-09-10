package commands

import (
	"errors"
	"fmt"
	"literedis/internal/storage"
	"literedis/pkg/protocol"
	"strconv"
	"strings"
	"time"
)

func registerStringCommands() {
	RegisterCommand("SET", handleSet)
	RegisterCommand("GET", handleGet)
	RegisterCommand("DEL", handleDel)
	RegisterCommand("EXISTS", handleExists)
	RegisterCommand("EXPIRE", handleExpire)
}

func handleSet(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) < 2 {
		return nil, errors.New("SET command requires at least two arguments")
	}

	key, value := args[0], args[1]
	var expiration time.Duration = 0

	if len(args) > 2 {
		if strings.ToUpper(args[2]) == "EX" && len(args) > 3 {
			seconds, err := strconv.Atoi(args[3])
			if err != nil {
				return nil, fmt.Errorf("Invalid expiration time: %v", err)
			}
			expiration = time.Duration(seconds) * time.Second
		} else if strings.ToUpper(args[2]) == "PX" && len(args) > 3 {
			milliseconds, err := strconv.Atoi(args[3])
			if err != nil {
				return nil, fmt.Errorf("Invalid expiration time: %v", err)
			}
			expiration = time.Duration(milliseconds) * time.Millisecond
		}
	}

	err := s.Set(key, []byte(value), expiration)
	if err != nil {
		return nil, err
	}

	return &protocol.Message{Type: "SimpleString", Content: "OK"}, nil
}

func handleGet(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 1 {
		return nil, errors.New("GET command requires one argument")
	}

	value, err := s.Get(args[0])
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			return &protocol.Message{Type: "BulkString", Content: nil}, nil
		}
		return nil, err
	}

	return &protocol.Message{Type: "BulkString", Content: value}, nil
}

func handleDel(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) < 1 {
		return nil, errors.New("DEL command requires at least one argument")
	}

	count := 0
	for _, key := range args {
		deleted, err := s.Del(key)
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
		if s.Exists(key) {
			count++
		}
	}

	return &protocol.Message{Type: "Integer", Content: count}, nil
}

func handleExpire(storage storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 2 {
		return nil, errors.New("EXPIRE command requires two arguments")
	}

	key := args[0]
	seconds, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, fmt.Errorf("Invalid expiration time: %v", err)
	}

	ok, err := storage.Expire(key, time.Duration(seconds)*time.Second)
	if err != nil {
		return nil, err
	}

	result := 0
	if ok {
		result = 1
	}

	return &protocol.Message{Type: "Integer", Content: result}, nil
}
