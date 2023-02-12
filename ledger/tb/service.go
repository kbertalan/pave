package tb

import (
	"context"
	"errors"
	"fmt"

	tigerbeetle_go "github.com/tigerbeetledb/tigerbeetle-go"
	"github.com/tigerbeetledb/tigerbeetle-go/pkg/types"

	"encore.app/ledger/model"
)

var (
	ErrLowBalance = errors.New("balance is too low")
)

type Service struct {
	factory *Factory
}

func NewTigerBeetleActivities(factory *Factory) *Service {
	return &Service{factory: factory}
}

func (tba *Service) getAccount(client tigerbeetle_go.Client, accountID model.AccountID) (*Account, error) {
	accounts, err := client.LookupAccounts([]types.Uint128{Uint128(uint64(accountID))})
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		return nil, fmt.Errorf("could not find account with id %d", accountID)
	}

	account := Account(accounts[0])
	return &account, nil
}

func (tba *Service) GetBalance(ctx context.Context, accountID model.AccountID) (model.Amount, error) {
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

func (tba *Service) GetAvailableBalance(ctx context.Context, accountID model.AccountID) (model.Amount, error) {
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

func (tba *Service) Authorize(ctx context.Context, req PendingAuthorizeRequest) error {
	c, err := tba.factory.NewClient()
	if err != nil {
		return err
	}
	defer c.Close()

	account, err := tba.getAccount(c, req.CreditAccountID)
	if err != nil {
		return err
	}

	_, err = c.CreateTransfers([]types.Transfer{
		{
			ID:              Uint128(uint64(req.ID)),
			DebitAccountID:  Uint128(uint64(GodID)),
			CreditAccountID: account.ID,
			Ledger:          LedgerNumber,
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

type CancelAuthorizeRequest struct {
	ID        model.CancelID
	PendingID model.PendingID
}

func (tba *Service) Cancel(ctx context.Context, req CancelAuthorizeRequest) error {
	c, err := tba.factory.NewClient()
	if err != nil {
		return err
	}
	defer c.Close()

	_, err = c.CreateTransfers([]types.Transfer{
		{
			ID:        Uint128(uint64(req.ID)),
			PendingID: Uint128(uint64(req.PendingID)),
			Flags: types.TransferFlags{
				VoidPendingTransfer: true,
			}.ToUint16(),
		},
	})
	if err != nil {
		return err
	}

	return nil
}

type TransferRequest struct {
	ID              model.TransferID
	CreditAccountID model.AccountID
	Amount          model.TransferAmount
}

func (tba *Service) Transfer(ctx context.Context, req TransferRequest) error {
	c, err := tba.factory.NewClient()
	if err != nil {
		return err
	}
	defer c.Close()

	account, err := tba.getAccount(c, req.CreditAccountID)
	if err != nil {
		return err
	}

	balance := account.AvailableBalance()
	if balance < 0 || uint64(balance) < uint64(req.Amount) {
		return ErrLowBalance
	}

	_, err = c.CreateTransfers([]types.Transfer{
		{
			ID:              Uint128(uint64(req.ID)),
			DebitAccountID:  Uint128(uint64(GodID)),
			CreditAccountID: Uint128(uint64(req.CreditAccountID)),
			Ledger:          LedgerNumber,
			Code:            1,
			Amount:          uint64(req.Amount),
		},
	})
	return err
}

type CompleteAuthorizeRequest struct {
	ID        model.TransferID
	PendingID model.PendingID
	Amount    model.TransferAmount
}

func (tba *Service) Complete(ctx context.Context, req CompleteAuthorizeRequest) error {
	c, err := tba.factory.NewClient()
	if err != nil {
		return err
	}
	defer c.Close()

	_, err = c.CreateTransfers([]types.Transfer{
		{
			ID:        Uint128(uint64(req.ID)),
			PendingID: Uint128(uint64(req.PendingID)),
			Amount:    uint64(req.Amount),
		},
	})
	return err
}
