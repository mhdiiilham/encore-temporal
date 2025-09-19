package infrastructure

import (
	"time"

	"encore.app/billing/domain"
	"encore.dev/rlog"
	"github.com/mitchellh/mapstructure"
	"go.temporal.io/sdk/workflow"
)

type Workflows struct {
	billingActivies domain.BillingActivities
}

func NewTemporalWorkflows(billingActivities domain.BillingActivities) *Workflows {
	return &Workflows{
		billingActivies: billingActivities,
	}
}

// BillingWorkflow is a Temporal workflow that manages the lifecycle of a Bill.
// It handles incoming signals to add line items or close the bill, updates the
// database via activities, calculates totals, and performs currency conversion
// when the bill is closed.
func (w *Workflows) BillingWorkflow(ctx workflow.Context, state *domain.Bill) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("starting billing workflows", "id", state.ID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	if err := workflow.ExecuteActivity(ctx, w.billingActivies.UpsertBillingToDBActivity, state).Get(ctx, nil); err != nil {
		rlog.Error("failed to execute upsertBillingToDB", "err", err)
		return err
	}

	if err := workflow.SetQueryHandler(ctx, domain.QueryTypeGetBilling, func() (domain.Bill, error) {
		state.Total = state.GetTotal()
		return *state, nil
	}); err != nil {
		logger.Info("SetQueryHandler failed.", "Error", err)
		return err
	}

	addItemLineCh := workflow.GetSignalChannel(ctx, domain.SignalAddLineItem)
	closeBillCh := workflow.GetSignalChannel(ctx, domain.SignalCloseBill)

	itemQueue := make([]domain.Item, 0)
	closeRequested := false
	var currencyExchangeReq domain.BillExchange

	for {
		if state.Status == domain.BillStatusClosed {
			rlog.Info("bill is already closed, ignoring all signals", "workflow_id", state.BillingID)
			break
		}

		selector := workflow.NewSelector(ctx)

		selector.AddReceive(addItemLineCh, func(c workflow.ReceiveChannel, _ bool) {
			var signal any
			c.Receive(ctx, &signal)

			var message domain.Item
			if err := mapstructure.Decode(signal, &message); err != nil {
				rlog.Error("invalid signal type", "err", err)
				return
			}

			if state.Status == domain.BillStatusClosed {
				rlog.Warn("attempted to add item to closed bill", "workflow_id", state.BillingID, "item", message.Name)
				return
			}

			rlog.Info("received line item signal", "workflow_id", state.BillingID)
			itemQueue = append(itemQueue, message)
		})

		selector.AddReceive(closeBillCh, func(c workflow.ReceiveChannel, _ bool) {
			var signal any
			c.Receive(ctx, &signal)

			var message domain.BillExchange
			if err := mapstructure.Decode(signal, &message); err != nil {
				rlog.Error("invalid signal type", "err", err)
				return
			}

			if state.Status == domain.BillStatusClosed {
				rlog.Warn("attempted to close already closed bill", "workflow_id", state.BillingID)
				return
			}

			closeRequested = true
			currencyExchangeReq = message
		})

		selector.Select(ctx)

		for _, item := range itemQueue {
			err := workflow.ExecuteActivity(ctx, w.billingActivies.InsertLineItemActivity, item).Get(ctx, nil)
			if err != nil {
				rlog.Error("failed to persist item to db", "err", err)
				continue
			}
			state.AddItem(item)
		}
		itemQueue = itemQueue[:0]

		if closeRequested {
			closedAt := workflow.Now(ctx)
			state.Conversion = currencyExchangeReq
			state.Close(closedAt)

			err := workflow.ExecuteActivity(ctx, w.billingActivies.SetBillingToCloseActivity, state).Get(ctx, nil)
			if err != nil {
				rlog.Error("failed to set billing to close", "err", err, "bill", state)
				return err
			}

			if state.Conversion.TargetCurrency != "" {
				if err := workflow.ExecuteActivity(ctx, w.billingActivies.InsertBillExchangeActivity, state).Get(ctx, nil); err != nil {
					rlog.Error("failed to set conversion", "err", err, "bill", state, "currency", state.Conversion.TargetCurrency)
					return err
				}
			}

			break
		}
	}

	return nil
}
