package transfer

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"encore.app/ledger/activity"
)

type activities interface {
	Authorize(ctx workflow.Context, req activity.PendingAuthorizeRequest) error
	Cancel(ctx workflow.Context, req activity.CancelAuthorizeRequest) error
	Transfer(ctx workflow.Context, req activity.TransferRequest) error
	Complete(ctx workflow.Context, req activity.CompleteAuthorizeRequest) error

	ScheduleCancelProcess(ctx workflow.Context, req CancelAuthorizationRequest) string
	TerminateCancelProcess(ctx workflow.Context, req TerminateCancelRequest) error
}

type transferActivities struct{}

func (a transferActivities) Authorize(ctx workflow.Context, request activity.PendingAuthorizeRequest) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 5,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        time.Second,
			BackoffCoefficient:     2.0,
			MaximumInterval:        100 * time.Second,
			MaximumAttempts:        10,
			NonRetryableErrorTypes: []string{},
		},
	})

	var tba *activity.TigerBeetleActivities
	err := workflow.ExecuteActivity(ctx, tba.Authorize, request).Get(ctx, nil)
	if err != nil {
		return nil // TODO handle errors
	}

	return nil
}

func (a transferActivities) Cancel(ctx workflow.Context, request activity.CancelAuthorizeRequest) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 5,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        time.Second,
			BackoffCoefficient:     2.0,
			MaximumInterval:        100 * time.Second,
			MaximumAttempts:        10,
			NonRetryableErrorTypes: []string{},
		},
	})

	var tba *activity.TigerBeetleActivities
	err := workflow.ExecuteActivity(ctx, tba.Cancel, request).Get(ctx, nil)
	if err != nil {
		return nil // TODO handle errors
	}

	return nil
}

func (a transferActivities) Transfer(ctx workflow.Context, request activity.TransferRequest) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 5,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        time.Second,
			BackoffCoefficient:     2.0,
			MaximumInterval:        100 * time.Second,
			MaximumAttempts:        10,
			NonRetryableErrorTypes: []string{},
		},
	})

	var tba *activity.TigerBeetleActivities
	err := workflow.ExecuteActivity(ctx, tba.Transfer, request).Get(ctx, nil)
	if err != nil {
		return nil // TODO handle errors
	}

	return nil
}

func (a transferActivities) Complete(ctx workflow.Context, request activity.CompleteAuthorizeRequest) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 5,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        time.Second,
			BackoffCoefficient:     2.0,
			MaximumInterval:        100 * time.Second,
			MaximumAttempts:        10,
			NonRetryableErrorTypes: []string{},
		},
	})

	var tba *activity.TigerBeetleActivities
	err := workflow.ExecuteActivity(ctx, tba.Complete, request).Get(ctx, nil)
	if err != nil {
		return nil // TODO handle errors
	}

	return nil
}

func (a transferActivities) ScheduleCancelProcess(ctx workflow.Context, req CancelAuthorizationRequest) string {
	childID := fmt.Sprintf("cancel-%s", req.ReferenceID)
	ctx = workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: childID,
	})
	workflow.ExecuteChildWorkflow(ctx, CancelWorkflow, req)

	return childID
}

func (a transferActivities) TerminateCancelProcess(ctx workflow.Context, req TerminateCancelRequest) error {
	return workflow.RequestCancelExternalWorkflow(ctx, req.WorkflowID, "").Get(ctx, nil)
}
