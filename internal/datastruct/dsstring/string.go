package dsstring

import (
	"time"
)

const (
	SDS_MAX_PREALLOC = 1024 * 1024 // 1MB
)

type SDS struct {
	buf      []byte
	len      int
	free     int
	expireAt time.Time
}

func NewSDS(init string) *SDS {
	buf := make([]byte, len(init), len(init)*2)
	copy(buf, init)
	return &SDS{
		buf:  buf,
		len:  len(init),
		free: len(init),
	}
}

func (s *SDS) Type() string {
	return "string"
}

func (s *SDS) Len() int64 {
	return int64(s.len)
}

func (s *SDS) Expire() time.Time {
	return s.expireAt
}

func (s *SDS) SetExpire(t time.Time) {
	s.expireAt = t
}

func (s *SDS) IsExpired() bool {
	return !s.expireAt.IsZero() && time.Now().After(s.expireAt)
}

func (s *SDS) Get() []byte {
	return s.buf[:s.len]
}

func (s *SDS) Set(value []byte) {
	s.buf = make([]byte, len(value), len(value)*2)
	copy(s.buf, value)
	s.len = len(value)
	s.free = len(value)
}

func (s *SDS) Append(value []byte) int {
	if s.free < len(value) {
		s.grow(len(value))
	}
	copy(s.buf[s.len:], value)
	s.len += len(value)
	s.free -= len(value)
	return s.len
}

func (s *SDS) grow(addLen int) {
	newLen := s.len + addLen
	if newLen < SDS_MAX_PREALLOC {
		newLen *= 2
	} else {
		newLen += SDS_MAX_PREALLOC
	}
	newBuf := make([]byte, newLen)
	copy(newBuf, s.buf)
	s.buf = newBuf
	s.free = newLen - s.len
}

func (s *SDS) GetRange(start, end int) []byte {
	if start < 0 {
		start = s.len + start
	}
	if end < 0 {
		end = s.len + end
	}
	if start < 0 {
		start = 0
	}
	if end >= s.len {
		end = s.len - 1
	}
	if start > end {
		return []byte{}
	}
	return s.buf[start : end+1]
}

func (s *SDS) SetRange(offset int, value []byte) int {
	if offset < 0 {
		return s.len
	}
	if offset >= s.len {
		s.Append(make([]byte, offset-s.len))
		s.Append(value)
	} else {
		endOffset := offset + len(value)
		if endOffset > s.len {
			s.Append(make([]byte, endOffset-s.len))
		}
		copy(s.buf[offset:], value)
	}
	return s.len
}

func (s *SDS) Bytes() []byte {
	return s.buf
}
