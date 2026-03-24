package types

import (
	"database/sql"
	"fmt"
	"time"
)

type AccountDTO struct {
	ID               int
	PlaidAccountID   string
	PlaidItemID      int
	Name             string
	OfficialName     sql.NullString
	Type             string
	Subtype          string
	CurrentBalance   sql.NullFloat64
	AvailableBalance sql.NullFloat64
	IsoCurrencyCode  string
	CreatedTime      time.Time
	LastModifiedTime time.Time
}

type Account struct {
	ID               int
	PlaidAccountID   string
	PlaidItemID      int
	Name             string
	OfficialName     AccountName
	Type             string
	Subtype          string
	CurrentBalance   Balance
	AvailableBalance Balance
	IsoCurrencyCode  string
	CreatedTime      time.Time
	LastModifiedTime time.Time
}

func (a Account) LastSynced() string {
	return a.LastModifiedTime.Format("Jan 2, 2006 3:04 PM")
}

type AccountName struct {
	Value *string
}

func (a AccountName) isNil() bool {
	return a.Value == nil
}

func (a AccountName) String() string {
	if a.isNil() {
		return ""
	}
	return *a.Value
}

func (a AccountName) Valid() bool {
	return !a.isNil()
}

type Balance struct {
	Value *float64
}

func (b Balance) isNil() bool {
	return b.Value == nil
}

func (b Balance) Float64() float64 {
	if b.isNil() {
		return 0
	}
	return *b.Value
}

func (b Balance) String() string {
	if b.isNil() {
		return "-"
	}
	return fmt.Sprintf("$%.2f", *b.Value)
}

func (b Balance) Valid() bool {
	return !b.isNil()
}
