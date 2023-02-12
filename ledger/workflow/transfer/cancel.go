package transfer

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

type CancelAuthorizationRequest struct {
	ReferenceID string
	ExpireAfter time.Duration
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
