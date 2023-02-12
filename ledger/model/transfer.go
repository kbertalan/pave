package model

type PendingID uint64
type CancelID uint64
type TransferAmount uint64

type AuthorizeResponse struct {
	ReferenceID string
}
