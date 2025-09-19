package domain

import (
	"context"
)

type BillingActivities interface {
	SetBillingToCloseActivity(ctx context.Context, bill Bill) error
	UpsertBillingToDBActivity(ctx context.Context, bill Bill) error
	InsertLineItemActivity(ctx context.Context, item Item) error
	InsertBillExchangeActivity(ctx context.Context, bill Bill) error
}
