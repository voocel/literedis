package dshash

import (
	"literedis/internal/datastruct/base"
)

type Hash interface {
	base.DataStructure
	HSet(field string, value string) int
	HGet(field string) (string, bool)
	HDel(fields ...string) int
	HExists(field string) bool
	HLen() int64
	HKeys() []string
	HVals() []string
	HGetAll() map[string]string
}

type hashImpl struct {
	data map[string]string
	base.DataStructure
}
