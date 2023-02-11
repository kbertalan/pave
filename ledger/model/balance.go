package model

type AccountID uint64

type BalanceResponse struct {
	AccountID AccountID `json:"account_id"`
	Balance   uint64    `json:"balance"`
}
