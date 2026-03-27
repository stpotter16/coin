package types

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"
)

type TransactionDTO struct {
	ID                    int
	PlaidTransactionID    string
	AccountID             int
	Amount                float64
	TransactionDate       string // YYYY-MM-DD
	Description           string
	MerchantName          sql.NullString
	Pending               bool
	PaymentChannel        string
	PlaidCategoryPrimary  sql.NullString
	PlaidCategoryDetailed sql.NullString
	CreatedTime           time.Time
	LastModifiedTime      time.Time
}

type Transaction struct {
	ID                    int
	PlaidTransactionID    string
	AccountID             int
	Amount                float64
	TransactionDate       time.Time
	Description           string
	MerchantName          MerchantName
	Pending               bool
	PaymentChannel        string
	PlaidCategoryPrimary  PlaidCategory
	PlaidCategoryDetailed PlaidCategory
	CreatedTime           time.Time
	LastModifiedTime      time.Time
}

type TransactionGroup struct {
	Date         string // formatted, e.g. "Mon, Mar 16"
	Transactions []Transaction
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

type PlaidCategory struct {
	Value *string
}

func (p PlaidCategory) Valid() bool {
	return p.Value != nil
}

func (p PlaidCategory) String() string {
	if !p.Valid() {
		return ""
	}
	words := strings.Split(strings.ToLower(*p.Value), "_")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
