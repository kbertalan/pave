package activity

import (
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

func (tba *TigerBeetleActivities) GetBalance(accountID model.AccountID) (uint64, error) {
	c, err := tba.factory.NewClient()
	if err != nil {
		return 0, err
	}
	defer c.Close()

	accounts, err := c.LookupAccounts([]types.Uint128{tb.Uint128(uint64(accountID))})
	if err != nil {
		return 0, err
	}

	return accounts[0].DebitsPending, nil
}
