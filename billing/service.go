package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"encore.app/billing/domain"
	"encore.app/billing/infrastructure"
	"encore.app/billing/usecases"
	"encore.app/pkg/clock"
	"encore.app/pkg/currency"
	"encore.app/pkg/generator"
	"encore.app/pkg/temporalclient"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// Service is the Encore service that wraps a Temporal client,
// providing methods to start workflows and interact with billing-related activities.
//
//encore:service
type Service struct {
	useCase usecases.BillingUseCase
	client  client.Client
	worker  worker.Worker
}

func initService() (*Service, error) {
	c, err := temporalclient.GetTemporalClient(client.Options{})
	if err != nil {
		return nil, err
	}

	idGenerator := generator.NewIDGenerator(time.Now().UnixNano())
	clock := clock.RealClock{}

	repository := infrastructure.NewRepository(billingdb)
	billingActivities := infrastructure.NewBillingActivity(repository)
	workflows := infrastructure.NewTemporalWorkflows(billingActivities)

	temporalClient := infrastructure.NewTemporalWorkflowClient(c, workflows)
	billingUseCase := usecases.NewBillingUseCase(repository, temporalClient, idGenerator, clock)

	rlog.Info("starting temporal worker")
	w := worker.New(c, domain.TemporalQueueName, worker.Options{})
	w.RegisterWorkflow(workflows.BillingWorkflow)
	w.RegisterActivity(billingActivities.UpsertBillingToDBActivity)
	w.RegisterActivity(billingActivities.SetBillingToCloseActivity)
	w.RegisterActivity(billingActivities.InsertLineItemActivity)
	w.RegisterActivity(billingActivities.InsertBillExchangeActivity)

	if err := w.Start(); err != nil {
		c.Close()
		return nil, fmt.Errorf("failed to start temporal worker: %v", err)
	}

	rlog.Info("success running temporal workder")
	return &Service{
		useCase: billingUseCase,
		worker:  w,
		client:  c,
	}, nil
}

// GetBill fetches the current state of a Bill by its ID.
// For open bills, it queries the running Temporal workflow (fastest).
// For closed bills, it queries the database directly.
//
//encore:api public method=GET path=/api/v1/bills/:id
func (s *Service) GetBill(ctx context.Context, id string) (*GetBillResponse, error) {
	bill, err := s.useCase.GetBill(ctx, id)
	if err != nil {
		var domainValidationErr domain.ValidationError
		if errors.As(err, &domainValidationErr) {
			return nil, errs.WrapCode(err, errs.InvalidArgument, err.Error())
		}

		return nil, errs.WrapCode(err, errs.Internal, "internal server error")
	}

	return &GetBillResponse{
		Bill: fromDomainBillToBillReponse(bill),
	}, nil
}

// AddItem add a new item to a running bill workflow
// Assumes the workflow is still running for active operations
//
//encore:api public method=POST path=/api/v1/bills/:id/items
func (s *Service) AddItem(ctx context.Context, id string, req *AddItemRequest) (*AddItemResponse, error) {
	bill, err := s.useCase.AddItem(ctx, usecases.AddItemRequest{
		BillingID: id,
		Name:      req.Name,
		Price:     req.Price,
	})

	if err != nil {
		var domainValidationErr domain.ValidationError
		if errors.As(err, &domainValidationErr) {
			return nil, errs.WrapCode(err, errs.InvalidArgument, err.Error())
		}

		return nil, errs.WrapCode(err, errs.Internal, "internal server error")
	}

	return &AddItemResponse{
		CurrentBill: fromDomainBillToBillReponse(bill),
	}, nil
}

// CloseBillingByID close an running bill workflow and return its total
// Assumes the workflow is still running for active operations
//
//encore:api public method=POST path=/api/v1/bills/:id
func (s *Service) CloseBillingByID(ctx context.Context, id string, req *CloseBillingRequest) (*CloseBillingResponse, error) {
	finalBill, err := s.useCase.CloseBill(ctx, usecases.CloseBillRequest{BillingID: id, Currency: req.Currency})
	if err != nil {
		var domainValidationErr domain.ValidationError
		if errors.As(err, &domainValidationErr) {
			return nil, errs.WrapCode(err, errs.InvalidArgument, err.Error())
		}

		return nil, errs.WrapCode(err, errs.Internal, "internal server error")
	}

	return &CloseBillingResponse{
		OriginalCurrencyTotal: Amount{
			Currency:        string(finalBill.Currency),
			Amount:          finalBill.Total,
			FormattedAmount: currency.FormatString(string(finalBill.Currency), finalBill.Total),
		},
		ConvertedCurrencyTotal: Amount{
			Currency:        string(finalBill.Conversion.TargetCurrency),
			Amount:          finalBill.Conversion.Total,
			FormattedAmount: currency.FormatString(string(finalBill.Conversion.TargetCurrency), finalBill.Conversion.Total),
		},
	}, nil
}

// OpenBilling handle open new biling by executing new workflows
//
//encore:api public method=POST path=/api/v1/bills
func (s *Service) OpenBilling(ctx context.Context, req *OpenBillingRequest) (*OpenBillingResponse, error) {
	billindID, err := s.useCase.CreateBill(ctx, usecases.CreateBillRequest{
		Currency: req.Currency,
	})
	if err != nil {
		var domainValidationErr domain.ValidationError
		if errors.As(err, &domainValidationErr) {
			return nil, errs.WrapCode(err, errs.InvalidArgument, err.Error())
		}

		return nil, errs.WrapCode(err, errs.Internal, "internal server error")
	}

	return &OpenBillingResponse{
		BillingID: billindID,
		Currency:  req.Currency,
	}, nil

}

// Shutdown hanlde graceful shutdown.
func (s *Service) Shutdown(force context.Context) {
	s.client.Close()
	s.worker.Stop()
}
