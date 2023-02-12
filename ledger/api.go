package ledger

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"

	"encore.app/ledger/model"
	"encore.app/ledger/workflow/transfer"
)

// GetBalance retrieves the actual balance of the account
//
//encore:api public method=GET path=/balance/:accountID
func (s *Service) GetBalance(ctx context.Context, accountID uint64) (*model.BalanceResponse, error) {
	result, err := s.activities.GetBalance(ctx, model.AccountID(accountID))
	if err != nil {
		return nil, err
	}

	return &model.BalanceResponse{
		AccountID: model.AccountID(accountID),
		Balance:   result,
	}, nil
}

// GetAvailableBalance retrieves the available balance of the account to spend
//
//encore:api public method=GET path=/available-balance/:accountID
func (s *Service) GetAvailableBalance(ctx context.Context, accountID uint64) (*model.AvailableBalanceResponse, error) {
	result, err := s.activities.GetAvailableBalance(ctx, model.AccountID(accountID))
	if err != nil {
		return nil, err
	}

	return &model.AvailableBalanceResponse{
		AccountID:        model.AccountID(accountID),
		AvailableBalance: result,
	}, nil
}

//encore:api public method=POST path=/authorize/:accountID/:amount
func (s *Service) Authorize(ctx context.Context, accountID uint64, amount uint64) (*model.AuthorizeResponse, error) {
	signalArg := transfer.AuthorizeSignal{
		ReferenceID: model.NewReferenceID(),
		Amount:      model.TransferAmount(amount),
		ExpireAfter: 10 * time.Second,
	}
	workflowName := fmt.Sprintf("%s-%d", transferWorkflowName, accountID)

	err := s.client.SignalWorkflow(ctx, workflowName, "", transfer.AuthorizeSignalName, signalArg)
	if err != nil {
		switch err.(type) {
		case *serviceerror.NotFound:
			_, err = s.client.SignalWithStartWorkflow(ctx, workflowName, transfer.AuthorizeSignalName, signalArg, client.StartWorkflowOptions{
				ID:        workflowName,
				TaskQueue: paveTaskQueue,
			}, transfer.Workflow, model.AccountID(accountID), nil)
		}
	}

	if err != nil {
		return nil, err
	}

	return &model.AuthorizeResponse{
		ReferenceID: signalArg.ReferenceID,
	}, nil
}

//encore:api public method=POST path=/present/:accountID/:amount
func (s *Service) Present(ctx context.Context, accountID uint64, amount uint64) (*model.PresentResponse, error) {
	signalArg := transfer.PresentSignal{
		ReferenceID: model.NewReferenceID(),
		Amount:      model.TransferAmount(amount),
	}
	workflowName := fmt.Sprintf("%s-%d", transferWorkflowName, accountID)

	err := s.client.SignalWorkflow(ctx, workflowName, "", transfer.PresentSignalName, signalArg)
	if err != nil {
		switch err.(type) {
		case *serviceerror.NotFound:
			_, err = s.client.SignalWithStartWorkflow(ctx, workflowName, transfer.PresentSignalName, signalArg, client.StartWorkflowOptions{
				ID:        workflowName,
				TaskQueue: paveTaskQueue,
			}, transfer.Workflow, model.AccountID(accountID), nil)
		}
	}

	if err != nil {
		return nil, err
	}

	return &model.PresentResponse{
		ReferenceID: signalArg.ReferenceID,
	}, err
}
