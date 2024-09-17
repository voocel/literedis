package commands

import (
	"errors"
	"literedis/internal/storage"
	"literedis/pkg/protocol"
	"strconv"
)

func registerZSetCommands() {
	RegisterCommand("ZADD", handleZAdd)
	RegisterCommand("ZSCORE", handleZScore)
	RegisterCommand("ZREM", handleZRem)
	RegisterCommand("ZRANGE", handleZRange)
	RegisterCommand("ZCARD", handleZCard)
}

func handleZAdd(storage storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) < 3 || len(args)%2 != 1 {
		return nil, errors.New("ZADD command requires at least one score-member pair")
	}

	key := args[0]
	added := 0

	for i := 1; i < len(args); i += 2 {
		score, err := strconv.ParseFloat(args[i], 64)
		if err != nil {
			return nil, errors.New("Invalid score")
		}
		member := args[i+1]

		n, err := storage.ZAdd(key, score, member)
		if err != nil {
			return nil, err
		}
		added += n
	}

	return &protocol.Message{Type: "Integer", Content: added}, nil
}

func handleZScore(storage storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 2 {
		return nil, errors.New("ZSCORE command requires exactly two arguments")
	}

	key := args[0]
	member := args[1]

	score, ok := storage.ZScore(key, member)
	if !ok {
		return &protocol.Message{Type: "Null"}, nil
	}

	return &protocol.Message{Type: "BulkString", Content: strconv.FormatFloat(score, 'f', -1, 64)}, nil
}

func handleZRem(storage storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) < 2 {
		return nil, errors.New("ZREM command requires at least two arguments")
	}

	key := args[0]
	members := args[1:]

	removed := 0
	for _, member := range members {
		n, err := storage.ZRem(key, member)
		if err != nil {
			return nil, err
		}
		removed += n
	}

	return &protocol.Message{Type: "Integer", Content: removed}, nil
}

func handleZRange(storage storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) < 3 {
		return nil, errors.New("ZRANGE command requires at least three arguments")
	}

	key := args[0]
	start, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, errors.New("Invalid start index")
	}
	stop, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return nil, errors.New("Invalid stop index")
	}

	members, err := storage.ZRange(key, start, stop)
	if err != nil {
		return nil, err
	}

	return &protocol.Message{Type: "Array", Content: members}, nil
}

func handleZCard(storage storage.Storage, args []string) (*protocol.Message, error) {
	if len(args) != 1 {
		return nil, errors.New("ZCARD command requires exactly one argument")
	}

	key := args[0]

	count, err := storage.ZCard(key)
	if err != nil {
		return nil, err
	}

	return &protocol.Message{Type: "Integer", Content: count}, nil
}
