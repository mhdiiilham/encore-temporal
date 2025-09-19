package usecases

import (
	"encoding/json"
	"time"

	"encore.app/billing/domain"
)

// CreateBillRequest represents the payload for creating a new bill.
// Currency must be either "USD" or "GEL".
type CreateBillRequest struct {
	Currency string `json:"currency"`
}

// AddItemRequest represents the payload to add a new item to an existing bill.
type AddItemRequest struct {
	BillingID string `json:"billingId"`
	Name      string `json:"name"`
	Price     int64  `json:"price"`
}

// CloseBillRequest represents the payload to close an existing bill.
// Currency must match one of the supported currencies.
type CloseBillRequest struct {
	BillingID string              `json:"billingId"`
	Currency  string              `json:"currency"`
	ClosedAt  time.Time           `json:"closedAt"`
	Exchange  domain.BillExchange `json:"exchange"`
}

// PayloadToBytes convert request argument `r` to []byte
// to generate idempotency key.
func PayloadToBytes(r any) []byte {
	b, _ := json.Marshal(r)
	return b
}
