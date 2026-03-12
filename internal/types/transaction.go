package types

import "time"

type Transaction struct {
	ID                    int
	PlaidTransactionID    string
	AccountID             int
	Amount                float64
	TransactionDate       string // YYYY-MM-DD
	Description           string
	MerchantName          *string
	Pending               bool
	PaymentChannel        string
	PlaidCategoryPrimary  *string
	PlaidCategoryDetailed *string
	CategoryID            *int
	LastModifiedBy        *int
	CreatedTime           time.Time
	LastModifiedTime      time.Time
}
