package storage

import (
	"sync"
	"time"
)

type ListData struct {
	Values [][]byte
	Expiry time.Time
}

type MemoryListStorage struct {
	data map[string]*ListData
	mu   sync.RWMutex
}

func NewMemoryListStorage() *MemoryListStorage {
	return &MemoryListStorage{
		data: make(map[string]*ListData),
	}
}

func (m *MemoryListStorage) LPush(key string, values ...[]byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	list, ok := m.data[key]
	if !ok {
		list = &ListData{Values: make([][]byte, 0)}
		m.data[key] = list
	}

	list.Values = append(append(make([][]byte, 0, len(values)+len(list.Values)), values...), list.Values...)
	return len(list.Values), nil
}

func (m *MemoryListStorage) RPush(key string, values ...[]byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	list, ok := m.data[key]
	if !ok {
		list = &ListData{Values: make([][]byte, 0)}
		m.data[key] = list
	}

	list.Values = append(list.Values, values...)
	return len(list.Values), nil
}

func (m *MemoryListStorage) LPop(key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	list, ok := m.data[key]
	if !ok || len(list.Values) == 0 {
		return nil, ErrKeyNotFound
	}

	value := list.Values[0]
	list.Values = list.Values[1:]

	if len(list.Values) == 0 {
		delete(m.data, key)
	}

	return value, nil
}

func (m *MemoryListStorage) RPop(key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	list, ok := m.data[key]
	if !ok || len(list.Values) == 0 {
		return nil, ErrKeyNotFound
	}

	lastIndex := len(list.Values) - 1
	value := list.Values[lastIndex]
	list.Values = list.Values[:lastIndex]

	if len(list.Values) == 0 {
		delete(m.data, key)
	}

	return value, nil
}

func (m *MemoryListStorage) LLen(key string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list, ok := m.data[key]
	if !ok {
		return 0, nil
	}

	return len(list.Values), nil
}

func (m *MemoryListStorage) LRange(key string, start, stop int) ([][]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list, ok := m.data[key]
	if !ok {
		return [][]byte{}, nil
	}

	length := len(list.Values)
	if start < 0 {
		start = length + start
	}
	if stop < 0 {
		stop = length + stop
	}

	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}

	if start > stop {
		return [][]byte{}, nil
	}

	return list.Values[start : stop+1], nil
}
