package usecases

import (
	"context"
	"fmt"

	"encore.app/billing/domain"
	"encore.app/pkg/clock"
	"encore.app/pkg/convertion"
	"encore.app/pkg/generator"
)

// billingUseCase implements BillingUseCase interface
type billingUseCase struct {
	repo           domain.Repository
	workflowClient WorkflowClient
	idGenerator    generator.IDProvider
	clock          clock.Clock
}

// NewBillingUseCase creates a new billing use case
func NewBillingUseCase(
	repo domain.Repository,
	workflowClient WorkflowClient,
	idGenerator generator.IDProvider,
	clock clock.Clock,
) BillingUseCase {
	return &billingUseCase{
		repo:           repo,
		workflowClient: workflowClient,
		idGenerator:    idGenerator,
		clock:          clock,
	}
}

// CreateBill creates a new bill and starts the workflow
func (u *billingUseCase) CreateBill(ctx context.Context, req CreateBillRequest) (string, error) {
	if err := u.validateCreateBillRequest(req); err != nil {
		return "", err
	}

	billingID := u.idGenerator.GenerateBillingID("Bill")
	bill := &domain.Bill{
		BillingID: billingID,
		Status:    domain.BillStatusOpen,
		Currency:  domain.Currency(req.Currency),
		Total:     0,
		Items:     []domain.Item{},
		CreatedAt: u.clock.Now(),
	}

	if err := u.workflowClient.StartWorkflow(ctx, bill.BillingID, bill); err != nil {
		return "", fmt.Errorf("failed to start workflow: %w", err)
	}

	return bill.BillingID, nil
}

// GetBill retrieves a bill by ID
func (u *billingUseCase) GetBill(ctx context.Context, billingID string) (domain.Bill, error) {
	if billingID == "" {
		return domain.Bill{}, domain.ValidationError{Field: "billingID", Message: "billing ID is required"}
	}

	bill, err := u.workflowClient.QueryWorkflow(ctx, billingID)
	if err == nil {
		return bill, nil
	}

	bill, err = u.repo.GetBill(ctx, billingID)
	if err != nil {
		return domain.Bill{}, domain.ErrBillNotFound
	}

	return bill, nil
}

// AddItem adds an item to a bill
func (u *billingUseCase) AddItem(ctx context.Context, req AddItemRequest) error {
	if err := u.validateAddItemRequest(req); err != nil {
		return err
	}

	bill, err := u.GetBill(ctx, req.BillingID)
	if err != nil {
		return err
	}

	if bill.IsClosed() {
		return domain.ErrBillClosed
	}

	idempotencyKey := u.idGenerator.GenerateIdempotencyKey("idem", PayloadToBytes(req))
	item := &domain.Item{
		BillingID:      req.BillingID,
		Name:           req.Name,
		Price:          req.Price,
		IdempotencyKey: idempotencyKey,
	}

	if err := u.workflowClient.SignalWorkflow(ctx, req.BillingID, domain.SignalAddLineItem, item); err != nil {
		return fmt.Errorf("failed to add item: %w", err)
	}

	return nil
}

// CloseBill closes a bill
func (u *billingUseCase) CloseBill(ctx context.Context, req CloseBillRequest) (domain.Bill, error) {
	if err := u.validateCloseBillRequest(req); err != nil {
		return domain.Bill{}, err
	}

	bill, err := u.GetBill(ctx, req.BillingID)
	if err != nil {
		return domain.Bill{}, err
	}

	if bill.IsClosed() {
		return domain.Bill{}, domain.ErrBillClosed
	}

	if req.Currency != "" {
		converted, rate, err := convertion.ConvertAmount(bill.Total, string(bill.Currency), req.Currency)
		if err != nil {
			return domain.Bill{}, domain.ErrFailedToConvertBill
		}

		bill.Conversion = domain.BillExchange{
			BillID:         bill.BillingID,
			BaseCurrency:   bill.Currency,
			TargetCurrency: domain.Currency(req.Currency),
			Rate:           rate,
			Total:          converted,
		}
	}

	if err := u.workflowClient.SignalWorkflow(ctx, req.BillingID, domain.SignalCloseBill, bill.Conversion); err != nil {
		return domain.Bill{}, fmt.Errorf("failed to close bill: %w", err)
	}

	closedAt := u.clock.Now()
	bill.Close(closedAt)
	return bill, nil
}

// Validation methods
func (u *billingUseCase) validateCreateBillRequest(req CreateBillRequest) error {
	if req.Currency == "" {
		return domain.ValidationError{Field: "currency", Message: "currency is required"}
	}
	if req.Currency != "USD" && req.Currency != "GEL" {
		return domain.ValidationError{Field: "currency", Message: "currency must be USD or GEL"}
	}
	return nil
}

func (u *billingUseCase) validateAddItemRequest(req AddItemRequest) error {
	if req.BillingID == "" {
		return domain.ValidationError{Field: "billingID", Message: "billing ID is required"}
	}
	if req.Name == "" {
		return domain.ValidationError{Field: "name", Message: "item name is required"}
	}
	if req.Price <= 0 {
		return domain.ValidationError{Field: "price", Message: "price must be greater than 0"}
	}
	return nil
}

func (u *billingUseCase) validateCloseBillRequest(req CloseBillRequest) error {
	if req.BillingID == "" {
		return domain.ValidationError{Field: "billingID", Message: "billing ID is required"}
	}

	if req.Currency != "" {
		if req.Currency != "USD" && req.Currency != "GEL" {
			return domain.ValidationError{Field: "currency", Message: "currency must be USD or GEL"}
		}
	}

	return nil
}
