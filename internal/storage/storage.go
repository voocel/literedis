package storage

import (
	"errors"
	"time"
)

var ErrKeyNotFound = errors.New("key does not exist")

type Storage interface {
	StringStorage
	HashStorage
	ListStorage

	Del(key string) (bool, error)
	Exists(key string) (bool, error)
	Expire(key string, expiration time.Duration) (bool, error)
	TTL(key string) (time.Duration, error)
	Type(key string) (string, error)
	Flush() error
}

type StringStorage interface {
	Set(key string, value []byte, expiration time.Duration) error
	Get(key string) ([]byte, error)
}

type HashStorage interface {
	HSet(key string, fields map[string][]byte) (int, error)
	HGet(key, field string) ([]byte, error)
	HDel(key string, fields ...string) (int, error)
	HLen(key string) (int, error)
}

type ListStorage interface {
	LPush(key string, values ...[]byte) (int, error)
	RPush(key string, values ...[]byte) (int, error)
	LPop(key string) ([]byte, error)
	RPop(key string) ([]byte, error)
	LLen(key string) (int, error)
	LRange(key string, start, stop int) ([][]byte, error)
}
