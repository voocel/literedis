package storage

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"hash/crc32"
	"io"
	"os"
	"sync/atomic"
	"time"

	"literedis/internal/datastruct/dslist"
	"literedis/pkg/log"
)

const currentVersion = 1

type rdbHeader struct {
	Version int
}

type RDBStorage struct {
	Filename         string
	Storage          *MemoryStorage
	savingInProgress atomic.Bool
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

	// 获取文件大小
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	fileSize := fileInfo.Size()

	// 读取除校验和外的所有数据
	data := make([]byte, fileSize-4) // 4 bytes for checksum
	if _, err := io.ReadFull(file, data); err != nil {
		return err
	}

	// 读取存储的校验和
	var storedChecksum uint32
	if err := binary.Read(file, binary.LittleEndian, &storedChecksum); err != nil {
		return err
	}

	// 计算实际校验和
	calculatedChecksum := crc32.ChecksumIEEE(data)

	// 验证校验和
	if storedChecksum != calculatedChecksum {
		return errors.New("RDB file is corrupted: checksum mismatch")
	}

	// 创建一个gzip读取器
	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	decoder := gob.NewDecoder(gzipReader)

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

func (r *RDBStorage) SaveIncremental() error {
	r.Storage.mu.RLock()
	defer r.Storage.mu.RUnlock()

	if len(r.Storage.dirtyKeys) == 0 {
		log.Info("No changes since last save, skipping RDB save")
		return nil
	}

	tempFilename := r.Filename + ".temp"
	file, err := os.Create(tempFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	defer gzipWriter.Close()

	encoder := gob.NewEncoder(gzipWriter)

	// 写入头部信息
	header := rdbHeader{Version: currentVersion}
	if err := encoder.Encode(header); err != nil {
		return err
	}

	// 写入增量数据
	for dbIndex, keys := range r.Storage.dirtyKeys {
		if err := encoder.Encode(dbIndex); err != nil {
			return err
		}
		if err := encoder.Encode(len(keys)); err != nil {
			return err
		}
		for key := range keys {
			if err := r.encodeKey(encoder, dbIndex, key); err != nil {
				return err
			}
		}
	}

	// 写入结束标记
	if err := encoder.Encode(-1); err != nil {
		return err
	}

	// 关闭gzip写入器
	if err := gzipWriter.Close(); err != nil {
		return err
	}

	// 计算校验和
	checksum := crc32.ChecksumIEEE(buf.Bytes())

	// 写入校验和
	if err := binary.Write(file, binary.LittleEndian, checksum); err != nil {
		return err
	}

	// 将缓冲区的内容写入文件
	if _, err := file.Write(buf.Bytes()); err != nil {
		return err
	}

	// 原子性地替换旧的RDB文件
	if err := os.Rename(tempFilename, r.Filename); err != nil {
		return err
	}

	// 清除脏键记录
	r.Storage.dirtyKeys = make(map[int]map[string]struct{})
	r.Storage.lastSaveTime = time.Now()

	return nil
}

func (r *RDBStorage) encodeKey(encoder *gob.Encoder, dbIndex int, key string) error {
	db := r.Storage.databases[dbIndex]
	if err := encoder.Encode(key); err != nil {
		return err
	}

	// 根据键的类型进行编码
	// 这里需要根据你的具体实现来编码不同类型的数据
	// 例如:
	if value, ok := db.stringStorage.data[key]; ok {
		if err := encoder.Encode("string"); err != nil {
			return err
		}
		return encoder.Encode(value)
	}
	// 对其他类型的数据进行类似的处理...

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

	// 编码列数据
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

func (r *RDBStorage) BackgroundSave() error {
	if !r.savingInProgress.CompareAndSwap(false, true) {
		return errors.New("background save already in progress")
	}

	go func() {
		defer r.savingInProgress.Store(false)

		if err := r.SaveIncremental(); err != nil {
			log.Errorf("Background RDB save failed: %v", err)
		} else {
			log.Info("Background RDB save completed successfully")
		}
	}()

	return nil
}
