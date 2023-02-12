package tb

import (
	"fmt"

	"github.com/tigerbeetledb/tigerbeetle-go/pkg/types"
)

func Uint128(value uint64) types.Uint128 {
	x, err := types.HexStringToUint128(fmt.Sprintf("%x", value))
	if err != nil {
		panic(err)
	}
	return x
}

type Account types.Account

func (a Account) Balance() int64 {
	return int64(a.DebitsPosted) - int64(a.CreditsPosted)
}

func (a Account) AvailableBalance() int64 {
	return int64(a.DebitsPosted) - int64(a.CreditsPosted) - int64(a.CreditsPending)
}
