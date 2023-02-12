package transfer

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"encore.app/ledger/model"
)

const (
	AuthorizeSignalName = "authorize"
	CancelSignalName    = "cancel"
)

type AuthorizeSignal struct {
	ReferenceID string
	Amount      model.TransferAmount
	ExpireAfter time.Duration
}

type CancelSignal struct {
	ReferenceID string
}

func TransferWorkflow(ctx workflow.Context, accountId model.AccountID, state *transferState) error {
	handledSignalsCount := 0
	if state == nil {
		state = &transferState{
			pending:   make([]*transfer, 0, 10),
			transfers: make(map[string]*transfer),
		}
	}

	var triggeredSignalName string
	var authorize AuthorizeSignal
	authorizeChan := workflow.GetSignalChannel(ctx, AuthorizeSignalName)

	var cancel CancelSignal
	cancelChan := workflow.GetSignalChannel(ctx, CancelSignalName)

	selector := workflow.NewSelector(ctx)
	selector.AddReceive(authorizeChan, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &authorize)
		triggeredSignalName = AuthorizeSignalName
	})
	selector.AddReceive(cancelChan, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &cancel)
		triggeredSignalName = CancelSignalName
	})

	for {
		triggeredSignalName = ""
		selector.Select(ctx)

		var err error
		switch triggeredSignalName {
		case AuthorizeSignalName:
			err = state.handleAuthorize(ctx, authorize, accountId, transferActivities{})
		case CancelSignalName:
			err = state.handleCancel(ctx, cancel, accountId, transferActivities{})
		}

		if err != nil {
			return err
		}
		handledSignalsCount++

		if !selector.HasPending() && handledSignalsCount >= 20 {
			break
		}
	}

	return workflow.NewContinueAsNewError(ctx, TransferWorkflow, accountId, state)
}
