package domain

import (
	"context"
)

// Repository defines the interface for all data operations
// Consolidated for simplicity - handles bills, items, and exchanges
type Repository interface {
	// Bill operations
	GetBill(ctx context.Context, billingID string) (Bill, error)
	SaveBill(ctx context.Context, bill *Bill) error
	CloseBilling(ctx context.Context, billing Bill) error
	RevertBillClosing(ctx context.Context, billingID string) error

	// Item operations
	SaveItem(ctx context.Context, item *Item) error
	GetItemsByBillID(ctx context.Context, billID string) ([]Item, error)

	// Exchange operations
	SaveExchange(ctx context.Context, bill *Bill) error
	GetExchangeByBillID(ctx context.Context, billID string) (BillExchange, error)
}
