package transfer

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"encore.app/ledger/activity"
	"encore.app/ledger/model"
)

type transferState struct {
	pending   []*transfer
	transfers map[model.ReferenceID]*transfer
}

type transfer struct {
	referenceID      model.ReferenceID
	pendingID        model.PendingID
	cancelWorkflowID string
	status           transferStatus
	amount           model.TransferAmount
	expireAfter      time.Duration
}

type transferStatus uint8

const (
	StatusRequested transferStatus = iota
	StatusPending
	StatusCompleted
	StatusFailed
	StatusCancelled
)

func (s *transferState) handleAuthorize(ctx workflow.Context, authorize AuthorizeSignal, accountId model.AccountID, a activities) error {
	req := activity.PendingAuthorizeRequest{
		ID:              model.PendingID(idFromTime(ctx)),
		CreditAccountID: accountId,
		Amount:          authorize.Amount,
	}

	tr := &transfer{
		referenceID: authorize.ReferenceID,
		pendingID:   req.ID,
		status:      StatusRequested,
		amount:      authorize.Amount,
		expireAfter: authorize.ExpireAfter,
	}

	s.pending = append(s.pending, tr)
	s.transfers[tr.referenceID] = tr

	err := a.Authorize(ctx, req)
	if err != nil {
		tr.status = StatusFailed
		return err // TODO error handling
	}

	creq := CancelAuthorizationRequest{
		ReferenceID: authorize.ReferenceID,
		ExpireAfter: authorize.ExpireAfter,
	}

	childID := a.ScheduleCancelProcess(ctx, creq)

	tr.status = StatusPending
	tr.cancelWorkflowID = childID
	return nil
}

func (s *transferState) handleCancel(ctx workflow.Context, cancel CancelSignal, accountID model.AccountID, a activities) error {
	tr := s.transfers[cancel.ReferenceID]

	if tr.status != StatusPending && tr.status != StatusFailed {
		return nil
	}

	req := activity.CancelAuthorizeRequest{
		ID:        model.CancelID(idFromTime(ctx)),
		PendingID: tr.pendingID,
	}

	err := a.Cancel(ctx, req)
	if err != nil {
		// keep transfer in pending state
		return err
	}
	tr.status = StatusCancelled
	return nil
}

func (s *transferState) handlePresent(ctx workflow.Context, present PresentSignal, accountID model.AccountID, a activities) error {
	tr := findFirstMatching(s.pending, present.Amount)
	if tr == nil {
		return s.handleTransfer(ctx, present, accountID, a)
	}

	s.transfers[present.ReferenceID] = tr

	req := activity.CompleteAuthorizeRequest{
		ID:        model.TransferID(idFromTime(ctx)),
		PendingID: tr.pendingID,
		Amount:    tr.amount,
	}

	err := a.Complete(ctx, req)
	if err != nil {
		tr.status = StatusFailed
		return nil
	}

	tr.status = StatusCompleted

	if tr.cancelWorkflowID != "" {
		treq := TerminateCancelRequest{
			WorkflowID: tr.cancelWorkflowID,
		}

		err = a.TerminateCancelProcess(ctx, treq)
		if err != nil {
			// ignoring, as cancel handler cancels only if it sees pending or failed transfer
			return nil
		}
	}

	return nil
}

func (s *transferState) handleTransfer(ctx workflow.Context, present PresentSignal, accountID model.AccountID, a activities) error {
	tr := &transfer{
		referenceID:      present.ReferenceID,
		pendingID:        0,
		cancelWorkflowID: "",
		status:           StatusRequested,
		amount:           present.Amount,
		expireAfter:      0,
	}
	s.transfers[tr.referenceID] = tr

	req := activity.TransferRequest{
		ID:              model.TransferID(idFromTime(ctx)),
		CreditAccountID: accountID,
		Amount:          present.Amount,
	}

	err := a.Transfer(ctx, req)
	if err != nil {
		tr.status = StatusFailed
		return err
	}

	tr.status = StatusCompleted
	return nil
}

// idFromTime is only for toy implementation, it cannot guarantee uniqueness
func idFromTime(ctx workflow.Context) uint64 {
	now := workflow.Now(ctx)
	return uint64(now.UnixNano())
}

func findFirstMatching(pending []*transfer, amount model.TransferAmount) *transfer {
	for _, v := range pending {
		if v != nil && v.status == StatusPending && v.amount == amount {
			return v
		}
	}
	return nil
}
