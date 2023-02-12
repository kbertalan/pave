package workflow

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"encore.app/ledger/activity"
	"encore.app/ledger/model"
)

const (
	AuthorizeSignalName = "authorize"
)

type AuthorizeSignal struct {
	Amount      model.TransferAmount
	ExpireAfter time.Duration
}

func Transfer(ctx workflow.Context, accountId model.AccountID) error {
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 5,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	var pendingAuthorizations []activity.PendingAuthorizeRequest
	handledSignalsCount := 0

	var authorize AuthorizeSignal
	authorizeChan := workflow.GetSignalChannel(ctx, AuthorizeSignalName)

	selector := workflow.NewSelector(ctx)
	selector.AddReceive(authorizeChan, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &authorize)
	})

	for {
		selector.Select(ctx)

		var err error
		switch {
		case authorize.ExpireAfter != 0:
			pendingAuthorizations, err = handleAuthorize(ctx, authorize, accountId, pendingAuthorizations)
		}

		if err != nil {
			return err
		}
		handledSignalsCount++

		if !selector.HasPending() && handledSignalsCount >= 20 {
			break
		}
	}

	return workflow.NewContinueAsNewError(ctx, Transfer, accountId)
}

func handleAuthorize(ctx workflow.Context, authorize AuthorizeSignal, accountId model.AccountID, pending []activity.PendingAuthorizeRequest) ([]activity.PendingAuthorizeRequest, error) {
	req := activity.PendingAuthorizeRequest{
		ID:              model.PendingID(idFromTime(ctx)),
		CreditAccountID: accountId,
		Amount:          authorize.Amount,
	}

	var tba *activity.TigerBeetleActivities
	err := workflow.ExecuteActivity(ctx, tba.Authorize, req).Get(ctx, nil)
	if err != nil {
		return pending, nil // TODO handle error cases
	}

	pending = append(pending, req)
	return pending, nil
}

func idFromTime(ctx workflow.Context) uint64 {
	time := workflow.Now(ctx)
	return uint64(time.UnixNano())
}
