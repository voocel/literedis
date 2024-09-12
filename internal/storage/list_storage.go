package storage

import (
	"literedis/internal/consts"
	"literedis/internal/datastruct/dslist"
	"sync"
	"time"
)

// MemoryListStorage implements ListStorage interface using in-memory storage
type MemoryListStorage struct {
	data map[string]*dslist.QuickList
	mu   sync.RWMutex
}

// NewMemoryListStorage creates a new MemoryListStorage
func NewMemoryListStorage() *MemoryListStorage {
	return &MemoryListStorage{
		data: make(map[string]*dslist.QuickList),
	}
}

// LPush inserts all the specified values at the head of the list stored at key
func (m *MemoryListStorage) LPush(key string, values ...[]byte) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	list, ok := m.data[key]
	if !ok {
		list = dslist.New()
		m.data[key] = list
	}

	return list.LPush(values...), nil
}

// RPush inserts all the specified values at the tail of the list stored at key
func (m *MemoryListStorage) RPush(key string, values ...[]byte) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	list, ok := m.data[key]
	if !ok {
		list = dslist.New()
		m.data[key] = list
	}

	return list.RPush(values...), nil
}

// LPop removes and returns the first element of the list stored at key
func (m *MemoryListStorage) LPop(key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	list, ok := m.data[key]
	if !ok {
		return nil, consts.ErrKeyNotFound
	}

	value, ok := list.LPop()
	if !ok {
		return nil, consts.ErrKeyNotFound
	}

	if list.Len() == 0 {
		delete(m.data, key)
	}

	return value, nil
}

// RPop removes and returns the last element of the list stored at key
func (m *MemoryListStorage) RPop(key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	list, ok := m.data[key]
	if !ok {
		return nil, consts.ErrKeyNotFound
	}

	value, ok := list.RPop()
	if !ok {
		return nil, consts.ErrKeyNotFound
	}

	if list.Len() == 0 {
		delete(m.data, key)
	}

	return value, nil
}

// LRange returns the specified elements of the list stored at key
func (m *MemoryListStorage) LRange(key string, start, stop int64) ([][]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list, ok := m.data[key]
	if !ok {
		return nil, consts.ErrKeyNotFound
	}

	return list.LRange(start, stop), nil
}

// LIndex returns the element at index in the list stored at key
func (m *MemoryListStorage) LIndex(key string, index int64) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list, ok := m.data[key]
	if !ok {
		return nil, consts.ErrKeyNotFound
	}

	value, ok := list.LIndex(index)
	if !ok {
		return nil, consts.ErrKeyNotFound
	}

	return value, nil
}

// LSet sets the list element at index to value
func (m *MemoryListStorage) LSet(key string, index int64, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	list, ok := m.data[key]
	if !ok {
		return consts.ErrKeyNotFound
	}

	if !list.LSet(index, value) {
		return consts.ErrIndexOutOfRange
	}

	return nil
}

// LLen returns the length of the list stored at key
func (m *MemoryListStorage) LLen(key string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list, ok := m.data[key]
	if !ok {
		return 0, consts.ErrKeyNotFound
	}

	return list.Len(), nil
}

// Expire sets a timeout on key
func (m *MemoryListStorage) Expire(key string, expiration time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	list, ok := m.data[key]
	if !ok {
		return false, nil
	}

	list.SetExpire(time.Now().Add(expiration))
	return true, nil
}

// TTL returns the remaining time to live of a key that has a timeout
func (m *MemoryListStorage) TTL(key string) (time.Duration, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list, ok := m.data[key]
	if !ok {
		return -2, nil // -2 indicates that the key does not exist
	}

	if list.Expire().IsZero() {
		return -1, nil // -1 indicates that the key has no expiration time
	}

	ttl := time.Until(list.Expire())
	if ttl < 0 {
		delete(m.data, key)
		return -2, nil
	}

	return ttl, nil
}