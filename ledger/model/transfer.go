package model

import "github.com/google/uuid"

type TransferID uint64

type PendingID uint64

type CancelID uint64

type TransferAmount uint64

type ReferenceID string

func NewReferenceID() ReferenceID {
	return ReferenceID(uuid.New().String())
}

type AuthorizeResponse struct {
	ReferenceID ReferenceID
}

type PresentResponse struct {
	ReferenceID ReferenceID
}
