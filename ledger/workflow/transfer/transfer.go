package transfer

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"encore.app/ledger/model"
)

const (
	AuthorizeSignalName = "authorize"
	CancelSignalName    = "cancel"

	PresentSignalName = "present"
)

type AuthorizeSignal struct {
	ReferenceID model.ReferenceID
	Amount      model.TransferAmount
	ExpireAfter time.Duration
}

type CancelSignal struct {
	ReferenceID model.ReferenceID
}

type PresentSignal struct {
	ReferenceID model.ReferenceID
	Amount      model.TransferAmount
}

func Workflow(ctx workflow.Context, accountId model.AccountID, state *transferState) error {
	handledSignalsCount := 0
	if state == nil {
		state = &transferState{
			pending:   make([]*transfer, 0, 10),
			transfers: make(map[model.ReferenceID]*transfer),
		}
	}

	var triggeredSignalName string
	var authorize AuthorizeSignal
	authorizeChan := workflow.GetSignalChannel(ctx, AuthorizeSignalName)

	var cancel CancelSignal
	cancelChan := workflow.GetSignalChannel(ctx, CancelSignalName)

	var present PresentSignal
	presentChan := workflow.GetSignalChannel(ctx, PresentSignalName)

	selector := workflow.NewSelector(ctx)
	selector.AddReceive(authorizeChan, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &authorize)
		triggeredSignalName = AuthorizeSignalName
	})
	selector.AddReceive(cancelChan, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &cancel)
		triggeredSignalName = CancelSignalName
	})
	selector.AddReceive(presentChan, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &present)
		triggeredSignalName = PresentSignalName
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
		case PresentSignalName:
			err = state.handlePresent(ctx, present, accountId, transferActivities{})
		}

		if err != nil {
			return err
		}
		handledSignalsCount++

		if !selector.HasPending() && handledSignalsCount >= 20 {
			break
		}
	}

	return workflow.NewContinueAsNewError(ctx, Workflow, accountId, state)
}
