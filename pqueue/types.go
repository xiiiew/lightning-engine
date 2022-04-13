package pqueue

import "github.com/shopspring/decimal"

// INodeValue 节点存储的值
type INodeValue interface {
	// GetId 返回节点值的唯一ID
	GetId() string
	// GetAmount 返回节点值的数量
	GetAmount() decimal.Decimal
	// GetUserId 返回节点值的用户id
	GetUserId() int64
	// SetAmount 更新节点值的数量
	SetAmount(decimal.Decimal)
}
