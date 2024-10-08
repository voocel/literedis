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
	RegisterCommand("APPEND", handleAppend)
	RegisterCommand("GETRANGE", handleGetRange)
	RegisterCommand("SETRANGE", handleSetRange)
}

func handleSet(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) < 2 {
		return nil, errors.New("SET command requires at least two arguments")
	}

	key, value := args[0], []byte(args[1])
	var expiration time.Duration = 0

	if len(args) > 2 {
		if strings.ToUpper(args[2]) == "EX" && len(args) > 3 {
			seconds, err := strconv.Atoi(args[3])
			if err != nil {
				return nil, fmt.Errorf("invalid expiration time: %v", err)
			}
			expiration = time.Duration(seconds) * time.Second
		} else if strings.ToUpper(args[2]) == "PX" && len(args) > 3 {
			milliseconds, err := strconv.Atoi(args[3])
			if err != nil {
				return nil, fmt.Errorf("invalid expiration time: %v", err)
			}
			expiration = time.Duration(milliseconds) * time.Millisecond
		}
	}

	err := s.Set(key, value)
	if err != nil {
		return nil, err
	}

	if expiration > 0 {
		_, err = s.Expire(key, expiration)
		if err != nil {
			return nil, err
		}
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

func handleAppend(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 2 {
		return nil, errors.New("APPEND command requires two arguments")
	}
	key, value := args[0], []byte(args[1])
	newLength, err := s.Append(key, value)
	if err != nil {
		return nil, err
	}
	return &protocol.Message{Type: "Integer", Content: int64(newLength)}, nil
}

func handleGetRange(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 3 {
		return nil, errors.New("GETRANGE command requires three arguments")
	}
	key := args[0]
	start, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, fmt.Errorf("invalid start index: %v", err)
	}
	end, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, fmt.Errorf("invalid end index: %v", err)
	}
	value, err := s.GetRange(key, start, end)
	if err != nil {
		return nil, err
	}
	return &protocol.Message{Type: "BulkString", Content: value}, nil
}

func handleSetRange(s storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 3 {
		return nil, errors.New("SETRANGE command requires three arguments")
	}
	key := args[0]
	offset, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, fmt.Errorf("invalid offset: %v", err)
	}
	value := []byte(args[2])
	newLength, err := s.SetRange(key, offset, value)
	if err != nil {
		return nil, err
	}
	return &protocol.Message{Type: "Integer", Content: int64(newLength)}, nil
}
