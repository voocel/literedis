package storage

import (
	"encoding/gob"
	"os"
	"time"

	"literedis/internal/datastruct/dslist"
	"literedis/pkg/log"
)

type RDBStorage struct {
	Filename string
	Storage  *MemoryStorage
}

func NewRDBStorage(filename string, storage *MemoryStorage) *RDBStorage {
	return &RDBStorage{
		Filename: filename,
		Storage:  storage,
	}
}

func (r *RDBStorage) Save() error {
	log.Infof("Saving RDB to file: %s", r.Filename)
	file, err := os.Create(r.Filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)

	r.Storage.mu.RLock()
	defer r.Storage.mu.RUnlock()

	// 编码数据库数量
	if err := encoder.Encode(len(r.Storage.databases)); err != nil {
		return err
	}

	// 编码每个数据库
	for i, db := range r.Storage.databases {
		if err := r.encodeDatabase(encoder, i, db); err != nil {
			return err
		}
	}

	log.Infof("RDB save completed")
	return nil
}

func (r *RDBStorage) Load() error {
	log.Infof("Loading RDB from file: %s", r.Filename)
	file, err := os.Open(r.Filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)

	// 解码数据库数量
	var dbCount int
	if err := decoder.Decode(&dbCount); err != nil {
		return err
	}

	r.Storage.mu.Lock()
	defer r.Storage.mu.Unlock()

	r.Storage.databases = make([]*Database, dbCount)

	// 解码每个数据库
	for i := 0; i < dbCount; i++ {
		if err := r.decodeDatabase(decoder); err != nil {
			return err
		}
	}

	log.Infof("RDB load completed")
	return nil
}

func (r *RDBStorage) encodeDatabase(encoder *gob.Encoder, index int, db *Database) error {
	// 编码数据库索引
	if err := encoder.Encode(index); err != nil {
		return err
	}

	// 编码字符串数据
	if err := encoder.Encode(db.stringStorage.data); err != nil {
		return err
	}

	// 编码哈希数据
	if err := encoder.Encode(db.hashStorage.data); err != nil {
		return err
	}

	// 编码列表数据
	if err := encoder.Encode(db.listStorage); err != nil {
		return err
	}

	// 编码集合数据
	if err := encoder.Encode(db.setStorage.data); err != nil {
		return err
	}

	// 编码有序集合数据
	if err := encoder.Encode(db.zsetStorage.data); err != nil {
		return err
	}

	// 编码过期时间数据
	return encoder.Encode(db.expiry)
}

func (r *RDBStorage) decodeDatabase(decoder *gob.Decoder) error {
	var dbIndex int
	if err := decoder.Decode(&dbIndex); err != nil {
		return err
	}

	db := &Database{
		stringStorage: NewMemoryStringStorage(),
		hashStorage:   NewMemoryHashStorage(),
		listStorage:   make(map[string]*dslist.QuickList),
		setStorage:    NewMemorySetStorage(),
		zsetStorage:   NewMemoryZSetStorage(),
		expiry:        make(map[string]time.Time),
	}

	// 解码字符串数据
	if err := decoder.Decode(&db.stringStorage.data); err != nil {
		return err
	}

	// 解码哈希数据
	if err := decoder.Decode(&db.hashStorage.data); err != nil {
		return err
	}

	// 解码列表数据
	if err := decoder.Decode(&db.listStorage); err != nil {
		return err
	}

	// 解码集合数据
	if err := decoder.Decode(&db.setStorage.data); err != nil {
		return err
	}

	// 解码有序集合数据
	if err := decoder.Decode(&db.zsetStorage.data); err != nil {
		return err
	}

	// 解码过期时间数据
	if err := decoder.Decode(&db.expiry); err != nil {
		return err
	}

	r.Storage.databases[dbIndex] = db
	return nil
}
