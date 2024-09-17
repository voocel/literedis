package dszset

import (
	"math/rand"
	"sync"
)

const (
	maxLevel    = 32 // Maximum number of levels in the skip list
	probability = 0.25
)

type scoreNode struct {
	Member string
	Score  float64
}

type skipListLevel struct {
	forward *node
	span    uint64
}

type node struct {
	sn       *scoreNode
	backward *node
	level    []skipListLevel
}

type SkipList struct {
	header *node
	tail   *node
	length int64
	level  int
	dict   map[string]*node
	mu     sync.RWMutex
}

func NewSkipList() *SkipList {
	return &SkipList{
		level: 1,
		header: &node{
			level: make([]skipListLevel, maxLevel),
		},
		dict: make(map[string]*node),
	}
}

func (sl *SkipList) randomLevel() int {
	level := 1
	for rand.Float64() < probability && level < maxLevel {
		level++
	}
	return level
}

func (sl *SkipList) Insert(score float64, member string) *node {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	if n, ok := sl.dict[member]; ok {
		if n.sn.Score == score {
			return n
		}
		sl.delete(n.sn.Score, member)
	}

	level := sl.randomLevel()
	if level > sl.level {
		for i := sl.level; i < level; i++ {
			sl.header.level[i] = skipListLevel{span: uint64(sl.length)}
		}
		sl.level = level
	}

	x := &node{
		sn:    &scoreNode{Member: member, Score: score},
		level: make([]skipListLevel, level),
	}

	sl.insertNode(score, x)
	sl.dict[member] = x
	return x
}

// insertNode inserts a new node into the skip list.
// It updates the necessary links and spans to maintain the skip list structure.
func (sl *SkipList) insertNode(score float64, x *node) {
	update := make([]*node, maxLevel)
	rank := make([]uint64, maxLevel)

	p := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		if i == sl.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}
		for p.level[i].forward != nil &&
			(p.level[i].forward.sn.Score < score ||
				(p.level[i].forward.sn.Score == score && p.level[i].forward.sn.Member < x.sn.Member)) {
			rank[i] += p.level[i].span
			p = p.level[i].forward
		}
		update[i] = p
	}

	for i := 0; i < len(x.level); i++ {
		x.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = x

		x.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = (rank[0] - rank[i]) + 1
	}

	for i := len(x.level); i < sl.level; i++ {
		update[i].level[i].span++
	}

	if update[0] == sl.header {
		x.backward = nil
	} else {
		x.backward = update[0]
	}
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x
	} else {
		sl.tail = x
	}
	sl.length++
}

func (sl *SkipList) Delete(score float64, member string) bool {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	return sl.delete(score, member)
}

func (sl *SkipList) delete(score float64, member string) bool {
	update := make([]*node, maxLevel)
	p := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		for p.level[i].forward != nil &&
			(p.level[i].forward.sn.Score < score ||
				(p.level[i].forward.sn.Score == score && p.level[i].forward.sn.Member < member)) {
			p = p.level[i].forward
		}
		update[i] = p
	}
	p = p.level[0].forward

	if p != nil && score == p.sn.Score && p.sn.Member == member {
		sl.deleteNode(p, update)
		delete(sl.dict, member)
		return true
	}
	return false
}

func (sl *SkipList) deleteNode(p *node, update []*node) {
	for i := 0; i < sl.level; i++ {
		if update[i].level[i].forward == p {
			update[i].level[i].span += p.level[i].span - 1
			update[i].level[i].forward = p.level[i].forward
		} else {
			update[i].level[i].span--
		}
	}
	if p.level[0].forward != nil {
		p.level[0].forward.backward = p.backward
	} else {
		sl.tail = p.backward
	}
	for sl.level > 1 && sl.header.level[sl.level-1].forward == nil {
		sl.level--
	}
	sl.length--
}

func (sl *SkipList) Score(member string) (float64, bool) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	if n, ok := sl.dict[member]; ok {
		return n.sn.Score, true
	}
	return 0, false
}

func (sl *SkipList) Range(start, stop int64) []string {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	var result []string
	if start < 0 {
		start = sl.length + start
	}
	if stop < 0 {
		stop = sl.length + stop
	}
	if start < 0 {
		start = 0
	}
	if stop < 0 {
		stop = 0
	}
	if start > stop || start >= sl.length {
		return result
	}
	if stop >= sl.length {
		stop = sl.length - 1
	}

	node := sl.header.level[0].forward
	for i := int64(0); i < start; i++ {
		node = node.level[0].forward
	}
	for i := start; i <= stop; i++ {
		result = append(result, node.sn.Member)
		node = node.level[0].forward
	}
	return result
}

func (sl *SkipList) RangeByScore(min, max float64) []string {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	var result []string
	node := sl.header.level[0].forward
	for node != nil && node.sn.Score <= max {
		if node.sn.Score >= min {
			result = append(result, node.sn.Member)
		}
		node = node.level[0].forward
	}
	return result
}

func (sl *SkipList) Len() int64 {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.length
}
