package billing

import (
	"time"

	"encore.app/billing/domain"
	"encore.app/pkg/currency"
)

type (
	// GetBillResponse represents the response returned by the GetBill API.
	// The Bill field contains the current state of the requested bill.
	GetBillResponse struct {
		Bill Bill `json:"bill"`
	}

	// AddItemRequest represents the payload to add a new line item to a bill,
	// including the item's name and price in the smallest currency unit.
	AddItemRequest struct {
		Name  string `json:"name"`
		Price int64  `json:"price"`
	}

	// AddItemResponse represents the response after attempting to add a line item,
	// indicating whether the operation was successful.
	AddItemResponse struct {
		CurrentBill Bill `json:"current_bill"`
	}

	// CloseBillingRequest represents the payload to request closing a bill,
	// including the target currency for conversion.
	CloseBillingRequest struct {
		Currency string `json:"currency"`
	}

	// CloseBillingResponse represents the response after closing a bill,
	// including totals in both the original and converted currencies.
	CloseBillingResponse struct {
		OriginalCurrencyTotal  Amount `json:"originalCurrencyTotal"`
		ConvertedCurrencyTotal Amount `json:"convertedCurrencyTotal"`
	}

	// Amount represents a monetary value in a specific currency.
	Amount struct {
		Currency        string `json:"currency"`
		Amount          int64  `json:"amount"`
		FormattedAmount string `json:"formattedAmount"`
	}

	// OpenBillingRequest represents the payload to create a new bill,
	// specifying the currency for the bill.
	OpenBillingRequest struct {
		Currency string `json:"currency"`
	}

	// OpenBillingResponse represents the response after creating a new bill,
	// including the billing ID and the currency of the bill.
	OpenBillingResponse struct {
		Currency  string `json:"currency"`
		BillingID string `json:"billingId"`
	}
)

// Bill represents a billing record containing multiple items, currency info,
// total amount, and status (open or closed).
type Bill struct {
	BillingID      string              `json:"billingId"`
	Status         string              `json:"status"`
	Currency       string              `json:"currency"`
	Total          int64               `json:"total"`
	Items          []Item              `json:"items"`
	Conversion     BillExchangeResonse `json:"conversion"`
	CreatedAt      time.Time           `json:"createdAt"`
	ClosedAt       *time.Time          `json:"closedAt"`
	FormattedTotal string              `json:"formattedTotal"`
}

func fromDomainBillToBillReponse(b domain.Bill) Bill {
	var items []Item
	for _, i := range b.Items {
		items = append(items, fromDomainItemToResponse(i))
	}

	return Bill{
		BillingID:      b.BillingID,
		Status:         string(b.Status),
		Currency:       string(b.Currency),
		Total:          b.GetTotal(),
		Items:          items,
		Conversion:     fromDomainBillingExchangeToResponse(b.Conversion),
		FormattedTotal: currency.FormatString(string(b.Currency), b.GetTotal()),
		CreatedAt:      b.CreatedAt,
		ClosedAt:       b.ClosedAt,
	}
}

// Item represents a line item in a bill, including price, name, and
// optional idempotency key to prevent duplicate entries.
type Item struct {
	Name  string `json:"name"`
	Price int64  `json:"price"`
}

func fromDomainItemToResponse(i domain.Item) Item {
	return Item{
		Name:  i.Name,
		Price: i.Price,
	}
}

// BillExchangeResonse represents a currency conversion entry associated with a Bill.
type BillExchangeResonse struct {
	BaseCurrency   string  `json:"baseCurrency"`
	TargetCurrency string  `json:"targetCurrency"`
	Rate           float64 `json:"rate"`
	Total          int64   `json:"total"`
}

func fromDomainBillingExchangeToResponse(exc domain.BillExchange) BillExchangeResonse {
	return BillExchangeResonse{
		BaseCurrency:   string(exc.BaseCurrency),
		TargetCurrency: string(exc.TargetCurrency),
		Rate:           exc.Rate,
		Total:          exc.Total,
	}
}
