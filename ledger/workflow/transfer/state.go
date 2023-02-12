package transfer

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"encore.app/ledger/activity"
	"encore.app/ledger/model"
)

type transferState struct {
	pending   []*transfer
	transfers map[string]*transfer
}

type transfer struct {
	refID            string
	pendingID        model.PendingID
	cancelWorkflowID string
	status           transferStatus
	amount           model.TransferAmount
	expireAfter      time.Duration
}

type transferStatus uint8

const (
	TransferRequested transferStatus = iota
	TransferPending
	TransferCompleted
	TransferFailed
	TransferCancelled
)

func (s *transferState) handleAuthorize(ctx workflow.Context, authorize AuthorizeSignal, accountId model.AccountID, a transferActivities) error {
	req := activity.PendingAuthorizeRequest{
		ID:              model.PendingID(idFromTime(ctx)),
		CreditAccountID: accountId,
		Amount:          authorize.Amount,
	}

	tr := &transfer{
		refID:       authorize.ReferenceID,
		pendingID:   req.ID,
		status:      TransferRequested,
		amount:      authorize.Amount,
		expireAfter: authorize.ExpireAfter,
	}

	s.pending = append(s.pending, tr)
	s.transfers[tr.refID] = tr

	err := a.Authorize(ctx, req)
	if err != nil {
		tr.status = TransferFailed
		return err // TODO error handling
	}

	creq := CancelAuthorizationRequest{
		ReferenceID: authorize.ReferenceID,
		ExpireAfter: authorize.ExpireAfter,
	}

	childID := a.ScheduleCancel(ctx, creq)

	tr.status = TransferPending
	tr.cancelWorkflowID = childID
	return nil
}

func (s *transferState) handleCancel(ctx workflow.Context, cancel CancelSignal, accountID model.AccountID, a transferActivities) error {
	tr := s.transfers[cancel.ReferenceID]

	if tr.status != TransferPending {
		// if somebody completed the transfer, do nothing
		return nil
	}

	req := activity.CancelAuthorizeRequest{
		ID:              model.CancelID(idFromTime(ctx)),
		PendingID:       tr.pendingID,
		CreditAccountID: accountID,
	}

	err := a.Cancel(ctx, req)
	if err != nil {
		// keep transfer in pending state
		return err
	}
	tr.status = TransferCancelled
	return nil
}

// idFromTime is only for toy implementation, it cannot guarantee uniqueness
func idFromTime(ctx workflow.Context) uint64 {
	now := workflow.Now(ctx)
	return uint64(now.UnixNano())
}
