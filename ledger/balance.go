package ledger

import (
	"context"
	"fmt"

	"encore.dev/rlog"
	"go.temporal.io/sdk/client"

	"encore.app/ledger/model"
	"encore.app/ledger/workflow/balance"
)

// GetBalance retrieves the actual balance of the account
//
//encore:api public method=GET path=/balance/:accountId
func (s *Service) GetBalance(ctx context.Context, accountId uint64) (*model.BalanceResponse, error) {
	options := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("get-balance-%d", accountId),
		TaskQueue: paveTaskQueue,
	}
	we, err := s.client.ExecuteWorkflow(ctx, options, balance.GetBalance, model.AccountID(accountId))
	if err != nil {
		return nil, err
	}
	rlog.Info("started workflow", "id", we.GetID(), "run_id", we.GetRunID())

	var br *model.BalanceResponse
	err = we.Get(ctx, &br)
	if err != nil {
		return nil, err
	}

	return br, nil
}
