package storage

import (
	"literedis/internal/consts"
	"literedis/internal/datastruct/dsstring"
	"sync"
	"time"
)

type StringData struct {
	Value  *dsstring.SDS
	Expiry time.Time
}

type MemoryStringStorage struct {
	data map[string]*StringData
	mu   sync.RWMutex
}

func NewMemoryStringStorage() *MemoryStringStorage {
	return &MemoryStringStorage{
		data: make(map[string]*StringData),
	}
}

func (m *MemoryStringStorage) Set(key string, value []byte, expiration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var expiry time.Time
	if expiration > 0 {
		expiry = time.Now().Add(expiration)
	}

	m.data[key] = &StringData{
		Value:  dsstring.NewSDS(string(value)),
		Expiry: expiry,
	}

	return nil
}

func (m *MemoryStringStorage) Get(key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, ok := m.data[key]
	if !ok {
		return nil, consts.ErrKeyNotFound
	}

	if !data.Expiry.IsZero() && time.Now().After(data.Expiry) {
		delete(m.data, key)
		return nil, consts.ErrKeyNotFound
	}

	return data.Value.Get(), nil
}

func (m *MemoryStringStorage) Del(key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, exists := m.data[key]
	delete(m.data, key)
	return exists, nil
}

func (m *MemoryStringStorage) Exists(key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.data[key]
	return exists, nil
}

func (m *MemoryStringStorage) Expire(key string, expiration time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, exists := m.data[key]
	if !exists {
		return false, nil
	}

	data.Expiry = time.Now().Add(expiration)
	return true, nil
}

func (m *MemoryStringStorage) TTL(key string) (time.Duration, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, exists := m.data[key]
	if !exists {
		return -2, nil // -2 indicates that the key does not exist
	}

	if data.Expiry.IsZero() {
		return -1, nil // -1 indicates that the key has no expiration time
	}

	ttl := time.Until(data.Expiry)
	if ttl < 0 {
		delete(m.data, key)
		return -2, nil
	}

	return ttl, nil
}

func (m *MemoryStringStorage) Append(key string, value []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, exists := m.data[key]
	if !exists {
		m.data[key] = &StringData{
			Value: dsstring.NewSDS(""),
		}
		data = m.data[key]
	}

	return data.Value.Append(value), nil
}

func (m *MemoryStringStorage) GetRange(key string, start, end int) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, exists := m.data[key]
	if !exists {
		return nil, consts.ErrKeyNotFound
	}

	if !data.Expiry.IsZero() && time.Now().After(data.Expiry) {
		delete(m.data, key)
		return nil, consts.ErrKeyNotFound
	}

	return data.Value.GetRange(start, end), nil
}

func (m *MemoryStringStorage) SetRange(key string, offset int, value []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, exists := m.data[key]
	if !exists {
		m.data[key] = &StringData{
			Value: dsstring.NewSDS(""),
		}
		data = m.data[key]
	}

	return data.Value.SetRange(offset, value), nil
}

func (m *MemoryStringStorage) StrLen(key string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, exists := m.data[key]
	if !exists {
		return 0, consts.ErrKeyNotFound
	}

	if !data.Expiry.IsZero() && time.Now().After(data.Expiry) {
		delete(m.data, key)
		return 0, consts.ErrKeyNotFound
	}

	return int(data.Value.Len()), nil
}
