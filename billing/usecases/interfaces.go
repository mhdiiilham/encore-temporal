package usecases

import (
	"context"

	"encore.app/billing/domain"
)

// BillingUseCase defines the interface for billing business operations
type BillingUseCase interface {
	CreateBill(ctx context.Context, req CreateBillRequest) (string, error)
	GetBill(ctx context.Context, billingID string) (domain.Bill, error)
	AddItem(ctx context.Context, req AddItemRequest) (domain.Bill, error)
	CloseBill(ctx context.Context, req CloseBillRequest) (domain.Bill, error)
}

// WorkflowClient defines the interface for workflow operations
type WorkflowClient interface {
	StartWorkflow(ctx context.Context, workflowID string, bill *domain.Bill) error
	QueryWorkflow(ctx context.Context, workflowID string) (domain.Bill, error)
	SignalWorkflow(ctx context.Context, workflowID string, signal string, data interface{}) error
	IsWorkflowRunning(ctx context.Context, workflowID string) (bool, error)
}
