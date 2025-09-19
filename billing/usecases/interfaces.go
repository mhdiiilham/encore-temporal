package usecases

import (
	"context"

	"encore.app/billing/domain"
)

// BillingUseCase defines the interface for billing business operations
type BillingUseCase interface {
	CreateBill(ctx context.Context, req CreateBillRequest) (string, error)
	GetBill(ctx context.Context, billingID string) (domain.Bill, error)
	AddItem(ctx context.Context, req AddItemRequest) error
	CloseBill(ctx context.Context, req CloseBillRequest) (domain.Bill, error)
}

// WorkflowClient defines the interface for workflow operations
type WorkflowClient interface {
	StartWorkflow(ctx context.Context, workflowID string, bill *domain.Bill) error
	QueryWorkflow(ctx context.Context, workflowID string) (domain.Bill, error)
	SignalWorkflow(ctx context.Context, workflowID string, signal string, data interface{}) error
	IsWorkflowRunning(ctx context.Context, workflowID string) (bool, error)
}

// CreateBillRequest represents the payload for creating a new bill.
// Currency must be either "USD" or "GEL".
type CreateBillRequest struct {
	Currency string `json:"currency" validate:"required,oneof=USD GEL"`
}

// AddItemRequest represents the payload to add a new item to an existing bill.
type AddItemRequest struct {
	BillingID string `json:"billingId" validate:"required,uuid"`
	Name      string `json:"name" validate:"required,min=1,max=100"`
	Price     int64  `json:"price" validate:"required,min=1,max=99999999"`
}

// CloseBillRequest represents the payload to close an existing bill.
// Currency must match one of the supported currencies.
type CloseBillRequest struct {
	BillingID string `json:"billingId" validate:"required,uuid"`
	Currency  string `json:"currency" validate:"required,oneof=USD GEL"`
}
