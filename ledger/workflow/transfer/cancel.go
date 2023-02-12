package transfer

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"encore.app/ledger/model"
)

type CancelAuthorizationRequest struct {
	ReferenceID model.ReferenceID
	ExpireAfter time.Duration
}

type TerminateCancelRequest struct {
	WorkflowID string
}

func CancelWorkflow(ctx workflow.Context, req CancelAuthorizationRequest) error {
	err := workflow.Sleep(ctx, req.ExpireAfter)
	if err != nil {
		return err
	}
	info := workflow.GetInfo(ctx)
	return workflow.SignalExternalWorkflow(ctx, info.ParentWorkflowExecution.ID, "", CancelSignalName, CancelSignal{
		ReferenceID: req.ReferenceID,
	}).Get(ctx, nil)
}
