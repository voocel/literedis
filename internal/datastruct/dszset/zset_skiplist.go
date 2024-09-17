package dszset

// SkipListZSet implements the ZSet interface using a skip list data structure.
// Skip lists provide O(log N) time complexity for add, remove, and search operations.
type SkipListZSet struct {
	sl *SkipList
}

func NewSkipListZSet() *SkipListZSet {
	return &SkipListZSet{
		sl: NewSkipList(),
	}
}

func (z *SkipListZSet) Add(score float64, member string) bool {
	return z.sl.Insert(score, member) != nil
}

func (z *SkipListZSet) Score(member string) (float64, bool) {
	return z.sl.Score(member)
}

func (z *SkipListZSet) Remove(member string) bool {
	score, ok := z.Score(member)
	if !ok {
		return false
	}
	return z.sl.Delete(score, member)
}

func (z *SkipListZSet) Range(start, stop int64) []string {
	return z.sl.Range(start, stop)
}

func (z *SkipListZSet) RangeByScore(min, max float64) []string {
	return z.sl.RangeByScore(min, max)
}

func (z *SkipListZSet) Len() int64 {
	return z.sl.Len()
}

func (z *SkipListZSet) IncrBy(increment float64, member string) float64 {
	score, exists := z.Score(member)
	if exists {
		newScore := score + increment
		z.Remove(member)
		z.Add(newScore, member)
		return newScore
	}

	z.Add(increment, member)
	return increment
}
