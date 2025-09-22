package domain

import (
	"context"
)

// BillingActivities defines the set of activities related to billing operations
// that can be executed within a Temporal workflow.
type BillingActivities interface {
	SetBillingToCloseActivity(ctx context.Context, bill Bill) error
	UpsertBillingToDBActivity(ctx context.Context, bill Bill) error
	InsertLineItemActivity(ctx context.Context, item Item) error
	InsertBillExchangeActivity(ctx context.Context, bill Bill) error
	RevertBillCloseActivity(ctx context.Context, bill Bill) error
}
