package storage

import (
	"errors"
	"literedis/config"
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
	KeyStorage
	ServerStorage
}

type RDBStats struct {
	LastSaveTime     time.Time
	LastSaveDuration time.Duration
	TotalSaves       int
	TotalKeysSaved   int
	LastSaveSize     int64
}

// StringStorage 接口定义了字符串类型的操作
type StringStorage interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	Append(key string, value []byte) (int, error)
	GetRange(key string, start, end int) ([]byte, error)
	SetRange(key string, offset int, value []byte) (int, error)
	StrLen(key string) (int, error)
}

// HashStorage 接口定义了哈希类型的操作
type HashStorage interface {
	HSet(key string, fields map[string][]byte) (int, error)
	HGet(key, field string) ([]byte, error)
	HDel(key string, fields ...string) (int, error)
	HLen(key string) (int, error)
}

// ListStorage 接口定义了列表类型的操作
type ListStorage interface {
	LPush(key string, values ...[]byte) (int, error)
	RPush(key string, values ...[]byte) (int, error)
	LPop(key string) ([]byte, error)
	RPop(key string) ([]byte, error)
	LLen(key string) (int, error)
	LRange(key string, start, stop int) ([][]byte, error)
	LIndex(key string, index int64) ([]byte, error)
	LSet(key string, index int64, value []byte) error
}

// SetStorage 接口定义了集合类型的操作
type SetStorage interface {
	SAdd(key string, members ...string) (int, error)
	SMembers(key string) ([]string, error)
	SRem(key string, members ...string) (int, error)
	SCard(key string) (int, error)
}

// ZSetStorage 接口定义了有序集合类型的操作
type ZSetStorage interface {
	ZAdd(key string, score float64, member string) (int, error)
	ZScore(key, member string) (float64, bool)
	ZRem(key string, member string) (int, error)
	ZRange(key string, start, stop int64) ([]string, error)
	ZCard(key string) (int64, error)
}

// KeyStorage 接口定义了通用的键操作
type KeyStorage interface {
	Keys(pattern string) []string
	Del(key string) (bool, error)
	Exists(key string) bool
	Expire(key string, expiration time.Duration) (bool, error)
	TTL(key string) (time.Duration, error)
	Type(key string) (string, error)
}

// ServerStorage 接口定义了服务器级别的操作
type ServerStorage interface {
	Flush() error
	FlushDB() error
	Select(index int) error

	// RDB 相关的方法
	SaveRDB() error
	LoadRDB() error
	GetRDBStats() RDBStats
	SetRDBConfig(config config.RDBConfig)
}
