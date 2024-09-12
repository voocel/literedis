package storage

import (
	"literedis/internal/cluster"
	"literedis/internal/consts"
	"literedis/internal/datastruct/dslist"
	"sync"
	"time"
)

const DefaultDBCount = 16

type Database struct {
	stringStorage *MemoryStringStorage
	hashStorage   *MemoryHashStorage
	listStorage   map[string]*dslist.QuickList
	setStorage    *MemorySetStorage
	expiry        map[string]time.Time
}

type MemoryStorage struct {
	databases      []*Database
	currentDBIndex int
	mu             sync.RWMutex
	cluster        *cluster.Cluster
}

func NewMemoryStorage() *MemoryStorage {
	ms := &MemoryStorage{
		databases:      make([]*Database, DefaultDBCount),
		currentDBIndex: 0,
	}
	for i := 0; i < DefaultDBCount; i++ {
		ms.databases[i] = &Database{
			stringStorage: NewMemoryStringStorage(),
			hashStorage:   NewMemoryHashStorage(),
			listStorage:   make(map[string]*dslist.QuickList),
			setStorage:    NewMemorySetStorage(),
			expiry:        make(map[string]time.Time),
		}
	}
	return ms
}

func (m *MemoryStorage) GetCluster() *cluster.Cluster {
	return m.cluster
}

func (m *MemoryStorage) SetCluster(c *cluster.Cluster) {
	m.cluster = c
}

func (m *MemoryStorage) getCurrentDB() *Database {
	return m.databases[m.currentDBIndex]
}

func (m *MemoryStorage) Select(index int) error {
	if index < 0 || index >= len(m.databases) {
		return ErrInvalidDBIndex
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentDBIndex = index
	return nil
}

// ########################## String operations ##########################

func (m *MemoryStorage) Set(key string, value []byte, expiration time.Duration) error {
	db := m.getCurrentDB()
	err := db.stringStorage.Set(key, value, expiration)
	if err == nil && expiration > 0 {
		db.expiry[key] = time.Now().Add(expiration)
	}
	return err
}

func (m *MemoryStorage) Get(key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
		return nil, ErrKeyNotFound
	}
	return db.stringStorage.Get(key)
}

func (m *MemoryStorage) Append(key string, value []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	return db.stringStorage.Append(key, value)
}

func (m *MemoryStorage) GetRange(key string, start, end int) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	return db.stringStorage.GetRange(key, start, end)
}

func (m *MemoryStorage) SetRange(key string, offset int, value []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	return db.stringStorage.SetRange(key, offset, value)
}

func (m *MemoryStorage) StrLen(key string) (int, error) {
	//TODO implement me
	panic("implement me")
}

// ########################## Hash operations ##########################

func (m *MemoryStorage) HSet(key string, fields map[string][]byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
	}
	return db.hashStorage.HSet(key, fields)
}

func (m *MemoryStorage) HGet(key, field string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
		return nil, ErrKeyNotFound
	}
	return db.hashStorage.HGet(key, field)
}

func (m *MemoryStorage) HDel(key string, fields ...string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
		return 0, nil
	}
	return db.hashStorage.HDel(key, fields...)
}

func (m *MemoryStorage) HLen(key string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
		return 0, nil
	}
	return db.hashStorage.HLen(key)
}

// ########################## List operations ##########################

func (m *MemoryStorage) LPush(key string, values ...[]byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	list, ok := db.listStorage[key]
	if !ok {
		list = dslist.New()
		db.listStorage[key] = list
	} else if list.IsExpired() {
		list = dslist.New()
		db.listStorage[key] = list
		delete(db.expiry, key)
	}
	length := list.LPush(values...)
	return int(length), nil
}

func (m *MemoryStorage) RPush(key string, values ...[]byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	list, ok := db.listStorage[key]
	if !ok {
		list = dslist.New()
		db.listStorage[key] = list
	} else if list.IsExpired() {
		list = dslist.New()
		db.listStorage[key] = list
	}
	length := list.RPush(values...)
	return int(length), nil
}

func (m *MemoryStorage) LPop(key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	list, ok := db.listStorage[key]
	if !ok {
		return nil, consts.ErrKeyNotFound
	}
	if list.IsExpired() {
		delete(db.listStorage, key)
		return nil, consts.ErrKeyNotFound
	}
	value, ok := list.LPop()
	if !ok {
		delete(db.listStorage, key)
		return nil, consts.ErrKeyNotFound
	}
	if list.Len() == 0 {
		delete(db.listStorage, key)
	}
	return value, nil
}

func (m *MemoryStorage) RPop(key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	list, ok := db.listStorage[key]
	if !ok {
		return nil, consts.ErrKeyNotFound
	}
	if list.IsExpired() {
		delete(db.listStorage, key)
		return nil, consts.ErrKeyNotFound
	}
	value, ok := list.RPop()
	if !ok {
		delete(db.listStorage, key)
		return nil, consts.ErrKeyNotFound
	}
	if list.Len() == 0 {
		delete(db.listStorage, key)
	}
	return value, nil
}

func (m *MemoryStorage) LRange(key string, start, stop int) ([][]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	list, ok := db.listStorage[key]
	if !ok {
		return nil, consts.ErrKeyNotFound
	}
	if list.IsExpired() {
		delete(db.listStorage, key)
		return nil, consts.ErrKeyNotFound
	}
	return list.LRange(int64(start), int64(stop)), nil
}

// LLen returns the length of the list stored at key
func (m *MemoryStorage) LLen(key string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	list, ok := db.listStorage[key]
	if !ok {
		return 0, nil
	}
	if m.isExpired(db, key) {
		delete(db.listStorage, key)
		delete(db.expiry, key)
		return 0, nil
	}
	return int(list.Len()), nil
}

func (m *MemoryStorage) LIndex(key string, index int64) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	list, ok := db.listStorage[key]
	if !ok {
		return nil, consts.ErrKeyNotFound
	}
	if list.IsExpired() {
		delete(db.listStorage, key)
		return nil, consts.ErrKeyNotFound
	}
	value, ok := list.LIndex(index)
	if !ok {
		return nil, consts.ErrKeyNotFound
	}
	return value, nil
}

func (m *MemoryStorage) LSet(key string, index int64, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	list, ok := db.listStorage[key]
	if !ok {
		return consts.ErrKeyNotFound
	}
	if list.IsExpired() {
		delete(db.listStorage, key)
		return consts.ErrKeyNotFound
	}
	if !list.LSet(index, value) {
		return consts.ErrIndexOutOfRange
	}
	return nil
}

// ########################## Set operations ##########################

func (m *MemoryStorage) SAdd(key string, members ...string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
	}
	return db.setStorage.SAdd(key, members...)
}

func (m *MemoryStorage) SMembers(key string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
		return []string{}, nil
	}
	return db.setStorage.SMembers(key)
}

func (m *MemoryStorage) SRem(key string, members ...string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
		return 0, nil
	}
	return db.setStorage.SRem(key, members...)
}

func (m *MemoryStorage) SCard(key string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
		return 0, nil
	}
	return db.setStorage.SCard(key)
}

// ########################## Generic operations ##########################

func (m *MemoryStorage) Del(key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	deleted := false
	if _, ok := db.stringStorage.data[key]; ok {
		delete(db.stringStorage.data, key)
		deleted = true
	}
	if _, ok := db.hashStorage.data[key]; ok {
		delete(db.hashStorage.data, key)
		deleted = true
	}
	if _, ok := db.listStorage[key]; ok {
		delete(db.listStorage, key)
		deleted = true
	}
	delete(db.expiry, key)
	return deleted, nil
}

func (m *MemoryStorage) Exists(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
		return false
	}
	_, existsString := db.stringStorage.data[key]
	_, existsHash := db.hashStorage.data[key]
	_, existsList := db.listStorage[key]
	return existsString || existsHash || existsList
}

func (m *MemoryStorage) Expire(key string, expiration time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	if _, ok := db.stringStorage.data[key]; ok {
		db.expiry[key] = time.Now().Add(expiration)
		return true, nil
	}
	if _, ok := db.hashStorage.data[key]; ok {
		db.expiry[key] = time.Now().Add(expiration)
		return true, nil
	}
	if _, ok := db.listStorage[key]; ok {
		db.expiry[key] = time.Now().Add(expiration)
		return true, nil
	}
	return false, nil
}

func (m *MemoryStorage) TTL(key string) (time.Duration, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	if expiry, ok := db.expiry[key]; ok {
		ttl := time.Until(expiry)
		if ttl < 0 {
			m.deleteKey(db, key)
			return -2, nil // -2 表示键不存在
		}
		return ttl, nil
	}
	if m.Exists(key) {
		return -1, nil // -1 表示键没有过期时间
	}
	return -2, nil // -2 表示键不存在
}

func (m *MemoryStorage) Type(key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	if m.isExpired(db, key) {
		m.deleteKey(db, key)
		return "", ErrKeyNotFound
	}
	if _, ok := db.stringStorage.data[key]; ok {
		return "string", nil
	}
	if _, ok := db.hashStorage.data[key]; ok {
		return "hash", nil
	}
	if _, ok := db.listStorage[key]; ok {
		return "list", nil
	}
	return "", ErrKeyNotFound
}

func (m *MemoryStorage) Flush() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.databases {
		m.databases[i] = &Database{
			stringStorage: NewMemoryStringStorage(),
			hashStorage:   NewMemoryHashStorage(),
			listStorage:   make(map[string]*dslist.QuickList),
			setStorage:    NewMemorySetStorage(),
			expiry:        make(map[string]time.Time),
		}
	}
	return nil
}

func (m *MemoryStorage) FlushDB() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.databases[m.currentDBIndex] = &Database{
		stringStorage: NewMemoryStringStorage(),
		hashStorage:   NewMemoryHashStorage(),
		listStorage:   make(map[string]*dslist.QuickList),
		setStorage:    NewMemorySetStorage(),
		expiry:        make(map[string]time.Time),
	}
	return nil
}

func (m *MemoryStorage) isExpired(db *Database, key string) bool {
	if expiry, ok := db.expiry[key]; ok {
		if time.Now().After(expiry) {
			return true
		}
	}
	return false
}

func (m *MemoryStorage) deleteKey(db *Database, key string) {
	delete(db.stringStorage.data, key)
	delete(db.hashStorage.data, key)
	delete(db.listStorage, key)
	delete(db.setStorage.data, key)
	delete(db.expiry, key)
}
