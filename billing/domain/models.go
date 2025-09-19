package domain

import "time"

// Bill represents the core domain entity for billing.
type Bill struct {
	ID         int64        `json:"id"`
	BillingID  string       `json:"billingId"`
	Status     BillStatus   `json:"status"`
	Currency   Currency     `json:"currency"`
	Total      int64        `json:"total"`
	Items      []Item       `json:"items"`
	Conversion BillExchange `json:"conversion"`
	CreatedAt  time.Time    `json:"createdAt"`
	ClosedAt   *time.Time   `json:"closedAt"`
}

// BillStatus represents the possible states of a bill.
type BillStatus string

const (
	// BillStatusOpen represents the status OPEN.
	BillStatusOpen BillStatus = "OPEN"

	// BillStatusClosed represents the status CLOSED.
	BillStatusClosed BillStatus = "CLOSED"
)

// Currency represents supported currencies.
type Currency string

const (
	// CurrencyUSD represent USD
	CurrencyUSD Currency = "USD"

	// CurrencyGEL represent GEL
	CurrencyGEL Currency = "GEL"
)

// Item represents a line item in a bill.
type Item struct {
	ID             int64  `json:"id"`
	BillingID      string `json:"billingId"`
	Name           string `json:"name"`
	Price          int64  `json:"price"`
	IdempotencyKey string `json:"idempotencyKey"`
}

// BillExchange represents currency conversion info for a bill.
type BillExchange struct {
	ID             int64    `json:"id"`
	BillID         string   `json:"billId"`
	BaseCurrency   Currency `json:"baseCurrency"`
	TargetCurrency Currency `json:"targetCurrency"`
	Rate           float64  `json:"rate"`
	Total          int64    `json:"total"`
}

// AddItem adds a line item to the bill and updates the total.
func (b *Bill) AddItem(item Item) {
	b.Items = append(b.Items, item)
	b.Total += item.Price
}

// Close marks the bill as closed at a given timestamp and updates the total.
func (b *Bill) Close(closedAt time.Time) {
	b.ClosedAt = &closedAt
	b.Status = BillStatusClosed
	b.Total = b.GetTotal()
}

// GetTotal calculates the sum of all item prices in the bill.
func (b *Bill) GetTotal() int64 {
	var total int64
	for _, item := range b.Items {
		total += item.Price
	}
	return total
}

// IsClosed returns true if the bill is closed.
func (b *Bill) IsClosed() bool {
	return b.Status == BillStatusClosed
}

// IsOpen returns true if the bill is open.
func (b *Bill) IsOpen() bool {
	return b.Status == BillStatusOpen
}
