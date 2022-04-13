package skiplist

import (
	"github.com/shopspring/decimal"
	"lightning-engine/pqueue"
)

type SkipListLevel struct {
	// 向前指针
	forward *SkipListNode
	// 到下一个node的距离
	span int64
}

type SkipListNode struct {
	// 向后指针
	backward *SkipListNode
	// 索引
	level []SkipListLevel
	// 存储的值，需要实现INodeValue接口
	value pqueue.INodeValue
	// 用于排序，使用高精度
	score decimal.Decimal
}

func NewSkipListNode(level int, score decimal.Decimal, value pqueue.INodeValue) *SkipListNode {
	return &SkipListNode{
		backward: nil,
		level:    make([]SkipListLevel, level),
		value:    value,
		score:    score,
	}
}

// Next 第i层的下个元素
func (node *SkipListNode) Next(i int) *SkipListNode {
	return node.level[i].forward
}

// SetNext 设置第i层的下个元素
func (node *SkipListNode) SetNext(i int, next *SkipListNode) {
	node.level[i].forward = next
}

// Span 第层的span值
func (node *SkipListNode) Span(i int) int64 {
	return node.level[i].span
}

// SetSpan 设置第i层的span
func (node *SkipListNode) SetSpan(i int, span int64) {
	node.level[i].span = span
}

// Pre 上一个元素
func (node *SkipListNode) Pre() *SkipListNode {
	return node.backward
}

// Value 获取节点值
func (node *SkipListNode) Value() pqueue.INodeValue {
	return node.value
}

// Score 获取节点分数
func (node *SkipListNode) Score() decimal.Decimal {
	return node.score
}
