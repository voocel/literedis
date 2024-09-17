package dsset

import (
	"sort"
)

type IntSet struct {
	contents []int64
}

func NewIntSet() *IntSet {
	return &IntSet{contents: make([]int64, 0)}
}

func (is *IntSet) Add(value int64) bool {
	index := sort.Search(len(is.contents), func(i int) bool {
		return is.contents[i] >= value
	})

	if index < len(is.contents) && is.contents[index] == value {
		return false
	}

	is.contents = append(is.contents, 0)
	copy(is.contents[index+1:], is.contents[index:])
	is.contents[index] = value
	return true
}

func (is *IntSet) Remove(value int64) bool {
	index := sort.Search(len(is.contents), func(i int) bool {
		return is.contents[i] >= value
	})

	if index < len(is.contents) && is.contents[index] == value {
		is.contents = append(is.contents[:index], is.contents[index+1:]...)
		return true
	}
	return false
}

func (is *IntSet) Contains(value int64) bool {
	index := sort.Search(len(is.contents), func(i int) bool {
		return is.contents[i] >= value
	})
	return index < len(is.contents) && is.contents[index] == value
}

func (is *IntSet) Len() int {
	return len(is.contents)
}

func (is *IntSet) ToSlice() []int64 {
	return append([]int64{}, is.contents...)
}
