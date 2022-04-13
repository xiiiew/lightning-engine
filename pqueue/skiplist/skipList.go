package skiplist

import (
	"errors"
	"github.com/shopspring/decimal"
	"lightning-engine/pqueue"
	"math/rand"
	"time"
)

const (
	SKIPLIST_DEFAULT_MAX_LEVEL = 36
	SKIPLIST_DEFAULT_P         = 0.25 //	跳表加一层索引的概率
)

// SkipList 跳表，从小到大排序
type SkipList struct {
	// 头节点，尾节点
	head, tail *SkipListNode
	// node总数
	size int64
	// 当前最高level
	level int
	// 最多有多少level
	maxLevel int
}

func NewSkipList(options ...Option) (*SkipList, error) {
	rand.Seed(time.Now().UnixNano())
	skipList := &SkipList{
		head:     nil,
		tail:     nil,
		size:     0,
		level:    1,
		maxLevel: SKIPLIST_DEFAULT_MAX_LEVEL,
	}

	for _, option := range options {
		if err := option(skipList); err != nil {
			return nil, err
		}
	}

	skipList.head = NewSkipListNode(skipList.maxLevel, decimal.NewFromInt(0), nil)

	return skipList, nil
}

// Option 可选参数设置
type Option func(skipList *SkipList) error

// MaxLevel 设置MaxLevel
func MaxLevel(maxLevel int) Option {
	return func(skipList *SkipList) error {
		if maxLevel <= 0 {
			return errors.New("SkipList max_level must gte than 1")
		}
		skipList.maxLevel = maxLevel
		return nil
	}
}

// Insert 插入节点
func (list *SkipList) Insert(score decimal.Decimal, value pqueue.INodeValue) *SkipListNode {
	rank := make([]int64, list.maxLevel)           // 新增节点每一层上是第几个节点
	update := make([]*SkipListNode, list.maxLevel) // 新增节点每一层的上一个节点
	p := list.head
	for i := list.level - 1; i >= 0; i-- {
		if i == list.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}
		// 下个节点存在，并且下个节点的score小于等于score时(score相同，按时间排序)
		for p.Next(i) != nil && p.Next(i).score.LessThanOrEqual(score) {
			rank[i] += p.level[i].span
			p = p.Next(i)
		}
		update[i] = p
	}

	level := list.randLevel()

	if level > list.level {
		for i := list.level; i < level; i++ {
			rank[i] = 0
			update[i] = list.head
			update[i].SetSpan(i, list.size)
		}
		list.level = level
	}
	newNode := NewSkipListNode(level, score, value)

	for i := 0; i < level; i++ {
		newNode.SetNext(i, update[i].Next(i))
		update[i].SetNext(i, newNode)

		newNode.SetSpan(i, update[i].Span(i)-(rank[0]-rank[i]))
		update[i].SetSpan(i, rank[0]-rank[i]+1)
	}

	// 处理新增节点的span
	for i := level; i < list.level; i++ {
		update[i].level[i].span++
	}
	// 处理新增节点的后退指针
	if update[0] == list.head {
		newNode.backward = nil
	} else {
		newNode.backward = update[0]
	}
	// 判断新插入的节点是不是最后一个节点
	if newNode.Next(0) != nil {
		newNode.Next(0).backward = newNode
	} else {
		list.tail = newNode
	}
	list.size++
	return newNode
}

// randLevel 随机索引层数
func (list *SkipList) randLevel() int {
	level := 1
	for rand.Int31n(100) < int32(100*SKIPLIST_DEFAULT_P) && level < list.maxLevel {
		level++
	}
	return level
}

// Find 查找节点，并返回路径
func (list *SkipList) Find(score decimal.Decimal, id string) (*SkipListNode, []*SkipListNode) {
	update := make([]*SkipListNode, list.maxLevel)

	p := list.head
	for i := list.level - 1; i >= 0; i-- {
		for p.Next(i) != nil && p.Next(i).score.LessThan(score) {
			p = p.Next(i)
		}
		update[i] = p
	}

	// 遍历最后一层，比较id
	for p.Next(0) != nil && p.Next(0).score.LessThanOrEqual(score) {
		p = p.Next(0)
		if p.score.Equal(score) && p.value.GetId() == id {
			break
		}
		update[0] = p
	}

	if p.score.Equal(score) && p.value.GetId() == id {
		return p, update
	}
	return nil, nil
}

// Update 更新节点值
func (list *SkipList) Update(score decimal.Decimal, id string, value pqueue.INodeValue) *SkipListNode {
	node, _ := list.Find(score, id)
	node.value = value
	return node
}

// Delete 删除节点
func (list *SkipList) Delete(score decimal.Decimal, id string) {
	node, update := list.Find(score, id)
	if node == nil || node == list.head {
		return
	}

	for i := 0; i < list.level; i++ {
		if update[i].Next(i) == node {
			// 修改span
			update[i].SetSpan(i, update[i].Span(i)+node.Span(i)-1)
			// 删除节点
			update[i].SetNext(i, node.Next(i))
		} else {
			update[i].level[i].span--
		}
	}

	// 处理node的后指针
	if node.Next(0) == nil {
		list.tail = update[0]
	} else {
		node.Next(0).backward = update[0]
	}

	// 处理删掉的是最高level的情况
	for list.level > 1 && list.head.Next(list.level-1) == nil {
		list.level--
	}
	list.size--
}

// First 获取第一个节点
func (list *SkipList) First() *SkipListNode {
	if list.size == 0 {
		return nil
	}
	return list.head.Next(0)
}
