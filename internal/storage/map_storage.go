package storage

import (
	"errors"
	"sync"
)

var ErrKeyNotFound = errors.New("key not found")

type MapStorage struct {
	data sync.Map
}

func NewMapStorage() *MapStorage {
	return &MapStorage{}
}

func (m *MapStorage) Get(key string) (any, bool) {
	val, ok := m.data.Load(key)
	return val, ok
}

func (m *MapStorage) Set(key string, value interface{}) {
	m.data.Store(key, value)
}

func (m *MapStorage) Del(key string) {
	m.data.Delete(key)
}

func (m *MapStorage) Keys() []string {
	var keys []string
	m.data.Range(func(key, value interface{}) bool {
		keys = append(keys, key.(string))
		return true
	})
	return keys
}

func (m *MapStorage) Size() int {
	length := 0
	m.data.Range(func(key, value any) bool {
		length++
		return true
	})
	return length
}

func (m *MapStorage) Flush() {
	m.data = sync.Map{}
}
