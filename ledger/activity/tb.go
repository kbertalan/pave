package activity

import (
	"context"
	"fmt"

	tigerbeetle_go "github.com/tigerbeetledb/tigerbeetle-go"
	"github.com/tigerbeetledb/tigerbeetle-go/pkg/types"

	"encore.app/ledger/model"
	"encore.app/ledger/tb"
)

type TigerBeetleActivities struct {
	factory *tb.Factory
}

func NewTigerBeetleActivities(factory *tb.Factory) *TigerBeetleActivities {
	return &TigerBeetleActivities{factory: factory}
}

func (tba *TigerBeetleActivities) getAccount(client tigerbeetle_go.Client, accountID model.AccountID) (*tb.Account, error) {
	accounts, err := client.LookupAccounts([]types.Uint128{tb.Uint128(uint64(accountID))})
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		return nil, fmt.Errorf("could not find account with id %d", accountID)
	}

	account := tb.Account(accounts[0])
	return &account, nil
}
func (tba *TigerBeetleActivities) GetBalance(ctx context.Context, accountID model.AccountID) (model.Amount, error) {
	c, err := tba.factory.NewClient()
	if err != nil {
		return 0, err
	}
	defer c.Close()

	account, err := tba.getAccount(c, accountID)
	if err != nil {
		return 0, err
	}
	return model.Amount(account.Balance()), nil
}

func (tba *TigerBeetleActivities) GetAvailableBalance(ctx context.Context, accountID model.AccountID) (model.Amount, error) {
	c, err := tba.factory.NewClient()
	if err != nil {
		return 0, err
	}
	defer c.Close()

	account, err := tba.getAccount(c, accountID)
	if err != nil {
		return 0, err
	}
	return model.Amount(account.AvailableBalance()), nil
}

type PendingAuthorizeRequest struct {
	ID              model.PendingID
	CreditAccountID model.AccountID
	Amount          model.TransferAmount
}

func (tba *TigerBeetleActivities) Authorize(ctx context.Context, req PendingAuthorizeRequest) error {
	c, err := tba.factory.NewClient()
	if err != nil {
		return err
	}
	defer c.Close()

	account, err := tba.getAccount(c, req.CreditAccountID)
	if err != nil {
		return err
	}

	//if available := account.AvailableBalance(); available < 0 || uint64(available) < uint64(req.Amount) {
	//	return fmt.Errorf("balance has not enough assets for authorization")
	//}

	_, err = c.CreateTransfers([]types.Transfer{
		{
			ID:              tb.Uint128(uint64(req.ID)),
			DebitAccountID:  tb.Uint128(uint64(tb.GodID)),
			CreditAccountID: account.ID,
			Ledger:          tb.LedgerNumber,
			Code:            1,
			Amount:          uint64(req.Amount),
			Flags: types.TransferFlags{
				Pending: true,
			}.ToUint16(),
		},
	})
	if err != nil {
		return err
	}

	return nil
}
