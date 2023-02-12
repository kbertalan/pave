package model

type AccountID uint64
type Amount int64

type BalanceResponse struct {
	AccountID AccountID `json:"account_id"`
	Balance   Amount    `json:"balance"`
}

type AvailableBalanceResponse struct {
	AccountID        AccountID `json:"account_id"`
	AvailableBalance Amount    `json:"available_balance"`
}
