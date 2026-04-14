package types

import (
	"database/sql"
	"fmt"
	"math"
	"time"
)

type TransactionDTO struct {
	ID               int
	AccountID        sql.NullInt64
	AccountName      string
	Amount           float64
	TransactionDate  string // YYYY-MM-DD
	Description      string
	MerchantName     sql.NullString
	Pending          bool
	PlanItemID       sql.NullInt64
	PlanItemName     sql.NullString
	CreatedTime      time.Time
	LastModifiedTime time.Time
}

// AssignedPlanItem holds the plan item a transaction has been assigned to.
type AssignedPlanItem struct {
	ID   int
	Name string
}

type Transaction struct {
	ID               int
	AccountID        *int
	AccountName      string
	Amount           float64
	TransactionDate  time.Time
	Description      string
	MerchantName     MerchantName
	Pending          bool
	PlanItem         *AssignedPlanItem // nil if unassigned
	CreatedTime      time.Time
	LastModifiedTime time.Time
}

func (t Transaction) IsAssigned() bool {
	return t.PlanItem != nil
}

type TransactionGroup struct {
	Date         string // formatted, e.g. "Mon, Mar 16"
	Transactions []Transaction
}

type TransactionPage struct {
	Transactions []Transaction
	HasMore      bool
}

func (t Transaction) DisplayName() string {
	if t.MerchantName.Valid() {
		return t.MerchantName.String()
	}
	return t.Description
}

func (t Transaction) FormattedAmount() string {
	return fmt.Sprintf("$%.2f", math.Abs(t.Amount))
}

func (t Transaction) AmountClass() string {
	if t.Amount < 0 {
		return "text-success"
	}
	return "text-error"
}

func (t Transaction) FormattedDate() string {
	return t.TransactionDate.Format("Mon, Jan 2, 2006")
}

func (t Transaction) GroupDate() string {
	return t.TransactionDate.Format("Mon, Jan 2")
}

// AbsAmount returns the absolute value of the amount for display in forms.
func (t Transaction) AbsAmount() float64 {
	return math.Abs(t.Amount)
}

// IsIncome returns true if the amount is negative (money in).
func (t Transaction) IsIncome() bool {
	return t.Amount < 0
}

// FormattedTransactionDate returns the date in YYYY-MM-DD format for form inputs.
func (t Transaction) FormattedTransactionDate() string {
	return t.TransactionDate.Format("2006-01-02")
}

// AccountIDInt returns the account ID as an int, or 0 if no account is set.
func (t Transaction) AccountIDInt() int {
	if t.AccountID == nil {
		return 0
	}
	return *t.AccountID
}

type MerchantName struct {
	Value *string
}

func (m MerchantName) Valid() bool {
	return m.Value != nil && *m.Value != ""
}

func (m MerchantName) String() string {
	if !m.Valid() {
		return ""
	}
	return *m.Value
}
