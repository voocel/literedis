package sortedset

import (
	"math/bits"
	"math/rand"
)

const (
	maxLevel = 16
)

type Element struct {
	Member string
	Score  float64
}

type Level struct {
	forward *node
	span    int64
}

type node struct {
	Element
	backward *node
	level    []*Level
}

type skiplist struct {
	header *node
	tail   *node
	length int64
	level  int16
}

func makeNode(level int16, score float64, member string) *node {
	n := &node{
		Element: Element{
			Score:  score,
			Member: member,
		},
		level: make([]*Level, level),
	}
	for i := range n.level {
		n.level[i] = new(Level)
	}
	return n
}

func makeSkiplist() *skiplist {
	return &skiplist{
		level:  1,
		header: makeNode(maxLevel, 0, ""),
	}
}

func randomLevel() int16 {
	total := uint64(1)<<uint64(maxLevel) - 1
	k := rand.Uint64() % total
	return maxLevel - int16(bits.Len64(k+1)) + 1
}

func (skiplist *skiplist) insert(member string, score float64) *node {
	update := make([]*node, maxLevel)
	rank := make([]int64, maxLevel)

	node := skiplist.header
	for i := skiplist.level - 1; i >= 0; i-- {
		if i == skiplist.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}
		if node.level[i] != nil {
			for node.level[i].forward != nil &&
				(node.level[i].forward.Score < score ||
					(node.level[i].forward.Score == score && node.level[i].forward.Member < member)) { // same score, different key
				rank[i] += node.level[i].span
				node = node.level[i].forward
			}
		}
		update[i] = node
	}

	level := randomLevel()
	if level > skiplist.level {
		for i := skiplist.level; i < level; i++ {
			rank[i] = 0
			update[i] = skiplist.header
			update[i].level[i].span = skiplist.length
		}
		skiplist.level = level
	}

	node = makeNode(level, score, member)
	for i := int16(0); i < level; i++ {
		node.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = node

		node.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = (rank[0] - rank[i]) + 1
	}

	for i := level; i < skiplist.level; i++ {
		update[i].level[i].span++
	}

	if update[0] == skiplist.header {
		node.backward = nil
	} else {
		node.backward = update[0]
	}
	if node.level[0].forward != nil {
		node.level[0].forward.backward = node
	} else {
		skiplist.tail = node
	}
	skiplist.length++
	return node
}

func (skiplist *skiplist) removeNode(node *node, update []*node) {
	for i := int16(0); i < skiplist.level; i++ {
		if update[i].level[i].forward == node {
			update[i].level[i].span += node.level[i].span - 1
			update[i].level[i].forward = node.level[i].forward
		} else {
			update[i].level[i].span--
		}
	}
	if node.level[0].forward != nil {
		node.level[0].forward.backward = node.backward
	} else {
		skiplist.tail = node.backward
	}
	for skiplist.level > 1 && skiplist.header.level[skiplist.level-1].forward == nil {
		skiplist.level--
	}
	skiplist.length--
}

func (skiplist *skiplist) remove(member string, score float64) bool {
	update := make([]*node, maxLevel)
	node := skiplist.header
	for i := skiplist.level - 1; i >= 0; i-- {
		for node.level[i].forward != nil &&
			(node.level[i].forward.Score < score ||
				(node.level[i].forward.Score == score &&
					node.level[i].forward.Member < member)) {
			node = node.level[i].forward
		}
		update[i] = node
	}
	node = node.level[0].forward
	if node != nil && score == node.Score && node.Member == member {
		skiplist.removeNode(node, update)
		return true
	}
	return false
}

func (skiplist *skiplist) getRank(member string, score float64) int64 {
	var rank int64 = 0
	x := skiplist.header
	for i := skiplist.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			(x.level[i].forward.Score < score ||
				(x.level[i].forward.Score == score &&
					x.level[i].forward.Member <= member)) {
			rank += x.level[i].span
			x = x.level[i].forward
		}

		if x.Member == member {
			return rank
		}
	}
	return 0
}

func (skiplist *skiplist) getByRank(rank int64) *node {
	var i int64 = 0
	n := skiplist.header

	for level := skiplist.level - 1; level >= 0; level-- {
		for n.level[level].forward != nil && (i+n.level[level].span) <= rank {
			i += n.level[level].span
			n = n.level[level].forward
		}
		if i == rank {
			return n
		}
	}
	return nil
}

func (skiplist *skiplist) hasInRange(min Border, max Border) bool {
	if min.isIntersected(max) {
		return false
	}

	n := skiplist.tail
	if n == nil || !min.less(&n.Element) {
		return false
	}

	n = skiplist.header.level[0].forward
	if n == nil || !max.greater(&n.Element) {
		return false
	}
	return true
}

func (skiplist *skiplist) getFirstInRange(min Border, max Border) *node {
	if !skiplist.hasInRange(min, max) {
		return nil
	}
	n := skiplist.header
	// scan from top level
	for level := skiplist.level - 1; level >= 0; level-- {
		// if forward is not in range than move forward
		for n.level[level].forward != nil && !min.less(&n.level[level].forward.Element) {
			n = n.level[level].forward
		}
	}
	n = n.level[0].forward
	if !max.greater(&n.Element) {
		return nil
	}
	return n
}

func (skiplist *skiplist) getLastInRange(min Border, max Border) *node {
	if !skiplist.hasInRange(min, max) {
		return nil
	}
	n := skiplist.header
	for level := skiplist.level - 1; level >= 0; level-- {
		for n.level[level].forward != nil && max.greater(&n.level[level].forward.Element) {
			n = n.level[level].forward
		}
	}
	if !min.less(&n.Element) {
		return nil
	}
	return n
}

func (skiplist *skiplist) RemoveRange(min Border, max Border, limit int) (removed []*Element) {
	update := make([]*node, maxLevel)
	removed = make([]*Element, 0)
	node := skiplist.header
	for i := skiplist.level - 1; i >= 0; i-- {
		for node.level[i].forward != nil {
			if min.less(&node.level[i].forward.Element) { // already in range
				break
			}
			node = node.level[i].forward
		}
		update[i] = node
	}

	node = node.level[0].forward

	for node != nil {
		if !max.greater(&node.Element) {
			break
		}
		next := node.level[0].forward
		removedElement := node.Element
		removed = append(removed, &removedElement)
		skiplist.removeNode(node, update)
		if limit > 0 && len(removed) == limit {
			break
		}
		node = next
	}
	return removed
}

func (skiplist *skiplist) RemoveRangeByRank(start int64, stop int64) (removed []*Element) {
	var i int64 = 0
	update := make([]*node, maxLevel)
	removed = make([]*Element, 0)

	node := skiplist.header
	for level := skiplist.level - 1; level >= 0; level-- {
		for node.level[level].forward != nil && (i+node.level[level].span) < start {
			i += node.level[level].span
			node = node.level[level].forward
		}
		update[level] = node
	}

	i++
	node = node.level[0].forward

	for node != nil && i < stop {
		next := node.level[0].forward
		removedElement := node.Element
		removed = append(removed, &removedElement)
		skiplist.removeNode(node, update)
		node = next
		i++
	}
	return removed
}
