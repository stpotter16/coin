package types

import "time"

type PlaidItem struct {
	ID                int
	PlaidItemID       string
	PlaidAccessToken  string // AES-256-GCM encrypted
	InstitutionID     string
	InstitutionName   string
	TransactionCursor *string // nil before first sync
	CreatedTime       time.Time
	LastModifiedTime  time.Time
}
