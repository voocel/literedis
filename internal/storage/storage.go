package storage

import (
	"errors"
	"time"
)

var ErrKeyNotFound = errors.New("key not found")
var ErrInvalidDBIndex = errors.New("invalid database index")

type Storage interface {
	StringStorage
	HashStorage
	ListStorage
	SetStorage
	ZSetStorage

	Del(key string) (bool, error)
	Exists(key string) bool
	Expire(key string, expiration time.Duration) (bool, error)
	TTL(key string) (time.Duration, error)
	Type(key string) (string, error)
	Flush() error
	FlushDB() error
	Select(index int) error
}

type StringStorage interface {
	Set(key string, value []byte) error // 修改这里，移除 expiration 参数
	Get(key string) ([]byte, error)
	Append(key string, value []byte) (int, error)
	GetRange(key string, start, end int) ([]byte, error)
	SetRange(key string, offset int, value []byte) (int, error)
	StrLen(key string) (int, error)
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

type SetStorage interface {
	SAdd(key string, members ...string) (int, error)
	SMembers(key string) ([]string, error)
	SRem(key string, members ...string) (int, error)
	SCard(key string) (int, error)
}

type ZSetStorage interface {
	ZAdd(key string, score float64, member string) (int, error)
	ZScore(key, member string) (float64, bool)
	ZRem(key string, member string) (int, error)
	ZRange(key string, start, stop int64) ([]string, error)
	ZCard(key string) (int64, error)
}
