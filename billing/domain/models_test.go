package domain_test

import (
	"testing"
	"time"

	"encore.app/billing/domain"
	"github.com/stretchr/testify/assert"
)

func TestBill_AddItem(t *testing.T) {
	bill := &domain.Bill{Status: domain.BillStatusOpen}
	item := domain.Item{Price: 1000, Name: "Test Item"}

	bill.AddItem(item)

	assert.Len(t, bill.Items, 1)
	assert.Equal(t, int64(1000), bill.Total)
}

func TestBill_GetTotal(t *testing.T) {
	bill := &domain.Bill{
		Items: []domain.Item{
			{Price: 1000},
			{Price: 500},
		},
	}

	total := bill.GetTotal()
	assert.Equal(t, int64(1500), total)
}

func TestBill_Close(t *testing.T) {
	bill := &domain.Bill{
		Items:  []domain.Item{{Price: 1000}, {Price: 500}},
		Status: domain.BillStatusOpen,
	}

	now := time.Now()
	bill.Close(now)

	assert.Equal(t, domain.BillStatusClosed, bill.Status)
	assert.Equal(t, now, *bill.ClosedAt)
	assert.Equal(t, int64(1500), bill.Total)
	assert.True(t, bill.IsClosed())
	assert.False(t, bill.IsOpen())
}

func TestBill_IsOpen_IsClosed(t *testing.T) {
	bill := &domain.Bill{Status: domain.BillStatusOpen}
	assert.True(t, bill.IsOpen())
	assert.False(t, bill.IsClosed())

	bill.Status = domain.BillStatusClosed
	assert.False(t, bill.IsOpen())
	assert.True(t, bill.IsClosed())
}
