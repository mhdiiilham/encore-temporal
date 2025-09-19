package usecases_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"encore.app/billing/domain"
	mock_domain "encore.app/billing/domain/mock"
	"encore.app/billing/usecases"
	mock_usecases "encore.app/billing/usecases/mock"
	mock_clock "encore.app/pkg/clock/mock"
	mock_generator "encore.app/pkg/generator/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

var mockTime = time.Date(2000, 12, 15, 0, 0, 0, 0, time.UTC)

type billingUseCaseTestSuite struct {
	suite.Suite

	mockController     *gomock.Controller
	mockRepository     *mock_domain.MockRepository
	mockWorkflowClient *mock_usecases.MockWorkflowClient
	mockIDGenerator    *mock_generator.MockIDProvider
	mockClock          *mock_clock.MockClock
}

func (suite *billingUseCaseTestSuite) SetupTest() {
	t := suite.T()

	suite.mockController = gomock.NewController(t)
	suite.mockRepository = mock_domain.NewMockRepository(suite.mockController)
	suite.mockWorkflowClient = mock_usecases.NewMockWorkflowClient(suite.mockController)
	suite.mockIDGenerator = mock_generator.NewMockIDProvider(suite.mockController)
	suite.mockClock = mock_clock.NewMockClock(suite.mockController)
}

func TestBillingUseCases(t *testing.T) {
	suite.Run(t, new(billingUseCaseTestSuite))
}

func (suite *billingUseCaseTestSuite) TestCreateBill() {
	mockBillingID := "Bill-billing-id"

	testCases := []struct {
		condition   string
		argument    usecases.CreateBillRequest
		expectedErr error
		expectedID  string
		doMock      func(ctx context.Context, mockWorkflow *mock_usecases.MockWorkflowClient, mockGenerator *mock_generator.MockIDProvider, mockClock *mock_clock.MockClock)
	}{
		{
			condition:   "validation failed: currency is empty",
			argument:    usecases.CreateBillRequest{Currency: ""},
			expectedErr: domain.ValidationError{Field: "currency", Message: "currency is required"},
			expectedID:  "",
			doMock: func(ctx context.Context, mockWorkflow *mock_usecases.MockWorkflowClient, mockGenerator *mock_generator.MockIDProvider, mockClock *mock_clock.MockClock) {
			},
		},
		{
			condition:   "validation failed: currency is invalid",
			argument:    usecases.CreateBillRequest{Currency: "IDR"},
			expectedErr: domain.ValidationError{Field: "currency", Message: "currency must be USD or GEL"},
			expectedID:  "",
			doMock: func(ctx context.Context, mockWorkflow *mock_usecases.MockWorkflowClient, mockGenerator *mock_generator.MockIDProvider, mockClock *mock_clock.MockClock) {
			},
		},
		{
			condition:   "success",
			argument:    usecases.CreateBillRequest{Currency: "USD"},
			expectedErr: nil,
			expectedID:  mockBillingID,
			doMock: func(ctx context.Context, mockWorkflow *mock_usecases.MockWorkflowClient, mockGenerator *mock_generator.MockIDProvider, mockClock *mock_clock.MockClock) {
				mockGenerator.EXPECT().GenerateBillingID("Bill").Return(mockBillingID).Times(1)
				mockClock.EXPECT().Now().Return(mockTime).Times(1)
				mockWorkflow.EXPECT().
					StartWorkflow(ctx, gomock.Any(), gomock.AssignableToTypeOf(&domain.Bill{
						BillingID: mockBillingID,
						Status:    domain.BillStatusOpen,
						Currency:  domain.CurrencyUSD,
						Total:     0,
						Items:     []domain.Item{},
						CreatedAt: mockTime,
					})).
					Return(nil)
			},
		},
		{
			condition:   "fail start workflow",
			argument:    usecases.CreateBillRequest{Currency: "USD"},
			expectedErr: fmt.Errorf("failed to start workflow: %w", errors.New("unexpected error")),
			expectedID:  "",
			doMock: func(ctx context.Context, mockWorkflow *mock_usecases.MockWorkflowClient, mockGenerator *mock_generator.MockIDProvider, mockClock *mock_clock.MockClock) {
				mockGenerator.EXPECT().GenerateBillingID("Bill").Return(mockBillingID).Times(1)
				mockClock.EXPECT().Now().Return(mockTime).Times(1)
				mockWorkflow.EXPECT().
					StartWorkflow(ctx, gomock.Any(), gomock.AssignableToTypeOf(&domain.Bill{})).
					Return(errors.New("unexpected error")).
					Times(1)
			},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.condition, func(t *testing.T) {
			uc := usecases.NewBillingUseCase(nil, suite.mockWorkflowClient, suite.mockIDGenerator, suite.mockClock)
			ctx := context.Background()
			assertion := assert.New(t)

			tc.doMock(ctx, suite.mockWorkflowClient, suite.mockIDGenerator, suite.mockClock)

			id, err := uc.CreateBill(ctx, tc.argument)
			assertion.Equal(tc.expectedID, id)
			assertion.Equal(tc.expectedErr, err)

		})
	}
}

func (suite *billingUseCaseTestSuite) TestGetBill() {
	mockCreatedAt := time.Now()
	mockClosedAt := mockCreatedAt.Add(7 * time.Hour)

	testCases := []struct {
		condition       string
		billingID       string
		expectedBilling domain.Bill
		expectedErr     error
		doMock          func(ctx context.Context, mockRepo *mock_domain.MockRepository, mockWorkflow *mock_usecases.MockWorkflowClient)
	}{
		{
			condition: "success get from databae",
			billingID: "mock-billing-id",
			expectedBilling: domain.Bill{
				ID:        1,
				BillingID: "mock-billing-id",
				Status:    domain.BillStatusClosed,
				Currency:  domain.CurrencyUSD,
				Total:     1000,
				Items: []domain.Item{
					{Name: "Sparkling", Price: 1000},
				},
				Conversion: domain.BillExchange{},
				CreatedAt:  mockCreatedAt,
				ClosedAt:   &mockClosedAt,
			},
			expectedErr: nil,
			doMock: func(ctx context.Context, mockRepo *mock_domain.MockRepository, mockWorkflow *mock_usecases.MockWorkflowClient) {
				mockWorkflow.EXPECT().QueryWorkflow(ctx, "mock-billing-id").Return(domain.Bill{}, errors.New("temporal error")).Times(1)

				mockRepo.EXPECT().GetBill(ctx, "mock-billing-id").Return(domain.Bill{
					ID:        1,
					BillingID: "mock-billing-id",
					Status:    domain.BillStatusClosed,
					Currency:  domain.CurrencyUSD,
					Total:     1000,
					Items: []domain.Item{
						{Name: "Sparkling", Price: 1000},
					},
					Conversion: domain.BillExchange{},
					CreatedAt:  mockCreatedAt,
					ClosedAt:   &mockClosedAt,
				}, nil).Times(1)
			},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.condition, func(t *testing.T) {
			uc := usecases.NewBillingUseCase(suite.mockRepository, suite.mockWorkflowClient, nil, nil)
			ctx := context.Background()
			assertion := assert.New(t)

			tc.doMock(ctx, suite.mockRepository, suite.mockWorkflowClient)

			actualBill, actualErr := uc.GetBill(ctx, tc.billingID)
			assertion.Equal(tc.expectedBilling, actualBill)
			assertion.Equal(tc.expectedErr, actualErr)
		})
	}

}

func (suite *billingUseCaseTestSuite) TestAddItem() {
	mockCreatedAt := time.Now().AddDate(0, 0, -1)
	mockClosedAt := mockCreatedAt.Add(1 * time.Hour)
	mockIdempotencyKey := "mock-idempotency"

	testCases := []struct {
		condition   string
		req         usecases.AddItemRequest
		expectedErr error
		doMock      func(ctx context.Context, mockRepo *mock_domain.MockRepository, mockWorkflow *mock_usecases.MockWorkflowClient, mockGenerator *mock_generator.MockIDProvider)
	}{
		{
			condition:   "success",
			req:         usecases.AddItemRequest{BillingID: "mock-billing-id", Name: "Sparkling", Price: 1000},
			expectedErr: nil,
			doMock: func(ctx context.Context, mockRepo *mock_domain.MockRepository, mockWorkflow *mock_usecases.MockWorkflowClient, mockGenerator *mock_generator.MockIDProvider) {
				mockGenerator.EXPECT().GenerateIdempotencyKey("idem", usecases.PayloadToBytes(usecases.AddItemRequest{BillingID: "mock-billing-id", Name: "Sparkling", Price: 1000})).Return(mockIdempotencyKey).Times(1)
				mockWorkflow.EXPECT().QueryWorkflow(ctx, "mock-billing-id").Return(domain.Bill{
					ID:        1,
					BillingID: "mock-billing-id",
					Status:    domain.BillStatusOpen,
					Currency:  domain.CurrencyUSD,
					Total:     1000,
					Items: []domain.Item{
						{Name: "Sparkling", Price: 1000},
					},
					Conversion: domain.BillExchange{},
					CreatedAt:  mockCreatedAt,
				}, nil).Times(1)

				mockWorkflow.EXPECT().SignalWorkflow(ctx, "mock-billing-id", domain.SignalAddLineItem, gomock.AssignableToTypeOf(&domain.Item{})).Return(nil).Times(1)
			},
		},
		{
			condition:   "bill is closed",
			req:         usecases.AddItemRequest{BillingID: "mock-billing-id", Name: "Sparkling", Price: 1000},
			expectedErr: domain.ErrBillClosed,
			doMock: func(ctx context.Context, mockRepo *mock_domain.MockRepository, mockWorkflow *mock_usecases.MockWorkflowClient, mockGenerator *mock_generator.MockIDProvider) {
				mockWorkflow.EXPECT().QueryWorkflow(ctx, "mock-billing-id").Return(domain.Bill{
					ID:        1,
					BillingID: "mock-billing-id",
					Status:    domain.BillStatusClosed,
					Currency:  domain.CurrencyUSD,
					Total:     1000,
					Items: []domain.Item{
						{Name: "Sparkling", Price: 1000},
					},
					Conversion: domain.BillExchange{},
					CreatedAt:  mockCreatedAt,
					ClosedAt:   &mockClosedAt,
				}, nil).Times(1)
			},
		},
		{
			condition:   "bill not found",
			req:         usecases.AddItemRequest{BillingID: "mock-billing-id", Name: "Sparkling", Price: 1000},
			expectedErr: domain.ErrBillNotFound,
			doMock: func(ctx context.Context, mockRepo *mock_domain.MockRepository, mockWorkflow *mock_usecases.MockWorkflowClient, mockGenerator *mock_generator.MockIDProvider) {
				mockWorkflow.EXPECT().QueryWorkflow(ctx, "mock-billing-id").Return(domain.Bill{}, domain.ErrBillNotFound).Times(1)
				mockRepo.EXPECT().GetBill(ctx, "mock-billing-id").Return(domain.Bill{}, domain.ErrBillNotFound).Times(1)
			},
		},
		{
			condition:   "billing id is empty",
			req:         usecases.AddItemRequest{Name: "Sparkling", Price: 1000},
			expectedErr: domain.ValidationError{Field: "billingID", Message: "billing ID is required"},
			doMock: func(ctx context.Context, mockRepo *mock_domain.MockRepository, mockWorkflow *mock_usecases.MockWorkflowClient, mockGenerator *mock_generator.MockIDProvider) {
			},
		},
		{
			condition:   "item name is empty",
			req:         usecases.AddItemRequest{BillingID: "Sparkling", Price: 1000},
			expectedErr: domain.ValidationError{Field: "name", Message: "item name is required"},
			doMock: func(ctx context.Context, mockRepo *mock_domain.MockRepository, mockWorkflow *mock_usecases.MockWorkflowClient, mockGenerator *mock_generator.MockIDProvider) {
			},
		},
		{
			condition:   "item price is empty",
			req:         usecases.AddItemRequest{BillingID: "mock-billing-id", Name: "Sparkling"},
			expectedErr: domain.ValidationError{Field: "price", Message: "price must be greater than 0"},
			doMock: func(ctx context.Context, mockRepo *mock_domain.MockRepository, mockWorkflow *mock_usecases.MockWorkflowClient, mockGenerator *mock_generator.MockIDProvider) {
			},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.condition, func(t *testing.T) {
			uc := usecases.NewBillingUseCase(suite.mockRepository, suite.mockWorkflowClient, suite.mockIDGenerator, nil)
			ctx := context.Background()
			assertion := assert.New(t)

			tc.doMock(ctx, suite.mockRepository, suite.mockWorkflowClient, suite.mockIDGenerator)
			err := uc.AddItem(ctx, tc.req)
			assertion.Equal(tc.expectedErr, err)
		})
	}
}

func (suite *billingUseCaseTestSuite) TestCloseBilling() {
	testCases := []struct {
		condition       string
		req             usecases.CloseBillRequest
		expectedErr     error
		expectedBilling domain.Bill
		doMock          func(ctx context.Context, mockRepo *mock_domain.MockRepository, mockWorkflow *mock_usecases.MockWorkflowClient, mockClock *mock_clock.MockClock)
	}{
		{
			condition:   "billing id empty",
			req:         usecases.CloseBillRequest{BillingID: "", Currency: "USD"},
			expectedErr: domain.ValidationError{Field: "billingID", Message: "billing ID is required"},
			doMock: func(ctx context.Context, mockRepo *mock_domain.MockRepository, mockWorkflow *mock_usecases.MockWorkflowClient, mockClock *mock_clock.MockClock) {
			},
		},
		{
			condition:   "invalid currency",
			req:         usecases.CloseBillRequest{BillingID: "mock-billing-id", Currency: "IDR"},
			expectedErr: domain.ValidationError{Field: "currency", Message: "currency must be USD or GEL"},
			doMock: func(ctx context.Context, mockRepo *mock_domain.MockRepository, mockWorkflow *mock_usecases.MockWorkflowClient, mockClock *mock_clock.MockClock) {
			},
		},
		{
			condition:   "success without conversion",
			req:         usecases.CloseBillRequest{BillingID: "mock-billing-id", Currency: ""},
			expectedErr: nil,
			doMock: func(ctx context.Context, mockRepo *mock_domain.MockRepository, mockWorkflow *mock_usecases.MockWorkflowClient, mockClock *mock_clock.MockClock) {
				suite.mockClock.EXPECT().Now().Return(mockTime).Times(1)
				mockWorkflow.EXPECT().QueryWorkflow(ctx, "mock-billing-id").Return(
					domain.Bill{
						ID:        1,
						BillingID: "mock-billing-id",
						Status:    domain.BillStatusOpen,
						Currency:  domain.CurrencyUSD,
						Total:     1000,
						Items: []domain.Item{
							{Name: "Sparkling", Price: 1000},
						},
						Conversion: domain.BillExchange{},
					},
					nil,
				).Times(1)

				mockWorkflow.EXPECT().SignalWorkflow(ctx, "mock-billing-id", domain.SignalCloseBill, gomock.AssignableToTypeOf(domain.BillExchange{})).Return(nil).Times(1)
			},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.condition, func(t *testing.T) {
			uc := usecases.NewBillingUseCase(suite.mockRepository, suite.mockWorkflowClient, nil, suite.mockClock)
			ctx := context.Background()
			assertion := assert.New(t)

			tc.doMock(ctx, suite.mockRepository, suite.mockWorkflowClient, suite.mockClock)
			_, err := uc.CloseBill(ctx, tc.req)
			assertion.Equal(tc.expectedErr, err)
		})
	}
}

func (suite *billingUseCaseTestSuite) TearDownTest() {
	suite.mockController.Finish()
}
