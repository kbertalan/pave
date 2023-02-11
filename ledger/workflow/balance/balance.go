package balance

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"encore.app/ledger/activity"
	"encore.app/ledger/model"
)

func GetBalance(ctx workflow.Context, accountId model.AccountID) (*model.BalanceResponse, error) {
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 5,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	var tba *activity.TigerBeetleActivities
	var result uint64
	err := workflow.ExecuteActivity(ctx, tba.GetBalance, accountId).Get(ctx, &result)
	if err != nil {
		return nil, err
	}

	return &model.BalanceResponse{
		AccountID: accountId,
		Balance:   result,
	}, nil
}
