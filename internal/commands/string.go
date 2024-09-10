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

func handleSet(storage storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) < 2 {
		return nil, errors.New("SET 命令需要至少两个参数")
	}

	key, value := args[0], args[1]
	var expiration time.Duration = 0

	if len(args) > 2 {
		if strings.ToUpper(args[2]) == "EX" && len(args) > 3 {
			seconds, err := strconv.Atoi(args[3])
			if err != nil {
				return nil, fmt.Errorf("无效的过期时间: %v", err)
			}
			expiration = time.Duration(seconds) * time.Second
		} else if strings.ToUpper(args[2]) == "PX" && len(args) > 3 {
			milliseconds, err := strconv.Atoi(args[3])
			if err != nil {
				return nil, fmt.Errorf("无效的过期时间: %v", err)
			}
			expiration = time.Duration(milliseconds) * time.Millisecond
		}
	}

	err := storage.Set(key, []byte(value), expiration)
	if err != nil {
		return nil, err
	}

	return &protocol.Message{Type: "SimpleString", Content: "OK"}, nil
}

func handleGet(storage storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 1 {
		return nil, errors.New("GET 命令需要一个参数")
	}

	value, err := storage.Get(args[0])
	if err != nil {
		if err == storage.ErrKeyNotFound {
			return &protocol.Message{Type: "BulkString", Content: nil}, nil
		}
		return nil, err
	}

	return &protocol.Message{Type: "BulkString", Content: value}, nil
}

func handleDel(storage storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) < 1 {
		return nil, errors.New("DEL 命令需要至少一个参数")
	}

	count := 0
	for _, key := range args {
		deleted, err := storage.Del(key)
		if err != nil {
			return nil, err
		}
		if deleted {
			count++
		}
	}

	return &protocol.Message{Type: "Integer", Content: count}, nil
}

func handleExists(storage storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) < 1 {
		return nil, errors.New("EXISTS 命令需要至少一个参数")
	}

	count := 0
	for _, key := range args {
		exists, err := storage.Exists(key)
		if err != nil {
			return nil, err
		}
		if exists {
			count++
		}
	}

	return &protocol.Message{Type: "Integer", Content: count}, nil
}

func handleExpire(storage storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 2 {
		return nil, errors.New("EXPIRE 命令需要两个参数")
	}

	key := args[0]
	seconds, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, fmt.Errorf("无效的过期时间: %v", err)
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
