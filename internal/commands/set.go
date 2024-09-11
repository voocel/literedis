package commands

import (
	"errors"
	"literedis/internal/storage"
	"literedis/pkg/protocol"
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

	added, err := storage.SAdd(key, members...)
	if err != nil {
		return nil, err
	}

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

	removed, err := storage.SRem(key, members...)
	if err != nil {
		return nil, err
	}

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
