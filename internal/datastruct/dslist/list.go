package dslist

import (
	"literedis/internal/datastruct/dsziplist"
	"time"
)

const (
	defaultListNodeSize = 64 // Default number of elements in each node
)

// ListNode represents a node in the quicklist
type ListNode struct {
	ziplist *dsziplist.ZipList
	prev    *ListNode
	next    *ListNode
}

// QuickList is our main list structure
type QuickList struct {
	head     *ListNode
	tail     *ListNode
	len      int
	expireAt time.Time
}

// New creates a new QuickList
func New() *QuickList {
	return &QuickList{
		head: &ListNode{ziplist: dsziplist.NewZipList()},
		tail: &ListNode{ziplist: dsziplist.NewZipList()},
	}
}

// Len returns the number of elements in the list
func (ql *QuickList) Len() int64 {
	return int64(ql.len)
}

// IsExpired checks if the list has expired
func (ql *QuickList) IsExpired() bool {
	return !ql.expireAt.IsZero() && time.Now().After(ql.expireAt)
}

// SetExpire sets the expiration time
func (ql *QuickList) SetExpire(t time.Time) {
	ql.expireAt = t
}

// Expire returns the expiration time
func (ql *QuickList) Expire() time.Time {
	return ql.expireAt
}

// LPush inserts all the specified values at the head of the list
func (ql *QuickList) LPush(values ...[]byte) int64 {
	for _, value := range values {
		if ql.head == nil || ql.head.ziplist.Len() >= defaultListNodeSize {
			newNode := &ListNode{ziplist: dsziplist.NewZipList()}
			newNode.next = ql.head
			if ql.head != nil {
				ql.head.prev = newNode
			}
			ql.head = newNode
			if ql.tail == nil {
				ql.tail = newNode
			}
		}
		err := ql.head.ziplist.Insert(value)
		if err != nil {
			// 处理错误,可能需要创建新的节点
			continue
		}
		ql.len++
	}
	return int64(ql.len)
}

// RPush inserts all the specified values at the tail of the list
func (ql *QuickList) RPush(values ...[]byte) int64 {
	for _, value := range values {
		if ql.tail == nil || ql.tail.ziplist.Len() >= defaultListNodeSize {
			newNode := &ListNode{ziplist: dsziplist.NewZipList()}
			newNode.prev = ql.tail
			if ql.tail != nil {
				ql.tail.next = newNode
			}
			ql.tail = newNode
			if ql.head == nil {
				ql.head = newNode
			}
		}
		err := ql.tail.ziplist.Insert(value)
		if err != nil {
			// 处理错误,可能需要创建新的节点
			continue
		}
		ql.len++
	}
	return int64(ql.len)
}

// LPop removes and returns the first element of the list
func (ql *QuickList) LPop() ([]byte, bool) {
	if ql.len == 0 {
		return nil, false
	}
	value, ok := ql.head.ziplist.Get(0)
	if !ok {
		return nil, false
	}
	if ql.head.ziplist.Delete(0) {
		ql.len--
		if ql.head.ziplist.Len() == 0 {
			ql.head = ql.head.next
			if ql.head == nil {
				ql.tail = nil
			} else {
				ql.head.prev = nil
			}
		}
		return value, true
	}
	return nil, false
}

// RPop removes and returns the last element of the list
func (ql *QuickList) RPop() ([]byte, bool) {
	if ql.len == 0 {
		return nil, false
	}
	lastIndex := int(ql.tail.ziplist.Len() - 1)
	value, ok := ql.tail.ziplist.Get(lastIndex)
	if !ok {
		return nil, false
	}
	if ql.tail.ziplist.Delete(lastIndex) {
		ql.len--
		if ql.tail.ziplist.Len() == 0 {
			ql.tail = ql.tail.prev
			if ql.tail == nil {
				ql.head = nil
			} else {
				ql.tail.next = nil
			}
		}
		return value, true
	}
	return nil, false
}

// LRange returns the specified elements of the list
func (ql *QuickList) LRange(start, stop int64) [][]byte {
	if start < 0 {
		start = int64(ql.len) + start
	}
	if stop < 0 {
		stop = int64(ql.len) + stop
	}
	if start < 0 {
		start = 0
	}
	if stop >= int64(ql.len) {
		stop = int64(ql.len) - 1
	}
	if start > stop {
		return [][]byte{}
	}

	result := make([][]byte, 0, stop-start+1)
	node := ql.head
	var index int64 = 0
	for node != nil && index <= stop {
		nodeLen := node.ziplist.Len()
		if index+nodeLen > start {
			startIndex := 0
			if index < start {
				startIndex = int(start - index)
			}
			endIndex := int(nodeLen)
			if index+nodeLen > stop+1 {
				endIndex = int(stop - index + 1)
			}
			for i := startIndex; i < endIndex; i++ {
				value, _ := node.ziplist.Get(i)
				result = append(result, value)
			}
		}
		index += nodeLen
		node = node.next
	}
	return result
}

// LIndex returns the element at index in the list
func (ql *QuickList) LIndex(index int64) ([]byte, bool) {
	if index < 0 {
		index = int64(ql.len) + index
	}
	if index < 0 || index >= int64(ql.len) {
		return nil, false
	}

	node := ql.head
	var currentIndex int64 = 0
	for node != nil {
		nodeLen := node.ziplist.Len()
		if currentIndex+nodeLen > index {
			return node.ziplist.Get(int(index - currentIndex))
		}
		currentIndex += nodeLen
		node = node.next
	}
	return nil, false
}

// LSet sets the list element at index to value
func (ql *QuickList) LSet(index int64, value []byte) bool {
	if index < 0 {
		index = int64(ql.len) + index
	}
	if index < 0 || index >= int64(ql.len) {
		return false
	}

	node := ql.head
	var currentIndex int64 = 0
	for node != nil {
		nodeLen := node.ziplist.Len()
		if currentIndex+nodeLen > index {
			return node.ziplist.Set(int(index-currentIndex), value)
		}
		currentIndex += nodeLen
		node = node.next
	}
	return false
}

// LGet returns the element at index in the list
func (ql *QuickList) LGet(index int64) ([]byte, bool) {
	if index < 0 {
		index = int64(ql.len) + index
	}
	if index < 0 || index >= int64(ql.len) {
		return nil, false
	}

	node := ql.head
	var currentIndex int64 = 0
	for node != nil {
		nodeLen := node.ziplist.Len()
		if currentIndex+nodeLen > index {
			return node.ziplist.Get(int(index - currentIndex))
		}
		currentIndex += nodeLen
		node = node.next
	}
	return nil, false
}
