package infrastructure

import (
	"context"
	"fmt"
	"strings"

	"encore.app/billing/domain"
	"encore.app/billing/usecases"
	"go.temporal.io/sdk/client"
)

// temporalWorkflowClient implements WorkflowClient interface
type temporalWorkflowClient struct {
	client    client.Client
	workflows *Workflows
}

// NewTemporalWorkflowClient creates a new Temporal workflow client
func NewTemporalWorkflowClient(c client.Client, workflows *Workflows) usecases.WorkflowClient {
	return &temporalWorkflowClient{client: c, workflows: workflows}
}

// StartWorkflow starts a new workflow
func (t *temporalWorkflowClient) StartWorkflow(ctx context.Context, workflowID string, bill *domain.Bill) error {
	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "billing-task-queue",
	}

	_, err := t.client.ExecuteWorkflow(ctx, options, t.workflows.BillingWorkflow, bill)
	if err != nil {
		return fmt.Errorf("failed to start workflow: %w", err)
	}

	return nil
}

// QueryWorkflow queries a workflow
func (t *temporalWorkflowClient) QueryWorkflow(ctx context.Context, workflowID string) (domain.Bill, error) {
	resp, err := t.client.QueryWorkflow(ctx, workflowID, "", domain.QueryTypeGetBilling)
	if err != nil {
		if strings.Contains(err.Error(), "workflow not found") {
			return domain.Bill{}, domain.ErrWorkflowNotFound
		}
		return domain.Bill{}, fmt.Errorf("failed to query workflow: %w", err)
	}

	var bill domain.Bill
	if err := resp.Get(&bill); err != nil {
		return domain.Bill{}, fmt.Errorf("failed to parse workflow result: %w", err)
	}

	return bill, nil
}

// SignalWorkflow sends a signal to a workflow
func (t *temporalWorkflowClient) SignalWorkflow(ctx context.Context, workflowID string, signal string, data interface{}) error {
	err := t.client.SignalWorkflow(ctx, workflowID, "", signal, data)
	if err != nil {
		return fmt.Errorf("failed to signal workflow: %w", err)
	}
	return nil
}

// IsWorkflowRunning checks if a workflow is running
func (t *temporalWorkflowClient) IsWorkflowRunning(ctx context.Context, workflowID string) (bool, error) {
	_, err := t.client.QueryWorkflow(ctx, workflowID, "", "getBill")
	if err != nil {
		if strings.Contains(err.Error(), "workflow not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
