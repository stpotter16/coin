package types

import "time"

type Account struct {
	ID               int
	PlaidAccountID   string
	PlaidItemID      int
	Name             string
	OfficialName     *string
	Type             string
	Subtype          string
	CurrentBalance   *float64
	AvailableBalance *float64
	IsoCurrencyCode  string
	CreatedTime      time.Time
	LastModifiedTime time.Time
}
