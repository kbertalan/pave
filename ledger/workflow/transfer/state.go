package transfer

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"encore.app/ledger/model"
	"encore.app/ledger/tb"
)

type State struct {
	Pending   []*Transfer
	Transfers map[model.ReferenceID]*Transfer
}

type Transfer struct {
	ReferenceID      model.ReferenceID
	PendingID        model.PendingID
	CancelWorkflowID string
	Status           Status
	Amount           model.TransferAmount
	ExpireAfter      time.Duration
}

type Status uint8

const (
	StatusRequested Status = iota
	StatusPending
	StatusCompleted
	StatusFailed
	StatusCancelled
)

func (s *State) handleAuthorize(ctx workflow.Context, authorize AuthorizeSignal, accountId model.AccountID, a activities) error {
	req := tb.PendingAuthorizeRequest{
		ID:              model.PendingID(idFromTime(ctx)),
		CreditAccountID: accountId,
		Amount:          authorize.Amount,
	}

	tr := &Transfer{
		ReferenceID: authorize.ReferenceID,
		PendingID:   req.ID,
		Status:      StatusRequested,
		Amount:      authorize.Amount,
		ExpireAfter: authorize.ExpireAfter,
	}

	s.Pending = append(s.Pending, tr)
	s.Transfers[tr.ReferenceID] = tr

	err := a.Authorize(ctx, req)
	if err != nil {
		tr.Status = StatusFailed
		return err // TODO error handling
	}

	creq := CancelAuthorizationRequest{
		ReferenceID: authorize.ReferenceID,
		ExpireAfter: authorize.ExpireAfter,
	}

	childID := a.ScheduleCancelProcess(ctx, creq)

	tr.Status = StatusPending
	tr.CancelWorkflowID = childID
	return nil
}

func (s *State) handleCancel(ctx workflow.Context, cancel CancelSignal, accountID model.AccountID, a activities) error {
	tr := s.Transfers[cancel.ReferenceID]

	if tr.Status != StatusPending && tr.Status != StatusFailed {
		return nil
	}

	req := tb.CancelAuthorizeRequest{
		ID:        model.CancelID(idFromTime(ctx)),
		PendingID: tr.PendingID,
	}

	err := a.Cancel(ctx, req)
	if err != nil {
		// keep Transfer in Pending state
		return err
	}
	tr.Status = StatusCancelled
	return nil
}

func (s *State) handlePresent(ctx workflow.Context, present PresentSignal, accountID model.AccountID, a activities) error {
	tr := findFirstMatching(s.Pending, present.Amount)
	if tr == nil {
		return s.handleTransfer(ctx, present, accountID, a)
	}

	s.Transfers[present.ReferenceID] = tr

	req := tb.CompleteAuthorizeRequest{
		ID:        model.TransferID(idFromTime(ctx)),
		PendingID: tr.PendingID,
		Amount:    tr.Amount,
	}

	err := a.Complete(ctx, req)
	if err != nil {
		tr.Status = StatusFailed
		return nil
	}

	tr.Status = StatusCompleted

	if tr.CancelWorkflowID != "" {
		treq := TerminateCancelRequest{
			WorkflowID: tr.CancelWorkflowID,
		}

		err = a.TerminateCancelProcess(ctx, treq)
		if err != nil {
			// ignoring, as cancel handler cancels only if it sees Pending or failed Transfer
			return nil
		}
	}

	return nil
}

func (s *State) handleTransfer(ctx workflow.Context, present PresentSignal, accountID model.AccountID, a activities) error {
	tr := &Transfer{
		ReferenceID:      present.ReferenceID,
		PendingID:        0,
		CancelWorkflowID: "",
		Status:           StatusRequested,
		Amount:           present.Amount,
		ExpireAfter:      0,
	}
	s.Transfers[tr.ReferenceID] = tr

	req := tb.TransferRequest{
		ID:              model.TransferID(idFromTime(ctx)),
		CreditAccountID: accountID,
		Amount:          present.Amount,
	}

	err := a.Transfer(ctx, req)
	if err != nil {
		tr.Status = StatusFailed
		return err
	}

	tr.Status = StatusCompleted
	return nil
}

// idFromTime is only for toy implementation, it cannot guarantee uniqueness
func idFromTime(ctx workflow.Context) uint64 {
	now := workflow.Now(ctx)
	return uint64(now.UnixNano())
}

func findFirstMatching(pending []*Transfer, amount model.TransferAmount) *Transfer {
	for _, v := range pending {
		if v != nil && v.Status == StatusPending && v.Amount == amount {
			return v
		}
	}
	return nil
}
