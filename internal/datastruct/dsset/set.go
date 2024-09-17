package dsset

import "time"

const (
	useIntSet = iota
	useBitmap
	useHashTable
)

type Set interface {
	Add(members ...string) int
	Remove(members ...string) int
	IsMember(member string) bool
	Members() []string
	Len() int
	SetExpire(expireAt time.Time)
	IsExpired() bool
	Union(other Set) Set
	Intersection(other Set) Set
	Difference(other Set) Set
}

// The implementation can be switched here without affecting the users of the Set.
func NewSet() Set {
	return NewBasicSet() // Use this line to use the basic implementation
	//return NewOptimizedSet() // Use this line to use the optimized implementation
}
