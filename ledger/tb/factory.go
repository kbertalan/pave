package tb

import (
	tigerbeetle_go "github.com/tigerbeetledb/tigerbeetle-go"
	"github.com/tigerbeetledb/tigerbeetle-go/pkg/types"

	"encore.app/ledger/model"
)

const LedgerNumber = 1

var GodID = model.AccountID((1 << 63) - 1)

type Config struct {
	ClusterID      uint32
	Addresses      []string
	MaxConcurrency uint
}

type Factory struct {
	cfg Config
}

func NewFactory(cfg Config) *Factory {
	return &Factory{cfg: cfg}
}

func (f *Factory) NewClient() (tigerbeetle_go.Client, error) {
	client, err := tigerbeetle_go.NewClient(f.cfg.ClusterID, f.cfg.Addresses, f.cfg.MaxConcurrency)
	return client, err
}

func (f *Factory) RegisterDemoAccounts(count uint) error {
	client, err := f.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()

	accounts := make([]types.Account, 0, count)
	for i := 1; uint(i) <= count; i++ {

		accounts = append(accounts, types.Account{
			ID:     Uint128(uint64(i)),
			Ledger: LedgerNumber,
			Code:   1,
		})
	}
	accounts = append(accounts, types.Account{
		ID:     Uint128(uint64(GodID)),
		Ledger: LedgerNumber,
		Code:   1,
	})

	_, err = client.CreateAccounts(accounts)
	return err
}
