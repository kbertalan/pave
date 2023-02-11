package tb

import (
	"strconv"

	tigerbeetle_go "github.com/tigerbeetledb/tigerbeetle-go"
	"github.com/tigerbeetledb/tigerbeetle-go/pkg/types"
)

const LedgerNumber = 1

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

	_, err = client.CreateAccounts(accounts)
	return err
}

func Uint128(value uint64) types.Uint128 {
	x, err := types.HexStringToUint128(strconv.FormatUint(value, 10))
	if err != nil {
		panic(err)
	}
	return x
}
