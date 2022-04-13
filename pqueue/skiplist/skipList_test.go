package skiplist

import (
	"github.com/shopspring/decimal"
	"strconv"
	"testing"
)

type entrust struct {
	Id     string
	Amount decimal.Decimal
}

func (e *entrust) GetId() string {
	return e.Id
}

func (e *entrust) GetAmount() decimal.Decimal {
	return e.Amount
}

var skiplist *SkipList
var skiplistDesc *SkipListDesc

func TestMain(m *testing.M) {
	skiplist, _ = NewSkipList()
	for i := 0; i < 100; i++ {
		e := &entrust{
			Id:     strconv.Itoa(i),
			Amount: decimal.NewFromInt(int64(i)),
		}
		skiplist.Insert(decimal.NewFromInt(int64(i)), e)
	}
	skiplistDesc, _ = NewSkipListDesc()
	for i := 0; i < 100; i++ {
		e := &entrust{
			Id:     strconv.Itoa(i),
			Amount: decimal.NewFromInt(int64(i)),
		}
		skiplistDesc.Insert(decimal.NewFromInt(int64(i)), e)
	}
	m.Run()
}

func TestSkipList_Insert(t *testing.T) {
	skiplist, _ := NewSkipList()

	for i := 0; i < 100000; i++ {
		e := &entrust{
			Id:     "1",
			Amount: decimal.Decimal{},
		}
		skiplist.Insert(decimal.NewFromInt(int64(i)), e)
	}
	t.Logf("size: %d, level: %d\n", skiplist.size, skiplist.level)
}

func TestSkipList_Find(t *testing.T) {
	success := 0
	failed := 0
	for i := 0; i < 100000; i++ {
		node, _ := skiplist.Find(decimal.NewFromInt(int64(i)), strconv.Itoa(i))
		if node == nil || !node.score.Equal(decimal.NewFromInt(int64(i))) {
			failed++
		} else {
			success++
		}
	}
	t.Logf("success:%d failed:%d", success, failed)
}

func TestRange(t *testing.T) {
	p := skiplist.head
	for p != nil {
		p = p.Next(0)
	}
}

func TestSkipList_Delete(t *testing.T) {
	for i := 0; i < 100; i++ {
		skiplist.Delete(decimal.NewFromInt(int64(i)), strconv.Itoa(i))
	}
	if skiplist.size > 0 || skiplist.level > 1 {
		t.Error("error", skiplist.size, skiplist.level)
	} else {
		t.Logf("success")
	}
}

func TestSkipListDesc_Delete(t *testing.T) {
	for i := 0; i < 100; i++ {
		skiplistDesc.Delete(decimal.NewFromInt(int64(i)), strconv.Itoa(i))
	}
	if skiplistDesc.size > 0 || skiplistDesc.level > 1 {
		t.Error("error", skiplistDesc.size, skiplistDesc.level)
	} else {
		t.Logf("success")
	}
}

func TestSkipList_First(t *testing.T) {
	f1 := skiplist.First()
	f2 := skiplistDesc.First()
	t.Logf("f1.first = %s, f2.first = %s", f1.Value().GetAmount(), f2.Value().GetAmount())
}
