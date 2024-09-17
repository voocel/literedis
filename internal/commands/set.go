package commands

import (
	"errors"
	"literedis/internal/storage"
	"literedis/pkg/protocol"
	"log"
	"strconv"
	"time"
)

func registerSetCommands() {
	RegisterCommand("SADD", handleSAdd)
	RegisterCommand("SMEMBERS", handleSMembers)
	RegisterCommand("SREM", handleSRem)
	RegisterCommand("SCARD", handleSCard)
}

func handleSAdd(storage storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) < 2 {
		return nil, errors.New("SADD command requires at least two arguments")
	}

	key := args[0]
	members := args[1:]

	// 检查是否所有成员都是整数
	allIntegers := true
	for _, member := range members {
		if _, err := strconv.ParseInt(member, 10, 64); err != nil {
			allIntegers = false
			break
		}
	}
	if allIntegers {
		log.Printf("Set %s: all members are integers, may use IntSet optimization", key)
	}

	startTime := time.Now()
	added, err := storage.SAdd(key, members...)
	duration := time.Since(startTime)

	if err != nil {
		return nil, err
	}

	log.Printf("SADD %s: added %d members, took %v", key, added, duration)

	return &protocol.Message{Type: "Integer", Content: added}, nil
}

func handleSMembers(storage storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 1 {
		return nil, errors.New("SMEMBERS command requires one argument")
	}

	members, err := storage.SMembers(args[0])
	if err != nil {
		return nil, err
	}

	return &protocol.Message{Type: "Array", Content: members}, nil
}

func handleSRem(storage storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) < 2 {
		return nil, errors.New("SREM command requires at least two arguments")
	}

	key := args[0]
	members := args[1:]

	startTime := time.Now()
	removed, err := storage.SRem(key, members...)
	duration := time.Since(startTime)

	if err != nil {
		return nil, err
	}

	log.Printf("SREM %s: removed %d members, took %v", key, removed, duration)

	return &protocol.Message{Type: "Integer", Content: removed}, nil
}

func handleSCard(storage storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 1 {
		return nil, errors.New("SCARD command requires one argument")
	}

	count, err := storage.SCard(args[0])
	if err != nil {
		return nil, err
	}

	return &protocol.Message{Type: "Integer", Content: count}, nil
}
