package dszset

// ZSet represents a sorted set data structure.
// It allows for efficient ranking and scoring of string members.
type ZSet interface {
	Add(score float64, member string) bool
	Score(member string) (float64, bool)
	Remove(member string) bool
	Range(start, stop int64) []string
	RangeByScore(min, max float64) []string
	Len() int64
	IncrBy(increment float64, member string) float64
}

// NewZSet returns a new ZSet implementation.
// The implementation can be switched here without affecting the users of the ZSet.
func NewZSet() ZSet {
	return NewSkipListZSet()
}
