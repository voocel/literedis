package dsziplist

import (
	"encoding/binary"
	"errors"
	"math"
	"time"
)

const (
	ZIP_END      = 255
	ZIP_BIGLEN   = 254
	ZIP_STR_MASK = 0xc0
	ZIP_INT_MASK = 0x30
	ZIP_STR_06B  = 0 << 6
	ZIP_STR_14B  = 1 << 6
	ZIP_STR_32B  = 2 << 6
	ZIP_INT_16B  = 0xc0 | 0<<4
	ZIP_INT_32B  = 0xc0 | 1<<4
	ZIP_INT_64B  = 0xc0 | 2<<4
	ZIP_INT_24B  = 0xc0 | 3<<4
	ZIP_INT_8B   = 0xfe
)

// ZipList represents a compact list structure
type ZipList struct {
	bytes      []byte    // Raw byte data of the ziplist
	length     uint16    // Number of entries in the ziplist
	tailOffset uint32    // Offset to the last entry in the ziplist
	expireAt   time.Time // Expiration time of the ziplist
}

// NewZipList creates and initializes a new ZipList
func NewZipList() *ZipList {
	zl := &ZipList{
		bytes:      make([]byte, 10),
		length:     0,
		tailOffset: 10,
	}
	// Initialize the ziplist header
	binary.LittleEndian.PutUint32(zl.bytes[0:4], 10)  // Total bytes
	binary.LittleEndian.PutUint16(zl.bytes[4:6], 0)   // Number of entries
	binary.LittleEndian.PutUint32(zl.bytes[6:10], 10) // Tail offset
	return zl
}

// Len returns the number of entries in the ziplist
func (zl *ZipList) Len() int64 {
	return int64(zl.length)
}

// Bytes returns the raw byte data of the ziplist
func (zl *ZipList) Bytes() []byte {
	return zl.bytes
}

// SetExpire sets the expiration time for the ziplist
func (zl *ZipList) SetExpire(t time.Time) {
	zl.expireAt = t
}

// IsExpired checks if the ziplist has expired
func (zl *ZipList) IsExpired() bool {
	return !zl.expireAt.IsZero() && time.Now().After(zl.expireAt)
}

// Insert adds a new entry to the end of the ziplist
func (zl *ZipList) Insert(value []byte) error {
	encodedValue, err := encodeEntry(value)
	if err != nil {
		return err
	}

	requiredSpace := uint32(len(encodedValue))
	if requiredSpace > math.MaxUint32-uint32(len(zl.bytes)) {
		return errors.New("ziplist too large")
	}

	newSize := len(zl.bytes) + int(requiredSpace)
	if cap(zl.bytes) < newSize {
		// Allocate new memory with extra capacity
		newBytes := make([]byte, newSize, newSize*2)
		copy(newBytes, zl.bytes)
		zl.bytes = newBytes
	} else {
		// Extend the existing slice
		zl.bytes = zl.bytes[:newSize]
	}

	copy(zl.bytes[zl.tailOffset:], encodedValue)

	zl.length++
	zl.tailOffset += requiredSpace

	// Update the ziplist header
	binary.LittleEndian.PutUint32(zl.bytes[0:4], uint32(len(zl.bytes)))
	binary.LittleEndian.PutUint16(zl.bytes[4:6], zl.length)
	binary.LittleEndian.PutUint32(zl.bytes[6:10], zl.tailOffset)

	return nil
}

// Delete removes an entry from the ziplist at the specified index
func (zl *ZipList) Delete(index int) bool {
	if index < 0 || index >= int(zl.length) {
		return false
	}

	// Find the offset of the entry to delete
	offset := uint32(10)
	for i := 0; i < index; i++ {
		_, entryLen := decodeEntry(zl.bytes[offset:])
		offset += uint32(entryLen)
	}

	// Remove the entry
	_, entryLen := decodeEntry(zl.bytes[offset:])
	copy(zl.bytes[offset:], zl.bytes[offset+uint32(entryLen):])
	zl.bytes = zl.bytes[:len(zl.bytes)-entryLen]

	// Update ziplist metadata
	zl.length--
	zl.tailOffset -= uint32(entryLen)

	// Update the ziplist header
	binary.LittleEndian.PutUint32(zl.bytes[0:4], uint32(len(zl.bytes)))
	binary.LittleEndian.PutUint16(zl.bytes[4:6], zl.length)
	binary.LittleEndian.PutUint32(zl.bytes[6:10], zl.tailOffset)

	return true
}

// encodeEntry encodes a byte slice into a ziplist entry
func encodeEntry(value []byte) ([]byte, error) {
	if len(value) <= 63 {
		return append([]byte{byte(ZIP_STR_06B | len(value))}, value...), nil
	} else if len(value) <= 16383 {
		header := []byte{byte(ZIP_STR_14B | (len(value) >> 8)), byte(len(value) & 0xff)}
		return append(header, value...), nil
	} else if len(value) <= math.MaxUint32 {
		header := make([]byte, 5)
		header[0] = ZIP_STR_32B
		binary.LittleEndian.PutUint32(header[1:], uint32(len(value)))
		return append(header, value...), nil
	}
	return nil, errors.New("value too large")
}

// decodeEntry decodes a ziplist entry into a byte slice
func decodeEntry(data []byte) ([]byte, int) {
	if len(data) == 0 {
		return nil, 0
	}

	encoding := data[0]
	switch {
	case encoding>>6 == ZIP_STR_06B>>6:
		length := int(encoding & 0x3f)
		return data[1 : 1+length], 1 + length
	case encoding>>6 == ZIP_STR_14B>>6:
		length := int(encoding&0x3f)<<8 | int(data[1])
		return data[2 : 2+length], 2 + length
	case encoding == ZIP_STR_32B:
		length := binary.LittleEndian.Uint32(data[1:5])
		return data[5 : 5+length], 5 + int(length)
	default:
		return nil, 0
	}
}

// encodeLength encodes a length value into a byte slice.
func encodeLength(length uint32) []byte {
	if length < ZIP_BIGLEN {
		return []byte{byte(length)}
	}
	buf := make([]byte, 5)
	buf[0] = ZIP_BIGLEN
	binary.LittleEndian.PutUint32(buf[1:], length)
	return buf
}

// decodeLength decodes a length value from a byte slice.
func decodeLength(data []byte) (uint32, int) {
	if data[0] < ZIP_BIGLEN {
		return uint32(data[0]), 1
	}
	return binary.LittleEndian.Uint32(data[1:5]), 5
}

// encodeInteger encodes an integer value into a byte slice.
func encodeInteger(value int64) []byte {
	if value >= 0 && value <= 12 {
		return []byte{byte(ZIP_INT_16B | value)}
	} else if value >= math.MinInt8 && value <= math.MaxInt8 {
		return []byte{ZIP_INT_8B, byte(value)}
	} else if value >= math.MinInt16 && value <= math.MaxInt16 {
		buf := []byte{ZIP_INT_16B}
		return binary.LittleEndian.AppendUint16(buf, uint16(value))
	} else if value >= math.MinInt32 && value <= math.MaxInt32 {
		buf := []byte{ZIP_INT_32B}
		return binary.LittleEndian.AppendUint32(buf, uint32(value))
	} else {
		buf := []byte{ZIP_INT_64B}
		return binary.LittleEndian.AppendUint64(buf, uint64(value))
	}
}

// decodeInteger decodes an integer value from a byte slice
func decodeInteger(data []byte) (int64, int) {
	encoding := data[0]
	switch {
	case encoding >= ZIP_INT_16B && encoding <= ZIP_INT_16B+12:
		return int64(encoding & 0x0f), 1
	case encoding == ZIP_INT_8B:
		return int64(int8(data[1])), 2
	case encoding == ZIP_INT_16B:
		return int64(int16(binary.LittleEndian.Uint16(data[1:3]))), 3
	case encoding == ZIP_INT_32B:
		return int64(int32(binary.LittleEndian.Uint32(data[1:5]))), 5
	case encoding == ZIP_INT_64B:
		return int64(binary.LittleEndian.Uint64(data[1:9])), 9
	default:
		return 0, 0
	}
}

// Get retrieves the value at the specified index
func (zl *ZipList) Get(index int) ([]byte, bool) {
	if index < 0 || index >= int(zl.length) {
		return nil, false
	}

	offset := uint32(10)
	for i := 0; i < index; i++ {
		_, entryLen := decodeEntry(zl.bytes[offset:])
		offset += uint32(entryLen)
	}

	value, _ := decodeEntry(zl.bytes[offset:])
	return value, true
}

// Set updates the value at the specified index
func (zl *ZipList) Set(index int, value []byte) bool {
	if index < 0 || index >= int(zl.length) {
		return false
	}

	offset := uint32(10)
	for i := 0; i < index; i++ {
		_, entryLen := decodeEntry(zl.bytes[offset:])
		offset += uint32(entryLen)
	}

	_, oldEntryLen := decodeEntry(zl.bytes[offset:])
	newValue, err := encodeEntry(value)
	if err != nil {
		return false
	}

	lenDiff := len(newValue) - oldEntryLen
	if lenDiff != 0 {
		if lenDiff > 0 {
			zl.bytes = append(zl.bytes, make([]byte, lenDiff)...)
		}
		copy(zl.bytes[offset+uint32(len(newValue)):], zl.bytes[offset+uint32(oldEntryLen):])
	}
	copy(zl.bytes[offset:], newValue)

	zl.tailOffset += uint32(lenDiff)
	binary.LittleEndian.PutUint32(zl.bytes[0:4], uint32(len(zl.bytes)))
	binary.LittleEndian.PutUint32(zl.bytes[6:10], zl.tailOffset)

	return true
}
