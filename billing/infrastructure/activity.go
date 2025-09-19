package infrastructure

import (
	"context"
	"fmt"

	"encore.app/billing/domain"
	"encore.dev/rlog"
)

// BillingActivities defines the set of Temporal activities related to billing.
// Each method on this struct represents an activity that can be executed
// asynchronously by a Temporal workflow. Activities should be **idempotent**
// and short-running, as recommended by Temporal best practices.
type BillingActivities struct {
	repository domain.Repository
}

// NewBillingActivity creates a new BillingActivities instance with the given repository.
// This is used to register the activities with the Temporal worker.
func NewBillingActivity(repository domain.Repository) domain.BillingActivities {
	return &BillingActivities{
		repository: repository,
	}
}

// SetBillingToCloseActivity calculates the total for a Bill and closes it by
// calling CloseBill in the database context.
func (a *BillingActivities) SetBillingToCloseActivity(ctx context.Context, bill domain.Bill) error {
	if bill.BillingID == "" {
		return fmt.Errorf("close bill: missing billing id")
	}

	bill.Total = bill.GetTotal()
	if err := a.repository.CloseBilling(ctx, bill); err != nil {
		return fmt.Errorf("close bill %s: %w", bill.BillingID, err)
	}

	return nil
}

// UpsertBillingToDBActivity inserts a new Bill or updates an existing one in the database.
func (a *BillingActivities) UpsertBillingToDBActivity(ctx context.Context, bill domain.Bill) error {
	if bill.BillingID == "" {
		return fmt.Errorf("upsert bill: missing billing id")
	}
	if bill.Currency == "" {
		return fmt.Errorf("upsert bill %s: missing currency", bill.BillingID)
	}
	if err := a.repository.SaveBill(ctx, &bill); err != nil {
		return fmt.Errorf("upsert bill %s: %w", bill.BillingID, err)
	}
	return nil
}

// InsertLineItemActivity inserts or updates a single Item in the database.
func (a *BillingActivities) InsertLineItemActivity(ctx context.Context, item domain.Item) error {
	if item.BillingID == "" {
		return fmt.Errorf("upsert item: missing billing id")
	}
	if item.Name == "" {
		return fmt.Errorf("upsert item: missing name")
	}
	if item.Price <= 0 {
		return fmt.Errorf("upsert item: invalid price %d", item.Price)
	}
	if err := a.repository.SaveItem(ctx, &item); err != nil {
		return fmt.Errorf("upsert item for bill %s: %w", item.BillingID, err)
	}
	return nil
}

// InsertBillExchangeActivity is the Temporal activity wrapper
func (a *BillingActivities) InsertBillExchangeActivity(ctx context.Context, bill domain.Bill) error {
	if bill.BillingID == "" {
		return fmt.Errorf("insert exchange: missing billing id")
	}
	if bill.Conversion.TargetCurrency == "" {
		return fmt.Errorf("insert exchange %s: missing target currency", bill.BillingID)
	}

	if err := a.repository.SaveExchange(ctx, &bill); err != nil {
		return fmt.Errorf("persist exchange for bill %s: %w", bill.BillingID, err)
	}

	return nil
}

// RevertBillCloseActivity is to handle revert bill closing
func (a *BillingActivities) RevertBillCloseActivity(ctx context.Context, bill domain.Bill) error {
	rlog.Info("BillingActivities.RevertBillCloseActivity", "billing-id", bill.BillingID)

	return a.repository.RevertBillClosing(ctx, bill.BillingID)
}
