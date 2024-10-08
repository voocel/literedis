package storage

import (
	"compress/gzip"
	"literedis/config"
	"literedis/internal/cluster"
	"literedis/internal/consts"
	"literedis/internal/datastruct/dslist"
	"path/filepath"
	"sync"
	"time"
)

// 确保实现了所有接口方法
var _ Storage = (*MemoryStorage)(nil)

const DefaultDBCount = 16

type Database struct {
	stringStorage *MemoryStringStorage
	hashStorage   *MemoryHashStorage
	listStorage   map[string]*dslist.QuickList
	setStorage    *MemorySetStorage
	zsetStorage   *MemoryZSetStorage
	expiry        map[string]time.Time
	mu            sync.RWMutex
}

type MemoryStorage struct {
	databases      []*Database
	currentDBIndex int
	mu             sync.RWMutex
	cluster        *cluster.Cluster
	RDB            *RDBStorage
	lastSaveTime   time.Time
	dirtyKeys      map[int]map[string]struct{} // 数据库索引 -> 脏键集合
	data           map[string]interface{}
	mutex          sync.RWMutex
}

func NewMemoryStorage(rdbConfig ...config.RDBConfig) Storage {
	ms := &MemoryStorage{
		databases:      make([]*Database, DefaultDBCount),
		currentDBIndex: 0,
		lastSaveTime:   time.Now(),
		dirtyKeys:      make(map[int]map[string]struct{}),
		data:           make(map[string]interface{}),
	}
	for i := 0; i < DefaultDBCount; i++ {
		ms.databases[i] = &Database{
			stringStorage: NewMemoryStringStorage(),
			hashStorage:   NewMemoryHashStorage(),
			listStorage:   make(map[string]*dslist.QuickList),
			setStorage:    NewMemorySetStorage(),
			zsetStorage:   NewMemoryZSetStorage(),
			expiry:        make(map[string]time.Time),
		}
	}

	var cfg config.RDBConfig
	if len(rdbConfig) > 0 {
		cfg = rdbConfig[0]
	} else {
		cfg = config.RDBConfig{
			Filename:         "dump.rdb",
			SaveInterval:     5 * time.Minute,
			CompressionLevel: gzip.DefaultCompression,
			AutoSaveChanges:  1000,
		}
	}

	ms.RDB = NewRDBStorage(cfg, ms)
	ms.StartExpirationChecker(time.Minute)
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

func (m *MemoryStorage) Set(key string, value []byte) error {
	db := m.getCurrentDB()
	db.mu.Lock()
	defer db.mu.Unlock()

	err := db.stringStorage.Set(key, value)
	if err != nil {
		return err
	}

	// 标记键为脏
	m.markDirty(m.currentDBIndex, key)
	m.IncrementRDBChanges()

	// 如果键已经存在过期时间，我们应该删除它
	delete(db.expiry, key)

	return nil
}

func (m *MemoryStorage) Get(key string) ([]byte, error) {
	db := m.getCurrentDB()
	db.mu.RLock()
	defer db.mu.RUnlock()

	if m.isExpired(key) {
		m.deleteKey(key)
		return nil, ErrKeyNotFound
	}

	return db.stringStorage.Get(key)
}

func (m *MemoryStorage) isExpired(key string) bool {
	db := m.getCurrentDB()
	expireTime, exists := db.expiry[key]
	if !exists {
		return false
	}
	return time.Now().After(expireTime)
}

func (m *MemoryStorage) deleteKey(key string) {
	db := m.getCurrentDB()
	delete(db.stringStorage.data, key)
	delete(db.hashStorage.data, key)
	delete(db.listStorage, key)

	if members, err := db.setStorage.SMembers(key); err == nil {
		_, _ = db.setStorage.SRem(key, members...)
	}
	if db.zsetStorage != nil {
		_, _ = db.zsetStorage.ZRem(key, "")
	}

	delete(db.expiry, key)
}

func (m *MemoryStorage) Append(key string, value []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	length, err := db.stringStorage.Append(key, value)
	if err == nil {
		m.markDirty(m.currentDBIndex, key)
		m.IncrementRDBChanges()
	}
	return length, err
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
	length, err := db.stringStorage.SetRange(key, offset, value)
	if err == nil {
		m.markDirty(m.currentDBIndex, key)
		m.IncrementRDBChanges()
	}
	return length, err
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
	if m.isExpired(key) {
		m.deleteKey(key)
	}
	count, err := db.hashStorage.HSet(key, fields)
	if err == nil {
		m.markDirty(m.currentDBIndex, key)
		m.IncrementRDBChanges()
	}
	return count, err
}

func (m *MemoryStorage) HGet(key, field string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	if m.isExpired(key) {
		m.deleteKey(key)
		return nil, ErrKeyNotFound
	}
	return db.hashStorage.HGet(key, field)
}

func (m *MemoryStorage) HDel(key string, fields ...string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	if m.isExpired(key) {
		m.deleteKey(key)
		return 0, nil
	}
	count, err := db.hashStorage.HDel(key, fields...)
	if err == nil {
		m.markDirty(m.currentDBIndex, key)
		m.IncrementRDBChanges()
	}
	return count, err
}

func (m *MemoryStorage) HLen(key string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	if m.isExpired(key) {
		m.deleteKey(key)
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
	m.markDirty(m.currentDBIndex, key)
	m.IncrementRDBChanges()
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
	m.markDirty(m.currentDBIndex, key)
	m.IncrementRDBChanges()
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
	m.markDirty(m.currentDBIndex, key)
	m.IncrementRDBChanges()
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
	m.markDirty(m.currentDBIndex, key)
	m.IncrementRDBChanges()
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
	if m.isExpired(key) {
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
	m.markDirty(m.currentDBIndex, key)
	m.IncrementRDBChanges()
	return nil
}

// ########################## Set operations ##########################

func (m *MemoryStorage) SAdd(key string, members ...string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	if m.isExpired(key) {
		m.deleteKey(key)
	}
	count, err := db.setStorage.SAdd(key, members...)
	if err == nil {
		m.markDirty(m.currentDBIndex, key)
		m.IncrementRDBChanges()
	}
	return count, err
}

func (m *MemoryStorage) SMembers(key string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	if m.isExpired(key) {
		m.deleteKey(key)
		return []string{}, nil
	}
	return db.setStorage.SMembers(key)
}

func (m *MemoryStorage) SRem(key string, members ...string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	if m.isExpired(key) {
		m.deleteKey(key)
		return 0, nil
	}
	count, err := db.setStorage.SRem(key, members...)
	if err == nil {
		m.markDirty(m.currentDBIndex, key)
		m.IncrementRDBChanges()
	}
	return count, err
}

func (m *MemoryStorage) SCard(key string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	if m.isExpired(key) {
		m.deleteKey(key)
		return 0, nil
	}
	return db.setStorage.SCard(key)
}

// ########################## ZSet operations ##########################

func (m *MemoryStorage) ZAdd(key string, score float64, member string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	count, err := db.zsetStorage.ZAdd(key, score, member)
	if err == nil {
		m.markDirty(m.currentDBIndex, key)
		m.IncrementRDBChanges()
	}
	return count, err
}

func (m *MemoryStorage) ZScore(key, member string) (float64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	return db.zsetStorage.ZScore(key, member)
}

func (m *MemoryStorage) ZRem(key string, member string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	count, err := db.zsetStorage.ZRem(key, member)
	if err == nil {
		m.markDirty(m.currentDBIndex, key)
		m.IncrementRDBChanges()
	}
	return count, err
}

func (m *MemoryStorage) ZRange(key string, start, stop int64) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	return db.zsetStorage.ZRange(key, start, stop)
}

func (m *MemoryStorage) ZRangeByScore(key string, min, max float64) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	return db.zsetStorage.ZRangeByScore(key, min, max)
}

func (m *MemoryStorage) ZCard(key string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	return db.zsetStorage.ZCard(key)
}

func (m *MemoryStorage) ZIncrBy(key string, increment float64, member string) (float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	db := m.getCurrentDB()
	score, err := db.zsetStorage.ZIncrBy(key, increment, member)
	if err == nil {
		m.markDirty(m.currentDBIndex, key)
		m.IncrementRDBChanges()
	}
	return score, err
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
	if deleted {
		m.markDirty(m.currentDBIndex, key)
		m.IncrementRDBChanges()
	}
	return deleted, nil
}

func (m *MemoryStorage) Exists(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	if m.isExpired(key) {
		m.deleteKey(key)
		return false
	}
	_, existsString := db.stringStorage.data[key]
	_, existsHash := db.hashStorage.data[key]
	_, existsList := db.listStorage[key]
	return existsString || existsHash || existsList
}

func (m *MemoryStorage) Expire(key string, expiration time.Duration) (bool, error) {
	db := m.getCurrentDB()
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := m.getAnyKey(key); !exists {
		return false, nil
	}

	if expiration > 0 {
		db.expiry[key] = time.Now().Add(expiration)
	} else {
		delete(db.expiry, key)
	}
	m.markDirty(m.currentDBIndex, key)
	m.IncrementRDBChanges()
	return true, nil
}

func (m *MemoryStorage) TTL(key string) (time.Duration, error) {
	db := m.getCurrentDB()
	db.mu.RLock()
	defer db.mu.RUnlock()

	expireTime, exists := db.expiry[key]
	if !exists {
		// 键存在，但没有设置过期时间
		if _, keyExists := m.getAnyKey(key); keyExists {
			return -1 * time.Second, nil
		}
		// 键不存在
		return -2 * time.Second, nil
	}

	ttl := time.Until(expireTime)
	if ttl <= 0 {
		m.deleteKey(key)
		return -2 * time.Second, nil
	}

	return ttl, nil
}

func (m *MemoryStorage) Type(key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db := m.getCurrentDB()
	if m.isExpired(key) {
		m.deleteKey(key)
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
			zsetStorage:   NewMemoryZSetStorage(),
			expiry:        make(map[string]time.Time),
		}
	}
	m.dirtyKeys = make(map[int]map[string]struct{})
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
		zsetStorage:   NewMemoryZSetStorage(),
		expiry:        make(map[string]time.Time),
	}
	m.dirtyKeys[m.currentDBIndex] = make(map[string]struct{})
	return nil
}

func (m *MemoryStorage) cleanExpired() {
	for _, db := range m.databases {
		db.mu.Lock()
		now := time.Now()
		for key, expireTime := range db.expiry {
			if now.After(expireTime) {
				m.deleteKey(key)
				delete(db.expiry, key)
			}
		}
		db.mu.Unlock()
	}
}

func (m *MemoryStorage) StartExpirationChecker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			m.cleanExpired()
		}
	}()
}

func (m *MemoryStorage) getAnyKey(key string) (interface{}, bool) {
	db := m.getCurrentDB()
	if _, ok := db.stringStorage.data[key]; ok {
		return nil, true
	}
	if _, ok := db.hashStorage.data[key]; ok {
		return nil, true
	}
	if _, ok := db.listStorage[key]; ok {
		return nil, true
	}
	// 使用 SCard 来检查集合是否存在
	if count, err := db.setStorage.SCard(key); err == nil && count > 0 {
		return nil, true
	}
	// 假设 zsetStorage 有一个 ZCard 方法
	if count, err := db.zsetStorage.ZCard(key); err == nil && count > 0 {
		return nil, true
	}
	return nil, false
}

// SaveRDB 保存 RDB 文件
func (m *MemoryStorage) SaveRDB() error {
	return m.RDB.SaveIncremental()
}

// LoadRDB 加载 RDB 文件
func (m *MemoryStorage) LoadRDB() error {
	return m.RDB.Load()
}

// 在每次修改操作后调用此方法
func (m *MemoryStorage) markDirty(dbIndex int, key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.dirtyKeys[dbIndex]; !ok {
		m.dirtyKeys[dbIndex] = make(map[string]struct{})
	}
	m.dirtyKeys[dbIndex][key] = struct{}{}
}

func (m *MemoryStorage) IncrementRDBChanges() {
	if m.RDB != nil {
		m.RDB.incrementChanges()
	}
}

func (m *MemoryStorage) GetRDBStats() RDBStats {
	if m.RDB != nil {
		return m.RDB.GetStats()
	}
	return RDBStats{}
}

func (m *MemoryStorage) SetRDBConfig(config config.RDBConfig) {
	if m.RDB == nil {
		m.RDB = NewRDBStorage(config, m)
	} else {
		m.RDB.Config = config
	}
}

func (m *MemoryStorage) Keys(pattern string) []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var keys []string
	for key := range m.data {
		matched, err := filepath.Match(pattern, key)
		if err == nil && matched {
			keys = append(keys, key)
		}
	}
	return keys
}
