package base

import (
	"time"
)

type DataStructure interface {
	Type() string
	Len() int64
	Expire() time.Time
	SetExpire(t time.Time)
	IsExpired() bool
}
